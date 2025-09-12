package genesis

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// L1Genesis represents the L1 blockchain genesis configuration
// Innovation: Bootstraps with staking contracts and tokenomics built-in
type L1Genesis struct {
	ChainID     *big.Int                    `json:"chainId"`
	Timestamp   uint64                      `json:"timestamp"`
	ExtraData   hexutil.Bytes              `json:"extraData"`
	GasLimit    uint64                      `json:"gasLimit"`
	Difficulty  *big.Int                    `json:"difficulty"`
	Config      *ChainConfig               `json:"config"`
	Alloc       map[common.Address]GenesisAccount `json:"alloc"`
	
	// L1-specific fields
	Validators  []GenesisValidator         `json:"validators"`
	Token       TokenGenesis              `json:"token"`
	Governance  GovernanceGenesis         `json:"governance"`
	Economics   EconomicsGenesis          `json:"economics"`
	
	// Computed fields
	Hash        common.Hash               `json:"hash,omitempty"`
	Number      uint64                    `json:"number"`
	ParentHash  common.Hash               `json:"parentHash"`
	StateRoot   common.Hash               `json:"stateRoot"`
	ReceiptRoot common.Hash               `json:"receiptRoot"`
}

// ChainConfig contains the chain configuration for the L1
type ChainConfig struct {
	ChainID             *big.Int `json:"chainId"`
	HomesteadBlock      *big.Int `json:"homesteadBlock"`
	EIP150Block         *big.Int `json:"eip150Block"`
	EIP155Block         *big.Int `json:"eip155Block"`
	EIP158Block         *big.Int `json:"eip158Block"`
	ByzantiumBlock      *big.Int `json:"byzantiumBlock"`
	ConstantinopleBlock *big.Int `json:"constantinopleBlock"`
	PetersburgBlock     *big.Int `json:"petersburgBlock"`
	IstanbulBlock       *big.Int `json:"istanbulBlock"`
	BerlinBlock         *big.Int `json:"berlinBlock"`
	LondonBlock         *big.Int `json:"londonBlock"`
	
	// L1-specific config
	HPoSBlock          *big.Int    `json:"hposBlock"`          // Block to activate HPoS consensus
	StakingBlock       *big.Int    `json:"stakingBlock"`       // Block to enable staking
	GovernanceBlock    *big.Int    `json:"governanceBlock"`    // Block to enable governance
	MaxValidators      int         `json:"maxValidators"`      // Maximum validators
	MinStake           *big.Int    `json:"minStake"`           // Minimum stake amount
	SlashingEnabled    bool        `json:"slashingEnabled"`    // Enable slashing
}

// GenesisAccount represents a genesis account allocation
type GenesisAccount struct {
	Balance *big.Int                       `json:"balance"`
	Code    hexutil.Bytes                  `json:"code,omitempty"`
	Storage map[common.Hash]common.Hash    `json:"storage,omitempty"`
	Nonce   uint64                         `json:"nonce,omitempty"`
}

// GenesisValidator represents a genesis validator
type GenesisValidator struct {
	Address     common.Address  `json:"address"`
	PubKey      hexutil.Bytes   `json:"pubkey"`
	Stake       *big.Int        `json:"stake"`
	Commission  uint64          `json:"commission"`  // Commission rate in basis points
	Active      bool            `json:"active"`
	Description ValidatorDesc   `json:"description"`
}

// ValidatorDesc contains validator description
type ValidatorDesc struct {
	Moniker         string `json:"moniker"`
	Identity        string `json:"identity"`
	Website         string `json:"website"`
	SecurityContact string `json:"security_contact"`
	Details         string `json:"details"`
}

// TokenGenesis contains genesis token configuration
type TokenGenesis struct {
	Name            string   `json:"name"`
	Symbol          string   `json:"symbol"`
	Decimals        uint8    `json:"decimals"`
	TotalSupply     *big.Int `json:"totalSupply"`
	MaxSupply       *big.Int `json:"maxSupply"`
	InitialInflation *big.Int `json:"initialInflation"`
	StakingRewards  *big.Int `json:"stakingRewards"`
}

