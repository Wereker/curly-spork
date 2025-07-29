package models

type Measure struct {
	ID   uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name string `json:"name"`
}
