package node

import (
	"context"
	"encoding/hex"
	"net"
	"sync"
	"time"

	"github.com/webstradev/blockstra/crypto"

	"github.com/webstradev/blockstra/proto"
	"github.com/webstradev/blockstra/types"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/peer"
)

const blockTime = time.Second * 5

// this is probably going to be a BSTin future
type MemPool struct {
	lock sync.RWMutex
	txx  map[string]*proto.Transaction
}

func NewMemPool() *MemPool {
	return &MemPool{
		txx: map[string]*proto.Transaction{},
	}
}

func (p *MemPool) Has(tx *proto.Transaction) bool {
	hash := hex.EncodeToString(types.MustHashTransaction(tx))
	_, ok := p.txx[hash]
	return ok
}

func (p *MemPool) Add(tx *proto.Transaction) bool {
	if p.Has(tx) {
		return false
	}
	p.lock.Lock()
	defer p.lock.Unlock()

	hash := hex.EncodeToString(types.MustHashTransaction(tx))
	p.txx[hash] = tx
	return true
}

func (p *MemPool) Clear() {
	p.lock.Lock()
	defer p.lock.Unlock()

	for hash := range p.txx {
		delete(p.txx, hash)
	}
}

type ServerConfig struct {
	Version    string
	ListenAddr string
	PrivateKey *crypto.PrivateKey
}

type Node struct {
	ServerConfig
	logger *zap.SugaredLogger

	bootstrapNodes []string

	peerLock sync.RWMutex
	peers    map[proto.NodeClient]*proto.Version

	memPool *MemPool

	proto.UnimplementedNodeServer
}

func New(cfg ServerConfig, logger *zap.SugaredLogger, bootstrapNodes []string) *Node {
	return &Node{
		ServerConfig: cfg,
		logger:       logger.With("source", cfg.ListenAddr),

		bootstrapNodes: bootstrapNodes,

		peers: map[proto.NodeClient]*proto.Version{},

		memPool: NewMemPool(),
	}
}

func (n *Node) Start() error {

	var (
		opts       = []grpc.ServerOption{}
		grpcServer = grpc.NewServer(opts...)
	)

	ln, err := net.Listen("tcp", n.ListenAddr)
	if err != nil {
		n.logger.Fatal(err)
	}

	proto.RegisterNodeServer(grpcServer, n)

	n.logger.Infow("Node Started", "port", n.ListenAddr)

	// Bootstrap the network with a list of already known nodes
	if len(n.bootstrapNodes) > 0 {
		go n.bootstrapNetwork(n.bootstrapNodes)
	}

	if n.PrivateKey != nil {
		go n.validatorLoop()
	}

	return grpcServer.Serve(ln)
}

func (n *Node) HandleTransaction(ctx context.Context, tx *proto.Transaction) (*proto.Ack, error) {
	peer, _ := peer.FromContext(ctx)
	hash := hex.EncodeToString(types.MustHashTransaction(tx))

	if n.memPool.Add(tx) {
		n.logger.Debugw("received tx from: ", "from", peer.Addr, "hash", hash)
		go func() {
			if err := n.broadcast(tx); err != nil {
				n.logger.Errorw("broadcast error", "err", err)
			}
		}()
	}

	return &proto.Ack{}, nil
}

func (n *Node) validatorLoop() {
	n.logger.Infow("starting validator loop", "pubkey", n.PrivateKey.Public(), "blockTime", blockTime)
	ticker := time.NewTicker(blockTime)
	for {
		<-ticker.C

		n.logger.Debugw("time to create a new block", "lenTx", len(n.memPool.txx))
		n.memPool.Clear()
	}
}

func (n *Node) broadcast(msg any) error {
	for peer := range n.peers {
		switch v := msg.(type) {
		case *proto.Transaction:
			_, err := peer.HandleTransaction(context.Background(), v)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (n *Node) Handshake(ctx context.Context, v *proto.Version) (*proto.Version, error) {
	c, err := makeNodeClient(v.ListenAddr)
	if err != nil {
		return nil, err
	}

	n.addPeer(c, v)

	return n.getVersion(), nil
}

func (n *Node) addPeer(client proto.NodeClient, version *proto.Version) {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()

	// handle the logic where we decide to accept or drop
	// the incoming node connection

	n.peers[client] = version
	n.logger.Debugw("new peer successfully connected", "addr", version.ListenAddr)

	// Connect to all peers in the received list of peers
	if len(version.PeerList) > 0 {
		go n.bootstrapNetwork(version.PeerList)
	}
}

func (n *Node) removePeer(c proto.NodeClient) {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()
	delete(n.peers, c)
}

func (n *Node) bootstrapNetwork(addrs []string) error {
	for _, addr := range addrs {
		if addr == n.ListenAddr {
			continue
		}

		if !n.canConnectWith(addr) {
			continue
		}

		n.logger.Debugw("dialing remote nodes", "remote", addr)
		client, version, err := n.dialRemoteNode(addr)
		if err != nil {
			n.logger.Error("dial error: ", err)
			continue
		}

		n.addPeer(client, version)
	}

	return nil
}

func (n *Node) dialRemoteNode(addr string) (proto.NodeClient, *proto.Version, error) {
	c, err := makeNodeClient(addr)
	if err != nil {
		return nil, nil, err
	}

	v, err := c.Handshake(context.Background(), n.getVersion())
	if err != nil {
		return nil, nil, err
	}

	return c, v, nil
}

func (n *Node) getVersion() *proto.Version {
	return &proto.Version{
		Version:    n.Version,
		Height:     0,
		ListenAddr: n.ListenAddr,
		PeerList:   n.getPeerList(),
	}
}

func (n *Node) canConnectWith(addr string) bool {
	// Don't attempt to connect with itself
	if n.ListenAddr == addr {
		return false
	}

	// Don't attempt to connect with already connected peers
	connectedPeers := n.getPeerList()
	for _, connectedAddr := range connectedPeers {
		if addr == connectedAddr {
			return false
		}
	}

	return true
}

func (n *Node) getPeerList() []string {
	n.peerLock.RLock()
	defer n.peerLock.RUnlock()

	peers := []string{}

	for _, version := range n.peers {
		peers = append(peers, version.ListenAddr)
	}

	return peers
}

func makeNodeClient(listenAddr string) (proto.NodeClient, error) {
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	c, err := grpc.Dial(listenAddr, opts...)
	if err != nil {
		return nil, err
	}

	return proto.NewNodeClient(c), nil
}
