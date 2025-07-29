package models

type Product struct {
	ID        uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string `json:"name"`
	Quantity  uint   `json:"quantity"`
	UnitCoast uint   `json:"unit_coast"`
	MeasureID uint   `json:"measureID" gorm:"default:1"`
}
