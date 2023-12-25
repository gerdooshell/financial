package federal

import (
	postgres_service "github.com/gerdooshell/financial/service/postgres"
	"github.com/gerdooshell/financial/tax/canada/bpa"
	"github.com/gerdooshell/financial/tax/canada/canada_region"
	"github.com/gerdooshell/financial/tax/canada/federal/abstract_plugin"
)

type TaxCalculator interface {
	GetCalculatedFederalTax(params []*InputParameters) (<-chan *OutputParameters, error)
}

func NewTaxCalculator() TaxCalculator {
	return &taxCalculatorImpl{}
}

type InputParameters struct {
	Year     int
	Salary   float32
	Province canada_region.Province
}

type OutputParameters struct {
	InputParameters
	CalculatedFederalTax float32
	FederalTaxRate       float32
	FederalBpa           float32
}

type taxCalculatorImpl struct {
	basicPersonalAmount bpa.BasicPersonalAmount
	federalPlugin       abstract_plugin.TaxService
	out                 chan *OutputParameters
}

func (t *taxCalculatorImpl) GetCalculatedFederalTax(params []*InputParameters) (<-chan *OutputParameters, error) {
	t.setBpaCalculator()
	if err := t.setFederalPlugin(); err != nil {
		return nil, err
	}
	bpaInputParams := mapInputParamsToBpaInput(params)
	bpaOutCha, err := t.basicPersonalAmount.GetCalculatedBpa(bpaInputParams)
	if err != nil {
		return nil, err
	}
	mapInputParams := mapInputParamsToFederalPluginInputParams(params)
	entityChan, err := t.federalPlugin.GetFederalTaxEntities(mapInputParams)
	if err != nil {
		return nil, err
	}
	t.setOutputChannel(len(params))
	go t.calcFederalTax(params, bpaOutCha, entityChan)
	return t.out, nil
}

func (t *taxCalculatorImpl) setBpaCalculator() {
	if t.basicPersonalAmount != nil {
		return
	}
	t.basicPersonalAmount = bpa.NewBasicPersonalAmount()
}

func (t *taxCalculatorImpl) setOutputChannel(size int) {
	t.out = make(chan *OutputParameters, size)
}

func (t *taxCalculatorImpl) setFederalPlugin() error {
	if t.federalPlugin != nil {
		return nil
	}
	plugin, err := postgres_service.NewPGService()
	if err != nil {
		return err
	}
	t.federalPlugin = plugin
	return nil
}

func (t *taxCalculatorImpl) calcFederalTax(params []*InputParameters, bpaChan <-chan *bpa.OutputParameters, entityChan <-chan []*abstract_plugin.TaxEntity) {
	for _, param := range params {
		entity := <-entityChan
		bpaOut := <-bpaChan
		remainingSalary := param.Salary - bpaOut.CalculatedBPA
		federalTax := calcFederalTax(entity, remainingSalary)
		t.out <- &OutputParameters{
			InputParameters:      *param,
			CalculatedFederalTax: federalTax,
			FederalTaxRate:       federalTax / remainingSalary,
			FederalBpa:           bpaOut.CalculatedBPA,
		}
	}
	close(t.out)
}

func mapInputParamsToBpaInput(params []*InputParameters) []*bpa.InputParameters {
	bpaInputs := make([]*bpa.InputParameters, 0, len(params))
	for _, param := range params {
		bpaInputs = append(bpaInputs, &bpa.InputParameters{
			Year:       param.Year,
			RegionCode: canada_region.Federal,
			Salary:     param.Salary,
		})
	}
	return bpaInputs
}

func mapInputParamsToFederalPluginInputParams(params []*InputParameters) []*abstract_plugin.TaxCalculatorPluginInputParameters {
	pluginParams := make([]*abstract_plugin.TaxCalculatorPluginInputParameters, 0, len(params))
	for _, param := range params {
		pluginParams = append(pluginParams, &abstract_plugin.TaxCalculatorPluginInputParameters{
			Year:   param.Year,
			Salary: param.Salary,
		})
	}
	return pluginParams
}

func calcFederalTax(entities []*abstract_plugin.TaxEntity, remainedSalary float32) float32 {
	var federalTax float32 = 0
	for _, entity := range entities {
		if remainedSalary > entity.MaxSalaryExcluded {
			federalTax += (entity.MaxSalaryExcluded - entity.MinSalaryIncluded) * entity.Rate
			continue
		} else if remainedSalary >= entity.MinSalaryIncluded {
			federalTax += (remainedSalary - entity.MinSalaryIncluded) * entity.Rate
			break
		}
	}
	return federalTax / 100
}
