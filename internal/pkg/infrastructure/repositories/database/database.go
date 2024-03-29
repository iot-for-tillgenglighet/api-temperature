package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/iot-for-tillgenglighet/api-temperature/internal/pkg/infrastructure/logging"
	"github.com/iot-for-tillgenglighet/api-temperature/internal/pkg/infrastructure/repositories/models"
)

//Datastore is an interface that is used to inject the database into different handlers to improve testability
type Datastore interface {
	AddTemperatureMeasurement(device *string, latitude, longitude, temp float64, water bool, when string) (*models.Temperature, error)
	GetLatestTemperatures() ([]models.Temperature, error)
	GetTemperaturesNearPoint(latitude, longitude float64, distance, resultLimit uint64) ([]models.Temperature, error)
	GetTemperaturesWithinRect(latitude0, longitude0, latitude1, longitude1 float64, resultLimit uint64) ([]models.Temperature, error)
}

var dbCtxKey = &databaseContextKey{"database"}

type databaseContextKey struct {
	name string
}

// Middleware packs a pointer to the datastore into context
func Middleware(db Datastore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), dbCtxKey, db)

			// and call the next with our new context
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

//GetFromContext extracts the database wrapper, if any, from the provided context
func GetFromContext(ctx context.Context) (Datastore, error) {
	db, ok := ctx.Value(dbCtxKey).(Datastore)
	if ok {
		return db, nil
	}

	return nil, errors.New("failed to decode database from context")
}

type myDB struct {
	impl *gorm.DB
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

//ConnectorFunc is used to inject a database connection method into NewDatabaseConnection
type ConnectorFunc func() (*gorm.DB, error)

//NewPostgreSQLConnector opens a connection to a postgresql database
func NewPostgreSQLConnector() ConnectorFunc {
	dbHost := os.Getenv("TEMPERATURE_DB_HOST")
	username := os.Getenv("TEMPERATURE_DB_USER")
	dbName := os.Getenv("TEMPERATURE_DB_NAME")
	password := os.Getenv("TEMPERATURE_DB_PASSWORD")
	sslMode := getEnv("TEMPERATURE_DB_SSLMODE", "require")

	dbURI := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=%s password=%s", dbHost, username, dbName, sslMode, password)

	return func() (*gorm.DB, error) {
		for {
			log.Printf("Connecting to database host %s ...\n", dbHost)
			db, err := gorm.Open(postgres.Open(dbURI), &gorm.Config{})
			if err != nil {
				log.Fatalf("Failed to connect to database %s \n", err)
				time.Sleep(3 * time.Second)
			} else {
				return db, nil
			}
		}
	}
}

//NewSQLiteConnector opens a connection to a local sqlite database
func NewSQLiteConnector() ConnectorFunc {
	return func() (*gorm.DB, error) {
		db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})

		if err == nil {
			db.Exec("PRAGMA foreign_keys = ON")
		}

		return db, err
	}
}

//NewDatabaseConnection initializes a new connection to the database and wraps it in a Datastore
func NewDatabaseConnection(log logging.Logger, connect ConnectorFunc) (Datastore, error) {
	impl, err := connect()
	if err != nil {
		return nil, err
	}

	db := &myDB{
		impl: impl.Debug(),
	}

	db.impl.AutoMigrate(&models.Temperature{})

	if db.impl.Migrator().HasIndex(&models.Temperature{}, "idx_device_timestamp") {
		db.impl.Migrator().DropIndex(&models.Temperature{}, "idx_device_timestamp")
	}

	return db, nil
}

//AddTemperatureMeasurement takes a device, position and a temp and adds a record to the database
func (db *myDB) AddTemperatureMeasurement(device *string, latitude, longitude, temp float64, water bool, when string) (*models.Temperature, error) {

	ts, err := time.Parse(time.RFC3339Nano, when)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp from %s : (%s)", when, err.Error())
	}

	measurement := &models.Temperature{
		Latitude:   latitude,
		Longitude:  longitude,
		Temp:       float32(temp),
		Water:      water,
		Timestamp2: ts,
	}

	if device != nil {
		measurement.Device = *device
	}

	db.impl.Create(measurement)

	return measurement, nil
}

//GetLatestTemperatures returns the most recent value for all temp sensors that
//have reported a value during the last 6 hours
func (db *myDB) GetLatestTemperatures() ([]models.Temperature, error) {
	// Get temperatures from the last 6 hours
	queryStart := time.Now().UTC().Add(time.Hour * -6)

	latestTemperatures := []models.Temperature{}
	db.impl.Table("temperatures").Select("DISTINCT ON (device) *").Where("timestamp2 > ?", queryStart).Order("device, timestamp2 desc").Find(&latestTemperatures)
	return latestTemperatures, nil
}

func (db *myDB) GetTemperaturesNearPoint(latitude, longitude float64, distance, resultLimit uint64) ([]models.Temperature, error) {
	// Make a crude estimation of the coordinate offset based on the distance
	d := float64(distance)
	lat_delta := (180.0 / math.Pi) * (d / 6378137.0)
	lon_delta := (180.0 / math.Pi) * (d / 6378137.0) / math.Cos(math.Pi/180.0*latitude)

	nw_lat := latitude + lat_delta
	nw_lon := longitude - lon_delta
	se_lat := latitude - lat_delta
	se_lon := longitude + lon_delta

	// TODO: This is not correct, but a good enough first approximation for the MVP. We should make use of PostGIS
	// and do a correct search for matches within a radius. Not within a "square" like this.
	return db.GetTemperaturesWithinRect(nw_lat, nw_lon, se_lat, se_lon, resultLimit)
}

func (db *myDB) GetTemperaturesWithinRect(nw_lat, nw_lon, se_lat, se_lon float64, resultLimit uint64) ([]models.Temperature, error) {
	temperatures := []models.Temperature{}

	result := db.impl.Where(
		"latitude > ? AND latitude < ? AND longitude > ? AND longitude < ?",
		se_lat, nw_lat, nw_lon, se_lon,
	).Limit(int(resultLimit)).Order("timestamp2 desc").Find(&temperatures)

	if result.Error != nil {
		return nil, result.Error
	}

	return temperatures, nil
}
