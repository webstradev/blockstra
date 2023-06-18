package types

import (
	"crypto/sha256"

	pb "github.com/golang/protobuf/proto"
	"github.com/webstradev/blockstra/crypto"
	"github.com/webstradev/blockstra/proto"
)

// SignTransactions hashes and then signs a transaction or panics if fialing to hash
func MustSignTransaction(pk *crypto.PrivateKey, b *proto.Transaction) *crypto.Signature {
	return pk.Sign(MustHashTransaction(b))
}

func MustHashTransaction(tx *proto.Transaction) []byte {
	b, err := pb.Marshal(tx)
	if err != nil {
		panic(err)
	}

	hash := sha256.Sum256(b)

	return hash[:]
}

func VerifyTransaction(tx *proto.Transaction) bool {
	for _, input := range tx.Inputs {
		var (
			sig    = crypto.SignatureFromBytes(input.Signature)
			pubKey = crypto.PublicKeyFromBytes(input.PublicKey)
		)

		// @FIX: Make sure we don't run into problems after verification
		// due to setting the signature to nil
		input.Signature = nil

		if !sig.Verify(pubKey, MustHashTransaction(tx)) {
			return false
		}
	}
	return true
}
