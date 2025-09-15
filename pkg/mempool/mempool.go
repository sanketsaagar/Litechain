package mempool

import (
	"fmt"
	"math/big"
	"sort"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// MemPool manages pending transactions with parallel execution support
// Innovation: Priority-based transaction ordering with dependency analysis
type MemPool struct {
	// Transaction storage
	pending map[common.Hash]*PoolTransaction
	queued  map[common.Address][]*PoolTransaction
	all     map[common.Hash]*PoolTransaction

	// Parallel execution support
	dependencyGraph *DependencyGraph
	parallelBatches [][]*PoolTransaction

	// Configuration
	config *MemPoolConfig

	// Metrics
	stats *MemPoolStats

	// Synchronization
	mu      sync.RWMutex
	priceMu sync.RWMutex

	// Price tracking
	priceList *PriceList
	gasPrice  *big.Int

	// Events
	newTxsCh chan *types.Transaction

	// Parallel execution workers
	workers   []*ExecutionWorker
	workQueue chan *ExecutionBatch
}

// PoolTransaction wraps a transaction with additional metadata
type PoolTransaction struct {
	Tx       *types.Transaction
	Hash     common.Hash
	Cost     *big.Int
	From     common.Address
	GasPrice *big.Int
	GasLimit uint64
	Nonce    uint64
	AddedAt  time.Time
	Priority int64

	// Parallel execution metadata
	ReadSet      map[common.Address][]common.Hash // Accounts/storage read
	WriteSet     map[common.Address][]common.Hash // Accounts/storage written
	Dependencies []*PoolTransaction               // Transaction dependencies
	CanParallel  bool                             // Can execute in parallel
}

// DependencyGraph analyzes transaction dependencies for parallel execution
type DependencyGraph struct {
	nodes map[common.Hash]*GraphNode
	edges map[common.Hash][]common.Hash
	mu    sync.RWMutex
}

// GraphNode represents a transaction in the dependency graph
type GraphNode struct {
	Tx           *PoolTransaction
	Dependencies []common.Hash
	Dependents   []common.Hash
	Level        int // Execution level (0 = no dependencies)
}

// ExecutionBatch represents a batch of transactions that can execute in parallel
type ExecutionBatch struct {
	Transactions []*PoolTransaction
	Level        int
	Timestamp    time.Time
}

// ExecutionWorker processes transaction batches in parallel
type ExecutionWorker struct {
	ID       int
	workCh   chan *ExecutionBatch
	resultCh chan *ExecutionResult
	stopCh   chan struct{}
}

// ExecutionResult contains the result of parallel execution
type ExecutionResult struct {
	Batch    *ExecutionBatch
	Results  []*TransactionResult
	Error    error
	WorkerID int
	Duration time.Duration
}

// TransactionResult contains individual transaction execution result
type TransactionResult struct {
	Tx      *PoolTransaction
	Receipt *types.Receipt
	GasUsed uint64
	Error   error
}

// MemPoolConfig contains mempool configuration
type MemPoolConfig struct {
	GlobalSlots  uint64        // Maximum number of transactions
	GlobalQueue  uint64        // Maximum queued transactions per account
	AccountSlots uint64        // Maximum transactions per account
	AccountQueue uint64        // Maximum queued transactions per account
	Lifetime     time.Duration // Maximum transaction lifetime
	PriceBump    uint64        // Minimum price increase for replacement

	// Parallel execution config
	WorkerCount    int           // Number of parallel execution workers
	BatchSize      int           // Target batch size for parallel execution
	DepthLimit     int           // Maximum dependency depth
	ConflictWindow time.Duration // Time window for conflict detection
}

// MemPoolStats tracks mempool performance metrics
type MemPoolStats struct {
	TotalTxs         uint64
	PendingTxs       uint64
	QueuedTxs        uint64
	ParallelBatches  uint64
	ConflictResolved uint64
	AvgBatchSize     float64
	AvgExecutionTime time.Duration
	ThroughputTPS    float64

	mu sync.RWMutex
}

// PriceList maintains transactions sorted by gas price
type PriceList struct {
	items []*PoolTransaction
	mu    sync.RWMutex
}

// Default configuration optimized for high throughput
func DefaultMemPoolConfig() *MemPoolConfig {
	return &MemPoolConfig{
		GlobalSlots:  100000, // 100K pending transactions
		GlobalQueue:  50000,  // 50K queued transactions
		AccountSlots: 1000,   // 1K transactions per account
		AccountQueue: 500,    // 500 queued per account
		Lifetime:     30 * time.Minute,
		PriceBump:    10, // 10% price increase for replacement

		// Parallel execution - optimized for performance
		WorkerCount:    8,   // 8 parallel workers
		BatchSize:      100, // 100 transactions per batch
		DepthLimit:     10,  // Maximum dependency depth
		ConflictWindow: 1 * time.Second,
	}
}

// NewMemPool creates a new high-performance mempool
func NewMemPool(config *MemPoolConfig) *MemPool {
	if config == nil {
		config = DefaultMemPoolConfig()
	}

	mp := &MemPool{
		pending:         make(map[common.Hash]*PoolTransaction),
		queued:          make(map[common.Address][]*PoolTransaction),
		all:             make(map[common.Hash]*PoolTransaction),
		dependencyGraph: NewDependencyGraph(),
		config:          config,
		stats:           &MemPoolStats{},
		priceList:       &PriceList{},
		gasPrice:        big.NewInt(1000000000), // 1 Gwei default
		newTxsCh:        make(chan *types.Transaction, 1000),
		workers:         make([]*ExecutionWorker, config.WorkerCount),
		workQueue:       make(chan *ExecutionBatch, 100),
	}

	// Initialize parallel execution workers
	mp.initializeWorkers()

	return mp
}

// AddTransaction adds a transaction to the mempool with dependency analysis
func (mp *MemPool) AddTransaction(tx *types.Transaction) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	// Basic validation
	if tx == nil {
		return fmt.Errorf("nil transaction")
	}

	hash := tx.Hash()
	if _, exists := mp.all[hash]; exists {
		return fmt.Errorf("transaction already exists")
	}

	// Create pool transaction with metadata
	poolTx, err := mp.createPoolTransaction(tx)
	if err != nil {
		return fmt.Errorf("failed to create pool transaction: %w", err)
	}

	// Analyze dependencies for parallel execution
	mp.analyzeDependencies(poolTx)

	// Add to mempool
	mp.all[hash] = poolTx
	mp.pending[hash] = poolTx

	// Update dependency graph
	mp.dependencyGraph.AddTransaction(poolTx)

	// Update statistics
	mp.updateStats()

	// Trigger parallel batch creation
	go mp.createParallelBatches()

	return nil
}

