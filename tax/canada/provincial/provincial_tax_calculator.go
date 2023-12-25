package provincial

import (
	postgres_service "github.com/gerdooshell/financial/service/postgres"
	"github.com/gerdooshell/financial/tax/canada/bpa"
	"github.com/gerdooshell/financial/tax/canada/canada_region"
	"github.com/gerdooshell/financial/tax/canada/provincial/abstract_plugin"
	"sort"
)

type SalaryInputParameters struct {
	Year     int
	Salary   float32
	Province canada_region.Province
}

type SalaryOutputParameters struct {
	SalaryInputParameters
	CalculatedProvincialTax float32
	ProvincialTaxRate       float32
	CalculatedBPA           float32
}

type NetIncomeInputParameters struct {
	Year      int
	NetIncome float32
	Province  canada_region.Province
}

type NetIncomeOutputParameters struct {
	NetIncomeInputParameters
	Salary                  float32
	CalculatedProvincialTax float32
	ProvincialTaxRate       float32
	CalculatedBPA           float32
}

type TaxCalculator interface {
	CalcProvinceTax([]*SalaryInputParameters) (<-chan *SalaryOutputParameters, error)
	GetSalaryForNet([]*NetIncomeInputParameters) (<-chan *NetIncomeOutputParameters, error)
}

type taxCalculatorImpl struct {
	basicPersonalAmount bpa.BasicPersonalAmount
	provincialPlugin    abstract_plugin.TaxService
	salaryOut           chan *SalaryOutputParameters
}

func NewTaxCalculator() TaxCalculator {
	return &taxCalculatorImpl{}
}

func (t *taxCalculatorImpl) setBpaCalculator() {
	if t.basicPersonalAmount != nil {
		return
	}
	t.basicPersonalAmount = bpa.NewBasicPersonalAmount()
}

func (t *taxCalculatorImpl) setSalaryOutputChannel(size int) {
	t.salaryOut = make(chan *SalaryOutputParameters, size)
}

func (t *taxCalculatorImpl) setProvincialTaxPlugin() error {
	if t.provincialPlugin != nil {
		return nil
	}
	plugin, err := postgres_service.NewPGService()
	if err != nil {
		return err
	}
	t.provincialPlugin = plugin
	return nil
}

// CalcProvinceTax returns calculated outputs in the same order of inputs
func (t *taxCalculatorImpl) CalcProvinceTax(params []*SalaryInputParameters) (<-chan *SalaryOutputParameters, error) {
	t.setBpaCalculator()
	if err := t.setProvincialTaxPlugin(); err != nil {
		return nil, err
	}
	bpaInputParams := mapInputParamsToBpaInput(params)
	bpaOutCha, err := t.basicPersonalAmount.GetCalculatedBpa(bpaInputParams)
	if err != nil {
		return nil, err
	}
	pluginParams := mapInputParamsToProvincialPluginInputParams(params)
	entityChan, err := t.provincialPlugin.GetProvincialTaxEntities(nil, pluginParams)
	if err != nil {
		return nil, err
	}
	t.setSalaryOutputChannel(len(params))
	go t.calcProvincialTax(bpaOutCha, entityChan, params)
	return t.salaryOut, nil
}

func (t *taxCalculatorImpl) GetSalaryForNet([]*NetIncomeInputParameters) (<-chan *NetIncomeOutputParameters, error) {
	panic("not implemented")
}

func mapInputParamsToBpaInput(params []*SalaryInputParameters) []*bpa.InputParameters {
	bpaInputs := make([]*bpa.InputParameters, 0, len(params))
	for _, param := range params {
		bpaInputs = append(bpaInputs, &bpa.InputParameters{
			Year:       param.Year,
			RegionCode: canada_region.RegionCode(param.Province),
			Salary:     param.Salary,
		})
	}
	return bpaInputs
}

func mapInputParamsToProvincialPluginInputParams(params []*SalaryInputParameters) []*abstract_plugin.TaxCalculatorPluginInputParameters {
	pluginInputs := make([]*abstract_plugin.TaxCalculatorPluginInputParameters, 0, len(params))
	for _, param := range params {
		pluginInputs = append(pluginInputs, &abstract_plugin.TaxCalculatorPluginInputParameters{
			Year:         param.Year,
			Salary:       param.Salary,
			ProvinceCode: param.Province,
		})
	}
	return pluginInputs
}

func (t *taxCalculatorImpl) calcProvincialTax(bpaChan <-chan *bpa.OutputParameters, pluginEntityChan <-chan []*abstract_plugin.TaxEntity, params []*SalaryInputParameters) {
	for _, param := range params {
		entities := <-pluginEntityChan
		bpaOut := <-bpaChan
		taxableSalary := param.Salary - bpaOut.CalculatedBPA
		provincialTax := calculateProvincialTax(entities, taxableSalary)
		t.salaryOut <- &SalaryOutputParameters{
			SalaryInputParameters:   *param,
			CalculatedProvincialTax: provincialTax,
			ProvincialTaxRate:       provincialTax / taxableSalary,
			CalculatedBPA:           bpaOut.CalculatedBPA,
		}
	}
	close(t.salaryOut)
}

func calculateProvincialTax(entities []*abstract_plugin.TaxEntity, taxableSalary float32) float32 {
	sort.SliceStable(entities, func(i, j int) bool { return entities[i].MinSalary <= entities[j].MinSalary })
	var totalTax float32 = 0
	for _, entity := range entities {
		if taxableSalary >= entity.MaxSalary {
			totalTax += (entity.MaxSalary - entity.MinSalary) * entity.Rate
		} else {
			totalTax += (taxableSalary - entity.MinSalary) * entity.Rate
			break
		}
	}
	return totalTax / 100
}
