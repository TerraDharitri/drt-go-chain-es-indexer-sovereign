//go:build integrationtests

package integrationtests

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/TerraDharitri/drt-go-chain-core/core"
	dataBlock "github.com/TerraDharitri/drt-go-chain-core/data/block"
	"github.com/TerraDharitri/drt-go-chain-core/data/dcdt"
	"github.com/TerraDharitri/drt-go-chain-core/data/outport"
	"github.com/TerraDharitri/drt-go-chain-core/data/transaction"
	indexerdata "github.com/TerraDharitri/drt-go-chain-es-indexer/process/dataindexer"
	"github.com/stretchr/testify/require"
)

func TestNFTUpdateMetadata(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	bigUri := bytes.Repeat([]byte("a"), 50000)
	dcdtCreateData := &dcdt.DCDigitalToken{
		TokenMetaData: &dcdt.MetaData{
			URIs: [][]byte{[]byte("uri"), []byte("uri"), bigUri, bigUri, bigUri},
		},
	}
	marshalizedCreate, _ := json.Marshal(dcdtCreateData)

	esProc, err := CreateElasticProcessor(esClient)
	require.Nil(t, err)

	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
		ShardID:   1,
	}
	body := &dataBlock.Body{}

	// CREATE NFT data
	address := "drt1w7jyzuj6cv4ngw8luhlkakatjpmjh3ql95lmxphd3vssc4vpymks82rg5q"
	pool := &outport.TransactionPool{
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString([]byte("h1")),
				Log: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address),
							Identifier: []byte(core.BuiltInFunctionDCDTNFTCreate),
							Topics:     [][]byte{[]byte("NFT-abcd"), big.NewInt(14).Bytes(), big.NewInt(1).Bytes(), marshalizedCreate},
						},
						nil,
					},
				},
			},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	ids := []string{"NFT-abcd-0e"}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/updateNFT/token.json"), string(genericResponse.Docs[0].Source))

	// Add URIS 1
	pool = &outport.TransactionPool{
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString([]byte("h1")),
				Log: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address),
							Identifier: []byte(core.BuiltInFunctionDCDTNFTAddURI),
							Topics:     [][]byte{[]byte("NFT-abcd"), big.NewInt(14).Bytes(), big.NewInt(0).Bytes(), []byte("uri"), bigUri},
						},
						nil,
					},
				},
			},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	// Add URIS 2 --- results should be the same
	pool = &outport.TransactionPool{
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString([]byte("h1")),
				Log: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address),
							Identifier: []byte(core.BuiltInFunctionDCDTNFTAddURI),
							Topics:     [][]byte{[]byte("NFT-abcd"), big.NewInt(14).Bytes(), big.NewInt(0).Bytes(), []byte("uri"), bigUri},
						},
						nil,
					},
				},
			},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	// Update attributes 1
	ids = []string{"NFT-abcd-0e"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/updateNFT/token-after-add-uris.json"), string(genericResponse.Docs[0].Source))

	pool = &outport.TransactionPool{
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString([]byte("h1")),
				Log: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address),
							Identifier: []byte(core.BuiltInFunctionDCDTNFTUpdateAttributes),
							Topics:     [][]byte{[]byte("NFT-abcd"), big.NewInt(14).Bytes(), big.NewInt(0).Bytes(), []byte("tags:test,free,fun;description:This is a test description for an awesome nft;metadata:metadata-test")},
						},
						nil,
					},
				},
			},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	ids = []string{"NFT-abcd-0e"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/updateNFT/token-after-update-attributes.json"), string(genericResponse.Docs[0].Source))

	// Update attributes 2

	pool = &outport.TransactionPool{
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString([]byte("h1")),
				Log: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address),
							Identifier: []byte(core.BuiltInFunctionDCDTNFTUpdateAttributes),
							Topics:     [][]byte{[]byte("NFT-abcd"), big.NewInt(14).Bytes(), big.NewInt(0).Bytes(), []byte("something")},
						},
						nil,
					},
				},
			},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	ids = []string{"NFT-abcd-0e"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/updateNFT/token-after-update-attributes-second.json"), string(genericResponse.Docs[0].Source))

	// Freeze nft
	pool = &outport.TransactionPool{
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString([]byte("h1")),
				Log: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address),
							Identifier: []byte(core.BuiltInFunctionDCDTFreeze),
							Topics:     [][]byte{[]byte("NFT-abcd"), big.NewInt(14).Bytes(), big.NewInt(0).Bytes(), []byte("something")},
						},
						nil,
					},
				},
			},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)
	ids = []string{"NFT-abcd-0e"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/updateNFT/token-after-freeze.json"), string(genericResponse.Docs[0].Source))

	// UnFreeze nft
	pool = &outport.TransactionPool{
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString([]byte("h1")),
				Log: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address),
							Identifier: []byte(core.BuiltInFunctionDCDTUnFreeze),
							Topics:     [][]byte{[]byte("NFT-abcd"), big.NewInt(14).Bytes(), big.NewInt(0).Bytes(), []byte("something")},
						},
						nil,
					},
				},
			},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)
	ids = []string{"NFT-abcd-0e"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/updateNFT/token-after-un-freeze.json"), string(genericResponse.Docs[0].Source))

	// Set new uris
	pool = &outport.TransactionPool{
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString([]byte("h1")),
				Log: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address),
							Identifier: []byte(core.DCDTSetNewURIs),
							Topics:     [][]byte{[]byte("NFT-abcd"), big.NewInt(14).Bytes(), big.NewInt(0).Bytes(), []byte("uri"), []byte("uri"), []byte("uri"), []byte("uri"), []byte("uri")},
						},
						nil,
					},
				},
			},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)
	ids = []string{"NFT-abcd-0e"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/updateNFT/token-after-set-new-uris.json"), string(genericResponse.Docs[0].Source))

	// new creator
	newCreator := "drt12m3x8jp6dl027pj5f2nw6ght2cyhhjfrs86cdwsa8xn83r375qfq7jkw93"
	pool = &outport.TransactionPool{
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString([]byte("h1")),
				Log: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(newCreator),
							Identifier: []byte(core.DCDTModifyCreator),
							Topics:     [][]byte{[]byte("NFT-abcd"), big.NewInt(14).Bytes(), big.NewInt(0).Bytes()},
						},
						nil,
					},
				},
			},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)
	ids = []string{"NFT-abcd-0e"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/updateNFT/token-after-new-creator.json"), string(genericResponse.Docs[0].Source))

	// new royalties
	pool = &outport.TransactionPool{
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString([]byte("h1")),
				Log: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address),
							Identifier: []byte(core.DCDTModifyRoyalties),
							Topics:     [][]byte{[]byte("NFT-abcd"), big.NewInt(14).Bytes(), big.NewInt(0).Bytes(), big.NewInt(100).Bytes()},
						},
						nil,
					},
				},
			},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)
	ids = []string{"NFT-abcd-0e"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/updateNFT/token-after-new-royalties.json"), string(genericResponse.Docs[0].Source))
}

