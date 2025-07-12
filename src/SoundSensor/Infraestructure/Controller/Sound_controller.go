package controller

import (
	application "DataConsumer/src/SoundSensor/Application"
	entities "DataConsumer/src/SoundSensor/Domain/Entities"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type SoundSensorController struct {
	service *application.SoundService
	upgrader websocket.Upgrader
}

func NewSoundSensorController(service *application.SoundService) *SoundSensorController {
	return &SoundSensorController{
		service: service,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true }, 
		},
	}
}

func (c *SoundSensorController) SaveSoundData(ctx *gin.Context) {
	var sensor entities.SoundSensor
	if err := ctx.ShouldBindJSON(&sensor); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Validate sensor data
	if sensor.SensorID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "SensorID is required"})
		return
	}
	if sensor.Nivel < 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Nivel must be non-negative"})
		return
	}
	if sensor.Timestamp == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Timestamp is required"})
		return
	}

	err := c.service.SaveSoundData(&sensor)
	if err != nil {
		log.Println("Error saving sound sensor data:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving sound sensor data"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Sound sensor data saved successfully"})
}

func (c *SoundSensorController) GetSoundData(ctx *gin.Context) {
	sensors, err := c.service.GetSoundData()
	if err != nil {
		log.Println("Error getting sound sensor data:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting sound sensor data"})
		return
	}

	ctx.JSON(http.StatusOK, sensors)
}

func (c *SoundSensorController) HandleWebSocket(ctx *gin.Context) {
	conn, err := c.upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Println("Error upgrading to WebSocket:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error upgrading to WebSocket"})
		return
	}
	defer conn.Close()

	c.service.AddClient(conn)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading WebSocket message:", err)
			break
		}
	}

	c.service.RemoveClient(conn)
}
