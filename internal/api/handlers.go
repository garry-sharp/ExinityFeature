package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"payment-gateway/db"
	"payment-gateway/internal/models"
	"payment-gateway/internal/services"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type ContentType string

const (
	JSON ContentType = "application/json"
	XML  ContentType = "application/xml"
)

func returnJSONError(message, detailedmessage string, statusCode int, w http.ResponseWriter) {
	enc := json.NewEncoder(w)
	enc.Encode(models.APIResponse[any]{
		StatusCode: statusCode,
		Error: &models.Error{
			Message:         message,
			DetailedMessage: detailedmessage,
		},
	})
}

func returnXMLError(message, detailedmessage string, statusCode int, w http.ResponseWriter) {
	enc := xml.NewEncoder(w)
	enc.Encode(models.APIResponse[any]{
		StatusCode: statusCode,
		Error: &models.Error{
			Message:         message,
			DetailedMessage: detailedmessage,
		},
	})
}

func returnError(message, detailedmessage string, statusCode int, w http.ResponseWriter, typ ContentType) {
	switch typ {
	case XML:
		returnXMLError(message, detailedmessage, statusCode, w)
	case JSON:
		returnJSONError(message, detailedmessage, statusCode, w)
	default:
		returnJSONError(message, detailedmessage, statusCode, w)
	}
}

func returnResponse[T any](response T, statusCode int, w http.ResponseWriter, typ ContentType) {
	switch typ {
	case XML:
		w.Header().Add("Content-Type", "application/xml")
		enc := xml.NewEncoder(w)
		err := enc.Encode(models.APIResponse[T]{
			StatusCode: statusCode,
			Data:       response,
		})
		fmt.Println(err)
	case JSON:
		w.Header().Add("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.Encode(models.APIResponse[T]{
			StatusCode: statusCode,
			Data:       response,
		})
	default:
		w.Header().Add("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.Encode(models.APIResponse[T]{
			StatusCode: statusCode,
			Data:       response,
		})
	}
}

// Takes a deposit request via POST HTTP verb. Sanity checks the request. Then creates a transaction in the DB and sends a message to Kafka.
func DepositPostHandler(w http.ResponseWriter, r *http.Request) {
	contentType, ok := r.Context().Value("contentType").(ContentType)
	if !ok {
		returnJSONError("unsupported context type", "", http.StatusBadRequest, w)
		return
	}
	_db, err := db.GetDB()
	if err != nil {
		returnError("unable to connect to DB", "", http.StatusInternalServerError, w, contentType)
		return
	}
	depositPostHandler(_db, r.Context(), w)
}

func depositPostHandler(_db *sql.DB, ctx context.Context, w http.ResponseWriter) {
	contentType, ok := ctx.Value("contentType").(ContentType)
	if !ok {
		returnJSONError("unsupported context type", "", http.StatusBadRequest, w)
		return
	}
	request, ok := ctx.Value("request").(models.DepositRequest)
	if !ok {
		returnError("unable to parse body", "", http.StatusBadRequest, w, contentType)
		return
	}

	rows, err := _db.QueryContext(ctx, "SELECT country_id from users where id = $1", request.UserID)
	if err != nil || !rows.Next() {
		returnError("User not found", "", http.StatusNotFound, w, contentType)
		return
	}
	defer rows.Close()
	countryID := 0
	if err := rows.Scan(&countryID); err != nil {
		returnError("User not found", "", http.StatusNotFound, w, contentType)
		return
	}

	//Validate that the currency requested is supported in the region
	if supported, err := db.CurrencySupportedInCountry(ctx, _db, request.Currency, countryID); !supported || err != nil {
		returnError("Currency not supported in country", "", http.StatusBadRequest, w, contentType)
		return
	}

	//Validate the amount requested is no more than 2 decimal places and non negative (or less an 0.01)
	if !services.CurrencyAmountIsValid(request.Amount) {
		returnError("Invalid amount", "Amount must be not be more than 2 decimal places", http.StatusBadRequest, w, contentType)
		return
	}

	gateway, err := db.GetRandomGateway(ctx, _db, countryID, request.Currency, string(contentType))
	if err != nil {
		returnError("unable to get gateway", err.Error(), http.StatusInternalServerError, w, contentType)
		return
	}

	txReq := models.TransactionRequest{
		Type:      "deposit",
		Amount:    request.Amount,
		UserID:    request.UserID,
		CountryID: countryID,
		Currency:  request.Currency,
		GatewayID: gateway.ID,
	}

	if err := services.RetryOperation(func() error {
		return SendKafkaMessageAndDB(ctx, _db, &txReq, string(contentType))
	}, 3); err != nil {
		returnError("unable to create transaction", err.Error(), http.StatusInternalServerError, w, contentType)
		return
	}

	returnTransaction(ctx, http.StatusCreated, w, contentType, _db, fmt.Sprint(txReq.TransactionID), db.DEPOSIT)
}

