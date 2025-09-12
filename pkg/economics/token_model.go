package economics

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// LightToken represents the native L1 token (LIGHT)
// Innovation: Deflationary token with performance-based rewards
type LightToken struct {
	// Token properties
	name         string
	symbol       string
	decimals     uint8
	totalSupply  *big.Int
	maxSupply    *big.Int
	
	// Economic parameters
	inflationRate     *big.Int // Annual inflation rate (basis points)
	deflationRate     *big.Int // Annual deflation rate (basis points)
	stakingAPY        *big.Int // Base staking APY (basis points)
	performanceBonus  *big.Int // Performance bonus multiplier (basis points)
	
	// Balances and allowances
	balances   map[common.Address]*big.Int
	allowances map[common.Address]map[common.Address]*big.Int
	
	// Staking data
	stakedBalances map[common.Address]*StakePosition
	totalStaked    *big.Int
	
	// Token economics
	burned        *big.Int
	minted        *big.Int
	lastReward    time.Time
	rewardPool    *big.Int
	
	// Governance
	governanceToken bool
	votingPower     map[common.Address]*big.Int
	
	// Synchronization
	mu sync.RWMutex
}

// StakePosition represents a staking position
type StakePosition struct {
	Amount       *big.Int
	StakedAt     time.Time
	LastReward   time.Time
	Performance  float64
	LockPeriod   time.Duration
	RewardDebt   *big.Int
	ValidatorID  common.Address
}

// GasModel represents the gas pricing model
// Innovation: Dynamic gas pricing based on network load and validator performance
type GasModel struct {
	// Base gas parameters
	baseGasPrice    *big.Int
	minGasPrice     *big.Int
	maxGasPrice     *big.Int
	gasTarget       uint64
	gasLimit        uint64
	
	// Dynamic pricing
	loadMultiplier     float64
	performanceDiscount float64
	priorityMultiplier  float64
	
	// Gas tracking
	recentBlocks    []BlockGasData
	averageGasUsed  uint64
	
	// Economic incentives
	validatorGasShare *big.Int // Percentage of gas fees going to validators (basis points)
	burnRatio         *big.Int // Percentage of gas fees burned (basis points)
	treasuryRatio     *big.Int // Percentage going to treasury (basis points)
	
	mu sync.RWMutex
}

// BlockGasData tracks gas usage in a block
type BlockGasData struct {
	BlockHeight uint64
	GasUsed     uint64
	GasLimit    uint64
	BaseFee     *big.Int
	Timestamp   time.Time
}

// TokenEconomics manages the overall economic model
type TokenEconomics struct {
	token    *LightToken
	gasModel *GasModel
	
	// Economic state
	treasury       *big.Int
	rewardPool     *big.Int
	burnedTokens   *big.Int
	
	// Parameters
	blockReward         *big.Int
	halvingInterval     uint64
	nextHalving         uint64
	transactionFeeRatio *big.Int
	
	mu sync.RWMutex
}

const (
	// Token constants
	InitialSupply  = "1000000000" // 1 billion LIGHT tokens
	MaxSupply      = "2100000000" // 2.1 billion max (similar to Bitcoin but larger)
	TokenDecimals  = 18
	
	// Economic constants (basis points)
	BaseInflationRate   = 200  // 2% annual
	BaseDeflationRate   = 100  // 1% annual (from burning)
	BaseStakingAPY      = 800  // 8% base APY
	MaxPerformanceBonus = 500  // 5% max performance bonus
	
	// Gas constants
	InitialGasPrice = 1000000000 // 1 Gwei
	MinGasPrice     = 100000000  // 0.1 Gwei
	MaxGasPrice     = 100000000000 // 100 Gwei
	TargetGasUsage  = 15000000   // 15M gas per block
	MaxGasLimit     = 30000000   // 30M gas per block
)

// NewLightToken creates a new LIGHT token instance
func NewLightToken() *LightToken {
	initialSupply, _ := new(big.Int).SetString(InitialSupply+"000000000000000000", 10) // Add 18 decimals
	maxSupply, _ := new(big.Int).SetString(MaxSupply+"000000000000000000", 10)
	
	return &LightToken{
		name:              "LightChain",
		symbol:            "LIGHT",
		decimals:          TokenDecimals,
		totalSupply:       initialSupply,
		maxSupply:         maxSupply,
		inflationRate:     big.NewInt(BaseInflationRate),
		deflationRate:     big.NewInt(BaseDeflationRate),
		stakingAPY:        big.NewInt(BaseStakingAPY),
		performanceBonus:  big.NewInt(MaxPerformanceBonus),
		balances:          make(map[common.Address]*big.Int),
		allowances:        make(map[common.Address]map[common.Address]*big.Int),
		stakedBalances:    make(map[common.Address]*StakePosition),
		totalStaked:       big.NewInt(0),
		burned:            big.NewInt(0),
		minted:            big.NewInt(0),
		lastReward:        time.Now(),
		rewardPool:        big.NewInt(0),
		governanceToken:   true,
		votingPower:       make(map[common.Address]*big.Int),
	}
}

