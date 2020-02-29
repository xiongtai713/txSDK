// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package sol

import (
	"pdx-chain/accounts/abi"
	"pdx-chain/accounts/abi/bind"
	"pdx-chain/common"
	"pdx-chain/core/types"
	"strings"
)

// ApinameABI is the input ABI used to generate the binding from.
const ApinameABI = "[{\"constant\":false,\"inputs\":[],\"name\":\"kill\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_newgreeting\",\"type\":\"string\"}],\"name\":\"setGreeting\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"greet\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_greeting\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"}]"

// Apiname is an auto generated Go binding around an Ethereum contract.
type Apiname struct {
	ApinameCaller     // Read-only binding to the contract
	ApinameTransactor // Write-only binding to the contract
	ApinameFilterer   // Log filterer for contract events
}

// ApinameCaller is an auto generated read-only Go binding around an Ethereum contract.
type ApinameCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ApinameTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ApinameTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ApinameFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ApinameFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ApinameSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ApinameSession struct {
	Contract     *Apiname          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ApinameCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ApinameCallerSession struct {
	Contract *ApinameCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// ApinameTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ApinameTransactorSession struct {
	Contract     *ApinameTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// ApinameRaw is an auto generated low-level Go binding around an Ethereum contract.
type ApinameRaw struct {
	Contract *Apiname // Generic contract binding to access the raw methods on
}

// ApinameCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ApinameCallerRaw struct {
	Contract *ApinameCaller // Generic read-only contract binding to access the raw methods on
}

// ApinameTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ApinameTransactorRaw struct {
	Contract *ApinameTransactor // Generic write-only contract binding to access the raw methods on
}

// NewApiname creates a new instance of Apiname, bound to a specific deployed contract.
func NewApiname(address common.Address, backend bind.ContractBackend) (*Apiname, error) {
	contract, err := bindApiname(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Apiname{ApinameCaller: ApinameCaller{contract: contract}, ApinameTransactor: ApinameTransactor{contract: contract}, ApinameFilterer: ApinameFilterer{contract: contract}}, nil
}

// NewApinameCaller creates a new read-only instance of Apiname, bound to a specific deployed contract.
func NewApinameCaller(address common.Address, caller bind.ContractCaller) (*ApinameCaller, error) {
	contract, err := bindApiname(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ApinameCaller{contract: contract}, nil
}

// NewApinameTransactor creates a new write-only instance of Apiname, bound to a specific deployed contract.
func NewApinameTransactor(address common.Address, transactor bind.ContractTransactor) (*ApinameTransactor, error) {
	contract, err := bindApiname(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ApinameTransactor{contract: contract}, nil
}

// NewApinameFilterer creates a new log filterer instance of Apiname, bound to a specific deployed contract.
func NewApinameFilterer(address common.Address, filterer bind.ContractFilterer) (*ApinameFilterer, error) {
	contract, err := bindApiname(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ApinameFilterer{contract: contract}, nil
}

// bindApiname binds a generic wrapper to an already deployed contract.
func bindApiname(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ApinameABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Apiname *ApinameRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Apiname.Contract.ApinameCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Apiname *ApinameRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Apiname.Contract.ApinameTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Apiname *ApinameRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Apiname.Contract.ApinameTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Apiname *ApinameCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Apiname.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Apiname *ApinameTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Apiname.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Apiname *ApinameTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Apiname.Contract.contract.Transact(opts, method, params...)
}

// Greet is a free data retrieval call binding the contract method 0xcfae3217.
//
// Solidity: function greet() constant returns(string)
func (_Apiname *ApinameCaller) Greet(opts *bind.CallOpts) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _Apiname.contract.Call(opts, out, "greet")
	return *ret0, err
}

// Greet is a free data retrieval call binding the contract method 0xcfae3217.
//
// Solidity: function greet() constant returns(string)
func (_Apiname *ApinameSession) Greet() (string, error) {
	return _Apiname.Contract.Greet(&_Apiname.CallOpts)
}

// Greet is a free data retrieval call binding the contract method 0xcfae3217.
//
// Solidity: function greet() constant returns(string)
func (_Apiname *ApinameCallerSession) Greet() (string, error) {
	return _Apiname.Contract.Greet(&_Apiname.CallOpts)
}

// Kill is a paid mutator transaction binding the contract method 0x41c0e1b5.
//
// Solidity: function kill() returns()
func (_Apiname *ApinameTransactor) Kill(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Apiname.contract.Transact(opts, "kill")
}

// Kill is a paid mutator transaction binding the contract method 0x41c0e1b5.
//
// Solidity: function kill() returns()
func (_Apiname *ApinameSession) Kill() (*types.Transaction, error) {
	return _Apiname.Contract.Kill(&_Apiname.TransactOpts)
}

// Kill is a paid mutator transaction binding the contract method 0x41c0e1b5.
//
// Solidity: function kill() returns()
func (_Apiname *ApinameTransactorSession) Kill() (*types.Transaction, error) {
	return _Apiname.Contract.Kill(&_Apiname.TransactOpts)
}

// SetGreeting is a paid mutator transaction binding the contract method 0xa4136862.
//
// Solidity: function setGreeting(_newgreeting string) returns()
func (_Apiname *ApinameTransactor) SetGreeting(opts *bind.TransactOpts, _newgreeting string) (*types.Transaction, error) {
	return _Apiname.contract.Transact(opts, "setGreeting", _newgreeting)
}

// SetGreeting is a paid mutator transaction binding the contract method 0xa4136862.
//
// Solidity: function setGreeting(_newgreeting string) returns()
func (_Apiname *ApinameSession) SetGreeting(_newgreeting string) (*types.Transaction, error) {
	return _Apiname.Contract.SetGreeting(&_Apiname.TransactOpts, _newgreeting)
}

// SetGreeting is a paid mutator transaction binding the contract method 0xa4136862.
//
// Solidity: function setGreeting(_newgreeting string) returns()
func (_Apiname *ApinameTransactorSession) SetGreeting(_newgreeting string) (*types.Transaction, error) {
	return _Apiname.Contract.SetGreeting(&_Apiname.TransactOpts, _newgreeting)
}
