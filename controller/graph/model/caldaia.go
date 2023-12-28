package model

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/stianeikeland/go-rpio/v4"
)

type BoilerConfig struct {
	Name                  string
	DefaultMinTemperature float64
	DefaultMaxTemperature float64
	SwitchPin             int
}

type Boiler struct {
	config BoilerConfig
	client *redis.Client
}

func NewBoiler(ctx context.Context, client *redis.Client, config BoilerConfig) (*Boiler, error) {
	boiler := Boiler{config, client}
	_, err := boiler.GetInfo(ctx)
	return &boiler, err
}

// Function to switch the relay on or off
// Accepts only two values: "on" or "off"
func (c *Boiler) Switch(ctx context.Context, targetState State) (*State, error) {
	err := rpio.Open()
	if err != nil {
		return nil, err
	}
	defer rpio.Close()
	pin := rpio.Pin(c.config.SwitchPin)
	pin.Output()
	info, err := c.GetInfo(ctx)
	if err != nil {
		return nil, err
	}
	switch targetState {
	case StateOn:
		fmt.Println("Switching relay on")
		pin.High()
	case StateOff:
		fmt.Println("Switching relay off")
		pin.Low()
	default:
		return &targetState, fmt.Errorf("Invalid state to set")
	}
	info.State = targetState
	err = c.save(ctx, info)
	return &info.State, err
}

func (c *Boiler) SetMinTemp(ctx context.Context, temp float64) (*float64, error) {
	info, err := c.GetInfo(ctx)
	if err != nil {
		return nil, err
	}
	if temp < info.MinTemp || temp > info.MaxTemp {
		return &info.MinTemp, fmt.Errorf("Invalid min temperature")
	}
	info.MinTemp = temp
	err = c.save(ctx, info)
	return &info.MinTemp, err
}

func (c *Boiler) SetMaxTemp(ctx context.Context, temp float64) (*float64, error) {
	info, err := c.GetInfo(ctx)
	if err != nil {
		return nil, err
	}
	if temp > info.MaxTemp || temp < info.MinTemp {
		return &info.MaxTemp, fmt.Errorf("Invalid max temperature")
	}
	info.MaxTemp = temp
	err = c.save(ctx, info)
	return &info.MaxTemp, err
}

func (c *Boiler) SetTargetTemp(ctx context.Context, temp float64) (*float64, error) {
	info, err := c.GetInfo(ctx)
	if err != nil {
		return nil, err
	}
	if temp < info.MinTemp || temp > info.MaxTemp {
		return &info.MaxTemp, fmt.Errorf("Invalid target temperature")
	}
	info.TargetTemp = &temp
	err = c.save(ctx, info)
	return info.TargetTemp, err
}

func (c *Boiler) Listen(ctx context.Context) (<-chan *BoilerInfo, error) {
	boilerUpdates := make(chan *BoilerInfo)
	go func() {
		sub := c.client.Subscribe(ctx, c.config.Name)
		for msg := range sub.Channel() {
			boiler := BoilerInfo{}
			err := json.Unmarshal([]byte(msg.Payload), &boiler)
			if err != nil {
				fmt.Println(err)
				break
			} else {
				boilerUpdates <- &boiler
			}
		}
		close(boilerUpdates)
	}()
	return boilerUpdates, nil
}

func (c *Boiler) GetInfo(ctx context.Context) (*BoilerInfo, error) {
	data := c.client.Get(ctx, c.config.Name).Val()
	if data == "" {
		defaultInfo := &BoilerInfo{
			State:               StateUnknown,
			MinTemp:             c.config.DefaultMinTemperature,
			MaxTemp:             c.config.DefaultMaxTemperature,
			TargetTemp:          nil,
			ProgrammedIntervals: nil,
		}
		err := c.save(ctx, defaultInfo) // Save default values
		return defaultInfo, err
	} else {
		var info BoilerInfo
		err := json.Unmarshal([]byte(data), &info)
		return &info, err
	}
}

func (c *Boiler) save(ctx context.Context, info *BoilerInfo) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	err = c.client.Set(ctx, c.config.Name, data, 0).Err()
	if err != nil {
		return err
	}
	err = c.client.Publish(ctx, c.config.Name, data).Err()
	return err
}
