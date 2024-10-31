package interfaces

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type ERC20Token interface {
	Approve(auth *bind.TransactOpts, spender common.Address, amount *big.Int) (*types.Transaction, error)
	Symbol(opts *bind.CallOpts) (string, error)
	BalanceOf(opts *bind.CallOpts, account common.Address) (*big.Int, error)
}
