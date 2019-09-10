package main

import (
	"encoding/json"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/streadway/amqp"
)

type IoTHubMessageOrigin struct {
	Device    string  `json:"device"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type IoTHubMessage struct {
	Origin    IoTHubMessageOrigin `json:"origin"`
	Timestamp string              `json:"timestamp"`
}

type TelemetryTemperature struct {
	IoTHubMessage
	Temp float32 `json:"temp"`
}

func receiveTemp() (*amqp.Connection, *amqp.Channel) {

	connection, err := amqp.Dial("amqp://user:bitnami@rabbitmq:5672/")
	if err != nil {
		time.Sleep(5 * time.Second)
		log.Fatal("Failed to connect to RabbitMQ: " + err.Error())
	}

	log.Info("Connected to RabbitMQ")

	channel, err := connection.Channel()
	if err != nil {
		log.Fatal("Failed to open a channel: " + err.Error())
	}
	log.Info("Opened a channel")

	err = channel.ExchangeDeclare(
		"iot-msg-exchange-topic", //name
		"topic",                  //type
		false,                    //durable
		false,                    //auto-deleted
		false,                    //internal
		false,                    //no-wait
		nil,                      //arguments
	)
	if err != nil {
		log.Fatal("Failed to declare an exchange: " + err.Error())
	}
	log.Info("Declared an exchange")

	q, err := channel.QueueDeclare(
		"",    //name
		false, //durable
		false, //delete when unused
		true,  //exclusive
		false, //no-wait
		nil,   //arguments
	)
	if err != nil {
		log.Fatal("Failed to declare a queue: " + err.Error())
	}
	log.Info("Declared a queue")

	err = channel.QueueBind(
		q.Name,                   //queue name
		"telemetry.temperature",  //routing key
		"iot-msg-exchange-topic", //exchange
		false,
		nil,
	)
	if err != nil {
		log.Fatal("Failed to bind a queue: " + err.Error())
	}
	log.Info("Bound to a queue")

	temps, err := channel.Consume(
		q.Name, //queue
		"",     //consumer
		true,   //auto ack
		false,  //exclusive
		false,  //no local
		false,  //no-wait
		nil,    //args
	)
	if err != nil {
		log.Fatal("Failed to register a consumer: " + err.Error())
	}
	log.Info("Registered a consumer")

	go func() {
		for data := range temps {
			log.Info("Message received from queue: " + string(data.Body))
			telTemp := &TelemetryTemperature{}
			err = json.Unmarshal(data.Body, telTemp)
			if err != nil {
				log.Error("Unmarshal problem")
				continue
			}

			newtemp := &Temperature{
				Device:    telTemp.Origin.Device,
				Latitude:  telTemp.Origin.Latitude,
				Longitude: telTemp.Origin.Longitude,
				Temp:      telTemp.Temp,
			}
			GetDB().Create(newtemp)
		}
	}()

	log.Info(" [*] Waiting for temperatures. To exit press CTRL+C")

	return connection, channel
}
