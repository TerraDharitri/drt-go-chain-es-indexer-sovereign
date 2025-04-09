package templatesAndPolicies

import (
	"bytes"

	"github.com/TerraDharitri/drt-go-chain-es-indexer/templates"
)

// TemplatesAndPoliciesHandler  defines the actions that a templates and policies handler should do
type TemplatesAndPoliciesHandler interface {
	GetElasticTemplatesAndPolicies() (map[string]*bytes.Buffer, map[string]*bytes.Buffer, error)
	GetExtraMappings() ([]templates.ExtraMapping, error)
}
