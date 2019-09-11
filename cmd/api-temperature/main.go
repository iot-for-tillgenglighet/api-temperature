package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type Temperature struct {
	gorm.Model
	Latitude  float64
	Longitude float64
	Device    string `gorm:"unique_index:idx_device_timestamp"`
	Temp      float32
	Timestamp string `gorm:"unique_index:idx_device_timestamp"`
}

var db *gorm.DB

func GetDB() *gorm.DB {
	return db
}

func handleTemperatureRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sensor := vars["sensor"]

	temp := &Temperature{}
	GetDB().Limit(1).Table("temperatures").Where("device = ?", sensor).Order("timestamp desc").Find(temp)

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

func connectToDB() {

	dbHost := os.Getenv("TEMPERATURE_DB_HOST")
	username := os.Getenv("TEMPERATURE_DB_USER")
	dbName := os.Getenv("TEMPERATURE_DB_NAME")
	password := os.Getenv("TEMPERATURE_DB_PASSWORD")

	dbURI := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s", dbHost, username, dbName, password)

	for {
		log.Printf("Connecting to database host %s ...\n", dbHost)
		conn, err := gorm.Open("postgres", dbURI)
		if err != nil {
			log.Fatal("Failed to connect to database: %s \n", err)
			time.Sleep(3 * time.Second)
		} else {
			db = conn
			db.Debug().AutoMigrate(&Temperature{})
			return
		}
		defer conn.Close()
	}
}

func main() {

	log.Info("Starting api-temperature")

	time.Sleep(30 * time.Second)

	connectToDB()

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
