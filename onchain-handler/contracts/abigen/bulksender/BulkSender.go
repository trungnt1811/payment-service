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
	ABI: "[{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"recipientsLength\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amountsLength\",\"type\":\"uint256\"}],\"name\":\"ArraysLengthMismatch\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"ERC20Transfer\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"NativeTransfer\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"recipients\",\"type\":\"address[]\"},{\"internalType\":\"uint256[]\",\"name\":\"amounts\",\"type\":\"uint256[]\"},{\"internalType\":\"address\",\"name\":\"tokenAddress\",\"type\":\"address\"}],\"name\":\"bulkTransfer\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"}]",
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

// BulkTransfer is a paid mutator transaction binding the contract method 0x89ef8292.
//
// Solidity: function bulkTransfer(address[] recipients, uint256[] amounts, address tokenAddress) payable returns()
func (_Bulksender *BulksenderTransactor) BulkTransfer(opts *bind.TransactOpts, recipients []common.Address, amounts []*big.Int, tokenAddress common.Address) (*types.Transaction, error) {
	return _Bulksender.contract.Transact(opts, "bulkTransfer", recipients, amounts, tokenAddress)
}

// BulkTransfer is a paid mutator transaction binding the contract method 0x89ef8292.
//
// Solidity: function bulkTransfer(address[] recipients, uint256[] amounts, address tokenAddress) payable returns()
func (_Bulksender *BulksenderSession) BulkTransfer(recipients []common.Address, amounts []*big.Int, tokenAddress common.Address) (*types.Transaction, error) {
	return _Bulksender.Contract.BulkTransfer(&_Bulksender.TransactOpts, recipients, amounts, tokenAddress)
}

// BulkTransfer is a paid mutator transaction binding the contract method 0x89ef8292.
//
// Solidity: function bulkTransfer(address[] recipients, uint256[] amounts, address tokenAddress) payable returns()
func (_Bulksender *BulksenderTransactorSession) BulkTransfer(recipients []common.Address, amounts []*big.Int, tokenAddress common.Address) (*types.Transaction, error) {
	return _Bulksender.Contract.BulkTransfer(&_Bulksender.TransactOpts, recipients, amounts, tokenAddress)
}

// BulksenderERC20TransferIterator is returned from FilterERC20Transfer and is used to iterate over the raw logs and unpacked data for ERC20Transfer events raised by the Bulksender contract.
type BulksenderERC20TransferIterator struct {
	Event *BulksenderERC20Transfer // Event containing the contract specifics and raw log

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
func (it *BulksenderERC20TransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BulksenderERC20Transfer)
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
		it.Event = new(BulksenderERC20Transfer)
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
func (it *BulksenderERC20TransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BulksenderERC20TransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BulksenderERC20Transfer represents a ERC20Transfer event raised by the Bulksender contract.
type BulksenderERC20Transfer struct {
	Token  common.Address
	From   common.Address
	To     common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterERC20Transfer is a free log retrieval operation binding the contract event 0x16e1b64802081839623a4bec223b20b6ee097d9edd8fcef3d4ceb3a94271306e.
//
// Solidity: event ERC20Transfer(address indexed token, address indexed from, address indexed to, uint256 amount)
func (_Bulksender *BulksenderFilterer) FilterERC20Transfer(opts *bind.FilterOpts, token []common.Address, from []common.Address, to []common.Address) (*BulksenderERC20TransferIterator, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _Bulksender.contract.FilterLogs(opts, "ERC20Transfer", tokenRule, fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &BulksenderERC20TransferIterator{contract: _Bulksender.contract, event: "ERC20Transfer", logs: logs, sub: sub}, nil
}

// WatchERC20Transfer is a free log subscription operation binding the contract event 0x16e1b64802081839623a4bec223b20b6ee097d9edd8fcef3d4ceb3a94271306e.
//
// Solidity: event ERC20Transfer(address indexed token, address indexed from, address indexed to, uint256 amount)
func (_Bulksender *BulksenderFilterer) WatchERC20Transfer(opts *bind.WatchOpts, sink chan<- *BulksenderERC20Transfer, token []common.Address, from []common.Address, to []common.Address) (event.Subscription, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _Bulksender.contract.WatchLogs(opts, "ERC20Transfer", tokenRule, fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BulksenderERC20Transfer)
				if err := _Bulksender.contract.UnpackLog(event, "ERC20Transfer", log); err != nil {
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

// ParseERC20Transfer is a log parse operation binding the contract event 0x16e1b64802081839623a4bec223b20b6ee097d9edd8fcef3d4ceb3a94271306e.
//
// Solidity: event ERC20Transfer(address indexed token, address indexed from, address indexed to, uint256 amount)
func (_Bulksender *BulksenderFilterer) ParseERC20Transfer(log types.Log) (*BulksenderERC20Transfer, error) {
	event := new(BulksenderERC20Transfer)
	if err := _Bulksender.contract.UnpackLog(event, "ERC20Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BulksenderNativeTransferIterator is returned from FilterNativeTransfer and is used to iterate over the raw logs and unpacked data for NativeTransfer events raised by the Bulksender contract.
type BulksenderNativeTransferIterator struct {
	Event *BulksenderNativeTransfer // Event containing the contract specifics and raw log

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
func (it *BulksenderNativeTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BulksenderNativeTransfer)
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
		it.Event = new(BulksenderNativeTransfer)
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
func (it *BulksenderNativeTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BulksenderNativeTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BulksenderNativeTransfer represents a NativeTransfer event raised by the Bulksender contract.
type BulksenderNativeTransfer struct {
	From   common.Address
	To     common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterNativeTransfer is a free log retrieval operation binding the contract event 0xce8688f853ffa65c042b72302433c25d7a230c322caba0901587534b6551091d.
//
// Solidity: event NativeTransfer(address indexed from, address indexed to, uint256 amount)
func (_Bulksender *BulksenderFilterer) FilterNativeTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*BulksenderNativeTransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _Bulksender.contract.FilterLogs(opts, "NativeTransfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &BulksenderNativeTransferIterator{contract: _Bulksender.contract, event: "NativeTransfer", logs: logs, sub: sub}, nil
}

// WatchNativeTransfer is a free log subscription operation binding the contract event 0xce8688f853ffa65c042b72302433c25d7a230c322caba0901587534b6551091d.
//
// Solidity: event NativeTransfer(address indexed from, address indexed to, uint256 amount)
func (_Bulksender *BulksenderFilterer) WatchNativeTransfer(opts *bind.WatchOpts, sink chan<- *BulksenderNativeTransfer, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _Bulksender.contract.WatchLogs(opts, "NativeTransfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BulksenderNativeTransfer)
				if err := _Bulksender.contract.UnpackLog(event, "NativeTransfer", log); err != nil {
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

// ParseNativeTransfer is a log parse operation binding the contract event 0xce8688f853ffa65c042b72302433c25d7a230c322caba0901587534b6551091d.
//
// Solidity: event NativeTransfer(address indexed from, address indexed to, uint256 amount)
func (_Bulksender *BulksenderFilterer) ParseNativeTransfer(log types.Log) (*BulksenderNativeTransfer, error) {
	event := new(BulksenderNativeTransfer)
	if err := _Bulksender.contract.UnpackLog(event, "NativeTransfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
