#!/bin/bash
# Docker-Kurtosis Bridge for LightChain L2
# Seamlessly switch between Docker Compose and Kurtosis environments

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
DOCKER_COMPOSE_FILE="docker-compose.yml"
KURTOSIS_SCRIPT="scripts/kurtosis-manager.sh"

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
        *)       echo -e "${PURPLE}[${timestamp}] $level:${NC} $message" ;;
    esac
}

# Check what's currently running
check_environment() {
    local docker_running=false
    local kurtosis_running=false
    
    # Check Docker Compose
    if docker-compose ps 2>/dev/null | grep -q "Up"; then
        docker_running=true
    fi
    
    # Check Kurtosis
    if command -v kurtosis &> /dev/null && kurtosis enclave ls 2>/dev/null | grep -q "lightchain-devnet"; then
        kurtosis_running=true
    fi
    
    echo "docker_running:$docker_running,kurtosis_running:$kurtosis_running"
}

# Stop all environments
stop_all() {
    log "INFO" "🛑 Stopping all LightChain L2 environments..."
    
    # Stop Docker Compose
    if docker-compose ps 2>/dev/null | grep -q "Up"; then
        log "INFO" "Stopping Docker Compose environment..."
        docker-compose down --timeout 30
    fi
    
    # Stop Kurtosis
    if command -v kurtosis &> /dev/null && kurtosis enclave ls 2>/dev/null | grep -q "lightchain-devnet"; then
        log "INFO" "Stopping Kurtosis environment..."
        kurtosis enclave stop lightchain-devnet
    fi
    
    log "SUCCESS" "All environments stopped"
}

# Start Docker environment
start_docker() {
    log "INFO" "🐳 Starting Docker Compose environment..."
    
    # Stop Kurtosis if running
    if command -v kurtosis &> /dev/null && kurtosis enclave ls 2>/dev/null | grep -q "lightchain-devnet"; then
        log "INFO" "Stopping Kurtosis environment first..."
        kurtosis enclave stop lightchain-devnet
    fi
    
    # Start Docker Compose
    ./scripts/network-lifecycle.sh start
    
    log "SUCCESS" "Docker environment started!"
    log "INFO" "🌐 Access points:"
    log "INFO" "   • RPC: http://localhost:8545"
    log "INFO" "   • Grafana: http://localhost:3000 (admin/admin123)"
    log "INFO" "   • Prometheus: http://localhost:9090"
}

# Start Kurtosis environment
start_kurtosis() {
    local validators=${1:-3}
    local sequencers=${2:-1}
    local archives=${3:-1}
    
    log "INFO" "🎯 Starting Kurtosis DevNet environment..."
    
    # Stop Docker if running
    if docker-compose ps 2>/dev/null | grep -q "Up"; then
        log "INFO" "Stopping Docker Compose environment first..."
        docker-compose down --timeout 30
    fi
    
    # Start Kurtosis
    $KURTOSIS_SCRIPT start $validators $sequencers $archives
    
    log "SUCCESS" "Kurtosis environment started!"
    $KURTOSIS_SCRIPT access
}

# Switch between environments
switch() {
    local target=${1:-"auto"}
    local env_status=$(check_environment)
    local docker_running=$(echo $env_status | cut -d',' -f1 | cut -d':' -f2)
    local kurtosis_running=$(echo $env_status | cut -d',' -f2 | cut -d':' -f2)
    
    log "INFO" "🔄 Switching LightChain L2 environment..."
    log "INFO" "Current state: Docker=$docker_running, Kurtosis=$kurtosis_running"
    
    case $target in
        "docker")
            start_docker
            ;;
        "kurtosis")
            start_kurtosis "${2:-3}" "${3:-1}" "${4:-1}"
            ;;
        "auto")
            if [ "$docker_running" = "true" ]; then
                log "INFO" "Docker is running, switching to Kurtosis..."
                start_kurtosis
            elif [ "$kurtosis_running" = "true" ]; then
                log "INFO" "Kurtosis is running, switching to Docker..."
                start_docker
            else
                log "INFO" "Nothing running, starting Docker (default)..."
                start_docker
            fi
            ;;
        *)
            log "ERROR" "Invalid target: $target. Use 'docker', 'kurtosis', or 'auto'"
            exit 1
            ;;
    esac
}

