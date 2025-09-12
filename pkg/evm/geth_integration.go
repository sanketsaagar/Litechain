package evm

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
)

// EVMExecutor provides full Ethereum Virtual Machine compatibility
type EVMExecutor struct {
	blockchain *core.BlockChain
	stateDB    *state.StateDB
	vmConfig   vm.Config
	chainID    *big.Int
}

// NewEVMExecutor creates a new EVM-compatible execution engine
func NewEVMExecutor(chainID *big.Int, stateDB *state.StateDB) *EVMExecutor {
	return &EVMExecutor{
		chainID:  chainID,
		stateDB:  stateDB,
		vmConfig: vm.Config{},
	}
}

// ExecuteTransaction processes a transaction through the EVM
func (e *EVMExecutor) ExecuteTransaction(tx *types.Transaction, header *types.Header) (*types.Receipt, error) {
	// Create EVM context
	blockContext := core.NewEVMBlockContext(header, e.blockchain, nil)
	txContext := core.NewEVMTxContext(tx)
	
	// Initialize EVM
	evm := vm.NewEVM(blockContext, txContext, e.stateDB, e.getChainConfig(), e.vmConfig)
	
	// Execute transaction
	result, err := core.ApplyTransaction(
		e.getChainConfig(),
		e.blockchain,
		nil, // coinbase will be set by consensus
		e.stateDB,
		header,
		tx,
		&header.GasUsed,
		evm,
	)
	
	return result, err
}

// DeployContract deploys a smart contract
func (e *EVMExecutor) DeployContract(from common.Address, code []byte, value *big.Int, gas uint64) (common.Address, error) {
	// Create contract creation transaction
	tx := types.NewContractCreation(
		e.stateDB.GetNonce(from),
		value,
		gas,
		e.getGasPrice(),
		code,
	)
	
	// Sign and execute (simplified - in reality needs proper signing)
	header := &types.Header{
		Number:     big.NewInt(1),
		GasLimit:   gas * 2,
		Time:       uint64(1234567890),
		Difficulty: big.NewInt(1),
	}
	
	receipt, err := e.ExecuteTransaction(tx, header)
	if err != nil {
		return common.Address{}, err
	}
	
	return receipt.ContractAddress, nil
}

// CallContract calls a smart contract function
func (e *EVMExecutor) CallContract(to common.Address, data []byte, value *big.Int) ([]byte, error) {
	// Create call message
	msg := &core.Message{
		To:         &to,
		From:       common.Address{}, // Will be set by caller
		Value:      value,
		Gas:        1000000, // Default gas limit
		GasPrice:   e.getGasPrice(),
		GasFeeCap:  e.getGasPrice(),
		GasTipCap:  big.NewInt(0),
		Data:       data,
		AccessList: nil,
	}
	
	// Create EVM context
	header := &types.Header{
		Number:     big.NewInt(1),
		GasLimit:   1000000,
		Time:       uint64(1234567890),
		Difficulty: big.NewInt(1),
	}
	
	blockContext := core.NewEVMBlockContext(header, e.blockchain, nil)
	txContext := core.NewEVMTxContext(types.NewTx(&types.LegacyTx{}))
	
	// Initialize EVM and execute
	evm := vm.NewEVM(blockContext, txContext, e.stateDB, e.getChainConfig(), e.vmConfig)
	result, _, err := evm.Call(vm.AccountRef(msg.From), *msg.To, msg.Data, msg.Gas, msg.Value)
	
	return result, err
}

// GetContractCode returns the bytecode of a deployed contract
func (e *EVMExecutor) GetContractCode(address common.Address) []byte {
	return e.stateDB.GetCode(address)
}

// EstimateGas estimates the gas needed for a transaction
func (e *EVMExecutor) EstimateGas(from, to common.Address, data []byte, value *big.Int) (uint64, error) {
	// Create estimation message
	msg := &core.Message{
		From:       from,
		To:         &to,
		Value:      value,
		Gas:        10000000, // High gas limit for estimation
		GasPrice:   e.getGasPrice(),
		GasFeeCap:  e.getGasPrice(),
		GasTipCap:  big.NewInt(0),
		Data:       data,
		AccessList: nil,
	}
	
	// Run estimation (simplified)
	header := &types.Header{
		Number:     big.NewInt(1),
		GasLimit:   10000000,
		Time:       uint64(1234567890),
		Difficulty: big.NewInt(1),
	}
	
	blockContext := core.NewEVMBlockContext(header, e.blockchain, nil)
	txContext := core.NewEVMTxContext(types.NewTx(&types.LegacyTx{}))
	
	evm := vm.NewEVM(blockContext, txContext, e.stateDB, e.getChainConfig(), e.vmConfig)
	
	// Binary search for optimal gas (simplified implementation)
	gasUsed, _, err := evm.Call(vm.AccountRef(msg.From), *msg.To, msg.Data, msg.Gas, msg.Value)
	if err != nil {
		return 0, err
	}
	
	return msg.Gas - gasUsed, nil
}

// getChainConfig returns the chain configuration compatible with Ethereum
func (e *EVMExecutor) getChainConfig() *params.ChainConfig {
	return &params.ChainConfig{
		ChainID:                 e.chainID,
		HomesteadBlock:          big.NewInt(0),
		EIP150Block:             big.NewInt(0),
		EIP155Block:             big.NewInt(0),
		EIP158Block:             big.NewInt(0),
		ByzantiumBlock:          big.NewInt(0),
		ConstantinopleBlock:     big.NewInt(0),
		PetersburgBlock:         big.NewInt(0),
		IstanbulBlock:           big.NewInt(0),
		BerlinBlock:             big.NewInt(0),
		LondonBlock:             big.NewInt(0),
		ArrowGlacierBlock:       big.NewInt(0),
		GrayGlacierBlock:        big.NewInt(0),
		MergeNetsplitBlock:      big.NewInt(0),
		ShanghaiBlock:           big.NewInt(0),
		CancunBlock:             big.NewInt(0),
		PragueBlock:             big.NewInt(0),
		VerkleBlock:             big.NewInt(0),
	}
}

// getGasPrice returns current gas price
func (e *EVMExecutor) getGasPrice() *big.Int {
	// Dynamic gas pricing - will be integrated with economics module
	return big.NewInt(1000000000) // 1 Gwei default
}

// GetBalance returns the balance of an account
func (e *EVMExecutor) GetBalance(address common.Address) *big.Int {
	return e.stateDB.GetBalance(address)
}

// GetNonce returns the nonce of an account
func (e *EVMExecutor) GetNonce(address common.Address) uint64 {
	return e.stateDB.GetNonce(address)
}

// GetStorageAt returns the storage value at a specific slot
func (e *EVMExecutor) GetStorageAt(address common.Address, slot common.Hash) common.Hash {
	return e.stateDB.GetState(address, slot)
}