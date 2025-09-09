#!/bin/bash
# LightChain L2 Auto-Start Script
# Automatically starts the blockchain when Docker starts

set -e

# Configuration
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOG_FILE="$PROJECT_DIR/logs/auto-start.log"

# Ensure log directory exists
mkdir -p "$PROJECT_DIR/logs"

# Logging function
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*" | tee -a "$LOG_FILE"
}

log "🚀 LightChain L2 Auto-Start initiated"
log "📁 Project directory: $PROJECT_DIR"

# Change to project directory
cd "$PROJECT_DIR"

# Wait for Docker to be ready
log "⏳ Waiting for Docker daemon to be ready..."
while ! docker info >/dev/null 2>&1; do
    sleep 5
done
log "✅ Docker daemon is ready"

# Start the network
log "🌐 Starting LightChain L2 network..."
if ./scripts/network-lifecycle.sh start; then
    log "✅ LightChain L2 network started successfully"
    
    # Start continuous monitoring
    log "👀 Starting continuous monitoring..."
    ./scripts/network-lifecycle.sh monitor &
    
    log "🎉 LightChain L2 is now running continuously"
    log "📊 Access points:"
    log "   - RPC: http://localhost:8545"
    log "   - Grafana: http://localhost:3000 (admin/admin123)"
    log "   - Prometheus: http://localhost:9090"
else
    log "❌ Failed to start LightChain L2 network"
    exit 1
fi
