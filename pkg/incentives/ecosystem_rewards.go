package incentives

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// EcosystemRewards manages incentives to drive adoption
type EcosystemRewards struct {
	totalRewardPool       *big.Int
	developerRewardPool   *big.Int
	dappIncentivePool     *big.Int
	validatorBonusPool    *big.Int
	liquidityRewardPool   *big.Int
	
	// Tracking
	activeDApps           map[common.Address]*DAppMetrics
	developerContributions map[common.Address]*DeveloperMetrics
	liquidityProviders    map[common.Address]*LiquidityMetrics
}

// NewEcosystemRewards creates the incentive system
func NewEcosystemRewards(totalPool *big.Int) *EcosystemRewards {
	// Allocate reward pools
	developerPool := new(big.Int).Div(totalPool, big.NewInt(4))     // 25% for developers
	dappPool := new(big.Int).Div(totalPool, big.NewInt(4))         // 25% for DApp rewards
	validatorPool := new(big.Int).Div(totalPool, big.NewInt(4))    // 25% for validator bonuses
	liquidityPool := new(big.Int).Sub(totalPool, developerPool)    // Remaining 25% for liquidity
	liquidityPool.Sub(liquidityPool, dappPool)
	liquidityPool.Sub(liquidityPool, validatorPool)
	
	return &EcosystemRewards{
		totalRewardPool:       totalPool,
		developerRewardPool:   developerPool,
		dappIncentivePool:     dappPool,
		validatorBonusPool:    validatorPool,
		liquidityRewardPool:   liquidityPool,
		activeDApps:           make(map[common.Address]*DAppMetrics),
		developerContributions: make(map[common.Address]*DeveloperMetrics),
		liquidityProviders:    make(map[common.Address]*LiquidityMetrics),
	}
}

// 🚀 DEVELOPER ATTRACTION INCENTIVES

// RegisterDeveloper rewards developers for building on the platform
func (er *EcosystemRewards) RegisterDeveloper(developer common.Address, githubProfile string) *DeveloperReward {
	// Onboarding bonus for new developers
	onboardingBonus := new(big.Int).Mul(big.NewInt(1000), big.NewInt(1e18)) // 1000 LIGHT
	
	er.developerContributions[developer] = &DeveloperMetrics{
		Developer:       developer,
		GithubProfile:   githubProfile,
		JoinedAt:        time.Now(),
		OnboardingBonus: onboardingBonus,
		TotalRewards:    onboardingBonus,
		ContractsDeployed: 0,
		TotalTVL:        big.NewInt(0),
		CommunityScore:  0,
	}
	
	return &DeveloperReward{
		Developer:    developer,
		RewardType:   "onboarding",
		Amount:       onboardingBonus,
		Description:  "Welcome to LightChain L1! Build the future of DeFi",
		VestingPeriod: 30 * 24 * time.Hour, // 30 day vesting
	}
}

// RewardContractDeployment provides massive rewards for deploying useful contracts
func (er *EcosystemRewards) RewardContractDeployment(developer, contract common.Address, category ContractCategory) *DeveloperReward {
	metrics := er.developerContributions[developer]
	if metrics == nil {
		return nil
	}
	
	// Base reward based on contract type
	var baseReward *big.Int
	var description string
	
	switch category {
	case CategoryDeFi:
		baseReward = new(big.Int).Mul(big.NewInt(10000), big.NewInt(1e18)) // 10,000 LIGHT
		description = "DeFi Protocol Deployment - Building the financial future!"
	case CategoryNFT:
		baseReward = new(big.Int).Mul(big.NewInt(5000), big.NewInt(1e18)) // 5,000 LIGHT
		description = "NFT Collection Deployment - Creating digital assets!"
	case CategoryGameFi:
		baseReward = new(big.Int).Mul(big.NewInt(15000), big.NewInt(1e18)) // 15,000 LIGHT
		description = "GameFi Protocol - Gaming meets DeFi!"
	case CategoryDAO:
		baseReward = new(big.Int).Mul(big.NewInt(8000), big.NewInt(1e18)) // 8,000 LIGHT
		description = "DAO Governance - Decentralized decision making!"
	case CategoryInfrastructure:
		baseReward = new(big.Int).Mul(big.NewInt(20000), big.NewInt(1e18)) // 20,000 LIGHT
		description = "Infrastructure Tool - Building the ecosystem!"
	default:
		baseReward = new(big.Int).Mul(big.NewInt(2000), big.NewInt(1e18)) // 2,000 LIGHT
		description = "Smart Contract Deployment"
	}
	
	// Bonus for early adopters (first 100 contracts get 2x reward)
	if len(er.activeDApps) < 100 {
		baseReward.Mul(baseReward, big.NewInt(2))
		description += " + Early Adopter Bonus!"
	}
	
	// Update metrics
	metrics.ContractsDeployed++
	metrics.TotalRewards.Add(metrics.TotalRewards, baseReward)
	
	// Track the DApp
	er.activeDApps[contract] = &DAppMetrics{
		ContractAddress: contract,
		Developer:      developer,
		Category:       category,
		DeployedAt:     time.Now(),
		TVL:           big.NewInt(0),
		Users:         0,
		Transactions:  0,
		Revenue:       big.NewInt(0),
	}
	
	return &DeveloperReward{
		Developer:     developer,
		RewardType:    "contract_deployment",
		Amount:        baseReward,
		Description:   description,
		VestingPeriod: 7 * 24 * time.Hour, // 7 day vesting
	}
}

