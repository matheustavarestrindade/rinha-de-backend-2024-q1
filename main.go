package main

import (
	"fmt"

	"github.com/matheustavarestrindade/rinha-de-backend-2024-q1.git/internal/app/api"
	buslock "github.com/matheustavarestrindade/rinha-de-backend-2024-q1.git/internal/app/bus"
	"github.com/matheustavarestrindade/rinha-de-backend-2024-q1.git/internal/app/database"
)

func main() {
	fmt.Println("testando app")

    // Initialize the desktop bus lock
	buslock.Init()

    // Initialize the database
	defer database.Close()
	database.Connect()

    // Start the API
	api.Start()
}
