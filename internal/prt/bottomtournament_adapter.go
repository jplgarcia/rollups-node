// (c) Cartesi and individual authors (see AUTHORS)
// SPDX-License-Identifier: Apache-2.0 (see LICENSE)

package prt

import (
	"math/big"

	. "github.com/cartesi/rollups-node/internal/model"
	"github.com/cartesi/rollups-node/pkg/contracts/bottomtournament"
	"github.com/cartesi/rollups-node/pkg/ethutil"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// BottomTournament Wrapper
type BottomTournamentAdapterImpl struct {
	tournament        *bottomtournament.BottomTournament
	client            *ethclient.Client
	tournamentAddress common.Address
	filter            ethutil.Filter
}

func NewBottomTournamentAdapter(
	tournamentAddress common.Address,
	client *ethclient.Client,
	filter ethutil.Filter,
) (TournamentAdapter, error) {
	tournamentContract, err := bottomtournament.NewBottomTournament(tournamentAddress, client)
	if err != nil {
		return nil, err
	}
	return &BottomTournamentAdapterImpl{
		tournament:        tournamentContract,
		tournamentAddress: tournamentAddress,
		client:            client,
		filter:            filter,
	}, nil
}

func (a *BottomTournamentAdapterImpl) Result(opts *bind.CallOpts) (bool, [32]byte, error) {
	finished, _, commitment, _, error := a.tournament.InnerTournamentWinner(opts)
	return finished, commitment, error
}

func (a *BottomTournamentAdapterImpl) Constants(opts *bind.CallOpts) (TournamentConstants, error) {
	c, error := a.tournament.TournamentLevelConstants(opts)
	return TournamentConstants{
		MaxLevel: c.MaxLevel,
		Level:    c.Level,
		Log2step: c.Log2step,
		Height:   c.Height,
	}, error
}

func (a *BottomTournamentAdapterImpl) TimeFinished(opts *bind.CallOpts) (bool, uint64, error) {
	finished, block, error := a.tournament.TimeFinished(opts)
	return finished, block, error
}

func buildBottomCommitmentJoinedFilterQuery(
	opts *bind.FilterOpts,
	tournamentAddress common.Address,
) (q ethereum.FilterQuery, err error) {
	c, err := bottomtournament.BottomTournamentMetaData.GetAbi()
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

func (a *BottomTournamentAdapterImpl) RetrieveCommitmentJoinedEvents(
	opts *bind.FilterOpts,
) ([]CommitmentJoinedEvent, error) {
	q, err := buildBottomCommitmentJoinedFilterQuery(opts, a.tournamentAddress)
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
		events = append(events, &BottomCommitmentJoinedEvent{ev})
	}
	return events, nil
}

func buildBottomMatchAdvancedFilterQuery(
	opts *bind.FilterOpts,
	tournamentAddress common.Address,
) (q ethereum.FilterQuery, err error) {
	c, err := bottomtournament.BottomTournamentMetaData.GetAbi()
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

func (a *BottomTournamentAdapterImpl) RetrieveMatchAdvancedEvents(
	opts *bind.FilterOpts,
) ([]MatchAdvancedEvent, error) {
	q, err := buildBottomMatchAdvancedFilterQuery(opts, a.tournamentAddress)
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
		events = append(events, &BottomMatchAdvancedEvent{ev})
	}
	return events, nil
}

func buildBottomMatchCreatedFilterQuery(
	opts *bind.FilterOpts,
	tournamentAddress common.Address,
) (q ethereum.FilterQuery, err error) {
	c, err := bottomtournament.BottomTournamentMetaData.GetAbi()
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

func (a *BottomTournamentAdapterImpl) RetrieveMatchCreatedEvents(
	opts *bind.FilterOpts,
) ([]MatchCreatedEvent, error) {
	q, err := buildBottomMatchCreatedFilterQuery(opts, a.tournamentAddress)
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
		events = append(events, &BottomMatchCreatedEvent{ev})
	}
	return events, nil
}

func buildBottomMatchDeletedFilterQuery(
	opts *bind.FilterOpts,
	tournamentAddress common.Address,
) (q ethereum.FilterQuery, err error) {
	c, err := bottomtournament.BottomTournamentMetaData.GetAbi()
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

func (a *BottomTournamentAdapterImpl) RetrieveMatchDeletedEvents(
	opts *bind.FilterOpts,
) ([]MatchDeletedEvent, error) {
	q, err := buildBottomMatchDeletedFilterQuery(opts, a.tournamentAddress)
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
		events = append(events, &BottomMatchDeletedEvent{ev})
	}
	return events, nil
}

func (a *BottomTournamentAdapterImpl) RetrieveNewInnerTournamentEvents(
	_ *bind.FilterOpts,
) ([]NewInnerTournamentEvent, error) {
	// BottomTournament has no NewInnerTournament event
	return []NewInnerTournamentEvent{}, nil
}

func buildBottomAllEventsFilterQuery(
	opts *bind.FilterOpts,
	tournamentAddress common.Address,
) (q ethereum.FilterQuery, err error) {
	c, err := bottomtournament.BottomTournamentMetaData.GetAbi()
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

func (a *BottomTournamentAdapterImpl) RetrieveAllEvents(
	opts *bind.FilterOpts,
) (*TournamentEvents, error) {
	q, err := buildBottomAllEventsFilterQuery(opts, a.tournamentAddress)
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

	c, err := bottomtournament.BottomTournamentMetaData.GetAbi()
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
			commitmentJoined = append(commitmentJoined, &BottomCommitmentJoinedEvent{ev})
		case c.Events[MonitoredEvent_MatchAdvanced.String()].ID:
			ev, err := a.tournament.ParseMatchAdvanced(*log)
			if err != nil {
				return nil, err
			}
			matchAdvanced = append(matchAdvanced, &BottomMatchAdvancedEvent{ev})
		case c.Events[MonitoredEvent_MatchCreated.String()].ID:
			ev, err := a.tournament.ParseMatchCreated(*log)
			if err != nil {
				return nil, err
			}
			matchCreated = append(matchCreated, &BottomMatchCreatedEvent{ev})
		case c.Events[MonitoredEvent_MatchDeleted.String()].ID:
			ev, err := a.tournament.ParseMatchDeleted(*log)
			if err != nil {
				return nil, err
			}
			matchDeleted = append(matchDeleted, &BottomMatchDeletedEvent{ev})
		}
	}

	return &TournamentEvents{
		CommitmentJoined:   commitmentJoined,
		MatchAdvanced:      matchAdvanced,
		MatchCreated:       matchCreated,
		MatchDeleted:       matchDeleted,
		NewInnerTournament: []NewInnerTournamentEvent{}, // BottomTournament has no NewInnerTournament event
	}, nil
}

