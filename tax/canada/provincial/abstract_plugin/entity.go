package abstract_plugin

import "github.com/gerdooshell/financial/tax/canada/canada_region"

type TaxEntity struct {
	Year         int
	ProvinceCode canada_region.Province
	ProvinceName string
	MinSalary    float32
	MaxSalary    float32
	Rate         float32
}