# Show current status
status() {
    log "INFO" "📊 LightChain L2 Environment Status"
    echo ""
    
    local env_status=$(check_environment)
    local docker_running=$(echo $env_status | cut -d',' -f1 | cut -d':' -f2)
    local kurtosis_running=$(echo $env_status | cut -d',' -f2 | cut -d':' -f2)
    
    echo -e "${BLUE}=== Environment Status ===${NC}"
    if [ "$docker_running" = "true" ]; then
        echo -e "🐳 Docker Compose: ${GREEN}RUNNING${NC}"
        echo "   Services:"
        docker-compose ps --format="table {{.Name}}\t{{.Status}}" | grep -v NAME | sed 's/^/   /'
    else
        echo -e "🐳 Docker Compose: ${RED}STOPPED${NC}"
    fi
    
    echo ""
    if [ "$kurtosis_running" = "true" ]; then
        echo -e "🎯 Kurtosis DevNet: ${GREEN}RUNNING${NC}"
        if command -v kurtosis &> /dev/null; then
            echo "   Services:"
            kurtosis service ls lightchain-devnet 2>/dev/null | tail -n +2 | sed 's/^/   /' || echo "   (Unable to list services)"
        fi
    else
        echo -e "🎯 Kurtosis DevNet: ${RED}STOPPED${NC}"
    fi
    
    echo ""
    if [ "$docker_running" = "true" ] || [ "$kurtosis_running" = "true" ]; then
        echo -e "${CYAN}🌐 Active Environment:${NC}"
        if [ "$docker_running" = "true" ]; then
            echo "   • Type: Docker Compose"
            echo "   • RPC: http://localhost:8545"
            echo "   • Grafana: http://localhost:3000"
            echo "   • Management: ./scripts/network-lifecycle.sh"
        elif [ "$kurtosis_running" = "true" ]; then
            echo "   • Type: Kurtosis DevNet"
            echo "   • Management: ./scripts/kurtosis-manager.sh"
            echo "   • Access Points: ./scripts/kurtosis-manager.sh access"
        fi
    else
        echo -e "${YELLOW}⚠️  No environment is currently running${NC}"
        echo "   Use: ./scripts/docker-kurtosis-bridge.sh start docker"
        echo "   Or:  ./scripts/docker-kurtosis-bridge.sh start kurtosis"
    fi
}

# Compare environments
compare() {
    $KURTOSIS_SCRIPT compare
}

