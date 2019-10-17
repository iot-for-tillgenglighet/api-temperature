package main

import (
	"time"

	log "github.com/sirupsen/logrus"

	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/iot-for-tillgenglighet/api-temperature/pkg/database"
	"github.com/iot-for-tillgenglighet/api-temperature/pkg/handler"
)

func main() {

	log.Info("Starting api-temperature")

	time.Sleep(30 * time.Second)

	database.ConnectToDB()

	connection, channel := receiveTemp()

	defer connection.Close()
	defer channel.Close()

	handler.Router()
}
