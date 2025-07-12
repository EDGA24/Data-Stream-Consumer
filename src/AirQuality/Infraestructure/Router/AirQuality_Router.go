package router

import (
	controller "DataConsumer/src/AirQuality/Infraestructure/Controller"
	"github.com/gin-gonic/gin"
)

func RegisterAirQualitySensorRoutes(apiWithJSON, airQualitySensorController *controller.AirQualityController) {
    airQualitySensorGroup := apiWithJSON.Group("/airqualitysensor")
    {
        airQualitySensorGroup.GET("/", airQualitySensorController.GetAirQualityData)

        airQualitySensorGroup.POST("/", airQualitySensorController.SaveAirQualityData)

        airQualitySensorGroup.GET("/ws/handshake/air", airQualitySensorController.HandleWebSocketConnection)
    }
}