// GovernanceGenesis contains governance parameters
type GovernanceGenesis struct {
	VotingPeriod        uint64   `json:"votingPeriod"`        // Blocks
	VotingDelay         uint64   `json:"votingDelay"`         // Blocks
	ProposalThreshold   *big.Int `json:"proposalThreshold"`   // Min tokens to propose
	QuorumThreshold     *big.Int `json:"quorumThreshold"`     // Min participation
	SuperMajority       uint64   `json:"superMajority"`       // Basis points for passing
	ExecutionDelay      uint64   `json:"executionDelay"`      // Blocks before execution
	GracePeriod         uint64   `json:"gracePeriod"`         // Blocks for execution
}

// EconomicsGenesis contains economic parameters
type EconomicsGenesis struct {
	BlockReward         *big.Int `json:"blockReward"`
	HalvingInterval     uint64   `json:"halvingInterval"`
	MinGasPrice         *big.Int `json:"minGasPrice"`
	MaxGasPrice         *big.Int `json:"maxGasPrice"`
	TargetGasUsage      uint64   `json:"targetGasUsage"`
	ValidatorFeeShare   uint64   `json:"validatorFeeShare"`   // Basis points
	BurnRatio           uint64   `json:"burnRatio"`           // Basis points
	TreasuryRatio       uint64   `json:"treasuryRatio"`       // Basis points
}

// GenesisBuilder helps construct L1 genesis blocks
type GenesisBuilder struct {
	genesis *L1Genesis
}

// NewGenesisBuilder creates a new genesis builder
func NewGenesisBuilder(chainID *big.Int) *GenesisBuilder {
	return &GenesisBuilder{
		genesis: &L1Genesis{
			ChainID:     chainID,
			Timestamp:   uint64(time.Now().Unix()),
			GasLimit:    30000000,
			Difficulty:  big.NewInt(1),
			Number:      0,
			ParentHash:  common.Hash{},
			Alloc:       make(map[common.Address]GenesisAccount),
			Validators:  make([]GenesisValidator, 0),
			ExtraData:   []byte("LightChain L1 Genesis"),
		},
	}
}

// DefaultL1Genesis creates a default L1 genesis configuration
func DefaultL1Genesis(chainID *big.Int) *L1Genesis {
	builder := NewGenesisBuilder(chainID)
	
	// Set chain configuration
	builder.SetChainConfig(&ChainConfig{
		ChainID:             chainID,
		HomesteadBlock:      big.NewInt(0),
		EIP150Block:         big.NewInt(0),
		EIP155Block:         big.NewInt(0),
		EIP158Block:         big.NewInt(0),
		ByzantiumBlock:      big.NewInt(0),
		ConstantinopleBlock: big.NewInt(0),
		PetersburgBlock:     big.NewInt(0),
		IstanbulBlock:       big.NewInt(0),
		BerlinBlock:         big.NewInt(0),
		LondonBlock:         big.NewInt(0),
		HPoSBlock:           big.NewInt(0),
		StakingBlock:        big.NewInt(0),
		GovernanceBlock:     big.NewInt(100), // Enable governance after 100 blocks
		MaxValidators:       21,
		MinStake:            new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18)), // 100 LIGHT tokens
		SlashingEnabled:     true,
	})
	
	// Set token configuration
	totalSupply, _ := new(big.Int).SetString("1000000000000000000000000000", 10) // 1B tokens with 18 decimals
	maxSupply, _ := new(big.Int).SetString("2100000000000000000000000000", 10)   // 2.1B max supply
	
	builder.SetTokenGenesis(TokenGenesis{
		Name:             "LightChain",
		Symbol:           "LIGHT",
		Decimals:         18,
		TotalSupply:      totalSupply,
		MaxSupply:        maxSupply,
		InitialInflation: big.NewInt(200), // 2% basis points
		StakingRewards:   big.NewInt(800), // 8% basis points
	})
	
	// Set governance parameters
	builder.SetGovernanceGenesis(GovernanceGenesis{
		VotingPeriod:      43200,  // ~1 day at 2s blocks
		VotingDelay:       7200,   // ~4 hours
		ProposalThreshold: new(big.Int).Mul(big.NewInt(10000), big.NewInt(1e18)), // 10k LIGHT
		QuorumThreshold:   big.NewInt(3000), // 30% basis points
		SuperMajority:     6000,             // 60% basis points
		ExecutionDelay:    14400,            // ~8 hours
		GracePeriod:       43200,            // ~1 day
	})
	
	// Set economic parameters
	blockReward, _ := new(big.Int).SetString("5000000000000000000", 10) // 5 LIGHT per block
	
	builder.SetEconomicsGenesis(EconomicsGenesis{
		BlockReward:       blockReward,
		HalvingInterval:   2100000,         // ~4 years
		MinGasPrice:       big.NewInt(100000000),  // 0.1 Gwei
		MaxGasPrice:       big.NewInt(100000000000), // 100 Gwei
		TargetGasUsage:    15000000,        // 15M gas
		ValidatorFeeShare: 6000,            // 60%
		BurnRatio:         2000,            // 20%
		TreasuryRatio:     2000,            // 20%
	})
	
	// Add default validators for initial bootstrap
	builder.AddGenesisValidators()
	
	// Add foundation accounts
	builder.AddFoundationAccounts()
	
	return builder.Build()
}

