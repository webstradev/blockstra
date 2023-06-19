package node

import (
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/webstradev/blockstra/proto"
	"github.com/webstradev/blockstra/types"
)

type HeaderList struct {
	lock    sync.RWMutex
	headers []*proto.Header
}

func NewHeaderlist() *HeaderList {
	return &HeaderList{
		headers: []*proto.Header{},
	}
}

func (l *HeaderList) Add(header *proto.Header) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.headers = append(l.headers, header)
}

func (l *HeaderList) Get(index int) *proto.Header {
	if index > l.Height() {
		panic("index to high")
	}

	return l.headers[index]
}

func (l *HeaderList) Height() int {
	return l.Len() - 1
}

func (l *HeaderList) Len() int {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return len(l.headers)
}

type Chain struct {
	blockStore BlockStorer
	headers    *HeaderList
}

func NewChain(bs BlockStorer) *Chain {
	return &Chain{
		blockStore: bs,
		headers:    NewHeaderlist(),
	}
}

func (c *Chain) Height() int {
	return c.headers.Height()
}

func (c *Chain) AddBlock(block *proto.Block) error {
	// add the header to the list of headers
	c.headers.Add(block.Header)

	// validation
	return c.blockStore.Put(block)
}

func (c *Chain) GetBlockByHash(hash []byte) (*proto.Block, error) {
	hashHex := hex.EncodeToString(hash)
	return c.blockStore.Get(hashHex)
}

func (c *Chain) GetBlockByHeight(height int) (*proto.Block, error) {
	if c.Height() < height {
		return nil, fmt.Errorf("provided height [%d] higher than chain height", height)
	}

	header := c.headers.Get(height)

	hash := types.MustHashHeader(header)
	return c.GetBlockByHash(hash)
}
