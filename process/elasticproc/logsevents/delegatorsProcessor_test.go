package logsevents

import (
	"encoding/hex"
	"math/big"
	"strconv"
	"testing"

	"github.com/TerraDharitri/drt-go-chain-core/core"
	"github.com/TerraDharitri/drt-go-chain-core/data/transaction"
	"github.com/TerraDharitri/drt-go-chain-es-indexer/data"
	"github.com/TerraDharitri/drt-go-chain-es-indexer/mock"
	"github.com/TerraDharitri/drt-go-chain-es-indexer/process/elasticproc/converters"
	"github.com/stretchr/testify/require"
)

func TestDelegatorsProcessor_ProcessEvent(t *testing.T) {
	t.Parallel()

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(delegateFunc),
		Topics:     [][]byte{big.NewInt(1000).Bytes(), big.NewInt(1000000000).Bytes(), big.NewInt(10).Bytes(), big.NewInt(1000000000).Bytes()},
	}
	args := &argsProcessEvent{
		timestamp:   1234,
		event:       event,
		logAddress:  []byte("contract"),
		selfShardID: core.MetachainShardId,
	}

	balanceConverter, _ := converters.NewBalanceConverter(10)
	delegatorsProcessor := newDelegatorsProcessor(&mock.PubkeyConverterMock{}, balanceConverter)

	res := delegatorsProcessor.processEvent(args)
	require.True(t, res.processed)
	require.Equal(t, &data.Delegator{
		Address:        "61646472",
		Contract:       "636f6e7472616374",
		ActiveStakeNum: 0.1,
		ActiveStake:    "1000000000",
		Timestamp:      1234,
	}, res.delegator)
}

func TestDelegatorProcessor_WithdrawWithDelete(t *testing.T) {
	t.Parallel()

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(withdrawFunc),
		Topics:     [][]byte{big.NewInt(1000).Bytes(), big.NewInt(0).Bytes(), big.NewInt(10).Bytes(), big.NewInt(1000000000).Bytes(), []byte(strconv.FormatBool(true)), []byte("a")},
	}
	args := &argsProcessEvent{
		timestamp:   1234,
		event:       event,
		logAddress:  []byte("contract"),
		selfShardID: core.MetachainShardId,
	}

	balanceConverter, _ := converters.NewBalanceConverter(10)
	delegatorsProcessor := newDelegatorsProcessor(&mock.PubkeyConverterMock{}, balanceConverter)

	res := delegatorsProcessor.processEvent(args)
	require.True(t, res.processed)
	require.Equal(t, &data.Delegator{
		Address:         "61646472",
		Contract:        "636f6e7472616374",
		ActiveStakeNum:  0,
		ActiveStake:     "0",
		ShouldDelete:    true,
		Timestamp:       1234,
		WithdrawFundIDs: []string{"61"},
	}, res.delegator)
}

func TestDelegatorProcessor_ClaimRewardsWithDelete(t *testing.T) {
	t.Parallel()

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(claimRewardsFunc),
		Topics:     [][]byte{big.NewInt(1000).Bytes(), []byte(strconv.FormatBool(true))},
	}
	args := &argsProcessEvent{
		timestamp:   1234,
		event:       event,
		logAddress:  []byte("contract"),
		selfShardID: core.MetachainShardId,
	}

	balanceConverter, _ := converters.NewBalanceConverter(10)
	delegatorsProcessor := newDelegatorsProcessor(&mock.PubkeyConverterMock{}, balanceConverter)

	res := delegatorsProcessor.processEvent(args)
	require.True(t, res.processed)
	require.Equal(t, &data.Delegator{
		Address:      "61646472",
		Contract:     "636f6e7472616374",
		ShouldDelete: true,
	}, res.delegator)
}

func TestDelegatorProcessor_ClaimRewardsContractAddressInTopics(t *testing.T) {
	t.Parallel()

	contractAddress := []byte("contract2")
	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(claimRewardsFunc),
		Topics:     [][]byte{big.NewInt(1000).Bytes(), []byte(strconv.FormatBool(true)), contractAddress},
	}
	args := &argsProcessEvent{
		timestamp:   1234,
		event:       event,
		logAddress:  []byte("contract1"),
		selfShardID: core.MetachainShardId,
	}

	balanceConverter, _ := converters.NewBalanceConverter(10)
	delegatorsProcessor := newDelegatorsProcessor(&mock.PubkeyConverterMock{}, balanceConverter)

	res := delegatorsProcessor.processEvent(args)
	require.True(t, res.processed)
	require.Equal(t, &data.Delegator{
		Address:      "61646472",
		Contract:     hex.EncodeToString(contractAddress),
		ShouldDelete: true,
	}, res.delegator)
}

func TestDelegatorProcessor_ClaimRewardsNoDelete(t *testing.T) {
	t.Parallel()

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(claimRewardsFunc),
		Topics:     [][]byte{big.NewInt(1000).Bytes(), []byte(strconv.FormatBool(false))},
	}
	args := &argsProcessEvent{
		timestamp:   1234,
		event:       event,
		logAddress:  []byte("contract"),
		selfShardID: core.MetachainShardId,
	}

	balanceConverter, _ := converters.NewBalanceConverter(10)
	delegatorsProcessor := newDelegatorsProcessor(&mock.PubkeyConverterMock{}, balanceConverter)

	res := delegatorsProcessor.processEvent(args)
	require.True(t, res.processed)
	require.Nil(t, res.delegator)
}

func TestDelegatorsProcessor_WithdrawalShouldWorkWith5Topics(t *testing.T) {
	t.Parallel()

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(withdrawFunc),
		Topics:     [][]byte{big.NewInt(1000).Bytes(), big.NewInt(0).Bytes(), big.NewInt(10).Bytes(), big.NewInt(1000000000).Bytes(), []byte(strconv.FormatBool(true))},
	}
	args := &argsProcessEvent{
		timestamp:   1234,
		event:       event,
		logAddress:  []byte("contract"),
		selfShardID: core.MetachainShardId,
	}

	balanceConverter, _ := converters.NewBalanceConverter(10)
	delegatorsProcessor := newDelegatorsProcessor(&mock.PubkeyConverterMock{}, balanceConverter)

	res := delegatorsProcessor.processEvent(args)
	require.True(t, res.processed)
	require.True(t, res.delegator.ShouldDelete)
	require.Equal(t, 0, len(res.delegator.WithdrawFundIDs))
}

func TestDelegatorsProcessor_WithdrawalShouldWorkWithNewTopics(t *testing.T) {
	t.Parallel()

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(withdrawFunc),
		Topics:     [][]byte{big.NewInt(1000).Bytes(), big.NewInt(0).Bytes(), big.NewInt(10).Bytes(), big.NewInt(1000000000).Bytes(), []byte(strconv.FormatBool(true)), []byte("id1"), []byte("id2")},
	}
	args := &argsProcessEvent{
		timestamp:   1234,
		event:       event,
		logAddress:  []byte("contract"),
		selfShardID: core.MetachainShardId,
	}

	balanceConverter, _ := converters.NewBalanceConverter(10)
	delegatorsProcessor := newDelegatorsProcessor(&mock.PubkeyConverterMock{}, balanceConverter)

	res := delegatorsProcessor.processEvent(args)
	require.True(t, res.processed)
	require.True(t, res.delegator.ShouldDelete)
	require.Equal(t, []string{"696431", "696432"}, res.delegator.WithdrawFundIDs)
}
