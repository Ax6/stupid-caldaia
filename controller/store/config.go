package store

import (
	"encoding/json"
	"os"

	"github.com/redis/go-redis/v9"
)

const (
	CONFIG_PATH = "../config.json"
)

type Config struct {
	Sensors []SensorOptions
	Redis  redis.Options
}

func LoadConfig() (Config, error) {
	// Read config file and parse it
	file, err := os.Open(CONFIG_PATH)	
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	jsonParser := json.NewDecoder(file)
	config := Config{}
	err = jsonParser.Decode(&config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}