package metrics

import (
	"net/http"
	"testing"

	"github.com/TerraDharitri/drt-go-chain-core/data/outport"
	"github.com/TerraDharitri/drt-go-chain-es-indexer/core/request"
	"github.com/stretchr/testify/require"
)

func TestStatusMetrics_AddIndexingDataAndGetMetrics(t *testing.T) {
	t.Parallel()

	statusMetricsHandler := NewStatusMetrics()
	require.False(t, statusMetricsHandler.IsInterfaceNil())

	topic1 := "test1_0"
	statusMetricsHandler.AddIndexingData(ArgsAddIndexingData{
		GotError:   true,
		MessageLen: 100,
		Topic:      topic1,
		Duration:   12,
		StatusCode: http.StatusBadRequest,
	})
	statusMetricsHandler.AddIndexingData(ArgsAddIndexingData{
		GotError:   false,
		MessageLen: 222,
		Topic:      topic1,
		Duration:   15,
	})

	metrics := statusMetricsHandler.GetMetrics()
	require.Equal(t, &request.MetricsResponse{
		OperationsCount:   2,
		TotalErrorsCount:  1,
		TotalIndexingTime: 27,
		TotalData:         322,
		ErrorsCount: map[int]uint64{
			http.StatusBadRequest: 1,
		},
	}, metrics[topic1])

	prometheusMetrics := statusMetricsHandler.GetMetricsForPrometheus()
	require.Equal(t, `# TYPE test1 counter
test1{operation="total_data",shardID="0"} 322

# TYPE test1 counter
test1{operation="errors_count",shardID="0"} 1

# TYPE test1 counter
test1{operation="operations_count",shardID="0"} 2

# TYPE test1 counter
test1{operation="total_time",shardID="0"} 0

# TYPE test1 gauge
test1{operation="requests_errors",shardID="0",errorCode="400"} 1

`, prometheusMetrics)
}

func TestCamelCaseToSnakeCase(t *testing.T) {
	t.Parallel()

	require.Equal(t, "settings", camelToSnake(outport.TopicSettings))
	require.Equal(t, "save_validators_pub_keys", camelToSnake(outport.TopicSaveValidatorsPubKeys))
	require.Equal(t, "000000000000000", camelToSnake("000000000000000"))
	require.Equal(t, "one_one_one", camelToSnake("One_One_One"))
	require.Equal(t, "req_block", camelToSnake("req_block"))
}
