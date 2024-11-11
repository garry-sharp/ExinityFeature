package services

import (
	"github.com/shopspring/decimal"
)

func CurrencyAmountIsValid(amount decimal.Decimal) bool {
	if amount.LessThan(decimal.NewFromFloat(0.01)) {
		return false
	}
	if amount.Exponent() < -2 {
		return false
	}
	return true
}
