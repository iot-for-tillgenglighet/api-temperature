package handler

import (
	"compress/flate"
	"math"
	"net/http"
	"os"
	"strings"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	gql "github.com/iot-for-tillgenglighet/api-temperature/internal/pkg/graphql"
	"github.com/iot-for-tillgenglighet/api-temperature/pkg/database"
	"github.com/iot-for-tillgenglighet/api-temperature/pkg/models"
	"github.com/iot-for-tillgenglighet/ngsi-ld-golang/pkg/datamodels/fiware"
	ngsi "github.com/iot-for-tillgenglighet/ngsi-ld-golang/pkg/ngsi-ld"
	"github.com/iot-for-tillgenglighet/ngsi-ld-golang/pkg/ngsi-ld/types"

	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
)

//RequestRouter wraps the concrete router implementation
type RequestRouter struct {
	impl *chi.Mux
}

func (router *RequestRouter) addGraphQLHandlers(db database.Datastore) {
	gqlServer := handler.New(gql.NewExecutableSchema(gql.Config{Resolvers: &gql.Resolver{}}))
	gqlServer.AddTransport(&transport.POST{})
	gqlServer.Use(extension.Introspection{})

	// TODO: Investigate some way to use closures instead of context even for GraphQL handlers
	router.impl.Use(database.Middleware(db))

	router.impl.Handle("/api/graphql/playground", playground.Handler("GraphQL playground", "/api/graphql"))
	router.impl.Handle("/api/graphql", gqlServer)
}

func (router *RequestRouter) addNGSIHandlers(contextRegistry ngsi.ContextRegistry) {
	router.Get("/ngsi-ld/v1/entities", ngsi.NewQueryEntitiesHandler(contextRegistry))
}

//Get accepts a pattern that should be routed to the handlerFn on a GET request
func (router *RequestRouter) Get(pattern string, handlerFn http.HandlerFunc) {
	router.impl.Get(pattern, handlerFn)
}

//newRequestRouter creates and returns a new router wrapper
func newRequestRouter() *RequestRouter {
	router := &RequestRouter{impl: chi.NewRouter()}

	router.impl.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		Debug:            false,
	}).Handler)

	// Enable gzip compression for ngsi-ld responses
	compressor := middleware.NewCompressor(flate.DefaultCompression, "application/json", "application/ld+json")
	router.impl.Use(compressor.Handler)
	router.impl.Use(middleware.Logger)

	return router
}

func createRequestRouter(contextRegistry ngsi.ContextRegistry, db database.Datastore) *RequestRouter {
	router := newRequestRouter()

	router.addGraphQLHandlers(db)
	router.addNGSIHandlers(contextRegistry)

	return router
}

//CreateRouterAndStartServing creates a request router, registers all handlers and starts serving requests
func CreateRouterAndStartServing(db database.Datastore) {

	contextRegistry := ngsi.NewContextRegistry()
	ctxSource := contextSource{db: db}
	contextRegistry.Register(ctxSource)

	router := createRequestRouter(contextRegistry, db)

	port := os.Getenv("TEMPERATURE_API_PORT")
	if port == "" {
		port = "8880"
	}

	log.Printf("Starting api-temperature on port %s.\n", port)

	log.Fatal(http.ListenAndServe(":"+port, router.impl))
}

type contextSource struct {
	db database.Datastore
}

func convertDatabaseRecordToWeatherObserved(r *models.Temperature) *fiware.WeatherObserved {
	if r != nil {
		entity := fiware.NewWeatherObserved("temperature:"+r.Device, r.Latitude, r.Longitude, r.Timestamp)
		entity.Temperature = types.NewNumberProperty(math.Round(float64(r.Temp*10)) / 10)
		return entity
	}

	return nil
}

func (cs contextSource) GetEntities(query ngsi.Query, callback ngsi.QueryEntitiesCallback) error {

	var temperatures []models.Temperature
	var err error

	temperatures, err = cs.db.GetLatestTemperatures()

	if err == nil {
		for _, v := range temperatures {
			err = callback(convertDatabaseRecordToWeatherObserved(&v))
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
	return strings.HasPrefix(entityID, "urn:ngsi-ld:WeatherObserved:")
}

func (cs contextSource) ProvidesType(typeName string) bool {
	return typeName == "WeatherObserved"
}

func (cs contextSource) UpdateEntityAttributes(entityID string, patch ngsi.Patch) error {
	return nil
}
