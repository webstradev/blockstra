package types

import (
	"crypto/sha256"

	pb "github.com/golang/protobuf/proto"
	"github.com/webstradev/blockstra/crypto"
	"github.com/webstradev/blockstra/proto"
)

// SignBlock
func SignBlock(pk *crypto.PrivateKey, b *proto.Block) *crypto.Signature {
	return pk.Sign(HashBlock(b))
}

// HasBlock returns a SHA256 of the header
func HashBlock(block *proto.Block) []byte {
	b, err := pb.Marshal(block)
	if err != nil {
		panic(err)
	}

	hash := sha256.Sum256(b)

	return hash[:]
}
