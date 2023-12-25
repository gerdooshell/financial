package abstract_plugin

import "github.com/gerdooshell/financial/tax/canada/canada_region"

type TaxCalculatorPluginInputParameters struct {
	Year         int
	Salary       float32
	ProvinceCode canada_region.Province
}
