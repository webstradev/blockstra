package types

import (
	"crypto/sha256"

	pb "github.com/golang/protobuf/proto"
	"github.com/webstradev/blockstra/crypto"
	"github.com/webstradev/blockstra/proto"
)

// SignBlock hashes and then signs a blocks or panics if fialing to hash
func MustSignBlock(pk *crypto.PrivateKey, b *proto.Block) *crypto.Signature {
	return pk.Sign(MustHashBlock(b))
}

// MustHashBlock returns a SHA256 of the header or panics if encountering an error
func MustHashBlock(block *proto.Block) []byte {
	return MustHashHeader(block.Header)
}

// MustHashHeader returns a SHA256 of the header or panics if encountering an error
func MustHashHeader(header *proto.Header) []byte {
	b, err := pb.Marshal(header)
	if err != nil {
		panic(err)
	}

	hash := sha256.Sum256(b)

	return hash[:]
}
