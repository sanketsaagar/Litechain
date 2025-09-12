package staking

import (
	"fmt"
	"math/big"
	"sort"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// StakingManager manages validator staking for the L1 chain
// Innovation: Performance-weighted staking with dynamic slashing
type StakingManager struct {
	// Staking state
	validators     map[common.Address]*ValidatorInfo
	delegations    map[common.Address]map[common.Address]*Delegation // validator -> delegator -> delegation
	totalStaked    *big.Int
	
	// Staking parameters
	minValidatorStake *big.Int
	minDelegation     *big.Int
	maxValidators     int
	unbondingPeriod   time.Duration
	slashingEnabled   bool
	
	// Performance tracking
	performanceWindow time.Duration
	performanceMetrics map[common.Address]*PerformanceMetrics
	
	// Reward distribution
	rewardPool        *big.Int
	commissionPool    map[common.Address]*big.Int
	
	// Slashing
	slashingRules     []SlashingRule
	slashingHistory   map[common.Address][]SlashingEvent
	
	// Governance integration
	governanceWeight  map[common.Address]*big.Int
	
	mu sync.RWMutex
}

// ValidatorInfo contains validator information and status
type ValidatorInfo struct {
	Address         common.Address
	PubKey          []byte
	Moniker         string
	Description     string
	Website         string
	Commission      uint64 // Basis points (0-10000)
	MaxCommission   uint64 // Max commission rate
	
	// Staking info
	SelfStake       *big.Int
	TotalStake      *big.Int
	DelegatorCount  int
	
	// Status
	Status          ValidatorStatus
	JailedUntil     time.Time
	CreatedAt       time.Time
	LastActive      time.Time
	
	// Performance
	Performance     *ValidatorPerformance
	
	// Slashing
	SlashingCount   uint64
	SlashedAmount   *big.Int
	
	// Rewards
	AccRewardsPerShare *big.Int
	LastRewardHeight   uint64
}

// ValidatorStatus represents validator status
type ValidatorStatus int

const (
	StatusUnknown ValidatorStatus = iota
	StatusActive
	StatusInactive
	StatusJailed
	StatusUnbonding
	StatusUnbonded
)

// ValidatorPerformance tracks validator performance metrics
type ValidatorPerformance struct {
	BlocksProposed     uint64
	BlocksMissed       uint64
	VotesSubmitted     uint64
	VotesMissed        uint64
	ResponseTime       time.Duration
	UptimePercentage   float64
	PerformanceScore   float64
	LastUpdated        time.Time
}

// Delegation represents a delegation to a validator
type Delegation struct {
	Delegator       common.Address
	Validator       common.Address
	Amount          *big.Int
	CreatedAt       time.Time
	
	// Unbonding
	IsUnbonding     bool
	UnbondingHeight uint64
	UnbondingTime   time.Time
	
	// Rewards
	RewardDebt      *big.Int
	AccRewards      *big.Int
}

// UnbondingDelegation represents an unbonding delegation
type UnbondingDelegation struct {
	Delegator       common.Address
	Validator       common.Address
	Amount          *big.Int
	CompletionTime  time.Time
	Height          uint64
}

// PerformanceMetrics tracks detailed performance metrics
type PerformanceMetrics struct {
	StartTime         time.Time
	BlocksProposed    uint64
	BlocksSigned      uint64
	BlocksMissed      uint64
	ResponseTimes     []time.Duration
	DowntimeEvents    []DowntimeEvent
	LastHeartbeat     time.Time
}

// DowntimeEvent records downtime incidents
type DowntimeEvent struct {
	StartTime    time.Time
	EndTime      time.Time
	Duration     time.Duration
	Reason       string
	BlockHeight  uint64
}

// SlashingRule defines conditions for slashing
type SlashingRule struct {
	Violation       ViolationType
	SlashPercentage uint64        // Basis points
	JailDuration    time.Duration
	Enabled         bool
}

// ViolationType represents different violation types
type ViolationType int

const (
	ViolationDoubleSign ViolationType = iota
	ViolationDowntime
	ViolationMissedBlocks
	ViolationInvalidProposal
	ViolationByzantineBehavior
)

// SlashingEvent records a slashing event
type SlashingEvent struct {
	Validator   common.Address
	Violation   ViolationType
	Amount      *big.Int
	Height      uint64
	Timestamp   time.Time
	Evidence    []byte
	JailTime    time.Duration
}

// Constants
const (
	DefaultMinValidatorStake = "100000000000000000000"    // 100 LIGHT
	DefaultMinDelegation     = "1000000000000000000"      // 1 LIGHT
	DefaultUnbondingPeriod   = 14 * 24 * time.Hour       // 14 days
	DefaultMaxValidators     = 21
	DefaultPerformanceWindow = 24 * time.Hour             // 24 hours
	
	// Slashing percentages (basis points)
	DoubleSignSlash     = 5000  // 50%
	DowntimeSlash       = 100   // 1%
	MissedBlocksSlash   = 50    // 0.5%
	InvalidProposalSlash = 1000  // 10%
	ByzantineSlash      = 10000 // 100%
)

// NewStakingManager creates a new staking manager
func NewStakingManager() *StakingManager {
	minValidatorStake, _ := new(big.Int).SetString(DefaultMinValidatorStake, 10)
	minDelegation, _ := new(big.Int).SetString(DefaultMinDelegation, 10)
	
	sm := &StakingManager{
		validators:         make(map[common.Address]*ValidatorInfo),
		delegations:        make(map[common.Address]map[common.Address]*Delegation),
		totalStaked:        big.NewInt(0),
		minValidatorStake:  minValidatorStake,
		minDelegation:      minDelegation,
		maxValidators:      DefaultMaxValidators,
		unbondingPeriod:    DefaultUnbondingPeriod,
		slashingEnabled:    true,
		performanceWindow:  DefaultPerformanceWindow,
		performanceMetrics: make(map[common.Address]*PerformanceMetrics),
		rewardPool:         big.NewInt(0),
		commissionPool:     make(map[common.Address]*big.Int),
		slashingHistory:    make(map[common.Address][]SlashingEvent),
		governanceWeight:   make(map[common.Address]*big.Int),
	}
	
	// Initialize slashing rules
	sm.initializeSlashingRules()
	
	return sm
}

// initializeSlashingRules sets up default slashing rules
func (sm *StakingManager) initializeSlashingRules() {
	sm.slashingRules = []SlashingRule{
		{
			Violation:       ViolationDoubleSign,
			SlashPercentage: DoubleSignSlash,
			JailDuration:    7 * 24 * time.Hour, // 7 days
			Enabled:         true,
		},
		{
			Violation:       ViolationDowntime,
			SlashPercentage: DowntimeSlash,
			JailDuration:    6 * time.Hour,
			Enabled:         true,
		},
		{
			Violation:       ViolationMissedBlocks,
			SlashPercentage: MissedBlocksSlash,
			JailDuration:    1 * time.Hour,
			Enabled:         true,
		},
		{
			Violation:       ViolationInvalidProposal,
			SlashPercentage: InvalidProposalSlash,
			JailDuration:    24 * time.Hour,
			Enabled:         true,
		},
		{
			Violation:       ViolationByzantineBehavior,
			SlashPercentage: ByzantineSlash,
			JailDuration:    0, // Permanent jail
			Enabled:         true,
		},
	}
}

// CreateValidator creates a new validator
func (sm *StakingManager) CreateValidator(
	address common.Address,
	pubKey []byte,
	moniker string,
	description string,
	website string,
	commission uint64,
	maxCommission uint64,
	selfStake *big.Int,
) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	// Validate inputs
	if selfStake.Cmp(sm.minValidatorStake) < 0 {
		return fmt.Errorf("self stake %s below minimum %s", selfStake.String(), sm.minValidatorStake.String())
	}
	
	if commission > 10000 {
		return fmt.Errorf("commission rate cannot exceed 100%%")
	}
	
	if maxCommission > 10000 || maxCommission < commission {
		return fmt.Errorf("invalid max commission rate")
	}
	
	// Check if validator already exists
	if _, exists := sm.validators[address]; exists {
		return fmt.Errorf("validator already exists")
	}
	
	// Check validator limit
	if len(sm.validators) >= sm.maxValidators {
		// Check if this validator would be in top validators by stake
		if !sm.wouldBeTopValidator(selfStake) {
			return fmt.Errorf("validator set full and stake insufficient")
		}
	}
	
	// Create validator
	validator := &ValidatorInfo{
		Address:       address,
		PubKey:        pubKey,
		Moniker:       moniker,
		Description:   description,
		Website:       website,
		Commission:    commission,
		MaxCommission: maxCommission,
		SelfStake:     new(big.Int).Set(selfStake),
		TotalStake:    new(big.Int).Set(selfStake),
		Status:        StatusActive,
		CreatedAt:     time.Now(),
		LastActive:    time.Now(),
		Performance:   &ValidatorPerformance{},
		SlashedAmount: big.NewInt(0),
		AccRewardsPerShare: big.NewInt(0),
	}
	
	sm.validators[address] = validator
	sm.delegations[address] = make(map[common.Address]*Delegation)
	sm.performanceMetrics[address] = &PerformanceMetrics{
		StartTime:     time.Now(),
		ResponseTimes: make([]time.Duration, 0),
	}
	sm.commissionPool[address] = big.NewInt(0)
	sm.totalStaked.Add(sm.totalStaked, selfStake)
	
	// Add self-delegation
	selfDelegation := &Delegation{
		Delegator: address,
		Validator: address,
		Amount:    new(big.Int).Set(selfStake),
		CreatedAt: time.Now(),
		RewardDebt: big.NewInt(0),
		AccRewards: big.NewInt(0),
	}
	sm.delegations[address][address] = selfDelegation
	
	return nil
}

