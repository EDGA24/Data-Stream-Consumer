package application

import (
	entities "DataConsumer/src/LightSensor/Domain/Entities"
	repositories "DataConsumer/src/LightSensor/Domain/Repositories"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type LightService struct {
	repo      repositories.LightRepository
	clients   map[*websocket.Conn]bool
	broadcast chan *entities.LightSensor
	mu        sync.Mutex
}

func NewLightService(repo repositories.LightRepository) *LightService {
	return &LightService{
		repo:      repo,
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan *entities.LightSensor),
	}
}

func (s *LightService) SaveLightData(light *entities.LightSensor) error {
	// ✅ Validación 1: Sensor nulo (ya existía)
	if light == nil {
		return errors.New("los datos del sensor de luz son nulos")
	}

	// ✅ Validación 2: SensorID vacío - AGREGADO
	if strings.TrimSpace(light.SensorID) == "" {
		return errors.New("SensorID es requerido y no puede estar vacío")
	}

	// ✅ Validación 3: Nivel negativo o extremo - AGREGADO
	if light.Nivel < 0 || light.Nivel > 2000 {
		return errors.New("Nivel de luz debe estar entre 0 y 2000 lux")
	}

	// ✅ Validación 4: Timestamp vacío y formato - AGREGADO
	if strings.TrimSpace(light.Timestamp) == "" {
		return errors.New("Timestamp es requerido")
	}

	if _, err := time.Parse(time.RFC3339, light.Timestamp); err != nil {
		return errors.New("Timestamp debe tener formato RFC3339 válido")
	}

	// Continuar con lógica original
	if err := s.repo.SaveLightData(light); err != nil {
		return err
	}

	s.broadcast <- light
	return nil
}

func (s *LightService) GetLightData() ([]*entities.LightSensor, error) {
	return s.repo.GetLightData()
}

func (s *LightService) HandleWebSocketConnection(conn *websocket.Conn) {
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

func (s *LightService) StartBroadcasting() {
	for lightData := range s.broadcast {
		s.mu.Lock()
		for client := range s.clients {
			if err := client.WriteJSON(lightData); err != nil {
				client.Close()
				delete(s.clients, client)
			}
		}
		s.mu.Unlock()
	}
}