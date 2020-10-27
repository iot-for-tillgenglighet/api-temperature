package context

import (
	"errors"
	"math"
	"strings"

	"github.com/iot-for-tillgenglighet/api-temperature/pkg/database"
	"github.com/iot-for-tillgenglighet/api-temperature/pkg/models"
	"github.com/iot-for-tillgenglighet/ngsi-ld-golang/pkg/datamodels/fiware"
	ngsi "github.com/iot-for-tillgenglighet/ngsi-ld-golang/pkg/ngsi-ld"
	"github.com/iot-for-tillgenglighet/ngsi-ld-golang/pkg/ngsi-ld/types"
)

type contextSource struct {
	db database.Datastore
}

//CreateSource instantiates and returns a Fiware ContextSource that wraps the provided db interface
func CreateSource(db database.Datastore) ngsi.ContextSource {
	return &contextSource{db: db}
}

func convertDatabaseRecordToWaterQualityObserved(r *models.Temperature) *fiware.WaterQualityObserved {
	if r != nil {
		entity := fiware.NewWaterQualityObserved("temperature:"+r.Device, r.Latitude, r.Longitude, r.Timestamp)
		entity.Temperature = types.NewNumberProperty(math.Round(float64(r.Temp*10)) / 10)
		return entity
	}

	return nil
}

func convertDatabaseRecordToWeatherObserved(r *models.Temperature) *fiware.WeatherObserved {
	if r != nil {
		entity := fiware.NewWeatherObserved("temperature:"+r.Device, r.Latitude, r.Longitude, r.Timestamp)
		entity.Temperature = types.NewNumberProperty(math.Round(float64(r.Temp*10)) / 10)
		return entity
	}

	return nil
}

func (cs contextSource) CreateEntity(typeName, entityID string, req ngsi.Request) error {
	return errors.New("CreateEntity not supported for type " + typeName)
}

func (cs contextSource) GetEntities(query ngsi.Query, callback ngsi.QueryEntitiesCallback) error {

	var temperatures []models.Temperature
	var err error

	if query == nil {
		return errors.New("GetEntities: query may not be nil")
	}

	includeAirTemperature := false
	includeWaterTemperature := false

	for _, typeName := range query.EntityTypes() {
		if typeName == "WeatherObserved" {
			includeAirTemperature = true
		} else if typeName == "WaterQualityObserved" {
			includeWaterTemperature = true
		}
	}

	if includeAirTemperature == false && includeWaterTemperature == false {
		// No provided type specified, but maybe the caller specified an attribute list instead?
		if queriedAttributesDoNotInclude(query.EntityAttributes(), "temperature") {
			return errors.New("GetEntities called without specifying a type that is provided by this service")
		}

		// Include both entity types as they both hold a temperature value
		includeAirTemperature = true
		includeWaterTemperature = true
	}

	temperatures, err = cs.db.GetLatestTemperatures()

	if err == nil {
		for _, v := range temperatures {
			if v.Water == false && includeAirTemperature {
				err = callback(convertDatabaseRecordToWeatherObserved(&v))
			} else if v.Water == true && includeWaterTemperature {
				err = callback(convertDatabaseRecordToWaterQualityObserved(&v))
			}
			if err != nil {
				break
			}
		}
	}

	return err
}

func (cs contextSource) ProvidesAttribute(attributeName string) bool {
	return attributeName == "temperature"
}

func (cs contextSource) ProvidesEntitiesWithMatchingID(entityID string) bool {
	return strings.HasPrefix(entityID, "urn:ngsi-ld:WeatherObserved:") ||
		strings.HasPrefix(entityID, "urn:ngsi-ld:WaterQualityObserved:")
}

func (cs contextSource) ProvidesType(typeName string) bool {
	return typeName == "WeatherObserved" || typeName == "WaterQualityObserved"
}

func (cs contextSource) UpdateEntityAttributes(entityID string, req ngsi.Request) error {
	return errors.New("UpdateEntityAttributes is not supported by this service")
}

func queriedAttributesDoNotInclude(attributes []string, requiredAttribute string) bool {
	for _, attr := range attributes {
		if attr == requiredAttribute {
			return false
		}
	}

	return true
}
