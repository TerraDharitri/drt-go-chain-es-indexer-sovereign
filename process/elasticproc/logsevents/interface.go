package logsevents

import (
	coreData "github.com/TerraDharitri/drt-go-chain-core/data"
	"github.com/TerraDharitri/drt-go-chain-core/data/outport"
	"github.com/TerraDharitri/drt-go-chain-es-indexer/data"
	"github.com/TerraDharitri/drt-go-chain-es-indexer/process/elasticproc/tokeninfo"
)

type argsProcessEvent struct {
	txHashHexEncoded        string
	scDeploys               map[string]*data.ScDeployInfo
	changeOwnerOperations   map[string]*data.OwnerData
	txs                     map[string]*data.Transaction
	scrs                    map[string]*data.ScResult
	event                   coreData.EventHandler
	tokens                  data.TokensHandler
	tokensSupply            data.TokensHandler
	tokenRolesAndProperties *tokeninfo.TokenRolesAndProperties
	txHashStatusInfoProc    txHashStatusInfoHandler
	timestamp               uint64
	logAddress              []byte
	selfShardID             uint32
	numOfShards             uint32
}

type argOutputProcessEvent struct {
	tokenInfo     *data.TokenInfo
	delegator     *data.Delegator
	updatePropNFT *data.NFTDataUpdate
	processed     bool
}

type eventsProcessor interface {
	processEvent(args *argsProcessEvent) argOutputProcessEvent
}

type txHashStatusInfoHandler interface {
	addRecord(hash string, statusInfo *outport.StatusInfo)
	getAllRecords() map[string]*outport.StatusInfo
}
