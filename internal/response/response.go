package response

type ErrorResponse struct {
	Error string `json:"error"`
}

type ProductCreate struct {
	Name      string `json:"name"`
	Quantity  uint   `json:"quantity"`
	UnitCoast uint   `json:"unit_coast"`
	MeasureID uint   `json:"measureID"`
}

type ProductUpdate struct {
	Name      string `json:"name"`
	Quantity  uint   `json:"quantity"`
	UnitCoast uint   `json:"unit_coast"`
	MeasureID uint   `json:"measureID"`
}

type ProductResponse struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Quantity  uint   `json:"quantity"`
	UnitCoast uint   `json:"unit_coast"`
	MeasureID uint   `json:"measureID"`
}

type MeasureCreate struct {
	Name string `json:"name"`
}

type MeasureUpdate struct {
	Name string `json:"name"`
}

type MeasureResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}
