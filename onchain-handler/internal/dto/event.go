package dto

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type TransferEventDTO struct {
	From  common.Address
	To    common.Address
	Value *big.Int
}
