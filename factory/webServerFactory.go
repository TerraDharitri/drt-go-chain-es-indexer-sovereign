package factory

import (
	"github.com/TerraDharitri/drt-go-chain-es-indexer/api/gin"
	"github.com/TerraDharitri/drt-go-chain-es-indexer/config"
	"github.com/TerraDharitri/drt-go-chain-es-indexer/core"
	"github.com/TerraDharitri/drt-go-chain-es-indexer/facade"
)

// CreateWebServer will create a new instance of core.WebServerHandler
func CreateWebServer(apiConfig config.ApiRoutesConfig, statusMetricsHandler core.StatusMetricsHandler) (core.WebServerHandler, error) {
	metricsFacade, err := facade.NewMetricsFacade(statusMetricsHandler)
	if err != nil {
		return nil, err
	}

	args := gin.ArgsWebServer{
		Facade:    metricsFacade,
		ApiConfig: apiConfig,
	}
	return gin.NewWebServer(args)
}
