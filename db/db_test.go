package db

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func AddDummyData() error {

	tx, err := db.Begin()
	if err != nil {
		tx.Rollback()
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	usd := Currency{Symbol: "USD"}
	aed := Currency{Symbol: "AED"}
	eur := Currency{Symbol: "EUR"}
	if err := CreateCurrency(ctx, tx, &usd); err != nil {
		tx.Rollback()
		return err
	}
	if err := CreateCurrency(ctx, tx, &aed); err != nil {
		tx.Rollback()
		return err
	}
	if err := CreateCurrency(ctx, tx, &eur); err != nil {
		tx.Rollback()
		return err
	}

	uae := Country{
		Name: "United Arab Emirates",
		Code: "AE",
	}

	usa := Country{
		Name: "United States of America",
		Code: "US",
	}

	if err := CreateCountry(ctx, tx, &uae); err != nil {
		tx.Rollback()
		return err
	}

	if err := CreateCountry(ctx, tx, &usa); err != nil {
		tx.Rollback()
		return err
	}

	if _, err := tx.ExecContext(ctx, `INSERT INTO country_currency (country_id, currency_id) VALUES ($1, $2)`, usa.ID, usd.ID); err != nil {
		tx.Rollback()
		return err
	}

	if _, err := tx.ExecContext(ctx, `INSERT INTO country_currency (country_id, currency_id) VALUES ($1, $2)`, uae.ID, aed.ID); err != nil {
		tx.Rollback()
		return err
	}

	if _, err := tx.ExecContext(ctx, `INSERT INTO country_currency (country_id, currency_id) VALUES ($1, $2)`, uae.ID, usd.ID); err != nil {
		tx.Rollback()
		return err
	}

	CreateUser(ctx, tx, &User{
		Username:  "johnsmith",
		Email:     "john.smith@example.com",
		CountryID: uae.ID,
	})

	gate1 := Gateway{
		Name:                "Gateway 1",
		DataFormatSupported: "application/json",
	}
	gate2 := Gateway{
		Name:                "Gateway 2",
		DataFormatSupported: "application/xml",
	}
	CreateGateway(ctx, tx, &gate1)
	CreateGateway(ctx, tx, &gate2)

	if _, err := tx.Exec("INSERT INTO gateway_countries (country_id, gateway_id) VALUES ($1, $2)", uae.ID, gate1.ID); err != nil {
		tx.Rollback()
		return err
	}
	if _, err := tx.Exec("INSERT INTO gateway_countries (country_id, gateway_id) VALUES ($1, $2)", uae.ID, gate2.ID); err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	return nil
}

func TestMain(m *testing.M) {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")

	dbURL := "postgres://" + dbUser + ":" + dbPassword + "@" + dbHost + ":" + dbPort + "/postgres?sslmode=disable"

	InitializeDB(dbURL)
	if _, err := db.Exec("DROP DATABASE IF EXISTS " + dbName); err != nil {
		log.Fatalln("Could not drop database:", err)
	}
	if _, err := db.Exec("CREATE DATABASE " + dbName); err != nil {
		log.Fatalln("Could not create database:", err)
	}

	db.Close()
	dbURL = "postgres://" + dbUser + ":" + dbPassword + "@" + dbHost + ":" + dbPort + "/" + dbName + "?sslmode=disable"
	InitializeDB(dbURL)

	fp, _ := filepath.Abs("init.sql")
	initScript, err := os.ReadFile(fp)
	if err != nil {
		log.Fatalln("Could not read init.sql:", err)
	}
	if _, err := db.Exec(string(initScript)); err != nil {
		log.Fatalln("Could not run init.sql:", err)
	}

	if err := AddDummyData(); err != nil {
		log.Fatalln("Could not add dummy data:", err)
	}

	os.Exit(m.Run())
}
