package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/webstradev/blockstra/crypto"
	"github.com/webstradev/blockstra/util"
)

func TestSignBlock(t *testing.T) {
	var (
		block   = util.RandomBlock()
		privKey = crypto.MustGeneratePrivateKey()
		pubKey  = privKey.Public()
	)

	sig := SignBlock(privKey, block)

	assert.Equal(t, 64, len(sig.Bytes()))

	assert.True(t, sig.Verify(pubKey, HashBlock(block)))

}

func TestHashBlock(t *testing.T) {
	block := util.RandomBlock()

	hash := HashBlock(block)

	assert.Equal(t, 32, len(hash))
}
