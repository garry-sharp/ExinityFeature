package services

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestCurrencyAmountIsValid(t *testing.T) {
	inputs := []decimal.Decimal{decimal.NewFromFloat(100.00), decimal.NewFromFloat(100.001), decimal.NewFromFloat(100.0)}
	outputs := []bool{true, false, true}
	for i, input := range inputs {
		if got := CurrencyAmountIsValid(input); got != outputs[i] {
			t.Errorf("Expected %v Received %v", outputs[i], got)
		}
	}
	t.Log("TestCurrencyAmountIsValid passed")
}
