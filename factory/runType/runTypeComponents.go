package runType

import (
	"github.com/TerraDharitri/drt-go-chain-es-indexer/process/elasticproc"
	"github.com/TerraDharitri/drt-go-chain-es-indexer/process/elasticproc/transactions"
)

type runTypeComponents struct {
	txHashExtractor    transactions.TxHashExtractor
	rewardTxData       transactions.RewardTxDataHandler
	indexTokensHandler elasticproc.IndexTokensHandler
}

// Close does nothing
func (rtc *runTypeComponents) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (rtc *runTypeComponents) IsInterfaceNil() bool {
	return rtc == nil
}
