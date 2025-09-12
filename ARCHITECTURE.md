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
    └── legacy/             # Legacy L2-specific implementations
```

## Core Components

### 1. L1 Chain Engine (`pkg/l1chain/`)
- Main blockchain implementation
- Coordinates all other components
- Manages block production and validation

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

## Design Principles

### L1-First Architecture
- **Independent blockchain**: Not a Layer 2 or sidechain
- **Native consensus**: HPoS consensus without external dependencies
- **Self-contained**: All validation and finality within the L1

### Performance-Weighted Validation
- **Merit-based selection**: Validators chosen by performance + stake
- **Real-time metrics**: Continuous performance monitoring
- **Dynamic rewards**: Rewards based on actual contribution

### Modular Design
- **Separation of concerns**: Each package has clear responsibility
- **Testability**: Components can be tested independently
- **Extensibility**: Easy to add new features and optimizations

## Future Extensions

The architecture is designed to support future enhancements:

- **RPC/API layer**: Full JSON-RPC and GraphQL support
- **Advanced cryptography**: Zero-knowledge proofs and privacy features
- **Cross-chain bridges**: Interoperability with other blockchains
- **Smart contracts**: EVM-compatible smart contract execution
- **Data availability**: Enhanced data availability guarantees