// Takes a withdrawal request via POST HTTP verb. Sanity checks the request. Then creates a transaction in the DB and sends a message to Kafka.
func WithdrawalPostHandler(w http.ResponseWriter, r *http.Request) {
	contentType, ok := r.Context().Value("contentType").(ContentType)
	if !ok {
		returnJSONError("unsupported context type", "", http.StatusBadRequest, w)
		return
	}
	_db, err := db.GetDB()
	if err != nil {
		returnError("unable to connect to DB", "", http.StatusInternalServerError, w, contentType)
		return
	}
	withdrawalPostHandler(_db, r.Context(), w)
}

func withdrawalPostHandler(_db *sql.DB, ctx context.Context, w http.ResponseWriter) {
	contentType, ok := ctx.Value("contentType").(ContentType)
	if !ok {
		returnJSONError("unsupported context type", "", http.StatusBadRequest, w)
		return
	}
	request, ok := ctx.Value("request").(models.WithdrawalRequest)
	if !ok {
		returnError("unable to parse body", "", http.StatusBadRequest, w, contentType)
		return
	}

	rows, err := _db.QueryContext(ctx, "SELECT country_id from users where id = $1", request.UserID)
	if err != nil || !rows.Next() {
		returnError("User not found", "", http.StatusNotFound, w, contentType)
		return
	}
	defer rows.Close()
	countryID := 0
	if err := rows.Scan(&countryID); err != nil {
		returnError("User not found", "", http.StatusNotFound, w, contentType)
		return
	}

	//Validate that the currency requested is supported in the region
	if supported, err := db.CurrencySupportedInCountry(ctx, _db, request.Currency, countryID); !supported || err != nil {
		returnError("Currency not supported in country", "", http.StatusBadRequest, w, contentType)
		return
	}

	//Validate the amount requested is no more than 2 decimal places and non negative (or less an 0.01)
	if !services.CurrencyAmountIsValid(request.Amount) {
		returnError("Invalid amount", "Amount must be not be more than 2 decimal places", http.StatusBadRequest, w, contentType)
		return
	}

	gateway, err := db.GetRandomGateway(ctx, _db, countryID, request.Currency, string(contentType))
	if err != nil {
		returnError("unable to get gateway", err.Error(), http.StatusInternalServerError, w, contentType)
		return
	}

	txReq := models.TransactionRequest{
		Type:      "withdrawal",
		Amount:    request.Amount,
		UserID:    request.UserID,
		CountryID: countryID,
		Currency:  request.Currency,
		GatewayID: gateway.ID,
	}

	if err := services.RetryOperation(func() error {
		return SendKafkaMessageAndDB(ctx, _db, &txReq, string(contentType))
	}, 3); err != nil {
		returnError("unable to create transaction", err.Error(), http.StatusInternalServerError, w, contentType)
		return
	}

	returnTransaction(ctx, http.StatusCreated, w, contentType, _db, fmt.Sprint(txReq.TransactionID), db.WITHDRAWAL)
}

