package unified

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
)

// EVMEngine implements an Erigon-inspired parallel execution engine
// Key features from Erigon: parallel execution, MDBX storage, stage-based processing
type EVMEngine struct {
	// Core EVM components
	chainConfig *params.ChainConfig
	vmConfig    *vm.Config

	// Erigon-inspired components
	parallelExecutor *ParallelExecutor
	stageSequence    *StageSequence
	txPool           *TransactionPool

	// Storage (MDBX-inspired)
	mdbxDB      *MDBXDatabase
	stateReader *StateReader
	stateWriter *StateWriter

	// Parallel execution management
	executorPool    *ExecutorPool
	dependencyGraph *DependencyGraph

	// Configuration
	config *EVMConfig

	// Runtime state
	mu      sync.RWMutex
	running bool

	// Metrics
	metrics *EVMMetrics
}

// EVMConfig holds EVM configuration
type EVMConfig struct {
	ChainID       uint64 `json:"chain_id"`
	NetworkID     uint64 `json:"network_id"`
	BlockGasLimit uint64 `json:"block_gas_limit"`
	BaseFee       uint64 `json:"base_fee"`

	// Erigon-inspired settings
	ParallelWorkers    int  `json:"parallel_workers"`
	EnableParallelExec bool `json:"enable_parallel_exec"`
	MDBXPageSize       int  `json:"mdbx_page_size"`
	MemoryLimit        int  `json:"memory_limit_mb"`

	// Stage execution settings
	EnableStages    bool `json:"enable_stages"`
	StageBufferSize int  `json:"stage_buffer_size"`
}

// ParallelExecutor manages parallel transaction execution
type ParallelExecutor struct {
	workers       []*ExecutionWorker
	taskQueue     chan *ExecutionTask
	resultQueue   chan *ExecutionResult
	dependencyMgr *DependencyManager

	workerCount    int
	maxConcurrency int
}

// StageSequence implements Erigon's stage-based processing
type StageSequence struct {
	stages       []Stage
	currentStage int
	stageData    map[string]interface{}
}

// Stage represents a processing stage (Erigon-inspired)
type Stage interface {
	Name() string
	Execute(ctx context.Context, batch *StageBatch) error
	Rollback(ctx context.Context, batch *StageBatch) error
}

// Transaction execution stages (based on Erigon)
const (
	StageHeaders             = "Headers"
	StageBodies              = "Bodies"
	StageSenders             = "Senders"
	StageExecution           = "Execution"
	StageHashState           = "HashState"
	StageIntermediateHashes  = "IntermediateHashes"
	StageCallTraces          = "CallTraces"
	StageAccountHistoryIndex = "AccountHistoryIndex"
	StageStorageHistoryIndex = "StorageHistoryIndex"
	StageLogIndex            = "LogIndex"
	StageTxLookup            = "TxLookup"
	StageFinish              = "Finish"
)

// ExecutionTask represents a transaction to be executed
type ExecutionTask struct {
	Transaction  *types.Transaction
	Header       *types.Header
	StateDB      *state.StateDB
	Index        int
	Dependencies []int // Indices of transactions this depends on
}

// ExecutionResult contains the result of transaction execution
type ExecutionResult struct {
	Receipt   *types.Receipt
	StateRoot common.Hash
	GasUsed   uint64
	Index     int
	Error     error

	// State changes for conflict detection
	StateChanges *StateChangeSet
}

