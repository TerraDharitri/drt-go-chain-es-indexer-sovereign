//go:build integrationtests

package integrationtests

import (
	"testing"

	"github.com/TerraDharitri/drt-go-chain-es-indexer/process/dataindexer"
	"github.com/stretchr/testify/require"
)

func TestMappingsOfDCDTsIndex(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	_, err = CreateElasticProcessor(esClient)
	require.Nil(t, err)

	mappings, err := getIndexMappings(dataindexer.DCDTsIndex)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/mappings/dcdts.json"), mappings)
}
