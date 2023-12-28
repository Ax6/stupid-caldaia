package main

import (
	"context"
	"fmt"
	"log"
	"stupid-caldaia/controller/graph/model"
	"stupid-caldaia/controller/store"
	"time"

	"github.com/parMaster/htu21"
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
)

func main() {
	ctx := context.Background()

	config, err := store.LoadConfig()
	if err != nil {
		panic(err)
	}
	close, err := config.OpenPin()
	if err != nil {
		panic(err)
	}
	defer close()

	_, sensors, boiler := config.CreateObjects(ctx)

	for {
		// zzz
		time.Sleep(1 * time.Second)

		// Read sensors...
		if _, err := host.Init(); err != nil {
			log.Fatal(err)
		}

		// Use i2creg I²C bus registry to find the first available I²C bus.
		b, err := i2creg.Open("1")
		if err != nil {
			log.Fatalf("failed to open I²C: %v", err)
		}

		htu21Device, err = htu21.NewI2C(b, 0x40)
		if err != nil {
			log.Fatalf("failed to initialize htu21: %v", err)
		}

		// Measure
		if err := htu21Device.Sense(&htu21Data); err != nil {
			log.Fatal(err)
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

		// And now use some logic

		// Get all the programmed intervals
		programmedIntervals, err := boiler.GetProgrammedIntervals(ctx)
		if err != nil {
			panic(err)
		}

		// Get the temperature of the last 20 minutes
		centrale := sensors["temperatura:centrale"]
		averageTemperature, err := centrale.GetAverage(ctx, time.Now().Add(-20*time.Minute), time.Now())
		if err != nil {
			panic(err)
		}
		currentTemperature := htu21Data.Temperature.Celsius()
		if averageTemperature != nil {
			currentTemperature = *averageTemperature
		}

		shouldHeat := false
		for _, programmedInterval := range programmedIntervals {
			// Check if the programmed interval is active
			projectedStartTime := time.Now().Add(-programmedInterval.Duration)
			ruleIsActive := projectedStartTime.Before(programmedInterval.Start)
			// print for debug
			fmt.Printf("Rule %s is active: %t\n", programmedInterval.ID, ruleIsActive)
			temperatureNotOk := currentTemperature < programmedInterval.TargetTemp
			// print for debug
			fmt.Printf("%f < %f: %t\n", currentTemperature, programmedInterval.TargetTemp, temperatureNotOk)
			shouldHeat = ruleIsActive && temperatureNotOk
			if shouldHeat {
				break
			}
		}
		if shouldHeat {
			boiler.Switch(ctx, model.StateOn)
		} else {
			boiler.Switch(ctx, model.StateOff)
		}
	}
}
