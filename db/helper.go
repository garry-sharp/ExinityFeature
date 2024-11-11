package db

import (
	"context"
	"database/sql"
	"fmt"
)

func CurrencySupportedInCountry(ctx context.Context, db *sql.DB, currencySymbol string, countryID int) (bool, error) {
	row := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM country_currency cc LEFT JOIN currencies cu on cc.currency_id = cu.id WHERE cc.country_id = $1 AND cu.symbol = $2", countryID, currencySymbol)
	if row.Err() != nil {
		return false, row.Err()
	}
	cnt := 0
	if err := row.Scan(&cnt); err != nil {
		return false, err
	}

	return cnt > 0, nil
}

func GetSupportedCurrenciesFromCountry(ctx context.Context, db *sql.DB, countryID int) ([]Currency, error) {
	rows, err := db.QueryContext(ctx, "select cu.id, cu.symbol from country_currency cc join currencies cu on cc.currency_id = cu.id where cc.country_id = $1", countryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var currencies []Currency
	for rows.Next() {
		var currency Currency
		if err := rows.Scan(&currency.ID, &currency.Symbol); err != nil {
			return nil, err
		}
		currencies = append(currencies, currency)
	}

	return currencies, nil
}

func GetRandomGateway(ctx context.Context, db *sql.DB, countryID int, currency, dataformat string) (Gateway, error) {
	rows, err := db.Query("SELECT g.id, g.name, g.data_format_supported, g.created_at, g.updated_at FROM gateway_country_currency g WHERE g.country_id = $1 and g.currency_symbol = $2 and g.data_format_supported = $3 ORDER BY random() LIMIT 1", countryID, currency, dataformat)
	if err != nil {
		return Gateway{}, fmt.Errorf("failed to get gateway: %v", err)
	}
	defer rows.Close()
	if !rows.Next() {
		return Gateway{}, fmt.Errorf("no gateway found for country %d, currency %s and format %s", countryID, currency, dataformat)
	}
	var gateway Gateway
	if err := rows.Scan(&gateway.ID, &gateway.Name, &gateway.DataFormatSupported, &gateway.CreatedAt, &gateway.UpdatedAt); err != nil {
		return Gateway{}, fmt.Errorf("failed to scan gateway: %v", err)
	}
	return gateway, nil
}
