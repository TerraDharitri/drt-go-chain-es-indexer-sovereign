//go:build integrationtests

package integrationtests

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/TerraDharitri/drt-go-chain-core/core"
	dataBlock "github.com/TerraDharitri/drt-go-chain-core/data/block"
	"github.com/TerraDharitri/drt-go-chain-core/data/dcdt"
	"github.com/TerraDharitri/drt-go-chain-core/data/outport"
	"github.com/TerraDharitri/drt-go-chain-core/data/smartContractResult"
	"github.com/TerraDharitri/drt-go-chain-core/data/transaction"
	"github.com/stretchr/testify/require"

	indexerData "github.com/TerraDharitri/drt-go-chain-es-indexer/process/dataindexer"
	"github.com/TerraDharitri/drt-go-chain-es-indexer/process/elasticproc"
)

type esToken struct {
	Identifier  string
	Value       *big.Int
	NumDecimals int64
}

type esNft struct {
	Collection string
	Nonce      uint64
	Data       dcdt.DCDigitalToken
}

func createTokens() ([]esToken, []esNft) {
	tokens := []esToken{}
	token1 := esToken{
		Identifier:  "TKN18-1a2b3c",
		Value:       big.NewInt(123),
		NumDecimals: 18,
	}
	tokens = append(tokens, token1)
	token2 := esToken{
		Identifier:  "TKN12-1c2b3a",
		Value:       big.NewInt(333),
		NumDecimals: 12,
	}
	tokens = append(tokens, token2)

	nfts := []esNft{}
	nft := esNft{
		Collection: "NFT-abc123",
		Nonce:      1,
		Data: dcdt.DCDigitalToken{
			Type:       uint32(core.NonFungibleV2),
			Value:      big.NewInt(1),
			Properties: []byte("3032"),
			TokenMetaData: &dcdt.MetaData{
				Nonce:     1,
				Name:      []byte("NFT"),
				Creator:   []byte("creator"),
				Royalties: uint32(2500),
			},
		},
	}
	nfts = append(nfts, nft)

	return tokens, nfts
}

func TestCrossChainTokensIndexingFromMainChain(t *testing.T) {
	setLogLevelDebug()

	mainChainEsClient, err := createMainChainESClient(esMainChainURL, true)
	require.Nil(t, err)

	tokens, nfts := createTokens()
	createTokensInSourceEs(t, mainChainEsClient, tokens, nfts)

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	esProc, err := CreateSovereignElasticProcessor(esClient, mainChainEsClient)
	require.Nil(t, err)

	allTokens := getAllTokensIDs(tokens, nfts)
	allTokens = append(allTokens, getAllNftIDs(nfts)...)
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), allTokens, indexerData.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	for _, token := range genericResponse.Docs {
		require.False(t, token.Found)
	}

	scrHash := []byte("scrHash")
	header := &dataBlock.Header{
		Round:     10,
		TimeStamp: 2500,
	}
	body := &dataBlock.Body{
		MiniBlocks: dataBlock.MiniBlockSlice{
			{
				Type:            dataBlock.SmartContractResultBlock,
				SenderShardID:   core.MainChainShardId,
				ReceiverShardID: core.SovereignChainShardId,
				TxHashes:        [][]byte{scrHash},
			},
		},
	}

	pool := &outport.TransactionPool{
		SmartContractResults: map[string]*outport.SCRInfo{
			hex.EncodeToString(scrHash): {SmartContractResult: &smartContractResult.SmartContractResult{
				Nonce:          11,
				Value:          big.NewInt(0),
				GasLimit:       0,
				SndAddr:        decodeAddress("drt1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqzllls6prdez"),
				RcvAddr:        decodeAddress("drt1kzrfl2tztgzjpeedwec37c8npcr0a2ulzh9lhmj7xufyg23zcxuqmyhpvw"),
				Data:           createMultiDcdtTransferData(tokens, nfts),
				OriginalTxHash: nil,
			}, FeeInfo: &outport.FeeInfo{}},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, 1))
	require.Nil(t, err)

	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), []string{hex.EncodeToString(scrHash)}, indexerData.ScResultsIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t,
		readExpectedResult("./testdata/incomingSCR/incoming-scr.json"),
		string(genericResponse.Docs[0].Source),
	)

	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), allTokens, indexerData.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	for _, token := range genericResponse.Docs {
		require.True(t, token.Found)
	}
}

