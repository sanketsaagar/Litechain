package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/yourusername/lightchain-l2/internal/config"
	"github.com/yourusername/lightchain-l2/internal/node"
	"github.com/yourusername/lightchain-l2/pkg/agglayer"
	"github.com/yourusername/lightchain-l2/pkg/consensus"
	"github.com/yourusername/lightchain-l2/pkg/networking"
	"github.com/yourusername/lightchain-l2/pkg/state"
)

const (
	defaultConfigPath = "configs/validator.yaml"
	appName           = "LightChain L2"
	version           = "v0.1.0"
)

func main() {
	var (
		configPath  = flag.String("config", defaultConfigPath, "Path to configuration file")
		nodeType    = flag.String("type", "validator", "Node type: validator, sequencer, archive")
		showVersion = flag.Bool("version", false, "Show version information")
		dataDir     = flag.String("data-dir", "./data", "Data directory for blockchain storage")
		logLevel    = flag.String("log-level", "info", "Log level: debug, info, warn, error")
	)
	flag.Parse()

	if *showVersion {
		fmt.Printf("%s %s\n", appName, version)
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Override config with command line flags
	if *dataDir != "./data" {
		cfg.DataDir = *dataDir
	}
	if *logLevel != "info" {
		cfg.LogLevel = *logLevel
	}
	cfg.NodeType = *nodeType

	// Validate node type
	if !isValidNodeType(cfg.NodeType) {
		log.Fatalf("Invalid node type: %s. Valid types are: validator, sequencer, archive", cfg.NodeType)
	}

	// Create data directory if it doesn't exist
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	// Initialize logger
	logger := setupLogger(cfg.LogLevel)
	logger.Printf("Starting %s %s", appName, version)
	logger.Printf("Node type: %s", cfg.NodeType)
	logger.Printf("Data directory: %s", cfg.DataDir)

	// Create main context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize components
	stateManager, err := state.New(cfg.State)
	if err != nil {
		log.Fatalf("Failed to initialize state manager: %v", err)
	}

	networkManager, err := networking.New(cfg.Network)
	if err != nil {
		log.Fatalf("Failed to initialize network manager: %v", err)
	}

	consensusEngine, err := consensus.New(cfg.Consensus, stateManager)
	if err != nil {
		log.Fatalf("Failed to initialize consensus engine: %v", err)
	}

	// Initialize AggLayer integration if enabled
	var aggLayerClient *agglayer.Client
	if cfg.AggLayer.Enabled {
		aggLayerClient, err = agglayer.NewClient(cfg.AggLayer)
		if err != nil {
			log.Fatalf("Failed to initialize AggLayer client: %v", err)
		}
		logger.Printf("AggLayer integration enabled")
	}

	// Create and start the node
	nodeInstance, err := node.New(&node.Config{
		Type:           cfg.NodeType,
		DataDir:        cfg.DataDir,
		State:          stateManager,
		Network:        networkManager,
		Consensus:      consensusEngine,
		AggLayerClient: aggLayerClient,
		Logger:         logger,
	})
	if err != nil {
		log.Fatalf("Failed to create node: %v", err)
	}

	// Start the node
	if err := nodeInstance.Start(ctx); err != nil {
		log.Fatalf("Failed to start node: %v", err)
	}

	logger.Printf("LightChain L2 node started successfully")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	logger.Printf("Shutting down LightChain L2 node...")

	// Graceful shutdown
	if err := nodeInstance.Stop(); err != nil {
		logger.Printf("Error during shutdown: %v", err)
	}

	logger.Printf("LightChain L2 node stopped")
}

func isValidNodeType(nodeType string) bool {
	validTypes := []string{"validator", "sequencer", "archive"}
	for _, validType := range validTypes {
		if nodeType == validType {
			return true
		}
	}
	return false
}

func setupLogger(level string) *log.Logger {
	// For now, use standard logger. In production, use structured logging like zap or logrus
	logger := log.New(os.Stdout, fmt.Sprintf("[%s] ", appName), log.LstdFlags|log.Lshortfile)
	return logger
}

func init() {
	// Set the working directory to the project root if running from cmd/lightchain
	if wd, err := os.Getwd(); err == nil {
		if filepath.Base(wd) == "lightchain" {
			if err := os.Chdir("../.."); err != nil {
				log.Printf("Warning: Failed to change to project root: %v", err)
			}
		}
	}
}
