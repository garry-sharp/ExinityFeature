package api

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"payment-gateway/db"
	"payment-gateway/internal/kafka"
	"payment-gateway/internal/models"
	"payment-gateway/internal/services"
	"strconv"
	"time"
)

func NewHandlerContext(duration time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), duration)
}

// Encrypts the kafka message
// Creates a DB transaction
// Uses a circuit breaker to publish the message to Kafka
// If there are any failures the DB tx rollsback otherwise commits
// Returns the transaction ID and error (if any)
func SendKafkaMessageAndDB(ctx context.Context, _db *sql.DB, txReq *models.TransactionRequest, requestContentType string) error {

	// Create a sql transaction from the txReq and write a transaction to the DB which will be committed on successful completion of the rest of the code
	tx, err := _db.Begin()
	if err != nil {
		return err
	}

	var typ db.TransactionType
	switch txReq.Type {
	case "deposit":
		typ = db.DEPOSIT
	case "withdrawal":
		typ = db.WITHDRAWAL
	default:
		return fmt.Errorf("invalid transaction type")
	}

	transaction := db.Transaction{
		Amount:    txReq.Amount,
		Type:      typ,
		UserID:    txReq.UserID,
		CountryID: txReq.CountryID,
		Status:    db.SENT,
		GatewayID: txReq.GatewayID,
	}

	if err := db.CreateTransaction(ctx, tx, &transaction); err != nil {
		tx.Rollback()
		return err
	}

	// Encode the kafka txReq to xml/json and then AES encrypt it.
	encryptedKafkaMessage, err := services.EncodeAndEncryptKafkaTransaction(txReq, requestContentType)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := kafka.PublishTransaction(context.Background(), fmt.Sprint(transaction.ID), encryptedKafkaMessage, requestContentType); err != nil {
		tx.Rollback()
		return err
	}

	txReq.TransactionID = transaction.ID

	return tx.Commit()
}

func returnTransaction(ctx context.Context, statusCode int, w http.ResponseWriter, contentType ContentType, _db *sql.DB, txid string, txType db.TransactionType) {
	if txid == "" {
		returnError("ID not passed", "", http.StatusBadRequest, w, contentType)
		return
	}

	if _db == nil {
		returnError("unable to connect to DB", "", http.StatusInternalServerError, w, contentType)
		return
	}

	id, err := strconv.ParseInt(txid, 10, 32)
	if err != nil {
		returnError("ID not a number", "", http.StatusBadRequest, w, contentType)
		return
	}

	tx, err := db.GetTransaction(ctx, _db, int(id), txType)
	if err != nil {
		returnError("unable to get transaction", err.Error(), http.StatusNotFound, w, contentType)
		return
	}

	returnResponse(tx, statusCode, w, contentType)
}