// SetChainConfig sets the chain configuration
func (gb *GenesisBuilder) SetChainConfig(config *ChainConfig) *GenesisBuilder {
	gb.genesis.Config = config
	return gb
}

// SetTokenGenesis sets the token configuration
func (gb *GenesisBuilder) SetTokenGenesis(token TokenGenesis) *GenesisBuilder {
	gb.genesis.Token = token
	return gb
}

// SetGovernanceGenesis sets governance parameters
func (gb *GenesisBuilder) SetGovernanceGenesis(governance GovernanceGenesis) *GenesisBuilder {
	gb.genesis.Governance = governance
	return gb
}

// SetEconomicsGenesis sets economic parameters
func (gb *GenesisBuilder) SetEconomicsGenesis(economics EconomicsGenesis) *GenesisBuilder {
	gb.genesis.Economics = economics
	return gb
}

// AddAccount adds a genesis account allocation
func (gb *GenesisBuilder) AddAccount(address common.Address, balance *big.Int) *GenesisBuilder {
	gb.genesis.Alloc[address] = GenesisAccount{
		Balance: balance,
		Nonce:   0,
	}
	return gb
}

// AddContractAccount adds a genesis contract account
func (gb *GenesisBuilder) AddContractAccount(address common.Address, balance *big.Int, code []byte, storage map[common.Hash]common.Hash) *GenesisBuilder {
	gb.genesis.Alloc[address] = GenesisAccount{
		Balance: balance,
		Code:    code,
		Storage: storage,
		Nonce:   1, // Contracts start with nonce 1
	}
	return gb
}

// AddValidator adds a genesis validator
func (gb *GenesisBuilder) AddValidator(validator GenesisValidator) *GenesisBuilder {
	gb.genesis.Validators = append(gb.genesis.Validators, validator)
	return gb
}

// AddFoundationAccounts adds foundation/team accounts with vesting
func (gb *GenesisBuilder) AddFoundationAccounts() *GenesisBuilder {
	// FOUNDER ALLOCATION - Replace with your actual wallet address
	// This is where YOU as the blockchain founder will receive tokens
	founderAddr := common.HexToAddress("0xYOUR_WALLET_ADDRESS_HERE")  // <-- UPDATE THIS WITH YOUR WALLET
	founderBalance, _ := new(big.Int).SetString("200000000000000000000000000", 10) // 200M LIGHT to founder
	gb.AddAccount(founderAddr, founderBalance)
	
	// Team allocation (for future team members)
	teamAddr := common.HexToAddress("0x1000000000000000000000000000000000000002")
	teamBalance, _ := new(big.Int).SetString("50000000000000000000000000", 10) // 50M LIGHT
	gb.AddAccount(teamAddr, teamBalance)
	
	// Ecosystem fund (for partnerships and development)
	ecosystemAddr := common.HexToAddress("0x1000000000000000000000000000000000000003")
	ecosystemBalance, _ := new(big.Int).SetString("100000000000000000000000000", 10) // 100M LIGHT
	gb.AddAccount(ecosystemAddr, ecosystemBalance)
	
	// Treasury (for operational expenses)
	treasuryAddr := common.HexToAddress("0x1000000000000000000000000000000000000004")
	treasuryBalance, _ := new(big.Int).SetString("150000000000000000000000000", 10) // 150M LIGHT
	gb.AddAccount(treasuryAddr, treasuryBalance)
	
	return gb
}

