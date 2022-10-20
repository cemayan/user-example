package dto

import "github.com/google/uuid"

type UpdateUser struct {
	ID        uuid.UUID `json:"id"`
	NickName  string    `json:"nickname"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Country   string    `json:"country"`
}
