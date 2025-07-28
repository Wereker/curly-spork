package models

import (
	"gorm.io/gorm"
)

type Measure struct {
	gorm.Model
	Name     string
	Products []Product
}
