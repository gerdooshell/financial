package rrsp

import (
	postgres_service "github.com/gerdooshell/financial/service/postgres"
	"github.com/gerdooshell/financial/tax/canada/canada_region"
	tbplugin "github.com/gerdooshell/financial/tax/canada/rrsp/abstract_tax_brackets_plugin"
	"github.com/gerdooshell/financial/tax/canada/total_tax"
	"sort"
)

type RRSavingPlan interface {
	GetOptimalReturns(params []*OptimalReturnsInputParameters) (<-chan []*OptimalReturnsOutputParameters, error)
}

type rrSavingPlan struct {
	taxCalculator     total_tax.TaxCalculator
	taxBracketsPlugin tbplugin.TaxBracketsService
	optimalReturnsOut chan []*OptimalReturnsOutputParameters
}

type OptimalReturnsInputParameters struct {
	Year       int
	Province   canada_region.Province
	Salary     float32
	IsEmployer bool
}

type OptimalReturnsOutputParameters struct {
	OptimalReturnsInputParameters
	NetRemained  float32
	Contribution float32
	TaxOnSalary  float32
	TaxOnEdge    float32
	TaxReturn    float32
	EdgeSalary   float32
}

func NewRRetirementSavingPlan() RRSavingPlan {
	return &rrSavingPlan{
		taxCalculator: total_tax.NewTotalTaxCalculator(),
	}
}

func (r *rrSavingPlan) setBracketsPlugin() error {
	if r.taxBracketsPlugin != nil {
		return nil
	}
	// TODO: set plugin
	plugin, err := postgres_service.NewPGService()
	if err != nil {
		return err
	}
	r.taxBracketsPlugin = plugin
	return nil
}

func (r *rrSavingPlan) GetOptimalReturns(params []*OptimalReturnsInputParameters) (<-chan []*OptimalReturnsOutputParameters, error) {
	if err := r.setBracketsPlugin(); err != nil {
		return nil, err
	}
	prBracketInputs := mapOptimalReturnInputToPrTaxBracketsInputParams(params)
	fedBracketInputs := mapOptimalReturnInputToFedTaxBracketsInputParams(params)
	prBracketsChan, err := r.taxBracketsPlugin.GetPrTaxBrackets(prBracketInputs)
	if err != nil {
		return nil, err
	}
	fedBracketsChan, err := r.taxBracketsPlugin.GetFedTaxBrackets(fedBracketInputs)
	if err != nil {
		return nil, err
	}
	taxCalcInputParams := mapOptimalReturnInputToTaxCalculatorInputParams(params)
	salaryTaxChan, err := r.taxCalculator.CalculateTotalTax(taxCalcInputParams)
	if err != nil {
		return nil, err
	}
	mergedBracketsChan := getMergedBrackets(params, fedBracketsChan, prBracketsChan)
	r.setOptimalReturnOutputChannel()
	go func() {
		defer close(r.optimalReturnsOut)
		for _, param := range params {
			mergedBrackets := <-mergedBracketsChan
			taxEdgeInputs := buildTaxCalcInputsForEdges(mergedBrackets)
			edgeTaxCalculator := total_tax.NewTotalTaxCalculator()
			edgeTaxChan, err := edgeTaxCalculator.CalculateTotalTax(taxEdgeInputs)
			taxSalary := <-salaryTaxChan
			if err != nil {
				r.optimalReturnsOut <- nil
			}
			outputs := make([]*OptimalReturnsOutputParameters, 0, len(mergedBrackets.Edges))
			for edgeTax := range edgeTaxChan {
				taxReturn := taxSalary.TotalTax - edgeTax.TotalTax
				outputs = append(outputs, &OptimalReturnsOutputParameters{
					OptimalReturnsInputParameters: *param,
					NetRemained:                   edgeTax.NetIncome + taxReturn,
					Contribution:                  taxSalary.NetIncome - edgeTax.NetIncome,
					TaxOnSalary:                   taxSalary.TotalTax,
					TaxOnEdge:                     edgeTax.TotalTax,
					TaxReturn:                     taxReturn,
					EdgeSalary:                    edgeTax.Salary,
				})
			}
			r.optimalReturnsOut <- outputs
		}
	}()
	return r.optimalReturnsOut, nil
}

func (r *rrSavingPlan) setOptimalReturnOutputChannel() {
	r.optimalReturnsOut = make(chan []*OptimalReturnsOutputParameters)
}

