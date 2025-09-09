#!/bin/bash
# LightChain L2 Network Lifecycle Management
# Handles network startup, continuous operation, and controlled upgrades

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Configuration
NETWORK_NAME="LightChain L2"
DATA_DIR="./data"
UPGRADE_FLAG_FILE="./NETWORK_UPGRADE_REQUIRED"
MAINTENANCE_FLAG_FILE="./MAINTENANCE_MODE"
STATUS_FILE="./network_status.json"

# Logging function
log() {
    local level=$1
    shift
    local message="$*"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    case $level in
        "INFO")  echo -e "${GREEN}[${timestamp}] INFO:${NC} $message" ;;
        "WARN")  echo -e "${YELLOW}[${timestamp}] WARN:${NC} $message" ;;
        "ERROR") echo -e "${RED}[${timestamp}] ERROR:${NC} $message" ;;
        "DEBUG") echo -e "${BLUE}[${timestamp}] DEBUG:${NC} $message" ;;
        *)       echo -e "${PURPLE}[${timestamp}] $level:${NC} $message" ;;
    esac
}

# Update network status
update_status() {
    local status=$1
    local message=$2
    cat > $STATUS_FILE << EOF
{
  "network": "$NETWORK_NAME",
  "status": "$status",
  "message": "$message",
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "uptime": "$(uptime -p 2>/dev/null || echo 'N/A')",
  "version": "v0.1.0"
}
EOF
}

# Check if network upgrade is required
check_upgrade_required() {
    if [ -f "$UPGRADE_FLAG_FILE" ]; then
        log "WARN" "Network upgrade flag detected: $UPGRADE_FLAG_FILE"
        return 0
    fi
    return 1
}

# Check if maintenance mode is enabled
check_maintenance_mode() {
    if [ -f "$MAINTENANCE_FLAG_FILE" ]; then
        log "WARN" "Maintenance mode flag detected: $MAINTENANCE_FLAG_FILE"
        return 0
    fi
    return 1
}

# Start the network
start_network() {
    log "INFO" "🚀 Starting $NETWORK_NAME..."
    update_status "starting" "Network initialization in progress"
    
    # Ensure data directories exist
    mkdir -p $DATA_DIR/{validator,sequencer,archive}/{state,logs}
    
    # Start all services
    docker-compose up -d
    
    # Wait for services to be healthy
    log "INFO" "⏳ Waiting for services to become healthy..."
    local max_wait=300  # 5 minutes
    local wait_time=0
    
    while [ $wait_time -lt $max_wait ]; do
        if docker-compose ps | grep -q "healthy"; then
            log "INFO" "✅ Network services are healthy"
            update_status "running" "Network is operational and mining blocks"
            return 0
        fi
        sleep 10
        wait_time=$((wait_time + 10))
        log "DEBUG" "Waiting for health checks... (${wait_time}s/${max_wait}s)"
    done
    
    log "ERROR" "❌ Network failed to start within $max_wait seconds"
    update_status "failed" "Network startup failed - health checks did not pass"
    return 1
}

