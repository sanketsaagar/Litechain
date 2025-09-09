#!/bin/bash

# LightChain L2 Development Environment Setup Script

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_NAME="LightChain L2"
REQUIRED_GO_VERSION="1.21"
DATA_DIR="./data"
KEYS_DIR="./keys"

echo -e "${BLUE}=== $PROJECT_NAME Development Setup ===${NC}"

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Go is installed
check_go() {
    print_status "Checking Go installation..."
    
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go $REQUIRED_GO_VERSION or later."
        echo "Visit: https://golang.org/dl/"
        exit 1
    fi
    
    # Check Go version
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    REQUIRED_VERSION_NUM=$(echo $REQUIRED_GO_VERSION | sed 's/\.//')
    CURRENT_VERSION_NUM=$(echo $GO_VERSION | cut -d. -f1,2 | sed 's/\.//')
    
    if [ "$CURRENT_VERSION_NUM" -lt "$REQUIRED_VERSION_NUM" ]; then
        print_error "Go version $GO_VERSION is too old. Please upgrade to $REQUIRED_GO_VERSION or later."
        exit 1
    fi
    
    print_status "Go version $GO_VERSION is compatible"
}

# Check if Docker is installed (optional)
check_docker() {
    print_status "Checking Docker installation..."
    
    if command -v docker &> /dev/null; then
        print_status "Docker is installed"
        
        # Check if Docker daemon is running
        if docker info &> /dev/null; then
            print_status "Docker daemon is running"
        else
            print_warning "Docker is installed but daemon is not running"
        fi
    else
        print_warning "Docker is not installed. Some features may not be available."
        echo "To install Docker, visit: https://docs.docker.com/get-docker/"
    fi
}

# Check if Kurtosis is installed (optional)
check_kurtosis() {
    print_status "Checking Kurtosis installation..."
    
    if command -v kurtosis &> /dev/null; then
        print_status "Kurtosis is installed"
    else
        print_warning "Kurtosis is not installed. Devnet features may not be available."
        echo "To install Kurtosis, visit: https://docs.kurtosis.com/install"
    fi
}

# Install Go dependencies
install_dependencies() {
    print_status "Installing Go dependencies..."
    
    go mod download
    go mod tidy
    go mod verify
    
    print_status "Dependencies installed successfully"
}

# Install development tools
install_dev_tools() {
    print_status "Installing development tools..."
    
    # Install linter
    if ! command -v golangci-lint &> /dev/null; then
        print_status "Installing golangci-lint..."
        go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    fi
    
    # Install goimports
    if ! command -v goimports &> /dev/null; then
        print_status "Installing goimports..."
        go install golang.org/x/tools/cmd/goimports@latest
    fi
    
    # Install security scanner
    if ! command -v gosec &> /dev/null; then
        print_status "Installing gosec..."
        go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
    fi
    
    print_status "Development tools installed successfully"
}

# Create necessary directories
create_directories() {
    print_status "Creating project directories..."
    
    mkdir -p $DATA_DIR/{validator,sequencer,archive}/{state,logs}
    mkdir -p $KEYS_DIR
    mkdir -p ./build
    mkdir -p ./releases
    mkdir -p ./logs
    
    print_status "Directories created successfully"
}

