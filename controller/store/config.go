package store

import (
	"context"
	"encoding/json"
	"os"
	"stupid-caldaia/controller/graph/model"

	"github.com/redis/go-redis/v9"
)

const (
	ConfigEnvVar      = "CONFIG_PATH"
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

func (c *Config) CreateObjects(ctx context.Context) (*redis.Client, map[string]*model.Sensor) {
	client := redis.NewClient(&c.Redis)
	sensors := make(map[string]*model.Sensor)
	for _, sensorOptions := range c.Sensors {
		sensor, err := model.NewSensor(ctx, client, &sensorOptions)
		if err != nil {
			panic(err)
		}
		sensors[sensor.Id] = sensor
	}
	return client, sensors
}
