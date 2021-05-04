package main

import (
	"github.com/iot-for-tillgenglighet/api-temperature/internal/pkg/infrastructure/logging"

	"github.com/iot-for-tillgenglighet/api-temperature/internal/pkg/application"
	"github.com/iot-for-tillgenglighet/api-temperature/internal/pkg/infrastructure/repositories/database"

	"github.com/iot-for-tillgenglighet/messaging-golang/pkg/messaging"
	"github.com/iot-for-tillgenglighet/messaging-golang/pkg/messaging/telemetry"
)

func main() {

	serviceName := "api-temperature"

	log := logging.NewLogger()

	log.Infof("Starting up %s ...", serviceName)

	config := messaging.LoadConfiguration(serviceName)
	messenger, _ := messaging.Initialize(config)

	defer messenger.Close()

	// Make sure that we have a proper connection to the database ...
	db, _ := database.NewDatabaseConnection(log)

	// ... before we start listening for temperature telemetry
	messenger.RegisterTopicMessageHandler(
		(&telemetry.Temperature{}).TopicName(),
		application.NewTemperatureReceiver(log, db),
	)
	messenger.RegisterTopicMessageHandler(
		(&telemetry.WaterTemperature{}).TopicName(),
		application.NewWaterTempReceiver(log, db),
	)

	application.CreateRouterAndStartServing(log, db)
}
