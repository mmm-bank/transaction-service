package storage

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mmm-bank/infra/security"
	"os"
)

var _ SearcherAccountID = searcherAccountID{}

type SearcherAccountID interface {
	ByPhone(phoneNumber string) (uuid.UUID, error)
	ByCard(cardNumber string) (uuid.UUID, error)
	ByAccount(accountNumber string) (uuid.UUID, error)
}

var key = os.Getenv("AES_KEY")

type searcherAccountID struct {
	db *pgxpool.Pool
}

func (s searcherAccountID) ByPhone(phoneNumber string) (uuid.UUID, error) {
	var accountID uuid.UUID
	phoneNumberBytea := security.Encrypt(phoneNumber, key)

	query := "SELECT account_id FROM phone_to_account WHERE phone_number = $1"
	err := s.db.QueryRow(context.Background(), query, phoneNumberBytea).Scan(&accountID)
	if err != nil {
		return uuid.UUID{}, err
	}
	return accountID, nil
}

func (s searcherAccountID) ByCard(cardNumber string) (uuid.UUID, error) {
	var accountID uuid.UUID
	cardNumberBytea := security.Encrypt(cardNumber, key)

	query := "SELECT account_id FROM card_to_account WHERE card_number = $1"
	err := s.db.QueryRow(context.Background(), query, cardNumberBytea).Scan(&accountID)
	if err != nil {
		return uuid.UUID{}, err
	}
	return accountID, nil
}

func (s searcherAccountID) ByAccount(accountNumber string) (uuid.UUID, error) {
	var accountID uuid.UUID
	accountNumberBytea := security.Encrypt(accountNumber, key)

	query := "SELECT account_id FROM accounts WHERE account_number = $1"
	err := s.db.QueryRow(context.Background(), query, accountNumberBytea).Scan(&accountID)
	if err != nil {
		return uuid.UUID{}, err
	}
	return accountID, nil
}
