package postgres

import (
	"context"
	bpaService "github.com/gerdooshell/financial/tax/canada/bpa/abstract_plugin"
	cppPlugin "github.com/gerdooshell/financial/tax/canada/canada_pension_plan/abstract_plugin"
	"github.com/gerdooshell/financial/tax/canada/canada_region"
	fedPlugin "github.com/gerdooshell/financial/tax/canada/federal/abstract_plugin"
	eiPlugin "github.com/gerdooshell/financial/tax/canada/federal_ei_premium/abstract_plugin"
	prPlugin "github.com/gerdooshell/financial/tax/canada/provincial/abstract_plugin"
)

type pgServiceMockImpl struct {
}

type PGServiceMock interface {
	prPlugin.TaxService
	fedPlugin.TaxService
	cppPlugin.CppService
	bpaService.BPAService
	eiPlugin.FedEiService
}

func NewPGServiceMock() (PGServiceMock, error) {
	return &pgServiceMockImpl{}, nil
}

func (p pgServiceMockImpl) GetProvincialTaxEntities(ctx context.Context, params []*prPlugin.TaxCalculatorPluginInputParameters) (<-chan []*prPlugin.TaxEntity, error) {
	o := make(chan []*prPlugin.TaxEntity, len(params))
	for range params {
		o <- getPr2023Mock()
	}
	close(o)
	return o, nil
}

func (p pgServiceMockImpl) GetFederalTaxEntities(params []*fedPlugin.TaxCalculatorPluginInputParameters) (<-chan []*fedPlugin.TaxEntity, error) {
	o := make(chan []*fedPlugin.TaxEntity, len(params))
	for range params {
		o <- getFed2023Mock()
	}
	close(o)
	return o, nil
}

func (p pgServiceMockImpl) GetCppEntities(params []*cppPlugin.CppPluginInputParameters) (<-chan *cppPlugin.CppEntity, error) {
	o := make(chan *cppPlugin.CppEntity, len(params))
	for range params {
		o <- getCpp2023Mock()

	}
	close(o)
	return o, nil
}

func (p pgServiceMockImpl) GetBPAEntities(params []*bpaService.BpaPluginInputParameters) (<-chan *bpaService.BpaEntity, error) {
	o := make(chan *bpaService.BpaEntity)
	go func() {
		defer close(o)
		for _, param := range params {
			switch param.RegionCode {
			case canada_region.RegionCode(canada_region.BritishColumbia):
				o <- getBCBpa2023Mock()
			case canada_region.Federal:
				fallthrough
			default:
				o <- getFederalBpa2023Mock()
			}
		}
	}()
	return o, nil
}

func (p pgServiceMockImpl) GetEIEntities(parameters []*eiPlugin.FedEiPluginInputParameters) (<-chan *eiPlugin.FedEiEntity, error) {
	//TODO implement me
	panic("implement me")
}

func getBCBpa2023Mock() *bpaService.BpaEntity {
	return &bpaService.BpaEntity{
		MinAmount:       11981,
		MaxAmount:       11981,
		MinAmountSalary: 2000000,
		MaxAmountSalary: 2000000,
	}
}

func getFederalBpa2023Mock() *bpaService.BpaEntity {
	return &bpaService.BpaEntity{
		MinAmount:       13521,
		MaxAmount:       15000,
		MinAmountSalary: 235675,
		MaxAmountSalary: 165430,
	}
}

func getCpp2023Mock() *cppPlugin.CppEntity {
	return &cppPlugin.CppEntity{Year: 2023, MaxAnnualPensionableEarning: 66600, BasicExemptionAmount: 3500, Rate: 5.95}
}

func getPr2023Mock() []*prPlugin.TaxEntity {
	return []*prPlugin.TaxEntity{
		{
			Year:         2023,
			ProvinceCode: canada_region.BritishColumbia,
			ProvinceName: "british columbia",
			MinSalary:    0,
			MaxSalary:    45654,
			Rate:         5.06,
		},
		{
			Year:         2023,
			ProvinceCode: canada_region.BritishColumbia,
			ProvinceName: "british columbia",
			MinSalary:    45654,
			MaxSalary:    91310,
			Rate:         7.7,
		},
		{
			Year:         2023,
			ProvinceCode: canada_region.BritishColumbia,
			ProvinceName: "british columbia",
			MinSalary:    91310,
			MaxSalary:    104835,
			Rate:         10.5,
		},
		{
			Year:         2023,
			ProvinceCode: canada_region.BritishColumbia,
			ProvinceName: "british columbia",
			MinSalary:    104835,
			MaxSalary:    127299,
			Rate:         12.29,
		},
		{
			Year:         2023,
			ProvinceCode: canada_region.BritishColumbia,
			ProvinceName: "british columbia",
			MinSalary:    127299,
			MaxSalary:    172602,
			Rate:         14.7,
		},
		{
			Year:         2023,
			ProvinceCode: canada_region.BritishColumbia,
			ProvinceName: "british columbia",
			MinSalary:    172602,
			MaxSalary:    240716,
			Rate:         16.8,
		},
		{
			Year:         2023,
			ProvinceCode: canada_region.BritishColumbia,
			ProvinceName: "british columbia",
			MinSalary:    240716,
			MaxSalary:    1000000000,
			Rate:         20.5,
		},
	}
}

func getFed2023Mock() []*fedPlugin.TaxEntity {
	return []*fedPlugin.TaxEntity{
		{Year: 2023, MinSalaryIncluded: 0, MaxSalaryExcluded: 53359, Rate: 15},
		{Year: 2023, MinSalaryIncluded: 53359, MaxSalaryExcluded: 106717, Rate: 20.5},
		{Year: 2023, MinSalaryIncluded: 106717, MaxSalaryExcluded: 165430, Rate: 26},
		{Year: 2023, MinSalaryIncluded: 165430, MaxSalaryExcluded: 235675, Rate: 29},
		{Year: 2023, MinSalaryIncluded: 235675, MaxSalaryExcluded: 1000000000, Rate: 33},
	}
}
