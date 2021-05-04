package main

import (
	"encoding/json"
	"math"

	"github.com/streadway/amqp"

	"github.com/iot-for-tillgenglighet/api-temperature/internal/pkg/infrastructure/logging"
	"github.com/iot-for-tillgenglighet/api-temperature/internal/pkg/infrastructure/repositories/database"
	"github.com/iot-for-tillgenglighet/messaging-golang/pkg/messaging"
	"github.com/iot-for-tillgenglighet/messaging-golang/pkg/messaging/telemetry"
)

func createTemperatureReceiver(log logging.Logger, db database.Datastore) messaging.TopicMessageHandler {
	return func(msg amqp.Delivery) {

		log.Infof("Message received from queue: %s", string(msg.Body))

		telTemp := &telemetry.Temperature{}
		err := json.Unmarshal(msg.Body, telTemp)

		if err != nil {
			log.Error("Failed to unmarshal message")
			return
		}

		if telTemp.Timestamp == "" {
			log.Info("Ignored temperature message with an empty timestamp.")
			return
		}

		db.AddTemperatureMeasurement(
			&telTemp.Origin.Device,
			telTemp.Origin.Latitude, telTemp.Origin.Longitude,
			float64(math.Round(telTemp.Temp*10)/10),
			false,
			telTemp.Timestamp,
		)
	}
}

func createWaterTempReceiver(log logging.Logger, db database.Datastore) messaging.TopicMessageHandler {
	return func(msg amqp.Delivery) {

		log.Infof("Message received from queue: %s", string(msg.Body))

		telTemp := &telemetry.WaterTemperature{}
		err := json.Unmarshal(msg.Body, telTemp)

		if err != nil {
			log.Error("Failed to unmarshal message")
			return
		}

		if telTemp.Timestamp == "" {
			log.Info("Ignored water temperature message with an empty timestamp.")
			return
		}

		db.AddTemperatureMeasurement(
			&telTemp.Origin.Device,
			telTemp.Origin.Latitude, telTemp.Origin.Longitude,
			float64(math.Round(telTemp.Temp*10)/10),
			true,
			telTemp.Timestamp,
		)
	}
}