// createPoolTransaction creates a PoolTransaction with metadata
func (mp *MemPool) createPoolTransaction(tx *types.Transaction) (*PoolTransaction, error) {
	// Extract sender (requires signature validation)
	from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
	if err != nil {
		return nil, fmt.Errorf("invalid signature: %w", err)
	}

	// Analyze transaction for parallel execution
	readSet, writeSet := mp.analyzeTransactionAccess(tx)

	poolTx := &PoolTransaction{
		Tx:          tx,
		Hash:        tx.Hash(),
		Cost:        tx.Cost(),
		From:        from,
		GasPrice:    tx.GasPrice(),
		GasLimit:    tx.Gas(),
		Nonce:       tx.Nonce(),
		AddedAt:     time.Now(),
		Priority:    mp.calculatePriority(tx),
		ReadSet:     readSet,
		WriteSet:    writeSet,
		CanParallel: mp.canExecuteInParallel(readSet, writeSet),
	}

	return poolTx, nil
}

// analyzeTransactionAccess analyzes what accounts/storage a transaction accesses
func (mp *MemPool) analyzeTransactionAccess(tx *types.Transaction) (map[common.Address][]common.Hash, map[common.Address][]common.Hash) {
	readSet := make(map[common.Address][]common.Hash)
	writeSet := make(map[common.Address][]common.Hash)

	// Basic analysis - in a full implementation, this would use static analysis
	// or execution simulation to determine exact access patterns

	// Always reads from sender
	from, _ := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
	readSet[from] = []common.Hash{}  // Balance read
	writeSet[from] = []common.Hash{} // Balance/nonce write

	// If transaction has a recipient
	if tx.To() != nil {
		to := *tx.To()
		readSet[to] = []common.Hash{}  // Balance read
		writeSet[to] = []common.Hash{} // Balance write

		// For contract calls, this would analyze the contract code
		// to determine storage access patterns
		if len(tx.Data()) > 0 {
			// Contract call - analyze bytecode (simplified)
			// In reality, this would require EVM simulation or static analysis
			storageSlot := common.BytesToHash(tx.Data()[:32])
			readSet[to] = append(readSet[to], storageSlot)
			writeSet[to] = append(writeSet[to], storageSlot)
		}
	}

	return readSet, writeSet
}

