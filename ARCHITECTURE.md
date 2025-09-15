# LightChain L1 Architecture

## Project Structure

```
lightchain-l1/
├── cmd/                    # Main applications
│   └── lightchain/         # Main L1 blockchain binary
├── pkg/                    # Reusable packages
│   ├── consensus/          # HPoS consensus engine
│   ├── network/            # P2P networking
│   ├── staking/            # Validator staking and management
│   ├── economics/          # Token economics and gas model
│   ├── genesis/            # Genesis block configuration
│   ├── l1chain/            # Main L1 blockchain implementation
│   ├── api/                # API interfaces (future)
│   ├── crypto/             # Cryptographic utilities (future)
│   ├── db/                 # Database interfaces (future)
│   ├── rpc/                # RPC server implementation (future)
│   ├── sync/               # Block synchronization (future)
│   └── mempool/            # Transaction pool (future)
├── internal/               # Private application code
│   └── config/             # Configuration management
├── docs/                   # Documentation
├── scripts/                # Deployment and management scripts
├── deployments/            # Deployment configurations
│   ├── docker/             # Docker configurations
│   ├── kurtosis/           # Kurtosis test environments
│   ├── grafana/            # Monitoring dashboards
│   └── prometheus/         # Metrics collection
├── configs/                # Configuration files
├── keys/                   # Key management
├── data/                   # Runtime data directory
└── archive/                # Archived legacy code
    └── legacy/             # Legacy implementations (if any)
```

## Core Components

### 1. ZK-Native L1 Chain Engine (`pkg/l1chain/`)
- Main blockchain implementation with integrated ZK engine
- Coordinates all components including ZK proof verification
- Manages block production, validation, and ZK rollup settlement
- Native support for privacy-preserving transactions

### 2. HPoS Consensus (`pkg/consensus/`)
- Hybrid Proof-of-Stake consensus mechanism
- Performance-weighted validator selection
- Byzantine fault tolerant block finalization

### 3. Validator Staking (`pkg/staking/`)
- Validator registration and management
- Stake management and delegation
- Performance tracking and rewards

### 4. Network Layer (`pkg/network/`)
- P2P networking for validator communication
- Block and transaction propagation
- Peer discovery and management

### 5. Token Economics (`pkg/economics/`)
- Dynamic gas pricing model
- Token supply management
- Fee burning and reward distribution

### 6. Genesis Management (`pkg/genesis/`)
- Genesis block configuration
- Initial validator set
- Token allocation and economics parameters

### 7. ZK Engine (`pkg/zk/`)
- **Native zero-knowledge capabilities**
- Multi-proof system support (SNARKs, STARKs, Bulletproofs)
- ZK rollup infrastructure and management
- Privacy-preserving transaction processing
- Universal ZK bridges for cross-chain privacy

## Design Principles

### ZK-Native L1-First Architecture
- **Independent ZK-enabled blockchain**: Not a Layer 2 or sidechain
- **Native ZK consensus**: HPoS consensus with integrated ZK verification
- **ZK-powered privacy**: Native privacy-preserving transactions
- **Self-contained**: All validation, finality, and ZK proofs within the L1

### Performance-Weighted Validation
- **Merit-based selection**: Validators chosen by performance + stake
- **Real-time metrics**: Continuous performance monitoring
- **Dynamic rewards**: Rewards based on actual contribution

### Modular Design
- **Separation of concerns**: Each package has clear responsibility
- **Testability**: Components can be tested independently
- **Extensibility**: Easy to add new features and optimizations

## Current ZK Features & Future Extensions

The ZK-native architecture currently includes:

### **✅ Implemented ZK Features**
- **Multi-proof systems**: SNARKs, STARKs, Bulletproofs
- **ZK rollup infrastructure**: Up to 100 rollups, 50K TPS each
- **Privacy-preserving transactions**: Hidden amounts and recipients
- **Universal ZK bridges**: Private cross-chain transfers
- **ZK-enhanced EVM**: Privacy extensions to Solidity contracts

### **🔮 Future ZK Enhancements**
- **ZK-ML integration**: Zero-knowledge machine learning
- **Advanced ZK circuits**: Custom privacy-preserving applications
- **ZK governance**: Private voting and proposal mechanisms
- **ZK identity**: Privacy-preserving identity verification
- **ZK compliance**: Regulatory compliance with privacy