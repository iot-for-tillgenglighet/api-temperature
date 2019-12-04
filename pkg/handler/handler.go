package handler

import (
	"encoding/json"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"

	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/iot-for-tillgenglighet/api-temperature/pkg/database"
	"github.com/iot-for-tillgenglighet/api-temperature/pkg/models"
)

func Router() {

	router := mux.NewRouter()

	router.HandleFunc("/api/temperature/{sensor}", handleTemperatureRequest).Methods("GET")
	router.HandleFunc("/api/temperature/{sensor}/{startdate}/{enddate}", getTemperatureHistory).Methods("GET")

	port := os.Getenv("TEMPERATURE_API_PORT")
	if port == "" {
		port = "8880"
	}

	log.Printf("Starting api-temperature on port %s.\n", port)

	err := http.ListenAndServe(":"+port, handlers.LoggingHandler(os.Stdout, router))
	if err != nil {
		log.Print(err)
	}

}

func handleTemperatureRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sensor := vars["sensor"]

	temp := &models.Temperature{}
	database.GetDB().Limit(1).Table("temperatures").Where("device = ?", sensor).Order("timestamp desc").Find(temp)

	if temp.ID == 0 {
		http.Error(w, "No temperature reported for that device", http.StatusNotFound)
		return
	}

	gurka, err := json.MarshalIndent(temp, "", " ")
	if err != nil {

		http.Error(w, "Marshal problem: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(gurka)

}

//kopierat den ovan för att ha en mall tillsvidare
func getTemperatureHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sensor := vars["sensor"]
	startdate := vars["startdate"]
	enddate := vars["enddate"]

	temp := &models.Temperature{}
	database.GetDB().Table("temperatures").Where("device = ? AND timestamp BETWEEN ? AND ?", sensor, startdate, enddate).Find(temp)

	if temp.ID == 0 {
		http.Error(w, "No temperature reported for that device", http.StatusNotFound)
		return
	}

	temperature, err := json.MarshalIndent(temp, "", " ")
	if err != nil {

		http.Error(w, "Marshal problem: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(temperature)

}
