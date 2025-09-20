// (c) Cartesi and individual authors (see AUTHORS)
// SPDX-License-Identifier: Apache-2.0 (see LICENSE)

package prt

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	. "github.com/cartesi/rollups-node/internal/model"
	"github.com/cartesi/rollups-node/internal/repository"
	"github.com/cartesi/rollups-node/pkg/contracts/idaveconsensus"
	"github.com/cartesi/rollups-node/pkg/contracts/itournament"
)

type prtRepository interface {
	ListApplications(ctx context.Context, f repository.ApplicationFilter,
		p repository.Pagination, descending bool) ([]*Application, uint64, error)
	UpdateApplicationState(ctx context.Context, appID int64, state ApplicationState, reason *string) error

	ListEpochs(ctx context.Context, nameOrAddress string, f repository.EpochFilter,
		p repository.Pagination, descending bool) ([]*Epoch, uint64, error)
	UpdateEpoch(ctx context.Context, nameOrAddress string, e *Epoch) error
	UpdateEpochStatus(ctx context.Context, nameOrAddress string, e *Epoch) error

	CreateTournament(ctx context.Context, nameOrAddress string, t *Tournament) error

	CreateCommitment(ctx context.Context, nameOrAddress string, c *Commitment) error
	UpdateMatch(ctx context.Context, nameOrAddress string, m *Match) error
	CreateMatch(ctx context.Context, nameOrAddress string, m *Match) error
	CreateMatchAdvanced(ctx context.Context, nameOrAddress string, m *MatchAdvanced) error

	SaveNodeConfigRaw(ctx context.Context, key string, rawJSON []byte) error
	LoadNodeConfigRaw(ctx context.Context, key string) (rawJSON []byte, createdAt, updatedAt time.Time, err error)
}

// EthClientInterface defines the methods we need from ethclient.Client
type EthClientInterface interface {
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	ChainID(ctx context.Context) (*big.Int, error)
}

func getAllRunningApplications(ctx context.Context, r prtRepository) ([]*Application, uint64, error) {
	f := repository.ApplicationFilter{State: Pointer(ApplicationState_Enabled), ConsensusType: Pointer(Consensus_PRT)}
	return r.ListApplications(ctx, f, repository.Pagination{}, false)
}

func getAllClaimComputedEpochs(ctx context.Context, r prtRepository, nameOrAddress string) ([]*Epoch, uint64, error) {
	f := repository.EpochFilter{Status: Pointer(EpochStatus_ClaimComputed)}
	return r.ListEpochs(ctx, nameOrAddress, f, repository.Pagination{}, false)
}

// setApplicationInoperable marks an application as inoperable with the given reason,
// logs any error that occurs during the update, and returns an error with the reason.
func (s *Service) setApplicationInoperable(ctx context.Context, app *Application, reasonFmt string, args ...any) error {
	reason := fmt.Sprintf(reasonFmt, args...)
	appAddress := app.IApplicationAddress.String()

	// Log the reason first
	s.Logger.Error(reason, "application", appAddress)

	// Update application state
	err := s.repository.UpdateApplicationState(ctx, app.ID, ApplicationState_Inoperable, &reason)
	if err != nil {
		s.Logger.Error("failed to update application state to inoperable", "app", appAddress, "err", err)
	}

	// Return the error with the reason
	return errors.New(reason)
}

