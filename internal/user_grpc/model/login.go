package model

// LoginInput is representation of the login payload
type LoginInput struct {
	Nickname string `json:"nickname"`
	Password string `json:"password"`
}
