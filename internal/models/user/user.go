package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID               uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Username         string    `json:"username" gorm:"size:150;uniqueIndex"`
	Email            string    `json:"email" gorm:"uniqueIndex;size:100"`
	Password         string    `json:"-"`
	FirstName        *string   `json:"first_name,omitempty" gorm:"size:100"`
	LastName         *string   `json:"last_name,omitempty" gorm:"size:100"`
	FatherName       *string   `json:"father_name,omitempty" gorm:"size:100"`
	Phone            *string   `json:"phone,omitempty" gorm:"size:100"`
	Address          *string   `json:"address,omitempty"`
	RegistrationDate time.Time `json:"registration_date" gorm:"autoCreateTime"`

	RightID uint   `json:"right_id"`
	Right   *Right `json:"right" gorm:"foreignKey:RightID"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	return u.HashPassword()
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {
	if tx.Statement.Changed("Password") {
		return u.HashPassword()
	}
	return nil
}

func (u *User) HashPassword() error {
	if u.Password == "" {
		return nil
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.Password = string(bytes)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
