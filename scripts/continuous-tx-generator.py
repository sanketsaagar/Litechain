#!/usr/bin/env python3
"""
LightChain L2 Continuous Transaction Generator
Generates realistic blockchain activity 24/7 like a real blockchain
"""

import requests
import time
import json
import random
import os
import sys
from datetime import datetime
import threading
import signal

# Configuration from environment variables
RPC_URL = os.getenv('RPC_URL', 'http://localhost:8545')
SEQUENCER_URL = os.getenv('SEQUENCER_URL', 'http://localhost:8555')
TX_INTERVAL = int(os.getenv('TX_INTERVAL', '3'))  # seconds between transactions
ENABLE_BURST_MODE = os.getenv('ENABLE_BURST_MODE', 'true').lower() == 'true'

# Pre-funded development accounts (from genesis.yaml)
ACCOUNTS = [
    {
        "address": "0x742A4D1A0Ac05A73A48F10C2E2d6b0E3f1b2e3F4",
        "private_key": "0x0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
    },
    {
        "address": "0x8B3A4D1A0Ac05A73A48F10C2E2d6b0E3f1b2e3F5", 
        "private_key": "0x1123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
    },
    {
        "address": "0x9C4A4D1A0Ac05A73A48F10C2E2d6b0E3f1b2e3F6",
        "private_key": "0x2123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
    }
]

# Transaction patterns (realistic blockchain activity)
TX_PATTERNS = [
    {"type": "transfer", "weight": 60, "amount_range": (1, 1000)},
    {"type": "large_transfer", "weight": 15, "amount_range": (1000, 10000)},
    {"type": "micro_transfer", "weight": 20, "amount_range": (1, 10)},
    {"type": "contract_call", "weight": 5, "amount_range": (0, 100)}
]

# Global statistics
stats = {
    "total_transactions": 0,
    "successful_transactions": 0,
    "failed_transactions": 0,
    "start_time": datetime.now(),
    "last_block_check": 0,
    "blocks_observed": 0
}

running = True

def log_message(msg, level="INFO"):
    timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    print(f"[{timestamp}] {level}: {msg}", flush=True)

def signal_handler(signum, frame):
    global running
    log_message("Received shutdown signal. Stopping transaction generator...")
    running = False

def send_rpc_request(url, method, params=None):
    """Send JSON-RPC request with retry logic"""
    if params is None:
        params = []
    
    payload = {
        "jsonrpc": "2.0",
        "method": method,
        "params": params,
        "id": random.randint(1, 1000)
    }
    
    max_retries = 3
    for attempt in range(max_retries):
        try:
            response = requests.post(url, json=payload, timeout=10)
            if response.status_code == 200:
                return response.json()
            else:
                log_message(f"HTTP {response.status_code} from {url}", "WARN")
        except requests.exceptions.RequestException as e:
            if attempt < max_retries - 1:
                log_message(f"RPC request failed (attempt {attempt + 1}): {e}", "WARN")
                time.sleep(2 ** attempt)  # Exponential backoff
            else:
                log_message(f"RPC request failed after {max_retries} attempts: {e}", "ERROR")
    
    return None

def check_network_health():
    """Check if blockchain network is healthy"""
    try:
        # Check validator
        result = send_rpc_request(RPC_URL, "web3_clientVersion")
        validator_ok = result is not None
        
        # Check sequencer
        result = send_rpc_request(SEQUENCER_URL, "web3_clientVersion")
        sequencer_ok = result is not None
        
        # Check block progression
        result = send_rpc_request(RPC_URL, "eth_blockNumber")
        current_block = 0
        if result and "result" in result:
            current_block = int(result["result"], 16)
            
        if current_block > stats["last_block_check"]:
            stats["blocks_observed"] += current_block - stats["last_block_check"]
            stats["last_block_check"] = current_block
            
        return validator_ok, sequencer_ok, current_block
        
    except Exception as e:
        log_message(f"Health check failed: {e}", "ERROR")
        return False, False, 0

def generate_realistic_transaction():
    """Generate a realistic transaction based on patterns"""
    pattern = random.choices(
        TX_PATTERNS, 
        weights=[p["weight"] for p in TX_PATTERNS]
    )[0]
    
    from_account = random.choice(ACCOUNTS)
    to_account = random.choice(ACCOUNTS)
    
    # Ensure different accounts
    while to_account["address"] == from_account["address"]:
        to_account = random.choice(ACCOUNTS)
    
    min_amount, max_amount = pattern["amount_range"]
    amount = random.randint(min_amount, max_amount)
    
    # Generate realistic gas prices (1-50 Gwei)
    gas_price = random.randint(1000000000, 50000000000)
    
    # Vary gas limit based on transaction type
    if pattern["type"] == "contract_call":
        gas_limit = random.randint(50000, 200000)
    else:
        gas_limit = 21000
    
    tx = {
        "from": from_account["address"],
        "to": to_account["address"],
        "value": hex(amount),
        "gas": hex(gas_limit),
        "gasPrice": hex(gas_price),
        "data": "0x" if pattern["type"] != "contract_call" else "0x" + "00" * random.randint(0, 100)
    }
    
    return tx, pattern["type"], amount

