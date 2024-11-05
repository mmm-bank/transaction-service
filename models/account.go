package models

import "github.com/google/uuid"

type Account struct {
	ID            uuid.UUID `json:"account_id" validate:"required"`
	UserID        uuid.UUID `json:"user_id" validate:"required"`
	AccountNumber string    `json:"account_number" validate:"required"`
	PhoneNumber   string    `json:"phone_number" validate:"required"`
	Currency      string    `json:"currency" validate:"required"`
	Balance       uint64    `json:"balance" validate:"required"`
}
