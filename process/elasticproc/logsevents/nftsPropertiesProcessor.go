package logsevents

import (
	"math/big"

	"github.com/TerraDharitri/drt-go-chain-core/core"
	"github.com/TerraDharitri/drt-go-chain-core/data/dcdt"
	"github.com/TerraDharitri/drt-go-chain-core/marshal"
	"github.com/TerraDharitri/drt-go-chain-es-indexer/data"
	"github.com/TerraDharitri/drt-go-chain-es-indexer/process/elasticproc/converters"
)

const minTopicsUpdate = 4

type nftsPropertiesProc struct {
	marshaller                 marshal.Marshalizer
	pubKeyConverter            core.PubkeyConverter
	propertiesChangeOperations map[string]struct{}
}

func newNFTsPropertiesProcessor(pubKeyConverter core.PubkeyConverter, marshaller marshal.Marshalizer) *nftsPropertiesProc {
	return &nftsPropertiesProc{
		marshaller:      marshaller,
		pubKeyConverter: pubKeyConverter,
		propertiesChangeOperations: map[string]struct{}{
			core.BuiltInFunctionDCDTNFTAddURI:           {},
			core.BuiltInFunctionDCDTNFTUpdateAttributes: {},
			core.BuiltInFunctionDCDTFreeze:              {},
			core.BuiltInFunctionDCDTUnFreeze:            {},
			core.BuiltInFunctionDCDTPause:               {},
			core.BuiltInFunctionDCDTUnPause:             {},
			core.DCDTMetaDataRecreate:                   {},
			core.DCDTMetaDataUpdate:                     {},
			core.DCDTSetNewURIs:                         {},
			core.DCDTModifyCreator:                      {},
			core.DCDTModifyRoyalties:                    {},
		},
	}
}

func (npp *nftsPropertiesProc) processEvent(args *argsProcessEvent) argOutputProcessEvent {
	//nolint
	eventIdentifier := string(args.event.GetIdentifier())
	_, ok := npp.propertiesChangeOperations[eventIdentifier]
	if !ok {
		return argOutputProcessEvent{}
	}

	callerAddress := npp.pubKeyConverter.SilentEncode(args.event.GetAddress(), log)
	if callerAddress == "" {
		return argOutputProcessEvent{
			processed: true,
		}
	}

	topics := args.event.GetTopics()
	if len(topics) == 1 {
		return npp.processPauseAndUnPauseEvent(eventIdentifier, string(topics[0]))
	}

	// topics contains:
	// [0] --> token identifier
	// [1] --> nonce of the NFT (bytes)
	// [2] --> value
	// [3:] --> modified data
	// [3] --> DCDT token data in case of DCDTMetaDataRecreate

	isModifyCreator := len(topics) == minTopicsUpdate-1 && eventIdentifier == core.DCDTModifyCreator
	if len(topics) < minTopicsUpdate && !isModifyCreator {
		return argOutputProcessEvent{
			processed: true,
		}
	}

	callerAddress = npp.pubKeyConverter.SilentEncode(args.event.GetAddress(), log)
	if callerAddress == "" {
		return argOutputProcessEvent{
			processed: true,
		}
	}

	nonceBig := big.NewInt(0).SetBytes(topics[1])
	if nonceBig.Uint64() == 0 {
		// this is a fungible token so we should return
		return argOutputProcessEvent{}
	}

	token := string(topics[0])
	identifier := converters.ComputeTokenIdentifier(token, nonceBig.Uint64())

	updateNFT := &data.NFTDataUpdate{
		Identifier: identifier,
		Address:    callerAddress,
	}

	switch eventIdentifier {
	case core.BuiltInFunctionDCDTNFTUpdateAttributes:
		updateNFT.NewAttributes = topics[3]
	case core.BuiltInFunctionDCDTNFTAddURI:
		updateNFT.URIsToAdd = topics[3:]
	case core.DCDTSetNewURIs:
		updateNFT.SetURIs = true
		updateNFT.URIsToAdd = topics[3:]
	case core.BuiltInFunctionDCDTFreeze:
		updateNFT.Freeze = true
	case core.BuiltInFunctionDCDTUnFreeze:
		updateNFT.UnFreeze = true
	case core.DCDTMetaDataRecreate, core.DCDTMetaDataUpdate:
		npp.processMetaDataUpdate(updateNFT, topics[3])
	case core.DCDTModifyCreator:
		updateNFT.NewCreator = callerAddress
	case core.DCDTModifyRoyalties:
		newRoyalties := uint32(big.NewInt(0).SetBytes(topics[3]).Uint64())
		updateNFT.NewRoyalties = core.OptionalUint32{
			Value:    newRoyalties,
			HasValue: true,
		}
	}

	return argOutputProcessEvent{
		processed:     true,
		updatePropNFT: updateNFT,
	}
}

func (npp *nftsPropertiesProc) processMetaDataUpdate(updateNFT *data.NFTDataUpdate, dcdtTokenBytes []byte) {
	dcdtToken := &dcdt.DCDigitalToken{}
	err := npp.marshaller.Unmarshal(dcdtToken, dcdtTokenBytes)
	if err != nil {
		log.Warn("nftsPropertiesProc.processMetaDataRecreate() cannot urmarshal", "error", err.Error())
		return
	}

	tokenMetaData := converters.PrepareTokenMetaData(convertMetaData(npp.pubKeyConverter, dcdtToken.TokenMetaData))
	updateNFT.NewMetaData = tokenMetaData
}

func (npp *nftsPropertiesProc) processPauseAndUnPauseEvent(eventIdentifier string, token string) argOutputProcessEvent {
	var updateNFT *data.NFTDataUpdate

	switch eventIdentifier {
	case core.BuiltInFunctionDCDTPause:
		updateNFT = &data.NFTDataUpdate{
			Identifier: token,
			Pause:      true,
		}
	case core.BuiltInFunctionDCDTUnPause:
		updateNFT = &data.NFTDataUpdate{
			Identifier: token,
			UnPause:    true,
		}
	}

	return argOutputProcessEvent{
		processed:     true,
		updatePropNFT: updateNFT,
	}
}