// canExecuteInParallel determines if a transaction can be executed in parallel
func (mp *MemPool) canExecuteInParallel(readSet, writeSet map[common.Address][]common.Hash) bool {
	// Simple heuristic: transactions with minimal state access can parallel
	totalAccesses := 0
	for _, slots := range readSet {
		totalAccesses += len(slots)
	}
	for _, slots := range writeSet {
		totalAccesses += len(slots)
	}

	// Transactions with few state accesses are good candidates for parallelization
	return totalAccesses <= 5
}

// analyzeDependencies analyzes transaction dependencies
func (mp *MemPool) analyzeDependencies(tx *PoolTransaction) {
	// Check for conflicts with existing pending transactions
	for _, existingTx := range mp.pending {
		if mp.hasConflict(tx, existingTx) {
			tx.Dependencies = append(tx.Dependencies, existingTx)
			tx.CanParallel = false
		}
	}
}

// hasConflict checks if two transactions have conflicting access patterns
func (mp *MemPool) hasConflict(tx1, tx2 *PoolTransaction) bool {
	// Same sender - must be sequential (nonce ordering)
	if tx1.From == tx2.From {
		return true
	}

	// Check for read-write or write-write conflicts
	for addr, writeSlots := range tx1.WriteSet {
		// Write-Write conflict
		if otherWrites, exists := tx2.WriteSet[addr]; exists {
			if mp.slotsOverlap(writeSlots, otherWrites) {
				return true
			}
		}

		// Write-Read conflict
		if otherReads, exists := tx2.ReadSet[addr]; exists {
			if mp.slotsOverlap(writeSlots, otherReads) {
				return true
			}
		}
	}

	// Check reverse direction
	for addr, writeSlots := range tx2.WriteSet {
		if otherReads, exists := tx1.ReadSet[addr]; exists {
			if mp.slotsOverlap(writeSlots, otherReads) {
				return true
			}
		}
	}

	return false
}

// slotsOverlap checks if two storage slot lists overlap
func (mp *MemPool) slotsOverlap(slots1, slots2 []common.Hash) bool {
	if len(slots1) == 0 || len(slots2) == 0 {
		return len(slots1) == 0 && len(slots2) == 0 // Both empty = both access balance
	}

	slotMap := make(map[common.Hash]bool)
	for _, slot := range slots1 {
		slotMap[slot] = true
	}

	for _, slot := range slots2 {
		if slotMap[slot] {
			return true
		}
	}

	return false
}

// calculatePriority calculates transaction priority
func (mp *MemPool) calculatePriority(tx *types.Transaction) int64 {
	// Priority based on gas price and gas limit
	gasPrice := tx.GasPrice().Int64()
	gasLimit := int64(tx.Gas())

	// Higher gas price = higher priority
	// Bonus for transactions that can execute in parallel
	priority := gasPrice * gasLimit / 1000000 // Normalize

	return priority
}

// createParallelBatches creates batches of transactions that can execute in parallel
func (mp *MemPool) createParallelBatches() {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	// Get all pending transactions sorted by dependency level
	levels := mp.dependencyGraph.GetExecutionLevels()

	mp.parallelBatches = make([][]*PoolTransaction, len(levels))

	for level, txs := range levels {
		// Sort transactions within level by priority
		sort.Slice(txs, func(i, j int) bool {
			return txs[i].Priority > txs[j].Priority
		})

		// Create batches of optimal size
		batchSize := mp.config.BatchSize
		for i := 0; i < len(txs); i += batchSize {
			end := i + batchSize
			if end > len(txs) {
				end = len(txs)
			}

			batch := txs[i:end]
			mp.parallelBatches[level] = append(mp.parallelBatches[level], batch...)
		}
	}

	// Send batches to workers
	go mp.dispatchBatches()
}

