# LightChain L2 - L1 to L2 Transaction Execution Flow

## 🔄 Transaction Execution Lifecycle

### **1. L1 → L2 (Deposits/Bridging)**

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Ethereum L1   │    │   AggLayer      │    │  LightChain L2  │
│                 │    │   (Bridge)      │    │                 │
│  User submits   │───▶│  Bridge locks   │───▶│  Tokens appear  │
│  bridge tx      │    │  tokens on L1   │    │  on L2          │
└─────────────────┘    └─────────────────┘    └─────────────────┘
        │                        │                        │
        ▼                        ▼                        ▼
   Tx to Bridge              Certificate              AggOracle
   Contract on L1            Generated               Updates L2
```

### **2. L2 Transaction Processing**

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   User Wallet   │    │   Sequencer     │    │   Validators    │
│                 │    │                 │    │                 │
│  Submit tx to   │───▶│  Orders & batch │───▶│  Validate &     │
│  L2 RPC         │    │  transactions   │    │  finalize blocks│
└─────────────────┘    └─────────────────┘    └─────────────────┘
        │                        │                        │
        ▼                        ▼                        ▼
   Instant Response          Batch Creation           Block Finality
   (pending)                 every 2-5s              (12 seconds)
```

### **3. L2 → L1 (Withdrawals)**

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  LightChain L2  │    │   AggLayer      │    │   Ethereum L1   │
│                 │    │   (Unified      │    │                 │
│  User initiates │───▶│   Bridge)       │───▶│  Tokens released│
│  withdrawal     │    │                 │    │  to user        │
└─────────────────┘    └─────────────────┘    └─────────────────┘
        │                        │                        │
        ▼                        ▼                        ▼
   Burn tokens on L2         Certificate              Pessimistic Proof
   Create exit proof         Submission               Verification
```

## 🏗️ Key Components

### **AggSender (in Sequencer)**
- Monitors bridge deposits from L1
- Creates certificates for state changes
- Submits batches to AggLayer every 10 minutes
- Handles L1 → L2 asset transfers

### **AggOracle (in Validators)**
- Fetches verified state from AggLayer
- Updates L2 with cross-chain state
- Enables L2 → L1 withdrawals
- Verifies pessimistic proofs

### **Unified Bridge**
- **No wrapped tokens** - native assets on both chains
- **Fast finality** - sub-5-second transfers
- **Pessimistic proofs** - security-first approach
- **Unified liquidity** - seamless cross-chain

## 📋 Transaction Types

### **1. Pure L2 Transactions**
```
User → L2 RPC → Sequencer → Validators → Block
```
- **Speed**: ~2 seconds
- **Cost**: Very low (L2 gas prices)
- **Security**: L2 consensus + AggLayer

### **2. L1 → L2 Bridge Transactions**
```
User → L1 Bridge Contract → AggLayer → AggOracle → L2 Balance Update
```
- **Speed**: ~5-10 minutes
- **Cost**: L1 gas + bridge fee
- **Security**: L1 finality + pessimistic proofs

### **3. L2 → L1 Withdrawal Transactions**
```
User → L2 Burn → Certificate → AggLayer → L1 Release
```
- **Speed**: ~15-30 minutes (security delay)
- **Cost**: L2 gas + L1 gas for claim
- **Security**: Pessimistic proof + challenge period

## 🔧 Configuration (from your setup)

### **Sequencer L1 Integration**
```yaml
l1:
  rpc_url: "https://eth-mainnet.alchemyapi.io/v2/YOUR_API_KEY"
  contract_address: "0x..." # L1 bridge contract
  private_key_path: "./keys/l1-submitter.key"
  gas_limit: 500000
  max_gas_price: "100000000000" # 100 Gwei
```

### **AggLayer Settings**
```yaml
agglayer:
  enabled: true
  rpc_url: "https://agglayer-rpc.polygon.technology"
  certificate_ttl: "30m"
  batch_size: 500
  sender:
    enabled: true
    poll_interval: "5s"
    batch_certificates: true
```

## 🚀 What Makes LightChain L2 Unique

### **vs Other L2s (Optimism, Arbitrum, Polygon)**

| Feature | LightChain L2 | Traditional L2 |
|---------|---------------|----------------|
| **Bridge** | Unified Bridge (no wrapped tokens) | Wrapped tokens required |
| **Security** | Pessimistic proofs | Optimistic rollups |
| **Finality** | ~5 seconds cross-chain | 7+ days withdrawal |
| **Liquidity** | Unified across chains | Fragmented |
| **Integration** | AggLayer native | Custom bridges |

### **Advantages**
- ✅ **No wrapped tokens** - native assets everywhere
- ✅ **Fast withdrawals** - minutes, not days
- ✅ **Unified liquidity** - seamless cross-chain
- ✅ **Pessimistic security** - assume guilt until proven innocent
- ✅ **AggLayer integration** - interoperability by design

## 💡 Development Tips

### **Testing L1-L2 Interactions**
```bash
# Monitor bridge activity
docker-compose logs -f sequencer | grep -i aggLayer

# Check certificate generation
docker-compose logs -f validator-1 | grep -i certificate

# Watch L1 submissions
docker-compose logs -f sequencer | grep -i "l1_submission"
```

### **Bridge Transaction Example**
```javascript
// L1 → L2 Bridge
const bridgeTx = await l1Bridge.deposit(
  tokenAddress,
  amount,
  l2RecipientAddress
);

// L2 → L1 Withdrawal
const withdrawTx = await l2Bridge.withdraw(
  tokenAddress,
  amount,
  l1RecipientAddress
);
```
