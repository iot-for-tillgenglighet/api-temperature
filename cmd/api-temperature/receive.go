package main

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"

	"github.com/streadway/amqp"

	"github.com/iot-for-tillgenglighet/api-temperature/pkg/database"
	"github.com/iot-for-tillgenglighet/api-temperature/pkg/models"
	"github.com/iot-for-tillgenglighet/messaging-golang/pkg/messaging/telemetry"
)

func receiveTemperature(msg amqp.Delivery) {

	log.Info("Message received from queue: " + string(msg.Body))
	telTemp := &telemetry.Temperature{}
	err := json.Unmarshal(msg.Body, telTemp)
	if err != nil {
		log.Error("Unmarshal problem")
		return
	}

	if telTemp.Timestamp == "" {
		log.Info("Ignored temperature message with an empty timestamp.")
		return
	}

	newtemp := &models.Temperature{
		Device:    telTemp.Origin.Device,
		Latitude:  telTemp.Origin.Latitude,
		Longitude: telTemp.Origin.Longitude,
		Temp:      float32(telTemp.Temp),
		Timestamp: telTemp.Timestamp,
	}

	database.GetDB().Create(newtemp)
}
