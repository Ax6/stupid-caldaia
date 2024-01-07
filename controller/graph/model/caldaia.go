package model

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/redis/go-redis/v9"
)

type BoilerConfig struct {
	Name                  string
	DefaultMinTemperature float64
	DefaultMaxTemperature float64
	SwitchPin             int
}

type Boiler struct {
	Config BoilerConfig
	client *redis.Client
}

func NewBoiler(ctx context.Context, client *redis.Client, config BoilerConfig) (*Boiler, error) {
	boiler := Boiler{Config: config, client: client}
	_, err := boiler.GetInfo(ctx)
	return &boiler, err
}

// Function to switch the relay on or off
// Accepts only two values: "on" or "off"
func (c *Boiler) Switch(ctx context.Context, targetState State) (*State, error) {
	info, err := c.GetInfo(ctx)
	if err != nil {
		return nil, err
	}
	if targetState != StateOn && targetState != StateOff {
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

func (c *Boiler) SetProgrammedInterval(ctx context.Context, opt *ProgrammedInterval) (*ProgrammedInterval, error) {
	info, err := c.GetInfo(ctx)
	if err != nil {
		return nil, err
	}
	if opt.TargetTemp < info.MinTemp || opt.TargetTemp > info.MaxTemp {
		return nil, fmt.Errorf("Target temperature out of bounds")
	}

	// Map programmed intervals to a map for easier lookup
	lookupProgrammedIntervals := make(map[string]*ProgrammedInterval)
	for _, interval := range info.ProgrammedIntervals {
		lookupProgrammedIntervals[interval.ID] = interval
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

func (c *Boiler) StartProgrammedInterval(ctx context.Context, id string) error {
	info, err := c.GetInfo(ctx)
	if err != nil {
		return err
	}

	for _, programmedInterval := range info.ProgrammedIntervals {
		err = fmt.Errorf("Could not find programmedInterval with id: %s", id)
		if programmedInterval.ID == id {
			programmedInterval.StoppedTime = time.Time{}
			programmedInterval.IsActive = true
			err = nil
			break
		}
	}
	if err != nil {
		return err
	}

	c.save(ctx, info)
	return nil
}

func (c *Boiler) StopProgrammedInterval(ctx context.Context, id string) error {
	info, err := c.GetInfo(ctx)
	if err != nil {
		return err
	}

	for _, programmedInterval := range info.ProgrammedIntervals {
		err = fmt.Errorf("Could not find programmedInterval with id: %s", id)
		if programmedInterval.ID == id {
			programmedInterval.StoppedTime = time.Now()
			programmedInterval.IsActive = false
			err = nil
			break
		}
	}
	if err != nil {
		return err
	}

	c.save(ctx, info)
	return nil
}

func (c *Boiler) DeleteProgrammedInterval(ctx context.Context, id string) error {
	info, err := c.GetInfo(ctx)
	if err != nil {
		return err
	}

	for index, programmedInterval := range info.ProgrammedIntervals {
		err = fmt.Errorf("Could not find programmed interval with id: %s", id)
		if programmedInterval.ID == id {
			info.ProgrammedIntervals = append(info.ProgrammedIntervals[:index], info.ProgrammedIntervals[index+1:]...)
			err = nil
			break
		}
	}
	if err != nil {
		return err
	}

	err = c.save(ctx, info)
	return err
}

func (c *Boiler) ListenProgrammedIntervals(ctx context.Context) (<-chan []*ProgrammedInterval, error) {
	programmedIntervalUpdates := make(chan []*ProgrammedInterval)
	boilerInfo, err := c.GetInfo(ctx)
	if err != nil {
		return nil, err
	}
	currentRules, err := json.Marshal(boilerInfo.ProgrammedIntervals)
	if err != nil {
		return nil, err
	}
	boilerListener, err := c.Listen(ctx)
	if err != nil {
		return nil, err
	}
	go func() {
		for boilerInfo := range boilerListener {
			newRules, err := json.Marshal(boilerInfo.ProgrammedIntervals)
			if err != nil {
				fmt.Println(fmt.Errorf("Error marshalling programmed intervals: %w", err))
				continue
			}
			if !cmp.Equal(currentRules, newRules) {
				programmedIntervalUpdates <- boilerInfo.ProgrammedIntervals
			}
			currentRules = newRules
		}
	}()
	return programmedIntervalUpdates, nil
}

func (c *Boiler) Listen(ctx context.Context) (<-chan *BoilerInfo, error) {
	boilerUpdates := make(chan *BoilerInfo)
	go func() {
		sub := c.client.Subscribe(ctx, c.Config.Name)
		defer sub.Close()
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
	data, err := c.client.Get(ctx, c.Config.Name).Result()
	switch err {
	case redis.Nil: // Data doesn't exist yet
		defaultInfo := &BoilerInfo{
			State:               StateUnknown,
			MinTemp:             c.Config.DefaultMinTemperature,
			MaxTemp:             c.Config.DefaultMaxTemperature,
			ProgrammedIntervals: nil,
		}
		err := c.save(ctx, defaultInfo) // Save default values
		return defaultInfo, err
	case nil: // No error
		var info BoilerInfo
		err := json.Unmarshal([]byte(data), &info)
		return &info, err
	default:
		return nil, err
	}
}

func (c *Boiler) save(ctx context.Context, info *BoilerInfo) error {
	// Serialise data
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	storedData, err := c.client.Get(ctx, c.Config.Name).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("Cannot update database. Error when getting current state: %w", err)
	}
	diff := cmp.Diff(data, []byte(storedData))
	if diff != "" {
		err = c.client.Set(ctx, c.Config.Name, data, 0).Err()
		if err != nil {
			return err
		}
		err = c.client.Publish(ctx, c.Config.Name, data).Err()
	}
	return err
}
