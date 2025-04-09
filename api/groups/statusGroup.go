package groups

import (
	"fmt"
	"net/http"

	"github.com/TerraDharitri/drt-go-chain-core/core/check"
	"github.com/TerraDharitri/drt-go-chain-es-indexer/api/shared"
	"github.com/TerraDharitri/drt-go-chain-es-indexer/core"
	"github.com/gin-gonic/gin"
)

const (
	metricsPath           = "/metrics"
	prometheusMetricsPath = "/prometheus-metrics"
)

type statusGroup struct {
	*baseGroup
	facade shared.FacadeHandler
}

// NewStatusGroup returns a new instance of status group
func NewStatusGroup(facade shared.FacadeHandler) (*statusGroup, error) {
	if check.IfNil(facade) {
		return nil, fmt.Errorf("%w for status group", core.ErrNilFacadeHandler)
	}

	sg := &statusGroup{
		facade:    facade,
		baseGroup: &baseGroup{},
	}

	endpoints := []*shared.EndpointHandlerData{
		{
			Path:    metricsPath,
			Handler: sg.getMetrics,
			Method:  http.MethodGet,
		},
		{
			Path:    prometheusMetricsPath,
			Handler: sg.getPrometheusMetrics,
			Method:  http.MethodGet,
		},
	}
	sg.endpoints = endpoints

	return sg, nil
}

// getMetrics will expose endpoints statistics in json format
func (sg *statusGroup) getMetrics(c *gin.Context) {
	metricsResults := sg.facade.GetMetrics()

	returnStatus(c, gin.H{"metrics": metricsResults}, http.StatusOK, "", "successful")
}

// getPrometheusMetrics will expose proxy metrics in prometheus format
func (sg *statusGroup) getPrometheusMetrics(c *gin.Context) {
	metricsResults := sg.facade.GetMetricsForPrometheus()

	c.String(http.StatusOK, metricsResults)
}

// IsInterfaceNil returns true if there is no value under the interface
func (sg *statusGroup) IsInterfaceNil() bool {
	return sg == nil
}

func returnStatus(c *gin.Context, data interface{}, httpStatus int, err string, code string) {
	c.JSON(
		httpStatus,
		shared.GenericAPIResponse{
			Data:  data,
			Error: err,
			Code:  code,
		},
	)
}
