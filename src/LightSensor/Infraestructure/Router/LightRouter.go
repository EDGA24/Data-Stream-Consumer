package router

import (
	controller "DataConsumer/src/LightSensor/Infraestructure/Controller"
	"github.com/gin-gonic/gin"
)

func RegisterLightSensorRoutes(apiWithJSON, lightSensorController *controller.LightController) {
	lightSensorGroup := apiWithJSON.Group("/lightsensor")
	{
		lightSensorGroup.GET("/", lightSensorController.GetLightData)

		lightSensorGroup.POST("/", lightSensorController.SaveLightData)

		lightSensorGroup.GET("/ws/handshake/light", lightSensorController.HandleWebSocket)

		
	}
}