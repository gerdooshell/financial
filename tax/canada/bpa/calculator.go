package bpa

import (
	postgres_service "github.com/gerdooshell/financial/service/postgres"
	"github.com/gerdooshell/financial/tax/canada/bpa/abstract_plugin"
	"github.com/gerdooshell/financial/tax/canada/canada_region"
)

type BasicPersonalAmount interface {
	GetCalculatedBpa([]*InputParameters) (<-chan *OutputParameters, error)
}

func NewBasicPersonalAmount() BasicPersonalAmount {
	return &basicPersonalAmount{}
}

type InputParameters struct {
	Year       int
	RegionCode canada_region.RegionCode
	Salary     float32
}

type OutputParameters struct {
	InputParameters
	MinBPA        float32
	MaxBPA        float32
	CalculatedBPA float32
}

type basicPersonalAmount struct {
	plugin abstract_plugin.BPAService
	out    chan *OutputParameters
}

func (b *basicPersonalAmount) setPlugin() error {
	if b.plugin != nil {
		return nil
	}
	plugin, err := postgres_service.NewPGService()
	if err != nil {
		return err
	}
	b.plugin = plugin
	return nil
}

func (b *basicPersonalAmount) setOutChannel(size int) {
	b.out = make(chan *OutputParameters, size)
}

func (b *basicPersonalAmount) GetCalculatedBpa(params []*InputParameters) (<-chan *OutputParameters, error) {
	if err := b.setPlugin(); err != nil {
		return nil, err
	}
	b.setOutChannel(len(params))
	pluginInputs := mapBpaInputToPluginInputParameters(params)
	entities, err := b.plugin.GetBPAEntities(pluginInputs)
	if err != nil {
		return nil, err
	}
	go b.getBPA(params, entities)
	return b.out, nil
}

func mapBpaInputToPluginInputParameters(params []*InputParameters) []*abstract_plugin.BpaPluginInputParameters {
	pluginInputs := make([]*abstract_plugin.BpaPluginInputParameters, 0, len(params))
	for _, param := range params {
		pluginInputs = append(pluginInputs, &abstract_plugin.BpaPluginInputParameters{
			Year:       param.Year,
			RegionCode: param.RegionCode,
		})
	}
	return pluginInputs
}

func (b *basicPersonalAmount) getBPA(params []*InputParameters, entities <-chan *abstract_plugin.BpaEntity) {
	defer close(b.out)
	var bpa float32 = 0
	for _, param := range params {
		entity := <-entities
		if param.Salary <= entity.MaxAmountSalary {
			bpa = entity.MaxAmount
		} else if param.Salary < entity.MinAmountSalary {
			bpa = calcBPA(param.Salary, entity)
		} else {
			bpa = entity.MinAmount
		}
		b.out <- &OutputParameters{
			InputParameters: *param,
			MaxBPA:          entity.MaxAmount,
			MinBPA:          entity.MinAmount,
			CalculatedBPA:   bpa,
		}
	}
}

func calcBPA(salary float32, entity *abstract_plugin.BpaEntity) float32 {
	diff := (entity.MaxAmount - entity.MinAmount) * (salary - entity.MinAmountSalary) / (entity.MinAmountSalary - entity.MaxAmountSalary)
	if diff < 0 {
		diff = 0
	}
	return entity.MinAmount + diff
}
