package node

import (
	"context"
	"net"
	"sync"

	"github.com/webstradev/blockstra/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/peer"
)

type Node struct {
	version    string
	listenAddr string
	logger     *zap.SugaredLogger

	bootstrapNodes []string

	peerLock sync.RWMutex
	peers    map[proto.NodeClient]*proto.Version

	proto.UnimplementedNodeServer
}

func New(version, listenAddr string, logger *zap.SugaredLogger, bootstrapNodes []string) *Node {
	return &Node{
		version:    version,
		listenAddr: listenAddr,
		logger:     logger.With("source", listenAddr),

		peers:          map[proto.NodeClient]*proto.Version{},
		bootstrapNodes: bootstrapNodes,
	}
}

func (n *Node) Start() error {

	var (
		opts       = []grpc.ServerOption{}
		grpcServer = grpc.NewServer(opts...)
	)

	ln, err := net.Listen("tcp", n.listenAddr)
	if err != nil {
		n.logger.Fatal(err)
	}

	proto.RegisterNodeServer(grpcServer, n)

	n.logger.Infow("Node Started", "port", n.listenAddr)

	// Bootstrap the network with a list of already known nodes
	if len(n.bootstrapNodes) > 0 {
		go n.bootstrapNetwork(n.bootstrapNodes)
	}

	return grpcServer.Serve(ln)
}

func (n *Node) HandleTransaction(ctx context.Context, tx *proto.Transaction) (*proto.Ack, error) {
	peer, _ := peer.FromContext(ctx)

	n.logger.Debug("received tx from: ", peer.Addr)
	return &proto.Ack{}, nil
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
		if addr == n.listenAddr {
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
		Version:    n.version,
		Height:     0,
		ListenAddr: n.listenAddr,
		PeerList:   n.getPeerList(),
	}
}

func (n *Node) canConnectWith(addr string) bool {
	// Don't attempt to connect with itself
	if n.listenAddr == addr {
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
