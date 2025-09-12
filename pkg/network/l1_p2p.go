package network

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// L1P2PNetwork implements P2P networking for the L1 blockchain
// Innovation: Hybrid topology with validator-priority routing
type L1P2PNetwork struct {
	// Network configuration
	nodeID       common.Address
	listenAddr   string
	externalAddr string
	maxPeers     int
	
	// Network state
	peers        map[common.Address]*Peer
	validators   map[common.Address]bool
	listener     net.Listener
	
	// Message routing
	msgHandler   MessageHandler
	msgBus       chan *NetworkMessage
	
	// Synchronization
	mu      sync.RWMutex
	ctx     context.Context
	cancel  context.CancelFunc
	running bool
	
	// Innovation: Priority routing for validators
	validatorPeers map[common.Address]*Peer
	
	// Discovery and bootstrapping
	bootstrapNodes []string
	peerDiscovery  *PeerDiscovery
	
	// Metrics
	sentMessages     uint64
	receivedMessages uint64
	activeConnections int
}

// Peer represents a network peer
type Peer struct {
	ID          common.Address
	Address     string
	Connection  net.Conn
	IsValidator bool
	LastSeen    time.Time
	Latency     time.Duration
	Version     string
	
	// Communication channels
	sendCh   chan []byte
	stopCh   chan struct{}
	
	// Metrics
	messagesSent     uint64
	messagesReceived uint64
	bytesTransferred uint64
}

// NetworkMessage represents a P2P message
type NetworkMessage struct {
	Type      MessageType        `json:"type"`
	From      common.Address     `json:"from"`
	To        common.Address     `json:"to,omitempty"` // Empty for broadcast
	Data      json.RawMessage    `json:"data"`
	Timestamp time.Time          `json:"timestamp"`
	Signature []byte             `json:"signature"`
}

// MessageType represents different message types
type MessageType string

const (
	MsgTypeProposal     MessageType = "proposal"
	MsgTypeVote         MessageType = "vote"
	MsgTypeCommit       MessageType = "commit"
	MsgTypeTransaction  MessageType = "transaction"
	MsgTypeBlockRequest MessageType = "block_request"
	MsgTypeBlockData    MessageType = "block_data"
	MsgTypePeerInfo     MessageType = "peer_info"
	MsgTypeHandshake    MessageType = "handshake"
	MsgTypePing         MessageType = "ping"
	MsgTypePong         MessageType = "pong"
)

// MessageHandler interface for handling network messages
type MessageHandler interface {
	HandleMessage(msg *NetworkMessage) error
}

// PeerDiscovery handles peer discovery
type PeerDiscovery struct {
	bootstrapNodes []string
	knownPeers    map[string]time.Time
	mu            sync.RWMutex
}

// HandshakeMessage represents the initial handshake
type HandshakeMessage struct {
	NodeID      common.Address `json:"node_id"`
	Version     string         `json:"version"`
	ChainID     *big.Int       `json:"chain_id"`
	IsValidator bool           `json:"is_validator"`
	ListenPort  int            `json:"listen_port"`
}

// NewL1P2PNetwork creates a new L1 P2P network
func NewL1P2PNetwork(nodeID common.Address, listenAddr string, maxPeers int, bootstrapNodes []string) *L1P2PNetwork {
	return &L1P2PNetwork{
		nodeID:         nodeID,
		listenAddr:     listenAddr,
		maxPeers:       maxPeers,
		peers:          make(map[common.Address]*Peer),
		validators:     make(map[common.Address]bool),
		validatorPeers: make(map[common.Address]*Peer),
		msgBus:         make(chan *NetworkMessage, 1000),
		bootstrapNodes: bootstrapNodes,
		peerDiscovery:  NewPeerDiscovery(bootstrapNodes),
	}
}

// Start begins the P2P network
func (n *L1P2PNetwork) Start(ctx context.Context) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	
	if n.running {
		return fmt.Errorf("network already running")
	}
	
	n.ctx, n.cancel = context.WithCancel(ctx)
	n.running = true
	
	// Start listening for connections
	listener, err := net.Listen("tcp", n.listenAddr)
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}
	n.listener = listener
	
	// Start network components
	go n.acceptConnections()
	go n.messageProcessor()
	go n.peerMaintenance()
	go n.networkDiscovery()
	
	// Bootstrap connections
	go n.bootstrapConnections()
	
	fmt.Printf("🌐 L1 P2P Network started on %s\n", n.listenAddr)
	fmt.Printf("   • Node ID: %s\n", n.nodeID.Hex()[:8])
	fmt.Printf("   • Max Peers: %d\n", n.maxPeers)
	fmt.Printf("   • Bootstrap Nodes: %d\n", len(n.bootstrapNodes))
	
	return nil
}

