#!/bin/bash
# LightChain L2 Multi-Window Blockchain Monitor
# Opens multiple terminal windows to monitor different aspects of the blockchain

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}🚀 LightChain L2 Blockchain Monitor${NC}"
echo -e "${GREEN}Setting up multiple monitoring windows...${NC}"
echo ""

# Function to open terminal with specific command
open_terminal_with_cmd() {
    local title="$1"
    local cmd="$2"
    local color="$3"
    
    echo -e "${color}📺 Opening: $title${NC}"
    
    # For macOS (Terminal.app)
    if [[ "$OSTYPE" == "darwin"* ]]; then
        osascript -e "tell application \"Terminal\" to do script \"cd '$PWD' && echo -e '\\033[1;36m=== $title ===\\033[0m' && $cmd\""
    # For Linux (gnome-terminal)
    elif command -v gnome-terminal >/dev/null 2>&1; then
        gnome-terminal --title="$title" -- bash -c "cd '$PWD' && echo -e '\\033[1;36m=== $title ===\\033[0m' && $cmd; exec bash"
    # For other Linux terminals
    elif command -v xterm >/dev/null 2>&1; then
        xterm -title "$title" -e "cd '$PWD' && echo -e '\\033[1;36m=== $title ===\\033[0m' && $cmd; exec bash" &
    else
        echo -e "${RED}⚠️  Could not detect terminal. Please run manually: $cmd${NC}"
    fi
    
    sleep 1
}

# Check if docker-compose is available
if ! command -v docker-compose >/dev/null 2>&1; then
    echo -e "${RED}❌ docker-compose not found. Please install Docker Compose.${NC}"
    exit 1
fi

# Check if blockchain is running
echo -e "${YELLOW}🔍 Checking blockchain status...${NC}"
if ! docker-compose ps | grep -q "Up"; then
    echo -e "${RED}❌ Blockchain not running. Starting it now...${NC}"
    docker-compose up -d
    echo -e "${GREEN}✅ Blockchain started. Waiting 10 seconds for initialization...${NC}"
    sleep 10
fi

echo ""
echo -e "${GREEN}📊 Opening monitoring windows:${NC}"
echo ""

# 1. Validator Nodes Logs
open_terminal_with_cmd "Validator Nodes" "docker-compose logs -f validator-1 validator-2" "${GREEN}"

# 2. Sequencer Logs  
open_terminal_with_cmd "Sequencer Node" "docker-compose logs -f sequencer" "${BLUE}"

# 3. Archive Node Logs
open_terminal_with_cmd "Archive Node" "docker-compose logs -f archive" "${YELLOW}"

# 4. Database & Infrastructure
open_terminal_with_cmd "Infrastructure (DB/Monitoring)" "docker-compose logs -f postgres prometheus grafana" "${RED}"

# 5. Activity Simulator
open_terminal_with_cmd "Blockchain Activity Simulator" "./scripts/simulate-activity.sh" "${GREEN}"

# 6. Service Status Monitor
open_terminal_with_cmd "Service Status" "watch -n 5 'docker-compose ps && echo && echo \"=== Resource Usage ===\" && docker stats --no-stream --format \"table {{.Name}}\\t{{.CPUPerc}}\\t{{.MemUsage}}\\t{{.NetIO}}\"'" "${BLUE}"

echo ""
echo -e "${GREEN}✅ All monitoring windows opened!${NC}"
echo ""
echo -e "${YELLOW}🎯 What you can now see:${NC}"
echo -e "  📺 ${GREEN}Validator Nodes${NC}: Real-time consensus and block validation logs"
echo -e "  📺 ${BLUE}Sequencer Node${NC}: Transaction ordering and batch creation logs" 
echo -e "  📺 ${YELLOW}Archive Node${NC}: Full blockchain history and data availability logs"
echo -e "  📺 ${RED}Infrastructure${NC}: Database, monitoring, and system service logs"
echo -e "  📺 ${GREEN}Activity Simulator${NC}: Continuous transaction generation"
echo -e "  📺 ${BLUE}Service Status${NC}: Live resource usage and health monitoring"
echo ""
echo -e "${BLUE}🌐 Access Points:${NC}"
echo -e "  🔗 Main RPC: ${GREEN}http://localhost:8545${NC}"
echo -e "  📊 Grafana Dashboard: ${GREEN}http://localhost:3000${NC} (admin/admin123)"
echo -e "  📈 Prometheus Metrics: ${GREEN}http://localhost:9090${NC}"
echo ""
echo -e "${YELLOW}💡 Tips:${NC}"
echo -e "  • Watch the ${GREEN}Activity Simulator${NC} window for realistic blockchain transactions"
echo -e "  • Monitor ${BLUE}Validator Nodes${NC} for consensus activity and block creation"
echo -e "  • Check ${RED}Service Status${NC} for resource usage and health"
echo -e "  • Visit ${GREEN}Grafana Dashboard${NC} for beautiful visualizations"
echo ""
echo -e "${GREEN}🎉 Your LightChain L2 blockchain is now alive with constant activity!${NC}"
