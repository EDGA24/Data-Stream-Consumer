package entities

type TemperatureHumiditySensor struct {
	ID          int     `json:"id"`
	SensorID    string  `json:"sensor_id"`
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Timestamp   string  `json:"timestamp"`
}