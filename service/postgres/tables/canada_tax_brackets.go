package tables

import (
	"github.com/gerdooshell/financial/tax/canada/canada_region"
	fedPlugin "github.com/gerdooshell/financial/tax/canada/federal/abstract_plugin"
	prPlugin "github.com/gerdooshell/financial/tax/canada/provincial/abstract_plugin"
	tbPlugin "github.com/gerdooshell/financial/tax/canada/rrsp/abstract_tax_brackets_plugin"
)

type CanadaTaxBracketsModel struct {
	Id           int     `gorm:"->;primaryKey"`
	BracketIndex int     `gorm:"->"`
	RegionCode   string  `gorm:"->"`
	RegionName   string  `gorm:"->"`
	Year         int     `gorm:"->"`
	MinSalary    float32 `gorm:"->"`
	MaxSalary    float32 `gorm:"->"`
	Rate         float32 `gorm:"->"`
}

func (ctb *CanadaTaxBracketsModel) TableName() string {
	return "tax.canada_tax_brackets"
}

func (ctb *CanadaTaxBracketsModel) ToFederalTaxEntity() *fedPlugin.TaxEntity {
	return &fedPlugin.TaxEntity{
		Year:              ctb.Year,
		MinSalaryIncluded: ctb.MinSalary,
		MaxSalaryExcluded: ctb.MaxSalary,
		Rate:              ctb.Rate,
	}
}

func (ctb *CanadaTaxBracketsModel) ToProvincialTaxEntity() *prPlugin.TaxEntity {
	return &prPlugin.TaxEntity{
		Year:         ctb.Year,
		ProvinceCode: canada_region.Province(ctb.RegionCode),
		ProvinceName: ctb.RegionName,
		MinSalary:    ctb.MinSalary,
		MaxSalary:    ctb.MaxSalary,
		Rate:         ctb.Rate,
	}
}

func (ctb *CanadaTaxBracketsModel) ToTaxBracketEntity() *tbPlugin.TaxBracketEntity {
	return &tbPlugin.TaxBracketEntity{
		Index:     ctb.BracketIndex,
		MinSalary: ctb.MinSalary,
		MaxSalary: ctb.MaxSalary,
		Rate:      ctb.Rate,
	}
}
