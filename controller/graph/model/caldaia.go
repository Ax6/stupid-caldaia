package model

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

const (
	MinimumTemperature = 0
	MaximumTemperature = 30
)

type BoilerStore struct {
	Boiler
	client *redis.Client
}

func GetCaldaia(ctx context.Context, client *redis.Client) (*BoilerStore, error) {
	caldaia := BoilerStore{
		Boiler{
			State:               StateUnknown,
			MinTemp:             MinimumTemperature,
			MaxTemp:             MaximumTemperature,
			TargetTemp:          nil,
			ProgrammedIntervals: nil,
		},
		client,
	}
	data := client.Get(ctx, "caldaia").Val()
	if data == "" {
		err := caldaia.save(ctx) // Save default values
		return &caldaia, err
	} else {
		err := json.Unmarshal([]byte(data), &caldaia)
		return &caldaia, err
	}
}

// Function to switch the relay on or off
// Accepts only two values: "on" or "off"
func (c *BoilerStore) Switch(ctx context.Context, targetState State) (State, error) {
	switch targetState {
	case StateOn:
		fmt.Println("Switching relay on")
	case StateOff:
		fmt.Println("Switching relay off")
	default:
		return targetState, fmt.Errorf("Invalid state to set")
	}
	c.State = targetState
	err := c.save(ctx)
	return c.State, err
}

func (c *BoilerStore) SetMinTemp(ctx context.Context, temp float64) (float64, error) {
	if temp < MinimumTemperature || temp > c.MaxTemp {
		return c.MinTemp, fmt.Errorf("Invalid min temperature")
	}
	c.MinTemp = temp
	err := c.save(ctx)
	return c.MinTemp, err
}

func (c *BoilerStore) SetMaxTemp(ctx context.Context, temp float64) (float64, error) {
	if temp > MaximumTemperature || temp < c.MinTemp {
		return c.MaxTemp, fmt.Errorf("Invalid max temperature")
	}
	c.MaxTemp = temp
	err := c.save(ctx)
	return c.MaxTemp, err
}

func (c *BoilerStore) SetTargetTemp(ctx context.Context, temp float64) (float64, error) {
	if temp < c.MinTemp || temp > c.MaxTemp {
		return c.MaxTemp, fmt.Errorf("Invalid target temperature")
	}
	c.TargetTemp = &temp
	err := c.save(ctx)
	return *c.TargetTemp, err
}

func (c *BoilerStore) save(ctx context.Context) error {
	data, err := json.Marshal(c.Boiler)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, "caldaia", data, 0).Err()
}
