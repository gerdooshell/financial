package tables

import (
	cppPlugin "github.com/gerdooshell/financial/tax/canada/canada_pension_plan/abstract_plugin"
)

type CanadaPensionPlanModel struct {
	Id                              int     `gorm:"->;primaryKey"`
	Year                            int     `gorm:"->"`
	MaximumAnnualPensionableEarning float32 `gorm:"->"`
	BasicExceptionAmount            float32 `gorm:"->"`
	Rate                            float32 `gorm:"->"`
}

func (c *CanadaPensionPlanModel) TableName() string {
	return "tax.canada_pension_plan"
}

func (c *CanadaPensionPlanModel) ToCppEntity() *cppPlugin.CppEntity {
	return &cppPlugin.CppEntity{
		Year:                        c.Year,
		MaxAnnualPensionableEarning: c.MaximumAnnualPensionableEarning,
		BasicExemptionAmount:        c.BasicExceptionAmount,
		Rate:                        c.Rate,
	}
}
