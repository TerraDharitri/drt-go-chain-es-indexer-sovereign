package logsevents

import (
	"testing"
	"time"

	"github.com/TerraDharitri/drt-go-chain-core/core"
	"github.com/TerraDharitri/drt-go-chain-core/data/transaction"
	"github.com/TerraDharitri/drt-go-chain-es-indexer/data"
	"github.com/TerraDharitri/drt-go-chain-es-indexer/mock"
	"github.com/stretchr/testify/require"
)

func TestIssueDCDTProcessor(t *testing.T) {
	t.Parallel()

	dcdtIssueProc := newDCDTIssueProcessor(&mock.PubkeyConverterMock{})

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(issueNonFungibleDCDTFunc),
		Topics:     [][]byte{[]byte("MYTOKEN-abcd"), []byte("my-token"), []byte("MYTOKEN"), []byte(core.NonFungibleDCDT)},
	}
	args := &argsProcessEvent{
		timestamp:   1234,
		event:       event,
		selfShardID: core.MetachainShardId,
	}

	res := dcdtIssueProc.processEvent(args)

	require.Equal(t, &data.TokenInfo{
		Token:        "MYTOKEN-abcd",
		Name:         "my-token",
		Ticker:       "MYTOKEN",
		Timestamp:    time.Duration(1234),
		Type:         core.NonFungibleDCDT,
		Issuer:       "61646472",
		CurrentOwner: "61646472",
		OwnersHistory: []*data.OwnerData{
			{
				Address:   "61646472",
				Timestamp: time.Duration(1234),
			},
		},
		Properties: &data.TokenProperties{},
	}, res.tokenInfo)
}

func TestIssueDCDTProcessor_TransferOwnership(t *testing.T) {
	t.Parallel()

	dcdtIssueProc := newDCDTIssueProcessor(&mock.PubkeyConverterMock{})

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(transferOwnershipFunc),
		Topics:     [][]byte{[]byte("MYTOKEN-abcd"), []byte("my-token"), []byte("MYTOKEN"), []byte(core.NonFungibleDCDT), []byte("newOwner")},
	}
	args := &argsProcessEvent{
		timestamp:   1234,
		event:       event,
		selfShardID: core.MetachainShardId,
	}

	res := dcdtIssueProc.processEvent(args)

	require.Equal(t, &data.TokenInfo{
		Token:        "MYTOKEN-abcd",
		Name:         "my-token",
		Ticker:       "MYTOKEN",
		Timestamp:    time.Duration(1234),
		Type:         core.NonFungibleDCDT,
		Issuer:       "61646472",
		CurrentOwner: "6e65774f776e6572",
		OwnersHistory: []*data.OwnerData{
			{
				Address:   "6e65774f776e6572",
				Timestamp: time.Duration(1234),
			},
		},
		TransferOwnership: true,
		Properties:        &data.TokenProperties{},
	}, res.tokenInfo)
}
