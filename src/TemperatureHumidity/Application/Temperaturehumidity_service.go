package application

import (
	entities "DataConsumer/src/TemperatureHumidity/Domain/Entities"
	repositories "DataConsumer/src/TemperatureHumidity/Domain/Repositories"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type TemperatureHumidityService struct {
	repo      repositories.TemperatureHumidityRepository
	clients   map[*websocket.Conn]bool
	broadcast chan *entities.TemperatureHumiditySensor
	mu        sync.Mutex
}

func NewTemperatureHumidityService(repo repositories.TemperatureHumidityRepository) *TemperatureHumidityService {
	return &TemperatureHumidityService{
		repo:      repo,
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan *entities.TemperatureHumiditySensor),
	}
}

func (s *TemperatureHumidityService) SaveTemperatureHumidityData(sensor *entities.TemperatureHumiditySensor) error {
	// ✅ Validación 1: Sensor nulo (ya existía)
	if sensor == nil {
		return errors.New("los datos del sensor de temperatura y humedad son nulos")
	}

	// ✅ Validación 2: SensorID vacío - AGREGADO
	if strings.TrimSpace(sensor.SensorID) == "" {
		return errors.New("SensorID es requerido y no puede estar vacío")
	}

	// ✅ Validación 3: Temperatura fuera de rango - AGREGADO
	if sensor.Temperature < -50 || sensor.Temperature > 80 {
		return errors.New("Temperatura debe estar entre -50 y 80 grados Celsius")
	}

	// ✅ Validación 4: Humedad fuera de rango - AGREGADO
	if sensor.Humidity < 0 || sensor.Humidity > 100 {
		return errors.New("Humedad debe estar entre 0 y 100 por ciento")
	}

	// ✅ Validación 5: Timestamp vacío y formato - AGREGADO
	if strings.TrimSpace(sensor.Timestamp) == "" {
		return errors.New("Timestamp es requerido")
	}

	if _, err := time.Parse(time.RFC3339, sensor.Timestamp); err != nil {
		return errors.New("Timestamp debe tener formato RFC3339 válido")
	}

	// Continuar con lógica original
	if err := s.repo.SaveTemperatureHumidityData(sensor); err != nil {
		return err
	}

	s.broadcast <- sensor
	return nil
}

func (s *TemperatureHumidityService) GetTemperatureHumidityData() ([]*entities.TemperatureHumiditySensor, error) {
	return s.repo.GetTemperatureHumidityData()
}

func (s *TemperatureHumidityService) HandleWebSocketConnection(conn *websocket.Conn) {
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

func (s *TemperatureHumidityService) StartBroadcasting() {
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