// NewGasModel creates a new gas pricing model
func NewGasModel() *GasModel {
	return &GasModel{
		baseGasPrice:        big.NewInt(InitialGasPrice),
		minGasPrice:         big.NewInt(MinGasPrice),
		maxGasPrice:         big.NewInt(MaxGasPrice),
		gasTarget:           TargetGasUsage,
		gasLimit:            MaxGasLimit,
		loadMultiplier:      1.0,
		performanceDiscount: 0.1, // 10% discount for high-performance validators
		priorityMultiplier:  1.5, // 50% premium for priority transactions
		recentBlocks:        make([]BlockGasData, 0, 100),
		validatorGasShare:   big.NewInt(6000), // 60%
		burnRatio:           big.NewInt(2000), // 20%
		treasuryRatio:       big.NewInt(2000), // 20%
	}
}

// NewTokenEconomics creates a new token economics system
func NewTokenEconomics() *TokenEconomics {
	blockReward, _ := new(big.Int).SetString("5000000000000000000", 10) // 5 LIGHT per block initially
	
	return &TokenEconomics{
		token:               NewLightToken(),
		gasModel:           NewGasModel(),
		treasury:           big.NewInt(0),
		rewardPool:         big.NewInt(0),
		burnedTokens:       big.NewInt(0),
		blockReward:        blockReward,
		halvingInterval:    2100000, // Approximately 4 years at 2s blocks
		nextHalving:        2100000,
		transactionFeeRatio: big.NewInt(10), // 0.1% of transaction value
	}
}

// Transfer transfers tokens between addresses
func (lt *LightToken) Transfer(from, to common.Address, amount *big.Int) error {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	
	if amount.Sign() <= 0 {
		return fmt.Errorf("transfer amount must be positive")
	}
	
	fromBalance := lt.getBalance(from)
	if fromBalance.Cmp(amount) < 0 {
		return fmt.Errorf("insufficient balance")
	}
	
	// Update balances
	lt.balances[from] = new(big.Int).Sub(fromBalance, amount)
	toBalance := lt.getBalance(to)
	lt.balances[to] = new(big.Int).Add(toBalance, amount)
	
	return nil
}

// Stake stakes tokens for validation rewards
func (lt *LightToken) Stake(staker common.Address, amount *big.Int, validatorID common.Address, lockPeriod time.Duration) error {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	
	if amount.Sign() <= 0 {
		return fmt.Errorf("stake amount must be positive")
	}
	
	balance := lt.getBalance(staker)
	if balance.Cmp(amount) < 0 {
		return fmt.Errorf("insufficient balance to stake")
	}
	
	// Transfer to staked balance
	lt.balances[staker] = new(big.Int).Sub(balance, amount)
	
	// Create or update stake position
	if existing, exists := lt.stakedBalances[staker]; exists {
		existing.Amount.Add(existing.Amount, amount)
		existing.StakedAt = time.Now()
	} else {
		lt.stakedBalances[staker] = &StakePosition{
			Amount:      new(big.Int).Set(amount),
			StakedAt:    time.Now(),
			LastReward:  time.Now(),
			Performance: 1.0,
			LockPeriod:  lockPeriod,
			RewardDebt:  big.NewInt(0),
			ValidatorID: validatorID,
		}
	}
	
	lt.totalStaked.Add(lt.totalStaked, amount)
	
	// Update voting power for governance
	if lt.governanceToken {
		votePower := lt.getVotingPower(staker)
		lt.votingPower[staker] = new(big.Int).Add(votePower, amount)
	}
	
	return nil
}

