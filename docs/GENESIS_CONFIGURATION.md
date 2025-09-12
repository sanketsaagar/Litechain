# 🎯 Genesis Configuration Guide

The `genesis.yaml` file is the **foundation** of your LightChain L1 blockchain - it defines the initial state when the blockchain starts from block 0.

## 🔧 **What is Genesis Configuration?**

Genesis configuration is like the **"DNA"** of your blockchain that defines:

- 🏗️ **Initial blockchain state** (block 0)
- 👑 **Genesis validators** who can produce the first blocks
- 💰 **Initial token allocation** (who gets tokens at launch)
- ⚙️ **Network parameters** (consensus rules, gas limits, etc.)
- 🔗 **Chain identity** (chain ID, network name)

## 📂 **Available Genesis Files**

| **File** | **Purpose** | **Chain ID** | **Use Case** |
|----------|-------------|--------------|--------------|
| `genesis.yaml` | **Testnet/Development** | 1337 | Local testing, development |
| `genesis-mainnet.yaml` | **Production Mainnet** | 1001 | Live production network |

## 🎯 **Chain ID Usage**

### **Chain ID 1337 (Testnet)**
- ✅ **Development and testing**
- ✅ **Local devnet deployments**  
- ✅ **Kurtosis test environments**
- ❌ **Not for production use**

### **Chain ID 1001 (Mainnet)**
- ✅ **Production mainnet deployment**
- ✅ **Real LIGHT token value**
- ✅ **Secure validator network**
- ❌ **Not for testing**

## 📋 **Genesis File Structure**

### **1. Chain Identity**
```yaml
chain_id: 1001 # Unique blockchain identifier
network_id: 1001 # Network identifier (usually same as chain_id)
chain_name: "LightChain L1 Mainnet" # Human-readable name
```

### **2. Genesis Block**
```yaml
genesis:
  timestamp: "2024-12-31T00:00:00Z" # Launch time
  parent_hash: "0x000...000" # Always zeros for genesis
  extra_data: "LightChain L1 - Developer-First Blockchain"
  gas_limit: "30000000" # Block gas limit
  difficulty: "0x1" # Always 1 for PoS
```

### **3. Consensus Parameters**
```yaml
consensus:
  type: "hpos" # Hybrid Proof of Stake
  epoch_length: 100 # Blocks per epoch
  block_time: "2s" # Target block time
  min_validators: 3 # Minimum validators
  max_validators: 21 # Maximum validators
  
  staking:
    min_stake: "1000000000000000000000" # 1000 LIGHT
    slash_fraction_double_sign: "0.05" # 5% slashing
    slash_fraction_downtime: "0.01" # 1% slashing
```

### **4. Genesis Validators**
```yaml
validators:
- address: "0xYOUR_VALIDATOR_ADDRESS" # Validator wallet
  name: "Genesis Validator 1" # Human name
  pub_key: "0xVALIDATOR_PUBLIC_KEY" # Consensus public key
  stake: "100000000000000000000000" # 100K LIGHT stake
  commission_rate: "0.05" # 5% commission
```

### **5. Token Allocation**
```yaml
allocations:
# Founder gets 200M LIGHT (20%)
- address: "0xYOUR_WALLET_ADDRESS"
  balance: "200000000000000000000000000"

# Validators get initial stakes
- address: "0xVALIDATOR_ADDRESS"
  balance: "100000000000000000000000"
```

### **6. System Contracts**
```yaml
contracts:
- name: "ValidatorSet" # Core validator management
  address: "0x0000000000000000000000000000000000002001"
- name: "StakingManager" # Staking logic
  address: "0x0000000000000000000000000000000000002002"
- name: "DeveloperRewards" # Developer incentives
  address: "0x0000000000000000000000000000000000002004"
```

## 🚀 **How to Use Genesis Files**

### **For Development (Testnet)**
```bash
# Use testnet genesis (chain ID 1337)
./lightchain --chain-id 1337 --genesis configs/genesis.yaml

# Or via Kurtosis
./scripts/kurtosis-manager.sh start 3 2 1
```

