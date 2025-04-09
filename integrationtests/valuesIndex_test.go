//go:build integrationtests

package integrationtests

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/TerraDharitri/drt-go-chain-es-indexer/mock"
	indexerData "github.com/TerraDharitri/drt-go-chain-es-indexer/process/dataindexer"
	"github.com/TerraDharitri/drt-go-chain-es-indexer/process/elasticproc"
	"github.com/TerraDharitri/drt-go-chain-es-indexer/process/elasticproc/factory"
)

func TestCheckVersionIsIndexer(t *testing.T) {
	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	version := "v1.4.5"
	args := factory.ArgElasticProcessorFactory{
		Marshalizer:              &mock.MarshalizerMock{},
		Hasher:                   &mock.HasherMock{},
		AddressPubkeyConverter:   pubKeyConverter,
		ValidatorPubkeyConverter: mock.NewPubkeyConverterMock(32),
		DBClient:                 esClient,
		Denomination:             18,
		Version:                  version,
		EnabledIndexes:           []string{indexerData.ValuesIndex},
		TxHashExtractor:          &mock.TxHashExtractorMock{},
		RewardTxData:             &mock.RewardTxDataMock{},
		IndexTokensHandler:       &elasticproc.IndexTokenHandlerMock{},
	}

	_, err = factory.CreateElasticProcessor(args)
	require.Nil(t, err)

	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), []string{"indexer-version"}, indexerData.ValuesIndex, true, genericResponse)
	require.Nil(t, err)
	require.Equal(t, fmt.Sprintf(`{"key":"indexer-version","value":"%s"}`, version), string(genericResponse.Docs[0].Source))
}
