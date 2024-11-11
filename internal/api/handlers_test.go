package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"payment-gateway/internal/models"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
)

func TestReturnError(t *testing.T) {
	// Setup a ResponseRecorder to capture the HTTP response
	rr := httptest.NewRecorder()

	// Call the function
	returnError("Test Error", "Detailed Test Error", http.StatusInternalServerError, rr, JSON)

	resp := rr.Result()
	defer resp.Body.Close()

	apiResponse := models.APIResponse[any]{}
	json.NewDecoder(resp.Body).Decode(&apiResponse)
	// Assert the response

	if apiResponse.Error == nil {
		t.Error("Api error response should not be nil")
	}

	if apiResponse.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected %v received %v", http.StatusInternalServerError, apiResponse.StatusCode)
	}

	expected := `{"status_code":500,"error":{"message":"Test Error","detailed_message":"Detailed Test Error"}}`
	if strings.TrimSpace(rr.Body.String()) != expected {
		t.Errorf("unexpected body: got %v, want %v", rr.Body.String(), expected)
	}
}

func bootstraptest(v interface{}, method, endpoint string, contentType ContentType) (*sql.DB, sqlmock.Sqlmock, error, []byte, *httptest.ResponseRecorder, *http.Request) {
	_db, mock, err := sqlmock.New()
	body, _ := json.Marshal(v)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(method, endpoint, bytes.NewReader(body))
	ctx := context.WithValue(context.Background(), "contentType", contentType)
	ctx = context.WithValue(ctx, "request", v)
	req = req.WithContext(ctx)
	return _db, mock, err, body, rr, req
}

func assertResponse(expected []byte, rr *httptest.ResponseRecorder, t *testing.T) {
	resp := rr.Result()
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)

	if strings.TrimSpace(string(expected)) != strings.TrimSpace(string(b)) {
		fmt.Println("bAD MATHC")
		t.Errorf("Expected %v\nReceived %v", string(expected), string(b))
	}
}

// This is the only function tested and only partially. There are time limits on this assessment I somewhat wanted to respect.
// This is an idea of how you would do it. You would mock the DB and read the response from the httptest.ResponseRecorder
// Here I'm passing in different request values and mocking the DB.
func TestDepositPostHandler(t *testing.T) {

	//Test User not Found
	_db, _, _, _, rr, req := bootstraptest(models.DepositRequest{
		Amount:   decimal.NewFromInt(20),
		UserID:   1,
		Currency: "GBP",
	}, "POST", "/withdrawal", JSON)
	depositPostHandler(_db, req.Context(), rr)
	assertResponse([]byte(`{"status_code":404,"error":{"message":"User not found"}}`), rr, t)

	//Test Currency Unsupported
	_db, mock, _, _, rr, req := bootstraptest(models.DepositRequest{
		Amount:   decimal.NewFromInt(20),
		UserID:   1,
		Currency: "GBP",
	}, "POST", "/withdrawal", JSON)
	mock.ExpectQuery("SELECT country_id from users where id = (.+)").WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"country_id"}).AddRow(1))
	depositPostHandler(_db, req.Context(), rr)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
	assertResponse([]byte(`{"status_code":400,"error":{"message":"Currency not supported in country"}}`), rr, t)

}
