//go:build integrationtests

package integrationtests

import (
	"context"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/TerraDharitri/drt-go-chain-core/core"
	dataBlock "github.com/TerraDharitri/drt-go-chain-core/data/block"
	"github.com/TerraDharitri/drt-go-chain-core/data/outport"
	"github.com/TerraDharitri/drt-go-chain-core/data/smartContractResult"
	"github.com/TerraDharitri/drt-go-chain-core/data/transaction"
	indexerdata "github.com/TerraDharitri/drt-go-chain-es-indexer/process/dataindexer"
	"github.com/stretchr/testify/require"
)

func TestTransactionWithClaimRewardsGasRefund(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	esProc, err := CreateElasticProcessor(esClient)
	require.Nil(t, err)

	txHash := []byte("claimRewards")
	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
		ShardID:   core.MetachainShardId,
	}

	scrHash1 := []byte("scrRefundGasReward")
	body := &dataBlock.Body{
		MiniBlocks: dataBlock.MiniBlockSlice{
			{
				Type:            dataBlock.TxBlock,
				SenderShardID:   0,
				ReceiverShardID: core.MetachainShardId,
				TxHashes:        [][]byte{txHash},
			},
			{
				Type:            dataBlock.SmartContractResultBlock,
				SenderShardID:   core.MetachainShardId,
				ReceiverShardID: 0,
				TxHashes:        [][]byte{scrHash1},
			},
		},
	}

	addressSender := "drt14wnzmpwhcm9up7lsrujcf7jne2lgnydcpkfwk0etlnndn5dcacksurylsq"
	addressReceiver := "drt1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqplllskzf8kp"

	refundValue, _ := big.NewInt(0).SetString("49320000000000", 10)
	scr1 := &smartContractResult.SmartContractResult{
		Nonce:          618,
		GasPrice:       1000000000,
		SndAddr:        decodeAddress(addressReceiver),
		RcvAddr:        decodeAddress(addressSender),
		Data:           []byte("@6f6b"),
		PrevTxHash:     txHash,
		OriginalTxHash: txHash,
		Value:          refundValue,
	}

	rewards, _ := big.NewInt(0).SetString("2932360285576807", 10)
	scrHash2 := []byte("scrRewards")
	scr2 := &smartContractResult.SmartContractResult{
		Nonce:          0,
		GasPrice:       1000000000,
		SndAddr:        decodeAddress(addressReceiver),
		RcvAddr:        decodeAddress(addressSender),
		PrevTxHash:     txHash,
		OriginalTxHash: txHash,
		Value:          rewards,
	}

	tx1 := &transaction.Transaction{
		Nonce:    617,
		SndAddr:  decodeAddress(addressSender),
		RcvAddr:  decodeAddress(addressReceiver),
		GasLimit: 6000000,
		GasPrice: 1000000000,
		Data:     []byte("claimRewards"),
		Value:    big.NewInt(0),
	}

	txInfo := &outport.TxInfo{
		Transaction: tx1,
		FeeInfo: &outport.FeeInfo{
			GasUsed:        1068000,
			Fee:            big.NewInt(78000000000000),
			InitialPaidFee: big.NewInt(127320000000000),
		},
		ExecutionOrder: 0,
	}

	pool := &outport.TransactionPool{
		Transactions: map[string]*outport.TxInfo{
			hex.EncodeToString(txHash): txInfo,
		},
		SmartContractResults: map[string]*outport.SCRInfo{
			hex.EncodeToString(scrHash2): {SmartContractResult: scr2, FeeInfo: &outport.FeeInfo{}},
			hex.EncodeToString(scrHash1): {SmartContractResult: scr1, FeeInfo: &outport.FeeInfo{}},
		},
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString(txHash),
				Log: &transaction.Log{
					Address: decodeAddress(addressSender),
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(addressSender),
							Identifier: []byte("writeLog"),
							Topics:     [][]byte{[]byte("something")},
						},
					},
				},
			},
		},
	}

	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	ids := []string{hex.EncodeToString(txHash)}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	require.JSONEq(t,
		readExpectedResult("./testdata/claimRewards/tx-claim-rewards.json"),
		string(genericResponse.Docs[0].Source),
	)
}
