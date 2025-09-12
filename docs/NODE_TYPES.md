# 🔧 LightChain L1 Node Types

LightChain L1 supports three distinct node types, each optimized for different use cases in our **Hybrid Proof-of-Stake (HPoS)** network.

## 📊 **Node Type Overview**

| **Node Type** | **Purpose** | **Resources** | **Rewards** | **Data Storage** |
|---------------|-------------|---------------|-------------|------------------|
| **Validator** | Consensus participation | High | High staking rewards | Recent blocks + state |
| **Full Node** | RPC services | Medium | None | Recent blocks + state |
| **Archive** | Historical data | Highest | None | Complete blockchain history |

---

## 🎯 **1. Validator Nodes**

### **Purpose**
- **Participate in HPoS consensus** by validating transactions and producing blocks
- **Stake LIGHT tokens** and earn rewards based on performance
- **Secure the network** through Byzantine fault-tolerant consensus

### **Requirements**
- **Minimum Stake**: 1,000 LIGHT tokens (1000 * 10^18 wei)
- **Hardware**: 4+ CPU cores, 8GB+ RAM, 500GB+ SSD
- **Network**: Stable internet with low latency
- **Uptime**: 99%+ for optimal rewards

### **Rewards System**
```
Base Reward = (Your Stake / Total Network Stake) * Block Rewards
Performance Bonus = Base Reward * (Performance Score * 0.3)
Total Reward = Base Reward + Performance Bonus
```

### **Performance Metrics**
- **Block Proposals**: Successfully proposed blocks
- **Vote Participation**: Timely prevotes and precommits  
- **Response Time**: Average response to consensus messages
- **Uptime**: Percentage of time online and participating

### **Configuration**
```yaml
# validator.yaml
node_type: "validator"
consensus:
  type: "hpos"
  validator:
    enabled: true
    stake_amount: "1000000000000000000000" # 1000 LIGHT
    commission_rate: "0.05" # 5% commission
    performance_tracking: true
```

### **Slashing Conditions**
- **Double signing**: -5% of stake
- **Downtime** (>10% missed blocks): -1% of stake
- **Byzantine behavior**: -20% of stake

---

## 🌐 **2. Full Nodes**

### **Purpose**
- **Serve RPC requests** for DApps and wallets
- **Relay transactions** and blocks across the network
- **Provide network access** without consensus participation

### **Requirements**
- **No staking required**
- **Hardware**: 2+ CPU cores, 4GB+ RAM, 250GB+ SSD
- **Network**: Stable internet connection
- **Uptime**: Recommended 95%+ for reliable service

### **Use Cases**
- **DApp backends** serving user requests
- **Wallet infrastructure** for transaction broadcasting
- **Analytics platforms** querying blockchain data
- **Development environments** for testing

### **Configuration**
```yaml
# fullnode.yaml
node_type: "fullnode"
consensus:
  validator:
    enabled: false # Full nodes don't validate
rpc:
  enabled: true
  api_modules: ["eth", "net", "web3", "lightchain"]
state:
  pruning:
    enabled: true # Keep only recent data
    keep_recent: 10000
```

### **Data Pruning**
- **Recent Blocks**: Last 10,000 blocks (~5.5 hours)
- **Recent State**: Current state + 1,000 block rollback
- **Transaction Pool**: Active mempool transactions

---

## 📚 **3. Archive Nodes**

### **Purpose**
- **Store complete blockchain history** from genesis
- **Provide historical queries** for analytics and auditing
- **Support blockchain explorers** and data services

### **Requirements**
- **No staking required**
- **Hardware**: 8+ CPU cores, 16GB+ RAM, 2TB+ SSD
- **Network**: High-bandwidth connection
- **Storage**: Growing ~100GB/month (estimated)

### **Use Cases**
- **Blockchain explorers** (Etherscan-like services)
- **Analytics platforms** requiring historical data
- **Compliance tools** for transaction auditing
- **Research and development** on network history

