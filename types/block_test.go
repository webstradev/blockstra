package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/webstradev/blockstra/crypto"
	"github.com/webstradev/blockstra/util"
)

func TestMustSignBlock(t *testing.T) {
	var (
		block   = util.RandomBlock()
		privKey = crypto.MustGeneratePrivateKey()
		pubKey  = privKey.Public()
	)

	sig := MustSignBlock(privKey, block)

	assert.Equal(t, 64, len(sig.Bytes()))

	assert.True(t, sig.Verify(pubKey, MustHashBlock(block)))

}

func TestMustHashBlock(t *testing.T) {
	block := util.RandomBlock()

	hash := MustHashBlock(block)

	assert.Equal(t, 32, len(hash))
}
