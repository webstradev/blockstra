package node

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/webstradev/blockstra/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/peer"
)

type Node struct {
	version    string
	listenAddr string

	peerLock sync.RWMutex
	peers    map[proto.NodeClient]*proto.Version

	proto.UnimplementedNodeServer
}

func New(version, listenAddr string) *Node {
	return &Node{
		version:    version,
		listenAddr: listenAddr,

		peers: map[proto.NodeClient]*proto.Version{},
	}
}

func (n *Node) addPeer(c proto.NodeClient, v *proto.Version) {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()

	fmt.Printf("[%s] new peer connected (%s) - height (%d)\n", n.listenAddr, v.ListenAddr, v.Height)
	n.peers[c] = v
}

func (n *Node) removePeer(c proto.NodeClient) {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()
	delete(n.peers, c)
}

func (n *Node) BootstrapNetwork(addrs []string) error {
	for _, addr := range addrs {
		c, err := makeNodeClient(addr)
		if err != nil {
			return err
		}

		v, err := c.Handshake(context.Background(), n.getVersion())
		if err != nil {
			fmt.Println("handshake error: ", err)
			continue
		}

		n.addPeer(c, v)
	}

	return nil
}

func (n *Node) Start() error {

	opts := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(opts...)

	ln, err := net.Listen("tcp", n.listenAddr)
	if err != nil {
		log.Fatal(err)
	}

	proto.RegisterNodeServer(grpcServer, n)

	return grpcServer.Serve(ln)
}

func (n *Node) HandleTransaction(ctx context.Context, tx *proto.Transaction) (*proto.Ack, error) {
	peer, _ := peer.FromContext(ctx)

	fmt.Println("received tx from: ", peer.Addr)
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

func (n *Node) getVersion() *proto.Version {
	return &proto.Version{
		Version:    n.version,
		Height:     0,
		ListenAddr: n.listenAddr,
	}
}

func makeNodeClient(listenAddr string) (proto.NodeClient, error) {
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	c, err := grpc.Dial(listenAddr, opts...)
	if err != nil {
		return nil, err
	}

	return proto.NewNodeClient(c), nil
}
