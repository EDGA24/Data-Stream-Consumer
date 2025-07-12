package router

import (
	controller "DataConsumer/src/TemperatureHumidity/Infraestructure/Controller"
	"github.com/gin-gonic/gin"
)

func RegisterTemperatureHumidityRoutes(apiWithJSON, temperatureHumidityController *controller.TemperatureHumidityController) {
	temperatureHumidityGroup := apiWithJSON.Group("/temperaturehumidity")
	{
		temperatureHumidityGroup.GET("/", temperatureHumidityController.GetTemperatureHumidityData)

		temperatureHumidityGroup.POST("/", temperatureHumidityController.SaveTemperatureHumidityData)

			temperatureHumidityGroup.GET("/ws/handshake/temperature", temperatureHumidityController.HandleWebSocketConnection)

		
	}
}