func TestCreateNFTAndMetaDataRecreate(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	dcdtCreateData := &dcdt.DCDigitalToken{
		TokenMetaData: &dcdt.MetaData{
			Name: []byte("token-token-token"),
			URIs: [][]byte{[]byte("uri"), []byte("uri")},
		},
	}
	marshalizedCreate, _ := json.Marshal(dcdtCreateData)

	esProc, err := CreateElasticProcessor(esClient)
	require.Nil(t, err)

	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
		ShardID:   1,
	}
	body := &dataBlock.Body{}

	// CREATE NFT data
	address := "drt1w7jyzuj6cv4ngw8luhlkakatjpmjh3ql95lmxphd3vssc4vpymks82rg5q"
	pool := &outport.TransactionPool{
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString([]byte("h1")),
				Log: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address),
							Identifier: []byte(core.BuiltInFunctionDCDTNFTCreate),
							Topics:     [][]byte{[]byte("NEW-abcd"), big.NewInt(100).Bytes(), big.NewInt(1).Bytes(), marshalizedCreate},
						},
						nil,
					},
				},
			},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	ids := []string{"NEW-abcd-64"}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/updateNFT/token-before-recreate.json"), string(genericResponse.Docs[0].Source))

	// RECREATE
	reCreate := &dcdt.DCDigitalToken{
		TokenMetaData: &dcdt.MetaData{
			Name: []byte("token"),
			URIs: [][]byte{[]byte("uri")},
			Hash: []byte("hash"),
		},
	}
	marshalizedReCreate, _ := json.Marshal(reCreate)

	pool = &outport.TransactionPool{
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString([]byte("h1")),
				Log: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address),
							Identifier: []byte(core.DCDTMetaDataRecreate),
							Topics:     [][]byte{[]byte("NEW-abcd"), big.NewInt(100).Bytes(), big.NewInt(0).Bytes(), marshalizedReCreate},
						},
						nil,
					},
				},
			},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/updateNFT/token-after-recreate.json"), string(genericResponse.Docs[0].Source))

	// UPDATE
	update := &dcdt.DCDigitalToken{
		TokenMetaData: &dcdt.MetaData{
			Name: []byte("token-second"),
			URIs: [][]byte{[]byte("uri")},
			Hash: []byte("hash"),
		},
	}
	marshalizedUpdate, _ := json.Marshal(update)

	pool = &outport.TransactionPool{
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString([]byte("h1")),
				Log: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address),
							Identifier: []byte(core.DCDTMetaDataUpdate),
							Topics:     [][]byte{[]byte("NEW-abcd"), big.NewInt(100).Bytes(), big.NewInt(0).Bytes(), marshalizedUpdate},
						},
						nil,
					},
				},
			},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/updateNFT/token-after-update.json"), string(genericResponse.Docs[0].Source))
}