// dispatchBatches sends transaction batches to parallel workers
func (mp *MemPool) dispatchBatches() {
	for level, txs := range mp.parallelBatches {
		if len(txs) == 0 {
			continue
		}

		batch := &ExecutionBatch{
			Transactions: txs,
			Level:        level,
			Timestamp:    time.Now(),
		}

		select {
		case mp.workQueue <- batch:
			mp.stats.mu.Lock()
			mp.stats.ParallelBatches++
			mp.stats.mu.Unlock()
		default:
			// Work queue full, skip this batch
		}
	}
}

// GetParallelBatches returns transaction batches ready for parallel execution
func (mp *MemPool) GetParallelBatches() [][]*PoolTransaction {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	// Return copy of parallel batches
	batches := make([][]*PoolTransaction, len(mp.parallelBatches))
	for i, batch := range mp.parallelBatches {
		batches[i] = make([]*PoolTransaction, len(batch))
		copy(batches[i], batch)
	}

	return batches
}

// RemoveTransaction removes a transaction from the mempool
func (mp *MemPool) RemoveTransaction(hash common.Hash) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	if tx, exists := mp.all[hash]; exists {
		delete(mp.all, hash)
		delete(mp.pending, hash)

		// Remove from dependency graph
		mp.dependencyGraph.RemoveTransaction(hash)

		// Remove from queued transactions
		if queued, exists := mp.queued[tx.From]; exists {
			for i, queuedTx := range queued {
				if queuedTx.Hash == hash {
					mp.queued[tx.From] = append(queued[:i], queued[i+1:]...)
					break
				}
			}
		}

		mp.updateStats()
	}
}

// GetStats returns mempool statistics
func (mp *MemPool) GetStats() *MemPoolStats {
	mp.stats.mu.RLock()
	defer mp.stats.mu.RUnlock()

	// Return copy
	return &MemPoolStats{
		TotalTxs:         mp.stats.TotalTxs,
		PendingTxs:       mp.stats.PendingTxs,
		QueuedTxs:        mp.stats.QueuedTxs,
		ParallelBatches:  mp.stats.ParallelBatches,
		ConflictResolved: mp.stats.ConflictResolved,
		AvgBatchSize:     mp.stats.AvgBatchSize,
		AvgExecutionTime: mp.stats.AvgExecutionTime,
		ThroughputTPS:    mp.stats.ThroughputTPS,
	}
}

// updateStats updates mempool statistics
func (mp *MemPool) updateStats() {
	mp.stats.mu.Lock()
	defer mp.stats.mu.Unlock()

	mp.stats.TotalTxs = uint64(len(mp.all))
	mp.stats.PendingTxs = uint64(len(mp.pending))

	queuedCount := uint64(0)
	for _, queue := range mp.queued {
		queuedCount += uint64(len(queue))
	}
	mp.stats.QueuedTxs = queuedCount

	// Calculate average batch size
	if mp.stats.ParallelBatches > 0 {
		mp.stats.AvgBatchSize = float64(mp.stats.TotalTxs) / float64(mp.stats.ParallelBatches)
	}
}

// initializeWorkers initializes parallel execution workers
func (mp *MemPool) initializeWorkers() {
	for i := 0; i < mp.config.WorkerCount; i++ {
		worker := &ExecutionWorker{
			ID:       i,
			workCh:   make(chan *ExecutionBatch, 10),
			resultCh: make(chan *ExecutionResult, 10),
			stopCh:   make(chan struct{}),
		}

		mp.workers[i] = worker

		// Start worker goroutine
		go mp.runWorker(worker)
	}
}

