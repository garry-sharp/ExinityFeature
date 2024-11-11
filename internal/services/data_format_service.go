package services

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"payment-gateway/internal/models"
)

// decodes the incoming request based on content type
func DecodeRequest[T any](r *http.Request, request *T) error {
	contentType := r.Header.Get("Content-Type")

	switch contentType {
	case "application/json":
		return json.NewDecoder(r.Body).Decode(request)
	case "text/xml":
		return xml.NewDecoder(r.Body).Decode(request)
	case "application/xml":
		return xml.NewDecoder(r.Body).Decode(request)
	default:
		return fmt.Errorf("unsupported content type")
	}
}

func TransactionRequestToEncrypted(txReq *models.TransactionRequest) (*models.TransactionRequestEncrypted, error) {
	tx := *txReq

	currency, err := Encrypt([]byte(tx.Currency))
	if err != nil {
		return nil, err
	}

	countryid, err := Encrypt([]byte(fmt.Sprint(tx.CountryID)))
	if err != nil {
		return nil, err
	}

	userid, err := Encrypt([]byte(fmt.Sprint(tx.UserID)))
	if err != nil {
		return nil, err
	}

	amount, err := Encrypt([]byte(tx.Amount.String()))
	if err != nil {
		return nil, err
	}

	typ, err := Encrypt([]byte(tx.Type))
	if err != nil {
		return nil, err
	}

	return &models.TransactionRequestEncrypted{
		GatewayID: tx.GatewayID, // this is no encrypted so it can picked up by consumer groups in kafka
		Type:      typ,
		Amount:    amount,
		UserID:    userid,
		CountryID: countryid,
		Currency:  currency,
	}, nil
}

func EncodeAndEncryptKafkaTransaction(kafkaTransacion *models.TransactionRequest, dataFormat string) ([]byte, error) {
	encrypted, err := TransactionRequestToEncrypted(kafkaTransacion)
	if err != nil {
		return nil, err
	}
	switch dataFormat {
	case "application/json":
		return json.Marshal(encrypted)
	case "text/xml", "application/xml":
		return xml.Marshal(encrypted)
	default:
		return nil, fmt.Errorf("unsupported data format")
	}
}