// Delegate delegates tokens to a validator
func (sm *StakingManager) Delegate(delegator, validator common.Address, amount *big.Int) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if amount.Cmp(sm.minDelegation) < 0 {
		return fmt.Errorf("delegation amount below minimum")
	}
	
	validatorInfo, exists := sm.validators[validator]
	if !exists {
		return fmt.Errorf("validator does not exist")
	}
	
	if validatorInfo.Status != StatusActive {
		return fmt.Errorf("validator not active")
	}
	
	// Update or create delegation
	if existing, exists := sm.delegations[validator][delegator]; exists {
		// Add to existing delegation
		existing.Amount.Add(existing.Amount, amount)
	} else {
		// Create new delegation
		delegation := &Delegation{
			Delegator:  delegator,
			Validator:  validator,
			Amount:     new(big.Int).Set(amount),
			CreatedAt:  time.Now(),
			RewardDebt: big.NewInt(0),
			AccRewards: big.NewInt(0),
		}
		sm.delegations[validator][delegator] = delegation
		validatorInfo.DelegatorCount++
	}
	
	// Update validator stake
	validatorInfo.TotalStake.Add(validatorInfo.TotalStake, amount)
	sm.totalStaked.Add(sm.totalStaked, amount)
	
	// Update governance weight
	if existing, exists := sm.governanceWeight[delegator]; exists {
		existing.Add(existing, amount)
	} else {
		sm.governanceWeight[delegator] = new(big.Int).Set(amount)
	}
	
	return nil
}