func createTokensInSourceEs(t *testing.T, esClient elasticproc.DatabaseClientHandler, tokens []esToken, nfts []esNft) {
	esProc, err := CreateElasticProcessor(esClient)
	require.Nil(t, err)

	body := &dataBlock.Body{}
	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
		ShardID:   core.MetachainShardId,
	}

	address1 := "drt1k04pxr6c0gvlcx4rd5fje0a4uy33axqxwz0fpcrgtfdy3nrqauqq4s3zqj"

	// create issue token and nft collection events
	events := make([]*transaction.Event, 0)
	for _, token := range tokens {
		events = append(events, &transaction.Event{
			Address:    decodeAddress(address1),
			Identifier: []byte("issue"),
			Topics:     [][]byte{[]byte(token.Identifier), []byte("TKN"), []byte("TKN"), []byte(core.FungibleDCDT), big.NewInt(token.NumDecimals).Bytes()},
		})
	}
	for _, nft := range nfts {
		events = append(events, &transaction.Event{
			Address:    decodeAddress(address1),
			Identifier: []byte("issueNonFungible"),
			Topics:     [][]byte{[]byte(nft.Collection), []byte("NFT"), []byte("NFT"), []byte(core.DCDTType(nft.Data.Type).String())},
		})
	}

	pool := &outport.TransactionPool{
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString([]byte("txHash1")),
				Log: &transaction.Log{
					Address: decodeAddress(address1),
					Events:  events,
				},
			},
		},
	}

	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	genericResponse := &GenericResponse{}
	allTokens := getAllTokensIDs(tokens, nfts)
	err = esClient.DoMultiGet(context.Background(), allTokens, indexerData.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	for _, token := range genericResponse.Docs {
		require.True(t, token.Found)
	}

	// create nft event
	events = make([]*transaction.Event, 0)
	for _, nft := range nfts {
		nftDataBytes, _ := json.Marshal(nft.Data)

		events = append(events, &transaction.Event{
			Address:    decodeAddress(address1),
			Identifier: []byte(core.BuiltInFunctionDCDTNFTCreate),
			Topics:     [][]byte{[]byte(nft.Collection), big.NewInt(0).SetUint64(nft.Nonce).Bytes(), nft.Data.Value.Bytes(), []byte(nftDataBytes)},
		})
	}

	header = &dataBlock.Header{
		Round:     51,
		TimeStamp: 5600,
		ShardID:   0,
	}

	pool = &outport.TransactionPool{
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString([]byte("txHash2")),
				Log: &transaction.Log{
					Address: decodeAddress(address1),
					Events:  events,
				},
			},
		},
	}

	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	allNfts := getAllNftIDs(nfts)
	err = esClient.DoMultiGet(context.Background(), allNfts, indexerData.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	for _, token := range genericResponse.Docs {
		require.True(t, token.Found)
	}
}

func getAllTokensIDs(tokens []esToken, nfts []esNft) []string {
	allTokens := make([]string, 0)
	for _, token := range tokens {
		allTokens = append(allTokens, token.Identifier)
	}
	for _, nft := range nfts {
		allTokens = append(allTokens, nft.Collection)
	}
	return allTokens
}

func getAllNftIDs(nfts []esNft) []string {
	allNfts := make([]string, 0)
	for _, nft := range nfts {
		nonceBytes := big.NewInt(0).SetUint64(nft.Nonce).Bytes()
		nonceHex := hex.EncodeToString(nonceBytes)
		nftIdentifier := nft.Collection + "-" + nonceHex

		allNfts = append(allNfts, nftIdentifier)

	}
	return allNfts
}

func createMultiDcdtTransferData(tokens []esToken, nfts []esNft) []byte {
	data := []byte(core.BuiltInFunctionMultiDCDTNFTTransfer +
		"@" + hex.EncodeToString(big.NewInt(int64(len(tokens)+len(nfts))).Bytes()))
	for _, token := range tokens {
		data = append(data, []byte(
			"@"+hex.EncodeToString([]byte(token.Identifier))+
				"@"+
				"@"+hex.EncodeToString(token.Value.Bytes()))...)
	}
	for _, nft := range nfts {
		nftDataBytes, _ := json.Marshal(nft.Data)
		data = append(data, []byte(
			"@"+hex.EncodeToString([]byte(nft.Collection))+
				"@"+hex.EncodeToString(big.NewInt(0).SetUint64(nft.Nonce).Bytes())+
				"@"+hex.EncodeToString(nftDataBytes))...)
	}

	return data
}
