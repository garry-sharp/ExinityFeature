package models

import (
	"time"

	decimal "github.com/shopspring/decimal"
)

// a standard request structure for the transactions
type TransactionRequest struct {
	TransactionID int             `json:"transaction_id" xml:"transaction_id"`
	Type          string          `json:"type" xml:"type"` // deposit or withdrawal
	Amount        decimal.Decimal `json:"amount" xml:"amount"`
	UserID        int             `json:"user_id" xml:"user_id"`
	CountryID     int             `json:"country_id" xml:"country_id"`
	Currency      string          `json:"currency" xml:"currency"`
	GatewayID     int             `json:"gateway_id" xml:"gateway_id"`
}

type TransactionRequestEncrypted struct {
	Type      string `json:"type" xml:"type"` // deposit or withdrawal
	Amount    string `json:"amount" xml:"amount"`
	UserID    string `json:"user_id" xml:"user_id"`
	CountryID string `json:"country_id" xml:"country_id"`
	Currency  string `json:"currency" xml:"currency"`
	GatewayID int    `json:"gateway_id" xml:"gateway_id"`
}

type Error struct {
	Message         string `json:"message,omitempty"`
	DetailedMessage string `json:"detailed_message,omitempty"`
}

// a standard response structure for the APIs
type APIResponse[T any] struct {
	StatusCode int    `json:"status_code" xml:"status_code"`
	Data       T      `json:"data,omitempty" xml:"data,omitempty"`
	Error      *Error `json:"error,omitempty" xml:"error,omitempty"`
}

type WithdrawalRequest struct {
	Amount   decimal.Decimal `json:"amount" xml:"amount"`
	UserID   int             `json:"user_id" xml:"user_id"`
	Currency string          `json:"currency" xml:"currency"`
}

type DepositRequest struct {
	Amount   decimal.Decimal `json:"amount" xml:"amount"`
	UserID   int             `json:"user_id" xml:"user_id"`
	Currency string          `json:"currency" xml:"currency"`
}

type WithdrawalPutRequest struct {
	TransactionID int    `json:"transaction_id" xml:"transaction_id"`
	Status        string `json:"status" xml:"status"`
}

type DepositPutRequest struct {
	TransactionID int    `json:"transaction_id" xml:"transaction_id"`
	Status        string `json:"status" xml:"status"`
}

type WithdrawalResponse struct {
	TransactionID int       `json:"transaction_id" xml:"transaction_id"`
	Created       time.Time `json:"created" xml:"created"`
	Status        string    `json:"status" xml:"status"`
}

type DepositResponse struct {
	TransactionID int       `json:"transaction_id" xml:"transaction_id"`
	Created       time.Time `json:"created" xml:"created"`
	Status        string    `json:"status" xml:"status"`
}
