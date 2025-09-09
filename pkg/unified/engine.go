package unified

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
)

// UnifiedEngine combines consensus and execution in a single system
// Inspired by Polygon's Heimdall (consensus) + Bor (execution) but unified
type UnifiedEngine struct {
	// Consensus components (Heimdall-inspired)
	consensus  *ConsensusEngine
	validators *ValidatorSet
	staking    *StakingManager

	// Execution components (Bor-inspired)
	evm        *EVMEngine
	blockchain *core.BlockChain
	txPool     *core.TxPool

	// Unified state management
	stateManager *UnifiedStateManager

	// L2 specific components
	aggLayer *AggLayerClient
	batcher  *BatchManager

	// Configuration
	config *UnifiedConfig

	// Runtime state
	mu           sync.RWMutex
	running      bool
	currentBlock *types.Block

	// Channels for coordination
	newBlockCh chan *types.Block
	stopCh     chan struct{}
}

// UnifiedConfig holds configuration for the unified engine
type UnifiedConfig struct {
	// Consensus config (Heimdall-like)
	BlockTime        time.Duration `json:"block_time"`
	ValidatorCount   int           `json:"validator_count"`
	ConsensusTimeout time.Duration `json:"consensus_timeout"`

	// Execution config (Bor-like)
	ChainID       uint64 `json:"chain_id"`
	NetworkID     uint64 `json:"network_id"`
	BlockGasLimit uint64 `json:"block_gas_limit"`
	BaseFee       uint64 `json:"base_fee"`

	// L2 specific config
	AggLayerEndpoint  string        `json:"agglayer_endpoint"`
	BatchInterval     time.Duration `json:"batch_interval"`
	L1SubmissionDelay time.Duration `json:"l1_submission_delay"`

	// Development settings
	EnableAutoMining bool          `json:"enable_auto_mining"`
	DevMode          bool          `json:"dev_mode"`
	DevPeriod        time.Duration `json:"dev_period"`
}

// NewUnifiedEngine creates a new unified blockchain engine
func NewUnifiedEngine(config *UnifiedConfig) (*UnifiedEngine, error) {
	// Initialize consensus engine (Heimdall-inspired)
	consensus, err := NewConsensusEngine(&ConsensusConfig{
		BlockTime:        config.BlockTime,
		ValidatorCount:   config.ValidatorCount,
		ConsensusTimeout: config.ConsensusTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create consensus engine: %w", err)
	}

	// Initialize validator set
	validators, err := NewValidatorSet(config.ValidatorCount)
	if err != nil {
		return nil, fmt.Errorf("failed to create validator set: %w", err)
	}

	// Initialize EVM engine (Bor-inspired)
	evm, err := NewEVMEngine(&EVMConfig{
		ChainID:       config.ChainID,
		NetworkID:     config.NetworkID,
		BlockGasLimit: config.BlockGasLimit,
		BaseFee:       config.BaseFee,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create EVM engine: %w", err)
	}

	// Initialize unified state manager
	stateManager, err := NewUnifiedStateManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create state manager: %w", err)
	}

	// Initialize AggLayer client
	aggLayer, err := NewAggLayerClient(config.AggLayerEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create AggLayer client: %w", err)
	}

	// Initialize batch manager
	batcher, err := NewBatchManager(&BatchConfig{
		Interval:          config.BatchInterval,
		L1SubmissionDelay: config.L1SubmissionDelay,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create batch manager: %w", err)
	}

	engine := &UnifiedEngine{
		consensus:    consensus,
		validators:   validators,
		evm:          evm,
		stateManager: stateManager,
		aggLayer:     aggLayer,
		batcher:      batcher,
		config:       config,
		newBlockCh:   make(chan *types.Block, 100),
		stopCh:       make(chan struct{}),
	}

	return engine, nil
}

// Start begins the unified blockchain engine
func (e *UnifiedEngine) Start(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.running {
		return fmt.Errorf("engine already running")
	}

	fmt.Println("🚀 Starting LightChain L2 Unified Engine")
	fmt.Printf("   • Consensus: CometBFT-inspired PoS\n")
	fmt.Printf("   • Execution: Geth-compatible EVM\n")
	fmt.Printf("   • Block Time: %v\n", e.config.BlockTime)
	fmt.Printf("   • Validators: %d\n", e.config.ValidatorCount)

	// Start consensus engine
	if err := e.consensus.Start(ctx); err != nil {
		return fmt.Errorf("failed to start consensus engine: %w", err)
	}

	// Start EVM engine
	if err := e.evm.Start(ctx); err != nil {
		return fmt.Errorf("failed to start EVM engine: %w", err)
	}

	// Start state manager
	if err := e.stateManager.Start(ctx); err != nil {
		return fmt.Errorf("failed to start state manager: %w", err)
	}

	// Start AggLayer integration
	if err := e.aggLayer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start AggLayer client: %w", err)
	}

	// Start batch manager
	if err := e.batcher.Start(ctx); err != nil {
		return fmt.Errorf("failed to start batch manager: %w", err)
	}

	e.running = true

	// Start main processing loop
	go e.processBlocks(ctx)

	// Start auto-mining if enabled (for development)
	if e.config.EnableAutoMining {
		go e.autoMiningLoop(ctx)
	}

	fmt.Println("✅ LightChain L2 Unified Engine started successfully")
	return nil
}

// processBlocks is the main unified processing loop
// This is where the magic happens - consensus and execution work together
func (e *UnifiedEngine) processBlocks(ctx context.Context) {
	fmt.Println("🔄 Starting unified block processing...")

	ticker := time.NewTicker(e.config.BlockTime)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-e.stopCh:
			return
		case <-ticker.C:
			if err := e.produceBlock(ctx); err != nil {
				fmt.Printf("❌ Block production failed: %v\n", err)
			}
		case block := <-e.newBlockCh:
			if err := e.processIncomingBlock(ctx, block); err != nil {
				fmt.Printf("❌ Block processing failed: %v\n", err)
			}
		}
	}
}