// 💰 DAPP SUCCESS INCENTIVES

// RewardDAppMilestone rewards DApps for reaching user/TVL milestones
func (er *EcosystemRewards) RewardDAppMilestone(contract common.Address, milestone MilestoneType, value *big.Int) *DAppReward {
	dapp := er.activeDApps[contract]
	if dapp == nil {
		return nil
	}
	
	var reward *big.Int
	var description string
	
	switch milestone {
	case MilestoneUsers:
		switch {
		case value.Cmp(big.NewInt(1000)) >= 0: // 1000 users
			reward = new(big.Int).Mul(big.NewInt(50000), big.NewInt(1e18))
			description = "🎉 1000 Users Milestone! Your DApp is gaining traction!"
		case value.Cmp(big.NewInt(10000)) >= 0: // 10,000 users
			reward = new(big.Int).Mul(big.NewInt(200000), big.NewInt(1e18))
			description = "🚀 10K Users! You're building something big!"
		case value.Cmp(big.NewInt(100000)) >= 0: // 100,000 users
			reward = new(big.Int).Mul(big.NewInt(1000000), big.NewInt(1e18))
			description = "💎 100K Users! You're a ecosystem leader!"
		}
	case MilestoneTVL:
		switch {
		case value.Cmp(new(big.Int).Mul(big.NewInt(1000000), big.NewInt(1e18))) >= 0: // $1M TVL
			reward = new(big.Int).Mul(big.NewInt(100000), big.NewInt(1e18))
			description = "💰 $1M TVL Milestone! Serious DeFi protocol!"
		case value.Cmp(new(big.Int).Mul(big.NewInt(10000000), big.NewInt(1e18))) >= 0: // $10M TVL
			reward = new(big.Int).Mul(big.NewInt(500000), big.NewInt(1e18))
			description = "🏆 $10M TVL! You're in the big leagues!"
		case value.Cmp(new(big.Int).Mul(big.NewInt(100000000), big.NewInt(1e18))) >= 0: // $100M TVL
			reward = new(big.Int).Mul(big.NewInt(2000000), big.NewInt(1e18))
			description = "👑 $100M TVL! Protocol legend status!"
		}
	case MilestoneRevenue:
		// Revenue sharing - 10% of protocol revenue as bonus
		reward = new(big.Int).Div(value, big.NewInt(10))
		description = fmt.Sprintf("💵 Revenue Milestone: %s LIGHT earned!", reward.String())
	}
	
	if reward == nil || reward.Sign() <= 0 {
		return nil
	}
	
	return &DAppReward{
		Contract:    contract,
		Developer:   dapp.Developer,
		RewardType:  string(milestone),
		Amount:      reward,
		Description: description,
	}
}

// 🔄 ONGOING ECOSYSTEM REWARDS

