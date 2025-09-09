package unified

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// ConsensusEngine implements a CometBFT-inspired consensus mechanism
// Simplified for L2 use case but maintains key properties
type ConsensusEngine struct {
	config     *ConsensusConfig
	validators *ValidatorSet

	// Current consensus state
	height uint64
	round  uint32
	step   ConsensusStep

	// Validator state
	isValidator    bool
	validatorIndex int
	privateKey     *ecdsa.PrivateKey

	// Consensus timing
	timeoutPropose   time.Duration
	timeoutPrevote   time.Duration
	timeoutPrecommit time.Duration

	// Vote tracking
	votes map[VoteType]*VoteSet

	mu      sync.RWMutex
	running bool
}

// ConsensusConfig holds consensus configuration
type ConsensusConfig struct {
	BlockTime        time.Duration `json:"block_time"`
	ValidatorCount   int           `json:"validator_count"`
	ConsensusTimeout time.Duration `json:"consensus_timeout"`
}

// ConsensusStep represents the current step in consensus
type ConsensusStep int

const (
	StepPropose ConsensusStep = iota
	StepPrevote
	StepPrecommit
	StepCommit
)

// VoteType represents different types of votes in consensus
type VoteType int

const (
	VoteTypePrevote VoteType = iota
	VoteTypePrecommit
)

// Vote represents a consensus vote
type Vote struct {
	Type      VoteType       `json:"type"`
	Height    uint64         `json:"height"`
	Round     uint32         `json:"round"`
	BlockHash common.Hash    `json:"block_hash"`
	Validator common.Address `json:"validator"`
	Signature []byte         `json:"signature"`
	Timestamp time.Time      `json:"timestamp"`
}

// VoteSet tracks votes for a specific type and round
type VoteSet struct {
	votes    map[common.Address]*Vote
	majority int
	total    int
}

// NewConsensusEngine creates a new consensus engine
func NewConsensusEngine(config *ConsensusConfig) (*ConsensusEngine, error) {
	// Generate validator private key (in production, this would be loaded)
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate validator key: %w", err)
	}

	engine := &ConsensusEngine{
		config:           config,
		privateKey:       privateKey,
		timeoutPropose:   config.ConsensusTimeout,
		timeoutPrevote:   config.ConsensusTimeout / 2,
		timeoutPrecommit: config.ConsensusTimeout / 2,
		votes:            make(map[VoteType]*VoteSet),
	}

	// Initialize vote sets
	engine.votes[VoteTypePrevote] = &VoteSet{
		votes:    make(map[common.Address]*Vote),
		majority: (config.ValidatorCount * 2 / 3) + 1,
		total:    config.ValidatorCount,
	}
	engine.votes[VoteTypePrecommit] = &VoteSet{
		votes:    make(map[common.Address]*Vote),
		majority: (config.ValidatorCount * 2 / 3) + 1,
		total:    config.ValidatorCount,
	}

	return engine, nil
}

// Start begins the consensus engine
func (c *ConsensusEngine) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		return fmt.Errorf("consensus engine already running")
	}

	fmt.Println("🔥 Starting Consensus Engine (CometBFT-inspired)")
	fmt.Printf("   • Timeout Propose: %v\n", c.timeoutPropose)
	fmt.Printf("   • Timeout Prevote: %v\n", c.timeoutPrevote)
	fmt.Printf("   • Timeout Precommit: %v\n", c.timeoutPrecommit)
	fmt.Printf("   • Validator Count: %d\n", c.config.ValidatorCount)
	fmt.Printf("   • Majority Required: %d votes\n", c.votes[VoteTypePrevote].majority)

	c.running = true
	c.height = 1
	c.round = 0
	c.step = StepPropose

	// Start consensus rounds
	go c.consensusLoop(ctx)

	return nil
}

// consensusLoop runs the main consensus algorithm
func (c *ConsensusEngine) consensusLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if err := c.runConsensusRound(ctx); err != nil {
				fmt.Printf("❌ Consensus round failed: %v\n", err)
				time.Sleep(time.Second) // Brief pause before retry
			}
		}
	}
}

