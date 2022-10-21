package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

// Base contains common columns for all tables.
type Base struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Deleted   gorm.DeletedAt
}

// User struct
type User struct {
	*Base
	NickName  string `gorm:"uniqueIndex" json:"nickname" validate:"required"   `
	Email     string `gorm:"uniqueIndex" json:"email"  validate:"required,email" `
	Password  string `gorm:"not null" json:"password"  validate:"required" `
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Country   string `json:"country"`
}

type UserData struct {
	ID        uuid.UUID `json:"id"`
	NickName  string    `json:"nickname,omitempty"`
	Email     string    `json:"email,omitempty"`
	Password  string    `json:"password,omitempty"`
	FirstName string    `json:"first_name,omitempty"`
	LastName  string    `json:"last_name,omitempty"`
	Country   string    `json:"country,omitempty"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	var base Base
	base = Base{
		ID:        uuid.UUID{},
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
		Deleted:   gorm.DeletedAt{},
	}
	u.Base = &base
	str := uuid.New()
	u.ID = str
	return err
}
