package http

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/mmm-bank/infra/middleware"
	"github.com/mmm-bank/transaction-service/models"
	"github.com/mmm-bank/transaction-service/storage"
	"go.uber.org/zap"
	"log"
	"net/http"
)

type CardTransactionRequest struct {
	SenderAccountID    string `json:"account_id" validate:"required"`
	ReceiverCardNumber string `json:"card_number" validate:"required"`
	Amount             uint64 `json:"amount" validate:"required"`
}

type PhoneTransactionRequest struct {
	SenderAccountID     string `json:"account_id" validate:"required"`
	ReceiverPhoneNumber string `json:"phone_number" validate:"required"`
	Amount              uint64 `json:"amount" validate:"required"`
}

type AccountTransactionRequest struct {
	SenderAccountID       string `json:"account_id" validate:"required"`
	ReceiverAccountNumber string `json:"account_number" validate:"required"`
	Amount                uint64 `json:"amount" validate:"required"`
}

type TransactionResponse struct {
	Status        string `json:"status"`
	TransactionID string `json:"transaction_id"`
}

type TransactionService struct {
	db     storage.Storage
	logger *zap.Logger
}

func NewCardService(db storage.Storage) *TransactionService {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to initialize zap logger: %v", err)
	}
	return &TransactionService{db: db, logger: logger}
}

func (t *TransactionService) parseCardTransaction(r *http.Request) (transaction models.Transaction, Err string) {
	var req CardTransactionRequest
	var err error
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		Err = "Failed to parse JSON request"
		return
	}
	if err = validator.New().Struct(req); err != nil {
		Err = "Missing fields"
		return
	}

	transaction.SenderAccountID, err = uuid.Parse(req.SenderAccountID)
	if err != nil {
		Err = "Invalid sender account ID"
		return
	}

	transaction.ReceiverAccountID, err = t.db.GetAccountID().ByCard(req.ReceiverCardNumber)
	if err != nil {
		Err = "Failed to get receiver account ID"
		return
	}

	transaction.ID = uuid.New()
	transaction.SenderID = r.Context().Value("user_id").(uuid.UUID)
	transaction.Amount = req.Amount
	return
}

