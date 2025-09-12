package consensus

import (
	"math/big"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// NewValidatorSet creates a new validator set
func NewValidatorSet() *ValidatorSet {
	return &ValidatorSet{
		validators: make(map[common.Address]*Validator),
		sorted:     make([]*Validator, 0),
		totalStake: big.NewInt(0),
	}
}

// AddValidator adds a validator to the set
func (vs *ValidatorSet) AddValidator(validator *Validator) {
	vs.mu.Lock()
	defer vs.mu.Unlock()
	
	if existing, exists := vs.validators[validator.Address]; exists {
		vs.totalStake.Sub(vs.totalStake, existing.Stake)
	}
	
	vs.validators[validator.Address] = validator
	vs.totalStake.Add(vs.totalStake, validator.Stake)
	vs.updateSorted()
}

// RemoveValidator removes a validator from the set
func (vs *ValidatorSet) RemoveValidator(address common.Address) {
	vs.mu.Lock()
	defer vs.mu.Unlock()
	
	if validator, exists := vs.validators[address]; exists {
		delete(vs.validators, address)
		vs.totalStake.Sub(vs.totalStake, validator.Stake)
		vs.updateSorted()
	}
}

// IsValidator checks if an address is a validator
func (vs *ValidatorSet) IsValidator(address common.Address) bool {
	vs.mu.RLock()
	defer vs.mu.RUnlock()
	
	_, exists := vs.validators[address]
	return exists
}

// GetValidator returns a validator by address
func (vs *ValidatorSet) GetValidator(address common.Address) (*Validator, bool) {
	vs.mu.RLock()
	defer vs.mu.RUnlock()
	
	validator, exists := vs.validators[address]
	return validator, exists
}

// GetSortedValidators returns validators sorted by power
func (vs *ValidatorSet) GetSortedValidators() []*Validator {
	vs.mu.RLock()
	defer vs.mu.RUnlock()
	
	// Return a copy to prevent external modification
	result := make([]*Validator, len(vs.sorted))
	copy(result, vs.sorted)
	return result
}

// SortByPerformance sorts validators by combined stake and performance
func (vs *ValidatorSet) SortByPerformance() {
	vs.mu.Lock()
	defer vs.mu.Unlock()
	
	vs.updateSorted()
}

// updateSorted updates the sorted validator list
func (vs *ValidatorSet) updateSorted() {
	vs.sorted = make([]*Validator, 0, len(vs.validators))
	for _, validator := range vs.validators {
		vs.sorted = append(vs.sorted, validator)
	}
	
	// Sort by combined score: stake * performance
	sort.Slice(vs.sorted, func(i, j int) bool {
		scoreI := float64(vs.sorted[i].Stake.Uint64()) * vs.sorted[i].Performance
		scoreJ := float64(vs.sorted[j].Stake.Uint64()) * vs.sorted[j].Performance
		return scoreI > scoreJ
	})
}

// GetTotalStake returns the total stake in the validator set
func (vs *ValidatorSet) GetTotalStake() *big.Int {
	vs.mu.RLock()
	defer vs.mu.RUnlock()
	
	return new(big.Int).Set(vs.totalStake)
}

// Size returns the number of validators
func (vs *ValidatorSet) Size() int {
	vs.mu.RLock()
	defer vs.mu.RUnlock()
	
	return len(vs.validators)
}

// NewPerformanceTracker creates a new performance tracker
func NewPerformanceTracker() *PerformanceTracker {
	return &PerformanceTracker{
		blockProposals: make(map[common.Address]uint64),
		blockSignings:  make(map[common.Address]uint64),
		missedBlocks:   make(map[common.Address]uint64),
		responseTime:   make(map[common.Address]time.Duration),
	}
}

// RecordProposal records a block proposal
func (pt *PerformanceTracker) RecordProposal(validator common.Address) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	
	pt.blockProposals[validator]++
}

// RecordVote records a vote/signing
func (pt *PerformanceTracker) RecordVote(validator common.Address) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	
	pt.blockSignings[validator]++
}

// RecordMissedBlock records a missed block
func (pt *PerformanceTracker) RecordMissedBlock(validator common.Address) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	
	pt.missedBlocks[validator]++
}

// UpdatePerformanceScores updates performance scores for all validators
func (pt *PerformanceTracker) UpdatePerformanceScores(validators *ValidatorSet) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	
	for addr := range validators.validators {
		performance := pt.calculatePerformanceScore(addr)
		if validator, exists := validators.validators[addr]; exists {
			validator.Performance = performance
		}
	}
}

// calculatePerformanceScore calculates performance score for a validator
func (pt *PerformanceTracker) calculatePerformanceScore(validator common.Address) float64 {
	proposals := pt.blockProposals[validator]
	signings := pt.blockSignings[validator]
	missed := pt.missedBlocks[validator]
	
	if proposals+signings+missed == 0 {
		return 1.0 // New validator gets full score
	}
	
	// Performance = (proposals * 2 + signings) / (total_expected)
	// Penalties for missed blocks
	totalActivity := proposals + signings + missed
	score := float64(proposals*2+signings) / float64(totalActivity*2)
	
	// Apply missed block penalty
	if missed > 0 {
		penalty := float64(missed) / float64(totalActivity)
		score -= penalty * 0.5 // 50% penalty weight
	}
	
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}
	
	return score
}

// CalculateScores calculates and updates all performance scores
func (pt *PerformanceTracker) CalculateScores() {
	// This would be called periodically to update scores
}

// GetNodePerformance returns performance score for a specific node
func (pt *PerformanceTracker) GetNodePerformance(validator common.Address) float64 {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	
	return pt.calculatePerformanceScore(validator)
}

// ResetEpochMetrics resets metrics for a new epoch
func (pt *PerformanceTracker) ResetEpochMetrics() {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	
	// Keep some history, but reset current epoch metrics
	for addr := range pt.blockProposals {
		pt.blockProposals[addr] = pt.blockProposals[addr] / 2 // Decay
		pt.blockSignings[addr] = pt.blockSignings[addr] / 2
		pt.missedBlocks[addr] = pt.missedBlocks[addr] / 2
	}
}