package storage

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mmm-bank/infra/security"
	"github.com/mmm-bank/transaction-service/models"
	"log"
	"time"
)

var _ Storage = PostgresCards{}

type Storage interface {
	GetAccountID() SearcherAccountID
	ProcessTransfer(transaction models.Transaction) error

	CreateAccount(account *models.Account) error
	LinkCardToAccount(cardNumber string, accountID uuid.UUID) error
	LinkPhoneToAccount(phoneNumber string, accountID uuid.UUID) error
}

type PostgresCards struct {
	db *pgxpool.Pool
}

func NewPostgresCards(connString string) PostgresCards {
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}
	return PostgresCards{pool}
}

func (p PostgresCards) GetAccountID() SearcherAccountID {
	return searcherAccountID{p.db}
}

func (p PostgresCards) ProcessTransfer(transaction models.Transaction) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := p.db.Exec(ctx, "CALL process_transfer($1, $2, $3, $4, $5)",
		transaction.ID, transaction.SenderID, transaction.SenderAccountID, transaction.ReceiverAccountID, transaction.Amount)
	return err
}

func (p PostgresCards) CreateAccount(account *models.Account) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO accounts (account_id, user_id, account_number, phone_number, currency, balance)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (account_id)
		DO NOTHING
	`
	accountNumber := security.Encrypt(account.AccountNumber, key)
	phoneNumber := security.Encrypt(account.PhoneNumber, key)
	_, err := p.db.Exec(ctx, query, account.ID, account.UserID, accountNumber, phoneNumber, account.Currency, account.Balance)
	return err
}

func (p PostgresCards) LinkCardToAccount(cardNumber string, accountID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO card_to_account (card_number, account_id)
		VALUES ($1, $2)
		ON CONFLICT (card_number) 
		DO UPDATE SET account_id = excluded.account_id
	`
	encryptedCardNumber := security.Encrypt(cardNumber, key)
	_, err := p.db.Exec(ctx, query, encryptedCardNumber, accountID)
	return err
}

func (p PostgresCards) LinkPhoneToAccount(phoneNumber string, accountID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO phone_to_account (phone_number, account_id)
		VALUES ($1, $2)
		ON CONFLICT (phone_number) 
		DO UPDATE SET account_id = excluded.account_id
	`
	encryptedPhoneNumber := security.Encrypt(phoneNumber, key)
	_, err := p.db.Exec(ctx, query, encryptedPhoneNumber, accountID)
	return err
}
