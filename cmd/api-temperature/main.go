package main

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"api-temperature/pkg"
)

type Temperature struct {
	gorm.Model
	Latitude  float64
	Longitude float64
	Device    string `gorm:"unique_index:idx_device_timestamp"`
	Temp      float32
	Timestamp string `gorm:"unique_index:idx_device_timestamp"`
}

func handleTemperatureRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sensor := vars["sensor"]

	temp := &Temperature{}
	&pkg.GetDB().Limit(1).Table("temperatures").Where("device = ?", sensor).Order("timestamp desc").Find(temp)

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

func main() {

	log.Info("Starting api-temperature")

	time.Sleep(30 * time.Second)

	&pkg.connectToDB()

	connection, channel := receiveTemp()

	defer connection.Close()
	defer channel.Close()

	router := mux.NewRouter()

	router.HandleFunc("/api/temperature/{sensor}", handleTemperatureRequest).Methods("GET")

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
