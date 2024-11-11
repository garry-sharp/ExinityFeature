package db

import (
	"context"
	"testing"
)

func TestCurrencySupportedInCountry(t *testing.T) {

	if usdUAE, err := CurrencySupportedInCountry(context.Background(), db, "USD", 1); err != nil || !usdUAE {
		t.Errorf("Expected USD to be supported in UAE, got value %v and error %v", usdUAE, err)
	}
	if aedUAE, err := CurrencySupportedInCountry(context.Background(), db, "AED", 1); err != nil || !aedUAE {
		t.Errorf("Expected AED to be supported in UAE, got value %v and error %v", aedUAE, err)
	}
	if eurUAE, err := CurrencySupportedInCountry(context.Background(), db, "EUR", 1); err != nil || eurUAE {
		t.Errorf("Expected EUR to not be supported in UAE, got value %v and error %v", eurUAE, err)
	}
	if nonexist, err := CurrencySupportedInCountry(context.Background(), db, "XXX", 5000); err != nil || nonexist {
		t.Errorf("Expected AED to be supported in UAE, got value %v and error %v", nonexist, err)
	}

}

func TestGetSupportedCurrenciesFromCountry(t *testing.T) {
	currencies, err := GetSupportedCurrenciesFromCountry(context.Background(), db, 1)
	if err != nil {
		t.Fatalf("Error getting supported currencies: %v", err)
	}

	if currencies[0].Symbol != "USD" || currencies[1].Symbol != "AED" {
		t.Fatalf("Expected currencies not found")
	}
}

func TestGetRandomGateway(t *testing.T) {
	gateway1, err := GetRandomGateway(context.Background(), db, 1, "USD", "application/json")
	if err != nil {
		t.Fatalf("Error getting random gateway: %v", err)
	}
	gateway2, err := GetRandomGateway(context.Background(), db, 1, "USD", "application/xml")
	if err != nil {
		t.Fatalf("Error getting random gateway: %v", err)
	}
	if gateway1.Name != "Gateway 1" {
		t.Fatalf("Expected Gateway 1, got %s", gateway1.Name)
	}
	if gateway2.Name != "Gateway 2" {
		t.Fatalf("Expected Gateway 2, got %s", gateway2.Name)
	}
}