// runWorker runs a parallel execution worker
func (mp *MemPool) runWorker(worker *ExecutionWorker) {
	for {
		select {
		case batch := <-mp.workQueue:
			start := time.Now()

			// Process batch in parallel
			results := make([]*TransactionResult, len(batch.Transactions))
			for i, tx := range batch.Transactions {
				// Simulate transaction execution
				result := &TransactionResult{
					Tx:      tx,
					GasUsed: tx.GasLimit,
					Error:   nil,
				}
				results[i] = result
			}

			// Send result
			execResult := &ExecutionResult{
				Batch:    batch,
				Results:  results,
				WorkerID: worker.ID,
				Duration: time.Since(start),
			}

			select {
			case worker.resultCh <- execResult:
			default:
				// Result channel full
			}

		case <-worker.stopCh:
			return
		}
	}
}

// Dependency Graph Implementation

// NewDependencyGraph creates a new dependency graph
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		nodes: make(map[common.Hash]*GraphNode),
		edges: make(map[common.Hash][]common.Hash),
	}
}

// AddTransaction adds a transaction to the dependency graph
func (dg *DependencyGraph) AddTransaction(tx *PoolTransaction) {
	dg.mu.Lock()
	defer dg.mu.Unlock()

	node := &GraphNode{
		Tx:           tx,
		Dependencies: make([]common.Hash, 0),
		Dependents:   make([]common.Hash, 0),
		Level:        0,
	}

	// Add dependencies
	for _, dep := range tx.Dependencies {
		node.Dependencies = append(node.Dependencies, dep.Hash)
		dg.edges[dep.Hash] = append(dg.edges[dep.Hash], tx.Hash)

		if depNode, exists := dg.nodes[dep.Hash]; exists {
			depNode.Dependents = append(depNode.Dependents, tx.Hash)
		}
	}

	dg.nodes[tx.Hash] = node
	dg.calculateLevels()
}

// RemoveTransaction removes a transaction from the dependency graph
func (dg *DependencyGraph) RemoveTransaction(hash common.Hash) {
	dg.mu.Lock()
	defer dg.mu.Unlock()

	if node, exists := dg.nodes[hash]; exists {
		// Remove from dependents
		for _, depHash := range node.Dependencies {
			if depNode, exists := dg.nodes[depHash]; exists {
				for i, dependent := range depNode.Dependents {
					if dependent == hash {
						depNode.Dependents = append(depNode.Dependents[:i], depNode.Dependents[i+1:]...)
						break
					}
				}
			}
		}

		delete(dg.nodes, hash)
		delete(dg.edges, hash)
	}
}

// calculateLevels calculates execution levels for transactions
func (dg *DependencyGraph) calculateLevels() {
	// Topological sort to assign levels
	visited := make(map[common.Hash]bool)

	var dfs func(hash common.Hash) int
	dfs = func(hash common.Hash) int {
		if visited[hash] {
			return dg.nodes[hash].Level
		}

		visited[hash] = true
		maxDepLevel := -1

		if node, exists := dg.nodes[hash]; exists {
			for _, depHash := range node.Dependencies {
				depLevel := dfs(depHash)
				if depLevel > maxDepLevel {
					maxDepLevel = depLevel
				}
			}

			node.Level = maxDepLevel + 1
		}

		return dg.nodes[hash].Level
	}

	for hash := range dg.nodes {
		dfs(hash)
	}
}

// GetExecutionLevels returns transactions grouped by execution level
func (dg *DependencyGraph) GetExecutionLevels() map[int][]*PoolTransaction {
	dg.mu.RLock()
	defer dg.mu.RUnlock()

	levels := make(map[int][]*PoolTransaction)

	for _, node := range dg.nodes {
		level := node.Level
		levels[level] = append(levels[level], node.Tx)
	}

	return levels
}

// Price List Implementation

func (pl *PriceList) Len() int { return len(pl.items) }

func (pl *PriceList) Less(i, j int) bool {
	return pl.items[i].GasPrice.Cmp(pl.items[j].GasPrice) > 0
}

func (pl *PriceList) Swap(i, j int) {
	pl.items[i], pl.items[j] = pl.items[j], pl.items[i]
}

func (pl *PriceList) Push(x interface{}) {
	pl.items = append(pl.items, x.(*PoolTransaction))
}

func (pl *PriceList) Pop() interface{} {
	old := pl.items
	n := len(old)
	item := old[n-1]
	pl.items = old[0 : n-1]
	return item
}