// acceptConnections accepts incoming peer connections
func (n *L1P2PNetwork) acceptConnections() {
	for {
		select {
		case <-n.ctx.Done():
			return
		default:
			conn, err := n.listener.Accept()
			if err != nil {
				select {
				case <-n.ctx.Done():
					return
				default:
					fmt.Printf("❌ Accept error: %v\n", err)
					continue
				}
			}
			
			go n.handleIncomingConnection(conn)
		}
	}
}

// handleIncomingConnection handles a new incoming connection
func (n *L1P2PNetwork) handleIncomingConnection(conn net.Conn) {
	defer conn.Close()
	
	// Perform handshake
	peer, err := n.performHandshake(conn, false)
	if err != nil {
		fmt.Printf("❌ Handshake failed: %v\n", err)
		return
	}
	
	// Add peer
	if n.addPeer(peer) {
		fmt.Printf("🤝 New peer connected: %s (%s)\n", peer.ID.Hex()[:8], peer.Address)
		n.handlePeerConnection(peer)
	}
}

// bootstrapConnections establishes initial connections to bootstrap nodes
func (n *L1P2PNetwork) bootstrapConnections() {
	for _, bootstrapAddr := range n.bootstrapNodes {
		select {
		case <-n.ctx.Done():
			return
		default:
			if err := n.connectToPeer(bootstrapAddr); err != nil {
				fmt.Printf("❌ Failed to connect to bootstrap node %s: %v\n", bootstrapAddr, err)
			}
			time.Sleep(1 * time.Second) // Stagger connections
		}
	}
}

// connectToPeer establishes a connection to a peer
func (n *L1P2PNetwork) connectToPeer(address string) error {
	conn, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to dial %s: %w", address, err)
	}
	
	// Perform handshake
	peer, err := n.performHandshake(conn, true)
	if err != nil {
		conn.Close()
		return fmt.Errorf("handshake failed: %w", err)
	}
	
	// Add peer
	if n.addPeer(peer) {
		fmt.Printf("🔗 Connected to peer: %s (%s)\n", peer.ID.Hex()[:8], peer.Address)
		go n.handlePeerConnection(peer)
		return nil
	}
	
	conn.Close()
	return fmt.Errorf("failed to add peer")
}

// performHandshake performs the initial handshake with a peer
func (n *L1P2PNetwork) performHandshake(conn net.Conn, isOutgoing bool) (*Peer, error) {
	// Create handshake message
	handshake := &HandshakeMessage{
		NodeID:      n.nodeID,
		Version:     "lightchain-l1/v1.0.0",
		ChainID:     big.NewInt(1337), // L1 chain ID
		IsValidator: false, // Will be updated based on validator status
		ListenPort:  8000,  // Extract from listenAddr
	}
	
	// Send handshake
	handshakeData, _ := json.Marshal(handshake)
	msg := &NetworkMessage{
		Type:      MsgTypeHandshake,
		From:      n.nodeID,
		Data:      handshakeData,
		Timestamp: time.Now(),
	}
	
	msgData, _ := json.Marshal(msg)
	if _, err := conn.Write(msgData); err != nil {
		return nil, fmt.Errorf("failed to send handshake: %w", err)
	}
	
	// Receive handshake response
	buffer := make([]byte, 4096)
	bytesRead, err := conn.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read handshake response: %w", err)
	}
	
	var responseMsg NetworkMessage
	if err := json.Unmarshal(buffer[:bytesRead], &responseMsg); err != nil {
		return nil, fmt.Errorf("failed to parse handshake response: %w", err)
	}
	
	var peerHandshake HandshakeMessage
	if err := json.Unmarshal(responseMsg.Data, &peerHandshake); err != nil {
		return nil, fmt.Errorf("failed to parse peer handshake: %w", err)
	}
	
	// Create peer
	peer := &Peer{
		ID:          peerHandshake.NodeID,
		Address:     conn.RemoteAddr().String(),
		Connection:  conn,
		IsValidator: peerHandshake.IsValidator,
		LastSeen:    time.Now(),
		Version:     peerHandshake.Version,
		sendCh:      make(chan []byte, 100),
		stopCh:      make(chan struct{}),
	}
	
	return peer, nil
}

