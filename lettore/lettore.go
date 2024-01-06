package main

import (
	"context"
	"fmt"
	"log"
	"stupid-caldaia/controller/graph/model"
	"stupid-caldaia/controller/store"
	"sync"
	"time"

	"github.com/parMaster/htu21"
	"github.com/stianeikeland/go-rpio/v4"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/host/v3"
)

const (
	ControllerHost = "localhost"
)

var (
	htu21Data   physic.Env
	htu21Device *htu21.Dev
	wg          sync.WaitGroup
)

func ObserveState(ctx context.Context, boiler *model.Boiler) {
	err := rpio.Open()
	if err != nil {
		log.Panic("Could not open gpio... 😱")
	}

	updateIO := func(info *model.BoilerInfo) {
		pin := rpio.Pin(boiler.Config.SwitchPin)
		pin.Output()
		fmt.Printf("State has changed 😮, updating gpio 👉 %d to %s\n", boiler.Config.SwitchPin, info.State)
		switch info.State {
		case model.StateOn:
			pin.High()
		case model.StateOff:
			pin.Low()
		default:
			break
		}
	}

	// Make sure we set I/O right from the start according to our state.
	onStartInfo, err := boiler.GetInfo(ctx)
	if err != nil {
		log.Panic("Could not get intial boiler state 😱")
	}
	updateIO(onStartInfo)

	listener, err := boiler.Listen(ctx)
	if err != nil {
		log.Panic("Could not listen to boiler... 🙉")
	}
	for info := range listener {
		select {
		case <-ctx.Done():
			return
		default:
			updateIO(info)
		}
	}
}

func ObserveSensor(ctx context.Context, sensors map[string]*model.Sensor) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// zzz
			time.Sleep(1 * time.Second)
			// Read sensors...
			if _, err := host.Init(); err != nil {
				log.Panic(err)
			}

			// Use i2creg I²C bus registry to find the first available I²C bus.
			b, err := i2creg.Open("1")
			if err != nil {
				log.Panicf("failed to open I²C: %v", err)
			}

			htu21Device, err = htu21.NewI2C(b, 0x40)
			if err != nil {
				log.Panicf("failed to initialize htu21: %v", err)
			}

			// Measure
			if err := htu21Device.Sense(&htu21Data); err != nil {
				log.Panic(err)
			}

			// Add to database
			sensors["temperatura:centrale"].AddSample(ctx, &model.Measure{
				Timestamp: time.Now(),
				Value:     htu21Data.Temperature.Celsius(),
			})

			sensors["umidita:centrale"].AddSample(ctx, &model.Measure{
				Timestamp: time.Now(),
				Value:     float64(htu21Data.Humidity / physic.MilliRH * physic.PercentRH),
			})
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	config, err := store.LoadConfig()
	if err != nil {
		log.Panic(err)
	}

	wg.Add(2)
	_, sensors, boiler := config.CreateObjects(ctx)

	// Start go routines
	go func() {
		defer wg.Done()
		ObserveSensor(ctx, sensors)
	}()

	go func() {
		defer wg.Done()
		ObserveState(ctx, boiler)
	}()

	defer func() {
		pin := rpio.Pin(boiler.Config.SwitchPin)
		pin.Output()
		pin.Low()
		rpio.Close()
	}()

	wg.Wait()
}
