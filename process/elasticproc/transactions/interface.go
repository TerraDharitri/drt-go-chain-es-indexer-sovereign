package transactions

import (
	coreData "github.com/TerraDharitri/drt-go-chain-core/data"
	"github.com/TerraDharitri/drt-go-chain-core/data/outport"
	datafield "github.com/TerraDharitri/drt-go-chain-vm-common/parsers/dataField"
)

// DataFieldParser defines what a data field parser should be able to do
type DataFieldParser interface {
	Parse(dataField []byte, sender, receiver []byte, numOfShards uint32) *datafield.ResponseParseData
}

type feeInfoHandler interface {
	GetFeeInfo() *outport.FeeInfo
}

// TxHashExtractor defines what tx hash extractor should be able to do
type TxHashExtractor interface {
	ExtractExecutedTxHashes(mbIndex int, mbTxHashes [][]byte, header coreData.HeaderHandler) [][]byte
	IsInterfaceNil() bool
}

// RewardTxDataHandler defines what rewards tx handler should be able to do
type RewardTxDataHandler interface {
	GetSender() string
	IsInterfaceNil() bool
}
