#!/bin/bash
# Generate secure keys for LightBeam testnet deployment
# This script generates cryptographically secure keys for validators and accounts

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

log() {
    local level=$1
    shift
    local message="$*"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')

    case $level in
        "INFO")  echo -e "${GREEN}[${timestamp}] INFO:${NC} $message" ;;
        "WARN")  echo -e "${YELLOW}[${timestamp}] WARN:${NC} $message" ;;
        "ERROR") echo -e "${RED}[${timestamp}] ERROR:${NC} $message" ;;
        "SUCCESS") echo -e "${CYAN}[${timestamp}] SUCCESS:${NC} $message" ;;
    esac
}

echo -e "${BLUE}╔══════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║            LIGHTBEAM TESTNET KEY GENERATOR                   ║${NC}"
echo -e "${BLUE}║              Secure Cryptographic Key Generation            ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════════════════╝${NC}"
echo ""

log "INFO" "🔐 Generating secure keys for LightBeam testnet..."

# Create keys directory if it doesn't exist
mkdir -p keys/testnet
mkdir -p configs/generated

# Function to generate a secure private key and address
generate_keypair() {
    local name=$1
    local private_key=$(openssl rand -hex 32)
    # Generate a deterministic address from private key using SHA256 (simplified for testnet)
    local hash=$(echo $private_key | openssl dgst -sha256 -hex | cut -d' ' -f2)
    local address="0x${hash:0:40}"
    # Generate public key from private key (simplified derivation)
    local public_key="0x04$(echo $private_key | openssl dgst -sha256 -hex | cut -d' ' -f2)"

    echo "$name,$private_key,$address,$public_key"
}

log "INFO" "🎯 Generating validator keypairs..."

# Generate 3 validator keypairs
validator1=$(generate_keypair "Genesis Validator 1")
validator2=$(generate_keypair "Genesis Validator 2")
validator3=$(generate_keypair "Genesis Validator 3")

# Generate treasury and development fund keypairs
treasury=$(generate_keypair "Treasury")
devfund=$(generate_keypair "Development Fund")

# Extract addresses and public keys
v1_addr=$(echo $validator1 | cut -d',' -f3)
v1_pubkey=$(echo $validator1 | cut -d',' -f4)
v2_addr=$(echo $validator2 | cut -d',' -f3)
v2_pubkey=$(echo $validator2 | cut -d',' -f4)
v3_addr=$(echo $validator3 | cut -d',' -f3)
v3_pubkey=$(echo $validator3 | cut -d',' -f4)
treasury_addr=$(echo $treasury | cut -d',' -f3)
devfund_addr=$(echo $devfund | cut -d',' -f3)

# Save private keys securely
cat > keys/testnet/validator-keys.json << EOF
{
  "validators": [
    {
      "name": "Genesis Validator 1",
      "private_key": "$(echo $validator1 | cut -d',' -f2)",
      "address": "$v1_addr",
      "public_key": "$v1_pubkey"
    },
    {
      "name": "Genesis Validator 2",
      "private_key": "$(echo $validator2 | cut -d',' -f2)",
      "address": "$v2_addr",
      "public_key": "$v2_pubkey"
    },
    {
      "name": "Genesis Validator 3",
      "private_key": "$(echo $validator3 | cut -d',' -f2)",
      "address": "$v3_addr",
      "public_key": "$v3_pubkey"
    }
  ],
  "treasury": {
    "private_key": "$(echo $treasury | cut -d',' -f2)",
    "address": "$treasury_addr"
  },
  "development_fund": {
    "private_key": "$(echo $devfund | cut -d',' -f2)",
    "address": "$devfund_addr"
  }
}
EOF

# Generate new genesis configuration with secure keys
cat > configs/generated/genesis-secure.yaml << EOF
# LightChain L1 Genesis Configuration - SECURE TESTNET VERSION
# Generated on: $(date)
# WARNING: These keys are for testnet use only!

# Chain information
chain_id: 1337
network_id: 1337
chain_name: "LightBeam Testnet"

# Genesis block configuration
genesis:
  timestamp: "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  parent_hash: "0x0000000000000000000000000000000000000000000000000000000000000000"
  extra_data: "LightBeam Testnet - Secure Deployment $(date +%Y%m%d)"
  gas_limit: "10000000"
  difficulty: "0x1"

# Consensus configuration
consensus:
  type: "hpos" # Hybrid Proof of Stake (L1 native)
  epoch_length: 100
  block_time: "2s"
  min_validators: 3
  max_validators: 100

  # Staking parameters
  staking:
    min_stake: "1000000000000000000000" # 1000 tokens (18 decimals)
    slash_fraction_double_sign: "0.05" # 5%
    slash_fraction_downtime: "0.01" # 1%
    downtime_jail_duration: "600s" # 10 minutes

# Initial validators - SECURE GENERATED KEYS
validators:
- address: "$v1_addr"
  name: "Genesis Validator 1"
  pub_key: "$v1_pubkey"
  stake: "5000000000000000000000" # 5000 tokens
  commission_rate: "0.10" # 10%

- address: "$v2_addr"
  name: "Genesis Validator 2"
  pub_key: "$v2_pubkey"
  stake: "5000000000000000000000" # 5000 tokens
  commission_rate: "0.10" # 10%

- address: "$v3_addr"
  name: "Genesis Validator 3"
  pub_key: "$v3_pubkey"
  stake: "5000000000000000000000" # 5000 tokens
  commission_rate: "0.10" # 10%