func (s *Service) saveCommitments(ctx context.Context, app *Application, epoch *Epoch,
	tournament *Tournament, events []*itournament.ITournamentCommitmentJoined) error {
	for _, ev := range events {
		c := Commitment{
			ApplicationID:     app.ID,
			EpochIndex:        epoch.Index,
			TournamentAddress: tournament.Address,
			Commitment:        ev.Commitment,
			FinalStateHash:    ev.FinalStateHash,
			SubmitterAddress:  ev.Submitter,
			BlockNumber:       ev.Raw.BlockNumber,
			TxHash:            ev.Raw.TxHash,
		}
		s.Logger.Info("CommitmentJoined event",
			"application", app.Name,
			"epoch_index", epoch.Index,
			"commitment", c.Commitment.String())
		err := s.repository.CreateCommitment(ctx, app.IApplicationAddress.Hex(), &c)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) saveMatches(ctx context.Context, app *Application, epoch *Epoch,
	tournament *Tournament, events []*itournament.ITournamentMatchCreated) error {
	for _, ev := range events {
		m := Match{
			ApplicationID:       app.ID,
			EpochIndex:          epoch.Index,
			TournamentAddress:   tournament.Address,
			IDHash:              ev.MatchIdHash,
			CommitmentOne:       ev.One,
			CommitmentTwo:       ev.Two,
			LeftOfTwo:           ev.LeftOfTwo,
			BlockNumber:         ev.Raw.BlockNumber,
			TxHash:              ev.Raw.TxHash,
			Winner:              WinnerCommitment_NONE,
			DeletionReason:      MatchDeletionReason_NOT_DELETED,
			DeletionBlockNumber: 0,
			DeletionTxHash:      common.Hash{},
		}
		s.Logger.Info("MatchCreated event",
			"application", app.Name,
			"epoch_index", epoch.Index,
			"tournament", tournament.Address.Hex(),
			"id_hash", m.IDHash.String(),
			"one", m.CommitmentOne.String(),
			"two", m.CommitmentTwo.String(),
			"leftOfTwo", m.LeftOfTwo.String())
		err := s.repository.CreateMatch(ctx, app.IApplicationAddress.Hex(), &m)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) saveMatchAdvanced(ctx context.Context, app *Application, epoch *Epoch,
	tournament *Tournament, events []*itournament.ITournamentMatchAdvanced) error {
	for _, ev := range events {
		m := &MatchAdvanced{
			ApplicationID:     app.ID,
			EpochIndex:        epoch.Index,
			TournamentAddress: tournament.Address,
			IDHash:            ev.MatchIdHash,
			OtherParent:       ev.OtherParent,
			LeftNode:          ev.LeftNode,
			BlockNumber:       ev.Raw.BlockNumber,
			TxHash:            ev.Raw.TxHash,
		}
		s.Logger.Info("MatchAdvanced event",
			"application", app.Name,
			"epoch_index", epoch.Index,
			"tournament", tournament.Address.Hex(),
			"id_hash", m.IDHash.String(),
			"other_parent", m.OtherParent.String(),
			"left_node", m.LeftNode.String())
		err := s.repository.CreateMatchAdvanced(ctx, app.IApplicationAddress.Hex(), m)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) saveMatchDeleted(ctx context.Context, app *Application, epoch *Epoch,
	tournament *Tournament, events []*itournament.ITournamentMatchDeleted) error {
	for _, ev := range events {
		m := Match{
			ApplicationID:       app.ID,
			EpochIndex:          epoch.Index,
			TournamentAddress:   tournament.Address,
			IDHash:              ev.MatchIdHash,
			CommitmentOne:       ev.One,
			CommitmentTwo:       ev.Two,
			Winner:              WinnerCommitmentFromUint8(ev.WinnerCommitment),
			DeletionReason:      MatchDeletionReasonFromUint8(ev.Reason),
			DeletionBlockNumber: ev.Raw.BlockNumber,
			DeletionTxHash:      ev.Raw.TxHash,
		}
		// For now, just log
		s.Logger.Info("MatchDeleted event",
			"application", app.Name,
			"epoch_index", epoch.Index,
			"tournament", tournament.Address.Hex(),
			"id_hash", ((common.Hash)(ev.MatchIdHash)).String(),
			"one", ((common.Hash)(ev.One)).String(),
			"two", ((common.Hash)(ev.Two)).String(),
			"winner", m.Winner.String(),
			"reason", m.DeletionReason.String(),
		)
		err := s.repository.UpdateMatch(ctx, app.IApplicationAddress.Hex(), &m)
		if err != nil {
			return err
		}

	}
	return nil
}

func (s *Service) checkFinalizedEpochs(ctx context.Context, app *Application) error {
	epochs, _, err := getAllClaimComputedEpochs(ctx, s.repository, app.Name)
	if err != nil {
		s.Logger.Error("failed to list epochs", "application", app.Name, "error", err)
		return err
	}
	if len(epochs) == 0 {
		return nil // nothing to do
	}

	// TODO: use adapters instead of direct contract calls
	// Type assertion to get the concrete client if possible
	ethClient, ok := s.client.(*ethclient.Client)
	if !ok {
		return fmt.Errorf("client is not an *ethclient.Client, cannot create dave consensus bind")
	}

	consensus, err := idaveconsensus.NewIDaveConsensus(app.IConsensusAddress, ethClient)
	if err != nil {
		s.Logger.Error("failed to bind dave consensus contract", "application", app.Name,
			"consensus_address", app.IConsensusAddress.String(), "error", err)
		return err
	}

	for _, epoch := range epochs {
		if epoch.ClaimTransactionHash == nil {
			break
		}
		receipt, err := ethClient.TransactionReceipt(ctx, *epoch.ClaimTransactionHash)
		if err != nil {
			s.Logger.Error("failed to fetch transaction receipt for epoch", "application", app.Name,
				"epoch", epoch.Index, "tx", epoch.ClaimTransactionHash, "error", err)
			return err
		}

		if receipt.Status != 1 {
			return fmt.Errorf("EpochSealed transaction hash points to failed transaction")
		}

		var event *idaveconsensus.IDaveConsensusEpochSealed
		for _, vLog := range receipt.Logs {
			event, err = consensus.ParseEpochSealed(*vLog)
			if err != nil {
				continue // Skip logs that don't match
			}
		}
		if event == nil {
			return fmt.Errorf("failed to find EpochSealed event in receipt logs")

		}

		if epoch.Index != event.EpochNumber.Uint64()-1 {
			return s.setApplicationInoperable(ctx, app, "Epoch %d has inconsistent index between off-chain (%d) and on-chain (%d)",
				epoch.Index, epoch.Index, event.EpochNumber.Uint64()-1)
		}
		if *epoch.MachineHash != event.InitialMachineStateHash {
			return s.setApplicationInoperable(ctx, app, "Epoch %d has inconsistent machine hash between off-chain (%s) and on-chain (%s)",
				epoch.Index, epoch.MachineHash.String(), hexutil.Encode(event.InitialMachineStateHash[:]))
		}
		if *epoch.ClaimHash != event.OutputsMerkleRoot {
			return s.setApplicationInoperable(ctx, app, "Epoch %d has inconsistent claim hash between off-chain (%s) and on-chain (%s)",
				epoch.Index, epoch.ClaimHash.String(), hexutil.Encode(event.OutputsMerkleRoot[:]))
		}

		err = s.fetchRootTournamentData(ctx, app, epoch)
		if err != nil {
			s.Logger.Error("failed to fetch tournament data", "application", app.Name,
				"epoch", epoch.Index, "tournament", epoch.TournamentAddress.String(), "error", err)
			return err
		}

		s.Logger.Info("Found finalized epoch. OutputsMerkleRoot matched. Setting claim as accepted",
			"application", app.Name,
			"epoch", epoch.Index,
			"event_block_number", event.Raw.BlockNumber,
			"claim_hash", fmt.Sprintf("%x", event.OutputsMerkleRoot),
			"tx", epoch.ClaimTransactionHash,
		)

		epoch.Status = EpochStatus_ClaimAccepted
		err = s.repository.UpdateEpochStatus(ctx, app.Name, epoch)
		if err != nil {
			s.Logger.Error("failed to update epoch status to claim accepted", "application", app.Name, "epoch", epoch.Index, "error", err)
			return err
		}
	}
	return nil
}

func (s *Service) saveTournamentEvents(ctx context.Context, app *Application, epoch *Epoch, t *Tournament, events *TournamentEvents) error {
	err := s.saveCommitments(ctx, app, epoch, t, events.CommitmentJoined)
	if err != nil {
		s.Logger.Error("failed to save commitments and matches", "application", app.Name,
			"epoch", epoch.Index, "tournament_address", t.Address.String(), "error", err)
		return err
	}

	err = s.saveMatches(ctx, app, epoch, t, events.MatchCreated)
	if err != nil {
		s.Logger.Error("failed to save commitments and matches", "application", app.Name,
			"epoch", epoch.Index, "tournament_address", t.Address.String(), "error", err)
		return err
	}

	err = s.saveMatchAdvanced(ctx, app, epoch, t, events.MatchAdvanced)
	if err != nil {
		s.Logger.Error("failed to save commitments and matches", "application", app.Name,
			"epoch", epoch.Index, "tournament_address", t.Address.String(), "error", err)
		return err
	}

	err = s.saveMatchDeleted(ctx, app, epoch, t, events.MatchDeleted)
	if err != nil {
		s.Logger.Error("failed to save commitments and matches", "application", app.Name,
			"epoch", epoch.Index, "tournament_address", t.Address.String(), "error", err)
		return err
	}
	return nil
}

func (s *Service) fetchRootTournamentData(ctx context.Context, app *Application, epoch *Epoch) error {
	// TODO: use adapters instead of direct contract calls
	// Type assertion to get the concrete client if possible
	ethClient, ok := s.client.(*ethclient.Client)
	if !ok {
		return fmt.Errorf("client is not an *ethclient.Client, cannot create dave consensus bind")
	}

	adapter, err := NewITournamentAdapter(*epoch.TournamentAddress, ethClient, s.filter)
	if err != nil {
		s.Logger.Error("failed to create top tournament adapter", "application", app.Name,
			"epoch", epoch.Index, "tournament_address", epoch.TournamentAddress.String(), "error", err)
		return err
	}

	constants, err := adapter.Constants(nil)
	if err != nil {
		s.Logger.Error("failed to fetch tournament constants", "application", app.Name,
			"epoch", epoch.Index, "tournament_address", epoch.TournamentAddress.String(), "error", err)
		return err
	}

	finished, timeFinished, err := adapter.TimeFinished(nil)
	if err != nil {
		s.Logger.Error("failed to fetch tournament finished at time", "application", app.Name,
			"epoch", epoch.Index, "tournament_address", epoch.TournamentAddress.String(), "error", err)
		return err
	}
	if !finished {
		s.Logger.Error("tournament should be finished", "application", app.Name,
			"epoch", epoch.Index, "tournament_address", epoch.TournamentAddress.String(), "error", err)
		return err
	}

	_, winnerCommitment, finalState, err := adapter.Result(nil)
	if err != nil {
		s.Logger.Error("failed to fetch tournament result", "application", app.Name,
			"epoch", epoch.Index, "tournament_address", epoch.TournamentAddress.String(), "error", err)
		return err
	}

	t := Tournament{
		ApplicationID:    app.ID,
		EpochIndex:       epoch.Index,
		Address:          *epoch.TournamentAddress,
		MaxLevel:         constants.MaxLevel,
		Level:            constants.Level,
		Log2Step:         constants.Log2step,
		Height:           constants.Height,
		WinnerCommitment: (*common.Hash)(&winnerCommitment),
		FinalStateHash:   (*common.Hash)(&finalState),
		FinishedAtBlock:  timeFinished,
	}

	err = s.repository.CreateTournament(ctx, app.IApplicationAddress.Hex(), &t)
	if err != nil {
		s.Logger.Error("failed to create tournament in database", "application", app.Name,
			"epoch", epoch.Index, "tournament_address", epoch.TournamentAddress.String(), "error", err)
		return err
	}

	opts := &bind.FilterOpts{
		Context: ctx,
		Start:   epoch.LastBlock,
		End:     &timeFinished, // To latest block
	}

	events, err := adapter.RetrieveAllEvents(opts)
	if err != nil {
		s.Logger.Error("failed to retrieve all events from tournament", "application", app.Name,
			"epoch", epoch.Index, "tournament_address", epoch.TournamentAddress.String(), "error", err)
		return err
	}

	// Print summary of events found
	s.Logger.Info("Retrieved events for root tournament", "address", t.Address.String(),
		"commitmentJoined", len(events.CommitmentJoined),
		"matchCreated", len(events.MatchCreated),
		"matchAdvanced", len(events.MatchAdvanced),
		"matchDeleted", len(events.MatchDeleted),
		"newInnerTournament", len(events.NewInnerTournament))

	err = s.saveTournamentEvents(ctx, app, epoch, &t, events)
	if err != nil {
		s.Logger.Error("failed to save events for root tournament", "application", app.Name,
			"epoch", epoch.Index, "tournament_address", t.Address.String(), "error", err)
		return err
	}

	for _, newInner := range events.NewInnerTournament {
		hashID := (common.Hash)(newInner.MatchIdHash)
		middleAddress := newInner.ChildTournament
		s.Logger.Info("NewInnerTournament event", "id_hash", hashID.String(), "tournament_address", middleAddress.String())
		err = s.fetchMiddleTournamentData(ctx, app, epoch, &hashID, &middleAddress)
		if err != nil {
			s.Logger.Error("failed to fetch middle tournament data", "application", app.Name,
				"tournament", middleAddress.String(), "error", err)
			return err
		}
	}

	return nil
}

func (s *Service) fetchMiddleTournamentData(ctx context.Context, app *Application, epoch *Epoch, hashID *common.Hash, middleAddress *common.Address) error {
	s.Logger.Info("Fetching middle tournament data", "application", app.Name, "tournament", middleAddress.String())
	// TODO: use adapters instead of direct contract calls
	// Type assertion to get the concrete client if possible
	ethClient, ok := s.client.(*ethclient.Client)
	if !ok {
		return fmt.Errorf("client is not an *ethclient.Client, cannot create dave consensus bind")
	}

	adapter, err := NewITournamentAdapter(*middleAddress, ethClient, s.filter)
	if err != nil {
		s.Logger.Error("failed to create middle tournament adapter", "application", app.Name,
			"epoch", epoch.Index, "tournament_address", middleAddress.String(), "error", err)
		return err
	}

	constants, err := adapter.Constants(nil)
	if err != nil {
		s.Logger.Error("failed to fetch middle tournament constants", "application", app.Name,
			"epoch", epoch.Index, "tournament_address", middleAddress.String(), "error", err)
		return err
	}

	finished, timeFinished, err := adapter.TimeFinished(nil)
	if err != nil {
		s.Logger.Error("failed to fetch tournament finished at time", "application", app.Name,
			"epoch", epoch.Index, "tournament_address", middleAddress.String(), "error", err)
		return err
	}
	if !finished {
		s.Logger.Error("tournament should be finished", "application", app.Name,
			"epoch", epoch.Index, "tournament_address", middleAddress.String(), "error", err)
		return err
	}

	_, winnerCommitment, finalState, err := adapter.Result(nil)
	if err != nil {
		s.Logger.Error("failed to fetch tournament result", "application", app.Name,
			"epoch", epoch.Index, "tournament_address", epoch.TournamentAddress.String(), "error", err)
		return err
	}

	t := Tournament{
		ApplicationID:           app.ID,
		EpochIndex:              epoch.Index,
		Address:                 *middleAddress,
		ParentMatchIDHash:       hashID,
		ParentTournamentAddress: epoch.TournamentAddress,
		MaxLevel:                constants.MaxLevel,
		Level:                   constants.Level,
		Log2Step:                constants.Log2step,
		Height:                  constants.Height,
		WinnerCommitment:        (*common.Hash)(&winnerCommitment),
		FinalStateHash:          (*common.Hash)(&finalState),
		FinishedAtBlock:         timeFinished,
	}

	err = s.repository.CreateTournament(ctx, app.IApplicationAddress.Hex(), &t)
	if err != nil {
		s.Logger.Error("failed to create middle tournament in database", "application", app.Name,
			"epoch", epoch.Index, "tournament_address", middleAddress.String(), "error", err)
		return err
	}

	opts := &bind.FilterOpts{
		Context: ctx,
		Start:   epoch.LastBlock,
		End:     &timeFinished, // To latest block
	}

	events, err := adapter.RetrieveAllEvents(opts)
	if err != nil {
		s.Logger.Error("failed to retrieve all events from tournament", "application", app.Name,
			"epoch", epoch.Index, "tournament_address", middleAddress.String(), "error", err)
		return err
	}

	// Print summary of events found
	s.Logger.Info("Retrieved events for middle tournament", "address", t.Address.String(),
		"commitmentJoined", len(events.CommitmentJoined),
		"matchCreated", len(events.MatchCreated),
		"matchAdvanced", len(events.MatchAdvanced),
		"matchDeleted", len(events.MatchDeleted),
		"newInnerTournament", len(events.NewInnerTournament))

	err = s.saveTournamentEvents(ctx, app, epoch, &t, events)
	if err != nil {
		s.Logger.Error("failed to save events for middle tournament", "application", app.Name,
			"epoch", epoch.Index, "tournament_address", t.Address.String(), "error", err)
		return err
	}

	for _, newInner := range events.NewInnerTournament {
		hashID := (common.Hash)(newInner.MatchIdHash)
		bottomAddress := newInner.ChildTournament
		s.Logger.Info("NewInnerTournament event", "id_hash", hashID.String(), "tournament_address", bottomAddress.String())
		err = s.fetchBottomTournamentData(ctx, app, epoch, &hashID, middleAddress, &bottomAddress)
		if err != nil {
			s.Logger.Error("failed to fetch bottom tournament data", "application", app.Name,
				"tournament", bottomAddress.String(), "error", err)
			return err
		}
	}
	return nil
}

func (s *Service) fetchBottomTournamentData(
	ctx context.Context,
	app *Application,
	epoch *Epoch,
	hashID *common.Hash,
	middleAddress, bottomAddress *common.Address,
) error {
	s.Logger.Info("Fetching bottom tournament data", "application", app.Name, "tournament", bottomAddress.String())
	// TODO: use adapters instead of direct contract calls
	// Type assertion to get the concrete client if possible
	ethClient, ok := s.client.(*ethclient.Client)
	if !ok {
		return fmt.Errorf("client is not an *ethclient.Client, cannot create dave consensus bind")
	}

	adapter, err := NewITournamentAdapter(*bottomAddress, ethClient, s.filter)
	if err != nil {
		s.Logger.Error("failed to create bottom tournament adapter", "application", app.Name,
			"epoch", epoch.Index, "tournament_address", bottomAddress.String(), "error", err)
		return err
	}

	constants, err := adapter.Constants(nil)
	if err != nil {
		s.Logger.Error("failed to fetch bottom tournament constants", "application", app.Name,
			"epoch", epoch.Index, "tournament_address", bottomAddress.String(), "error", err)
		return err
	}

	finished, timeFinished, err := adapter.TimeFinished(nil)
	if err != nil {
		s.Logger.Error("failed to fetch tournament finished at time", "application", app.Name,
			"epoch", epoch.Index, "tournament_address", bottomAddress.String(), "error", err)
		return err
	}
	if !finished {
		s.Logger.Error("tournament should be finished", "application", app.Name,
			"epoch", epoch.Index, "tournament_address", bottomAddress.String(), "error", err)
		return err
	}

	_, winnerCommitment, finalState, err := adapter.Result(nil)
	if err != nil {
		s.Logger.Error("failed to fetch tournament result", "application", app.Name,
			"epoch", epoch.Index, "tournament_address", epoch.TournamentAddress.String(), "error", err)
		return err
	}

	t := Tournament{
		ApplicationID:           app.ID,
		EpochIndex:              epoch.Index,
		Address:                 *bottomAddress,
		ParentMatchIDHash:       hashID,
		ParentTournamentAddress: middleAddress, // Assuming middle is the parent
		MaxLevel:                constants.MaxLevel,
		Level:                   constants.Level,
		Log2Step:                constants.Log2step,
		Height:                  constants.Height,
		WinnerCommitment:        (*common.Hash)(&winnerCommitment),
		FinalStateHash:          (*common.Hash)(&finalState),
		FinishedAtBlock:         timeFinished,
	}

	err = s.repository.CreateTournament(ctx, app.IApplicationAddress.Hex(), &t)
	if err != nil {
		s.Logger.Error("failed to create bottom tournament in database", "application", app.Name,
			"epoch", epoch.Index, "tournament_address", bottomAddress.String(), "error", err)
		return err
	}

	opts := &bind.FilterOpts{
		Context: ctx,
		Start:   epoch.LastBlock,
		End:     &timeFinished, // To latest block
	}

	events, err := adapter.RetrieveAllEvents(opts)
	if err != nil {
		s.Logger.Error("failed to retrieve all events from tournament", "application", app.Name,
			"epoch", epoch.Index, "tournament_address", bottomAddress.String(), "error", err)
		return err
	}

	// Print summary of events found
	s.Logger.Info("Retrieved events for bottom tournament", "address", bottomAddress.String(),
		"commitmentJoined", len(events.CommitmentJoined),
		"matchCreated", len(events.MatchCreated),
		"matchAdvanced", len(events.MatchAdvanced),
		"matchDeleted", len(events.MatchDeleted))

	err = s.saveTournamentEvents(ctx, app, epoch, &t, events)
	if err != nil {
		s.Logger.Error("failed to save events for bottom tournament", "application", app.Name,
			"epoch", epoch.Index, "tournament_address", t.Address.String(), "error", err)
		return err
	}

	return nil
}

func (s *Service) validateApplication(ctx context.Context, app *Application) error {
	s.Logger.Debug("Syncing PTR tournaments", "application", app.Name)
	return s.checkFinalizedEpochs(ctx, app)
}
