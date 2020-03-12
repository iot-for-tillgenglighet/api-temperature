// THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.
package graphql

import (
	"context"
	"math"

	"github.com/iot-for-tillgenglighet/api-temperature/pkg/database"
	"github.com/iot-for-tillgenglighet/api-temperature/pkg/models"
)

type Resolver struct{}

func (r *entityResolver) FindDeviceByID(ctx context.Context, id string) (*Device, error) {
	return &Device{ID: id}, nil
}

func convertDatabaseRecordToGQL(measurement *models.Temperature) *Temperature {
	if measurement != nil {
		temp := &Temperature{
			From: &Origin{
				Pos: &WGS84Position{
					Lat: measurement.Latitude,
					Lon: measurement.Longitude,
				},
				Device: &Device{
					ID: measurement.Device,
				},
			},
			When: measurement.Timestamp,
			Temp: math.Round(float64(measurement.Temp*10)) / 10,
		}

		return temp
	}

	return nil
}

func (r *queryResolver) Temperatures(ctx context.Context) ([]*Temperature, error) {
	temperatures, err := database.GetLatestTemperatures()

	if err != nil {
		panic("Failed to query latest temperatures.")
	}

	tempcount := len(temperatures)

	if tempcount == 0 {
		return []*Temperature{}, nil
	}

	gqltemps := make([]*Temperature, 0, tempcount)

	for _, v := range temperatures {
		gqltemps = append(gqltemps, convertDatabaseRecordToGQL(&v))
	}

	return gqltemps, nil
}

func (r *Resolver) Entity() EntityResolver { return &entityResolver{r} }
func (r *Resolver) Query() QueryResolver   { return &queryResolver{r} }

type entityResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
