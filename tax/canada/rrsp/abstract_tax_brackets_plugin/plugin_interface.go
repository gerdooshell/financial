package abstract_tax_brackets_plugin

type TaxBracketsService interface {
	GetFedTaxBrackets([]*FedTaxBracketsInputParams) (<-chan *FedTaxBracketOutputParams, error)
	GetPrTaxBrackets([]*PrTaxBracketsInputParams) (<-chan *PrTaxBracketOutputParams, error)
}
