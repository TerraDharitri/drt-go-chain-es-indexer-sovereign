package logsevents

import (
	"encoding/hex"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/TerraDharitri/drt-go-chain-core/core"
	"github.com/TerraDharitri/drt-go-chain-core/data/dcdt"
	"github.com/TerraDharitri/drt-go-chain-core/data/transaction"
	"github.com/TerraDharitri/drt-go-chain-es-indexer/data"
	"github.com/TerraDharitri/drt-go-chain-es-indexer/mock"
	"github.com/stretchr/testify/require"
)

func TestProcessNFTProperties_Update(t *testing.T) {
	t.Parallel()

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte("DCDTNFTUpdateAttributes"),
		Topics:     [][]byte{[]byte("TOUC-aaaa"), big.NewInt(1).Bytes(), nil, []byte("new-something")},
	}
	args := &argsProcessEvent{
		timestamp: 1234,
		event:     event,
	}

	nftsPropertiesP := newNFTsPropertiesProcessor(&mock.PubkeyConverterMock{}, &mock.MarshalizerMock{})

	res := nftsPropertiesP.processEvent(args)
	require.True(t, res.processed)
	require.Equal(t, &data.NFTDataUpdate{
		Identifier:    "TOUC-aaaa-01",
		NewAttributes: []byte("new-something"),
		Address:       "61646472",
	}, res.updatePropNFT)
}

func TestProcessNFTProperties_AddUris(t *testing.T) {
	t.Parallel()

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte("DCDTNFTAddURI"),
		Topics:     [][]byte{[]byte("TOUC-aaaa"), big.NewInt(1).Bytes(), nil, []byte("uri1"), []byte("uri2")},
	}
	args := &argsProcessEvent{
		timestamp: 1234,
		event:     event,
	}

	nftsPropertiesP := newNFTsPropertiesProcessor(&mock.PubkeyConverterMock{}, &mock.MarshalizerMock{})

	res := nftsPropertiesP.processEvent(args)
	require.True(t, res.processed)
	require.Equal(t, &data.NFTDataUpdate{
		Identifier: "TOUC-aaaa-01",
		URIsToAdd:  [][]byte{[]byte("uri1"), []byte("uri2")},
		Address:    "61646472",
	}, res.updatePropNFT)
}

func TestProcessNFTMetaDataRecreate(t *testing.T) {
	nftsPropertiesP := newNFTsPropertiesProcessor(&mock.PubkeyConverterMock{}, &mock.MarshalizerMock{})

	dcdtData := &dcdt.DCDigitalToken{
		TokenMetaData: &dcdt.MetaData{
			Creator: []byte("creator"),
		},
	}
	dcdtDataBytes, _ := json.Marshal(dcdtData)

	nonce := uint64(19)
	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(core.DCDTMetaDataRecreate),
		Topics:     [][]byte{[]byte("my-token"), big.NewInt(0).SetUint64(nonce).Bytes(), big.NewInt(1).Bytes(), dcdtDataBytes},
	}
	args := &argsProcessEvent{
		timestamp: 1234,
		event:     event,
	}

	res := nftsPropertiesP.processEvent(args)
	require.True(t, res.processed)
	require.NotNil(t, res.updatePropNFT.NewMetaData)
	require.Equal(t, hex.EncodeToString([]byte("creator")), res.updatePropNFT.NewMetaData.Creator)
}

func TestProcessNFTProperties_FreezeAndUnFreeze(t *testing.T) {
	t.Parallel()

	// freeze
	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte("DCDTFreeze"),
		Topics:     [][]byte{[]byte("TOUC-aaaa"), big.NewInt(1).Bytes(), nil, []byte("something")},
	}
	args := &argsProcessEvent{
		timestamp: 1234,
		event:     event,
	}

	nftsPropertiesP := newNFTsPropertiesProcessor(&mock.PubkeyConverterMock{}, &mock.MarshalizerMock{})

	res := nftsPropertiesP.processEvent(args)
	require.True(t, res.processed)
	require.True(t, res.updatePropNFT.Freeze)

	// unFreeze
	event = &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte("DCDTUnFreeze"),
		Topics:     [][]byte{[]byte("TOUC-aaaa"), big.NewInt(1).Bytes(), nil, []byte("something")},
	}
	args = &argsProcessEvent{
		timestamp: 1234,
		event:     event,
	}

	res = nftsPropertiesP.processEvent(args)
	require.True(t, res.processed)
	require.True(t, res.updatePropNFT.UnFreeze)
}

func TestProcessPauseAndUnPauseEvent(t *testing.T) {
	npp := &nftsPropertiesProc{}

	// test pause event
	result := npp.processPauseAndUnPauseEvent(core.BuiltInFunctionDCDTPause, "token1")
	require.True(t, result.processed, "Expected processed to be true")
	require.Equal(t, "token1", result.updatePropNFT.Identifier, "Expected identifier to be token1")
	require.True(t, result.updatePropNFT.Pause, "Expected pause to be true")
	require.False(t, result.updatePropNFT.UnPause, "Expected unpause to be false")

	// test unpause event
	result = npp.processPauseAndUnPauseEvent(core.BuiltInFunctionDCDTUnPause, "token2")
	require.True(t, result.processed, "Expected processed to be true")
	require.Equal(t, "token2", result.updatePropNFT.Identifier, "Expected identifier to be token2")
	require.False(t, result.updatePropNFT.Pause, "Expected pause to be false")
	require.True(t, result.updatePropNFT.UnPause, "Expected unpause to be true")

	// test wrong event
	result = npp.processPauseAndUnPauseEvent("wrong", "token2")
	require.Nil(t, result.updatePropNFT, "Expected updatePropNFT to be nil")
	require.True(t, result.processed, "Expected processed to be true")
}