# Backup current state
backup() {
    local env_status=$(check_environment)
    local docker_running=$(echo $env_status | cut -d',' -f1 | cut -d':' -f2)
    local kurtosis_running=$(echo $env_status | cut -d',' -f2 | cut -d':' -f2)
    
    local backup_dir="./backups/environment_$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$backup_dir"
    
    log "INFO" "💾 Creating environment backup..."
    
    if [ "$docker_running" = "true" ]; then
        log "INFO" "Backing up Docker environment..."
        ./scripts/network-lifecycle.sh backup
        cp -r ./backups/$(ls -t ./backups | head -1)/* "$backup_dir/" 2>/dev/null || true
        echo "docker" > "$backup_dir/environment_type"
    fi
    
    if [ "$kurtosis_running" = "true" ]; then
        log "INFO" "Backing up Kurtosis environment..."
        # Kurtosis data is ephemeral, but we can save configuration
        cp -r ./deployments/kurtosis "$backup_dir/"
        echo "kurtosis" > "$backup_dir/environment_type"
    fi
    
    # Save current configurations
    cp docker-compose.yml "$backup_dir/" 2>/dev/null || true
    cp -r configs "$backup_dir/" 2>/dev/null || true
    cp -r keys "$backup_dir/" 2>/dev/null || true
    
    log "SUCCESS" "Backup created: $backup_dir"
}

# Test current environment
test_env() {
    local env_status=$(check_environment)
    local docker_running=$(echo $env_status | cut -d',' -f1 | cut -d':' -f2)
    local kurtosis_running=$(echo $env_status | cut -d',' -f2 | cut -d':' -f2)
    
    if [ "$docker_running" = "true" ]; then
        log "INFO" "🧪 Testing Docker environment..."
        # Test Docker RPC
        local rpc_url="http://localhost:8545"
        local block_number=$(curl -s -X POST -H "Content-Type: application/json" \
            --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
            "$rpc_url" | jq -r '.result' 2>/dev/null || echo "failed")
        
        if [ "$block_number" != "failed" ]; then
            local block_decimal=$((16#${block_number#0x}))
            log "SUCCESS" "Docker environment is healthy - Block: #$block_decimal"
        else
            log "ERROR" "Docker environment is not responding"
        fi
        
    elif [ "$kurtosis_running" = "true" ]; then
        log "INFO" "🧪 Testing Kurtosis environment..."
        $KURTOSIS_SCRIPT test
        
    else
        log "WARN" "No environment is running to test"
    fi
}

show_help() {
    cat << EOF
${BLUE}LightChain L2 Docker-Kurtosis Bridge${NC}

${YELLOW}USAGE:${NC}
    $0 [COMMAND] [OPTIONS]

${YELLOW}COMMANDS:${NC}
    ${GREEN}start docker${NC}
        Start Docker Compose environment (stops Kurtosis if running)
        
    ${GREEN}start kurtosis [validators] [sequencers] [archives]${NC}
        Start Kurtosis DevNet environment (stops Docker if running)
        Default: 3 validators, 1 sequencer, 1 archive
        
    ${GREEN}switch [auto|docker|kurtosis]${NC}
        Switch between environments (auto detects current and switches)
        
    ${GREEN}stop${NC}
        Stop all environments (Docker and Kurtosis)
        
    ${GREEN}status${NC}
        Show status of both environments
        
    ${GREEN}compare${NC}
        Compare Docker vs Kurtosis features
        
    ${GREEN}backup${NC}
        Backup current environment state
        
    ${GREEN}test${NC}
        Test current environment functionality
        
    ${GREEN}help${NC}
        Show this help message

${YELLOW}EXAMPLES:${NC}
    # Start Docker environment
    $0 start docker
    
    # Start Kurtosis with custom config
    $0 start kurtosis 5 2 1
    
    # Switch to opposite environment
    $0 switch auto
    
    # Check what's running
    $0 status
    
    # Test current environment
    $0 test

${YELLOW}FEATURES:${NC}
    • Seamless switching between Docker and Kurtosis
    • Automatic environment detection
    • State backup and restore
    • Unified testing interface
    • Both environments support continuous operation
    • Both include auto-mining and transaction generation

${YELLOW}USE CASES:${NC}
    • Docker: Quick local development, simple testing
    • Kurtosis: Professional testing, complex scenarios, team environments
EOF
}

# Main command dispatcher
main() {
    case "${1:-help}" in
        "start")
            case "${2:-docker}" in
                "docker")
                    start_docker
                    ;;
                "kurtosis")
                    start_kurtosis "${3:-3}" "${4:-1}" "${5:-1}"
                    ;;
                *)
                    log "ERROR" "Invalid environment: ${2}. Use 'docker' or 'kurtosis'"
                    exit 1
                    ;;
            esac
            ;;
        "switch")
            switch "${2:-auto}" "${3:-3}" "${4:-1}" "${5:-1}"
            ;;
        "stop")
            stop_all
            ;;
        "status")
            status
            ;;
        "compare")
            compare
            ;;
        "backup")
            backup
            ;;
        "test")
            test_env
            ;;
        "help"|*)
            show_help
            ;;
    esac
}

# Run main function
main "$@"