func DepositPutHandler(w http.ResponseWriter, r *http.Request) {
	request := r.Context().Value("request").(models.DepositPutRequest)
	contentType := r.Context().Value("contentType").(ContentType)

	_db, err := db.GetDB()
	if err != nil {
		returnError("unable to connect to DB", "", http.StatusInternalServerError, w, contentType)
	}
	tx, err := db.GetTransaction(r.Context(), _db, request.TransactionID, db.DEPOSIT)
	if err != nil {
		returnError("unable to get transaction", err.Error(), http.StatusInternalServerError, w, contentType)
		return
	}

	if tx.Status != db.SENT {
		returnError("Transaction already processed", "", http.StatusBadRequest, w, contentType)
		return
	}

	if strings.ToLower(request.Status) == "success" {
		_, err := _db.ExecContext(r.Context(), "UPDATE transactions SET status = $1 WHERE id = $2 and type = $3", db.SUCCESS, request.TransactionID, db.DEPOSIT)
		if err != nil {
			returnError("unable to update transaction", err.Error(), http.StatusInternalServerError, w, contentType)
			return
		}
	} else if strings.ToLower(request.Status) == "failed" {
		_, err := _db.ExecContext(r.Context(), "UPDATE transactions SET status = $1 WHERE id = $2 and type = $3", db.FAILED, request.TransactionID, db.DEPOSIT)
		if err != nil {
			returnError("unable to update transaction", err.Error(), http.StatusInternalServerError, w, contentType)
			return
		}
	} else {
		returnError("Invalid status", "", http.StatusBadRequest, w, contentType)
		return
	}

	returnTransaction(r.Context(), http.StatusOK, w, contentType, _db, fmt.Sprint(request.TransactionID), db.DEPOSIT)
}

func WithdrawalPutHandler(w http.ResponseWriter, r *http.Request) {
	request := r.Context().Value("request").(models.WithdrawalPutRequest)
	contentType := r.Context().Value("contentType").(ContentType)

	_db, err := db.GetDB()
	if err != nil {
		returnError("unable to connect to DB", "", http.StatusInternalServerError, w, contentType)
	}
	tx, err := db.GetTransaction(r.Context(), _db, request.TransactionID, db.WITHDRAWAL)
	if err != nil {
		returnError("unable to get transaction", err.Error(), http.StatusInternalServerError, w, contentType)
		return
	}

	if tx.Status != db.SENT {
		returnError("Transaction already processed", "", http.StatusBadRequest, w, contentType)
		return
	}

	if strings.ToLower(request.Status) == "success" {
		_, err := _db.ExecContext(r.Context(), "UPDATE transactions SET status = $1 WHERE id = $2 and type = $3", db.SUCCESS, request.TransactionID, db.WITHDRAWAL)
		if err != nil {
			returnError("unable to update transaction", err.Error(), http.StatusInternalServerError, w, contentType)
			return
		}
	} else if strings.ToLower(request.Status) == "failed" {
		_, err := _db.ExecContext(r.Context(), "UPDATE transactions SET status = $1 WHERE id = $2 and type = $3", db.FAILED, request.TransactionID, db.WITHDRAWAL)
		if err != nil {
			returnError("unable to update transaction", err.Error(), http.StatusInternalServerError, w, contentType)
			return
		}
	} else {
		returnError("Invalid status", "", http.StatusBadRequest, w, contentType)
		return
	}

	returnTransaction(r.Context(), http.StatusOK, w, contentType, _db, fmt.Sprint(request.TransactionID), db.WITHDRAWAL)
}

func DepositGetHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := NewHandlerContext(time.Second * 5)
	defer cancel()
	contentType := JSON
	if ct := r.Header.Get("Content-Type"); ct == "application/xml" || ct == "text/xml" {
		contentType = XML
	}

	idstr := mux.Vars(r)["id"]
	_db, err := db.GetDB()
	if err != nil {
		returnError("unable to connect to DB", "", http.StatusInternalServerError, w, contentType)
		return
	}

	returnTransaction(ctx, http.StatusOK, w, contentType, _db, idstr, db.DEPOSIT)
}

// TODO should return type based on "Accept" header?
func WithdrawalGetHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := NewHandlerContext(time.Second * 5)
	defer cancel()
	contentType := JSON
	if ct := r.Header.Get("Content-Type"); ct == "application/xml" || ct == "text/xml" {
		contentType = XML
	}

	idstr := mux.Vars(r)["id"]
	_db, err := db.GetDB()
	if err != nil {
		returnError("unable to connect to DB", "", http.StatusInternalServerError, w, contentType)
		return
	}

	returnTransaction(ctx, http.StatusOK, w, contentType, _db, idstr, db.WITHDRAWAL)
}