// Wrappers for BottomTournament events
type BottomCommitmentJoinedEvent struct {
	*bottomtournament.BottomTournamentCommitmentJoined
}

func (e *BottomCommitmentJoinedEvent) Root() [32]byte      { return e.BottomTournamentCommitmentJoined.Root }
func (e *BottomCommitmentJoinedEvent) BlockNumber() uint64 { return e.Raw.BlockNumber }
func (e *BottomCommitmentJoinedEvent) TxHash() common.Hash { return e.Raw.TxHash }

type BottomMatchCreatedEvent struct {
	*bottomtournament.BottomTournamentMatchCreated
}

func (e *BottomMatchCreatedEvent) One() [32]byte { return e.BottomTournamentMatchCreated.One }
func (e *BottomMatchCreatedEvent) Two() [32]byte { return e.BottomTournamentMatchCreated.Two }
func (e *BottomMatchCreatedEvent) LeftOfTwo() [32]byte {
	return e.BottomTournamentMatchCreated.LeftOfTwo
}
func (e *BottomMatchCreatedEvent) BlockNumber() uint64 { return e.Raw.BlockNumber }
func (e *BottomMatchCreatedEvent) TxHash() common.Hash { return e.Raw.TxHash }

type BottomMatchAdvancedEvent struct {
	*bottomtournament.BottomTournamentMatchAdvanced
}

func (e *BottomMatchAdvancedEvent) Arg0() [32]byte      { return e.BottomTournamentMatchAdvanced.Arg0 }
func (e *BottomMatchAdvancedEvent) Parent() [32]byte    { return e.BottomTournamentMatchAdvanced.Parent }
func (e *BottomMatchAdvancedEvent) Left() [32]byte      { return e.BottomTournamentMatchAdvanced.Left }
func (e *BottomMatchAdvancedEvent) BlockNumber() uint64 { return e.Raw.BlockNumber }
func (e *BottomMatchAdvancedEvent) TxHash() common.Hash { return e.Raw.TxHash }

type BottomMatchDeletedEvent struct {
	*bottomtournament.BottomTournamentMatchDeleted
}

func (e *BottomMatchDeletedEvent) Arg0() [32]byte { return e.BottomTournamentMatchDeleted.Arg0 }
