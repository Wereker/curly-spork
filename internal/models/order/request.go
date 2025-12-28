package models

import (
	models "app/internal/models/user"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Request struct {
	ID             uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Number         string `json:"number" gorm:"uniqueIndex;not null"`
	UserComment    string `json:"user_comment,omitempty" gorm:"size:128"`
	ManagerComment string `json:"manager_comment,omitempty" gorm:"size:128"`

	OrderID         uint           `json:"order_id"`
	Order           *Order         `json:"order" gorm:"foreignKey:OrderID"`
	UserID          uint           `json:"user_id"`
	User            *models.User   `json:"user" gorm:"foreignKey:UserID"`
	RequestStatusID uint           `json:"request_status_id"`
	RequestStatus   *RequestStatus `json:"request_status" gorm:"foreignKey:RequestStatusID"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Request) TableName() string {
	return "requests"
}

func (r *Request) BeforeCreate(tx *gorm.DB) error {
	if r.Number == "" {
		r.Number = generateRequestNumber()
	}
	return r.ensureUniqueRequestNumber(tx)
}

func generateRequestNumber() string {
	id := uuid.New()
	uuidStr := strings.ReplaceAll(id.String(), "-", "")
	if len(uuidStr) > 20 {
		uuidStr = uuidStr[:20]
	}
	return fmt.Sprintf("R-%s", strings.ToUpper(uuidStr))
}

func (r *Request) ensureUniqueRequestNumber(tx *gorm.DB) error {
	baseNumber := r.Number
	counter := 1

	for {
		var count int64
		query := tx.Model(&Request{}).Where("number = ?", r.Number)

		if r.ID != 0 {
			query = query.Where("id != ?", r.ID)
		}

		query.Count(&count)

		if count == 0 {
			break
		}

		r.Number = fmt.Sprintf("%s-%d", baseNumber, counter)
		counter++

		if counter > 100 {
			return fmt.Errorf("failed to generate unique request number after 100 attempts")
		}
	}

	return nil
}
