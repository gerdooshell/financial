package abstract_tax_brackets_plugin

type FedTaxBracketOutputParams struct {
	FedTaxBracketsInputParams
	Entities []*TaxBracketEntity
}

type PrTaxBracketOutputParams struct {
	PrTaxBracketsInputParams
	Entities []*TaxBracketEntity
}

type TaxBracketEntity struct {
	Index     int
	MinSalary float32
	MaxSalary float32
	Rate      float32
}
