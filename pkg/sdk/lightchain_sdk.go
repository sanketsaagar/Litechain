package sdk

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

// LightChainSDK provides easy-to-use development tools
type LightChainSDK struct {
	client     *ethclient.Client
	rpcClient  *rpc.Client
	chainID    *big.Int
	privateKey *ecdsa.PrivateKey
}

// Config for SDK initialization
type SDKConfig struct {
	NodeURL    string `json:"nodeUrl"`
	PrivateKey string `json:"privateKey,omitempty"`
	ChainID    int64  `json:"chainId"`
}

// NewSDK creates a new LightChain SDK instance
func NewSDK(config SDKConfig) (*LightChainSDK, error) {
	// Connect to node
	rpcClient, err := rpc.Dial(config.NodeURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to node: %w", err)
	}
	
	client := ethclient.NewClient(rpcClient)
	
	sdk := &LightChainSDK{
		client:    client,
		rpcClient: rpcClient,
		chainID:   big.NewInt(config.ChainID),
	}
	
	// Set private key if provided
	if config.PrivateKey != "" {
		privateKey, err := crypto.HexToECDSA(config.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("invalid private key: %w", err)
		}
		sdk.privateKey = privateKey
	}
	
	return sdk, nil
}

// Account management
func (sdk *LightChainSDK) CreateAccount() (*Account, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	
	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	
	return &Account{
		Address:    address,
		PrivateKey: privateKey,
	}, nil
}

// GetBalance returns account balance
func (sdk *LightChainSDK) GetBalance(address common.Address) (*big.Int, error) {
	return sdk.client.BalanceAt(context.Background(), address, nil)
}

// SendTransaction sends a transaction
func (sdk *LightChainSDK) SendTransaction(to common.Address, amount *big.Int, data []byte) (*types.Transaction, error) {
	if sdk.privateKey == nil {
		return nil, fmt.Errorf("private key not set")
	}
	
	from := crypto.PubkeyToAddress(sdk.privateKey.PublicKey)
	nonce, err := sdk.client.PendingNonceAt(context.Background(), from)
	if err != nil {
		return nil, err
	}
	
	gasPrice, err := sdk.client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}
	
	gasLimit := uint64(21000)
	if len(data) > 0 {
		gasLimit = 100000 // Higher limit for contract calls
	}
	
	tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, data)
	
	signer := types.NewEIP155Signer(sdk.chainID)
	signedTx, err := types.SignTx(tx, signer, sdk.privateKey)
	if err != nil {
		return nil, err
	}
	
	err = sdk.client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, err
	}
	
	return signedTx, nil
}

// DeployContract deploys a smart contract
func (sdk *LightChainSDK) DeployContract(abi, bytecode string, params ...interface{}) (*ContractDeployment, error) {
	if sdk.privateKey == nil {
		return nil, fmt.Errorf("private key not set")
	}
	
	from := crypto.PubkeyToAddress(sdk.privateKey.PublicKey)
	
	// Create transactor
	auth, err := bind.NewKeyedTransactorWithChainID(sdk.privateKey, sdk.chainID)
	if err != nil {
		return nil, err
	}
	
	// Get gas price
	gasPrice, err := sdk.client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}
	auth.GasPrice = gasPrice
	
	// Deploy contract (simplified - would need proper ABI parsing)
	// This is a placeholder for contract deployment logic
	tx := &types.Transaction{}
	
	return &ContractDeployment{
		Transaction: tx,
		Address:     from, // Placeholder
	}, nil
}

// CallContract calls a smart contract function
func (sdk *LightChainSDK) CallContract(contractAddress common.Address, methodName string, params ...interface{}) ([]byte, error) {
	// Placeholder for contract call logic
	// In real implementation, would encode ABI call data
	data := []byte{} // Encoded call data would go here
	
	msg := map[string]interface{}{
		"to":   contractAddress,
		"data": fmt.Sprintf("0x%x", data),
	}
	
	var result string
	err := sdk.rpcClient.Call(&result, "eth_call", msg, "latest")
	if err != nil {
		return nil, err
	}
	
	return common.FromHex(result), nil
}

