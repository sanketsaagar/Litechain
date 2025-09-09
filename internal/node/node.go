package node

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/sanketsaagar/Litechain/pkg/agglayer"
)

// NodeType represents the type of node
type NodeType string

const (
	ValidatorNode NodeType = "validator"
	SequencerNode NodeType = "sequencer"
	ArchiveNode   NodeType = "archive"
)

// Config contains node configuration
type Config struct {
	Type           string
	DataDir        string
	State          *state.Manager
	Network        *networking.Manager
	Consensus      *consensus.Engine
	AggLayerClient *agglayer.Client
	Logger         *log.Logger
}

// Node represents a LightChain L2 node
type Node struct {
	config   *Config
	nodeType NodeType

	// Core components
	stateManager    *state.Manager
	networkManager  *networking.Manager
	consensusEngine *consensus.Engine
	aggLayerClient  *agglayer.Client

	// Runtime state
	running bool
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	logger  *log.Logger
}

// New creates a new node instance
func New(config *Config) (*Node, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Validate node type
	nodeType := NodeType(config.Type)
	if !isValidNodeType(nodeType) {
		return nil, fmt.Errorf("invalid node type: %s", config.Type)
	}

	return &Node{
		config:          config,
		nodeType:        nodeType,
		stateManager:    config.State,
		networkManager:  config.Network,
		consensusEngine: config.Consensus,
		aggLayerClient:  config.AggLayerClient,
		logger:          config.Logger,
	}, nil
}

// Start starts the node and all its components
func (n *Node) Start(ctx context.Context) error {
	if n.running {
		return fmt.Errorf("node is already running")
	}

	n.ctx, n.cancel = context.WithCancel(ctx)
	n.running = true

	n.logger.Printf("Starting %s node...", n.nodeType)

	// Start core components in order
	if err := n.startStateManager(); err != nil {
		return fmt.Errorf("failed to start state manager: %w", err)
	}

	if err := n.startNetworkManager(); err != nil {
		return fmt.Errorf("failed to start network manager: %w", err)
	}

	if err := n.startConsensusEngine(); err != nil {
		return fmt.Errorf("failed to start consensus engine: %w", err)
	}

	// Start AggLayer integration if enabled
	if n.aggLayerClient != nil {
		if err := n.startAggLayerIntegration(); err != nil {
			return fmt.Errorf("failed to start AggLayer integration: %w", err)
		}
	}

	// Start node-specific services
	if err := n.startNodeSpecificServices(); err != nil {
		return fmt.Errorf("failed to start node-specific services: %w", err)
	}

	n.logger.Printf("%s node started successfully", n.nodeType)
	return nil
}

// Stop gracefully stops the node and all its components
func (n *Node) Stop() error {
	if !n.running {
		return nil
	}

	n.logger.Printf("Stopping %s node...", n.nodeType)

	// Cancel context to signal shutdown
	n.cancel()

	// Stop components in reverse order
	if err := n.stopNodeSpecificServices(); err != nil {
		n.logger.Printf("Error stopping node-specific services: %v", err)
	}

	if n.aggLayerClient != nil {
		if err := n.stopAggLayerIntegration(); err != nil {
			n.logger.Printf("Error stopping AggLayer integration: %v", err)
		}
	}

	if err := n.stopConsensusEngine(); err != nil {
		n.logger.Printf("Error stopping consensus engine: %v", err)
	}

	if err := n.stopNetworkManager(); err != nil {
		n.logger.Printf("Error stopping network manager: %v", err)
	}

	if err := n.stopStateManager(); err != nil {
		n.logger.Printf("Error stopping state manager: %v", err)
	}

	// Wait for all goroutines to finish
	n.wg.Wait()

	n.running = false
	n.logger.Printf("%s node stopped", n.nodeType)
	return nil
}

// IsRunning returns whether the node is currently running
func (n *Node) IsRunning() bool {
	return n.running
}

// GetNodeType returns the type of this node
func (n *Node) GetNodeType() NodeType {
	return n.nodeType
}

