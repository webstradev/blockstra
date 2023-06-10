package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMustGeneratePrivateKey(t *testing.T) {
	privKey := MustGeneratePrivateKey()

	assert.Equal(t, len(privKey.Bytes()), privKeyLen)

	pubKey := privKey.Public()

	assert.Equal(t, len(pubKey.Bytes()), pubKeyLen)
}

func TestMustCreatePrivateKeyFromString(t *testing.T) {
	var (
		seedStr    = "e22f87e4add94968d0dc7dd9be75968ef4b2cb5686ae6641fddecfb6db8cb893"
		privKey    = MustCreatePrivateKeyFromString(seedStr)
		addressStr = "744c5a4919736642ff4a7b3529a174244428560e"
	)

	assert.Equal(t, privKeyLen, len(privKey.Bytes()))

	address := privKey.Public().Address()

	assert.Equal(t, addressStr, address.String())
}

func TestPrivateKeySign(t *testing.T) {
	privKey := MustGeneratePrivateKey()
	pubKey := privKey.Public()

	msg := []byte("foo bar baz")

	sig := privKey.Sign(msg)

	// Test with valid message and correct pub key
	assert.True(t, sig.Verify(pubKey, msg))

	// Test with invalid message and correct pub key
	assert.False(t, sig.Verify(pubKey, []byte("foo")))

	// Test with valid message and another pub key
	otherPrivKey := MustGeneratePrivateKey()
	otherPubKey := otherPrivKey.Public()

	assert.False(t, sig.Verify(otherPubKey, msg))
}

func TestPublicKeyToAddress(t *testing.T) {
	privKey := MustGeneratePrivateKey()
	pubkey := privKey.Public()

	address := pubkey.Address()

	assert.Equal(t, addressLen, len(address.Bytes()))

}
