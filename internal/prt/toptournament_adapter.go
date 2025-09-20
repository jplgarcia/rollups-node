// (c) Cartesi and individual authors (see AUTHORS)
// SPDX-License-Identifier: Apache-2.0 (see LICENSE)

package prt

import (
	"math/big"

	. "github.com/cartesi/rollups-node/internal/model"
	"github.com/cartesi/rollups-node/pkg/contracts/toptournament"
	"github.com/cartesi/rollups-node/pkg/ethutil"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// TopTournament Wrapper
type TopTournamentAdapterImpl struct {
	tournament        *toptournament.TopTournament
	client            *ethclient.Client
	tournamentAddress common.Address
	filter            ethutil.Filter
}

func NewTopTournamentAdapter(
	tournamentAddress common.Address,
	client *ethclient.Client,
	filter ethutil.Filter,
) (TournamentAdapter, error) {
	tournamentContract, err := toptournament.NewTopTournament(tournamentAddress, client)
	if err != nil {
		return nil, err
	}
	return &TopTournamentAdapterImpl{
		tournament:        tournamentContract,
		tournamentAddress: tournamentAddress,
		client:            client,
		filter:            filter,
	}, nil
}

func (a *TopTournamentAdapterImpl) Result(opts *bind.CallOpts) (bool, [32]byte, error) {
	finished, commitment, _, error := a.tournament.ArbitrationResult(opts)
	return finished, commitment, error
}

func (a *TopTournamentAdapterImpl) Constants(opts *bind.CallOpts) (TournamentConstants, error) {
	c, error := a.tournament.TournamentLevelConstants(opts)
	return TournamentConstants{
		MaxLevel: c.MaxLevel,
		Level:    c.Level,
		Log2step: c.Log2step,
		Height:   c.Height,
	}, error
}

func (a *TopTournamentAdapterImpl) TimeFinished(opts *bind.CallOpts) (bool, uint64, error) {
	return a.tournament.TimeFinished(opts)
}

func buildCommitmentJoinedFilterQuery(
	opts *bind.FilterOpts,
	tournamentAddress common.Address,
) (q ethereum.FilterQuery, err error) {
	c, err := toptournament.TopTournamentMetaData.GetAbi()
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

func (a *TopTournamentAdapterImpl) RetrieveCommitmentJoinedEvents(
	opts *bind.FilterOpts,
) ([]CommitmentJoinedEvent, error) {
	q, err := buildCommitmentJoinedFilterQuery(opts, a.tournamentAddress)
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
		events = append(events, &TopCommitmentJoinedEvent{ev})
	}
	return events, nil
}

func buildMatchAdvancedFilterQuery(
	opts *bind.FilterOpts,
	tournamentAddress common.Address,
) (q ethereum.FilterQuery, err error) {
	c, err := toptournament.TopTournamentMetaData.GetAbi()
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

func (a *TopTournamentAdapterImpl) RetrieveMatchAdvancedEvents(
	opts *bind.FilterOpts,
) ([]MatchAdvancedEvent, error) {
	q, err := buildMatchAdvancedFilterQuery(opts, a.tournamentAddress)
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
		events = append(events, &TopMatchAdvancedEvent{ev})
	}
	return events, nil
}

func buildMatchCreatedFilterQuery(
	opts *bind.FilterOpts,
	tournamentAddress common.Address,
) (q ethereum.FilterQuery, err error) {
	c, err := toptournament.TopTournamentMetaData.GetAbi()
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

func (a *TopTournamentAdapterImpl) RetrieveMatchCreatedEvents(
	opts *bind.FilterOpts,
) ([]MatchCreatedEvent, error) {
	q, err := buildMatchCreatedFilterQuery(opts, a.tournamentAddress)
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
		events = append(events, &TopMatchCreatedEvent{ev})
	}
	return events, nil
}

func buildMatchDeletedFilterQuery(
	opts *bind.FilterOpts,
	tournamentAddress common.Address,
) (q ethereum.FilterQuery, err error) {
	c, err := toptournament.TopTournamentMetaData.GetAbi()
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

func (a *TopTournamentAdapterImpl) RetrieveMatchDeletedEvents(
	opts *bind.FilterOpts,
) ([]MatchDeletedEvent, error) {
	q, err := buildMatchDeletedFilterQuery(opts, a.tournamentAddress)
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
		events = append(events, &TopMatchDeletedEvent{ev})
	}
	return events, nil
}

func buildNewInnerTournamentFilterQuery(
	opts *bind.FilterOpts,
	tournamentAddress common.Address,
) (q ethereum.FilterQuery, err error) {
	c, err := toptournament.TopTournamentMetaData.GetAbi()
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

func (a *TopTournamentAdapterImpl) RetrieveNewInnerTournamentEvents(
	opts *bind.FilterOpts,
) ([]NewInnerTournamentEvent, error) {
	q, err := buildNewInnerTournamentFilterQuery(opts, a.tournamentAddress)
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
		events = append(events, &TopNewInnerTournamentEvent{ev})
	}
	return events, nil
}

func buildAllEventsFilterQuery(
	opts *bind.FilterOpts,
	tournamentAddress common.Address,
) (q ethereum.FilterQuery, err error) {
	c, err := toptournament.TopTournamentMetaData.GetAbi()
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

func (a *TopTournamentAdapterImpl) RetrieveAllEvents(
	opts *bind.FilterOpts,
) (*TournamentEvents, error) {
	q, err := buildAllEventsFilterQuery(opts, a.tournamentAddress)
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

	c, err := toptournament.TopTournamentMetaData.GetAbi()
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
			commitmentJoined = append(commitmentJoined, &TopCommitmentJoinedEvent{ev})
		case c.Events[MonitoredEvent_MatchAdvanced.String()].ID:
			ev, err := a.tournament.ParseMatchAdvanced(*log)
			if err != nil {
				return nil, err
			}
			matchAdvanced = append(matchAdvanced, &TopMatchAdvancedEvent{ev})
		case c.Events[MonitoredEvent_MatchCreated.String()].ID:
			ev, err := a.tournament.ParseMatchCreated(*log)
			if err != nil {
				return nil, err
			}
			matchCreated = append(matchCreated, &TopMatchCreatedEvent{ev})
		case c.Events[MonitoredEvent_MatchDeleted.String()].ID:
			ev, err := a.tournament.ParseMatchDeleted(*log)
			if err != nil {
				return nil, err
			}
			matchDeleted = append(matchDeleted, &TopMatchDeletedEvent{ev})
		case c.Events[MonitoredEvent_NewInnerTournament.String()].ID:
			ev, err := a.tournament.ParseNewInnerTournament(*log)
			if err != nil {
				return nil, err
			}
			newInnerTournament = append(newInnerTournament, &TopNewInnerTournamentEvent{ev})
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

// Wrappers for TopTournament events
type TopCommitmentJoinedEvent struct {
	*toptournament.TopTournamentCommitmentJoined
}

func (e *TopCommitmentJoinedEvent) Root() [32]byte      { return e.TopTournamentCommitmentJoined.Root }
func (e *TopCommitmentJoinedEvent) BlockNumber() uint64 { return e.Raw.BlockNumber }
func (e *TopCommitmentJoinedEvent) TxHash() common.Hash { return e.Raw.TxHash }

type TopMatchCreatedEvent struct {
	*toptournament.TopTournamentMatchCreated
}

func (e *TopMatchCreatedEvent) One() [32]byte       { return e.TopTournamentMatchCreated.One }
func (e *TopMatchCreatedEvent) Two() [32]byte       { return e.TopTournamentMatchCreated.Two }
func (e *TopMatchCreatedEvent) LeftOfTwo() [32]byte { return e.TopTournamentMatchCreated.LeftOfTwo }
func (e *TopMatchCreatedEvent) BlockNumber() uint64 { return e.Raw.BlockNumber }
func (e *TopMatchCreatedEvent) TxHash() common.Hash { return e.Raw.TxHash }

type TopMatchAdvancedEvent struct {
	*toptournament.TopTournamentMatchAdvanced
}

func (e *TopMatchAdvancedEvent) Arg0() [32]byte      { return e.TopTournamentMatchAdvanced.Arg0 }
func (e *TopMatchAdvancedEvent) Parent() [32]byte    { return e.TopTournamentMatchAdvanced.Parent }
func (e *TopMatchAdvancedEvent) Left() [32]byte      { return e.TopTournamentMatchAdvanced.Left }
func (e *TopMatchAdvancedEvent) BlockNumber() uint64 { return e.Raw.BlockNumber }
func (e *TopMatchAdvancedEvent) TxHash() common.Hash { return e.Raw.TxHash }

type TopMatchDeletedEvent struct {
	*toptournament.TopTournamentMatchDeleted
}

func (e *TopMatchDeletedEvent) Arg0() [32]byte { return e.TopTournamentMatchDeleted.Arg0 }

type TopNewInnerTournamentEvent struct {
	*toptournament.TopTournamentNewInnerTournament
}

func (e *TopNewInnerTournamentEvent) Arg0() [32]byte { return e.TopTournamentNewInnerTournament.Arg0 }
func (e *TopNewInnerTournamentEvent) Arg1() common.Address {
	return e.TopTournamentNewInnerTournament.Arg1
}