### **Configuration**
```yaml
# archive.yaml (generated from fullnode.yaml)
node_type: "archive"
consensus:
  validator:
    enabled: false
state:
  pruning:
    enabled: false # Never prune data
archive:
  full_state_history: true
  transaction_indexing: true
  log_indexing: true
```

### **Data Storage**
- **All Blocks**: Complete block data from genesis
- **All Transactions**: Full transaction history with receipts
- **All State**: Historical state at every block
- **All Events**: Complete event logs and traces

---

## 🔄 **Node Type Comparison**

### **Startup Sync Times**
- **Validator**: ~2 hours (fast sync)
- **Full Node**: ~1 hour (fast sync)  
- **Archive**: ~24 hours (full historical sync)

### **Storage Growth**
- **Validator**: ~50GB + 10GB/month
- **Full Node**: ~30GB + 5GB/month
- **Archive**: ~500GB + 100GB/month

### **Revenue Models**
- **Validator**: Staking rewards (5-15% APY)
- **Full Node**: Service fees from DApps
- **Archive**: Premium data services

---

## 🚀 **Getting Started**

### **Quick Start Commands**
```bash
# Start a validator node
./lightchain --config configs/validator.yaml --keystore ./keys/validator

# Start a full node  
./lightchain --config configs/fullnode.yaml

# Start an archive node
./lightchain --config configs/archive.yaml --sync-mode full
```

### **Docker Deployment**
```bash
# Validator with staking
docker run -v ./configs:/configs -v ./keys:/keys \
  lightchain:latest --config /configs/validator.yaml

# Full node for RPC services
docker run -p 8545:8545 -v ./configs:/configs \
  lightchain:latest --config /configs/fullnode.yaml
```

### **Kurtosis DevNet**
```bash
# Deploy mixed network for testing
./scripts/kurtosis-manager.sh start 3 2 1
# 3 validators, 2 full nodes, 1 archive node
```

---

## ⚡ **Performance Optimization**

### **Validator Optimization**
- **SSD storage** for fast state access
- **Dedicated CPU cores** for consensus
- **Low-latency network** for quick propagation
- **Memory tuning** for signature verification

### **Full Node Optimization**  
- **RPC connection pooling** for high throughput
- **Caching layers** for frequent queries
- **Load balancing** across multiple instances
- **Monitoring** for uptime and performance

### **Archive Node Optimization**
- **High IOPS storage** for historical queries
- **Large RAM** for caching frequently accessed data
- **Compression** for efficient storage usage
- **Indexing strategies** for fast data retrieval

---

## 🔐 **Security Considerations**

### **All Node Types**
- **Firewall configuration** (only necessary ports open)
- **DDoS protection** for public RPC endpoints
- **Regular updates** to latest software versions
- **Monitoring** for unusual activity

### **Validator Specific**
- **Hardware wallet** integration for signing keys
- **Key rotation** strategies for long-term security
- **Slashing protection** to prevent double signing
- **High availability** setup to minimize downtime

---

## 📈 **Monitoring and Metrics**

### **Key Metrics to Track**
- **Sync status**: Current vs latest block
- **Peer connections**: Active P2P connections
- **Resource usage**: CPU, RAM, disk, network
- **RPC performance**: Request latency and throughput

### **Monitoring Stack**
- **Prometheus**: Metrics collection
- **Grafana**: Dashboard visualization  
- **AlertManager**: Critical alerts
- **Log aggregation**: Centralized logging

### **Health Checks**
```bash
# Check node status
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"lightchain_nodeStatus","params":[],"id":1}' \
  http://localhost:8545

# Check sync status  
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_syncing","params":[],"id":1}' \
  http://localhost:8545
```

---

**Choose your node type based on your needs:**
- 💰 **Want to earn rewards?** → Validator Node
- 🔌 **Building DApps?** → Full Node  
- 📊 **Need historical data?** → Archive Node

**All three node types are essential for a healthy, decentralized LightChain L1 network!** 🌟