// runConsensusRound executes a single round of consensus
func (c *ConsensusEngine) runConsensusRound(ctx context.Context) error {
	c.mu.Lock()
	height := c.height
	round := c.round
	c.mu.Unlock()

	fmt.Printf("🔄 Consensus Round H:%d R:%d\n", height, round)

	// Step 1: Propose (if we're the proposer)
	if c.isProposer(height, round) {
		fmt.Printf("🎯 We are proposer for H:%d R:%d\n", height, round)
		if err := c.propose(ctx, height, round); err != nil {
			return fmt.Errorf("propose failed: %w", err)
		}
	}

	// Step 2: Prevote
	if err := c.prevote(ctx, height, round); err != nil {
		return fmt.Errorf("prevote failed: %w", err)
	}

	// Step 3: Precommit
	if err := c.precommit(ctx, height, round); err != nil {
		return fmt.Errorf("precommit failed: %w", err)
	}

	// Step 4: Commit (if we have majority)
	if c.hasMajority(VoteTypePrecommit) {
		if err := c.commit(ctx, height, round); err != nil {
			return fmt.Errorf("commit failed: %w", err)
		}

		// Move to next height
		c.mu.Lock()
		c.height++
		c.round = 0
		c.clearVotes()
		c.mu.Unlock()
	} else {
		// Move to next round
		c.mu.Lock()
		c.round++
		c.clearVotes()
		c.mu.Unlock()
	}

	return nil
}

// propose creates and broadcasts a block proposal
func (c *ConsensusEngine) propose(ctx context.Context, height uint64, round uint32) error {
	c.mu.Lock()
	c.step = StepPropose
	c.mu.Unlock()

	// In a real implementation, this would create a proper block proposal
	// For now, we simulate the proposal process

	fmt.Printf("📝 Creating block proposal for H:%d R:%d\n", height, round)

	// Simulate block creation time
	time.Sleep(100 * time.Millisecond)

	fmt.Printf("📡 Broadcasting proposal for H:%d R:%d\n", height, round)

	return nil
}

// prevote casts a prevote for the current proposal
func (c *ConsensusEngine) prevote(ctx context.Context, height uint64, round uint32) error {
	c.mu.Lock()
	c.step = StepPrevote
	c.mu.Unlock()

	// Wait for proposal or timeout
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(c.timeoutPropose):
		// Timeout, prevote nil
	}

	// Create prevote (simplified - in reality would validate proposal)
	vote := &Vote{
		Type:      VoteTypePrevote,
		Height:    height,
		Round:     round,
		BlockHash: common.Hash{}, // nil vote for simplicity
		Validator: crypto.PubkeyToAddress(c.privateKey.PublicKey),
		Timestamp: time.Now(),
	}

	// Sign vote
	signature, err := c.signVote(vote)
	if err != nil {
		return fmt.Errorf("failed to sign vote: %w", err)
	}
	vote.Signature = signature

	// Add our vote
	c.addVote(vote)

	fmt.Printf("🗳️  Prevoted for H:%d R:%d\n", height, round)

	// Wait for majority or timeout
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(c.timeoutPrevote):
		// Continue to precommit
	}

	return nil
}

// precommit casts a precommit vote
func (c *ConsensusEngine) precommit(ctx context.Context, height uint64, round uint32) error {
	c.mu.Lock()
	c.step = StepPrecommit
	c.mu.Unlock()

	// Check if we have majority prevotes
	hasMajorityPrevotes := c.hasMajority(VoteTypePrevote)

	// Create precommit vote
	vote := &Vote{
		Type:      VoteTypePrecommit,
		Height:    height,
		Round:     round,
		BlockHash: common.Hash{}, // Would be actual block hash if we have proposal
		Validator: crypto.PubkeyToAddress(c.privateKey.PublicKey),
		Timestamp: time.Now(),
	}

	// Only precommit if we have majority prevotes
	if !hasMajorityPrevotes {
		// Precommit nil
		vote.BlockHash = common.Hash{}
	}

	// Sign vote
	signature, err := c.signVote(vote)
	if err != nil {
		return fmt.Errorf("failed to sign vote: %w", err)
	}
	vote.Signature = signature

	// Add our vote
	c.addVote(vote)

	fmt.Printf("✅ Precommitted for H:%d R:%d (majority prevotes: %v)\n", height, round, hasMajorityPrevotes)

	// Wait for majority or timeout
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(c.timeoutPrecommit):
		// Continue
	}

	return nil
}

