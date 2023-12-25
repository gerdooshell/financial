package abstract_plugin

type TaxService interface {
	GetFederalTaxEntities(params []*TaxCalculatorPluginInputParameters) (<-chan []*TaxEntity, error)
}