# Initial token allocation - SECURE GENERATED ADDRESSES
allocations:
# Pre-funded validator accounts
- address: "$v1_addr"
  balance: "100000000000000000000000000" # 100M tokens

- address: "$v2_addr"
  balance: "100000000000000000000000000" # 100M tokens

- address: "$v3_addr"
  balance: "100000000000000000000000000" # 100M tokens

# Treasury account - SECURE GENERATED
- address: "$treasury_addr"
  balance: "500000000000000000000000000" # 500M tokens

# Development fund - SECURE GENERATED
- address: "$devfund_addr"
  balance: "200000000000000000000000000" # 200M tokens

# Token configuration
token:
  name: "LightChain Token"
  symbol: "LIGHT"
  decimals: 18
  total_supply: "1000000000000000000000000000" # 1B tokens

# Gas configuration
gas:
  min_gas_price: "1000000000" # 1 Gwei
  gas_price_oracle:
    enabled: true
    base_fee_max_change_denominator: 8
    elasticity_multiplier: 2

# Fee configuration
fees:
  transaction_fee: "21000" # Base transaction fee in gas
  contract_creation_fee: "53000" # Contract creation fee in gas
  contract_call_fee: "25000" # Contract call fee in gas

# L1 Native Features
l1_native:
  enabled: true
  finality_blocks: 3
  cross_chain_bridges: true
  native_staking: true

# Precompiled contracts
precompiles:
# Standard Ethereum precompiles
- address: "0x0000000000000000000000000000000000000001" # ecRecover
  enabled: true
- address: "0x0000000000000000000000000000000000000002" # sha256
  enabled: true
- address: "0x0000000000000000000000000000000000000003" # ripemd160
  enabled: true
- address: "0x0000000000000000000000000000000000000004" # identity
  enabled: true
- address: "0x0000000000000000000000000000000000000005" # modexp
  enabled: true
- address: "0x0000000000000000000000000000000000000006" # bn256Add
  enabled: true
- address: "0x0000000000000000000000000000000000000007" # bn256ScalarMul
  enabled: true
- address: "0x0000000000000000000000000000000000000008" # bn256Pairing
  enabled: true
- address: "0x0000000000000000000000000000000000000009" # blake2f
  enabled: true

# LightChain L1 specific precompiles
- address: "0x0000000000000000000000000000000000001001" # Cross-chain bridges
  enabled: true
- address: "0x0000000000000000000000000000000000001002" # HPoS staking operations
  enabled: true
- address: "0x0000000000000000000000000000000000001003" # Developer rewards
  enabled: true

# Smart contracts to deploy at genesis
contracts:
# Core system contracts
- name: "ValidatorSet"
  address: "0x0000000000000000000000000000000000002001"
  bytecode: "0x" # To be compiled and added

- name: "StakingManager"
  address: "0x0000000000000000000000000000000000002002"
  bytecode: "0x" # To be compiled and added

- name: "BridgeContract"
  address: "0x0000000000000000000000000000000000002003"
  bytecode: "0x" # To be compiled and added

- name: "GovernanceContract"
  address: "0x0000000000000000000000000000000000002004"
  bytecode: "0x" # To be compiled and added

# Network bootstrapping
bootstrap:
  # Bootnode configurations
  bootnodes:
  - "/ip4/127.0.0.1/tcp/30301/p2p/16Uiu2HAm7Vs8StL5RjKaRqN7BkZZS6Zw9G1XriBaYGZq4TdB5Hj4"
  - "/ip4/127.0.0.1/tcp/30302/p2p/16Uiu2HAm8Kp8TxQ9RXrBGzgP9kBm7LJyGj5G4TdB5Hj4KKr2TZnE"

  # DNS seeds (for production)
  dns_seeds:
  - "seed1.lightchain.network"
  - "seed2.lightchain.network"
  - "seed3.lightchain.network"

# Hard fork schedule
forks:
# Future upgrades can be scheduled here
- name: "Genesis"
  block: 0
  enabled: true

# Example future fork
- name: "Upgrade1"
  block: 1000000
  enabled: false
  description: "Enhanced ZK integration"

# Development settings (testnet only)
development:
  enable_dev_accounts: false # Disabled for secure deployment
  auto_mine_blocks: true
  dev_period: "2s" # Mining period for testnet
  unlock_accounts: [] # No pre-unlocked accounts for security
EOF

# Set secure permissions
chmod 600 keys/testnet/validator-keys.json
chmod 644 configs/generated/genesis-secure.yaml

log "SUCCESS" "🎉 Secure keys generated successfully!"
echo ""
log "INFO" "📝 Generated files:"
echo "   • keys/testnet/validator-keys.json (private keys - keep secure!)"
echo "   • configs/generated/genesis-secure.yaml (public configuration)"
echo ""
log "INFO" "🔐 Key Summary:"
echo "   • Validator 1: $v1_addr"
echo "   • Validator 2: $v2_addr"
echo "   • Validator 3: $v3_addr"
echo "   • Treasury:    $treasury_addr"
echo "   • Dev Fund:    $devfund_addr"
echo ""
log "WARN" "⚠️  SECURITY NOTES:"
echo "   • Private keys are stored in keys/testnet/validator-keys.json"
echo "   • This file has restricted permissions (600)"
echo "   • NEVER commit private keys to version control"
echo "   • Use these keys for testnet deployment only"
echo "   • For production, use hardware security modules (HSMs)"
echo ""
log "INFO" "🚀 To use the secure genesis configuration:"
echo "   cp configs/generated/genesis-secure.yaml configs/genesis.yaml"
echo "   ./build/lightchain --chain-id 1337 --genesis configs/genesis.yaml"