// DistributeMonthlyRewards distributes ongoing rewards to active ecosystem participants
func (er *EcosystemRewards) DistributeMonthlyRewards() []*MonthlyDistribution {
	distributions := []*MonthlyDistribution{}
	
	// Calculate total ecosystem activity
	totalTVL := big.NewInt(0)
	totalUsers := int64(0)
	totalTransactions := int64(0)
	
	for _, dapp := range er.activeDApps {
		totalTVL.Add(totalTVL, dapp.TVL)
		totalUsers += dapp.Users
		totalTransactions += dapp.Transactions
	}
	
	// Distribute rewards based on contribution
	monthlyPool := new(big.Int).Div(er.dappIncentivePool, big.NewInt(12)) // Monthly allocation
	
	for contractAddr, dapp := range er.activeDApps {
		// Calculate share based on TVL and activity
		tvlShare := new(big.Int).Div(dapp.TVL, totalTVL)
		if totalTVL.Sign() == 0 {
			tvlShare = big.NewInt(0)
		}
		
		userShare := big.NewInt(dapp.Users)
		if totalUsers > 0 {
			userShare.Div(userShare, big.NewInt(totalUsers))
		}
		
		// Weighted average (70% TVL, 30% users)
		tvlWeight := new(big.Int).Mul(tvlShare, big.NewInt(70))
		userWeight := new(big.Int).Mul(userShare, big.NewInt(30))
		totalWeight := new(big.Int).Add(tvlWeight, userWeight)
		totalWeight.Div(totalWeight, big.NewInt(100))
		
		// Calculate monthly reward
		monthlyReward := new(big.Int).Mul(monthlyPool, totalWeight)
		monthlyReward.Div(monthlyReward, big.NewInt(100))
		
		if monthlyReward.Sign() > 0 {
			distributions = append(distributions, &MonthlyDistribution{
				Contract:   contractAddr,
				Developer:  dapp.Developer,
				Amount:     monthlyReward,
				Reason:     fmt.Sprintf("Monthly ecosystem reward - %s TVL, %d users", dapp.TVL.String(), dapp.Users),
				Period:     time.Now().Format("2006-01"),
			})
		}
	}
	
	return distributions
}

// 🌉 CROSS-CHAIN BRIDGE INCENTIVES

// RewardBridgeUsage incentivizes cross-chain activity
func (er *EcosystemRewards) RewardBridgeUsage(user common.Address, amount *big.Int, sourceChain string) *BridgeReward {
	// 1% of bridged amount as reward (max 1000 LIGHT per transaction)
	rewardAmount := new(big.Int).Div(amount, big.NewInt(100))
	maxReward := new(big.Int).Mul(big.NewInt(1000), big.NewInt(1e18))
	
	if rewardAmount.Cmp(maxReward) > 0 {
		rewardAmount = maxReward
	}
	
	// Higher rewards for bridging from major chains
	multiplier := big.NewInt(1)
	switch sourceChain {
	case "ethereum":
		multiplier = big.NewInt(3) // 3x for Ethereum bridges
	case "polygon", "arbitrum", "optimism":
		multiplier = big.NewInt(2) // 2x for major L2s
	case "bsc", "avalanche":
		multiplier = big.NewInt(2) // 2x for major altchains
	}
	
	rewardAmount.Mul(rewardAmount, multiplier)
	
	return &BridgeReward{
		User:        user,
		Amount:      rewardAmount,
		SourceChain: sourceChain,
		BridgedAmount: amount,
		Description: fmt.Sprintf("Bridge reward: %s LIGHT for bridging from %s", rewardAmount.String(), sourceChain),
	}
}

// 🏆 VALIDATOR PERFORMANCE BONUSES

// CalculateValidatorBonus provides performance-based validator rewards
func (er *EcosystemRewards) CalculateValidatorBonus(validator common.Address, performance *ValidatorPerformance) *ValidatorBonus {
	// Base monthly validator reward
	baseReward := new(big.Int).Mul(big.NewInt(10000), big.NewInt(1e18)) // 10,000 LIGHT
	
	// Performance multipliers
	performanceMultiplier := big.NewInt(100) // Start at 1.0x
	
	// Uptime bonus (up to 1.5x)
	if performance.Uptime >= 0.99 {
		performanceMultiplier.Add(performanceMultiplier, big.NewInt(50)) // +0.5x
	} else if performance.Uptime >= 0.95 {
		performanceMultiplier.Add(performanceMultiplier, big.NewInt(25)) // +0.25x
	}
	
	// Block production bonus
	if performance.BlocksProduced > performance.ExpectedBlocks {
		excess := performance.BlocksProduced - performance.ExpectedBlocks
		bonusPercent := (excess * 100) / performance.ExpectedBlocks
		if bonusPercent > 50 {
			bonusPercent = 50 // Cap at 50% bonus
		}
		performanceMultiplier.Add(performanceMultiplier, big.NewInt(int64(bonusPercent)))
	}
	
	// Governance participation bonus
	if performance.GovernanceVotes > 0 {
		performanceMultiplier.Add(performanceMultiplier, big.NewInt(10)) // +0.1x for governance participation
	}
	
	// Calculate final reward
	finalReward := new(big.Int).Mul(baseReward, performanceMultiplier)
	finalReward.Div(finalReward, big.NewInt(100))
	
	return &ValidatorBonus{
		Validator:   validator,
		BaseReward:  baseReward,
		Multiplier:  float64(performanceMultiplier.Int64()) / 100.0,
		FinalReward: finalReward,
		Performance: performance,
		Description: fmt.Sprintf("Validator performance bonus: %.2fx multiplier", float64(performanceMultiplier.Int64())/100.0),
	}
}