// Undelegate initiates unbonding of delegated tokens
func (sm *StakingManager) Undelegate(delegator, validator common.Address, amount *big.Int) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	delegation, exists := sm.delegations[validator][delegator]
	if !exists {
		return fmt.Errorf("delegation does not exist")
	}
	
	if delegation.Amount.Cmp(amount) < 0 {
		return fmt.Errorf("insufficient delegated amount")
	}
	
	validatorInfo := sm.validators[validator]
	
	// Check if this would leave validator below minimum
	newTotalStake := new(big.Int).Sub(validatorInfo.TotalStake, amount)
	if delegator == validator && newTotalStake.Cmp(sm.minValidatorStake) < 0 {
		return fmt.Errorf("would leave validator below minimum self stake")
	}
	
	// Update delegation
	delegation.Amount.Sub(delegation.Amount, amount)
	delegation.IsUnbonding = true
	delegation.UnbondingTime = time.Now().Add(sm.unbondingPeriod)
	
	// Remove delegation if amount becomes zero
	if delegation.Amount.Sign() == 0 {
		delete(sm.delegations[validator], delegator)
		validatorInfo.DelegatorCount--
	}
	
	// Update validator stake
	validatorInfo.TotalStake.Sub(validatorInfo.TotalStake, amount)
	sm.totalStaked.Sub(sm.totalStaked, amount)
	
	// Update governance weight
	if govWeight, exists := sm.governanceWeight[delegator]; exists {
		govWeight.Sub(govWeight, amount)
		if govWeight.Sign() == 0 {
			delete(sm.governanceWeight, delegator)
		}
	}
	
	return nil
}

