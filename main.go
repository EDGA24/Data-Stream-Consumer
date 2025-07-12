package main

import (
	core "DataConsumer/Core"
	"log"
	"time"

	middleware "DataConsumer/src/middleware"

	airqualityrouter "DataConsumer/src/AirQuality/Infraestructure/Router"
	lightrouter "DataConsumer/src/LightSensor/Infraestructure/Router"
	soundrouter "DataConsumer/src/SoundSensor/Infraestructure/Router"
	temphumidityrouter "DataConsumer/src/TemperatureHumidity/Infraestructure/Router"

	airqualitycontroller "DataConsumer/src/AirQuality/Infraestructure/Controller"
	lightcontroller "DataConsumer/src/LightSensor/Infraestructure/Controller"
	soundcontroller "DataConsumer/src/SoundSensor/Infraestructure/Controller"
	temperaturecontroller "DataConsumer/src/TemperatureHumidity/Infraestructure/Controller"

	airqualityapp "DataConsumer/src/AirQuality/Application"
	lightapp "DataConsumer/src/LightSensor/Application"
	soundapp "DataConsumer/src/SoundSensor/Application"
	temperatureapp "DataConsumer/src/TemperatureHumidity/Application"

	airqualitydb "DataConsumer/src/AirQuality/Infraestructure/Database"
	lightdb "DataConsumer/src/LightSensor/Infraestructure/Database"
	sounddb "DataConsumer/src/SoundSensor/Infraestructure/Database"
	tempdb "DataConsumer/src/TemperatureHumidity/Infraestructure/Database"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	db, err := core.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, 
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	api.Use

	airQualityRepo := airqualitydb.NewAirQualityRepository(db)
	lightRepo := lightdb.NewLightRepository(db)
	temperatureRepo := tempdb.NewTemperatureHumidityRepository(db)
	soundRepo := sounddb.NewSoundRepository(db)

	airQualityService := airqualityapp.NewAirQualityService(airQualityRepo)
	lightService := lightapp.NewLightService(lightRepo)
	temperatureService := temperatureapp.NewTemperatureHumidityService(temperatureRepo)
	soundService := soundapp.NewSoundService(soundRepo)

	go airQualityService.StartBroadcasting()
	go lightService.StartBroadcasting()
	go temperatureService.StartBroadcasting()
	go soundService.StartBroadcasting()

	airQualityCtrl := airqualitycontroller.NewAirQualityController(airQualityService)
	lightCtrl := lightcontroller.NewLightController(lightService)
	temperatureCtrl := temperaturecontroller.NewTemperatureHumidityController(temperatureService)
	soundCtrl := soundcontroller.NewSoundSensorController(soundService)

	// OPCIÓN 2: Crear grupos de rutas con validación específica
	api := router.Group("/api/v1")
	
	// Grupo para rutas que requieren JSON (POST/PUT)
	apiWithJSON := api.Group("")
	apiWithJSON.Use(ValidateContentType())
	apiWithJSON.Use(ValidateRequestBodySize())

	airqualityrouter.RegisterAirQualitySensorRoutes(apiWithJSON, airQualityCtrl)
	lightrouter.RegisterLightSensorRoutes(apiWithJSON, lightCtrl)
	temphumidityrouter.RegisterTemperatureHumidityRoutes(apiWithJSON, temperatureCtrl)
	soundrouter.RegisterSoundSensorRoutes(apiWithJSON, soundCtrl)

	log.Println("Starting server on :8082")
	if err := apiWithJSON.Run(":8082"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}