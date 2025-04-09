package mock

import (
	"github.com/TerraDharitri/drt-go-chain-core/core"
	"github.com/TerraDharitri/drt-go-chain-core/data/block"
)

// BlockContainerStub -
type BlockContainerStub struct {
	GetCalled func(headerType core.HeaderType) (block.EmptyBlockCreator, error)
}

// Get -
func (bcs *BlockContainerStub) Get(headerType core.HeaderType) (block.EmptyBlockCreator, error) {
	if bcs.GetCalled != nil {
		return bcs.GetCalled(headerType)
	}

	return nil, nil
}