// SlashValidator applies slashing to a validator
func (sm *StakingManager) SlashValidator(validator common.Address, violation ViolationType, evidence []byte, height uint64) error {
	if !sm.slashingEnabled {
		return nil
	}
	
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	validatorInfo, exists := sm.validators[validator]
	if !exists {
		return fmt.Errorf("validator does not exist")
	}
	
	// Find slashing rule
	var rule *SlashingRule
	for i := range sm.slashingRules {
		if sm.slashingRules[i].Violation == violation && sm.slashingRules[i].Enabled {
			rule = &sm.slashingRules[i]
			break
		}
	}
	
	if rule == nil {
		return fmt.Errorf("no slashing rule found for violation")
	}
	
	// Calculate slash amount
	slashAmount := new(big.Int).Mul(validatorInfo.TotalStake, big.NewInt(int64(rule.SlashPercentage)))
	slashAmount.Div(slashAmount, big.NewInt(10000))
	
	// Apply slashing
	validatorInfo.TotalStake.Sub(validatorInfo.TotalStake, slashAmount)
	validatorInfo.SlashedAmount.Add(validatorInfo.SlashedAmount, slashAmount)
	validatorInfo.SlashingCount++
	sm.totalStaked.Sub(sm.totalStaked, slashAmount)
	
	// Jail validator if required
	if rule.JailDuration > 0 {
		validatorInfo.Status = StatusJailed
		if rule.JailDuration == 0 {
			// Permanent jail (tombstone)
			validatorInfo.JailedUntil = time.Time{}
		} else {
			validatorInfo.JailedUntil = time.Now().Add(rule.JailDuration)
		}
	}
	
	// Record slashing event
	event := SlashingEvent{
		Validator: validator,
		Violation: violation,
		Amount:    slashAmount,
		Height:    height,
		Timestamp: time.Now(),
		Evidence:  evidence,
		JailTime:  rule.JailDuration,
	}
	
	sm.slashingHistory[validator] = append(sm.slashingHistory[validator], event)
	
	return nil
}

// UpdatePerformance updates validator performance metrics
func (sm *StakingManager) UpdatePerformance(validator common.Address, metrics *PerformanceMetrics) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if validatorInfo, exists := sm.validators[validator]; exists {
		sm.performanceMetrics[validator] = metrics
		
		// Calculate performance score
		score := sm.calculatePerformanceScore(metrics)
		validatorInfo.Performance.PerformanceScore = score
		validatorInfo.Performance.LastUpdated = time.Now()
		validatorInfo.LastActive = time.Now()
	}
}

// calculatePerformanceScore calculates a performance score from metrics
func (sm *StakingManager) calculatePerformanceScore(metrics *PerformanceMetrics) float64 {
	if metrics.BlocksProposed+metrics.BlocksSigned == 0 {
		return 1.0 // New validator gets full score
	}
	
	// Calculate uptime percentage
	totalBlocks := metrics.BlocksProposed + metrics.BlocksSigned + metrics.BlocksMissed
	if totalBlocks == 0 {
		return 1.0
	}
	
	uptime := float64(metrics.BlocksProposed+metrics.BlocksSigned) / float64(totalBlocks)
	
	// Calculate response time score (lower is better)
	avgResponseTime := sm.calculateAverageResponseTime(metrics.ResponseTimes)
	responseScore := 1.0
	if avgResponseTime > 0 {
		// Score decreases as response time increases
		responseScore = 1.0 / (1.0 + avgResponseTime.Seconds())
	}
	
	// Combine metrics (weighted average)
	score := (uptime * 0.7) + (responseScore * 0.3)
	
	// Penalize for downtime events
	if len(metrics.DowntimeEvents) > 0 {
		penalty := float64(len(metrics.DowntimeEvents)) * 0.05 // 5% penalty per event
		score -= penalty
	}
	
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}
	
	return score
}

// calculateAverageResponseTime calculates average response time
func (sm *StakingManager) calculateAverageResponseTime(responseTimes []time.Duration) time.Duration {
	if len(responseTimes) == 0 {
		return 0
	}
	
	var total time.Duration
	for _, rt := range responseTimes {
		total += rt
	}
	
	return total / time.Duration(len(responseTimes))
}

// GetActiveValidators returns the current active validator set
func (sm *StakingManager) GetActiveValidators() []*ValidatorInfo {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	var active []*ValidatorInfo
	for _, validator := range sm.validators {
		if validator.Status == StatusActive && time.Now().After(validator.JailedUntil) {
			active = append(active, validator)
		}
	}
	
	// Sort by total stake (descending)
	sort.Slice(active, func(i, j int) bool {
		return active[i].TotalStake.Cmp(active[j].TotalStake) > 0
	})
	
	// Limit to max validators
	if len(active) > sm.maxValidators {
		active = active[:sm.maxValidators]
	}
	
	return active
}

// GetTopValidatorsByPerformance returns validators ranked by performance-weighted stake
func (sm *StakingManager) GetTopValidatorsByPerformance() []*ValidatorInfo {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	var validators []*ValidatorInfo
	for _, validator := range sm.validators {
		if validator.Status == StatusActive && time.Now().After(validator.JailedUntil) {
			validators = append(validators, validator)
		}
	}
	
	// Sort by performance-weighted stake
	sort.Slice(validators, func(i, j int) bool {
		scoreI := float64(validators[i].TotalStake.Uint64()) * validators[i].Performance.PerformanceScore
		scoreJ := float64(validators[j].TotalStake.Uint64()) * validators[j].Performance.PerformanceScore
		return scoreI > scoreJ
	})
	
	return validators
}

