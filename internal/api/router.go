package api

import (
	"net/http"
	"payment-gateway/internal/models"
	"time"

	"github.com/gorilla/mux"
)

func SetupRouter() *mux.Router {
	router := mux.NewRouter()

	router.Handle("/withdrawal", BodyParseAndTimeout[models.WithdrawalRequest](time.Second*5)(http.HandlerFunc(WithdrawalPostHandler))).Methods(http.MethodPost)
	router.Handle("/withdrawal", BodyParseAndTimeout[models.WithdrawalPutRequest](time.Second*5)(http.HandlerFunc(WithdrawalPutHandler))).Methods(http.MethodPut)
	router.Handle("/withdrawal/{id}", http.HandlerFunc(WithdrawalGetHandler)).Methods(http.MethodGet)

	router.Handle("/deposit", BodyParseAndTimeout[models.DepositRequest](time.Second*5)(http.HandlerFunc(DepositPostHandler))).Methods(http.MethodPost)
	router.Handle("/deposit", BodyParseAndTimeout[models.DepositPutRequest](time.Second*5)(http.HandlerFunc(DepositPutHandler))).Methods(http.MethodPut)
	router.Handle("/deposit/{id}", http.HandlerFunc(DepositGetHandler)).Methods(http.MethodGet)

	return router

}
