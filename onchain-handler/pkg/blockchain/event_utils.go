package blockchain

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/pkg/logger"
)

// TransferEvent represents the structure of a transfer event.
type TransferEvent struct {
	From  common.Address
	To    common.Address
	Value *big.Int
}

func UnpackTransferEvent(vLog types.Log, parsedABI abi.ABI) (TransferEvent, error) {
	var transferEvent TransferEvent

	// Ensure the number of topics matches the expected event (Transfer has 3 topics: event signature, from, to)
	if len(vLog.Topics) != 3 {
		return transferEvent, fmt.Errorf("invalid number of topics in log")
	}

	// The first topic is the event signature, so we skip it.
	// The second topic is the "from" address, and the third is the "to" address.
	transferEvent.From = common.HexToAddress(vLog.Topics[1].Hex())
	transferEvent.To = common.HexToAddress(vLog.Topics[2].Hex())

	// Unpack the value (the non-indexed parameter) from the data field
	err := parsedABI.UnpackIntoInterface(&transferEvent, constants.TransferEventName, vLog.Data)
	if err != nil {
		logger.GetLogger().Errorf("Failed to unpack transfer event: %v", err)
		return transferEvent, err
	}

	return transferEvent, nil
}
