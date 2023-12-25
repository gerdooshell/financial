package abstract_tax_brackets_plugin

import "github.com/gerdooshell/financial/tax/canada/canada_region"

type FedTaxBracketsInputParams struct {
	Year   int
	Region canada_region.RegionCode
}

type PrTaxBracketsInputParams struct {
	Year     int
	Province canada_region.Province
}
