package models

import "github.com/google/uuid"

type Transaction struct {
	ID                uuid.UUID
	SenderID          uuid.UUID
	SenderAccountID   uuid.UUID
	ReceiverAccountID uuid.UUID
	Amount            uint64
}