// Unstake unstakes tokens (with lock period check)
func (lt *LightToken) Unstake(staker common.Address, amount *big.Int) error {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	
	position, exists := lt.stakedBalances[staker]
	if !exists {
		return fmt.Errorf("no stake position found")
	}
	
	if position.Amount.Cmp(amount) < 0 {
		return fmt.Errorf("insufficient staked amount")
	}
	
	// Check lock period
	if time.Since(position.StakedAt) < position.LockPeriod {
		return fmt.Errorf("stake is still locked")
	}
	
	// Update stake position
	position.Amount.Sub(position.Amount, amount)
	if position.Amount.Sign() == 0 {
		delete(lt.stakedBalances, staker)
	}
	
	// Return tokens to balance
	balance := lt.getBalance(staker)
	lt.balances[staker] = new(big.Int).Add(balance, amount)
	
	lt.totalStaked.Sub(lt.totalStaked, amount)
	
	// Update voting power
	if lt.governanceToken {
		votePower := lt.getVotingPower(staker)
		lt.votingPower[staker] = new(big.Int).Sub(votePower, amount)
	}
	
	return nil
}

// CalculateStakingRewards calculates rewards for a staker
func (lt *LightToken) CalculateStakingRewards(staker common.Address) (*big.Int, error) {
	lt.mu.RLock()
	defer lt.mu.RUnlock()
	
	position, exists := lt.stakedBalances[staker]
	if !exists {
		return big.NewInt(0), fmt.Errorf("no stake position found")
	}
	
	timeSinceLastReward := time.Since(position.LastReward)
	if timeSinceLastReward < time.Hour {
		return big.NewInt(0), nil // Rewards calculated hourly
	}
	
	// Calculate base reward
	annualReward := new(big.Int).Mul(position.Amount, lt.stakingAPY)
	annualReward.Div(annualReward, big.NewInt(10000)) // Convert from basis points
	
	// Apply performance bonus
	performanceMultiplier := big.NewInt(int64(position.Performance * 10000))
	bonusReward := new(big.Int).Mul(annualReward, lt.performanceBonus)
	bonusReward.Mul(bonusReward, performanceMultiplier)
	bonusReward.Div(bonusReward, big.NewInt(100000000)) // Normalize
	
	totalAnnualReward := new(big.Int).Add(annualReward, bonusReward)
	
	// Calculate reward for time period
	secondsInYear := big.NewInt(365 * 24 * 60 * 60)
	timeReward := new(big.Int).Mul(totalAnnualReward, big.NewInt(int64(timeSinceLastReward.Seconds())))
	timeReward.Div(timeReward, secondsInYear)
	
	return timeReward, nil
}

// DistributeStakingRewards distributes rewards to a staker
func (lt *LightToken) DistributeStakingRewards(staker common.Address) error {
	reward, err := lt.CalculateStakingRewards(staker)
	if err != nil {
		return err
	}
	
	if reward.Sign() <= 0 {
		return nil
	}
	
	lt.mu.Lock()
	defer lt.mu.Unlock()
	
	// Mint new tokens for reward (controlled inflation)
	if new(big.Int).Add(lt.totalSupply, reward).Cmp(lt.maxSupply) <= 0 {
		balance := lt.getBalance(staker)
		lt.balances[staker] = new(big.Int).Add(balance, reward)
		lt.totalSupply.Add(lt.totalSupply, reward)
		lt.minted.Add(lt.minted, reward)
		
		// Update last reward time
		if position, exists := lt.stakedBalances[staker]; exists {
			position.LastReward = time.Now()
		}
	}
	
	return nil
}

// Burn burns tokens (deflationary mechanism)
func (lt *LightToken) Burn(amount *big.Int) {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	
	if amount.Sign() > 0 && lt.totalSupply.Cmp(amount) >= 0 {
		lt.totalSupply.Sub(lt.totalSupply, amount)
		lt.burned.Add(lt.burned, amount)
	}
}

// CalculateGasPrice calculates dynamic gas price
func (gm *GasModel) CalculateGasPrice(blockHeight uint64, networkLoad float64, validatorPerformance float64, isPriority bool) *big.Int {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	
	basePrice := new(big.Int).Set(gm.baseGasPrice)
	
	// Apply network load multiplier
	loadAdjustment := big.NewInt(int64(networkLoad * 10000))
	loadPrice := new(big.Int).Mul(basePrice, loadAdjustment)
	loadPrice.Div(loadPrice, big.NewInt(10000))
	
	// Apply validator performance discount
	if validatorPerformance > 0.9 {
		discount := big.NewInt(int64(gm.performanceDiscount * 10000))
		discountAmount := new(big.Int).Mul(loadPrice, discount)
		discountAmount.Div(discountAmount, big.NewInt(10000))
		loadPrice.Sub(loadPrice, discountAmount)
	}
	
	// Apply priority multiplier
	if isPriority {
		priorityBonus := big.NewInt(int64(gm.priorityMultiplier * 10000))
		bonusAmount := new(big.Int).Mul(loadPrice, priorityBonus)
		bonusAmount.Div(bonusAmount, big.NewInt(10000))
		loadPrice.Add(loadPrice, bonusAmount)
	}
	
	// Ensure within bounds
	if loadPrice.Cmp(gm.minGasPrice) < 0 {
		loadPrice.Set(gm.minGasPrice)
	}
	if loadPrice.Cmp(gm.maxGasPrice) > 0 {
		loadPrice.Set(gm.maxGasPrice)
	}
	
	return loadPrice
}

