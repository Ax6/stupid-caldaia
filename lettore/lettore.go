package main

import (
	"context"
	"fmt"
	"stupid-caldaia/controller/store"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	CONFIG_PATH = "../config.json"
)

func main() {

	ctx := context.Background()

	config, err := store.LoadConfig()
	if err != nil {
		panic(err)
	}

	client := redis.NewClient(&config.Redis)

	sensors := make([]*store.Sensor, len(config.Sensors))
	for index, sensorOptions := range config.Sensors {
		sensor, err := store.NewSensor(ctx, client, &sensorOptions)
		if err != nil {
			panic(err)
		}
		sensors[index] = sensor
	}

	for {
		for _, sensor := range sensors {
			measure, err := sensor.Sample()
			if err != nil {
				panic(err)
			}
			fmt.Printf("%s %s %f %f\n", sensor.Name, sensor.Position, measure.Timestamp, measure.Value)
			err = sensor.Add(ctx, measure)
		}
		time.Sleep(1000 * time.Millisecond)
	}
}