// Data structures for tracking metrics and rewards
type ContractCategory string
type MilestoneType string

const (
	CategoryDeFi           ContractCategory = "defi"
	CategoryNFT            ContractCategory = "nft"
	CategoryGameFi         ContractCategory = "gamefi"
	CategoryDAO            ContractCategory = "dao"
	CategoryInfrastructure ContractCategory = "infrastructure"
	CategoryOther          ContractCategory = "other"
	
	MilestoneUsers   MilestoneType = "users"
	MilestoneTVL     MilestoneType = "tvl"
	MilestoneRevenue MilestoneType = "revenue"
)

type DeveloperMetrics struct {
	Developer         common.Address
	GithubProfile     string
	JoinedAt          time.Time
	ContractsDeployed int
	TotalTVL          *big.Int
	TotalRewards      *big.Int
	OnboardingBonus   *big.Int
	CommunityScore    int
}

type DAppMetrics struct {
	ContractAddress common.Address
	Developer       common.Address
	Category        ContractCategory
	DeployedAt      time.Time
	TVL             *big.Int
	Users           int64
	Transactions    int64
	Revenue         *big.Int
}

type LiquidityMetrics struct {
	Provider      common.Address
	TotalLiquidity *big.Int
	RewardsEarned  *big.Int
	StartDate      time.Time
}

type DeveloperReward struct {
	Developer     common.Address
	RewardType    string
	Amount        *big.Int
	Description   string
	VestingPeriod time.Duration
}

type DAppReward struct {
	Contract    common.Address
	Developer   common.Address
	RewardType  string
	Amount      *big.Int
	Description string
}

type MonthlyDistribution struct {
	Contract  common.Address
	Developer common.Address
	Amount    *big.Int
	Reason    string
	Period    string
}

type BridgeReward struct {
	User          common.Address
	Amount        *big.Int
	SourceChain   string
	BridgedAmount *big.Int
	Description   string
}

type ValidatorPerformance struct {
	Uptime          float64
	BlocksProduced  int64
	ExpectedBlocks  int64
	GovernanceVotes int64
	SlashingEvents  int64
}

type ValidatorBonus struct {
	Validator   common.Address
	BaseReward  *big.Int
	Multiplier  float64
	FinalReward *big.Int
	Performance *ValidatorPerformance
	Description string
}

// GetEcosystemStatus returns overall ecosystem health metrics
func (er *EcosystemRewards) GetEcosystemStatus() *EcosystemStatus {
	totalDApps := len(er.activeDApps)
	totalDevelopers := len(er.developerContributions)
	
	totalTVL := big.NewInt(0)
	totalUsers := int64(0)
	totalTransactions := int64(0)
	
	for _, dapp := range er.activeDApps {
		totalTVL.Add(totalTVL, dapp.TVL)
		totalUsers += dapp.Users
		totalTransactions += dapp.Transactions
	}
	
	totalRewardsDistributed := big.NewInt(0)
	for _, dev := range er.developerContributions {
		totalRewardsDistributed.Add(totalRewardsDistributed, dev.TotalRewards)
	}
	
	return &EcosystemStatus{
		TotalDApps:              totalDApps,
		TotalDevelopers:         totalDevelopers,
		TotalTVL:               totalTVL,
		TotalUsers:             totalUsers,
		TotalTransactions:      totalTransactions,
		TotalRewardsDistributed: totalRewardsDistributed,
		RemainingRewardPool:    new(big.Int).Sub(er.totalRewardPool, totalRewardsDistributed),
	}
}

type EcosystemStatus struct {
	TotalDApps              int
	TotalDevelopers         int
	TotalTVL               *big.Int
	TotalUsers             int64
	TotalTransactions      int64
	TotalRewardsDistributed *big.Int
	RemainingRewardPool    *big.Int
}