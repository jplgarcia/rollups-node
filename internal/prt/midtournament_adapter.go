// (c) Cartesi and individual authors (see AUTHORS)
// SPDX-License-Identifier: Apache-2.0 (see LICENSE)

package prt

import (
	"math/big"

	. "github.com/cartesi/rollups-node/internal/model"
	"github.com/cartesi/rollups-node/pkg/contracts/middletournament"
	"github.com/cartesi/rollups-node/pkg/ethutil"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// MiddleTournament Wrapper
type MiddleTournamentAdapterImpl struct {
	tournament        *middletournament.MiddleTournament
	client            *ethclient.Client
	tournamentAddress common.Address
	filter            ethutil.Filter
}

func NewMiddleTournamentAdapter(
	tournamentAddress common.Address,
	client *ethclient.Client,
	filter ethutil.Filter,
) (TournamentAdapter, error) {
	tournamentContract, err := middletournament.NewMiddleTournament(tournamentAddress, client)
	if err != nil {
		return nil, err
	}
	return &MiddleTournamentAdapterImpl{
		tournament:        tournamentContract,
		tournamentAddress: tournamentAddress,
		client:            client,
		filter:            filter,
	}, nil
}

func (a *MiddleTournamentAdapterImpl) Result(opts *bind.CallOpts) (bool, [32]byte, error) {
	finished, _, commitment, _, error := a.tournament.InnerTournamentWinner(opts)
	return finished, commitment, error
}

func (a *MiddleTournamentAdapterImpl) Constants(opts *bind.CallOpts) (TournamentConstants, error) {
	c, error := a.tournament.TournamentLevelConstants(opts)
	return TournamentConstants{
		MaxLevel: c.MaxLevel,
		Level:    c.Level,
		Log2step: c.Log2step,
		Height:   c.Height,
	}, error
}

func (a *MiddleTournamentAdapterImpl) TimeFinished(opts *bind.CallOpts) (bool, uint64, error) {
	finished, block, error := a.tournament.TimeFinished(opts)
	println("MiddleTournament TimeFinished:", finished, block, error)
	return finished, block, error
	// return a.tournament.TimeFinished(opts)
}

func buildMiddleCommitmentJoinedFilterQuery(
	opts *bind.FilterOpts,
	tournamentAddress common.Address,
) (q ethereum.FilterQuery, err error) {
	c, err := middletournament.MiddleTournamentMetaData.GetAbi()
	if err != nil {
		return q, err
	}

	topics, err := abi.MakeTopics(
		[]any{c.Events["commitmentJoined"].ID},
	)
	if err != nil {
		return q, err
	}

	q = ethereum.FilterQuery{
		Addresses: []common.Address{tournamentAddress},
		FromBlock: new(big.Int).SetUint64(opts.Start),
		Topics:    topics,
	}
	if opts.End != nil {
		q.ToBlock = new(big.Int).SetUint64(*opts.End)
	}
	return q, err
}

func (a *MiddleTournamentAdapterImpl) RetrieveCommitmentJoinedEvents(
	opts *bind.FilterOpts,
) ([]CommitmentJoinedEvent, error) {
	q, err := buildMiddleCommitmentJoinedFilterQuery(opts, a.tournamentAddress)
	if err != nil {
		return nil, err
	}

	itr, err := a.filter.ChunkedFilterLogs(opts.Context, a.client, q)
	if err != nil {
		return nil, err
	}

	var events []CommitmentJoinedEvent
	for log, err := range itr {
		if err != nil {
			return nil, err
		}
		ev, err := a.tournament.ParseCommitmentJoined(*log)
		if err != nil {
			return nil, err
		}
		events = append(events, &MiddleCommitmentJoinedEvent{ev})
	}
	return events, nil
}

func buildMiddleMatchAdvancedFilterQuery(
	opts *bind.FilterOpts,
	tournamentAddress common.Address,
) (q ethereum.FilterQuery, err error) {
	c, err := middletournament.MiddleTournamentMetaData.GetAbi()
	if err != nil {
		return q, err
	}

	topics, err := abi.MakeTopics(
		[]any{c.Events["matchAdvanced"].ID},
	)
	if err != nil {
		return q, err
	}

	q = ethereum.FilterQuery{
		Addresses: []common.Address{tournamentAddress},
		FromBlock: new(big.Int).SetUint64(opts.Start),
		Topics:    topics,
	}
	if opts.End != nil {
		q.ToBlock = new(big.Int).SetUint64(*opts.End)
	}
	return q, err
}

func (a *MiddleTournamentAdapterImpl) RetrieveMatchAdvancedEvents(
	opts *bind.FilterOpts,
) ([]MatchAdvancedEvent, error) {
	q, err := buildMiddleMatchAdvancedFilterQuery(opts, a.tournamentAddress)
	if err != nil {
		return nil, err
	}

	itr, err := a.filter.ChunkedFilterLogs(opts.Context, a.client, q)
	if err != nil {
		return nil, err
	}

	var events []MatchAdvancedEvent
	for log, err := range itr {
		if err != nil {
			return nil, err
		}
		ev, err := a.tournament.ParseMatchAdvanced(*log)
		if err != nil {
			return nil, err
		}
		events = append(events, &MiddleMatchAdvancedEvent{ev})
	}
	return events, nil
}

func buildMiddleMatchCreatedFilterQuery(
	opts *bind.FilterOpts,
	tournamentAddress common.Address,
) (q ethereum.FilterQuery, err error) {
	c, err := middletournament.MiddleTournamentMetaData.GetAbi()
	if err != nil {
		return q, err
	}

	topics, err := abi.MakeTopics(
		[]any{c.Events["matchCreated"].ID},
	)
	if err != nil {
		return q, err
	}

	q = ethereum.FilterQuery{
		Addresses: []common.Address{tournamentAddress},
		FromBlock: new(big.Int).SetUint64(opts.Start),
		Topics:    topics,
	}
	if opts.End != nil {
		q.ToBlock = new(big.Int).SetUint64(*opts.End)
	}
	return q, err
}

func (a *MiddleTournamentAdapterImpl) RetrieveMatchCreatedEvents(
	opts *bind.FilterOpts,
) ([]MatchCreatedEvent, error) {
	q, err := buildMiddleMatchCreatedFilterQuery(opts, a.tournamentAddress)
	if err != nil {
		return nil, err
	}

	itr, err := a.filter.ChunkedFilterLogs(opts.Context, a.client, q)
	if err != nil {
		return nil, err
	}

	var events []MatchCreatedEvent
	for log, err := range itr {
		if err != nil {
			return nil, err
		}
		ev, err := a.tournament.ParseMatchCreated(*log)
		if err != nil {
			return nil, err
		}
		events = append(events, &MiddleMatchCreatedEvent{ev})
	}
	return events, nil
}

func buildMiddleMatchDeletedFilterQuery(
	opts *bind.FilterOpts,
	tournamentAddress common.Address,
) (q ethereum.FilterQuery, err error) {
	c, err := middletournament.MiddleTournamentMetaData.GetAbi()
	if err != nil {
		return q, err
	}

	topics, err := abi.MakeTopics(
		[]any{c.Events["matchDeleted"].ID},
	)
	if err != nil {
		return q, err
	}

	q = ethereum.FilterQuery{
		Addresses: []common.Address{tournamentAddress},
		FromBlock: new(big.Int).SetUint64(opts.Start),
		Topics:    topics,
	}
	if opts.End != nil {
		q.ToBlock = new(big.Int).SetUint64(*opts.End)
	}
	return q, err
}

func (a *MiddleTournamentAdapterImpl) RetrieveMatchDeletedEvents(
	opts *bind.FilterOpts,
) ([]MatchDeletedEvent, error) {
	q, err := buildMiddleMatchDeletedFilterQuery(opts, a.tournamentAddress)
	if err != nil {
		return nil, err
	}

	itr, err := a.filter.ChunkedFilterLogs(opts.Context, a.client, q)
	if err != nil {
		return nil, err
	}

	var events []MatchDeletedEvent
	for log, err := range itr {
		if err != nil {
			return nil, err
		}
		ev, err := a.tournament.ParseMatchDeleted(*log)
		if err != nil {
			return nil, err
		}
		events = append(events, &MiddleMatchDeletedEvent{ev})
	}
	return events, nil
}

func buildMiddleNewInnerTournamentFilterQuery(
	opts *bind.FilterOpts,
	tournamentAddress common.Address,
) (q ethereum.FilterQuery, err error) {
	c, err := middletournament.MiddleTournamentMetaData.GetAbi()
	if err != nil {
		return q, err
	}

	topics, err := abi.MakeTopics(
		[]any{c.Events["newInnerTournament"].ID},
	)
	if err != nil {
		return q, err
	}

	q = ethereum.FilterQuery{
		Addresses: []common.Address{tournamentAddress},
		FromBlock: new(big.Int).SetUint64(opts.Start),
		Topics:    topics,
	}
	if opts.End != nil {
		q.ToBlock = new(big.Int).SetUint64(*opts.End)
	}
	return q, err
}

func (a *MiddleTournamentAdapterImpl) RetrieveNewInnerTournamentEvents(
	opts *bind.FilterOpts,
) ([]NewInnerTournamentEvent, error) {
	q, err := buildMiddleNewInnerTournamentFilterQuery(opts, a.tournamentAddress)
	if err != nil {
		return nil, err
	}

	itr, err := a.filter.ChunkedFilterLogs(opts.Context, a.client, q)
	if err != nil {
		return nil, err
	}

	var events []NewInnerTournamentEvent
	for log, err := range itr {
		if err != nil {
			return nil, err
		}
		ev, err := a.tournament.ParseNewInnerTournament(*log)
		if err != nil {
			return nil, err
		}
		events = append(events, &MiddleNewInnerTournamentEvent{ev})
	}
	return events, nil
}

func buildMiddleAllEventsFilterQuery(
	opts *bind.FilterOpts,
	tournamentAddress common.Address,
) (q ethereum.FilterQuery, err error) {
	c, err := middletournament.MiddleTournamentMetaData.GetAbi()
	if err != nil {
		return q, err
	}

	topics, err := abi.MakeTopics(
		[]any{
			c.Events[MonitoredEvent_CommitmentJoined.String()].ID,
			c.Events[MonitoredEvent_MatchAdvanced.String()].ID,
			c.Events[MonitoredEvent_MatchCreated.String()].ID,
			c.Events[MonitoredEvent_MatchDeleted.String()].ID,
			c.Events[MonitoredEvent_NewInnerTournament.String()].ID,
		},
	)
	if err != nil {
		return q, err
	}

	q = ethereum.FilterQuery{
		Addresses: []common.Address{tournamentAddress},
		FromBlock: new(big.Int).SetUint64(opts.Start),
		Topics:    topics,
	}
	if opts.End != nil {
		q.ToBlock = new(big.Int).SetUint64(*opts.End)
	}
	return q, err
}

func (a *MiddleTournamentAdapterImpl) RetrieveAllEvents(
	opts *bind.FilterOpts,
) (*TournamentEvents, error) {
	q, err := buildMiddleAllEventsFilterQuery(opts, a.tournamentAddress)
	if err != nil {
		return nil, err
	}

	itr, err := a.filter.ChunkedFilterLogs(opts.Context, a.client, q)
	if err != nil {
		return nil, err
	}

	var commitmentJoined []CommitmentJoinedEvent
	var matchAdvanced []MatchAdvancedEvent
	var matchCreated []MatchCreatedEvent
	var matchDeleted []MatchDeletedEvent
	var newInnerTournament []NewInnerTournamentEvent

	c, err := middletournament.MiddleTournamentMetaData.GetAbi()
	if err != nil {
		return nil, err
	}

	for log, err := range itr {
		if err != nil {
			return nil, err
		}

		switch log.Topics[0] {
		case c.Events[MonitoredEvent_CommitmentJoined.String()].ID:
			ev, err := a.tournament.ParseCommitmentJoined(*log)
			if err != nil {
				return nil, err
			}
			commitmentJoined = append(commitmentJoined, &MiddleCommitmentJoinedEvent{ev})
		case c.Events[MonitoredEvent_MatchAdvanced.String()].ID:
			ev, err := a.tournament.ParseMatchAdvanced(*log)
			if err != nil {
				return nil, err
			}
			matchAdvanced = append(matchAdvanced, &MiddleMatchAdvancedEvent{ev})
		case c.Events[MonitoredEvent_MatchCreated.String()].ID:
			ev, err := a.tournament.ParseMatchCreated(*log)
			if err != nil {
				return nil, err
			}
			matchCreated = append(matchCreated, &MiddleMatchCreatedEvent{ev})
		case c.Events[MonitoredEvent_MatchDeleted.String()].ID:
			ev, err := a.tournament.ParseMatchDeleted(*log)
			if err != nil {
				return nil, err
			}
			matchDeleted = append(matchDeleted, &MiddleMatchDeletedEvent{ev})
		case c.Events[MonitoredEvent_NewInnerTournament.String()].ID:
			ev, err := a.tournament.ParseNewInnerTournament(*log)
			if err != nil {
				return nil, err
			}
			newInnerTournament = append(newInnerTournament, &MiddleNewInnerTournamentEvent{ev})
		}
	}

	return &TournamentEvents{
		CommitmentJoined:   commitmentJoined,
		MatchAdvanced:      matchAdvanced,
		MatchCreated:       matchCreated,
		MatchDeleted:       matchDeleted,
		NewInnerTournament: newInnerTournament,
	}, nil
}

// Wrappers for MiddleTournament events
type MiddleCommitmentJoinedEvent struct {
	*middletournament.MiddleTournamentCommitmentJoined
}

func (e *MiddleCommitmentJoinedEvent) Root() [32]byte      { return e.MiddleTournamentCommitmentJoined.Root }
func (e *MiddleCommitmentJoinedEvent) BlockNumber() uint64 { return e.Raw.BlockNumber }
func (e *MiddleCommitmentJoinedEvent) TxHash() common.Hash { return e.Raw.TxHash }

type MiddleMatchCreatedEvent struct {
	*middletournament.MiddleTournamentMatchCreated
}

func (e *MiddleMatchCreatedEvent) One() [32]byte { return e.MiddleTournamentMatchCreated.One }
func (e *MiddleMatchCreatedEvent) Two() [32]byte { return e.MiddleTournamentMatchCreated.Two }
func (e *MiddleMatchCreatedEvent) LeftOfTwo() [32]byte {
	return e.MiddleTournamentMatchCreated.LeftOfTwo
}
func (e *MiddleMatchCreatedEvent) BlockNumber() uint64 { return e.Raw.BlockNumber }
func (e *MiddleMatchCreatedEvent) TxHash() common.Hash { return e.Raw.TxHash }

type MiddleMatchAdvancedEvent struct {
	*middletournament.MiddleTournamentMatchAdvanced
}

func (e *MiddleMatchAdvancedEvent) Arg0() [32]byte      { return e.MiddleTournamentMatchAdvanced.Arg0 }
func (e *MiddleMatchAdvancedEvent) Parent() [32]byte    { return e.MiddleTournamentMatchAdvanced.Parent }
func (e *MiddleMatchAdvancedEvent) Left() [32]byte      { return e.MiddleTournamentMatchAdvanced.Left }
func (e *MiddleMatchAdvancedEvent) BlockNumber() uint64 { return e.Raw.BlockNumber }
func (e *MiddleMatchAdvancedEvent) TxHash() common.Hash { return e.Raw.TxHash }

type MiddleMatchDeletedEvent struct {
	*middletournament.MiddleTournamentMatchDeleted
}

func (e *MiddleMatchDeletedEvent) Arg0() [32]byte { return e.MiddleTournamentMatchDeleted.Arg0 }

type MiddleNewInnerTournamentEvent struct {
	*middletournament.MiddleTournamentNewInnerTournament
}

func (e *MiddleNewInnerTournamentEvent) Arg0() [32]byte {
	return e.MiddleTournamentNewInnerTournament.Arg0
}

func (e *MiddleNewInnerTournamentEvent) Arg1() common.Address {
	return e.MiddleTournamentNewInnerTournament.Arg1
}
