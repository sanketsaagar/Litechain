package unified

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sanketsaagar/Litechain/pkg/agglayer"
)

// MinimalConfig holds configuration for the minimal blockchain
type MinimalConfig struct {
	NodeType  string
	DataDir   string
	ChainID   uint64
	NetworkID uint64

	// RPC Configuration
	RPCEnabled bool
	RPCHost    string
	RPCPort    int
	WSEnabled  bool
	WSHost     string
	WSPort     int

	// Auto-mining Configuration
	AutoMining  bool
	BlockTime   time.Duration
	EmptyBlocks bool

	// Legacy compatibility
	WorkerCount     int
	TxPoolSize      int
	BlockGasLimit   uint64
	AggLayerEnabled bool
	AggLayerRPC     string
	AggLayerClient  *agglayer.Client
}

// MinimalBlockchain is a simple blockchain implementation for testing RPC
type MinimalBlockchain struct {
	config    *MinimalConfig
	rpcServer *http.Server
	wsServer  *http.Server

	// State
	blockNumber *big.Int
	running     bool
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewMinimalBlockchain creates a minimal blockchain for testing
func NewMinimalBlockchain(config *MinimalConfig) (*MinimalBlockchain, error) {
	return &MinimalBlockchain{
		config:      config,
		blockNumber: big.NewInt(0),
	}, nil
}

// Alias for compatibility
func NewSimpleUnifiedEngine(config interface{}) (*MinimalBlockchain, error) {
	// Convert to our minimal config - just use defaults for now
	minimalConfig := &MinimalConfig{
		NodeType:    "validator",
		ChainID:     1337,
		NetworkID:   1337,
		RPCEnabled:  true,
		RPCHost:     "0.0.0.0",
		RPCPort:     8545,
		WSEnabled:   true,
		WSHost:      "0.0.0.0",
		WSPort:      8546,
		AutoMining:  true,
		BlockTime:   2 * time.Second,
		EmptyBlocks: true,
	}

	// Try to extract values from the passed config if possible
	if configMap, ok := config.(map[string]interface{}); ok {
		if nodeType, exists := configMap["NodeType"]; exists {
			if nt, ok := nodeType.(string); ok {
				minimalConfig.NodeType = nt
			}
		}
		if rpcPort, exists := configMap["RPCPort"]; exists {
			if port, ok := rpcPort.(int); ok {
				minimalConfig.RPCPort = port
			}
		}
		if wsPort, exists := configMap["WSPort"]; exists {
			if port, ok := wsPort.(int); ok {
				minimalConfig.WSPort = port
			}
		}
	}

	return NewMinimalBlockchain(minimalConfig)
}

// Start starts the minimal blockchain
func (bc *MinimalBlockchain) Start(ctx context.Context) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	if bc.running {
		return fmt.Errorf("blockchain already running")
	}

	bc.ctx, bc.cancel = context.WithCancel(ctx)

	fmt.Println("🚀 Starting LightChain L2 Minimal Blockchain")
	fmt.Printf("   • Node Type: %s\n", bc.config.NodeType)
	fmt.Printf("   • Chain ID: %d\n", bc.config.ChainID)

	// Start RPC servers
	if bc.config.RPCEnabled {
		if err := bc.startRPCServer(); err != nil {
			return fmt.Errorf("failed to start RPC server: %w", err)
		}
	}
	if bc.config.WSEnabled {
		if err := bc.startWSServer(); err != nil {
			return fmt.Errorf("failed to start WebSocket server: %w", err)
		}
	}

	bc.running = true

	// Start auto-mining if enabled
	if bc.config.AutoMining {
		go bc.autoMiningLoop()
		fmt.Printf("⛏️  Auto-mining started: blocks every %v\n", bc.config.BlockTime)
	}

	fmt.Println("✅ LightChain L2 Minimal Blockchain started successfully")
	return nil
}

// autoMiningLoop continuously mines new blocks
func (bc *MinimalBlockchain) autoMiningLoop() {
	ticker := time.NewTicker(bc.config.BlockTime)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			bc.mineBlock()
		case <-bc.ctx.Done():
			return
		}
	}
}

// mineBlock creates a new block
func (bc *MinimalBlockchain) mineBlock() {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	bc.blockNumber.Add(bc.blockNumber, big.NewInt(1))

	fmt.Printf("⛏️  Block #%d mined: gas: 420000, validator: %s\n",
		bc.blockNumber.Uint64(),
		"0x742A4D1A0Ac05A73A48F10C2E2d6b0E3f1b2e3F1")
}

