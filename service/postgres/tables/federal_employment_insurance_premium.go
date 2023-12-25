package tables

import "github.com/gerdooshell/financial/tax/canada/federal_ei_premium/abstract_plugin"

type FederalEmploymentInsuranceModel struct {
	Id                  int     `gorm:"->;primaryKey"`
	Year                int     `gorm:"->"`
	MaxInsurableEarning float32 `gorm:"->"`
	EmployeeRate        float32 `gorm:"->"`
	EmployerRate        float32 `gorm:"->"`
}

func (ei *FederalEmploymentInsuranceModel) TableName() string {
	return "tax.federal_employment_insurance_premium"
}

func (ei *FederalEmploymentInsuranceModel) ToEiEntity() *abstract_plugin.FedEiEntity {
	return &abstract_plugin.FedEiEntity{
		MaxInsurableEarning: ei.MaxInsurableEarning,
		EmployeeRate:        ei.EmployeeRate,
		EmployerRate:        ei.EmployerRate,
	}
}
