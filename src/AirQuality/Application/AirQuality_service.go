package application
//prueva
import (
	entities "DataConsumer/src/AirQuality/Domain/Entities"
	repositories "DataConsumer/src/AirQuality/Domain/Repositories"
	"errors"
	"log"
	"sync"

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
	if sensor == nil {
		return errors.New("los datos del sensor de calidad del aire son nulos")
	}

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