// RPC Server implementation
func (bc *MinimalBlockchain) startRPCServer() error {
	addr := fmt.Sprintf("%s:%d", bc.config.RPCHost, bc.config.RPCPort)

	mux := http.NewServeMux()
	mux.HandleFunc("/", bc.handleRPCRequest)
	mux.HandleFunc("/health", bc.handleHealthCheck)

	bc.rpcServer = &http.Server{
		Addr:    addr,
		Handler: bc.corsMiddleware(mux),
	}

	go func() {
		fmt.Printf("🌐 RPC Server started on http://%s\n", addr)
		if err := bc.rpcServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("❌ RPC Server error: %v\n", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	return nil
}

func (bc *MinimalBlockchain) startWSServer() error {
	addr := fmt.Sprintf("%s:%d", bc.config.WSHost, bc.config.WSPort)

	mux := http.NewServeMux()
	mux.HandleFunc("/", bc.handleWSConnection)

	bc.wsServer = &http.Server{
		Addr:    addr,
		Handler: bc.corsMiddleware(mux),
	}

	go func() {
		fmt.Printf("📡 WebSocket Server started on ws://%s\n", addr)
		if err := bc.wsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("❌ WebSocket Server error: %v\n", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	return nil
}

func (bc *MinimalBlockchain) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RPC Types
type MinimalRPCRequest struct {
	Version string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      interface{}   `json:"id"`
}

type MinimalRPCResponse struct {
	Version string           `json:"jsonrpc"`
	Result  interface{}      `json:"result,omitempty"`
	Error   *MinimalRPCError `json:"error,omitempty"`
	ID      interface{}      `json:"id"`
}

type MinimalRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (bc *MinimalBlockchain) handleRPCRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MinimalRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		bc.sendRPCError(w, req.ID, -32700, "Parse error")
		return
	}

	fmt.Printf("📥 RPC Request: %s\n", req.Method)

	result, err := bc.processRPCMethod(req.Method, req.Params)
	if err != nil {
		bc.sendRPCError(w, req.ID, -32000, err.Error())
		return
	}

	response := MinimalRPCResponse{
		Version: "2.0",
		Result:  result,
		ID:      req.ID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (bc *MinimalBlockchain) processRPCMethod(method string, params []interface{}) (interface{}, error) {
	switch method {
	case "eth_chainId":
		return fmt.Sprintf("0x%x", bc.config.ChainID), nil

	case "eth_blockNumber":
		bc.mu.RLock()
		blockNum := bc.blockNumber
		bc.mu.RUnlock()
		return fmt.Sprintf("0x%x", blockNum.Uint64()), nil

	case "eth_getBalance":
		return "0x56bc75e2d630e8000", nil // 100 ETH in wei (hex)

	case "eth_gasPrice":
		return "0x3b9aca00", nil // 1 Gwei

	case "net_version":
		return fmt.Sprintf("%d", bc.config.NetworkID), nil

	case "web3_clientVersion":
		return "LightChain-L2/v0.1.0", nil

	case "eth_sendTransaction":
		txHash := common.BytesToHash([]byte(fmt.Sprintf("tx_%d", time.Now().UnixNano())))
		fmt.Printf("📤 Transaction received: %s\n", txHash.Hex())
		return txHash.Hex(), nil

	case "eth_getTransactionCount":
		return "0x0", nil

	case "eth_estimateGas":
		return "0x5208", nil // 21000 gas

	case "lightchain_mineBlock":
		bc.mineBlock()
		return "Block mined successfully", nil

	case "lightchain_status":
		return map[string]interface{}{
			"nodeType":    bc.config.NodeType,
			"chainId":     bc.config.ChainID,
			"blockNumber": bc.blockNumber.Uint64(),
			"autoMining":  bc.config.AutoMining,
			"blockTime":   bc.config.BlockTime.String(),
			"running":     bc.running,
		}, nil

	default:
		return nil, fmt.Errorf("method %s not found", method)
	}
}

func (bc *MinimalBlockchain) sendRPCError(w http.ResponseWriter, id interface{}, code int, message string) {
	response := MinimalRPCResponse{
		Version: "2.0",
		Error: &MinimalRPCError{
			Code:    code,
			Message: message,
		},
		ID: id,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (bc *MinimalBlockchain) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":      "healthy",
		"nodeType":    bc.config.NodeType,
		"blockNumber": bc.blockNumber.Uint64(),
		"rpcEnabled":  bc.config.RPCEnabled,
		"wsEnabled":   bc.config.WSEnabled,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

func (bc *MinimalBlockchain) handleWSConnection(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"message": "WebSocket endpoint ready",
		"status":  "available",
	}
	json.NewEncoder(w).Encode(response)
}

// Stop stops the blockchain
func (bc *MinimalBlockchain) Stop() error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	if !bc.running {
		return nil
	}

	fmt.Println("🛑 Stopping LightChain L2 Minimal Blockchain...")

	if bc.cancel != nil {
		bc.cancel()
	}

	if bc.rpcServer != nil {
		bc.rpcServer.Shutdown(context.Background())
	}
	if bc.wsServer != nil {
		bc.wsServer.Shutdown(context.Background())
	}

	bc.running = false
	fmt.Println("✅ LightChain L2 Minimal Blockchain stopped")
	return nil
}

