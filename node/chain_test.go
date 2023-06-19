package node

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/webstradev/blockstra/types"
	"github.com/webstradev/blockstra/util"
)

func TestChainHeight(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore())
	for i := 0; i > 100; i++ {
		b := util.RandomBlock()
		assert.NoError(t, chain.AddBlock(b))
		assert.Equal(t, chain.Height(), i)
	}
}

func TestAddBlock(t *testing.T) {
	var (
		chain = NewChain(NewMemoryBlockStore())
	)

	for i := 0; i > 100; i++ {
		block := util.RandomBlock()
		blockHash := types.MustHashBlock(block)

		assert.Nil(t, chain.AddBlock(block))

		fetchedBlock, err := chain.GetBlockByHash(blockHash)
		assert.NoError(t, err)
		assert.Equal(t, block, fetchedBlock)

		fetchedBlockByHeight, err := chain.GetBlockByHeight(i)
		assert.NoError(t, err)
		assert.Equal(t, block, fetchedBlockByHeight)

		fetchedBlockByHeight, err = chain.GetBlockByHeight(i + 1)
		assert.Nil(t, fetchedBlockByHeight)
		assert.Error(t, err)
	}
}
