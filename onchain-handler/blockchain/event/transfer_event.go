package event

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type TransferEvent struct {
	From  common.Address
	To    common.Address
	Value *big.Int
}