### **For Production (Mainnet)**
```bash
# Use mainnet genesis (chain ID 1001)  
./lightchain --chain-id 1001 --genesis configs/genesis-mainnet.yaml

# Connect to mainnet
lightchain-cli --chain-id 1001 --node https://rpc.lightchain.network
```

## ⚠️ **Critical Setup Steps for Mainnet**

### **1. Update Your Wallet Address**
```yaml
# In genesis-mainnet.yaml, replace:
- address: "0xYOUR_WALLET_ADDRESS_HERE" # <-- UPDATE THIS
  balance: "200000000000000000000000000" # 200M LIGHT
```

### **2. Set Genesis Validators**
```yaml
# Replace with actual validator addresses and public keys:
validators:
- address: "0xYOUR_VALIDATOR_ADDRESS_1" # <-- UPDATE
  pub_key: "0xYOUR_VALIDATOR_PUBKEY_1"  # <-- UPDATE
```

### **3. Configure Bootstrap Nodes**
```yaml
bootstrap:
  bootnodes:
  - "/ip4/YOUR_NODE_IP_1/tcp/30301/p2p/NODE_ID_1" # <-- UPDATE
  - "/ip4/YOUR_NODE_IP_2/tcp/30302/p2p/NODE_ID_2" # <-- UPDATE
```

## 🔒 **Security Considerations**

### **Genesis Validator Keys**
- 🔐 **Use hardware wallets** for genesis validator keys
- 🔄 **Generate fresh keys** specifically for validators
- 💾 **Backup private keys securely** (they can't be recovered)
- 🚫 **Never share private keys**

### **Token Allocation**
- ✅ **Verify all addresses** before mainnet launch
- ✅ **Double-check token amounts** (18 decimal places)
- ✅ **Ensure total supply** = 1,000,000,000 LIGHT tokens
- ❌ **Never allocate more than total supply**

## 📊 **Token Distribution (Mainnet)**

| **Category** | **Amount** | **Percentage** | **Purpose** |
|--------------|------------|----------------|-------------|
| **Founder** | 200M LIGHT | 20% | You as blockchain creator |
| **Team** | 50M LIGHT | 5% | Future team members |
| **Ecosystem** | 100M LIGHT | 10% | Partnerships, development |
| **Treasury** | 150M LIGHT | 15% | Operations, marketing |
| **Validators** | 0.3M LIGHT | 0.03% | Initial stakes |
| **Mining Pool** | 499.7M LIGHT | 49.97% | Rewards, incentives |

## 🔄 **Genesis Hash Verification**

After genesis is created, verify the genesis hash:

```bash
# Calculate genesis hash
./lightchain --genesis configs/genesis-mainnet.yaml --genesis-hash

# Expected output:
# Genesis Hash: 0x1234567890abcdef...
# Block Number: 0
# Chain ID: 1001
```

⚠️ **Important**: All nodes must use **identical genesis configurations** or they won't be able to communicate!

## 🛠️ **Troubleshooting**

### **Common Issues**

1. **"Invalid genesis validators"**
   - Ensure validator addresses and public keys are correct
   - Check minimum stake requirements are met

2. **"Total allocation exceeds supply"**
   - Verify all balance amounts add up to ≤ 1B tokens
   - Remember: 1 LIGHT = 10^18 wei

3. **"Chain ID mismatch"**  
   - Ensure all nodes use the same chain ID
   - Use 1337 for testnet, 1001 for mainnet

4. **"Bootstrap nodes unreachable"**
   - Verify bootstrap node addresses are correct
   - Check network connectivity and firewall rules

## 📚 **Related Documentation**

- 🔧 [Node Types Guide](NODE_TYPES.md)
- 🚀 [Quick Start Guide](QUICKSTART.md)
- 🏗️ [Architecture Overview](../ARCHITECTURE.md)
- 💰 [Developer Guide](../DEVELOPER_GUIDE.md)

---

**Your genesis configuration is the foundation of your blockchain empire! Configure it carefully and launch with confidence.** 🌟