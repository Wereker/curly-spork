package response

import "time"

type ShipmentResponse struct {
	ID               uint                    `json:"id"`
	Number           string                  `json:"number"`
	Price            float64                 `json:"price"`
	Address          string                  `json:"address,omitempty"`
	ShipmentMethodID uint                    `json:"shipment_method_id"`
	ShipmentMethod   *ShipmentMethodResponse `json:"shipment_method,omitempty"`
	ShipmentStatusID uint                    `json:"shipment_status_id"`
	ShipmentStatus   *ShipmentStatusResponse `json:"shipment_status,omitempty"`
	CreatedAt        time.Time               `json:"created_at"`
	UpdatedAt        time.Time               `json:"updated_at"`
}

type ShipmentCreate struct {
	Price            float64 `json:"price" validate:"gte=0"`
	Address          string  `json:"address,omitempty"`
	ShipmentMethodID uint    `json:"shipment_method_id" validate:"required"`
	ShipmentStatusID uint    `json:"shipment_status_id" validate:"required"`
}

type ShipmentUpdate struct {
	Price            float64 `json:"price" validate:"gte=0"`
	Address          string  `json:"address,omitempty"`
	ShipmentMethodID uint    `json:"shipment_method_id" validate:"required"`
	ShipmentStatusID uint    `json:"shipment_status_id" validate:"required"`
}

type ShipmentPatch struct {
	Price            *float64 `json:"price,omitempty" validate:"omitempty,gte=0"`
	Address          *string  `json:"address,omitempty"`
	ShipmentMethodID *uint    `json:"shipment_method_id,omitempty"`
	ShipmentStatusID *uint    `json:"shipment_status_id,omitempty"`
}

type ShipmentListResponse struct {
	Data  []ShipmentResponse `json:"data"`
	Total int64              `json:"total"`
	Page  int                `json:"page,omitempty"`
	Limit int                `json:"limit,omitempty"`
}
