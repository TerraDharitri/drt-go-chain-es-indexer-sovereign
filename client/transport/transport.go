package transport

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/TerraDharitri/drt-go-chain-core/core/check"
	"github.com/TerraDharitri/drt-go-chain-es-indexer/core"
	"github.com/TerraDharitri/drt-go-chain-es-indexer/core/request"
	"github.com/TerraDharitri/drt-go-chain-es-indexer/metrics"
)

var errNilRequest = errors.New("nil request")

type metricsTransport struct {
	statusMetrics core.StatusMetricsHandler
	transport     http.RoundTripper
}

// NewMetricsTransport will create a new instance of metricsTransport
func NewMetricsTransport(statusMetrics core.StatusMetricsHandler) (*metricsTransport, error) {
	if check.IfNil(statusMetrics) {
		return nil, core.ErrNilMetricsHandler
	}

	return &metricsTransport{
		statusMetrics: statusMetrics,
		transport:     http.DefaultTransport,
	}, nil
}

// RoundTrip implements the http.RoundTripper interface and is used as a wrapper around the underlying
// transport to collect and record metrics related to the HTTP request/response cycle.
func (m *metricsTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req == nil {
		return nil, errNilRequest
	}

	startTime := time.Now()
	size := req.ContentLength

	var statusCode int
	resp, err := m.transport.RoundTrip(req)
	if err == nil {
		statusCode = resp.StatusCode
	}

	duration := time.Since(startTime)

	valueFromCtx := req.Context().Value(request.ContextKey)
	if valueFromCtx == nil {
		return resp, err
	}
	topic := fmt.Sprintf("%s", valueFromCtx)

	m.statusMetrics.AddIndexingData(metrics.ArgsAddIndexingData{
		StatusCode: statusCode,
		GotError:   err != nil,
		MessageLen: uint64(size),
		Topic:      topic,
		Duration:   duration,
	})

	return resp, err
}
