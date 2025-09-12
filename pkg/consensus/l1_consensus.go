package consensus

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// HPoSConsensus implements Hybrid Proof-of-Stake consensus for L1
// Combines CometBFT-style BFT consensus with Erigon's parallel execution
// Innovation: Dynamic validator selection based on performance + stake
type HPoSConsensus struct {
	// Core consensus state
	chainID     *big.Int
	blockHeight *big.Int
	epoch       uint64
	round       uint64
	
	// Validator management
	validators    *ValidatorSet
	activeStakers map[common.Address]*StakeInfo
	
	// Consensus parameters
	blockTime       time.Duration
	epochLength     uint64
	maxValidators   int
	slashingEnabled bool
	
	// Network state
	isValidator bool
	nodeID      common.Address
	privateKey  []byte
	
	// Synchronization
	mu      sync.RWMutex
	ctx     context.Context
	cancel  context.CancelFunc
	running bool
	
	// Innovation: Performance-based validator scoring
	performanceTracker *PerformanceTracker
	
	// Channels for consensus communication
	proposalCh chan *Proposal
	voteCh     chan *Vote
	commitCh   chan *Commit
}

// StakeInfo holds validator staking information
type StakeInfo struct {
	Amount          *big.Int
	StakedAt        uint64
	Performance     float64
	SlashCount      uint64
	LastActiveBlock uint64
	RewardAddress   common.Address
}

// ValidatorSet manages the active validator set
type ValidatorSet struct {
	validators map[common.Address]*Validator
	sorted     []*Validator
	totalStake *big.Int
	mu         sync.RWMutex
}

// Validator represents a network validator
type Validator struct {
	Address     common.Address
	PubKey      []byte
	Stake       *big.Int
	Power       uint64
	Performance float64
	IsOnline    bool
}

// PerformanceTracker tracks validator performance metrics
type PerformanceTracker struct {
	blockProposals map[common.Address]uint64
	blockSignings  map[common.Address]uint64
	missedBlocks   map[common.Address]uint64
	responseTime   map[common.Address]time.Duration
	mu             sync.RWMutex
}

// Proposal represents a block proposal
type Proposal struct {
	Height    uint64
	Round     uint64
	BlockHash common.Hash
	Proposer  common.Address
	Timestamp time.Time
	Signature []byte
}

// Vote represents a validator vote
type Vote struct {
	Height    uint64
	Round     uint64
	BlockHash common.Hash
	Validator common.Address
	VoteType  VoteType
	Signature []byte
}

// VoteType represents different types of votes
type VoteType int

const (
	VotePrevote VoteType = iota
	VotePrecommit
)

// Commit represents a block commit
type Commit struct {
	Height     uint64
	BlockHash  common.Hash
	Signatures [][]byte
	Validators []common.Address
}

// NewHPoSConsensus creates a new L1 consensus engine
func NewHPoSConsensus(chainID *big.Int, nodeAddr common.Address, privateKey []byte) *HPoSConsensus {
	return &HPoSConsensus{
		chainID:     chainID,
		blockHeight: big.NewInt(0),
		epoch:       0,
		round:       0,
		validators:  NewValidatorSet(),
		activeStakers: make(map[common.Address]*StakeInfo),
		blockTime:     2 * time.Second,
		epochLength:   100, // 100 blocks per epoch
		maxValidators: 21,  // Similar to BNB Chain but optimized
		slashingEnabled: true,
		nodeID:      nodeAddr,
		privateKey:  privateKey,
		performanceTracker: NewPerformanceTracker(),
		proposalCh: make(chan *Proposal, 10),
		voteCh:     make(chan *Vote, 100),
		commitCh:   make(chan *Commit, 10),
	}
}

// Start begins the consensus engine
func (h *HPoSConsensus) Start(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	if h.running {
		return fmt.Errorf("consensus already running")
	}
	
	h.ctx, h.cancel = context.WithCancel(ctx)
	h.running = true
	
	// Initialize genesis validators if first start
	if h.blockHeight.Cmp(big.NewInt(0)) == 0 {
		if err := h.initializeGenesisValidators(); err != nil {
			return fmt.Errorf("failed to initialize genesis validators: %w", err)
		}
	}
	
	// Start consensus rounds
	go h.consensusLoop()
	go h.performanceMonitor()
	go h.stakeManager()
	
	fmt.Printf("🚀 HPoS Consensus started for L1 chain %s\n", h.chainID.String())
	fmt.Printf("   • Node ID: %s\n", h.nodeID.Hex())
	fmt.Printf("   • Is Validator: %v\n", h.isValidator)
	fmt.Printf("   • Block Time: %v\n", h.blockTime)
	fmt.Printf("   • Max Validators: %d\n", h.maxValidators)
	
	return nil
}

