package facade

import (
	"github.com/TerraDharitri/drt-go-chain-core/core/check"
	"github.com/TerraDharitri/drt-go-chain-es-indexer/core"
	"github.com/TerraDharitri/drt-go-chain-es-indexer/core/request"
)

type metricsFacade struct {
	statusMetrics core.StatusMetricsHandler
}

// NewMetricsFacade will create a new instance of metricsFacade
func NewMetricsFacade(statusMetrics core.StatusMetricsHandler) (*metricsFacade, error) {
	if check.IfNil(statusMetrics) {
		return nil, core.ErrNilMetricsHandler
	}

	return &metricsFacade{
		statusMetrics: statusMetrics,
	}, nil
}

// GetMetrics will return metrics in json format
func (mf *metricsFacade) GetMetrics() map[string]*request.MetricsResponse {
	return mf.statusMetrics.GetMetrics()
}

// GetMetricsForPrometheus will return metrics in prometheus format
func (mf *metricsFacade) GetMetricsForPrometheus() string {
	return mf.statusMetrics.GetMetricsForPrometheus()
}

// IsInterfaceNil returns true if there is no value under the interface
func (mf *metricsFacade) IsInterfaceNil() bool {
	return mf == nil
}