// AddGenesisValidators adds initial validator set
func (gb *GenesisBuilder) AddGenesisValidators() *GenesisBuilder {
	// Add 5 genesis validators with equal stake
	genesisStake, _ := new(big.Int).SetString("1000000000000000000000", 10) // 1000 LIGHT
	
	validators := []struct {
		addr    string
		moniker string
	}{
		{"0x2000000000000000000000000000000000000001", "Genesis-1"},
		{"0x2000000000000000000000000000000000000002", "Genesis-2"},
		{"0x2000000000000000000000000000000000000003", "Genesis-3"},
		{"0x2000000000000000000000000000000000000004", "Genesis-4"},
		{"0x2000000000000000000000000000000000000005", "Genesis-5"},
	}
	
	for i, val := range validators {
		pubKey := make([]byte, 33) // Placeholder public key
		pubKey[0] = 0x02 // Compressed public key prefix
		for j := 1; j < 33; j++ {
			pubKey[j] = byte(i + 1) // Simple pattern for genesis
		}
		
		validator := GenesisValidator{
			Address:    common.HexToAddress(val.addr),
			PubKey:     pubKey,
			Stake:      new(big.Int).Set(genesisStake),
			Commission: 1000, // 10% commission
			Active:     true,
			Description: ValidatorDesc{
				Moniker:  val.moniker,
				Identity: fmt.Sprintf("genesis-validator-%d", i+1),
				Website:  "https://lightchain.org",
				Details:  "Genesis validator for LightChain L1",
			},
		}
		
		gb.AddValidator(validator)
		
		// Give validators some initial balance for operations
		validatorBalance, _ := new(big.Int).SetString("10000000000000000000000", 10) // 10k LIGHT
		gb.AddAccount(validator.Address, validatorBalance)
	}
	
	return gb
}

// SetTimestamp sets the genesis timestamp
func (gb *GenesisBuilder) SetTimestamp(timestamp uint64) *GenesisBuilder {
	gb.genesis.Timestamp = timestamp
	return gb
}

// SetExtraData sets genesis extra data
func (gb *GenesisBuilder) SetExtraData(data []byte) *GenesisBuilder {
	gb.genesis.ExtraData = data
	return gb
}

// Build finalizes and returns the genesis block
func (gb *GenesisBuilder) Build() *L1Genesis {
	// Calculate state root and receipt root based on allocations
	gb.genesis.StateRoot = gb.calculateStateRoot()
	gb.genesis.ReceiptRoot = common.Hash{} // Empty for genesis
	
	// Calculate genesis hash
	gb.genesis.Hash = gb.calculateGenesisHash()
	
	return gb.genesis
}

// calculateStateRoot calculates the state root from allocations
func (gb *GenesisBuilder) calculateStateRoot() common.Hash {
	// Simplified state root calculation
	// In production, this would use proper Merkle Patricia Trie
	hasher := sha256.New()
	
	// Hash all allocations
	for addr, account := range gb.genesis.Alloc {
		hasher.Write(addr.Bytes())
		hasher.Write(account.Balance.Bytes())
		if len(account.Code) > 0 {
			hasher.Write(account.Code)
		}
	}
	
	// Hash validators
	for _, validator := range gb.genesis.Validators {
		hasher.Write(validator.Address.Bytes())
		hasher.Write(validator.Stake.Bytes())
		hasher.Write(validator.PubKey)
	}
	
	return common.BytesToHash(hasher.Sum(nil))
}

