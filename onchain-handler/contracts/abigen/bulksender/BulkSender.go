// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bulksender

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

// BulksenderMetaData contains all meta data concerning the Bulksender contract.
var BulksenderMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"totalAmount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"recipientCount\",\"type\":\"uint256\"}],\"name\":\"BulkTransfer\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"address[]\",\"name\":\"recipients\",\"type\":\"address[]\"},{\"internalType\":\"uint256[]\",\"name\":\"values\",\"type\":\"uint256[]\"}],\"name\":\"bulkTransfer\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
}

// BulksenderABI is the input ABI used to generate the binding from.
// Deprecated: Use BulksenderMetaData.ABI instead.
var BulksenderABI = BulksenderMetaData.ABI

// Bulksender is an auto generated Go binding around an Ethereum contract.
type Bulksender struct {
	BulksenderCaller     // Read-only binding to the contract
	BulksenderTransactor // Write-only binding to the contract
	BulksenderFilterer   // Log filterer for contract events
}

// BulksenderCaller is an auto generated read-only Go binding around an Ethereum contract.
type BulksenderCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BulksenderTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BulksenderTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BulksenderFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BulksenderFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BulksenderSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BulksenderSession struct {
	Contract     *Bulksender       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BulksenderCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BulksenderCallerSession struct {
	Contract *BulksenderCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// BulksenderTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BulksenderTransactorSession struct {
	Contract     *BulksenderTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// BulksenderRaw is an auto generated low-level Go binding around an Ethereum contract.
type BulksenderRaw struct {
	Contract *Bulksender // Generic contract binding to access the raw methods on
}

// BulksenderCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BulksenderCallerRaw struct {
	Contract *BulksenderCaller // Generic read-only contract binding to access the raw methods on
}

// BulksenderTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BulksenderTransactorRaw struct {
	Contract *BulksenderTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBulksender creates a new instance of Bulksender, bound to a specific deployed contract.
func NewBulksender(address common.Address, backend bind.ContractBackend) (*Bulksender, error) {
	contract, err := bindBulksender(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Bulksender{BulksenderCaller: BulksenderCaller{contract: contract}, BulksenderTransactor: BulksenderTransactor{contract: contract}, BulksenderFilterer: BulksenderFilterer{contract: contract}}, nil
}

// NewBulksenderCaller creates a new read-only instance of Bulksender, bound to a specific deployed contract.
func NewBulksenderCaller(address common.Address, caller bind.ContractCaller) (*BulksenderCaller, error) {
	contract, err := bindBulksender(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BulksenderCaller{contract: contract}, nil
}

// NewBulksenderTransactor creates a new write-only instance of Bulksender, bound to a specific deployed contract.
func NewBulksenderTransactor(address common.Address, transactor bind.ContractTransactor) (*BulksenderTransactor, error) {
	contract, err := bindBulksender(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BulksenderTransactor{contract: contract}, nil
}

// NewBulksenderFilterer creates a new log filterer instance of Bulksender, bound to a specific deployed contract.
func NewBulksenderFilterer(address common.Address, filterer bind.ContractFilterer) (*BulksenderFilterer, error) {
	contract, err := bindBulksender(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BulksenderFilterer{contract: contract}, nil
}

// bindBulksender binds a generic wrapper to an already deployed contract.
func bindBulksender(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := BulksenderMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Bulksender *BulksenderRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Bulksender.Contract.BulksenderCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Bulksender *BulksenderRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bulksender.Contract.BulksenderTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Bulksender *BulksenderRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Bulksender.Contract.BulksenderTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Bulksender *BulksenderCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Bulksender.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Bulksender *BulksenderTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bulksender.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Bulksender *BulksenderTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Bulksender.Contract.contract.Transact(opts, method, params...)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Bulksender *BulksenderCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Bulksender.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Bulksender *BulksenderSession) Owner() (common.Address, error) {
	return _Bulksender.Contract.Owner(&_Bulksender.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Bulksender *BulksenderCallerSession) Owner() (common.Address, error) {
	return _Bulksender.Contract.Owner(&_Bulksender.CallOpts)
}

// BulkTransfer is a paid mutator transaction binding the contract method 0xe886dade.
//
// Solidity: function bulkTransfer(address token, address[] recipients, uint256[] values) returns()
func (_Bulksender *BulksenderTransactor) BulkTransfer(opts *bind.TransactOpts, token common.Address, recipients []common.Address, values []*big.Int) (*types.Transaction, error) {
	return _Bulksender.contract.Transact(opts, "bulkTransfer", token, recipients, values)
}

// BulkTransfer is a paid mutator transaction binding the contract method 0xe886dade.
//
// Solidity: function bulkTransfer(address token, address[] recipients, uint256[] values) returns()
func (_Bulksender *BulksenderSession) BulkTransfer(token common.Address, recipients []common.Address, values []*big.Int) (*types.Transaction, error) {
	return _Bulksender.Contract.BulkTransfer(&_Bulksender.TransactOpts, token, recipients, values)
}

// BulkTransfer is a paid mutator transaction binding the contract method 0xe886dade.
//
// Solidity: function bulkTransfer(address token, address[] recipients, uint256[] values) returns()
func (_Bulksender *BulksenderTransactorSession) BulkTransfer(token common.Address, recipients []common.Address, values []*big.Int) (*types.Transaction, error) {
	return _Bulksender.Contract.BulkTransfer(&_Bulksender.TransactOpts, token, recipients, values)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Bulksender *BulksenderTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bulksender.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Bulksender *BulksenderSession) RenounceOwnership() (*types.Transaction, error) {
	return _Bulksender.Contract.RenounceOwnership(&_Bulksender.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Bulksender *BulksenderTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Bulksender.Contract.RenounceOwnership(&_Bulksender.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Bulksender *BulksenderTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Bulksender.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Bulksender *BulksenderSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Bulksender.Contract.TransferOwnership(&_Bulksender.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Bulksender *BulksenderTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Bulksender.Contract.TransferOwnership(&_Bulksender.TransactOpts, newOwner)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Bulksender *BulksenderTransactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Bulksender.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Bulksender *BulksenderSession) Receive() (*types.Transaction, error) {
	return _Bulksender.Contract.Receive(&_Bulksender.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Bulksender *BulksenderTransactorSession) Receive() (*types.Transaction, error) {
	return _Bulksender.Contract.Receive(&_Bulksender.TransactOpts)
}

// BulksenderBulkTransferIterator is returned from FilterBulkTransfer and is used to iterate over the raw logs and unpacked data for BulkTransfer events raised by the Bulksender contract.
type BulksenderBulkTransferIterator struct {
	Event *BulksenderBulkTransfer // Event containing the contract specifics and raw log

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
func (it *BulksenderBulkTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BulksenderBulkTransfer)
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
		it.Event = new(BulksenderBulkTransfer)
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
func (it *BulksenderBulkTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BulksenderBulkTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BulksenderBulkTransfer represents a BulkTransfer event raised by the Bulksender contract.
type BulksenderBulkTransfer struct {
	Token          common.Address
	TotalAmount    *big.Int
	RecipientCount *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterBulkTransfer is a free log retrieval operation binding the contract event 0xda0ea7dc6081f48fde4813039347dfa02348fa095aa9e7c152e07ef4e474e2f6.
//
// Solidity: event BulkTransfer(address indexed token, uint256 totalAmount, uint256 recipientCount)
func (_Bulksender *BulksenderFilterer) FilterBulkTransfer(opts *bind.FilterOpts, token []common.Address) (*BulksenderBulkTransferIterator, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _Bulksender.contract.FilterLogs(opts, "BulkTransfer", tokenRule)
	if err != nil {
		return nil, err
	}
	return &BulksenderBulkTransferIterator{contract: _Bulksender.contract, event: "BulkTransfer", logs: logs, sub: sub}, nil
}

// WatchBulkTransfer is a free log subscription operation binding the contract event 0xda0ea7dc6081f48fde4813039347dfa02348fa095aa9e7c152e07ef4e474e2f6.
//
// Solidity: event BulkTransfer(address indexed token, uint256 totalAmount, uint256 recipientCount)
func (_Bulksender *BulksenderFilterer) WatchBulkTransfer(opts *bind.WatchOpts, sink chan<- *BulksenderBulkTransfer, token []common.Address) (event.Subscription, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _Bulksender.contract.WatchLogs(opts, "BulkTransfer", tokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BulksenderBulkTransfer)
				if err := _Bulksender.contract.UnpackLog(event, "BulkTransfer", log); err != nil {
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

// ParseBulkTransfer is a log parse operation binding the contract event 0xda0ea7dc6081f48fde4813039347dfa02348fa095aa9e7c152e07ef4e474e2f6.
//
// Solidity: event BulkTransfer(address indexed token, uint256 totalAmount, uint256 recipientCount)
func (_Bulksender *BulksenderFilterer) ParseBulkTransfer(log types.Log) (*BulksenderBulkTransfer, error) {
	event := new(BulksenderBulkTransfer)
	if err := _Bulksender.contract.UnpackLog(event, "BulkTransfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BulksenderOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Bulksender contract.
type BulksenderOwnershipTransferredIterator struct {
	Event *BulksenderOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *BulksenderOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BulksenderOwnershipTransferred)
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
		it.Event = new(BulksenderOwnershipTransferred)
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
func (it *BulksenderOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BulksenderOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BulksenderOwnershipTransferred represents a OwnershipTransferred event raised by the Bulksender contract.
type BulksenderOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Bulksender *BulksenderFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*BulksenderOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Bulksender.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &BulksenderOwnershipTransferredIterator{contract: _Bulksender.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Bulksender *BulksenderFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *BulksenderOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Bulksender.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BulksenderOwnershipTransferred)
				if err := _Bulksender.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_Bulksender *BulksenderFilterer) ParseOwnershipTransferred(log types.Log) (*BulksenderOwnershipTransferred, error) {
	event := new(BulksenderOwnershipTransferred)
	if err := _Bulksender.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
