package canada_region

type RegionCode string
type Province RegionCode

const (
	Federal         RegionCode = "ca"
	BritishColumbia Province   = "bc"
)