// commit finalizes the block if we have majority precommits
func (c *ConsensusEngine) commit(ctx context.Context, height uint64, round uint32) error {
	c.mu.Lock()
	c.step = StepCommit
	c.mu.Unlock()

	fmt.Printf("🔒 Committing block for H:%d R:%d\n", height, round)

	// In a real implementation, this would finalize the block
	// For now, we simulate the commit process

	time.Sleep(50 * time.Millisecond)

	fmt.Printf("✅ Block committed for H:%d R:%d\n", height, round)

	return nil
}

// ShouldProduceBlock determines if this validator should produce a block
func (c *ConsensusEngine) ShouldProduceBlock() (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.running {
		return false, nil
	}

	// Simple round-robin for now
	// In production, this would be more sophisticated
	return c.isProposer(c.height, c.round), nil
}

// ValidateBlock validates a block according to consensus rules
func (c *ConsensusEngine) ValidateBlock(block *types.Block) error {
	// Basic validation (in production, this would be comprehensive)
	if block == nil {
		return fmt.Errorf("block is nil")
	}

	if block.Number() == nil {
		return fmt.Errorf("block number is nil")
	}

	if block.Time() == 0 {
		return fmt.Errorf("block timestamp is zero")
	}

	// Additional consensus-specific validation would go here

	return nil
}

// FinalizeBlock finalizes a block after execution
func (c *ConsensusEngine) FinalizeBlock(block *types.Block) error {
	// In a real implementation, this would:
	// 1. Collect validator signatures
	// 2. Create consensus proof
	// 3. Update validator set if needed

	fmt.Printf("🔐 Finalizing block #%d with consensus\n", block.Number().Uint64())

	return nil
}

// GetCurrentValidator returns the current validator address
func (c *ConsensusEngine) GetCurrentValidator() common.Address {
	return crypto.PubkeyToAddress(c.privateKey.PublicKey)
}

// Helper methods

// isProposer determines if this validator is the proposer for a given height/round
func (c *ConsensusEngine) isProposer(height uint64, round uint32) bool {
	// Simple deterministic proposer selection
	// In production, this would be based on validator set and weights
	proposerIndex := (height + uint64(round)) % uint64(c.config.ValidatorCount)
	return int(proposerIndex) == c.validatorIndex
}

// signVote signs a consensus vote
func (c *ConsensusEngine) signVote(vote *Vote) ([]byte, error) {
	// Create vote hash
	voteBytes := fmt.Sprintf("%d:%d:%d:%s:%s",
		vote.Type, vote.Height, vote.Round, vote.BlockHash.Hex(), vote.Validator.Hex())

	hash := crypto.Keccak256Hash([]byte(voteBytes))

	// Sign the hash
	signature, err := crypto.Sign(hash.Bytes(), c.privateKey)
	if err != nil {
		return nil, err
	}

	return signature, nil
}

// addVote adds a vote to the appropriate vote set
func (c *ConsensusEngine) addVote(vote *Vote) {
	c.mu.Lock()
	defer c.mu.Unlock()

	voteSet := c.votes[vote.Type]
	voteSet.votes[vote.Validator] = vote
}

// hasMajority checks if we have majority votes for a type
func (c *ConsensusEngine) hasMajority(voteType VoteType) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	voteSet := c.votes[voteType]
	return len(voteSet.votes) >= voteSet.majority
}

// clearVotes clears all votes for the next round
func (c *ConsensusEngine) clearVotes() {
	for _, voteSet := range c.votes {
		voteSet.votes = make(map[common.Address]*Vote)
	}
}

// Stop stops the consensus engine
func (c *ConsensusEngine) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running {
		return nil
	}

	c.running = false
	fmt.Println("🛑 Consensus engine stopped")

	return nil
}
