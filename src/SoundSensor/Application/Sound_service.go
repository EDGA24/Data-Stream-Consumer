package application

import (
	entities "DataConsumer/src/SoundSensor/Domain/Entities"
	repositories "DataConsumer/src/SoundSensor/Domain/Repositories"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"sync"
	"time"
	"github.com/gorilla/websocket"
)

type SoundService struct {
	repo      repositories.SoundSensor
	clients   map[*websocket.Conn]bool
	broadcast chan *entities.SoundSensor
	mu        sync.Mutex
}


func NewSoundService(repo repositories.SoundSensor) *SoundService {
	return &SoundService{
		repo:      repo,
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan *entities.SoundSensor),
	}
}

func (s *SoundService) SaveSoundData(sensor *entities.SoundSensor) error {
	// ✅ Validación 1: Sensor nulo (ya existía)
	if sensor == nil {
		return errors.New("los datos del sensor de sonido son nulos")
	}

	// ✅ Validación 2: SensorID vacío - AGREGADO
	if strings.TrimSpace(sensor.SensorID) == "" {
		return errors.New("SensorID es requerido y no puede estar vacío")
	}

	// ✅ Validación 3: Nivel negativo o extremo - AGREGADO
	if sensor.Nivel < 0 || sensor.Nivel > 120 {
		return errors.New("Nivel de sonido debe estar entre 0 y 120 dB")
	}

	// ✅ Validación 4: Timestamp vacío y formato - AGREGADO
	if strings.TrimSpace(sensor.Timestamp) == "" {
		return errors.New("Timestamp es requerido")
	}

	if _, err := time.Parse(time.RFC3339, sensor.Timestamp); err != nil {
		return errors.New("Timestamp debe tener formato RFC3339 válido")
	}

	// Continuar con lógica original
	if err := s.repo.SaveSoundData(sensor); err != nil {
		return err
	}

	s.broadcast <- sensor
	return nil
}

func (s *SoundService) GetSoundData() ([]*entities.SoundSensor, error) {
	return s.repo.GetSoundData()
}

// ✅ CORREGIDO: Implementar métodos que estaban con panic - AGREGADO
func (s *SoundService) AddClient(conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[conn] = true
}

func (s *SoundService) RemoveClient(conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.clients, conn)
	conn.Close()
}

func (s *SoundService) HandleWebSocketConnection(conn *websocket.Conn) {
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

func (s *SoundService) StartBroadcasting() {
	for sensorData := range s.broadcast {
		message, err := json.Marshal(sensorData)
		if err != nil {
			log.Println("Error al serializar datos:", err)
			continue
		}

		s.mu.Lock()
		for client := range s.clients {
			if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Println("Error al enviar mensaje:", err)
				client.Close()
				delete(s.clients, client)
			}
		}
		s.mu.Unlock()
	}
}

