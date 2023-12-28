package main

import (
	"context"
	"fmt"
	"stupid-caldaia/controller/graph/model"
	"stupid-caldaia/controller/store"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {

	ctx := context.Background()

	config, err := store.LoadConfig()
	if err != nil {
		panic(err)
	}

	client := redis.NewClient(&config.Redis)

	sensors := make([]*model.Sensor, len(config.Sensors))
	for index, sensorOptions := range config.Sensors {
		sensor, err := model.NewSensor(ctx, client, &sensorOptions)
		if err != nil {
			panic(err)
		}
		sensors[index] = sensor
	}

	for {
		for _, sensor := range sensors {
			measure, err := sensor.Sample(ctx)
			if err != nil {
				panic(err)
			}
			fmt.Printf("%s %s %s %.2f\n", sensor.Name, sensor.Position, measure.Timestamp.Local(), measure.Value)
			err = sensor.Add(ctx, measure)
		}
		time.Sleep(1000 * time.Millisecond)
	}
}
