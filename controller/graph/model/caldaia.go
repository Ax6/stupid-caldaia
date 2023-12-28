package model

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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

func (c *Boiler) GetProgrammedIntervals(ctx context.Context) (map[string]*ProgrammedInterval, error) {
	info, err := c.GetInfo(ctx)
	if err != nil {
		return nil, err
	}
	programmedIntervals := make(map[string]*ProgrammedInterval)
	for _, interval := range info.ProgrammedIntervals {
		programmedIntervals[interval.ID] = interval
	}
	return programmedIntervals, nil
}

func (c *Boiler) SetProgrammedInterval(ctx context.Context, opt *ProgrammedInterval) (*ProgrammedInterval, error) {
	info, err := c.GetInfo(ctx)
	if err != nil {
		return nil, err
	}
	if opt.TargetTemp < info.MinTemp || opt.TargetTemp > info.MaxTemp {
		return nil, fmt.Errorf("Target temperature out of bounds")
	}

	// Map programmed intervals to a map for easier lookup
	lookupProgrammedIntervals, err := c.GetProgrammedIntervals(ctx)
	if err != nil {
		return nil, err
	}

	// If ID is present in the opt, use that, otherwise generate a new one
	if opt.ID == "" {
		// Create ours if not present
		opt.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	lookupProgrammedIntervals[opt.ID] = opt

	// Convert back to a slice
	programmedIntervals := make([]*ProgrammedInterval, 0, len(lookupProgrammedIntervals))
	for _, interval := range lookupProgrammedIntervals {
		programmedIntervals = append(programmedIntervals, interval)
	}
	info.ProgrammedIntervals = programmedIntervals

	err = c.save(ctx, info)
	return opt, err
}

func (c *Boiler) DeleteProgrammedInterval(ctx context.Context, id string) (bool, error) {
	info, err := c.GetInfo(ctx)
	if err != nil {
		return false, err
	}

	// Map programmed intervals to a map for easier lookup
	lookupProgrammedIntervals := make(map[string]*ProgrammedInterval)
	for _, interval := range info.ProgrammedIntervals {
		lookupProgrammedIntervals[interval.ID] = interval
	}

	if lookupProgrammedIntervals[id] == nil {
		return false, fmt.Errorf("Specified ID not present in programmed intervals")
	}

	delete(lookupProgrammedIntervals, id)

	// Convert back to a slice
	programmedIntervals := make([]*ProgrammedInterval, 0, len(lookupProgrammedIntervals))
	for _, interval := range lookupProgrammedIntervals {
		programmedIntervals = append(programmedIntervals, interval)
	}
	info.ProgrammedIntervals = programmedIntervals

	err = c.save(ctx, info)
	return true, err
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