// GetStateManager returns the state manager
func (n *Node) GetStateManager() *state.Manager {
	return n.stateManager
}

// GetNetworkManager returns the network manager
func (n *Node) GetNetworkManager() *networking.Manager {
	return n.networkManager
}

// GetConsensusEngine returns the consensus engine
func (n *Node) GetConsensusEngine() *consensus.Engine {
	return n.consensusEngine
}

// GetAggLayerClient returns the AggLayer client
func (n *Node) GetAggLayerClient() *agglayer.Client {
	return n.aggLayerClient
}

// Private methods for starting/stopping components

func (n *Node) startStateManager() error {
	n.logger.Printf("Starting state manager...")
	return n.stateManager.Start(n.ctx)
}

func (n *Node) stopStateManager() error {
	n.logger.Printf("Stopping state manager...")
	return n.stateManager.Stop()
}

func (n *Node) startNetworkManager() error {
	n.logger.Printf("Starting network manager...")
	return n.networkManager.Start(n.ctx)
}

func (n *Node) stopNetworkManager() error {
	n.logger.Printf("Stopping network manager...")
	return n.networkManager.Stop()
}

func (n *Node) startConsensusEngine() error {
	n.logger.Printf("Starting consensus engine...")
	return n.consensusEngine.Start(n.ctx)
}

func (n *Node) stopConsensusEngine() error {
	n.logger.Printf("Stopping consensus engine...")
	return n.consensusEngine.Stop()
}

func (n *Node) startAggLayerIntegration() error {
	n.logger.Printf("Starting AggLayer integration...")

	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		if err := n.aggLayerClient.Run(n.ctx); err != nil {
			n.logger.Printf("AggLayer client error: %v", err)
		}
	}()

	return nil
}

func (n *Node) stopAggLayerIntegration() error {
	n.logger.Printf("Stopping AggLayer integration...")
	return n.aggLayerClient.Stop()
}

func (n *Node) startNodeSpecificServices() error {
	switch n.nodeType {
	case ValidatorNode:
		return n.startValidatorServices()
	case SequencerNode:
		return n.startSequencerServices()
	case ArchiveNode:
		return n.startArchiveServices()
	default:
		return fmt.Errorf("unknown node type: %s", n.nodeType)
	}
}

func (n *Node) stopNodeSpecificServices() error {
	switch n.nodeType {
	case ValidatorNode:
		return n.stopValidatorServices()
	case SequencerNode:
		return n.stopSequencerServices()
	case ArchiveNode:
		return n.stopArchiveServices()
	default:
		return nil
	}
}

func (n *Node) startValidatorServices() error {
	n.logger.Printf("Starting validator-specific services...")
	// TODO: Implement validator-specific services
	// - Stake management
	// - Block validation
	// - Reward distribution
	return nil
}

func (n *Node) stopValidatorServices() error {
	n.logger.Printf("Stopping validator-specific services...")
	// TODO: Implement validator-specific cleanup
	return nil
}

func (n *Node) startSequencerServices() error {
	n.logger.Printf("Starting sequencer-specific services...")
	// TODO: Implement sequencer-specific services
	// - Transaction ordering
	// - Batch creation
	// - L1 submission
	return nil
}

func (n *Node) stopSequencerServices() error {
	n.logger.Printf("Stopping sequencer-specific services...")
	// TODO: Implement sequencer-specific cleanup
	return nil
}

func (n *Node) startArchiveServices() error {
	n.logger.Printf("Starting archive-specific services...")
	// TODO: Implement archive-specific services
	// - Historical data serving
	// - Data availability
	// - Backup management
	return nil
}

func (n *Node) stopArchiveServices() error {
	n.logger.Printf("Stopping archive-specific services...")
	// TODO: Implement archive-specific cleanup
	return nil
}

func isValidNodeType(nodeType NodeType) bool {
	switch nodeType {
	case ValidatorNode, SequencerNode, ArchiveNode:
		return true
	default:
		return false
	}
}
