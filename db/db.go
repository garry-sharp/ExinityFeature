package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"payment-gateway/internal/services"
	"time"

	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
)

var db *sql.DB

type Execer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

type User struct {
	ID        int
	Username  string
	Email     string
	CountryID int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Gateway struct {
	ID                  int
	Name                string
	DataFormatSupported string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type Country struct {
	ID        int
	Name      string
	Code      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Currency struct {
	ID     int
	Symbol string
}

type TransactionType string
type TransactionStatus string

const (
	DEPOSIT    TransactionType = "DEPOSIT"
	WITHDRAWAL TransactionType = "WITHDRAWAL"
)

const (
	DRAFT   TransactionStatus = "DRAFT"
	SENT    TransactionStatus = "SENT"
	SUCCESS TransactionStatus = "SUCCESS"
	FAILED  TransactionStatus = "FAILED"
)

type Transaction struct {
	ID        int
	Amount    decimal.Decimal
	Type      TransactionType
	Status    TransactionStatus
	UserID    int
	GatewayID int
	CountryID int
	CreatedAt time.Time
}

// InitializeDB initializes the database connection
func InitializeDB(dataSourceName string) {
	var err error

	err = services.RetryOperation(func() error {
		db, err = sql.Open("postgres", dataSourceName)
		if err != nil {
			return err
		}

		return db.Ping()
	}, 5)

	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}

	log.Println("Successfully connected to the database.")
}

func GetDB() (*sql.DB, error) {
	if db == nil {
		return nil, fmt.Errorf("db not initialized")
	}
	return db, nil
}

func CreateUser(ctx context.Context, db Execer, user *User) error {
	query := `INSERT INTO users (username, email, country_id, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5) RETURNING id`

	err := db.QueryRow(query, user.Username, user.Email, user.CountryID, time.Now(), time.Now()).Scan(&user.ID)
	if err != nil {
		return fmt.Errorf("failed to insert user: %v", err)
	}
	return nil
}

func GetUsers(ctx context.Context, db *sql.DB) ([]User, error) {
	rows, err := db.Query(`SELECT id, username, email, country_id, created_at, updated_at FROM users`)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %v", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.CountryID, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan user: %v", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func CreateGateway(ctx context.Context, db Execer, gateway *Gateway) error {
	query := `INSERT INTO gateways (name, data_format_supported, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4) RETURNING id`

	err := db.QueryRow(query, gateway.Name, gateway.DataFormatSupported, time.Now(), time.Now()).Scan(&gateway.ID)
	if err != nil {
		return fmt.Errorf("failed to insert gateway: %v", err)
	}
	return nil
}

func GetGateways(ctx context.Context, db *sql.DB) ([]Gateway, error) {
	rows, err := db.Query(`SELECT id, name, data_format_supported, created_at, updated_at FROM gateways`)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch gateways: %v", err)
	}
	defer rows.Close()

	var gateways []Gateway
	for rows.Next() {
		var gateway Gateway
		if err := rows.Scan(&gateway.ID, &gateway.Name, &gateway.DataFormatSupported, &gateway.CreatedAt, &gateway.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan gateway: %v", err)
		}
		gateways = append(gateways, gateway)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return gateways, nil
}

func CreateCountry(ctx context.Context, db Execer, country *Country) error {
	query := `INSERT INTO countries (name, code, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4) RETURNING id`

	err := db.QueryRow(query, country.Name, country.Code, time.Now(), time.Now()).Scan(&country.ID)
	if err != nil {
		return fmt.Errorf("failed to insert country: %v", err)
	}
	return nil
}

func GetCountries(ctx context.Context, db *sql.DB) ([]Country, error) {
	rows, err := db.Query(`SELECT id, name, code, created_at, updated_at FROM countries`)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch countries: %v", err)
	}
	defer rows.Close()

	var countries []Country
	for rows.Next() {
		var country Country
		if err := rows.Scan(&country.ID, &country.Name, &country.Code, &country.CreatedAt, &country.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan country: %v", err)
		}
		countries = append(countries, country)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return countries, nil
}

func CreateCurrency(ctx context.Context, db Execer, currency *Currency) error {
	query := `INSERT INTO currencies (symbol) 
			  VALUES ($1) RETURNING id`

	err := db.QueryRow(query, currency.Symbol).Scan(&currency.ID)
	if err != nil {
		return fmt.Errorf("failed to insert currency: %v", err)
	}
	return nil
}

func GetCurrencies(ctx context.Context, db *sql.DB) ([]Currency, error) {
	rows, err := db.Query(`SELECT id, symbol FROM currencies`)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch currencies: %v", err)
	}
	defer rows.Close()

	var currencies []Currency
	for rows.Next() {
		var currency Currency
		if err := rows.Scan(&currency.ID, &currency.Symbol); err != nil {
			return nil, fmt.Errorf("failed to scan country: %v", err)
		}
		currencies = append(currencies, currency)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return currencies, nil
}

func CreateTransaction(ctx context.Context, db Execer, transaction *Transaction) error {
	query := `INSERT INTO transactions (amount, type, status, gateway_id, country_id, user_id, created_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`

	err := db.QueryRow(query, transaction.Amount, transaction.Type, transaction.Status, transaction.GatewayID, transaction.CountryID, transaction.UserID, time.Now()).Scan(&transaction.ID)
	if err != nil {
		return fmt.Errorf("failed to insert transaction: %v", err)
	}
	return nil
}

func GetTransactions(ctx context.Context, db *sql.DB) ([]Transaction, error) {
	rows, err := db.Query(`SELECT id, amount, type, status, user_id, gateway_id, country_id, created_at FROM transactions`)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transactions: %v", err)
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		var transaction Transaction
		if err := rows.Scan(&transaction.ID, &transaction.Amount, &transaction.Type, &transaction.Status, &transaction.UserID, &transaction.GatewayID, &transaction.CountryID, &transaction.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %v", err)
		}
		transactions = append(transactions, transaction)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return transactions, nil
}

func GetTransaction(ctx context.Context, db *sql.DB, transactionID int, txType TransactionType) (Transaction, error) {
	rows, err := db.Query(`SELECT id, amount, type, status, user_id, gateway_id, country_id, created_at FROM transactions WHERE id = $1 and type = $2`, transactionID, txType)
	if err != nil {
		return Transaction{}, fmt.Errorf("failed to get transaction %d: %v", transactionID, err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Err(); err != nil {
			return Transaction{}, err
		}
		var transaction Transaction
		if err := rows.Scan(&transaction.ID, &transaction.Amount, &transaction.Type, &transaction.Status, &transaction.UserID, &transaction.GatewayID, &transaction.CountryID, &transaction.CreatedAt); err != nil {
			return Transaction{}, fmt.Errorf("failed to scan transaction: %v", err)
		}
		return transaction, nil
	} else {
		return Transaction{}, fmt.Errorf("no transaction found with id %d", transactionID)
	}
}

// func GetSupportedCountriesByGateway(db *sql.DB, gatewayID int) ([]Country, error) {
// 	query := `
// 		SELECT c.id AS country_id, c.name AS country_name
// 		FROM countries c
// 		JOIN gateway_countries gc ON c.id = gc.country_id
// 		WHERE gc.gateway_id = $1
// 		ORDER BY c.name
// 	`

// 	rows, err := db.Query(query, gatewayID)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to fetch countries for gateway %d: %v", gatewayID, err)
// 	}
// 	defer rows.Close()

// 	var countries []Country
// 	for rows.Next() {
// 		var country Country
// 		if err := rows.Scan(&country.ID, &country.Name); err != nil {
// 			return nil, fmt.Errorf("failed to scan country: %v", err)
// 		}
// 		countries = append(countries, country)
// 	}

// 	if err := rows.Err(); err != nil {
// 		return nil, fmt.Errorf("failed to iterate over rows: %v", err)
// 	}

// 	return countries, nil
// }
