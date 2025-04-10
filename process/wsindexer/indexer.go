package wsindexer

import (
	"errors"
	"fmt"
	"time"

	"github.com/TerraDharitri/drt-go-chain-core/core/check"
	"github.com/TerraDharitri/drt-go-chain-core/data/outport"
	"github.com/TerraDharitri/drt-go-chain-core/marshal"
	"github.com/TerraDharitri/drt-go-chain-es-indexer/core"
	"github.com/TerraDharitri/drt-go-chain-es-indexer/metrics"
	"github.com/TerraDharitri/drt-go-chain-es-indexer/process/dataindexer"
	logger "github.com/TerraDharitri/drt-go-chain-logger"
)

var (
	log               = logger.GetOrCreate("process/wsindexer")
	errNilDataIndexer = errors.New("nil data indexer")
)

// ArgsIndexer holds all the components needed to create a new instance of indexer
type ArgsIndexer struct {
	Marshaller    marshal.Marshalizer
	DataIndexer   DataIndexer
	StatusMetrics core.StatusMetricsHandler
}

type indexer struct {
	marshaller    marshal.Marshalizer
	di            DataIndexer
	statusMetrics core.StatusMetricsHandler
	actions       map[string]func(marshalledData []byte) error
}

// NewIndexer will create a new instance of *indexer
func NewIndexer(args ArgsIndexer) (*indexer, error) {
	if check.IfNil(args.Marshaller) {
		return nil, dataindexer.ErrNilMarshalizer
	}
	if check.IfNil(args.DataIndexer) {
		return nil, errNilDataIndexer
	}
	if check.IfNil(args.StatusMetrics) {
		return nil, core.ErrNilMetricsHandler
	}

	payloadIndexer := &indexer{
		marshaller:    args.Marshaller,
		di:            args.DataIndexer,
		statusMetrics: args.StatusMetrics,
	}
	payloadIndexer.initActionsMap()

	return payloadIndexer, nil
}

// GetOperationsMap returns the map with all the operations that will index data
func (i *indexer) initActionsMap() {
	i.actions = map[string]func(d []byte) error{
		outport.TopicSaveBlock:             i.saveBlock,
		outport.TopicRevertIndexedBlock:    i.revertIndexedBlock,
		outport.TopicSaveRoundsInfo:        i.saveRounds,
		outport.TopicSaveValidatorsRating:  i.saveValidatorsRating,
		outport.TopicSaveValidatorsPubKeys: i.saveValidatorsPubKeys,
		outport.TopicSaveAccounts:          i.saveAccounts,
		outport.TopicFinalizedBlock:        i.finalizedBlock,
		outport.TopicSettings:              i.setSettings,
	}
}

// ProcessPayload will proces the provided payload based on the topic
func (i *indexer) ProcessPayload(payload []byte, topic string, version uint32) error {
	if version != 1 {
		log.Warn("received a payload with a different version", "version", version)
	}

	payloadTypeAction, ok := i.actions[topic]
	if !ok {
		log.Warn("invalid payload type", "topic", topic)
		return nil
	}

	shardID, err := i.getShardID(payload)
	if err != nil {
		log.Warn("indexer.ProcessPayload: cannot get shardID from payload", "error", err)
	}

	start := time.Now()
	err = payloadTypeAction(payload)
	duration := time.Since(start)

	topicKey := fmt.Sprintf("%s_%d", topic, shardID)
	i.statusMetrics.AddIndexingData(metrics.ArgsAddIndexingData{
		GotError:   err != nil,
		MessageLen: uint64(len(payload)),
		Topic:      topicKey,
		Duration:   duration,
	})

	return err
}

func (i *indexer) saveBlock(marshalledData []byte) error {
	outportBlock := &outport.OutportBlock{}
	err := i.marshaller.Unmarshal(outportBlock, marshalledData)
	if err != nil {
		return err
	}

	return i.di.SaveBlock(outportBlock)
}

func (i *indexer) revertIndexedBlock(marshalledData []byte) error {
	blockData := &outport.BlockData{}
	err := i.marshaller.Unmarshal(blockData, marshalledData)
	if err != nil {
		return err
	}

	return i.di.RevertIndexedBlock(blockData)
}

func (i *indexer) saveRounds(marshalledData []byte) error {
	roundsInfo := &outport.RoundsInfo{}
	err := i.marshaller.Unmarshal(roundsInfo, marshalledData)
	if err != nil {
		return err
	}

	return i.di.SaveRoundsInfo(roundsInfo)
}

func (i *indexer) saveValidatorsRating(marshalledData []byte) error {
	ratingData := &outport.ValidatorsRating{}
	err := i.marshaller.Unmarshal(ratingData, marshalledData)
	if err != nil {
		return err
	}

	return i.di.SaveValidatorsRating(ratingData)
}

func (i *indexer) saveValidatorsPubKeys(marshalledData []byte) error {
	validatorsPubKeys := &outport.ValidatorsPubKeys{}
	err := i.marshaller.Unmarshal(validatorsPubKeys, marshalledData)
	if err != nil {
		return err
	}

	return i.di.SaveValidatorsPubKeys(validatorsPubKeys)
}

func (i *indexer) saveAccounts(marshalledData []byte) error {
	accounts := &outport.Accounts{}
	err := i.marshaller.Unmarshal(accounts, marshalledData)
	if err != nil {
		return err
	}

	return i.di.SaveAccounts(accounts)
}

func (i *indexer) finalizedBlock(_ []byte) error {
	return nil
}

func (i *indexer) setSettings(marshalledData []byte) error {
	settings := outport.OutportConfig{}
	err := i.marshaller.Unmarshal(&settings, marshalledData)
	if err != nil {
		return err
	}

	return i.di.SetCurrentSettings(settings)
}

// Close will close the indexer
func (i *indexer) Close() error {
	return i.di.Close()
}

// IsInterfaceNil returns true if underlying object is nil
func (i *indexer) IsInterfaceNil() bool {
	return i == nil
}

func (i *indexer) getShardID(payload []byte) (uint32, error) {
	shard := &outport.Shard{}
	err := i.marshaller.Unmarshal(shard, payload)
	if err != nil {
		return 0, err
	}

	return shard.ShardID, nil
}