// addPeer adds a peer to the network
func (n *L1P2PNetwork) addPeer(peer *Peer) bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	
	// Check if already connected
	if _, exists := n.peers[peer.ID]; exists {
		return false
	}
	
	// Check peer limits
	if len(n.peers) >= n.maxPeers {
		return false
	}
	
	// Add peer
	n.peers[peer.ID] = peer
	n.activeConnections++
	
	// Add to validator peers if applicable
	if peer.IsValidator {
		n.validatorPeers[peer.ID] = peer
	}
	
	return true
}

// handlePeerConnection handles ongoing communication with a peer
func (n *L1P2PNetwork) handlePeerConnection(peer *Peer) {
	defer n.removePeer(peer.ID)
	
	// Start send goroutine
	go n.peerSender(peer)
	
	// Handle incoming messages
	buffer := make([]byte, 4096)
	for {
		select {
		case <-peer.stopCh:
			return
		case <-n.ctx.Done():
			return
		default:
			peer.Connection.SetReadDeadline(time.Now().Add(30 * time.Second))
			bytesRead, err := peer.Connection.Read(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				fmt.Printf("❌ Peer %s read error: %v\n", peer.ID.Hex()[:8], err)
				return
			}
			
			// Process message
			var msg NetworkMessage
			if err := json.Unmarshal(buffer[:bytesRead], &msg); err != nil {
				fmt.Printf("❌ Failed to unmarshal message from %s: %v\n", peer.ID.Hex()[:8], err)
				continue
			}
			
			peer.messagesReceived++
			peer.bytesTransferred += uint64(bytesRead)
			peer.LastSeen = time.Now()
			n.receivedMessages++
			
			// Route to message bus
			select {
			case n.msgBus <- &msg:
			default:
				fmt.Printf("⚠️  Message bus full, dropping message from %s\n", peer.ID.Hex()[:8])
			}
		}
	}
}

// peerSender handles sending messages to a peer
func (n *L1P2PNetwork) peerSender(peer *Peer) {
	for {
		select {
		case <-peer.stopCh:
			return
		case <-n.ctx.Done():
			return
		case data := <-peer.sendCh:
			peer.Connection.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if _, err := peer.Connection.Write(data); err != nil {
				fmt.Printf("❌ Failed to send to peer %s: %v\n", peer.ID.Hex()[:8], err)
				return
			}
			peer.messagesSent++
			peer.bytesTransferred += uint64(len(data))
			n.sentMessages++
		}
	}
}

// messageProcessor processes messages from the message bus
func (n *L1P2PNetwork) messageProcessor() {
	for {
		select {
		case <-n.ctx.Done():
			return
		case msg := <-n.msgBus:
			if n.msgHandler != nil {
				if err := n.msgHandler.HandleMessage(msg); err != nil {
					fmt.Printf("❌ Message handler error: %v\n", err)
				}
			}
		}
	}
}

// peerMaintenance performs periodic peer maintenance
func (n *L1P2PNetwork) peerMaintenance() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			n.maintainPeers()
		}
	}
}

// maintainPeers maintains peer connections
func (n *L1P2PNetwork) maintainPeers() {
	n.mu.Lock()
	defer n.mu.Unlock()
	
	now := time.Now()
	for peerID, peer := range n.peers {
		// Remove stale peers
		if now.Sub(peer.LastSeen) > 2*time.Minute {
			fmt.Printf("🔌 Removing stale peer: %s\n", peerID.Hex()[:8])
			delete(n.peers, peerID)
			delete(n.validatorPeers, peerID)
			close(peer.stopCh)
			peer.Connection.Close()
			n.activeConnections--
		}
	}
	
	// Try to maintain target peer count
	if len(n.peers) < n.maxPeers/2 {
		go n.discoverMorePeers()
	}
}

// networkDiscovery performs network discovery
func (n *L1P2PNetwork) networkDiscovery() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			n.requestPeerInfo()
		}
	}
}

// requestPeerInfo requests peer information from connected peers
func (n *L1P2PNetwork) requestPeerInfo() {
	msg := &NetworkMessage{
		Type:      MsgTypePeerInfo,
		From:      n.nodeID,
		Timestamp: time.Now(),
	}
	
	n.Broadcast(msg)
}

