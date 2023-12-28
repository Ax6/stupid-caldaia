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
	Boiler  model.BoilerConfig
}

func LoadConfig() (Config, error) {
	// Read config file and parse it
	configPath := os.Getenv(ConfigEnvVar)
	if configPath == "" {
		configPath = DefaultConfigPath
	}
	file, err := os.Open(configPath)
	defer file.Close()
	if err != nil {
		panic(err)
	}

	config := Config{}
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		panic(err)
	}
	return config, nil
}

func (c *Config) CreateObjects(ctx context.Context) (*redis.Client, map[string]*model.Sensor, *model.Boiler) {
	// DB client
	client := redis.NewClient(&c.Redis)

	// Sensors
	sensors := make(map[string]*model.Sensor)
	for _, sensorOptions := range c.Sensors {
		sensor, err := model.NewSensor(ctx, client, &sensorOptions)
		if err != nil {
			panic(err)
		}
		sensors[sensor.Id] = sensor
	}

	// Boiler
	boiler, err := model.NewBoiler(ctx, client, c.Boiler)
	if err != nil {
		panic(err)
	}
	return client, sensors, boiler
}