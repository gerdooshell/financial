package abstract_plugin

import "github.com/gerdooshell/financial/tax/canada/canada_region"

type BpaPluginInputParameters struct {
	Year       int
	RegionCode canada_region.RegionCode
}
