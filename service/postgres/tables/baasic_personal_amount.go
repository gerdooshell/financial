package tables

import (
	bpaPlugin "github.com/gerdooshell/financial/tax/canada/bpa/abstract_plugin"
)

type BasicPersonalAmountModel struct {
	Id              int     `gorm:"->;primaryKey"`
	Year            int     `gorm:"->"`
	RegionCode      string  `gorm:"->"`
	MinAmount       float32 `gorm:"->"`
	MinAmountSalary float32 `gorm:"->"`
	MaxAmount       float32 `gorm:"->"`
	MaxAmountSalary float32 `gorm:"->"`
}

func (b *BasicPersonalAmountModel) TableName() string {
	return "tax.basic_personal_amount"
}

func (b *BasicPersonalAmountModel) ToBpaEntity() *bpaPlugin.BpaEntity {
	return &bpaPlugin.BpaEntity{
		MinAmount:       b.MinAmount,
		MinAmountSalary: b.MinAmountSalary,
		MaxAmount:       b.MaxAmount,
		MaxAmountSalary: b.MaxAmountSalary,
	}
}
