//go:build integrationtests

package integrationtests

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"testing"

	"github.com/TerraDharitri/drt-go-chain-core/core"
	"github.com/TerraDharitri/drt-go-chain-core/data/alteredAccount"
	dataBlock "github.com/TerraDharitri/drt-go-chain-core/data/block"
	"github.com/TerraDharitri/drt-go-chain-core/data/dcdt"
	"github.com/TerraDharitri/drt-go-chain-core/data/outport"
	"github.com/TerraDharitri/drt-go-chain-core/data/transaction"
	indexerdata "github.com/TerraDharitri/drt-go-chain-es-indexer/process/dataindexer"
	"github.com/stretchr/testify/require"
)

func TestAccountsDCDTDeleteOnRollback(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	dcdtToken := &dcdt.DCDigitalToken{
		Value:      big.NewInt(1000),
		Properties: []byte("3032"),
		TokenMetaData: &dcdt.MetaData{
			Creator: []byte("creator"),
		},
	}
	addr := "drt1sqy2ywvswp09ef7qwjhv8zwr9kzz3xas6y2ye5nuryaz0wcnfzzswuc7c0"
	coreAlteredAccounts := map[string]*alteredAccount.AlteredAccount{
		addr: {
			Address: addr,
			Tokens: []*alteredAccount.AccountTokenData{
				{
					Identifier: "TOKEN-eeee",
					Nonce:      2,
					Balance:    "1000",
					MetaData: &alteredAccount.TokenMetaData{
						Creator: "creator",
					},
					Properties: "3032",
				},
			},
		},
	}

	esProc, err := CreateElasticProcessor(esClient)
	require.Nil(t, err)

	// CREATE SEMI-FUNGIBLE TOKEN
	dcdtDataBytes, _ := json.Marshal(dcdtToken)
	pool := &outport.TransactionPool{
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString([]byte("h1")),
				Log: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(addr),
							Identifier: []byte(core.BuiltInFunctionDCDTNFTCreate),
							Topics:     [][]byte{[]byte("TOKEN-eeee"), big.NewInt(2).Bytes(), big.NewInt(1).Bytes(), dcdtDataBytes},
						},
						nil,
					},
				},
			},
		},
	}

	body := &dataBlock.Body{}
	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
		ShardID:   2,
	}

	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, coreAlteredAccounts, testNumOfShards))
	require.Nil(t, err)

	ids := []string{fmt.Sprintf("%s-TOKEN-eeee-02", addr)}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.AccountsDCDTIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/accountsDCDTRollback/account-after-create.json"), string(genericResponse.Docs[0].Source))

	// DO ROLLBACK
	err = esProc.RemoveAccountsDCDT(5040, 2)
	require.Nil(t, err)

	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.AccountsDCDTIndex, true, genericResponse)
	require.Nil(t, err)
	require.False(t, genericResponse.Docs[0].Found)
}