func buildTaxCalcInputsForEdges(mergedBrackets *MergeTaxBracketModel) []*total_tax.InputParameters {
	taxInputs := make([]*total_tax.InputParameters, 0, len(mergedBrackets.Edges))
	for _, edge := range mergedBrackets.Edges {
		taxInputs = append(taxInputs, &total_tax.InputParameters{
			Year:       mergedBrackets.Year,
			Salary:     edge - 0.01,
			Province:   mergedBrackets.Province,
			IsEmployer: mergedBrackets.IsEmployer,
		})
	}
	return taxInputs
}

func mapOptimalReturnInputToTaxCalculatorInputParams(params []*OptimalReturnsInputParameters) []*total_tax.InputParameters {
	taxInputs := make([]*total_tax.InputParameters, 0, len(params))
	for _, param := range params {
		taxInputs = append(taxInputs, &total_tax.InputParameters{
			Year:       param.Year,
			Salary:     param.Salary,
			Province:   param.Province,
			IsEmployer: param.IsEmployer,
		})
	}
	return taxInputs
}

func getMergedBrackets(params []*OptimalReturnsInputParameters, fedBracketsChan <-chan *tbplugin.FedTaxBracketOutputParams, prBracketsChan <-chan *tbplugin.PrTaxBracketOutputParams) <-chan *MergeTaxBracketModel {
	mergeChan := make(chan *MergeTaxBracketModel, len(params))

	go func() {
		defer close(mergeChan)
		for _, param := range params {
			fedBrackets := <-fedBracketsChan
			prBrackets := <-prBracketsChan
			edges := getSortedMergedEdges(fedBrackets, prBrackets, param.Salary)
			mergeChan <- &MergeTaxBracketModel{
				Year:       fedBrackets.Year,
				Region:     fedBrackets.Region,
				Province:   prBrackets.Province,
				Salary:     param.Salary,
				Edges:      edges,
				IsEmployer: param.IsEmployer,
			}
		}
	}()
	return mergeChan
}

type MergeTaxBracketModel struct {
	Year       int
	Region     canada_region.RegionCode
	Province   canada_region.Province
	Salary     float32
	Edges      []float32
	IsEmployer bool
}

func getSortedMergedEdges(fedBrackets *tbplugin.FedTaxBracketOutputParams, prBrackets *tbplugin.PrTaxBracketOutputParams, salary float32) []float32 {
	unOrderedEdges := make([]float32, 0, len(fedBrackets.Entities)+len(prBrackets.Entities))
	for _, fedEntity := range fedBrackets.Entities {
		if fedEntity.MinSalary == 0 {
			continue
		}
		if fedEntity.MinSalary > salary {
			break
		}
		unOrderedEdges = append(unOrderedEdges, fedEntity.MinSalary)
	}
	for _, prEntity := range prBrackets.Entities {
		if prEntity.MinSalary == 0 {
			continue
		}
		if prEntity.MinSalary > salary {
			break
		}
		unOrderedEdges = append(unOrderedEdges, prEntity.MinSalary)
	}
	sort.SliceStable(unOrderedEdges, func(i, j int) bool { return unOrderedEdges[i] > unOrderedEdges[j] })
	return unOrderedEdges
}

func mapOptimalReturnInputToPrTaxBracketsInputParams(params []*OptimalReturnsInputParameters) []*tbplugin.PrTaxBracketsInputParams {
	prTaxBracketsParams := make([]*tbplugin.PrTaxBracketsInputParams, 0, len(params))
	for _, param := range params {
		prTaxBracketsParams = append(prTaxBracketsParams, &tbplugin.PrTaxBracketsInputParams{
			Year:     param.Year,
			Province: param.Province,
		})
	}
	return prTaxBracketsParams
}

func mapOptimalReturnInputToFedTaxBracketsInputParams(params []*OptimalReturnsInputParameters) []*tbplugin.FedTaxBracketsInputParams {
	fedTaxBracketsParams := make([]*tbplugin.FedTaxBracketsInputParams, 0, len(params))
	for _, param := range params {
		fedTaxBracketsParams = append(fedTaxBracketsParams, &tbplugin.FedTaxBracketsInputParams{
			Year:   param.Year,
			Region: canada_region.Federal,
		})
	}
	return fedTaxBracketsParams
}

type CalculatorInputParameters struct {
	Year         int
	Salary       float32
	Contribution float32
}

type CalculatorOutputParameters struct {
	CalculatorInputParameters
	TaxOnSalary float32
}

func (r *rrSavingPlan) CalculateRRSP(params []*CalculatorInputParameters) (<-chan *CalculatorOutputParameters, error) {
	panic("not implemented")
}
