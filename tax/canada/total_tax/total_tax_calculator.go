package total_tax

import (
	"github.com/gerdooshell/financial/tax/canada/canada_pension_plan"
	"github.com/gerdooshell/financial/tax/canada/canada_region"
	"github.com/gerdooshell/financial/tax/canada/federal"
	"github.com/gerdooshell/financial/tax/canada/federal_ei_premium"
	"github.com/gerdooshell/financial/tax/canada/provincial"
)

type totalTaxCalculator struct {
	ei  federal_ei_premium.FederalEmploymentInsurance
	cpp canada_pension_plan.CanadaPensionPlan
	pr  provincial.TaxCalculator
	fed federal.TaxCalculator
	out chan *OutputParameters
}

type TaxCalculator interface {
	CalculateTotalTax([]*InputParameters) (<-chan *OutputParameters, error)
}

func NewTotalTaxCalculator() TaxCalculator {
	return &totalTaxCalculator{
		ei:  federal_ei_premium.NewFederalEmploymentInsurance(),
		cpp: canada_pension_plan.NewCanadaPensionPlan(),
		pr:  provincial.NewTaxCalculator(),
		fed: federal.NewTaxCalculator(),
	}
}

type InputParameters struct {
	Year       int
	Salary     float32
	Province   canada_region.Province
	IsEmployer bool
}

type OutputParameters struct {
	InputParameters
	EI                float32
	CPP               float32
	ProvincialTax     float32
	ProvincialTaxRate float32
	ProvincialBPA     float32
	FederalTax        float32
	FederalTaxRate    float32
	FederalBPA        float32
	TotalTax          float32
	TaxRate           float32
	NetIncome         float32
}

func (t *totalTaxCalculator) CalculateTotalTax(params []*InputParameters) (<-chan *OutputParameters, error) {
	t.setOutputChannel(len(params))
	eiInputs := mapInputToEiInputParameters(params)
	eiChan, err := t.ei.CalculateEI(eiInputs)
	if err != nil {
		return nil, err
	}
	cppInputs := mapInputToCppInputParameters(params)
	cppChan, err := t.cpp.GetCppForSalary(cppInputs)
	if err != nil {
		return nil, err
	}
	prInputs := mapInputToProvincialInputParameters(params)
	prChan, err := t.pr.CalcProvinceTax(prInputs)
	if err != nil {
		return nil, err
	}
	fedInputs := mapInputToFederalInputParameters(params)
	fedChan, err := t.fed.GetCalculatedFederalTax(fedInputs)
	if err != nil {
		return nil, err
	}
	go func() {
		defer close(t.out)
		for _, param := range params {
			eiOut := <-eiChan
			cppOut := <-cppChan
			prOut := <-prChan
			fedOut := <-fedChan
			totalTax := cppOut.CalculatedContribution + prOut.CalculatedProvincialTax + fedOut.CalculatedFederalTax + eiOut.CalculatedEI
			netPay := param.Salary - totalTax
			t.out <- &OutputParameters{
				InputParameters:   *param,
				EI:                eiOut.CalculatedEI,
				CPP:               cppOut.CalculatedContribution,
				ProvincialTax:     prOut.CalculatedProvincialTax,
				ProvincialTaxRate: prOut.ProvincialTaxRate * 100,
				ProvincialBPA:     prOut.CalculatedBPA,
				FederalTax:        fedOut.CalculatedFederalTax,
				FederalTaxRate:    fedOut.FederalTaxRate * 100,
				FederalBPA:        fedOut.FederalBpa,
				TotalTax:          totalTax,
				TaxRate:           totalTax / param.Salary * 100,
				NetIncome:         netPay,
			}
		}

	}()
	return t.out, nil
}

func (t *totalTaxCalculator) setOutputChannel(size int) {
	t.out = make(chan *OutputParameters, size)
}

func mapInputToEiInputParameters(params []*InputParameters) []*federal_ei_premium.InputParameters {
	eiInputs := make([]*federal_ei_premium.InputParameters, 0, len(params))
	for _, param := range params {
		eiInputs = append(eiInputs, &federal_ei_premium.InputParameters{
			Year:       param.Year,
			Salary:     param.Salary,
			IsEmployer: param.IsEmployer,
		})
	}
	return eiInputs
}

func mapInputToCppInputParameters(params []*InputParameters) []*canada_pension_plan.InputParameters {
	cppInputs := make([]*canada_pension_plan.InputParameters, 0, len(params))
	for _, param := range params {
		cppInputs = append(cppInputs, &canada_pension_plan.InputParameters{
			Year:   param.Year,
			Salary: param.Salary,
		})
	}
	return cppInputs
}

func mapInputToProvincialInputParameters(params []*InputParameters) []*provincial.SalaryInputParameters {
	prInputs := make([]*provincial.SalaryInputParameters, 0, len(params))
	for _, param := range params {
		prInputs = append(prInputs, &provincial.SalaryInputParameters{
			Year:     param.Year,
			Salary:   param.Salary,
			Province: param.Province,
		})
	}
	return prInputs
}

func mapInputToFederalInputParameters(params []*InputParameters) []*federal.InputParameters {
	fedInputs := make([]*federal.InputParameters, 0, len(params))
	for _, param := range params {
		fedInputs = append(fedInputs, &federal.InputParameters{
			Year:     param.Year,
			Salary:   param.Salary,
			Province: param.Province,
		})
	}
	return fedInputs
}
