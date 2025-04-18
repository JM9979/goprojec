package entity

type UsernameRequest struct {
	Username string `json:"username" validate:"required,min=2,max=20"`
}