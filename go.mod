module github.com/iot-for-tillgenglighet/api-temperature

go 1.14

require (
	github.com/99designs/gqlgen v0.11.3
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/iot-for-tillgenglighet/messaging-golang v0.0.0-20201009211140-579335ad3c49
	github.com/iot-for-tillgenglighet/ngsi-ld-golang v0.0.0-20201027142841-6a1bb73c1a6f
	github.com/lib/pq v1.8.0 // indirect
	github.com/mitchellh/mapstructure v1.3.3 // indirect
	github.com/rs/cors v1.7.0
	github.com/sirupsen/logrus v1.7.0
	github.com/streadway/amqp v1.0.0
	github.com/vektah/gqlparser/v2 v2.0.1
	gopkg.in/yaml.v2 v2.3.0 // indirect
	gorm.io/driver/postgres v1.0.5
	gorm.io/gorm v1.20.5
)