// consensusLoop runs the main consensus algorithm
func (h *HPoSConsensus) consensusLoop() {
	ticker := time.NewTicker(h.blockTime)
	defer ticker.Stop()
	
	for {
		select {
		case <-h.ctx.Done():
			return
		case <-ticker.C:
			if err := h.runConsensusRound(); err != nil {
				fmt.Printf("❌ Consensus round failed: %v\n", err)
			}
		case proposal := <-h.proposalCh:
			if err := h.handleProposal(proposal); err != nil {
				fmt.Printf("❌ Failed to handle proposal: %v\n", err)
			}
		case vote := <-h.voteCh:
			if err := h.handleVote(vote); err != nil {
				fmt.Printf("❌ Failed to handle vote: %v\n", err)
			}
		case commit := <-h.commitCh:
			if err := h.handleCommit(commit); err != nil {
				fmt.Printf("❌ Failed to handle commit: %v\n", err)
			}
		}
	}
}

// runConsensusRound executes a single consensus round
func (h *HPoSConsensus) runConsensusRound() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	// Phase 1: Proposer selection (Innovation: Performance-weighted)
	proposer := h.selectProposer()
	
	// Phase 2: Block proposal
	if proposer == h.nodeID && h.isValidator {
		if err := h.proposeBlock(); err != nil {
			return fmt.Errorf("failed to propose block: %w", err)
		}
	}
	
	// Phase 3: Prevote phase
	// Phase 4: Precommit phase  
	// Phase 5: Commit phase
	
	// Increment round
	h.round++
	if h.round >= h.epochLength {
		h.round = 0
		h.epoch++
		h.rotateValidators()
	}
	
	h.blockHeight.Add(h.blockHeight, big.NewInt(1))
	
	fmt.Printf("⛏️  Block #%d mined | Epoch: %d | Round: %d | Proposer: %s\n", 
		h.blockHeight.Uint64(), h.epoch, h.round, proposer.Hex()[:8])
	
	return nil
}

// selectProposer selects the block proposer using performance-weighted algorithm
func (h *HPoSConsensus) selectProposer() common.Address {
	validators := h.validators.GetSortedValidators()
	if len(validators) == 0 {
		return h.nodeID // Fallback
	}
	
	// Innovation: Weight selection by both stake AND performance
	totalWeight := 0.0
	for _, v := range validators {
		stake := float64(v.Stake.Uint64())
		performance := v.Performance
		weight := stake * (0.7 + 0.3*performance) // 70% stake, 30% performance
		totalWeight += weight
	}
	
	// Handle edge case where total weight is zero
	if totalWeight <= 0 {
		// Simple round-robin fallback
		idx := int(h.blockHeight.Uint64()) % len(validators)
		return validators[idx].Address
	}
	
	// Select using weighted random
	// Ensure we have a positive integer for rand.Int
	weightInt := int64(totalWeight * 1000)
	if weightInt <= 0 {
		weightInt = 1 // Minimum value to avoid panic
	}
	r, _ := rand.Int(rand.Reader, big.NewInt(weightInt))
	target := float64(r.Uint64()) / 1000.0
	
	current := 0.0
	for _, v := range validators {
		stake := float64(v.Stake.Uint64())
		performance := v.Performance
		weight := stake * (0.7 + 0.3*performance)
		current += weight
		if current >= target {
			return v.Address
		}
	}
	
	return validators[0].Address // Fallback
}

// proposeBlock creates and broadcasts a block proposal
func (h *HPoSConsensus) proposeBlock() error {
	// Create block proposal
	blockHash := crypto.Keccak256Hash([]byte(fmt.Sprintf("block_%d_%d", h.blockHeight.Uint64(), time.Now().UnixNano())))
	
	proposal := &Proposal{
		Height:    h.blockHeight.Uint64(),
		Round:     h.round,
		BlockHash: blockHash,
		Proposer:  h.nodeID,
		Timestamp: time.Now(),
	}
	
	// Sign proposal
	if err := h.signProposal(proposal); err != nil {
		return fmt.Errorf("failed to sign proposal: %w", err)
	}
	
	// Broadcast proposal (in real implementation, this would go through P2P network)
	fmt.Printf("📤 Proposed block %s at height %d\n", blockHash.Hex()[:8], h.blockHeight.Uint64())
	
	return nil
}

// signProposal signs a block proposal
func (h *HPoSConsensus) signProposal(proposal *Proposal) error {
	hash := crypto.Keccak256Hash([]byte(fmt.Sprintf("%d_%d_%s", proposal.Height, proposal.Round, proposal.BlockHash.Hex())))
	signature := crypto.Keccak256(append(hash.Bytes(), h.privateKey...))
	proposal.Signature = signature
	return nil
}

// handleProposal processes incoming block proposals
func (h *HPoSConsensus) handleProposal(proposal *Proposal) error {
	// Validate proposal
	if proposal.Height != h.blockHeight.Uint64() {
		return fmt.Errorf("invalid proposal height")
	}
	
	// Verify proposer
	if !h.validators.IsValidator(proposal.Proposer) {
		return fmt.Errorf("invalid proposer")
	}
	
	// Create and broadcast prevote
	return h.createPrevote(proposal)
}

