package abstract_plugin

import "context"

type TaxService interface {
	// GetProvincialTaxEntities returns an array of tax bracket info for each given input
	GetProvincialTaxEntities(context.Context, []*TaxCalculatorPluginInputParameters) (<-chan []*TaxEntity, error)
}
