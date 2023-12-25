package canada_pension_plan

import (
	postgres_service "github.com/gerdooshell/financial/service/postgres"
	"github.com/gerdooshell/financial/tax/canada/canada_pension_plan/abstract_plugin"
)

type InputParameters struct {
	Year   int
	Salary float32
}

type OutputParameters struct {
	MaxAnnualPensionableEarning       float32
	BasicExemptionAmount              float32
	MaxContributoryEarning            float32
	Rate                              float32
	MaxAnnualContribution             float32
	MaxAnnualContributionSelfEmployed float32
	CalculatedContribution            float32
	InputParameters
}

type CanadaPensionPlan interface {
	// GetCppForSalary returns calculated outputs in the same order of inputs
	GetCppForSalary(params []*InputParameters) (<-chan *OutputParameters, error)
}

func NewCanadaPensionPlan() CanadaPensionPlan {
	return &canadaPensionPlanImpl{}
}

type canadaPensionPlanImpl struct {
	cppPlugin abstract_plugin.CppService
	out       chan *OutputParameters
}

func (c *canadaPensionPlanImpl) GetCppForSalary(params []*InputParameters) (<-chan *OutputParameters, error) {
	if err := c.setCppPlugin(); err != nil {
		return nil, err
	}

	pluginInputs := mapInputParamsToPluginInputParams(params)
	cppEntities, err := c.cppPlugin.GetCppEntities(pluginInputs)
	if err != nil {
		return nil, err
	}
	c.setOutputChannel(len(params))
	go c.calcCpp(cppEntities, params)
	return c.out, nil
}

func (c *canadaPensionPlanImpl) setOutputChannel(size int) {
	c.out = make(chan *OutputParameters, size)
}

func (c *canadaPensionPlanImpl) setCppPlugin() error {
	if c.cppPlugin != nil {
		return nil
	}
	plugin, err := postgres_service.NewPGService()
	if err != nil {
		return err
	}
	c.cppPlugin = plugin
	return nil
}

func mapInputParamsToPluginInputParams(params []*InputParameters) []*abstract_plugin.CppPluginInputParameters {
	pluginInputs := make([]*abstract_plugin.CppPluginInputParameters, 0, len(params))
	for _, param := range params {
		pluginInputs = append(pluginInputs, &abstract_plugin.CppPluginInputParameters{
			Year: param.Year,
		})
	}
	return pluginInputs
}

func (c *canadaPensionPlanImpl) calcCpp(cppEntities <-chan *abstract_plugin.CppEntity, params []*InputParameters) {
	index := 0
	for entity := range cppEntities {
		param := params[index]
		salary := param.Salary
		maxAnnualContribution := calcMaxAnnualContribution(entity)
		c.out <- &OutputParameters{
			InputParameters:                   *param,
			Rate:                              entity.Rate,
			BasicExemptionAmount:              entity.BasicExemptionAmount,
			MaxAnnualPensionableEarning:       entity.MaxAnnualPensionableEarning,
			MaxContributoryEarning:            entity.MaxAnnualPensionableEarning - entity.BasicExemptionAmount,
			MaxAnnualContribution:             maxAnnualContribution,
			MaxAnnualContributionSelfEmployed: maxAnnualContribution * 2,
			CalculatedContribution:            calcCppContributionForSalary(salary, entity),
		}
		index++
	}
	close(c.out)
}

func calcMaxAnnualContribution(entity *abstract_plugin.CppEntity) float32 {
	return (entity.MaxAnnualPensionableEarning - entity.BasicExemptionAmount) * entity.Rate / 100
}

func calcCppContributionForSalary(salary float32, entity *abstract_plugin.CppEntity) float32 {
	if salary > entity.MaxAnnualPensionableEarning {
		return calcMaxAnnualContribution(entity)
	}
	if salary < entity.BasicExemptionAmount {
		return 0
	}
	return (salary - entity.BasicExemptionAmount) * entity.Rate / 100
}
