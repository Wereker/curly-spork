package response

import "time"

type UserResponse struct {
	ID               uint           `json:"id"`
	Username         string         `json:"username"`
	Email            string         `json:"email"`
	FirstName        *string        `json:"first_name,omitempty"`
	LastName         *string        `json:"last_name,omitempty"`
	FatherName       *string        `json:"father_name,omitempty"`
	Phone            *string        `json:"phone,omitempty"`
	Address          *string        `json:"address,omitempty"`
	RegistrationDate time.Time      `json:"registration_date"`
	RightID          uint           `json:"right_id"`
	Right            *RightResponse `json:"right,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

type UserCreate struct {
	Username   string  `json:"username" validate:"required,min=3,max=150"`
	Email      string  `json:"email" validate:"required,email,max=100"`
	Password   string  `json:"password" validate:"required,min=6"`
	FirstName  *string `json:"first_name,omitempty" validate:"omitempty,max=100"`
	LastName   *string `json:"last_name,omitempty" validate:"omitempty,max=100"`
	FatherName *string `json:"father_name,omitempty" validate:"omitempty,max=100"`
	Phone      *string `json:"phone,omitempty" validate:"omitempty,max=100"`
	Address    *string `json:"address,omitempty"`
	RightID    uint    `json:"right_id" validate:"required"`
}

type UserUpdate struct {
	Username   string  `json:"username" validate:"required,min=3,max=150"`
	Email      string  `json:"email" validate:"required,email,max=100"`
	FirstName  *string `json:"first_name,omitempty" validate:"omitempty,max=100"`
	LastName   *string `json:"last_name,omitempty" validate:"omitempty,max=100"`
	FatherName *string `json:"father_name,omitempty" validate:"omitempty,max=100"`
	Phone      *string `json:"phone,omitempty" validate:"omitempty,max=100"`
	Address    *string `json:"address,omitempty"`
	RightID    uint    `json:"right_id" validate:"required"`
}

type UserPatch struct {
	Username   *string `json:"username,omitempty" validate:"omitempty,min=3,max=150"`
	Email      *string `json:"email,omitempty" validate:"omitempty,email,max=100"`
	FirstName  *string `json:"first_name,omitempty" validate:"omitempty,max=100"`
	LastName   *string `json:"last_name,omitempty" validate:"omitempty,max=100"`
	FatherName *string `json:"father_name,omitempty" validate:"omitempty,max=100"`
	Phone      *string `json:"phone,omitempty" validate:"omitempty,max=100"`
	Address    *string `json:"address,omitempty"`
	RightID    *uint   `json:"right_id,omitempty"`
}

type UserPasswordUpdate struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6"`
}

type UserListResponse struct {
	Data  []UserResponse `json:"data"`
	Total int64          `json:"total"`
	Page  int            `json:"page,omitempty"`
	Limit int            `json:"limit,omitempty"`
}
