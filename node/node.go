package node

import (
	"context"
	"fmt"

	"github.com/webstradev/blockstra/proto"
	"google.golang.org/grpc/peer"
)

type Node struct {
	version string

	proto.UnimplementedNodeServer
}

func New(version string) *Node {
	return &Node{
		version: version,
	}
}

func (n *Node) HandleTransaction(ctx context.Context, tx *proto.Transaction) (*proto.Ack, error) {
	peer, _ := peer.FromContext(ctx)

	fmt.Println("received tx from: ", peer.Addr)
	return &proto.Ack{}, nil
}

func (n *Node) Handshake(ctx context.Context, v *proto.Version) (*proto.Version, error) {
	ownVersion := &proto.Version{
		Version: n.version,
		Height:  100,
	}

	peer, _ := peer.FromContext(ctx)

	// Accept or not logic
	fmt.Printf("received version from %s: %+v\n", peer.Addr, v)

	return ownVersion, nil
}