def send_transaction(tx, tx_type, amount):
    """Send transaction to blockchain"""
    stats["total_transactions"] += 1
    
    # Try both validator and sequencer endpoints
    endpoints = [
        ("Validator", RPC_URL),
        ("Sequencer", SEQUENCER_URL)
    ]
    
    for endpoint_name, url in endpoints:
        result = send_rpc_request(url, "eth_sendTransaction", [tx])
        if result and "result" in result:
            stats["successful_transactions"] += 1
            log_message(
                f"✅ {tx_type.title()} tx sent via {endpoint_name}: "
                f"{tx['from'][:10]}...→{tx['to'][:10]}... "
                f"({amount} wei) Hash: {result['result'][:10]}..."
            )
            return result["result"]
        elif result and "error" in result:
            log_message(f"❌ Transaction error on {endpoint_name}: {result['error']}", "WARN")
    
    stats["failed_transactions"] += 1
    log_message(f"❌ Transaction failed on all endpoints", "ERROR")
    return None

def generate_burst_activity():
    """Generate burst of activity (like DEX trading, NFT minting, etc.)"""
    if not ENABLE_BURST_MODE:
        return
        
    # 5% chance of burst activity every minute
    if random.random() < 0.05:
        burst_size = random.randint(5, 15)
        log_message(f"🚀 Generating burst activity: {burst_size} transactions")
        
        for i in range(burst_size):
            if not running:
                break
                
            tx, tx_type, amount = generate_realistic_transaction()
            send_transaction(tx, f"burst_{tx_type}", amount)
            time.sleep(0.5)  # Rapid-fire transactions

def print_statistics():
    """Print periodic statistics"""
    uptime = datetime.now() - stats["start_time"]
    success_rate = (stats["successful_transactions"] / max(stats["total_transactions"], 1)) * 100
    
    log_message(
        f"📊 Stats: {stats['total_transactions']} total txs, "
        f"{stats['successful_transactions']} successful ({success_rate:.1f}%), "
        f"{stats['blocks_observed']} blocks observed, "
        f"Uptime: {uptime}"
    )

def monitor_blockchain():
    """Background thread to monitor blockchain health"""
    while running:
        try:
            validator_ok, sequencer_ok, current_block = check_network_health()
            
            if current_block > 0:
                log_message(f"🧱 Block #{current_block} | Validator: {'✅' if validator_ok else '❌'} | Sequencer: {'✅' if sequencer_ok else '❌'}")
            
            if not validator_ok and not sequencer_ok:
                log_message("⚠️  All endpoints down, waiting for recovery...", "WARN")
                time.sleep(30)
            else:
                time.sleep(60)  # Check every minute
                
        except Exception as e:
            log_message(f"Monitor error: {e}", "ERROR")
            time.sleep(60)

def main():
    """Main transaction generation loop"""
    global running
    
    # Set up signal handlers
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)
    
    log_message("🚀 Starting LightChain L2 Continuous Transaction Generator")
    log_message(f"🎯 Target: Validator={RPC_URL}, Sequencer={SEQUENCER_URL}")
    log_message(f"⏱️  Interval: {TX_INTERVAL}s, Burst mode: {ENABLE_BURST_MODE}")
    
    # Start monitoring thread
    monitor_thread = threading.Thread(target=monitor_blockchain, daemon=True)
    monitor_thread.start()
    
    # Wait for blockchain to be ready
    log_message("⏳ Waiting for blockchain to be ready...")
    while running:
        validator_ok, sequencer_ok, _ = check_network_health()
        if validator_ok or sequencer_ok:
            log_message("✅ Blockchain is ready, starting transaction generation")
            break
        time.sleep(10)
    
    if not running:
        return
    
    # Main transaction generation loop
    transaction_count = 0
    last_stats_time = time.time()
    
    try:
        while running:
            # Generate and send transaction
            tx, tx_type, amount = generate_realistic_transaction()
            send_transaction(tx, tx_type, amount)
            transaction_count += 1
            
            # Generate burst activity occasionally
            if transaction_count % 20 == 0:  # Every 20 transactions
                generate_burst_activity()
            
            # Print statistics every 5 minutes
            if time.time() - last_stats_time > 300:
                print_statistics()
                last_stats_time = time.time()
            
            # Sleep with some randomness (realistic timing)
            sleep_time = TX_INTERVAL + random.uniform(-1, 1)
            time.sleep(max(1, sleep_time))
            
    except KeyboardInterrupt:
        log_message("Interrupted by user")
    except Exception as e:
        log_message(f"Unexpected error: {e}", "ERROR")
    finally:
        print_statistics()
        log_message("🛑 Transaction generator stopped")

if __name__ == "__main__":
    main()