// discoverMorePeers attempts to discover and connect to more peers
func (n *L1P2PNetwork) discoverMorePeers() {
	// Implementation would use DHT or other discovery mechanisms
	fmt.Printf("🔍 Discovering more peers... Current: %d/%d\n", len(n.peers), n.maxPeers)
}

// removePeer removes a peer from the network
func (n *L1P2PNetwork) removePeer(peerID common.Address) {
	n.mu.Lock()
	defer n.mu.Unlock()
	
	if peer, exists := n.peers[peerID]; exists {
		delete(n.peers, peerID)
		delete(n.validatorPeers, peerID)
		close(peer.stopCh)
		peer.Connection.Close()
		n.activeConnections--
		fmt.Printf("🔌 Peer disconnected: %s\n", peerID.Hex()[:8])
	}
}

// Broadcast sends a message to all connected peers
func (n *L1P2PNetwork) Broadcast(msg *NetworkMessage) error {
	return n.broadcastToPeers(msg, n.getAllPeers())
}

// BroadcastToValidators sends a message only to validator peers (priority routing)
func (n *L1P2PNetwork) BroadcastToValidators(msg *NetworkMessage) error {
	n.mu.RLock()
	validatorPeers := make([]*Peer, 0, len(n.validatorPeers))
	for _, peer := range n.validatorPeers {
		validatorPeers = append(validatorPeers, peer)
	}
	n.mu.RUnlock()
	
	return n.broadcastToPeers(msg, validatorPeers)
}

// SendToPeer sends a message to a specific peer
func (n *L1P2PNetwork) SendToPeer(peerID common.Address, msg *NetworkMessage) error {
	n.mu.RLock()
	peer, exists := n.peers[peerID]
	n.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("peer not found: %s", peerID.Hex())
	}
	
	msg.To = peerID
	msgData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	
	select {
	case peer.sendCh <- msgData:
		return nil
	default:
		return fmt.Errorf("peer send channel full")
	}
}

// broadcastToPeers broadcasts to a specific set of peers
func (n *L1P2PNetwork) broadcastToPeers(msg *NetworkMessage, peers []*Peer) error {
	if len(peers) == 0 {
		return fmt.Errorf("no peers to broadcast to")
	}
	
	msgData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	
	successCount := 0
	for _, peer := range peers {
		select {
		case peer.sendCh <- msgData:
			successCount++
		default:
			fmt.Printf("⚠️  Failed to send to peer %s: channel full\n", peer.ID.Hex()[:8])
		}
	}
	
	fmt.Printf("📤 Broadcasted %s to %d/%d peers\n", msg.Type, successCount, len(peers))
	return nil
}

// getAllPeers returns all connected peers
func (n *L1P2PNetwork) getAllPeers() []*Peer {
	n.mu.RLock()
	defer n.mu.RUnlock()
	
	peers := make([]*Peer, 0, len(n.peers))
	for _, peer := range n.peers {
		peers = append(peers, peer)
	}
	return peers
}

// SetMessageHandler sets the message handler
func (n *L1P2PNetwork) SetMessageHandler(handler MessageHandler) {
	n.msgHandler = handler
}

// Stop shuts down the P2P network
func (n *L1P2PNetwork) Stop() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	
	if !n.running {
		return nil
	}
	
	if n.cancel != nil {
		n.cancel()
	}
	
	if n.listener != nil {
		n.listener.Close()
	}
	
	// Close all peer connections
	for _, peer := range n.peers {
		close(peer.stopCh)
		peer.Connection.Close()
	}
	
	n.running = false
	fmt.Println("🌐 L1 P2P Network stopped")
	return nil
}

// GetNetworkStatus returns current network status
func (n *L1P2PNetwork) GetNetworkStatus() map[string]interface{} {
	n.mu.RLock()
	defer n.mu.RUnlock()
	
	return map[string]interface{}{
		"nodeId":            n.nodeID.Hex(),
		"activeConnections": n.activeConnections,
		"totalPeers":        len(n.peers),
		"validatorPeers":    len(n.validatorPeers),
		"sentMessages":      n.sentMessages,
		"receivedMessages":  n.receivedMessages,
		"maxPeers":         n.maxPeers,
		"running":          n.running,
	}
}

// NewPeerDiscovery creates a new peer discovery instance
func NewPeerDiscovery(bootstrapNodes []string) *PeerDiscovery {
	return &PeerDiscovery{
		bootstrapNodes: bootstrapNodes,
		knownPeers:    make(map[string]time.Time),
	}
}