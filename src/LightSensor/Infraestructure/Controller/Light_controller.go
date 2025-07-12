package controller

import (
	application "DataConsumer/src/LightSensor/Application"
	entities "DataConsumer/src/LightSensor/Domain/Entities"
	repositories "DataConsumer/src/LightSensor/Domain/Repositories"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type LightController struct {
	service *application.LightService
}

func NewLightController(repo repositories.LightRepository) *LightController {
	service := application.NewLightService(repo)
	return &LightController{
		service: service,
	}
}

func (c *LightController) SaveLightData(ctx *gin.Context) {
	var lightData entities.LightSensor
	
	if err := ctx.ShouldBindJSON(&lightData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Error al decodificar los datos"})
		return
	}

	// Validar los datos del sensor
	if lightData.SensorID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "El campo 'sensor_id' es obligatorio"})
		return
	}
	if lightData.Nivel < 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "El campo 'nivel' debe ser mayor o igual a 0"})
		return
	}
	if lightData.Timestamp == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "El campo 'timestamp' es obligatorio"})
		return
	}

	if err := c.service.SaveLightData(&lightData); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error al guardar los datos en la base de datos"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Datos de luz guardados correctamente"})
}


func (c *LightController) GetLightData(ctx *gin.Context) {
	lightData, err := c.service.GetLightData()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, lightData)
}

func (c *LightController) HandleWebSocket(ctx *gin.Context) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true }, 
	}

	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error al establecer la conexión WebSocket"})
		return
	}

	c.service.HandleWebSocketConnection(conn)
}

func (c *LightController) StartBroadcasting() {
	go c.service.StartBroadcasting()
}