# Generate example keys
generate_keys() {
    print_status "Generating example keys..."
    
    # Create example key files
    cat > $KEYS_DIR/example-validator.key << EOF
{
  "address": "742a4d1a0ac05a73a48f10c2e2d6b0e3f1b2e3f4",
  "private_key": "0x0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
  "public_key": "0x0459d7455d76b4e123c5f7e8b0a9b8c7d6e5f4a3b2c1d0e9f8a7b6c5d4e3f2a1b0"
}
EOF

    cat > $KEYS_DIR/example-sequencer.key << EOF
{
  "address": "8b3a4d1a0ac05a73a48f10c2e2d6b0e3f1b2e3f5",
  "private_key": "0x1123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
  "public_key": "0x0459d7455d76b4e123c5f7e8b0a9b8c7d6e5f4a3b2c1d0e9f8a7b6c5d4e3f2a1b1"
}
EOF

    cat > $KEYS_DIR/password.txt << EOF
development_password_123
EOF

    chmod 600 $KEYS_DIR/*.key
    chmod 600 $KEYS_DIR/password.txt
    
    print_status "Example keys generated successfully"
}

# Build the project
build_project() {
    print_status "Building LightChain L2..."
    
    make build
    
    if [ $? -eq 0 ]; then
        print_status "Build completed successfully"
    else
        print_error "Build failed"
        exit 1
    fi
}

# Run tests
run_tests() {
    print_status "Running tests..."
    
    make test
    
    if [ $? -eq 0 ]; then
        print_status "All tests passed"
    else
        print_warning "Some tests failed. Please check the output above."
    fi
}

# Create example environment file
create_env_file() {
    print_status "Creating example environment file..."
    
    cat > .env.example << EOF
# LightChain L2 Environment Configuration

# Node Configuration
NODE_TYPE=validator
DATA_DIR=./data
LOG_LEVEL=info

# Network Configuration
NETWORK_LISTEN_ADDR=0.0.0.0:30303
NETWORK_EXTERNAL_ADDR=
MAX_PEERS=50

# RPC Configuration
RPC_LISTEN_ADDR=127.0.0.1:8545
WS_LISTEN_ADDR=127.0.0.1:8546
METRICS_LISTEN_ADDR=127.0.0.1:9090

# Database Configuration
DB_TYPE=badger
DB_PATH=./data/state
DB_CACHE_SIZE=128MB

# AggLayer Configuration
AGGLAYER_ENABLED=true
AGGLAYER_RPC_URL=https://agglayer-rpc.polygon.technology
AGGLAYER_PRIVATE_KEY_PATH=./keys/aggsender.key

# Security Configuration
KEYSTORE_PATH=./keys
PASSWORD_FILE=./keys/password.txt

# Development Configuration (devnet only)
DEV_MODE=true
AUTO_MINE=true
UNLOCK_ACCOUNTS=0x742A4D1A0Ac05A73A48F10C2E2d6b0E3f1b2e3F4,0x8B3A4D1A0Ac05A73A48F10C2E2d6b0E3f1b2e3F5
EOF

    print_status "Environment file created at .env.example"
}

# Print final instructions
print_instructions() {
    echo ""
    echo -e "${GREEN}=== Setup Complete! ===${NC}"
    echo ""
    echo "Next steps:"
    echo "1. Copy .env.example to .env and customize as needed:"
    echo "   cp .env.example .env"
    echo ""
    echo "2. Start a single validator node:"
    echo "   make run-validator"
    echo ""
    echo "3. Start the full development network with Docker:"
    echo "   docker-compose up -d"
    echo ""
    echo "4. Start the development network with Kurtosis:"
    echo "   make dev-network"
    echo ""
    echo "5. Run tests:"
    echo "   make test"
    echo ""
    echo "6. Build for production:"
    echo "   make build-all"
    echo ""
    echo "Useful commands:"
    echo "  make help           - Show all available commands"
    echo "  make dev-setup      - Re-run development setup"
    echo "  make clean          - Clean build artifacts"
    echo "  make lint           - Run code linter"
    echo "  make format         - Format code"
    echo ""
    echo "Documentation:"
    echo "  README.md           - Project overview and quick start"
    echo "  docs/               - Detailed documentation"
    echo ""
    echo -e "${BLUE}Happy coding with $PROJECT_NAME!${NC}"
}

# Main execution
main() {
    check_go
    check_docker
    check_kurtosis
    install_dependencies
    install_dev_tools
    create_directories
    generate_keys
    create_env_file
    build_project
    run_tests
    print_instructions
}

# Handle script interruption
trap 'print_error "Setup interrupted"; exit 1' INT TERM

# Run main function
main
