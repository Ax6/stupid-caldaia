package store

import (
	"encoding/json"
	"os"
	"stupid-caldaia/controller/graph/model"

	"github.com/redis/go-redis/v9"
)


const (
	ConfigEnvVar = "CONFIG_PATH"
	DefaultConfigPath = "../config.json"
)


type Config struct {
	Sensors []model.SensorOptions
	Redis   redis.Options
}


func LoadConfig() (Config, error) {
	// Read config file and parse it
	configPath := os.Getenv(ConfigEnvVar)
	if configPath == "" {
		configPath = DefaultConfigPath
	}
	file, err := os.Open(configPath)
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
