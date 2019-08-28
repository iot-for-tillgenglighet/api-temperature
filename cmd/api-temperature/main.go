package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func handleTemperatureRequest(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Temperature requests not implemented yet!", http.StatusNotImplemented)
}

func main() {

	// TODO: Setup connection to database ...

	// TODO: Setup connection to message queue ...

	router := mux.NewRouter()

	router.HandleFunc("/api/temperature/{sensor}", handleTemperatureRequest).Methods("GET")

	port := os.Getenv("TEMPERATURE_API_PORT")
	if port == "" {
		port = "8880"
	}

	fmt.Printf("Starting api-temperature on port %s.\n", port)

	err := http.ListenAndServe(":"+port, handlers.LoggingHandler(os.Stdout, router))
	if err != nil {
		fmt.Print(err)
	}
}
