package logsevents

import (
	"encoding/hex"
	"math/big"
	"strconv"
	"testing"

	"github.com/TerraDharitri/drt-go-chain-core/core"
	"github.com/TerraDharitri/drt-go-chain-core/data/transaction"
	"github.com/TerraDharitri/drt-go-chain-es-indexer/mock"
	"github.com/TerraDharitri/drt-go-chain-es-indexer/process/elasticproc/tokeninfo"
	"github.com/stretchr/testify/require"
)

func TestDcdtPropertiesProcCreateRoleShouldWork(t *testing.T) {
	t.Parallel()

	dcdtPropProc := newDcdtPropertiesProcessor(&mock.PubkeyConverterMock{})

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(core.BuiltInFunctionSetDCDTRole),
		Topics:     [][]byte{[]byte("MYTOKEN-abcd"), big.NewInt(0).Bytes(), big.NewInt(0).Bytes(), []byte(core.DCDTRoleNFTCreate)},
	}

	tokenRolesAndProperties := tokeninfo.NewTokenRolesAndProperties()
	dcdtPropProc.processEvent(&argsProcessEvent{
		event:                   event,
		tokenRolesAndProperties: tokenRolesAndProperties,
	})

	expected := map[string][]*tokeninfo.RoleData{
		core.DCDTRoleNFTCreate: {
			{
				Token:   "MYTOKEN-abcd",
				Set:     true,
				Address: "61646472",
			},
		},
	}
	require.Equal(t, expected, tokenRolesAndProperties.GetRoles())
}

func TestDcdtPropertiesProcTransferCreateRole(t *testing.T) {
	t.Parallel()

	dcdtPropProc := newDcdtPropertiesProcessor(&mock.PubkeyConverterMock{})

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(core.BuiltInFunctionDCDTNFTCreateRoleTransfer),
		Topics:     [][]byte{[]byte("MYTOKEN-abcd"), big.NewInt(0).Bytes(), big.NewInt(0).Bytes(), []byte(strconv.FormatBool(true))},
	}

	tokenRolesAndProperties := tokeninfo.NewTokenRolesAndProperties()
	dcdtPropProc.processEvent(&argsProcessEvent{
		event:                   event,
		tokenRolesAndProperties: tokenRolesAndProperties,
	})

	expected := map[string][]*tokeninfo.RoleData{
		core.DCDTRoleNFTCreate: {
			{
				Token:   "MYTOKEN-abcd",
				Set:     true,
				Address: "61646472",
			},
		},
	}
	require.Equal(t, expected, tokenRolesAndProperties.GetRoles())
}

func TestDcdtPropertiesProcUpgradeProperties(t *testing.T) {
	t.Parallel()

	dcdtPropProc := newDcdtPropertiesProcessor(&mock.PubkeyConverterMock{})

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(upgradePropertiesEvent),
		Topics:     [][]byte{[]byte("MYTOKEN-abcd"), big.NewInt(0).Bytes(), []byte("canMint"), []byte("true"), []byte("canBurn"), []byte("false")},
	}

	tokenRolesAndProperties := tokeninfo.NewTokenRolesAndProperties()
	dcdtPropProc.processEvent(&argsProcessEvent{
		event:                   event,
		tokenRolesAndProperties: tokenRolesAndProperties,
	})

	expected := []*tokeninfo.PropertiesData{
		{
			Token: "MYTOKEN-abcd",
			Properties: map[string]bool{
				"canMint": true,
				"canBurn": false,
			},
		},
	}
	require.Equal(t, expected, tokenRolesAndProperties.GetAllTokensWithProperties())
}

func TestCheckRolesBytes(t *testing.T) {
	t.Parallel()

	role1, _ := hex.DecodeString("01")
	role2, _ := hex.DecodeString("02")
	rolesBytes := [][]byte{role1, role2}
	require.False(t, checkRolesBytes(rolesBytes))

	role1 = []byte("DCDTRoleNFTCreate")
	rolesBytes = [][]byte{role1}
	require.True(t, checkRolesBytes(rolesBytes))
}
