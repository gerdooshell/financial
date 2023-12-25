package federal_ei_premium

import (
	postgres_service "github.com/gerdooshell/financial/service/postgres"
	"github.com/gerdooshell/financial/tax/canada/federal_ei_premium/abstract_plugin"
)

type FederalEmploymentInsurance interface {
	CalculateEI(params []*InputParameters) (<-chan *OutputParameters, error)
}

type federalEmploymentInsurance struct {
	plugin abstract_plugin.FedEiService
	out    chan *OutputParameters
}

type InputParameters struct {
	Year       int
	Salary     float32
	IsEmployer bool
}

type OutputParameters struct {
	InputParameters
	CalculatedEI float32
}

func NewFederalEmploymentInsurance() FederalEmploymentInsurance {
	return &federalEmploymentInsurance{}
}

func (f *federalEmploymentInsurance) CalculateEI(params []*InputParameters) (<-chan *OutputParameters, error) {
	if err := f.setPlugin(); err != nil {
		return nil, err
	}
	f.setOutputChan(len(params))
	pluginInputs := mapInputToPluginInputParameters(params)
	eiChan, err := f.plugin.GetEIEntities(pluginInputs)
	if err != nil {
		return nil, err
	}
	go f.calcEI(params, eiChan)

	return f.out, nil
}

func (f *federalEmploymentInsurance) setPlugin() error {
	if f.plugin != nil {
		return nil
	}
	plugin, err := postgres_service.NewPGService()
	if err != nil {
		return err
	}
	f.plugin = plugin
	return nil
}

func (f *federalEmploymentInsurance) setOutputChan(size int) {
	f.out = make(chan *OutputParameters, size)
}

func (f *federalEmploymentInsurance) calcEI(params []*InputParameters, eiChan <-chan *abstract_plugin.FedEiEntity) {
	defer close(f.out)
	var eiAmount float32
	for _, param := range params {
		eiOut := <-eiChan
		rate := eiOut.EmployeeRate
		if param.IsEmployer {
			rate = eiOut.EmployerRate
		}
		if param.Salary > eiOut.MaxInsurableEarning {
			eiAmount = eiOut.MaxInsurableEarning * rate
		} else {
			eiAmount = param.Salary * rate
		}
		f.out <- &OutputParameters{
			InputParameters: *param,
			CalculatedEI:    eiAmount / 100,
		}
	}
}

func mapInputToPluginInputParameters(params []*InputParameters) []*abstract_plugin.FedEiPluginInputParameters {
	pluginInputs := make([]*abstract_plugin.FedEiPluginInputParameters, 0, len(params))
	for _, param := range params {
		pluginInputs = append(pluginInputs, &abstract_plugin.FedEiPluginInputParameters{
			Year: param.Year,
		})
	}
	return pluginInputs
}
