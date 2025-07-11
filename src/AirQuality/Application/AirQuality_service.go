package application

import (
	entities "DataConsumer/src/AirQuality/Domain/Entities"
	repositories "DataConsumer/src/AirQuality/Domain/Repositories"
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type AirQualityService struct {
	repo      repositories.AirQualityRepository
	clients   map[*websocket.Conn]bool
	broadcast chan *entities.AirQualitySensor
	mu        sync.Mutex
}

func NewAirQualityService(repo repositories.AirQualityRepository) *AirQualityService {
	return &AirQualityService{
		repo:      repo,
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan *entities.AirQualitySensor),
	}
}

func (s *AirQualityService) SaveAirQualityData(sensor *entities.AirQualitySensor) error {
	// ✅ Validación 1: Sensor nulo (ya existía)
	if sensor == nil {
		return errors.New("los datos del sensor de calidad del aire son nulos")
	}

	// ✅ Validación 2: SensorID vacío - AGREGADO
	if strings.TrimSpace(sensor.SensorID) == "" {
		return errors.New("SensorID es requerido y no puede estar vacío")
	}

	// ✅ Validación 3: CO2PPM fuera de rango - AGREGADO
	if sensor.CO2PPM < 0 || sensor.CO2PPM > 5000 {
		return errors.New("CO2PPM debe estar entre 0 y 5000 ppm")
	}

	// ✅ Validación 4: Air_level fuera de rango - AGREGADO
	if sensor.Air_level < 0 || sensor.Air_level > 100 {
		return errors.New("Air_level debe estar entre 0 y 100 por ciento")
	}

	// ✅ Validación 5: Timestamp vacío y formato - AGREGADO
	if strings.TrimSpace(sensor.Timestamp) == "" {
		return errors.New("Timestamp es requerido")
	}

	if _, err := time.Parse(time.RFC3339, sensor.Timestamp); err != nil {
		return errors.New("Timestamp debe tener formato RFC3339 válido")
	}

	// Continuar con lógica original
	if err := s.repo.SaveAirQualityData(sensor); err != nil {
		log.Printf("Error al guardar los datos de calidad del aire: %v", err)
		return err
	}

	s.broadcast <- sensor
	return nil
}

func (s *AirQualityService) GetAirQualityData() ([]*entities.AirQualitySensor, error) {
	return s.repo.GetAllAirQualityData()
}

func (s *AirQualityService) GetAllAirQualityData() ([]*entities.AirQualitySensor, error) {
	return s.repo.GetAllAirQualityData()
}

func (s *AirQualityService) HandleWebSocketConnection(conn *websocket.Conn) {
	s.mu.Lock()
	s.clients[conn] = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.clients, conn)
		s.mu.Unlock()
		conn.Close()
	}()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

func (s *AirQualityService) StartBroadcasting() {
	for sensorData := range s.broadcast {
		s.mu.Lock()
		for client := range s.clients {
			if err := client.WriteJSON(sensorData); err != nil {
				client.Close()
				delete(s.clients, client)
			}
		}
		s.mu.Unlock()
	}
}
