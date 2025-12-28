package models

import (
	"time"

	"gorm.io/gorm"
)

type Right struct {
	ID               uint    `json:"id" gorm:"primaryKey;autoIncrement"`
	Title            string  `json:"title" gorm:"size:50;not null"`
	Description      *string `json:"description,omitempty" gorm:"type:text"`
	CanViewCatalog   bool    `json:"can_view_catalog" gorm:"default:false"`
	CanMakeOrder     bool    `json:"can_make_order" gorm:"default:false"`
	CanManageOrder   bool    `json:"can_manage_order" gorm:"default:false"`
	CanManageProduct bool    `json:"can_manage_product" gorm:"default:false"`
	CanManageUser    bool    `json:"can_manage_user" gorm:"default:false"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Right) TableName() string {
	return "rights"
}