func TestMultipleDCDTMetadataRecreate(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	dcdtCreateData := &dcdt.DCDigitalToken{
		TokenMetaData: &dcdt.MetaData{
			Name: []byte("YELLOW"),
			URIs: [][]byte{[]byte("uri"), []byte("uri")},
		},
	}
	marshalizedCreate, _ := json.Marshal(dcdtCreateData)

	esProc, err := CreateElasticProcessor(esClient)
	require.Nil(t, err)

	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
		ShardID:   1,
	}
	body := &dataBlock.Body{}

	// CREATE NFT data
	address := "drt1w7jyzuj6cv4ngw8luhlkakatjpmjh3ql95lmxphd3vssc4vpymks82rg5q"
	pool := &outport.TransactionPool{
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString([]byte("h1")),
				Log: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address),
							Identifier: []byte(core.BuiltInFunctionDCDTNFTCreate),
							Topics:     [][]byte{[]byte("COLORS-df0e82"), big.NewInt(1).Bytes(), big.NewInt(1).Bytes(), marshalizedCreate},
						},
						nil,
					},
				},
			},
			{
				TxHash: hex.EncodeToString([]byte("h2")),
				Log: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address),
							Identifier: []byte(core.BuiltInFunctionDCDTNFTCreate),
							Topics:     [][]byte{[]byte("COLORS-df0e82"), big.NewInt(2).Bytes(), big.NewInt(1).Bytes(), marshalizedCreate},
						},
						nil,
					},
				},
			},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	// RECREATE
	reCreate := &dcdt.DCDigitalToken{
		TokenMetaData: &dcdt.MetaData{
			Name: []byte("GREEN"),
			URIs: [][]byte{[]byte("uri")},
			Hash: []byte("hash"),
		},
	}
	marshalizedReCreate, _ := json.Marshal(reCreate)

	pool = &outport.TransactionPool{
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString([]byte("h1")),
				Log: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address),
							Identifier: []byte(core.DCDTMetaDataRecreate),
							Topics:     [][]byte{[]byte("COLORS-df0e82"), big.NewInt(1).Bytes(), big.NewInt(0).Bytes(), marshalizedReCreate},
						},
						nil,
					},
				},
			},
			{
				TxHash: hex.EncodeToString([]byte("h2")),
				Log: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address),
							Identifier: []byte(core.DCDTMetaDataRecreate),
							Topics:     [][]byte{[]byte("COLORS-df0e82"), big.NewInt(2).Bytes(), big.NewInt(0).Bytes(), marshalizedReCreate},
						},
						nil,
					},
				},
			},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	ids := []string{"COLORS-df0e82-01", "COLORS-df0e82-02"}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/updateNFT/token-color-1.json"), string(genericResponse.Docs[0].Source))
	require.JSONEq(t, readExpectedResult("./testdata/updateNFT/token-color-2.json"), string(genericResponse.Docs[1].Source))
}