# Monitor network health
monitor_network() {
    log "INFO" "👀 Starting network health monitoring..."
    
    while true; do
        if check_upgrade_required; then
            log "WARN" "🔄 Network upgrade required. Initiating graceful shutdown..."
            update_status "upgrading" "Network upgrade in progress"
            graceful_upgrade
            break
        fi
        
        if check_maintenance_mode; then
            log "WARN" "🔧 Maintenance mode enabled. Pausing operations..."
            update_status "maintenance" "Network in maintenance mode"
            wait_for_maintenance_end
            continue
        fi
        
        # Check service health
        if ! docker-compose ps | grep -q "Up"; then
            log "ERROR" "❌ Some services are down. Attempting restart..."
            update_status "degraded" "Some services are down, attempting recovery"
            docker-compose restart
            sleep 30
        fi
        
        # Check block progression
        local current_block=$(curl -s -X POST -H "Content-Type: application/json" \
            --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
            http://localhost:8545 2>/dev/null | jq -r '.result // "0x0"' | xargs printf "%d\n" 2>/dev/null || echo "0")
        
        if [ "$current_block" -gt 0 ]; then
            log "DEBUG" "🧱 Current block: #$current_block"
            update_status "running" "Network operational - Block #$current_block"
        else
            log "WARN" "⚠️  No block progression detected"
        fi
        
        sleep 60  # Check every minute
    done
}

# Wait for maintenance mode to end
wait_for_maintenance_end() {
    log "INFO" "⏸️  Network in maintenance mode. Waiting for completion..."
    
    while check_maintenance_mode; do
        log "DEBUG" "Still in maintenance mode..."
        sleep 30
    done
    
    log "INFO" "▶️  Maintenance mode ended. Resuming operations..."
    update_status "running" "Network resumed after maintenance"
}

# Graceful network upgrade
graceful_upgrade() {
    log "INFO" "🔄 Starting graceful network upgrade..."
    
    # Stop accepting new transactions (this would be implemented in the actual blockchain)
    log "INFO" "🛑 Stopping transaction acceptance..."
    
    # Wait for pending transactions to complete
    log "INFO" "⏳ Waiting for pending transactions to complete..."
    sleep 30
    
    # Create backup
    log "INFO" "💾 Creating pre-upgrade backup..."
    backup_network_state
    
    # Stop services gracefully
    log "INFO" "🔽 Stopping network services gracefully..."
    docker-compose down --timeout 60
    
    # Remove upgrade flag
    rm -f "$UPGRADE_FLAG_FILE"
    
    log "INFO" "✅ Network upgrade completed. Ready for new version deployment."
    update_status "stopped" "Network stopped for upgrade - ready for new version"
}

# Backup network state
backup_network_state() {
    local backup_dir="./backups/$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$backup_dir"
    
    log "INFO" "📦 Creating backup in $backup_dir..."
    
    # Backup data directories
    if [ -d "$DATA_DIR" ]; then
        cp -r "$DATA_DIR" "$backup_dir/"
    fi
    
    # Backup configuration
    cp -r configs "$backup_dir/"
    cp -r keys "$backup_dir/"
    
    # Create backup manifest
    cat > "$backup_dir/manifest.json" << EOF
{
  "backup_time": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "network_version": "v0.1.0",
  "data_size": "$(du -sh $DATA_DIR 2>/dev/null | cut -f1 || echo 'unknown')",
  "backup_type": "pre_upgrade"
}
EOF
    
    log "INFO" "✅ Backup completed: $backup_dir"
}

# Trigger network upgrade
trigger_upgrade() {
    log "INFO" "🔄 Triggering network upgrade..."
    echo "Network upgrade requested at $(date)" > "$UPGRADE_FLAG_FILE"
    log "INFO" "✅ Upgrade flag created. Network will shutdown gracefully."
}

# Enable maintenance mode
enable_maintenance() {
    log "INFO" "🔧 Enabling maintenance mode..."
    echo "Maintenance mode enabled at $(date)" > "$MAINTENANCE_FLAG_FILE"
    log "INFO" "✅ Maintenance mode enabled."
}

# Disable maintenance mode
disable_maintenance() {
    log "INFO" "▶️  Disabling maintenance mode..."
    rm -f "$MAINTENANCE_FLAG_FILE"
    log "INFO" "✅ Maintenance mode disabled."
}

# Show network status
show_status() {
    echo -e "${BLUE}=== $NETWORK_NAME Status ===${NC}"
    
    if [ -f "$STATUS_FILE" ]; then
        cat "$STATUS_FILE" | jq '.' 2>/dev/null || cat "$STATUS_FILE"
    else
        echo "Status file not found"
    fi
    
    echo -e "\n${BLUE}=== Docker Services ===${NC}"
    docker-compose ps 2>/dev/null || echo "Docker Compose not running"
    
    echo -e "\n${BLUE}=== Resource Usage ===${NC}"
    docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}" 2>/dev/null || echo "No containers running"
}

# Show help
show_help() {
    cat << EOF
${BLUE}$NETWORK_NAME Lifecycle Management${NC}

Usage: $0 [COMMAND]

Commands:
  start              Start the network
  stop               Stop the network gracefully
  restart            Restart the network
  status             Show network status
  monitor            Start continuous monitoring
  upgrade            Trigger network upgrade
  maintenance on     Enable maintenance mode
  maintenance off    Disable maintenance mode
  backup             Create manual backup
  logs               Show network logs
  help               Show this help

Examples:
  $0 start           # Start the network
  $0 monitor         # Start monitoring (runs continuously)
  $0 upgrade         # Trigger graceful upgrade
  $0 status          # Check current status

The network will run continuously until 'upgrade' is triggered.
EOF
}

# Main command handler
case "${1:-help}" in
    "start")
        start_network
        ;;
    "stop")
        log "INFO" "🔽 Stopping $NETWORK_NAME..."
        update_status "stopping" "Network shutdown in progress"
        docker-compose down --timeout 60
        update_status "stopped" "Network stopped manually"
        ;;
    "restart")
        log "INFO" "🔄 Restarting $NETWORK_NAME..."
        docker-compose restart
        ;;
    "status")
        show_status
        ;;
    "monitor")
        monitor_network
        ;;
    "upgrade")
        trigger_upgrade
        ;;
    "maintenance")
        case "${2:-}" in
            "on") enable_maintenance ;;
            "off") disable_maintenance ;;
            *) log "ERROR" "Usage: $0 maintenance [on|off]" ;;
        esac
        ;;
    "backup")
        backup_network_state
        ;;
    "logs")
        docker-compose logs -f --tail=50
        ;;
    "help"|*)
        show_help
        ;;
esac