// calculateGenesisHash calculates the genesis block hash
func (gb *GenesisBuilder) calculateGenesisHash() common.Hash {
	hasher := sha256.New()
	
	// Hash core genesis data
	hasher.Write(gb.genesis.ChainID.Bytes())
	hasher.Write(gb.genesis.ParentHash.Bytes())
	hasher.Write(gb.genesis.StateRoot.Bytes())
	hasher.Write(gb.genesis.ReceiptRoot.Bytes())
	hasher.Write(gb.genesis.ExtraData)
	
	// Add timestamp and gas limit
	timestampBytes := make([]byte, 8)
	for i := 0; i < 8; i++ {
		timestampBytes[i] = byte(gb.genesis.Timestamp >> (8 * (7 - i)))
	}
	hasher.Write(timestampBytes)
	
	gasLimitBytes := make([]byte, 8)
	for i := 0; i < 8; i++ {
		gasLimitBytes[i] = byte(gb.genesis.GasLimit >> (8 * (7 - i)))
	}
	hasher.Write(gasLimitBytes)
	
	return common.BytesToHash(hasher.Sum(nil))
}

// ToJSON exports genesis to JSON
func (genesis *L1Genesis) ToJSON() ([]byte, error) {
	return json.MarshalIndent(genesis, "", "  ")
}

// FromJSON imports genesis from JSON
func FromJSON(data []byte) (*L1Genesis, error) {
	var genesis L1Genesis
	if err := json.Unmarshal(data, &genesis); err != nil {
		return nil, fmt.Errorf("failed to unmarshal genesis: %w", err)
	}
	return &genesis, nil
}

// Validate validates the genesis configuration
func (genesis *L1Genesis) Validate() error {
	if genesis.ChainID == nil || genesis.ChainID.Sign() <= 0 {
		return fmt.Errorf("invalid chain ID")
	}
	
	if genesis.Config == nil {
		return fmt.Errorf("missing chain config")
	}
	
	if len(genesis.Validators) == 0 {
		return fmt.Errorf("no genesis validators")
	}
	
	// Validate validators
	totalStake := big.NewInt(0)
	for i, validator := range genesis.Validators {
		if validator.Stake == nil || validator.Stake.Sign() <= 0 {
			return fmt.Errorf("validator %d has invalid stake", i)
		}
		totalStake.Add(totalStake, validator.Stake)
		
		if validator.Commission > 10000 {
			return fmt.Errorf("validator %d has commission > 100%%", i)
		}
	}
	
	// Validate total supply
	if genesis.Token.TotalSupply == nil || genesis.Token.TotalSupply.Sign() <= 0 {
		return fmt.Errorf("invalid total supply")
	}
	
	if genesis.Token.MaxSupply.Cmp(genesis.Token.TotalSupply) < 0 {
		return fmt.Errorf("max supply less than initial supply")
	}
	
	return nil
}

// GetValidatorStakeTotal returns total validator stake
func (genesis *L1Genesis) GetValidatorStakeTotal() *big.Int {
	total := big.NewInt(0)
	for _, validator := range genesis.Validators {
		total.Add(total, validator.Stake)
	}
	return total
}

// GetAllocTotal returns total allocated tokens
func (genesis *L1Genesis) GetAllocTotal() *big.Int {
	total := big.NewInt(0)
	for _, account := range genesis.Alloc {
		total.Add(total, account.Balance)
	}
	return total
}

// Summary returns a summary of the genesis
func (genesis *L1Genesis) Summary() map[string]interface{} {
	return map[string]interface{}{
		"chainId":         genesis.ChainID.String(),
		"timestamp":       genesis.Timestamp,
		"hash":           genesis.Hash.Hex(),
		"validatorCount":  len(genesis.Validators),
		"accountCount":    len(genesis.Alloc),
		"totalStake":     genesis.GetValidatorStakeTotal().String(),
		"totalAlloc":     genesis.GetAllocTotal().String(),
		"tokenName":      genesis.Token.Name,
		"tokenSymbol":    genesis.Token.Symbol,
		"totalSupply":    genesis.Token.TotalSupply.String(),
		"maxSupply":      genesis.Token.MaxSupply.String(),
		"blockReward":    genesis.Economics.BlockReward.String(),
		"maxValidators":  genesis.Config.MaxValidators,
	}
}