// WaitForTransaction waits for transaction confirmation
func (sdk *LightChainSDK) WaitForTransaction(txHash common.Hash) (*types.Receipt, error) {
	ctx := context.Background()
	
	for i := 0; i < 60; i++ { // Wait up to 2 minutes
		receipt, err := sdk.client.TransactionReceipt(ctx, txHash)
		if err == nil {
			return receipt, nil
		}
		
		// Wait 2 seconds (block time)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-context.Background().Done():
			// Continue waiting
		}
	}
	
	return nil, fmt.Errorf("transaction not confirmed after 2 minutes")
}

// GetTransactionReceipt gets transaction receipt
func (sdk *LightChainSDK) GetTransactionReceipt(txHash common.Hash) (*types.Receipt, error) {
	return sdk.client.TransactionReceipt(context.Background(), txHash)
}

// EstimateGas estimates gas for a transaction
func (sdk *LightChainSDK) EstimateGas(to common.Address, data []byte) (uint64, error) {
	from := common.Address{}
	if sdk.privateKey != nil {
		from = crypto.PubkeyToAddress(sdk.privateKey.PublicKey)
	}
	
	msg := map[string]interface{}{
		"from": from,
		"to":   to,
		"data": fmt.Sprintf("0x%x", data),
	}
	
	var result string
	err := sdk.rpcClient.Call(&result, "eth_estimateGas", msg)
	if err != nil {
		return 0, err
	}
	
	// Simplified gas estimation conversion
	hexBytes := common.FromHex(result)
	if len(hexBytes) == 0 {
		return 21000, nil // Default gas limit
	}
	return uint64(hexBytes[0]) * 256, nil
}

// Validator operations
func (sdk *LightChainSDK) StakeTokens(amount *big.Int) (*types.Transaction, error) {
	// Placeholder for staking logic
	// Would call staking contract or native staking function
	return sdk.SendTransaction(common.Address{}, amount, nil)
}

// GetStakingInfo returns staking information
func (sdk *LightChainSDK) GetStakingInfo(validator common.Address) (*StakingInfo, error) {
	// Placeholder for staking info retrieval
	return &StakingInfo{
		Validator:     validator,
		StakedAmount:  big.NewInt(0),
		Rewards:       big.NewInt(0),
		Performance:   0.95,
		IsActive:      true,
	}, nil
}

// Event handling
func (sdk *LightChainSDK) SubscribeToLogs(contractAddress common.Address, topics [][]common.Hash) (<-chan types.Log, error) {
	// Create subscription for contract events
	logs := make(chan types.Log)
	
	// Placeholder implementation
	go func() {
		defer close(logs)
		// Event subscription logic would go here
	}()
	
	return logs, nil
}

// Data structures
type Account struct {
	Address    common.Address
	PrivateKey *ecdsa.PrivateKey
}

type ContractDeployment struct {
	Transaction *types.Transaction
	Address     common.Address
}

type StakingInfo struct {
	Validator     common.Address
	StakedAmount  *big.Int
	Rewards       *big.Int
	Performance   float64
	IsActive      bool
}

// Utility functions
func (sdk *LightChainSDK) ToWei(amount float64) *big.Int {
	wei := new(big.Float).Mul(big.NewFloat(amount), big.NewFloat(1e18))
	result, _ := wei.Int(nil)
	return result
}

func (sdk *LightChainSDK) FromWei(wei *big.Int) float64 {
	f := new(big.Float).SetInt(wei)
	f = f.Quo(f, big.NewFloat(1e18))
	result, _ := f.Float64()
	return result
}

func (sdk *LightChainSDK) IsValidAddress(address string) bool {
	return common.IsHexAddress(address)
}

func (sdk *LightChainSDK) AddressFromString(address string) common.Address {
	return common.HexToAddress(address)
}

func (sdk *LightChainSDK) Close() {
	if sdk.client != nil {
		sdk.client.Close()
	}
	if sdk.rpcClient != nil {
		sdk.rpcClient.Close()
	}
}