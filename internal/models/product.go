package models

import (
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	Name      string
	Quantity  uint
	UnitCoast uint
	MeasureID uint
	Measure   Measure
}