// wouldBeTopValidator checks if a stake amount would qualify for top validator set
func (sm *StakingManager) wouldBeTopValidator(stake *big.Int) bool {
	active := sm.GetActiveValidators()
	if len(active) < sm.maxValidators {
		return true
	}
	
	// Check if stake is higher than lowest validator
	lowestStake := active[len(active)-1].TotalStake
	return stake.Cmp(lowestStake) > 0
}

// DistributeRewards distributes block rewards to validators and delegators
func (sm *StakingManager) DistributeRewards(totalRewards *big.Int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if totalRewards.Sign() <= 0 {
		return
	}
	
	activeValidators := sm.GetActiveValidators()
	if len(activeValidators) == 0 {
		return
	}
	
	// Distribute rewards proportionally to stake
	for _, validator := range activeValidators {
		// Calculate validator's share of total stake
		validatorShare := new(big.Int).Mul(totalRewards, validator.TotalStake)
		validatorShare.Div(validatorShare, sm.totalStaked)
		
		if validatorShare.Sign() <= 0 {
			continue
		}
		
		// Calculate commission
		commission := new(big.Int).Mul(validatorShare, big.NewInt(int64(validator.Commission)))
		commission.Div(commission, big.NewInt(10000))
		
		// Add commission to validator
		if existing, exists := sm.commissionPool[validator.Address]; exists {
			existing.Add(existing, commission)
		} else {
			sm.commissionPool[validator.Address] = new(big.Int).Set(commission)
		}
		
		// Distribute remaining to delegators
		delegatorRewards := new(big.Int).Sub(validatorShare, commission)
		sm.distributeToDelegators(validator.Address, delegatorRewards)
	}
}

// distributeToDelegators distributes rewards to validator's delegators
func (sm *StakingManager) distributeToDelegators(validator common.Address, totalRewards *big.Int) {
	validatorInfo := sm.validators[validator]
	if validatorInfo.TotalStake.Sign() == 0 {
		return
	}
	
	// Calculate rewards per share
	rewardsPerShare := new(big.Int).Mul(totalRewards, big.NewInt(1e18))
	rewardsPerShare.Div(rewardsPerShare, validatorInfo.TotalStake)
	
	// Update accumulated rewards per share
	validatorInfo.AccRewardsPerShare.Add(validatorInfo.AccRewardsPerShare, rewardsPerShare)
	
	// Update individual delegator rewards
	for _, delegation := range sm.delegations[validator] {
		// Calculate pending rewards
		pendingRewards := new(big.Int).Mul(delegation.Amount, validatorInfo.AccRewardsPerShare)
		pendingRewards.Div(pendingRewards, big.NewInt(1e18))
		pendingRewards.Sub(pendingRewards, delegation.RewardDebt)
		
		// Add to accumulated rewards
		delegation.AccRewards.Add(delegation.AccRewards, pendingRewards)
		
		// Update reward debt
		delegation.RewardDebt = new(big.Int).Mul(delegation.Amount, validatorInfo.AccRewardsPerShare)
		delegation.RewardDebt.Div(delegation.RewardDebt, big.NewInt(1e18))
	}
}

// GetValidatorInfo returns validator information
func (sm *StakingManager) GetValidatorInfo(address common.Address) (*ValidatorInfo, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	validator, exists := sm.validators[address]
	return validator, exists
}

// GetDelegation returns delegation information
func (sm *StakingManager) GetDelegation(delegator, validator common.Address) (*Delegation, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	if delegations, exists := sm.delegations[validator]; exists {
		delegation, exists := delegations[delegator]
		return delegation, exists
	}
	return nil, false
}

// GetStakingStatus returns overall staking status
func (sm *StakingManager) GetStakingStatus() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	activeCount := 0
	jailedCount := 0
	for _, validator := range sm.validators {
		switch validator.Status {
		case StatusActive:
			activeCount++
		case StatusJailed:
			jailedCount++
		}
	}
	
	return map[string]interface{}{
		"totalStaked":      sm.totalStaked.String(),
		"totalValidators":  len(sm.validators),
		"activeValidators": activeCount,
		"jailedValidators": jailedCount,
		"maxValidators":    sm.maxValidators,
		"minValidatorStake": sm.minValidatorStake.String(),
		"minDelegation":    sm.minDelegation.String(),
		"unbondingPeriod":  sm.unbondingPeriod.String(),
		"slashingEnabled":  sm.slashingEnabled,
	}
}