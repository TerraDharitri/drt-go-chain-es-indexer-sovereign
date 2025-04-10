package runType

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/TerraDharitri/drt-go-chain-es-indexer/process/elasticproc/factory"
)

func TestSovereignRunTypeComponentsFactory_CreateAndClose(t *testing.T) {
	t.Parallel()

	srtcf := NewSovereignRunTypeComponentsFactory(factory.ElasticConfig{}, "sov")
	require.False(t, srtcf.IsInterfaceNil())

	srtc, err := srtcf.Create()
	require.NotNil(t, srtc)
	require.NoError(t, err)

	require.NoError(t, srtc.Close())
}
