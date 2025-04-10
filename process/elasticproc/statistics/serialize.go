package statistics

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/TerraDharitri/drt-go-chain-core/data/outport"
	"github.com/TerraDharitri/drt-go-chain-es-indexer/data"
	logger "github.com/TerraDharitri/drt-go-chain-logger"
)

var log = logger.GetOrCreate("indexer/process/statistics")

type statisticsProcessor struct {
}

// NewStatisticsProcessor will create a new instance of a statisticsProcessor
func NewStatisticsProcessor() *statisticsProcessor {
	return &statisticsProcessor{}
}

// SerializeRoundsInfo will serialize information about rounds
func (sp *statisticsProcessor) SerializeRoundsInfo(rounds *outport.RoundsInfo) *bytes.Buffer {
	buff := &bytes.Buffer{}
	for _, info := range rounds.RoundsInfo {
		serializedRoundInfo, meta := serializeRoundInfo(&data.RoundInfo{
			Round:            info.Round,
			SignersIndexes:   info.SignersIndexes,
			BlockWasProposed: info.BlockWasProposed,
			ShardId:          info.ShardId,
			Epoch:            info.Epoch,
			Timestamp:        time.Duration(info.Timestamp),
		})

		buff.Grow(len(meta) + len(serializedRoundInfo))
		_, err := buff.Write(meta)
		if err != nil {
			log.Warn("generalInfoProcessor.SaveRoundsInfo cannot write meta", "error", err)
		}

		_, err = buff.Write(serializedRoundInfo)
		if err != nil {
			log.Warn("generalInfoProcessor.SaveRoundsInfo cannot write serialized round info", "error", err)
		}
	}

	return buff
}

func serializeRoundInfo(info *data.RoundInfo) ([]byte, []byte) {
	meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%d_%d" } }%s`, info.ShardId, info.Round, "\n"))

	serializedInfo, err := json.Marshal(info)
	if err != nil {
		log.Warn("serializeRoundInfo could not serialize round info, will skip indexing this round info", "error", err)
		return nil, nil
	}

	serializedInfo = append(serializedInfo, "\n"...)

	return serializedInfo, meta
}
