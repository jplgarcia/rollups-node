// (c) Cartesi and individual authors (see AUTHORS)
// SPDX-License-Identifier: Apache-2.0 (see LICENSE)

package claimer

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"

	"github.com/cartesi/rollups-node/internal/config"
	"github.com/cartesi/rollups-node/internal/model"
	"github.com/cartesi/rollups-node/pkg/contracts/iconsensus"
	"github.com/cartesi/rollups-node/pkg/ethutil"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type iclaimerBlockchain interface {
	findClaimSubmittedEventAndSucc(
		ctx context.Context,
		application *model.Application,
		epoch *model.Epoch,
		endBlock *big.Int,
	) (
		*iconsensus.IConsensus,
		*iconsensus.IConsensusClaimSubmitted,
		*iconsensus.IConsensusClaimSubmitted,
		error,
	)

	submitClaimToBlockchain(
		ic *iconsensus.IConsensus,
		application *model.Application,
		epoch *model.Epoch,
	) (common.Hash, error)

	pollTransaction(
		ctx context.Context,
		txHash common.Hash,
		endBlock *big.Int,
	) (bool, *types.Receipt, error)

	findClaimAcceptedEventAndSucc(
		ctx context.Context,
		application *model.Application,
		epoch *model.Epoch,
		endBlock *big.Int,
	) (
		*iconsensus.IConsensus,
		*iconsensus.IConsensusClaimAccepted,
		*iconsensus.IConsensusClaimAccepted,
		error,
	)

	getBlockNumber(ctx context.Context) (*big.Int, error)

	getConsensusAddress(
		ctx context.Context,
		app *model.Application,
	) (common.Address, error)
}

type claimerBlockchain struct {
	client       *ethclient.Client
	txOpts       *bind.TransactOpts
	logger       *slog.Logger
	filter       ethutil.Filter
	defaultBlock config.DefaultBlock
}

func (self *claimerBlockchain) submitClaimToBlockchain(
	ic *iconsensus.IConsensus,
	application *model.Application,
	epoch *model.Epoch,
) (common.Hash, error) {
	txHash := common.Hash{}
	lastBlockNumber := new(big.Int).SetUint64(epoch.LastBlock)
	tx, err := ic.SubmitClaim(self.txOpts, application.IApplicationAddress,
		lastBlockNumber, *epoch.OutputsMerkleRoot)
	if err != nil {
		self.logger.Error("submitClaimToBlockchain:failed",
			"appContractAddress", application.IApplicationAddress,
			"claimHash", *epoch.OutputsMerkleRoot,
			"last_block", epoch.LastBlock,
			"error", err)
	} else {
		txHash = tx.Hash()
		self.logger.Debug("submitClaimToBlockchain:success",
			"appContractAddress", application.IApplicationAddress,
			"claimHash", *epoch.OutputsMerkleRoot,
			"last_block", epoch.LastBlock,
			"TxHash", txHash)
	}
	return txHash, err
}

// scan the event stream for a claimSubmitted event that matches claim.
// return this event and its successor
func (self *claimerBlockchain) findClaimSubmittedEventAndSucc(
	ctx context.Context,
	application *model.Application,
	epoch *model.Epoch,
	endBlock *big.Int,
) (
	*iconsensus.IConsensus,
	*iconsensus.IConsensusClaimSubmitted,
	*iconsensus.IConsensusClaimSubmitted,
	error,
) {
	ic, err := iconsensus.NewIConsensus(application.IConsensusAddress, self.client)
	if err != nil {
		return nil, nil, nil, err
	}

	oracle := func(ctx context.Context, block uint64) (*big.Int, error) {
		callOpts := &bind.CallOpts{
			Context:     ctx,
			BlockNumber: new(big.Int).SetUint64(block),
		}
		numSubmittedClaims, err := ic.GetNumberOfSubmittedClaims(callOpts)

		if err != nil {
			return nil, fmt.Errorf("failed to get number of submitted claims at block %d: %w", block, err)
		}
		return numSubmittedClaims, nil
	}

	events := []*iconsensus.IConsensusClaimSubmitted{}
	onHit := func(block uint64) error {
		filterOpts := &bind.FilterOpts{
			Context: ctx,
			Start:   block,
			End:     &block,
		}
		claimSubmittedEvents, err := ic.FilterClaimSubmitted(filterOpts, nil, []common.Address{application.IApplicationAddress})
		if err != nil {
			return fmt.Errorf("failed to retrieve ClaimSubmitted events at block %d: %w", block, err)
		}
		defer claimSubmittedEvents.Close()
		for claimSubmittedEvents.Next() {
			event := claimSubmittedEvents.Event
			if (len(events) == 0) && claimSubmittedEventMatches(application, epoch, event) {
				events = append(events, event)
			} else if len(events) != 0 {
				events = append(events, event)
			}
		}
		err = claimSubmittedEvents.Error()
		if err != nil {
			return fmt.Errorf("failed to iterate ClaimSubmitted events: %w", err)
		}
		return nil
	}

	numSubmittedClaims, err := oracle(ctx, epoch.LastBlock)
	if err != nil {
		return nil, nil, nil, err
	}
	_, err = ethutil.FindTransitions(ctx, epoch.LastBlock, endBlock.Uint64(), numSubmittedClaims, oracle, onHit)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to walk ClaimSubmitted transitions: %w", err)
	}

	if len(events) == 0 {
		return ic, nil, nil, nil
	} else if len(events) == 1 {
		return ic, events[0], nil, nil
	} else {
		return ic, events[0], events[1], nil
	}
}