// produceBlock creates a new block using unified consensus + execution
func (e *UnifiedEngine) produceBlock(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 1. Check if we should produce a block (consensus check)
	shouldProduce, err := e.consensus.ShouldProduceBlock()
	if err != nil {
		return fmt.Errorf("consensus check failed: %w", err)
	}

	if !shouldProduce {
		return nil // Not our turn or network conditions not met
	}

	// 2. Get pending transactions from EVM mempool
	pendingTxs := e.evm.GetPendingTransactions()

	// 3. Create block header (unified approach)
	header := &types.Header{
		ParentHash: e.stateManager.GetCurrentBlockHash(),
		Number:     e.stateManager.GetCurrentBlockNumber().Add(e.stateManager.GetCurrentBlockNumber(), common.Big1),
		Time:       uint64(time.Now().Unix()),
		GasLimit:   e.config.BlockGasLimit,
		BaseFee:    new(big.Int).SetUint64(e.config.BaseFee),
		Difficulty: common.Big1, // PoS doesn't use difficulty
	}

	// 4. Execute transactions in EVM (Bor-inspired)
	receipts, err := e.evm.ExecuteTransactions(header, pendingTxs)
	if err != nil {
		return fmt.Errorf("transaction execution failed: %w", err)
	}

	// 5. Create block with executed transactions
	block := types.NewBlock(header, pendingTxs, nil, receipts, nil)

	// 6. Consensus validation (Heimdall-inspired)
	if err := e.consensus.ValidateBlock(block); err != nil {
		return fmt.Errorf("consensus validation failed: %w", err)
	}

	// 7. Apply state changes atomically
	if err := e.stateManager.ApplyBlock(block, receipts); err != nil {
		return fmt.Errorf("state application failed: %w", err)
	}

	// 8. Finalize consensus (get validator signatures)
	if err := e.consensus.FinalizeBlock(block); err != nil {
		return fmt.Errorf("consensus finalization failed: %w", err)
	}

	// 9. Update current block
	e.currentBlock = block

	// 10. Submit to AggLayer for L1 settlement
	if err := e.batcher.AddBlock(block); err != nil {
		fmt.Printf("⚠️  Failed to add block to batch: %v\n", err)
		// Don't fail block production for batching issues
	}

	// 11. Broadcast block to network
	e.broadcastBlock(block)

	fmt.Printf("🧱 Block #%d produced: %d txs, %s gas, validator: %s\n",
		block.Number().Uint64(),
		len(block.Transactions()),
		block.GasUsed(),
		e.consensus.GetCurrentValidator().String()[:10])

	return nil
}

