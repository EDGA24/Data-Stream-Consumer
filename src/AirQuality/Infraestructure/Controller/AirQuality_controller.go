package controller

import (
	application "DataConsumer/src/AirQuality/Application"
	entities "DataConsumer/src/AirQuality/Domain/Entities"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AirQualityController struct {
	service *application.AirQualityService
}

func NewAirQualityController(service *application.AirQualityService) *AirQualityController {
	return &AirQualityController{
		service: service,
	}
}

func (c *AirQualityController) SaveAirQualityData(ctx *gin.Context) {
	var sensor entities.AirQualitySensor
     
	if err := ctx.ShouldBindJSON(&sensor); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Error al decodificar los datos"})
		return
	}

	// Validar los datos del sensor
	if sensor.ID <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	if sensor.SensorID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "SensorID es requerido"})
		return
	}
	if sensor.CO2PPM < 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "CO2PPM debe ser mayor o igual a 0"})
		return
	}
	if sensor.Air_level < 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Air_level debe ser mayor o igual a 0"})
		return
	}
	if sensor.Timestamp == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Timestamp es requerido"})
		return
	}

	if err := c.service.SaveAirQualityData(&sensor); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error al guardar los datos en la base de datos"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Datos guardados correctamente"})
}

func (c *AirQualityController) GetAirQualityData(ctx *gin.Context) {
	data, err := c.service.GetAirQualityData()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener los datos"})
		return
	}

	ctx.JSON(http.StatusOK, data)
}

func (c *AirQualityController) HandleWebSocketConnection(ctx *gin.Context) {
	
}