package config

import (
	"fmt"
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the main configuration structure
type Config struct {
	NodeType string `yaml:"node_type"`
	DataDir  string `yaml:"data_dir"`
	LogLevel string `yaml:"log_level"`

	Network     NetworkConfig     `yaml:"network"`
	Consensus   ConsensusConfig   `yaml:"consensus"`
	State       StateConfig       `yaml:"state"`
	RPC         RPCConfig         `yaml:"rpc"`
	WebSocket   WebSocketConfig   `yaml:"websocket"`
	Metrics     MetricsConfig     `yaml:"metrics"`
	AggLayer    AggLayerConfig    `yaml:"agglayer"`
	Security    SecurityConfig    `yaml:"security"`
	Performance PerformanceConfig `yaml:"performance"`

	// Node-specific configurations
	Sequencer *SequencerConfig `yaml:"sequencer,omitempty"`
	Archive   *ArchiveConfig   `yaml:"archive,omitempty"`
	L1        *L1Config        `yaml:"l1,omitempty"`
}

// NetworkConfig contains P2P networking settings
type NetworkConfig struct {
	ListenAddr       string   `yaml:"listen_addr"`
	ExternalAddr     string   `yaml:"external_addr"`
	BootstrapNodes   []string `yaml:"bootstrap_nodes"`
	MaxPeers         int      `yaml:"max_peers"`
	DiscoveryEnabled bool     `yaml:"discovery_enabled"`
	NATEnabled       bool     `yaml:"nat_enabled"`
}

// ConsensusConfig contains consensus mechanism settings
type ConsensusConfig struct {
	Type        string        `yaml:"type"`
	BlockTime   time.Duration `yaml:"block_time"`
	EpochLength int           `yaml:"epoch_length"`

	Validator *ValidatorConfig `yaml:"validator,omitempty"`
	Sequencer *struct {
		Enabled              bool          `yaml:"enabled"`
		BatchSize            int           `yaml:"batch_size"`
		BatchTimeout         time.Duration `yaml:"batch_timeout"`
		L1SubmissionInterval time.Duration `yaml:"l1_submission_interval"`
	} `yaml:"sequencer,omitempty"`
	Archive *struct {
		Enabled           bool `yaml:"enabled"`
		FullSync          bool `yaml:"full_sync"`
		ServeLightClients bool `yaml:"serve_light_clients"`
	} `yaml:"archive,omitempty"`
}

// ValidatorConfig contains validator-specific settings
type ValidatorConfig struct {
	Enabled        bool   `yaml:"enabled"`
	StakeAmount    string `yaml:"stake_amount"`
	CommissionRate string `yaml:"commission_rate"`
	AutoCompound   bool   `yaml:"auto_compound"`
}

// StateConfig contains state management settings
type StateConfig struct {
	Database DatabaseConfig  `yaml:"database"`
	Pruning  PruningConfig   `yaml:"pruning"`
	Indexing *IndexingConfig `yaml:"indexing,omitempty"`
}

// DatabaseConfig contains database settings
type DatabaseConfig struct {
	Type      string `yaml:"type"`
	Path      string `yaml:"path"`
	CacheSize string `yaml:"cache_size"`
}

// PruningConfig contains state pruning settings
type PruningConfig struct {
	Enabled    bool `yaml:"enabled"`
	KeepRecent int  `yaml:"keep_recent"`
	KeepEvery  int  `yaml:"keep_every"`
}

// IndexingConfig contains indexing settings for archive nodes
type IndexingConfig struct {
	Enabled          bool `yaml:"enabled"`
	TransactionIndex bool `yaml:"transaction_index"`
	ReceiptIndex     bool `yaml:"receipt_index"`
	LogIndex         bool `yaml:"log_index"`
	TraceIndex       bool `yaml:"trace_index"`
}

// RPCConfig contains JSON-RPC settings
type RPCConfig struct {
	Enabled        bool     `yaml:"enabled"`
	ListenAddr     string   `yaml:"listen_addr"`
	CORSOrigins    []string `yaml:"cors_origins"`
	APIModules     []string `yaml:"api_modules"`
	MaxConnections int      `yaml:"max_connections"`
	Timeout        string   `yaml:"timeout,omitempty"`
}

// WebSocketConfig contains WebSocket settings
type WebSocketConfig struct {
	Enabled        bool   `yaml:"enabled"`
	ListenAddr     string `yaml:"listen_addr"`
	MaxConnections int    `yaml:"max_connections"`
}

// MetricsConfig contains metrics and monitoring settings
type MetricsConfig struct {
	Enabled    bool   `yaml:"enabled"`
	ListenAddr string `yaml:"listen_addr"`
	Path       string `yaml:"path"`
}

// AggLayerConfig contains AggLayer integration settings
type AggLayerConfig struct {
	Enabled        bool                 `yaml:"enabled"`
	RPCURL         string               `yaml:"rpc_url"`
	CertificateTTL time.Duration        `yaml:"certificate_ttl"`
	BatchSize      int                  `yaml:"batch_size"`
	Sender         AggLayerSenderConfig `yaml:"sender"`
	Oracle         AggLayerOracleConfig `yaml:"oracle"`
}

// AggLayerSenderConfig contains AggSender settings
type AggLayerSenderConfig struct {
	Enabled           bool          `yaml:"enabled"`
	PrivateKeyPath    string        `yaml:"private_key_path"`
	PollInterval      time.Duration `yaml:"poll_interval"`
	BatchCertificates bool          `yaml:"batch_certificates,omitempty"`
}

// AggLayerOracleConfig contains AggOracle settings
type AggLayerOracleConfig struct {
	Enabled           bool          `yaml:"enabled"`
	UpdateInterval    time.Duration `yaml:"update_interval"`
	VerificationDepth int           `yaml:"verification_depth"`
}

// SecurityConfig contains security settings
type SecurityConfig struct {
	KeystorePath string    `yaml:"keystore_path"`
	PasswordFile string    `yaml:"password_file"`
	TLS          TLSConfig `yaml:"tls"`
}

// TLSConfig contains TLS settings
type TLSConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

// PerformanceConfig contains performance tuning settings
type PerformanceConfig struct {
	MaxTxPoolSize int    `yaml:"max_tx_pool_size"`
	BlockGasLimit string `yaml:"block_gas_limit"`
	TxTimeout     string `yaml:"tx_timeout"`
	SyncMode      string `yaml:"sync_mode"`
	GCPercent     int    `yaml:"gc_percent"`
}

// SequencerConfig contains sequencer-specific settings
type SequencerConfig struct {
	Mempool       MempoolConfig       `yaml:"mempool"`
	Batching      BatchingConfig      `yaml:"batching"`
	FeeManagement FeeManagementConfig `yaml:"fee_management"`
}

// MempoolConfig contains mempool settings
type MempoolConfig struct {
	MaxSize         int    `yaml:"max_size"`
	MaxTxPerAccount int    `yaml:"max_tx_per_account"`
	PriceLimit      string `yaml:"price_limit"`
	PriceBump       int    `yaml:"price_bump"`
}

// BatchingConfig contains transaction batching settings
type BatchingConfig struct {
	MaxBatchSize       int           `yaml:"max_batch_size"`
	MaxBatchBytes      string        `yaml:"max_batch_bytes"`
	BatchTimeout       time.Duration `yaml:"batch_timeout"`
	CompressionEnabled bool          `yaml:"compression_enabled"`
}

// FeeManagementConfig contains fee management settings
type FeeManagementConfig struct {
	BaseFeeEnabled   bool `yaml:"base_fee_enabled"`
	FeeHistoryBlocks int  `yaml:"fee_history_blocks"`
	FeeCapMultiplier int  `yaml:"fee_cap_multiplier"`
}

// ArchiveConfig contains archive node specific settings
type ArchiveConfig struct {
	DataAvailability  DataAvailabilityConfig  `yaml:"data_availability"`
	HistoricalQueries HistoricalQueriesConfig `yaml:"historical_queries"`
	Backup            BackupConfig            `yaml:"backup"`
}

// DataAvailabilityConfig contains data availability settings
type DataAvailabilityConfig struct {
	Enabled         bool   `yaml:"enabled"`
	RetentionPeriod string `yaml:"retention_period"`
	ServeData       bool   `yaml:"serve_data"`
}

// HistoricalQueriesConfig contains historical query settings
type HistoricalQueriesConfig struct {
	MaxRange  int    `yaml:"max_range"`
	CacheSize string `yaml:"cache_size"`
	Timeout   string `yaml:"timeout"`
}

// BackupConfig contains backup settings
type BackupConfig struct {
	Enabled       bool   `yaml:"enabled"`
	Interval      string `yaml:"interval"`
	Path          string `yaml:"path"`
	Compression   bool   `yaml:"compression"`
	RetentionDays int    `yaml:"retention_days"`
}

// L1Config contains Layer 1 integration settings
type L1Config struct {
	RPCURL          string `yaml:"rpc_url"`
	ContractAddress string `yaml:"contract_address"`
	PrivateKeyPath  string `yaml:"private_key_path"`
	GasLimit        int    `yaml:"gas_limit"`
	MaxGasPrice     string `yaml:"max_gas_price"`
}

// Load reads and parses a configuration file
func Load(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.NodeType == "" {
		return fmt.Errorf("node_type is required")
	}

	if c.DataDir == "" {
		return fmt.Errorf("data_dir is required")
	}

	validNodeTypes := map[string]bool{
		"validator": true,
		"sequencer": true,
		"archive":   true,
	}

	if !validNodeTypes[c.NodeType] {
		return fmt.Errorf("invalid node_type: %s", c.NodeType)
	}

	// Node-specific validation
	switch c.NodeType {
	case "sequencer":
		if c.Sequencer == nil {
			return fmt.Errorf("sequencer configuration is required for sequencer nodes")
		}
	case "archive":
		if c.Archive == nil {
			return fmt.Errorf("archive configuration is required for archive nodes")
		}
	}

	return nil
}

// GetTimeout converts timeout string to time.Duration
func (r *RPCConfig) GetTimeout() time.Duration {
	if r.Timeout == "" {
		return 30 * time.Second
	}

	if duration, err := time.ParseDuration(r.Timeout); err == nil {
		return duration
	}

	return 30 * time.Second
}
