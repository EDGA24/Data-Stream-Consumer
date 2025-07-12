package router

import (
	controller "DataConsumer/src/SoundSensor/Infraestructure/Controller"

	"github.com/gin-gonic/gin"
)

func RegisterSoundSensorRoutes(apiWithJSON, soundSensorController *controller.SoundSensorController) {
	soundSensorGroup := apiWithJSON.Group("/soundsensor")
	{
		soundSensorGroup.GET("/", soundSensorController.GetSoundData)

		soundSensorGroup.POST("/", soundSensorController.SaveSoundData)

		soundSensorGroup.GET("/ws/handshake/sound", soundSensorController.HandleWebSocket)

		
	}
}
