// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package lifepointstaking

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// LifepointstakingMetaData contains all meta data concerning the Lifepointstaking contract.
var LifepointstakingMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"lockDuration\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"Staked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"Withdrawn\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"tokenAddress\",\"type\":\"address\"}],\"name\":\"addLifePointVersion\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lockDuration\",\"type\":\"uint256\"}],\"name\":\"backendStakeLifePoint\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"}],\"name\":\"backendWithdraw\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"lifePointVersions\",\"outputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lockDuration\",\"type\":\"uint256\"}],\"name\":\"stakeLifePoint\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"stakes\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"startTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lockDuration\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"withdrawn\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"}],\"name\":\"withdrawLifePoint\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// LifepointstakingABI is the input ABI used to generate the binding from.
// Deprecated: Use LifepointstakingMetaData.ABI instead.
var LifepointstakingABI = LifepointstakingMetaData.ABI

// Lifepointstaking is an auto generated Go binding around an Ethereum contract.
type Lifepointstaking struct {
	LifepointstakingCaller     // Read-only binding to the contract
	LifepointstakingTransactor // Write-only binding to the contract
	LifepointstakingFilterer   // Log filterer for contract events
}

// LifepointstakingCaller is an auto generated read-only Go binding around an Ethereum contract.
type LifepointstakingCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// LifepointstakingTransactor is an auto generated write-only Go binding around an Ethereum contract.
type LifepointstakingTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// LifepointstakingFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type LifepointstakingFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// LifepointstakingSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type LifepointstakingSession struct {
	Contract     *Lifepointstaking // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// LifepointstakingCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type LifepointstakingCallerSession struct {
	Contract *LifepointstakingCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts           // Call options to use throughout this session
}

// LifepointstakingTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type LifepointstakingTransactorSession struct {
	Contract     *LifepointstakingTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts           // Transaction auth options to use throughout this session
}

// LifepointstakingRaw is an auto generated low-level Go binding around an Ethereum contract.
type LifepointstakingRaw struct {
	Contract *Lifepointstaking // Generic contract binding to access the raw methods on
}

// LifepointstakingCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type LifepointstakingCallerRaw struct {
	Contract *LifepointstakingCaller // Generic read-only contract binding to access the raw methods on
}

// LifepointstakingTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type LifepointstakingTransactorRaw struct {
	Contract *LifepointstakingTransactor // Generic write-only contract binding to access the raw methods on
}