// ProcessBlockRewards processes block rewards for validators
func (te *TokenEconomics) ProcessBlockRewards(blockHeight uint64, validator common.Address, gasFeesCollected *big.Int) {
	te.mu.Lock()
	defer te.mu.Unlock()
	
	// Check for halving
	if blockHeight >= te.nextHalving {
		te.blockReward.Div(te.blockReward, big.NewInt(2))
		te.nextHalving += te.halvingInterval
	}
	
	// Distribute block reward
	validatorBalance := te.token.getBalance(validator)
	te.token.balances[validator] = new(big.Int).Add(validatorBalance, te.blockReward)
	te.token.totalSupply.Add(te.token.totalSupply, te.blockReward)
	te.token.minted.Add(te.token.minted, te.blockReward)
	
	// Process gas fees
	if gasFeesCollected.Sign() > 0 {
		te.distributeGasFees(gasFeesCollected, validator)
	}
}

// distributeGasFees distributes gas fees according to the economic model
func (te *TokenEconomics) distributeGasFees(totalFees *big.Int, validator common.Address) {
	// Validator share
	validatorShare := new(big.Int).Mul(totalFees, te.gasModel.validatorGasShare)
	validatorShare.Div(validatorShare, big.NewInt(10000))
	validatorBalance := te.token.getBalance(validator)
	te.token.balances[validator] = new(big.Int).Add(validatorBalance, validatorShare)
	
	// Burn share
	burnShare := new(big.Int).Mul(totalFees, te.gasModel.burnRatio)
	burnShare.Div(burnShare, big.NewInt(10000))
	te.token.Burn(burnShare)
	te.burnedTokens.Add(te.burnedTokens, burnShare)
	
	// Treasury share
	treasuryShare := new(big.Int).Mul(totalFees, te.gasModel.treasuryRatio)
	treasuryShare.Div(treasuryShare, big.NewInt(10000))
	te.treasury.Add(te.treasury, treasuryShare)
}

// Helper methods
func (lt *LightToken) getBalance(address common.Address) *big.Int {
	if balance, exists := lt.balances[address]; exists {
		return new(big.Int).Set(balance)
	}
	return big.NewInt(0)
}

func (lt *LightToken) getVotingPower(address common.Address) *big.Int {
	if power, exists := lt.votingPower[address]; exists {
		return new(big.Int).Set(power)
	}
	return big.NewInt(0)
}

// Getters for external access
func (lt *LightToken) GetBalance(address common.Address) *big.Int {
	lt.mu.RLock()
	defer lt.mu.RUnlock()
	return lt.getBalance(address)
}

func (lt *LightToken) GetTotalSupply() *big.Int {
	lt.mu.RLock()
	defer lt.mu.RUnlock()
	return new(big.Int).Set(lt.totalSupply)
}

func (lt *LightToken) GetStakedAmount(address common.Address) *big.Int {
	lt.mu.RLock()
	defer lt.mu.RUnlock()
	
	if position, exists := lt.stakedBalances[address]; exists {
		return new(big.Int).Set(position.Amount)
	}
	return big.NewInt(0)
}

func (te *TokenEconomics) GetEconomicStatus() map[string]interface{} {
	te.mu.RLock()
	defer te.mu.RUnlock()
	
	return map[string]interface{}{
		"totalSupply":    te.token.totalSupply.String(),
		"maxSupply":      te.token.maxSupply.String(),
		"totalStaked":    te.token.totalStaked.String(),
		"burned":         te.token.burned.String(),
		"minted":         te.token.minted.String(),
		"treasury":       te.treasury.String(),
		"blockReward":    te.blockReward.String(),
		"nextHalving":    te.nextHalving,
		"baseGasPrice":   te.gasModel.baseGasPrice.String(),
		"inflationRate":  te.token.inflationRate.String(),
		"deflationRate":  te.token.deflationRate.String(),
	}
}