package api

import (
	"fmt"
	"io"

	// "github.com/goccy/go-json"
	"net/http"
	"strconv"
	"time"

	"github.com/pquerna/ffjson/ffjson"

	"github.com/matheustavarestrindade/rinha-de-backend-2024-q1.git/internal/app/database"
)

type TransactionRequest struct {
	Value       int    `json:"valor"`
	Type        string `json:"tipo"`
	Description string `json:"descricao"`
}

func Start() {

	srv := &http.Server{
		Addr:         ":8080",
        ReadTimeout:  5 * time.Second,
        WriteTimeout: 10 * time.Second,
        


	}

	http.HandleFunc("GET /clientes/{clientId}/extrato", func(w http.ResponseWriter, r *http.Request) {
		clientIdParam := r.PathValue("clientId")
		//Check if clientId is valid
		clientId, err := strconv.Atoi(clientIdParam)
		if err != nil || clientId > 5 || clientId < 1 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		jsonBuffer, err := database.GetClientByIdWithTransactions(clientId)
        defer jsonBuffer.Reset()

		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}


		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonBuffer.Bytes())
	})

	http.HandleFunc("POST /clientes/{clientId}/transacoes", func(w http.ResponseWriter, r *http.Request) {
		clientIdParam := r.PathValue("clientId")
		clientId, err := strconv.Atoi(clientIdParam)
		if err != nil || clientId > 5 || clientId < 1 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		trx := TransactionRequest{}
		body, err := io.ReadAll(r.Body)
        if err != nil {
            w.WriteHeader(http.StatusUnprocessableEntity)
            return
        }

		err = ffjson.Unmarshal(body, &trx)

		if err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		if trx.Value <= 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		if trx.Type != "d" && trx.Type != "c" {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		if len(trx.Description) == 0 || len(trx.Description) > 10 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		if trx.Type == "d" {
			withdrawLimit, finalBalance, err := database.CreateClientTransaction(clientId, trx.Type, -1*trx.Value, trx.Description)
			if err != nil {
				w.WriteHeader(http.StatusUnprocessableEntity)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"limite":` + strconv.Itoa(withdrawLimit) + `,"saldo":` + strconv.Itoa(finalBalance) + `}`))
			return
		}

		withdrawLimit, finalBalance, err := database.CreateClientTransaction(clientId, trx.Type, trx.Value, trx.Description)
		if err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"limite":` + strconv.Itoa(withdrawLimit) + `,"saldo":` + strconv.Itoa(finalBalance) + `}`))
	})

	fmt.Println("Server started at :8080")
    httpError := srv.ListenAndServe()
    if httpError != nil {
        fmt.Println(httpError)
    }
}