// NewLifepointstaking creates a new instance of Lifepointstaking, bound to a specific deployed contract.
func NewLifepointstaking(address common.Address, backend bind.ContractBackend) (*Lifepointstaking, error) {
	contract, err := bindLifepointstaking(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Lifepointstaking{LifepointstakingCaller: LifepointstakingCaller{contract: contract}, LifepointstakingTransactor: LifepointstakingTransactor{contract: contract}, LifepointstakingFilterer: LifepointstakingFilterer{contract: contract}}, nil
}

// NewLifepointstakingCaller creates a new read-only instance of Lifepointstaking, bound to a specific deployed contract.
func NewLifepointstakingCaller(address common.Address, caller bind.ContractCaller) (*LifepointstakingCaller, error) {
	contract, err := bindLifepointstaking(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &LifepointstakingCaller{contract: contract}, nil
}

// NewLifepointstakingTransactor creates a new write-only instance of Lifepointstaking, bound to a specific deployed contract.
func NewLifepointstakingTransactor(address common.Address, transactor bind.ContractTransactor) (*LifepointstakingTransactor, error) {
	contract, err := bindLifepointstaking(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &LifepointstakingTransactor{contract: contract}, nil
}

// NewLifepointstakingFilterer creates a new log filterer instance of Lifepointstaking, bound to a specific deployed contract.
func NewLifepointstakingFilterer(address common.Address, filterer bind.ContractFilterer) (*LifepointstakingFilterer, error) {
	contract, err := bindLifepointstaking(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &LifepointstakingFilterer{contract: contract}, nil
}

// bindLifepointstaking binds a generic wrapper to an already deployed contract.
func bindLifepointstaking(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := LifepointstakingMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Lifepointstaking *LifepointstakingRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Lifepointstaking.Contract.LifepointstakingCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Lifepointstaking *LifepointstakingRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Lifepointstaking.Contract.LifepointstakingTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Lifepointstaking *LifepointstakingRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Lifepointstaking.Contract.LifepointstakingTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Lifepointstaking *LifepointstakingCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Lifepointstaking.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Lifepointstaking *LifepointstakingTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Lifepointstaking.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Lifepointstaking *LifepointstakingTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Lifepointstaking.Contract.contract.Transact(opts, method, params...)
}

// LifePointVersions is a free data retrieval call binding the contract method 0xc8c28039.
//
// Solidity: function lifePointVersions(uint256 ) view returns(address)
func (_Lifepointstaking *LifepointstakingCaller) LifePointVersions(opts *bind.CallOpts, arg0 *big.Int) (common.Address, error) {
	var out []interface{}
	err := _Lifepointstaking.contract.Call(opts, &out, "lifePointVersions", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// LifePointVersions is a free data retrieval call binding the contract method 0xc8c28039.
//
// Solidity: function lifePointVersions(uint256 ) view returns(address)
func (_Lifepointstaking *LifepointstakingSession) LifePointVersions(arg0 *big.Int) (common.Address, error) {
	return _Lifepointstaking.Contract.LifePointVersions(&_Lifepointstaking.CallOpts, arg0)
}

// LifePointVersions is a free data retrieval call binding the contract method 0xc8c28039.
//
// Solidity: function lifePointVersions(uint256 ) view returns(address)
func (_Lifepointstaking *LifepointstakingCallerSession) LifePointVersions(arg0 *big.Int) (common.Address, error) {
	return _Lifepointstaking.Contract.LifePointVersions(&_Lifepointstaking.CallOpts, arg0)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Lifepointstaking *LifepointstakingCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Lifepointstaking.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Lifepointstaking *LifepointstakingSession) Owner() (common.Address, error) {
	return _Lifepointstaking.Contract.Owner(&_Lifepointstaking.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Lifepointstaking *LifepointstakingCallerSession) Owner() (common.Address, error) {
	return _Lifepointstaking.Contract.Owner(&_Lifepointstaking.CallOpts)
}

// Stakes is a free data retrieval call binding the contract method 0x584b62a1.
//
// Solidity: function stakes(address , uint256 ) view returns(uint256 amount, uint256 startTime, uint256 lockDuration, bool withdrawn)
func (_Lifepointstaking *LifepointstakingCaller) Stakes(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) (struct {
	Amount       *big.Int
	StartTime    *big.Int
	LockDuration *big.Int
	Withdrawn    bool
}, error) {
	var out []interface{}
	err := _Lifepointstaking.contract.Call(opts, &out, "stakes", arg0, arg1)

	outstruct := new(struct {
		Amount       *big.Int
		StartTime    *big.Int
		LockDuration *big.Int
		Withdrawn    bool
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Amount = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.StartTime = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.LockDuration = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.Withdrawn = *abi.ConvertType(out[3], new(bool)).(*bool)

	return *outstruct, err

}

// Stakes is a free data retrieval call binding the contract method 0x584b62a1.
//
// Solidity: function stakes(address , uint256 ) view returns(uint256 amount, uint256 startTime, uint256 lockDuration, bool withdrawn)
func (_Lifepointstaking *LifepointstakingSession) Stakes(arg0 common.Address, arg1 *big.Int) (struct {
	Amount       *big.Int
	StartTime    *big.Int
	LockDuration *big.Int
	Withdrawn    bool
}, error) {
	return _Lifepointstaking.Contract.Stakes(&_Lifepointstaking.CallOpts, arg0, arg1)
}

// Stakes is a free data retrieval call binding the contract method 0x584b62a1.
//
// Solidity: function stakes(address , uint256 ) view returns(uint256 amount, uint256 startTime, uint256 lockDuration, bool withdrawn)
func (_Lifepointstaking *LifepointstakingCallerSession) Stakes(arg0 common.Address, arg1 *big.Int) (struct {
	Amount       *big.Int
	StartTime    *big.Int
	LockDuration *big.Int
	Withdrawn    bool
}, error) {
	return _Lifepointstaking.Contract.Stakes(&_Lifepointstaking.CallOpts, arg0, arg1)
}

// AddLifePointVersion is a paid mutator transaction binding the contract method 0x893b91f3.
//
// Solidity: function addLifePointVersion(uint256 version, address tokenAddress) returns()
func (_Lifepointstaking *LifepointstakingTransactor) AddLifePointVersion(opts *bind.TransactOpts, version *big.Int, tokenAddress common.Address) (*types.Transaction, error) {
	return _Lifepointstaking.contract.Transact(opts, "addLifePointVersion", version, tokenAddress)
}

// AddLifePointVersion is a paid mutator transaction binding the contract method 0x893b91f3.
//
// Solidity: function addLifePointVersion(uint256 version, address tokenAddress) returns()
func (_Lifepointstaking *LifepointstakingSession) AddLifePointVersion(version *big.Int, tokenAddress common.Address) (*types.Transaction, error) {
	return _Lifepointstaking.Contract.AddLifePointVersion(&_Lifepointstaking.TransactOpts, version, tokenAddress)
}

// AddLifePointVersion is a paid mutator transaction binding the contract method 0x893b91f3.
//
// Solidity: function addLifePointVersion(uint256 version, address tokenAddress) returns()
func (_Lifepointstaking *LifepointstakingTransactorSession) AddLifePointVersion(version *big.Int, tokenAddress common.Address) (*types.Transaction, error) {
	return _Lifepointstaking.Contract.AddLifePointVersion(&_Lifepointstaking.TransactOpts, version, tokenAddress)
}

// BackendStakeLifePoint is a paid mutator transaction binding the contract method 0xea73f59a.
//
// Solidity: function backendStakeLifePoint(address user, uint256 version, uint256 amount, uint256 lockDuration) returns()
func (_Lifepointstaking *LifepointstakingTransactor) BackendStakeLifePoint(opts *bind.TransactOpts, user common.Address, version *big.Int, amount *big.Int, lockDuration *big.Int) (*types.Transaction, error) {
	return _Lifepointstaking.contract.Transact(opts, "backendStakeLifePoint", user, version, amount, lockDuration)
}

// BackendStakeLifePoint is a paid mutator transaction binding the contract method 0xea73f59a.
//
// Solidity: function backendStakeLifePoint(address user, uint256 version, uint256 amount, uint256 lockDuration) returns()
func (_Lifepointstaking *LifepointstakingSession) BackendStakeLifePoint(user common.Address, version *big.Int, amount *big.Int, lockDuration *big.Int) (*types.Transaction, error) {
	return _Lifepointstaking.Contract.BackendStakeLifePoint(&_Lifepointstaking.TransactOpts, user, version, amount, lockDuration)
}

// BackendStakeLifePoint is a paid mutator transaction binding the contract method 0xea73f59a.
//
// Solidity: function backendStakeLifePoint(address user, uint256 version, uint256 amount, uint256 lockDuration) returns()
func (_Lifepointstaking *LifepointstakingTransactorSession) BackendStakeLifePoint(user common.Address, version *big.Int, amount *big.Int, lockDuration *big.Int) (*types.Transaction, error) {
	return _Lifepointstaking.Contract.BackendStakeLifePoint(&_Lifepointstaking.TransactOpts, user, version, amount, lockDuration)
}

// BackendWithdraw is a paid mutator transaction binding the contract method 0x8b5caa66.
//
// Solidity: function backendWithdraw(address user, uint256 version) returns()
func (_Lifepointstaking *LifepointstakingTransactor) BackendWithdraw(opts *bind.TransactOpts, user common.Address, version *big.Int) (*types.Transaction, error) {
	return _Lifepointstaking.contract.Transact(opts, "backendWithdraw", user, version)
}

// BackendWithdraw is a paid mutator transaction binding the contract method 0x8b5caa66.
//
// Solidity: function backendWithdraw(address user, uint256 version) returns()
func (_Lifepointstaking *LifepointstakingSession) BackendWithdraw(user common.Address, version *big.Int) (*types.Transaction, error) {
	return _Lifepointstaking.Contract.BackendWithdraw(&_Lifepointstaking.TransactOpts, user, version)
}

// BackendWithdraw is a paid mutator transaction binding the contract method 0x8b5caa66.
//
// Solidity: function backendWithdraw(address user, uint256 version) returns()
func (_Lifepointstaking *LifepointstakingTransactorSession) BackendWithdraw(user common.Address, version *big.Int) (*types.Transaction, error) {
	return _Lifepointstaking.Contract.BackendWithdraw(&_Lifepointstaking.TransactOpts, user, version)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Lifepointstaking *LifepointstakingTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Lifepointstaking.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Lifepointstaking *LifepointstakingSession) RenounceOwnership() (*types.Transaction, error) {
	return _Lifepointstaking.Contract.RenounceOwnership(&_Lifepointstaking.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Lifepointstaking *LifepointstakingTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Lifepointstaking.Contract.RenounceOwnership(&_Lifepointstaking.TransactOpts)
}

// StakeLifePoint is a paid mutator transaction binding the contract method 0xffdcf240.
//
// Solidity: function stakeLifePoint(uint256 version, uint256 amount, uint256 lockDuration) returns()
func (_Lifepointstaking *LifepointstakingTransactor) StakeLifePoint(opts *bind.TransactOpts, version *big.Int, amount *big.Int, lockDuration *big.Int) (*types.Transaction, error) {
	return _Lifepointstaking.contract.Transact(opts, "stakeLifePoint", version, amount, lockDuration)
}

// StakeLifePoint is a paid mutator transaction binding the contract method 0xffdcf240.
//
// Solidity: function stakeLifePoint(uint256 version, uint256 amount, uint256 lockDuration) returns()
func (_Lifepointstaking *LifepointstakingSession) StakeLifePoint(version *big.Int, amount *big.Int, lockDuration *big.Int) (*types.Transaction, error) {
	return _Lifepointstaking.Contract.StakeLifePoint(&_Lifepointstaking.TransactOpts, version, amount, lockDuration)
}

// StakeLifePoint is a paid mutator transaction binding the contract method 0xffdcf240.
//
// Solidity: function stakeLifePoint(uint256 version, uint256 amount, uint256 lockDuration) returns()
func (_Lifepointstaking *LifepointstakingTransactorSession) StakeLifePoint(version *big.Int, amount *big.Int, lockDuration *big.Int) (*types.Transaction, error) {
	return _Lifepointstaking.Contract.StakeLifePoint(&_Lifepointstaking.TransactOpts, version, amount, lockDuration)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Lifepointstaking *LifepointstakingTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Lifepointstaking.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Lifepointstaking *LifepointstakingSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Lifepointstaking.Contract.TransferOwnership(&_Lifepointstaking.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Lifepointstaking *LifepointstakingTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Lifepointstaking.Contract.TransferOwnership(&_Lifepointstaking.TransactOpts, newOwner)
}

// WithdrawLifePoint is a paid mutator transaction binding the contract method 0x6eb98aa8.
//
// Solidity: function withdrawLifePoint(uint256 version) returns()
func (_Lifepointstaking *LifepointstakingTransactor) WithdrawLifePoint(opts *bind.TransactOpts, version *big.Int) (*types.Transaction, error) {
	return _Lifepointstaking.contract.Transact(opts, "withdrawLifePoint", version)
}

// WithdrawLifePoint is a paid mutator transaction binding the contract method 0x6eb98aa8.
//
// Solidity: function withdrawLifePoint(uint256 version) returns()
func (_Lifepointstaking *LifepointstakingSession) WithdrawLifePoint(version *big.Int) (*types.Transaction, error) {
	return _Lifepointstaking.Contract.WithdrawLifePoint(&_Lifepointstaking.TransactOpts, version)
}

// WithdrawLifePoint is a paid mutator transaction binding the contract method 0x6eb98aa8.
//
// Solidity: function withdrawLifePoint(uint256 version) returns()
func (_Lifepointstaking *LifepointstakingTransactorSession) WithdrawLifePoint(version *big.Int) (*types.Transaction, error) {
	return _Lifepointstaking.Contract.WithdrawLifePoint(&_Lifepointstaking.TransactOpts, version)
}

// LifepointstakingOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Lifepointstaking contract.
type LifepointstakingOwnershipTransferredIterator struct {
	Event *LifepointstakingOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *LifepointstakingOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(LifepointstakingOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(LifepointstakingOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *LifepointstakingOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *LifepointstakingOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// LifepointstakingOwnershipTransferred represents a OwnershipTransferred event raised by the Lifepointstaking contract.
type LifepointstakingOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Lifepointstaking *LifepointstakingFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*LifepointstakingOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Lifepointstaking.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &LifepointstakingOwnershipTransferredIterator{contract: _Lifepointstaking.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Lifepointstaking *LifepointstakingFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *LifepointstakingOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Lifepointstaking.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(LifepointstakingOwnershipTransferred)
				if err := _Lifepointstaking.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Lifepointstaking *LifepointstakingFilterer) ParseOwnershipTransferred(log types.Log) (*LifepointstakingOwnershipTransferred, error) {
	event := new(LifepointstakingOwnershipTransferred)
	if err := _Lifepointstaking.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// LifepointstakingStakedIterator is returned from FilterStaked and is used to iterate over the raw logs and unpacked data for Staked events raised by the Lifepointstaking contract.
type LifepointstakingStakedIterator struct {
	Event *LifepointstakingStaked // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *LifepointstakingStakedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(LifepointstakingStaked)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(LifepointstakingStaked)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *LifepointstakingStakedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *LifepointstakingStakedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// LifepointstakingStaked represents a Staked event raised by the Lifepointstaking contract.
type LifepointstakingStaked struct {
	User         common.Address
	Version      *big.Int
	Amount       *big.Int
	LockDuration *big.Int
	Timestamp    *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterStaked is a free log retrieval operation binding the contract event 0x9cfd25589d1eb8ad71e342a86a8524e83522e3936c0803048c08f6d9ad974f40.
//
// Solidity: event Staked(address indexed user, uint256 version, uint256 amount, uint256 lockDuration, uint256 timestamp)
func (_Lifepointstaking *LifepointstakingFilterer) FilterStaked(opts *bind.FilterOpts, user []common.Address) (*LifepointstakingStakedIterator, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _Lifepointstaking.contract.FilterLogs(opts, "Staked", userRule)
	if err != nil {
		return nil, err
	}
	return &LifepointstakingStakedIterator{contract: _Lifepointstaking.contract, event: "Staked", logs: logs, sub: sub}, nil
}

// WatchStaked is a free log subscription operation binding the contract event 0x9cfd25589d1eb8ad71e342a86a8524e83522e3936c0803048c08f6d9ad974f40.
//
// Solidity: event Staked(address indexed user, uint256 version, uint256 amount, uint256 lockDuration, uint256 timestamp)
func (_Lifepointstaking *LifepointstakingFilterer) WatchStaked(opts *bind.WatchOpts, sink chan<- *LifepointstakingStaked, user []common.Address) (event.Subscription, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _Lifepointstaking.contract.WatchLogs(opts, "Staked", userRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(LifepointstakingStaked)
				if err := _Lifepointstaking.contract.UnpackLog(event, "Staked", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseStaked is a log parse operation binding the contract event 0x9cfd25589d1eb8ad71e342a86a8524e83522e3936c0803048c08f6d9ad974f40.
//
// Solidity: event Staked(address indexed user, uint256 version, uint256 amount, uint256 lockDuration, uint256 timestamp)
func (_Lifepointstaking *LifepointstakingFilterer) ParseStaked(log types.Log) (*LifepointstakingStaked, error) {
	event := new(LifepointstakingStaked)
	if err := _Lifepointstaking.contract.UnpackLog(event, "Staked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// LifepointstakingWithdrawnIterator is returned from FilterWithdrawn and is used to iterate over the raw logs and unpacked data for Withdrawn events raised by the Lifepointstaking contract.
type LifepointstakingWithdrawnIterator struct {
	Event *LifepointstakingWithdrawn // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *LifepointstakingWithdrawnIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(LifepointstakingWithdrawn)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(LifepointstakingWithdrawn)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *LifepointstakingWithdrawnIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *LifepointstakingWithdrawnIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// LifepointstakingWithdrawn represents a Withdrawn event raised by the Lifepointstaking contract.
type LifepointstakingWithdrawn struct {
	User      common.Address
	Version   *big.Int
	Amount    *big.Int
	Timestamp *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterWithdrawn is a free log retrieval operation binding the contract event 0x75e161b3e824b114fc1a33274bd7091918dd4e639cede50b78b15a4eea956a21.
//
// Solidity: event Withdrawn(address indexed user, uint256 version, uint256 amount, uint256 timestamp)
func (_Lifepointstaking *LifepointstakingFilterer) FilterWithdrawn(opts *bind.FilterOpts, user []common.Address) (*LifepointstakingWithdrawnIterator, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _Lifepointstaking.contract.FilterLogs(opts, "Withdrawn", userRule)
	if err != nil {
		return nil, err
	}
	return &LifepointstakingWithdrawnIterator{contract: _Lifepointstaking.contract, event: "Withdrawn", logs: logs, sub: sub}, nil
}

// WatchWithdrawn is a free log subscription operation binding the contract event 0x75e161b3e824b114fc1a33274bd7091918dd4e639cede50b78b15a4eea956a21.
//
// Solidity: event Withdrawn(address indexed user, uint256 version, uint256 amount, uint256 timestamp)
func (_Lifepointstaking *LifepointstakingFilterer) WatchWithdrawn(opts *bind.WatchOpts, sink chan<- *LifepointstakingWithdrawn, user []common.Address) (event.Subscription, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _Lifepointstaking.contract.WatchLogs(opts, "Withdrawn", userRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(LifepointstakingWithdrawn)
				if err := _Lifepointstaking.contract.UnpackLog(event, "Withdrawn", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseWithdrawn is a log parse operation binding the contract event 0x75e161b3e824b114fc1a33274bd7091918dd4e639cede50b78b15a4eea956a21.
//
// Solidity: event Withdrawn(address indexed user, uint256 version, uint256 amount, uint256 timestamp)
func (_Lifepointstaking *LifepointstakingFilterer) ParseWithdrawn(log types.Log) (*LifepointstakingWithdrawn, error) {
	event := new(LifepointstakingWithdrawn)
	if err := _Lifepointstaking.contract.UnpackLog(event, "Withdrawn", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