// scan the event stream for a claimAccepted event that matches claim.
// return this event and its successor
func (self *claimerBlockchain) findClaimAcceptedEventAndSucc(
	ctx context.Context,
	application *model.Application,
	epoch *model.Epoch,
	endBlock *big.Int,
) (
	*iconsensus.IConsensus,
	*iconsensus.IConsensusClaimAccepted,
	*iconsensus.IConsensusClaimAccepted,
	error,
) {
	ic, err := iconsensus.NewIConsensus(application.IConsensusAddress, self.client)
	if err != nil {
		return nil, nil, nil, err
	}

	oracle := func(ctx context.Context, block uint64) (*big.Int, error) {
		callOpts := &bind.CallOpts{
			Context:     ctx,
			BlockNumber: new(big.Int).SetUint64(block),
		}
		numAcceptedClaims, err := ic.GetNumberOfAcceptedClaims(callOpts)

		if err != nil {
			return nil, fmt.Errorf("failed to get number of accepted claims at block %d: %w", block, err)
		}
		return numAcceptedClaims, nil
	}

	events := []*iconsensus.IConsensusClaimAccepted{}
	onHit := func(block uint64) error {
		filterOpts := &bind.FilterOpts{
			Context: ctx,
			Start:   block,
			End:     &block,
		}
		claimAcceptedEvents, err := ic.FilterClaimAccepted(filterOpts, []common.Address{application.IApplicationAddress})
		if err != nil {
			return fmt.Errorf("failed to retrieve claimAccepted events at block %d: %w", block, err)
		}
		defer claimAcceptedEvents.Close()
		for claimAcceptedEvents.Next() {
			event := claimAcceptedEvents.Event
			if (len(events) == 0) && claimAcceptedEventMatches(application, epoch, event) {
				events = append(events, claimAcceptedEvents.Event)
			} else if len(events) != 0 {
				events = append(events, claimAcceptedEvents.Event)
			}
		}
		err = claimAcceptedEvents.Error()
		if err != nil {
			return fmt.Errorf("failed to iterate submitted claim events: %w", err)
		}
		return nil
	}

	numAcceptedClaims, err := oracle(ctx, epoch.LastBlock)
	if err != nil {
		return nil, nil, nil, err
	}
	_, err = ethutil.FindTransitions(ctx, epoch.LastBlock, endBlock.Uint64(), numAcceptedClaims, oracle, onHit)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to walk ClaimAccepted transitions: %w", err)
	}

	if len(events) == 0 {
		return ic, nil, nil, nil
	} else if len(events) == 1 {
		return ic, events[0], nil, nil
	} else {
		return ic, events[0], events[1], nil
	}
}

func (self *claimerBlockchain) getConsensusAddress(
	ctx context.Context,
	app *model.Application,
) (common.Address, error) {
	return ethutil.GetConsensus(ctx, self.client, app.IApplicationAddress)
}

/* poll a transaction hash for its submission status and receipt */
func (self *claimerBlockchain) pollTransaction(
	ctx context.Context,
	txHash common.Hash,
	endBlock *big.Int,
) (bool, *types.Receipt, error) {
	_, isPending, err := self.client.TransactionByHash(ctx, txHash)
	if err != nil || isPending {
		return false, nil, err
	}

	receipt, err := self.client.TransactionReceipt(ctx, txHash)
	if err != nil {
		return false, nil, err
	}

	if receipt.BlockNumber.Cmp(endBlock) >= 0 {
		return false, receipt, err
	}

	return receipt.Status == 1, receipt, err
}

/* Retrieve the block number of "DefaultBlock" */
func (self *claimerBlockchain) getBlockNumber(ctx context.Context) (*big.Int, error) {
	var nr int64
	switch self.defaultBlock {
	case model.DefaultBlock_Pending:
		nr = rpc.PendingBlockNumber.Int64()
	case model.DefaultBlock_Latest:
		nr = rpc.LatestBlockNumber.Int64()
	case model.DefaultBlock_Finalized:
		nr = rpc.FinalizedBlockNumber.Int64()
	case model.DefaultBlock_Safe:
		nr = rpc.SafeBlockNumber.Int64()
	default:
		return nil, fmt.Errorf("default block '%v' not supported", self.defaultBlock)
	}

	hdr, err := self.client.HeaderByNumber(ctx, big.NewInt(nr))
	if err != nil {
		return nil, err
	}
	return hdr.Number, nil
}