func (t *TransactionService) postCardTransferHandler(w http.ResponseWriter, r *http.Request) {
	transaction, Err := t.parseCardTransaction(r)
	if Err != "" {
		http.Error(w, Err, http.StatusBadRequest)
		return
	}

	if err := t.db.ProcessTransfer(transaction); err != nil {
		ReceiverCardNumber := struct {
			Value string `json:"card_number"`
		}{}
		_ = json.NewDecoder(r.Body).Decode(&ReceiverCardNumber)

		t.logger.Error("Transfer failed",
			zap.Error(err),
			zap.String("sender_id", transaction.SenderID.String()),
			zap.String("sender_account_id", transaction.SenderAccountID.String()),
			zap.String("receiver_card_number", ReceiverCardNumber.Value),
			zap.Uint64("amount", transaction.Amount),
		)
		http.Error(w, "Transaction failed", http.StatusInternalServerError)
		return
	}

	resp := TransactionResponse{
		TransactionID: transaction.ID.String(),
		Status:        "Success",
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		t.logger.Error("Failed to encode JSON response", zap.Error(err))
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (t *TransactionService) parsePhoneTransaction(r *http.Request) (transaction models.Transaction, Err string) {
	var req PhoneTransactionRequest
	var err error
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		Err = "Failed to parse JSON request"
		return
	}
	if err = validator.New().Struct(req); err != nil {
		Err = "Missing fields"
		return
	}

	transaction.SenderAccountID, err = uuid.Parse(req.SenderAccountID)
	if err != nil {
		Err = "Invalid sender account ID"
		return
	}

	transaction.ReceiverAccountID, err = t.db.GetAccountID().ByPhone(req.ReceiverPhoneNumber)
	if err != nil {
		Err = "Failed to get receiver account ID"
		return
	}

	transaction.ID = uuid.New()
	transaction.SenderID = r.Context().Value("user_id").(uuid.UUID)
	transaction.Amount = req.Amount
	return
}

func (t *TransactionService) postPhoneTransferHandler(w http.ResponseWriter, r *http.Request) {
	transaction, Err := t.parsePhoneTransaction(r)
	if Err != "" {
		http.Error(w, Err, http.StatusBadRequest)
		return
	}

	if err := t.db.ProcessTransfer(transaction); err != nil {
		ReceiverPhoneNumber := struct {
			Value string `json:"phone_number"`
		}{}
		_ = json.NewDecoder(r.Body).Decode(&ReceiverPhoneNumber)

		t.logger.Error("Transfer failed",
			zap.Error(err),
			zap.String("sender_id", transaction.SenderID.String()),
			zap.String("sender_account_id", transaction.SenderAccountID.String()),
			zap.String("receiver_phone_number", ReceiverPhoneNumber.Value),
			zap.Uint64("amount", transaction.Amount),
		)
		http.Error(w, "Transaction failed", http.StatusInternalServerError)
		return
	}

	resp := TransactionResponse{
		TransactionID: transaction.ID.String(),
		Status:        "Success",
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		t.logger.Error("Failed to encode JSON response", zap.Error(err))
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (t *TransactionService) parseAccountTransaction(r *http.Request) (transaction models.Transaction, Err string) {
	var req AccountTransactionRequest
	var err error
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		Err = "Failed to parse JSON request"
		return
	}
	if err = validator.New().Struct(req); err != nil {
		Err = "Missing fields"
		return
	}

	transaction.SenderAccountID, err = uuid.Parse(req.SenderAccountID)
	if err != nil {
		Err = "Invalid sender account ID"
		return
	}

	transaction.ReceiverAccountID, err = t.db.GetAccountID().ByAccount(req.ReceiverAccountNumber)
	if err != nil {
		Err = "Failed to get receiver account ID"
		return
	}

	transaction.ID = uuid.New()
	transaction.SenderID = r.Context().Value("user_id").(uuid.UUID)
	transaction.Amount = req.Amount
	return
}

func (t *TransactionService) postAccountTransferHandler(w http.ResponseWriter, r *http.Request) {
	transaction, Err := t.parseAccountTransaction(r)
	if Err != "" {
		http.Error(w, Err, http.StatusBadRequest)
		return
	}

	if err := t.db.ProcessTransfer(transaction); err != nil {
		ReceiverAccountNumber := struct {
			Value string `json:"account_number"`
		}{}
		_ = json.NewDecoder(r.Body).Decode(&ReceiverAccountNumber)

		t.logger.Error("Transfer failed",
			zap.Error(err),
			zap.String("sender_id", transaction.SenderID.String()),
			zap.String("sender_account_id", transaction.SenderAccountID.String()),
			zap.String("receiver_account_number", ReceiverAccountNumber.Value),
			zap.Uint64("amount", transaction.Amount),
		)
		http.Error(w, "Transaction failed", http.StatusInternalServerError)
		return
	}

	resp := TransactionResponse{
		TransactionID: transaction.ID.String(),
		Status:        "Success",
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		t.logger.Error("Failed to encode JSON response", zap.Error(err))
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

type AccountRequest struct {
	models.Account
	CardNumber string `json:"card_number" validate:"required"`
}

func parseAccountInfo(r *http.Request) (req AccountRequest, Err string) {
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Err = "Failed to parse JSON request"
		return
	}
	if err := validator.New().Struct(req); err != nil {
		Err = "Missing fields"
	}
	return
}

func (t *TransactionService) postCreateAccountAndLinkCard(w http.ResponseWriter, r *http.Request) {
	req, Err := parseAccountInfo(r)
	if Err != "" {
		http.Error(w, Err, http.StatusBadRequest)
		return
	}

	if err := t.db.CreateAccount(&req.Account); err != nil {
		http.Error(w, "Failed to create account", http.StatusInternalServerError)
		t.logger.Error("Failed to create account", zap.Error(err))
		return
	}

	if err := t.db.LinkCardToAccount(req.CardNumber, req.Account.ID); err != nil {
		http.Error(w, "Failed to link card to account", http.StatusInternalServerError)
		t.logger.Error("Failed to link card account", zap.Error(err))
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (t *TransactionService) postLinkCardToAccount(w http.ResponseWriter, r *http.Request) {
	var data struct {
		CardNumber string    `json:"card_number"`
		AccountID  uuid.UUID `json:"account_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid card data", http.StatusBadRequest)
		return
	}

	if len(data.CardNumber) != 16 {
		http.Error(w, "Invalid card details", http.StatusBadRequest)
		return
	}

	if data.AccountID == uuid.Nil {
		http.Error(w, "Invalid account ID", http.StatusBadRequest)
		return
	}

	err := t.db.LinkCardToAccount(data.CardNumber, data.AccountID)
	if err != nil {
		t.logger.Error("Failed to link card to account", zap.Error(err))
		http.Error(w, "Failed to link card to account", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (t *TransactionService) postLinkPhoneToAccount(w http.ResponseWriter, r *http.Request) {
	var data struct {
		PhoneNumber string    `json:"phone_number"`
		AccountID   uuid.UUID `json:"account_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid phone data", http.StatusBadRequest)
		return
	}

	if len(data.PhoneNumber) != 12 {
		http.Error(w, "Invalid phone number", http.StatusBadRequest)
		return
	}

	if data.AccountID == uuid.Nil {
		http.Error(w, "Invalid account ID", http.StatusBadRequest)
		return
	}

	err := t.db.LinkPhoneToAccount(data.PhoneNumber, data.AccountID)
	if err != nil {
		t.logger.Error("Failed to link phone to account", zap.Error(err))
		http.Error(w, "Failed to link phone to account", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func CreateAndRunServer(t *TransactionService, addr string) error {
	r := chi.NewRouter()
	r.Route("/transaction", func(r chi.Router) {
		r.Use(mymiddleware.ExtractPayload)
		r.Route("/transfer", func(r chi.Router) {
			r.Post("/card", t.postCardTransferHandler)
			r.Post("/phone", t.postPhoneTransferHandler)
			r.Post("/account", t.postAccountTransferHandler)
		})
	})

	r.Route("/service", func(r chi.Router) {
		r.Route("/account", func(r chi.Router) {
			r.Post("/create", t.postCreateAccountAndLinkCard)
			r.Post("/link/card", t.postLinkCardToAccount)
			r.Post("/link/phone", t.postLinkPhoneToAccount)
		})
	})
	return http.ListenAndServe(addr, r)
}
