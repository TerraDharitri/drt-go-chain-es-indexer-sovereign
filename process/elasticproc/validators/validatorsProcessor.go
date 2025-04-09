package validators

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/TerraDharitri/drt-go-chain-core/core"
	"github.com/TerraDharitri/drt-go-chain-core/core/check"
	"github.com/TerraDharitri/drt-go-chain-core/data/outport"
	"github.com/TerraDharitri/drt-go-chain-es-indexer/data"
	"github.com/TerraDharitri/drt-go-chain-es-indexer/process/dataindexer"
)

type validatorsProcessor struct {
	bulkSizeMaxSize          int
	validatorPubkeyConverter core.PubkeyConverter
}

// NewValidatorsProcessor will create a new instance of validatorsProcessor
func NewValidatorsProcessor(validatorPubkeyConverter core.PubkeyConverter, bulkSizeMaxSize int) (*validatorsProcessor, error) {
	if check.IfNil(validatorPubkeyConverter) {
		return nil, dataindexer.ErrNilPubkeyConverter
	}

	return &validatorsProcessor{
		bulkSizeMaxSize:          bulkSizeMaxSize,
		validatorPubkeyConverter: validatorPubkeyConverter,
	}, nil
}

// PrepareAnSerializeValidatorsPubKeys will prepare validators public keys and serialize them
func (vp *validatorsProcessor) PrepareAnSerializeValidatorsPubKeys(validatorsPubKeys *outport.ValidatorsPubKeys) ([]*bytes.Buffer, error) {
	buffSlice := data.NewBufferSlice(vp.bulkSizeMaxSize)

	for shardID, validatorPk := range validatorsPubKeys.ShardValidatorsPubKeys {
		err := vp.prepareAndSerializeValidatorsKeysForShard(shardID, validatorsPubKeys.Epoch, validatorPk.Keys, buffSlice)
		if err != nil {
			return nil, err
		}
	}

	return buffSlice.Buffers(), nil
}

func (vp *validatorsProcessor) prepareAndSerializeValidatorsKeysForShard(shardID uint32, epoch uint32, keys [][]byte, buffSlice *data.BufferSlice) error {
	preparedValidatorsPubKeys := &data.ValidatorsPublicKeys{
		PublicKeys: make([]string, 0),
	}

	for _, key := range keys {
		// it will never throw an error here
		strValidatorPk, _ := vp.validatorPubkeyConverter.Encode(key)
		preparedValidatorsPubKeys.PublicKeys = append(preparedValidatorsPubKeys.PublicKeys, strValidatorPk)
	}

	id := fmt.Sprintf("%d_%d", shardID, epoch)
	meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, id, "\n"))

	serializedData, err := json.Marshal(preparedValidatorsPubKeys)
	if err != nil {
		return err
	}

	err = buffSlice.PutData(meta, serializedData)
	if err != nil {
		return err
	}

	return nil
}
