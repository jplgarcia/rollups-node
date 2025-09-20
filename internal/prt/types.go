// (c) Cartesi and individual authors (see AUTHORS)
// SPDX-License-Identifier: Apache-2.0 (see LICENSE)

package prt

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type BlockchainEvent interface {
	Raw() types.Log // Blockchain specific contextual infos
}

// Interfaces for tournament events
type CommitmentJoinedEvent interface {
	Root() [32]byte
	BlockNumber() uint64
	TxHash() common.Hash
}

type MatchCreatedEvent interface {
	One() [32]byte
	Two() [32]byte
	LeftOfTwo() [32]byte
	BlockNumber() uint64
	TxHash() common.Hash
}

type MatchAdvancedEvent interface {
	Arg0() [32]byte
	Parent() [32]byte
	Left() [32]byte
	BlockNumber() uint64
	TxHash() common.Hash
}

type MatchDeletedEvent interface {
	Arg0() [32]byte
}

type NewInnerTournamentEvent interface {
	Arg0() [32]byte
	Arg1() common.Address
}

type TournamentConstants struct {
	MaxLevel uint64
	Level    uint64
	Log2step uint64
	Height   uint64
}

// Interface for Tournament reading
type TournamentAdapter interface {
	RetrieveCommitmentJoinedEvents(opts *bind.FilterOpts) ([]CommitmentJoinedEvent, error)
	RetrieveMatchAdvancedEvents(opts *bind.FilterOpts) ([]MatchAdvancedEvent, error)
	RetrieveMatchCreatedEvents(opts *bind.FilterOpts) ([]MatchCreatedEvent, error)
	RetrieveMatchDeletedEvents(opts *bind.FilterOpts) ([]MatchDeletedEvent, error)
	RetrieveNewInnerTournamentEvents(opts *bind.FilterOpts) ([]NewInnerTournamentEvent, error)
	RetrieveAllEvents(opts *bind.FilterOpts) (*TournamentEvents, error)
	Result(opts *bind.CallOpts) (bool, [32]byte, error)
	Constants(opts *bind.CallOpts) (TournamentConstants, error)
	TimeFinished(opts *bind.CallOpts) (bool, uint64, error)
}

// Struct to hold all events retrieved at once
type TournamentEvents struct {
	CommitmentJoined   []CommitmentJoinedEvent
	MatchAdvanced      []MatchAdvancedEvent
	MatchCreated       []MatchCreatedEvent
	MatchDeleted       []MatchDeletedEvent
	NewInnerTournament []NewInnerTournamentEvent
}
