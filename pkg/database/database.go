package database

import (
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/iot-for-tillgenglighet/api-temperature/pkg/models"
)

var db *gorm.DB

func GetDB() *gorm.DB {
	return db
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func ConnectToDB() {

	dbHost := os.Getenv("TEMPERATURE_DB_HOST")
	username := os.Getenv("TEMPERATURE_DB_USER")
	dbName := os.Getenv("TEMPERATURE_DB_NAME")
	password := os.Getenv("TEMPERATURE_DB_PASSWORD")
	sslMode := getEnv("TEMPERATURE_DB_SSLMODE", "require")

	dbURI := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=%s password=%s", dbHost, username, dbName, sslMode, password)

	for {
		log.Printf("Connecting to database host %s ...\n", dbHost)
		conn, err := gorm.Open("postgres", dbURI)
		if err != nil {
			log.Fatalf("Failed to connect to database %s \n", err)
			time.Sleep(3 * time.Second)
		} else {
			db = conn
			db.Debug().AutoMigrate(&models.Temperature{})
			return
		}
		defer conn.Close()
	}
}

//GetLatestTemperatures returns the most recent value for all sensors that have reported
//a value during the last 24 hours
func GetLatestTemperatures() ([]models.Temperature, error) {
	// Get temperatures from the last 24 hours
	queryStart := time.Now().UTC().AddDate(0, 0, -1).Format(time.RFC3339)

	latestTemperatures := []models.Temperature{}
	GetDB().Table("temperatures").Select("DISTINCT ON (device) *").Where("timestamp > ?", queryStart).Order("device, timestamp desc").Find(&latestTemperatures)
	return latestTemperatures, nil
}