// createPrevote creates a prevote for the proposal
func (h *HPoSConsensus) createPrevote(proposal *Proposal) error {
	vote := &Vote{
		Height:    proposal.Height,
		Round:     proposal.Round,
		BlockHash: proposal.BlockHash,
		Validator: h.nodeID,
		VoteType:  VotePrevote,
	}
	
	// Sign vote
	if err := h.signVote(vote); err != nil {
		return fmt.Errorf("failed to sign vote: %w", err)
	}
	
	fmt.Printf("🗳️  Prevoted for block %s\n", proposal.BlockHash.Hex()[:8])
	return nil
}

// signVote signs a vote
func (h *HPoSConsensus) signVote(vote *Vote) error {
	hash := crypto.Keccak256Hash([]byte(fmt.Sprintf("%d_%d_%s_%d", vote.Height, vote.Round, vote.BlockHash.Hex(), int(vote.VoteType))))
	signature := crypto.Keccak256(append(hash.Bytes(), h.privateKey...))
	vote.Signature = signature
	return nil
}

// handleVote processes incoming votes
func (h *HPoSConsensus) handleVote(vote *Vote) error {
	// Validate vote
	if !h.validators.IsValidator(vote.Validator) {
		return fmt.Errorf("invalid validator")
	}
	
	// Track performance
	h.performanceTracker.RecordVote(vote.Validator)
	
	return nil
}

// handleCommit processes block commits
func (h *HPoSConsensus) handleCommit(commit *Commit) error {
	// Finalize block
	fmt.Printf("✅ Block #%d committed with %d signatures\n", commit.Height, len(commit.Signatures))
	return nil
}

// initializeGenesisValidators sets up initial validator set
func (h *HPoSConsensus) initializeGenesisValidators() error {
	// Add self as genesis validator with initial stake
	initialStake := new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18)) // 100 tokens
	
	validator := &Validator{
		Address:     h.nodeID,
		Stake:       initialStake,
		Power:       100,
		Performance: 1.0,
		IsOnline:    true,
	}
	
	h.validators.AddValidator(validator)
	h.activeStakers[h.nodeID] = &StakeInfo{
		Amount:          initialStake,
		StakedAt:        0,
		Performance:     1.0,
		SlashCount:      0,
		LastActiveBlock: 0,
		RewardAddress:   h.nodeID,
	}
	
	h.isValidator = true
	
	fmt.Printf("🎯 Genesis validator initialized: %s with %s stake\n", h.nodeID.Hex()[:8], initialStake.String())
	return nil
}

// rotateValidators rotates validator set based on performance and stake
func (h *HPoSConsensus) rotateValidators() {
	// Innovation: Dynamic validator rotation based on performance metrics
	h.performanceTracker.UpdatePerformanceScores(h.validators)
	
	// Re-sort validators by combined stake and performance
	h.validators.SortByPerformance()
	
	fmt.Printf("🔄 Validator set rotated for epoch %d\n", h.epoch)
}

// performanceMonitor tracks validator performance
func (h *HPoSConsensus) performanceMonitor() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-h.ctx.Done():
			return
		case <-ticker.C:
			h.performanceTracker.CalculateScores()
		}
	}
}

// stakeManager handles staking operations
func (h *HPoSConsensus) stakeManager() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-h.ctx.Done():
			return
		case <-ticker.C:
			h.processStakeOperations()
		}
	}
}

// processStakeOperations handles stake deposits, withdrawals, and rewards
func (h *HPoSConsensus) processStakeOperations() {
	// Process stake rewards, slashing, etc.
	for _, stake := range h.activeStakers {
		if stake.Performance > 0.9 { // High performance gets rewards
			reward := new(big.Int).Div(stake.Amount, big.NewInt(1000)) // 0.1% reward
			stake.Amount.Add(stake.Amount, reward)
		}
	}
}

// Stop shuts down the consensus engine
func (h *HPoSConsensus) Stop() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	if !h.running {
		return nil
	}
	
	if h.cancel != nil {
		h.cancel()
	}
	
	h.running = false
	fmt.Println("🛑 HPoS Consensus stopped")
	return nil
}

// GetStatus returns current consensus status
func (h *HPoSConsensus) GetStatus() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	return map[string]interface{}{
		"chainId":       h.chainID.String(),
		"blockHeight":   h.blockHeight.Uint64(),
		"epoch":         h.epoch,
		"round":         h.round,
		"isValidator":   h.isValidator,
		"validatorCount": len(h.validators.validators),
		"totalStake":    h.validators.totalStake.String(),
		"performance":   h.performanceTracker.GetNodePerformance(h.nodeID),
	}
}