// NewEVMEngine creates a new Erigon-inspired EVM engine
func NewEVMEngine(config *EVMConfig) (*EVMEngine, error) {
	// Initialize chain configuration
	chainConfig := &params.ChainConfig{
		ChainID:             new(big.Int).SetUint64(config.ChainID),
		HomesteadBlock:      common.Big0,
		DAOForkBlock:        nil,
		DAOForkSupport:      false,
		EIP150Block:         common.Big0,
		EIP155Block:         common.Big0,
		EIP158Block:         common.Big0,
		ByzantiumBlock:      common.Big0,
		ConstantinopleBlock: common.Big0,
		PetersburgBlock:     common.Big0,
		IstanbulBlock:       common.Big0,
		BerlinBlock:         common.Big0,
		LondonBlock:         common.Big0,
		ArrowGlacierBlock:   common.Big0,
		GrayGlacierBlock:    common.Big0,
		MergeNetsplitBlock:  common.Big0,
		ShanghaiTime:        new(uint64),
		CancunTime:          new(uint64),
	}

	// Initialize VM configuration for optimized execution
	vmConfig := &vm.Config{
		EnablePreimageRecording: false, // Disabled for performance
		Tracer:                  nil,
		NoBaseFee:               false,
	}

	// Initialize MDBX database (Erigon's storage engine)
	mdbxDB, err := NewMDBXDatabase(&MDBXConfig{
		PageSize:    config.MDBXPageSize,
		MemoryLimit: config.MemoryLimit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize MDBX database: %w", err)
	}

	// Initialize parallel executor
	parallelExecutor, err := NewParallelExecutor(&ParallelConfig{
		Workers:        config.ParallelWorkers,
		MaxConcurrency: config.ParallelWorkers * 2,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create parallel executor: %w", err)
	}

	// Initialize stage sequence (Erigon's staged sync)
	stageSequence := NewStageSequence()

	// Initialize transaction pool
	txPool, err := NewTransactionPool(&TxPoolConfig{
		BufferSize: config.StageBufferSize,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction pool: %w", err)
	}

	// Initialize state reader/writer
	stateReader := NewStateReader(mdbxDB)
	stateWriter := NewStateWriter(mdbxDB)

	engine := &EVMEngine{
		chainConfig:      chainConfig,
		vmConfig:         vmConfig,
		parallelExecutor: parallelExecutor,
		stageSequence:    stageSequence,
		txPool:           txPool,
		mdbxDB:           mdbxDB,
		stateReader:      stateReader,
		stateWriter:      stateWriter,
		config:           config,
		metrics:          NewEVMMetrics(),
	}

	return engine, nil
}

// Start begins the EVM engine
func (e *EVMEngine) Start(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.running {
		return fmt.Errorf("EVM engine already running")
	}

	fmt.Println("🔥 Starting Erigon-inspired EVM Engine")
	fmt.Printf("   • Parallel Workers: %d\n", e.config.ParallelWorkers)
	fmt.Printf("   • MDBX Page Size: %d KB\n", e.config.MDBXPageSize/1024)
	fmt.Printf("   • Memory Limit: %d MB\n", e.config.MemoryLimit)
	fmt.Printf("   • Parallel Execution: %v\n", e.config.EnableParallelExec)

	// Start MDBX database
	if err := e.mdbxDB.Start(ctx); err != nil {
		return fmt.Errorf("failed to start MDBX database: %w", err)
	}

	// Start parallel executor
	if err := e.parallelExecutor.Start(ctx); err != nil {
		return fmt.Errorf("failed to start parallel executor: %w", err)
	}

	// Start transaction pool
	if err := e.txPool.Start(ctx); err != nil {
		return fmt.Errorf("failed to start transaction pool: %w", err)
	}

	e.running = true

	fmt.Println("✅ Erigon-inspired EVM Engine started")
	return nil
}

// ExecuteTransactions executes a batch of transactions with Erigon-style parallel processing
func (e *EVMEngine) ExecuteTransactions(header *types.Header, txs types.Transactions) ([]*types.Receipt, error) {
	if !e.config.EnableParallelExec || len(txs) < 4 {
		// Fall back to sequential execution for small batches
		return e.executeSequential(header, txs)
	}

	// Use parallel execution for larger batches (Erigon's approach)
	return e.executeParallel(header, txs)
}

// executeParallel performs parallel transaction execution (Erigon-inspired)
func (e *EVMEngine) executeParallel(header *types.Header, txs types.Transactions) ([]*types.Receipt, error) {
	startTime := time.Now()

	fmt.Printf("🚀 Parallel execution: %d transactions across %d workers\n",
		len(txs), e.config.ParallelWorkers)

	// Step 1: Analyze dependencies (Erigon's dependency analysis)
	dependencies, err := e.analyzeDependencies(txs)
	if err != nil {
		return nil, fmt.Errorf("dependency analysis failed: %w", err)
	}

	// Step 2: Create execution tasks
	tasks := make([]*ExecutionTask, len(txs))
	for i, tx := range txs {
		stateDB, err := e.stateReader.GetState(header.Root)
		if err != nil {
			return nil, fmt.Errorf("failed to get state for tx %d: %w", i, err)
		}

		tasks[i] = &ExecutionTask{
			Transaction:  tx,
			Header:       header,
			StateDB:      stateDB.Copy(),
			Index:        i,
			Dependencies: dependencies[i],
		}
	}

	// Step 3: Execute in parallel waves based on dependencies
	receipts := make([]*types.Receipt, len(txs))
	executed := make([]bool, len(txs))

	for wave := 0; len(executed) < len(txs); wave++ {
		// Find transactions ready to execute (no unresolved dependencies)
		readyTasks := []*ExecutionTask{}
		for i, task := range tasks {
			if executed[i] {
				continue
			}

			ready := true
			for _, depIndex := range task.Dependencies {
				if !executed[depIndex] {
					ready = false
					break
				}
			}

			if ready {
				readyTasks = append(readyTasks, task)
			}
		}

		if len(readyTasks) == 0 {
			return nil, fmt.Errorf("circular dependency detected in transaction batch")
		}

		// Execute ready transactions in parallel
		results, err := e.parallelExecutor.ExecuteBatch(readyTasks)
		if err != nil {
			return nil, fmt.Errorf("parallel execution failed in wave %d: %w", wave, err)
		}

		// Process results
		for _, result := range results {
			if result.Error != nil {
				return nil, fmt.Errorf("transaction %d failed: %w", result.Index, result.Error)
			}
			receipts[result.Index] = result.Receipt
			executed[result.Index] = true
		}

		fmt.Printf("⚡ Wave %d: executed %d transactions\n", wave, len(readyTasks))
	}

	duration := time.Since(startTime)
	e.metrics.RecordParallelExecution(len(txs), duration)

	fmt.Printf("✅ Parallel execution completed: %d txs in %v (%.2f TPS)\n",
		len(txs), duration, float64(len(txs))/duration.Seconds())

	return receipts, nil
}

// executeSequential performs sequential transaction execution
func (e *EVMEngine) executeSequential(header *types.Header, txs types.Transactions) ([]*types.Receipt, error) {
	receipts := make([]*types.Receipt, 0, len(txs))

	// Get initial state
	stateDB, err := e.stateReader.GetState(header.Root)
	if err != nil {
		return nil, fmt.Errorf("failed to get initial state: %w", err)
	}

	// Create EVM instance
	blockContext := core.NewEVMBlockContext(header, nil, nil)
	evm := vm.NewEVM(blockContext, vm.TxContext{}, stateDB, e.chainConfig, *e.vmConfig)

	gasUsed := uint64(0)

	// Execute transactions sequentially
	for i, tx := range txs {
		// Create transaction context
		txContext := core.NewEVMTxContext(tx)
		evm.Reset(txContext, stateDB)

		// Apply transaction
		receipt, err := e.applyTransaction(evm, stateDB, header, tx, &gasUsed, i)
		if err != nil {
			return nil, fmt.Errorf("transaction %d failed: %w", i, err)
		}

		receipts = append(receipts, receipt)
	}

	return receipts, nil
}

// analyzeDependencies analyzes transaction dependencies for parallel execution
func (e *EVMEngine) analyzeDependencies(txs types.Transactions) ([][]int, error) {
	dependencies := make([][]int, len(txs))
	addressAccess := make(map[common.Address][]int) // address -> transaction indices

	for i, tx := range txs {
		deps := []int{}

		// Check for dependencies based on address access
		from, err := types.Sender(types.LatestSignerForChainID(e.chainConfig.ChainID), tx)
		if err != nil {
			return nil, fmt.Errorf("failed to get sender for tx %d: %w", i, err)
		}

		// Add dependency on previous transactions that accessed the same addresses
		addresses := []common.Address{from}
		if tx.To() != nil {
			addresses = append(addresses, *tx.To())
		}

		for _, addr := range addresses {
			if prevIndices, exists := addressAccess[addr]; exists {
				// Add dependency on the most recent access
				if len(prevIndices) > 0 {
					deps = append(deps, prevIndices[len(prevIndices)-1])
				}
			}
			addressAccess[addr] = append(addressAccess[addr], i)
		}

		dependencies[i] = deps
	}

	return dependencies, nil
}

// applyTransaction applies a single transaction
func (e *EVMEngine) applyTransaction(evm *vm.EVM, stateDB *state.StateDB, header *types.Header,
	tx *types.Transaction, gasUsed *uint64, txIndex int) (*types.Receipt, error) {

	// Get transaction sender
	from, err := types.Sender(types.LatestSignerForChainID(e.chainConfig.ChainID), tx)
	if err != nil {
		return nil, err
	}

	// Check nonce
	if stateDB.GetNonce(from) != tx.Nonce() {
		return nil, fmt.Errorf("invalid nonce")
	}

	// Check balance
	gasPrice := tx.GasPrice()
	if gasPrice == nil {
		gasPrice = header.BaseFee
	}
	if stateDB.GetBalance(from).Cmp(new(big.Int).Mul(gasPrice, new(big.Int).SetUint64(tx.Gas()))) < 0 {
		return nil, fmt.Errorf("insufficient balance")
	}

	// Create snapshot for rollback
	snapshot := stateDB.Snapshot()

	// Apply transaction
	result, err := core.ApplyMessage(evm, types.NewMessage(
		from,
		tx.To(),
		tx.Nonce(),
		tx.Value(),
		tx.Gas(),
		gasPrice,
		gasPrice,
		gasPrice,
		tx.Data(),
		tx.AccessList(),
		false,
	), new(core.GasPool).AddGas(header.GasLimit))

	if err != nil {
		stateDB.RevertToSnapshot(snapshot)
		return nil, err
	}

	// Update gas used
	*gasUsed += result.UsedGas

	// Create receipt
	receipt := &types.Receipt{
		Type:              tx.Type(),
		PostState:         stateDB.IntermediateRoot(true).Bytes(),
		CumulativeGasUsed: *gasUsed,
		TxHash:            tx.Hash(),
		GasUsed:           result.UsedGas,
		BlockHash:         header.Hash(),
		BlockNumber:       header.Number,
		TransactionIndex:  uint(txIndex),
	}

	// Add logs
	receipt.Logs = stateDB.GetLogs(tx.Hash(), header.Hash())
	receipt.Bloom = types.CreateBloom(types.Receipts{receipt})

	return receipt, nil
}

// GetPendingTransactions returns pending transactions from the pool
func (e *EVMEngine) GetPendingTransactions() types.Transactions {
	return e.txPool.GetPending()
}

// SubmitTransaction submits a transaction to the pool
func (e *EVMEngine) SubmitTransaction(tx *types.Transaction) error {
	return e.txPool.AddLocal(tx)
}

// Stop stops the EVM engine
func (e *EVMEngine) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.running {
		return nil
	}

	fmt.Println("🛑 Stopping Erigon-inspired EVM Engine...")

	// Stop components
	e.txPool.Stop()
	e.parallelExecutor.Stop()
	e.mdbxDB.Stop()

	e.running = false

	fmt.Println("✅ EVM engine stopped")
	return nil
}
