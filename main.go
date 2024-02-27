package main

import (
	"fmt"
	// "log"
	// "os"
	// "runtime/pprof"

	"github.com/matheustavarestrindade/rinha-de-backend-2024-q1.git/internal/app/api"
	buslock "github.com/matheustavarestrindade/rinha-de-backend-2024-q1.git/internal/app/bus"
	"github.com/matheustavarestrindade/rinha-de-backend-2024-q1.git/internal/app/database"
)

func main() {
	fmt.Println("testando app")
    // f, err := os.Create("profile.prof")
    // if err != nil {
    //     log.Fatal(err)
    // }
    // pprof.StartCPUProfile(f)
    //
    // defer pprof.StopCPUProfile()
    // Initialize the desktop bus lock
	buslock.Init()

    // Initialize the database
	defer database.Close()
	database.Connect()

    // Start the API
	api.Start()
}
