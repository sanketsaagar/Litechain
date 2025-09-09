#!/bin/bash
# LightChain L2 Activity Simulator
# Creates realistic blockchain activity by sending mock transactions

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Configuration
RPC_URL="http://localhost:8545"
SEQUENCER_URL="http://localhost:8555"
INTERVAL=3 # seconds between transactions

# Pre-funded accounts from genesis
ACCOUNTS=(
    "0x742A4D1A0Ac05A73A48F10C2E2d6b0E3f1b2e3F4"
    "0x8B3A4D1A0Ac05A73A48F10C2E2d6b0E3f1b2e3F5"
    "0x9C4A4D1A0Ac05A73A48F10C2E2d6b0E3f1b2e3F6"
)

# Function to print colored log messages
log_info() {
    echo -e "${GREEN}[$(date '+%H:%M:%S')] INFO:${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[$(date '+%H:%M:%S')] WARN:${NC} $1"
}

log_error() {
    echo -e "${RED}[$(date '+%H:%M:%S')] ERROR:${NC} $1"
}

log_activity() {
    echo -e "${BLUE}[$(date '+%H:%M:%S')] ACTIVITY:${NC} $1"
}

log_block() {
    echo -e "${PURPLE}[$(date '+%H:%M:%S')] BLOCK:${NC} $1"
}

# Function to generate random transaction data
generate_transaction() {
    local from_idx=$((RANDOM % ${#ACCOUNTS[@]}))
    local to_idx=$((RANDOM % ${#ACCOUNTS[@]}))
    
    # Ensure from and to are different
    while [ $from_idx -eq $to_idx ]; do
        to_idx=$((RANDOM % ${#ACCOUNTS[@]}))
    done
    
    local from_addr="${ACCOUNTS[$from_idx]}"
    local to_addr="${ACCOUNTS[$to_idx]}"
    local value=$((RANDOM % 1000 + 1)) # 1-1000 wei
    local gas_price=$((RANDOM % 20000000000 + 1000000000)) # 1-20 Gwei
    
    echo "$from_addr,$to_addr,$value,$gas_price"
}

# Function to simulate RPC calls
simulate_rpc_call() {
    local method="$1"
    local description="$2"
    
    # Simulate network delay
    sleep 0.1
    
    # Simulate success/failure (95% success rate)
    if [ $((RANDOM % 100)) -lt 95 ]; then
        log_activity "$description ✅"
        return 0
    else
        log_warn "$description ❌ (simulated error)"
        return 1
    fi
}

# Function to simulate block creation
simulate_block_creation() {
    local block_num=$1
    local tx_count=$((RANDOM % 5 + 1)) # 1-5 transactions per block
    
    log_block "Creating block #$block_num with $tx_count transactions"
    
    # Simulate block validation
    sleep 0.5
    log_activity "Block #$block_num validated by consensus engine"
    
    # Simulate state update
    sleep 0.2
    log_activity "State updated for block #$block_num"
    
    # Simulate AggLayer certificate
    if [ $((RANDOM % 3)) -eq 0 ]; then
        log_activity "AggLayer certificate generated for block #$block_num"
    fi
}

# Function to check if blockchain is running
check_blockchain_status() {
    log_info "Checking blockchain node status..."
    
    # Check validator
    if curl -s --max-time 2 "$RPC_URL" >/dev/null 2>&1; then
        log_info "✅ Validator node responding at $RPC_URL"
    else
        log_warn "⚠️  Validator node not responding at $RPC_URL"
    fi
    
    # Check sequencer
    if curl -s --max-time 2 "$SEQUENCER_URL" >/dev/null 2>&1; then
        log_info "✅ Sequencer node responding at $SEQUENCER_URL"
    else
        log_warn "⚠️  Sequencer node not responding at $SEQUENCER_URL"
    fi
}

# Main activity simulation loop
run_activity_simulation() {
    log_info "🚀 Starting LightChain L2 Activity Simulator"
    log_info "🔄 Generating blockchain activity every $INTERVAL seconds"
    log_info "📊 Press Ctrl+C to stop"
    echo ""
    
    local tx_count=0
    local block_count=1
    local last_block_time=$(date +%s)
    
    # Simulate some startup activity
    log_activity "Initializing blockchain state..."
    log_activity "Loading validator set..."
    log_activity "Connecting to AggLayer..."
    log_activity "Ready to process transactions"
    echo ""
    
    while true; do
        # Generate transaction activity
        local tx_data=$(generate_transaction)
        IFS=',' read -r from_addr to_addr value gas_price <<< "$tx_data"
        
        tx_count=$((tx_count + 1))
        
        # Simulate transaction lifecycle
        log_activity "📩 Transaction #$tx_count received: ${from_addr:0:10}...→${to_addr:0:10}... (${value} wei)"
        simulate_rpc_call "eth_sendTransaction" "Transaction added to mempool"
        
        # Simulate sequencer processing
        if [ $((RANDOM % 2)) -eq 0 ]; then
            log_activity "🔄 Sequencer batching transaction #$tx_count"
        fi
        
        # Simulate block creation every 6-10 transactions or every 20-30 seconds
        local current_time=$(date +%s)
        local time_since_block=$((current_time - last_block_time))
        
        if [ $((tx_count % 7)) -eq 0 ] || [ $time_since_block -gt 25 ]; then
            echo ""
            simulate_block_creation $block_count
            block_count=$((block_count + 1))
            last_block_time=$current_time
            echo ""
        fi
        
        # Simulate various blockchain activities
        case $((RANDOM % 8)) in
            0) log_activity "🔍 Validator performing state verification" ;;
            1) log_activity "🌐 P2P sync with peer nodes" ;;
            2) log_activity "💾 Database checkpoint completed" ;;
            3) log_activity "📊 Metrics updated (TPS: $((RANDOM % 100 + 50)))" ;;
            4) log_activity "🔐 Cryptographic signature verification" ;;
            5) log_activity "⚡ Gas limit adjustment (Current: ${gas_price})" ;;
            6) log_activity "🎯 Transaction pool optimization" ;;
            7) log_activity "📡 Broadcasting block to network" ;;
        esac
        
        # Show progress
        if [ $((tx_count % 10)) -eq 0 ]; then
            log_info "📈 Total activity: $tx_count transactions, $((block_count - 1)) blocks"
        fi
        
        sleep $INTERVAL
    done
}

# Handle script termination
cleanup() {
    echo ""
    log_info "🛑 Stopping activity simulator"
    log_info "📊 Final stats: $tx_count transactions, $((block_count - 1)) blocks"
    exit 0
}

trap cleanup SIGINT SIGTERM

# Main execution
case "${1:-run}" in
    "check"|"status")
        check_blockchain_status
        ;;
    "run"|*)
        check_blockchain_status
        echo ""
        run_activity_simulation
        ;;
esac