// processIncomingBlock processes blocks received from other validators
func (e *UnifiedEngine) processIncomingBlock(ctx context.Context, block *types.Block) error {
	// 1. Consensus validation
	if err := e.consensus.ValidateBlock(block); err != nil {
		return fmt.Errorf("consensus validation failed: %w", err)
	}

	// 2. Execute transactions to verify state
	receipts, err := e.evm.ExecuteTransactions(block.Header(), block.Transactions())
	if err != nil {
		return fmt.Errorf("transaction execution failed: %w", err)
	}

	// 3. Verify state root matches
	if block.Root() != receipts[len(receipts)-1].PostState {
		return fmt.Errorf("state root mismatch")
	}

	// 4. Apply state changes
	if err := e.stateManager.ApplyBlock(block, receipts); err != nil {
		return fmt.Errorf("state application failed: %w", err)
	}

	// 5. Update current block
	e.mu.Lock()
	e.currentBlock = block
	e.mu.Unlock()

	fmt.Printf("✅ Block #%d processed: %d txs\n", block.Number().Uint64(), len(block.Transactions()))
	return nil
}

// autoMiningLoop produces blocks automatically for development
func (e *UnifiedEngine) autoMiningLoop(ctx context.Context) {
	fmt.Printf("⛏️  Auto-mining enabled: blocks every %v\n", e.config.DevPeriod)

	ticker := time.NewTicker(e.config.DevPeriod)
	defer ticker.Stop()

	blockCount := 0

	for {
		select {
		case <-ctx.Done():
			return
		case <-e.stopCh:
			return
		case <-ticker.C:
			if err := e.produceBlock(ctx); err != nil {
				fmt.Printf("❌ Auto-mining failed: %v\n", err)
			} else {
				blockCount++
				if blockCount%10 == 0 {
					fmt.Printf("📊 Auto-mined %d blocks\n", blockCount)
				}
			}
		}
	}
}

// broadcastBlock sends the block to other validators
func (e *UnifiedEngine) broadcastBlock(block *types.Block) {
	// TODO: Implement P2P broadcasting
	// For now, just log
	fmt.Printf("📡 Broadcasting block #%d to network\n", block.Number().Uint64())
}

// GetCurrentBlock returns the current block
func (e *UnifiedEngine) GetCurrentBlock() *types.Block {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.currentBlock
}

// GetBlockByNumber returns a block by number
func (e *UnifiedEngine) GetBlockByNumber(number uint64) (*types.Block, error) {
	return e.stateManager.GetBlockByNumber(number)
}

// SubmitTransaction submits a transaction to the unified mempool
func (e *UnifiedEngine) SubmitTransaction(tx *types.Transaction) error {
	return e.evm.SubmitTransaction(tx)
}

// GetBalance returns the balance of an account
func (e *UnifiedEngine) GetBalance(address common.Address) (*big.Int, error) {
	return e.stateManager.GetBalance(address)
}

// Stop stops the unified engine
func (e *UnifiedEngine) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.running {
		return nil
	}

	fmt.Println("🛑 Stopping LightChain L2 Unified Engine...")

	close(e.stopCh)

	// Stop all components
	e.batcher.Stop()
	e.aggLayer.Stop()
	e.stateManager.Stop()
	e.evm.Stop()
	e.consensus.Stop()

	e.running = false

	fmt.Println("✅ LightChain L2 Unified Engine stopped")
	return nil
}
