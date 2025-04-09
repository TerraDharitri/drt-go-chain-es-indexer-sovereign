package runType

import (
	"net/http"

	"github.com/elastic/go-elasticsearch/v7"

	"github.com/TerraDharitri/drt-go-chain-es-indexer/client"
	"github.com/TerraDharitri/drt-go-chain-es-indexer/client/disabled"
	"github.com/TerraDharitri/drt-go-chain-es-indexer/client/logging"
	"github.com/TerraDharitri/drt-go-chain-es-indexer/process/elasticproc"
	"github.com/TerraDharitri/drt-go-chain-es-indexer/process/elasticproc/factory"
	"github.com/TerraDharitri/drt-go-chain-es-indexer/process/elasticproc/tokens"
	"github.com/TerraDharitri/drt-go-chain-es-indexer/process/elasticproc/transactions"
)

type sovereignRunTypeComponentsFactory struct {
	mainChainElastic factory.ElasticConfig
	dcdtPrefix       string
}

// NewSovereignRunTypeComponentsFactory will return a new instance of sovereign run type components factory
func NewSovereignRunTypeComponentsFactory(mainChainElastic factory.ElasticConfig, dcdtPrefix string) *sovereignRunTypeComponentsFactory {
	return &sovereignRunTypeComponentsFactory{
		mainChainElastic: mainChainElastic,
		dcdtPrefix:       dcdtPrefix,
	}
}

// Create will create the run type components
func (srtcf *sovereignRunTypeComponentsFactory) Create() (*runTypeComponents, error) {
	mainChainElasticClient, err := createMainChainElasticClient(srtcf.mainChainElastic)
	if err != nil {
		return nil, err
	}

	sovIndexTokensHandler, err := tokens.NewSovereignIndexTokensHandler(mainChainElasticClient, srtcf.dcdtPrefix)
	if err != nil {
		return nil, err
	}

	return &runTypeComponents{
		txHashExtractor:    transactions.NewSovereignTxHashExtractor(),
		rewardTxData:       transactions.NewSovereignRewardTxData(),
		indexTokensHandler: sovIndexTokensHandler,
	}, nil
}

func createMainChainElasticClient(mainChainElastic factory.ElasticConfig) (elasticproc.MainChainDatabaseClientHandler, error) {
	if mainChainElastic.Enabled {
		argsEsClient := elasticsearch.Config{
			Addresses:     []string{mainChainElastic.Url},
			Username:      mainChainElastic.UserName,
			Password:      mainChainElastic.Password,
			Logger:        &logging.CustomLogger{},
			RetryOnStatus: []int{http.StatusConflict},
			RetryBackoff:  client.RetryBackOff,
		}
		esClient, err := client.NewElasticClient(argsEsClient)
		if err != nil {
			return nil, err
		}

		return client.NewMainChainElasticClient(esClient, mainChainElastic.Enabled)
	} else {
		return disabled.NewDisabledElasticClient(), nil
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (srtcf *sovereignRunTypeComponentsFactory) IsInterfaceNil() bool {
	return srtcf == nil
}
