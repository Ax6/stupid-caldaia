package store

import (
	"context"
	"encoding/json"
	"fmt"
	"stupid-caldaia/controller/graph/model"
	"time"

	"github.com/google/go-cmp/cmp"
)

func ShouldHeat(programmedIntervals map[string]*model.ProgrammedInterval, referenceTemperature float64) bool {
	for _, programmedInterval := range programmedIntervals {
		// Check if the programmed interval is active
		projectedStartTime := time.Now().Add(-programmedInterval.Duration)
		ruleIsActive := projectedStartTime.Before(programmedInterval.Start)
		// print for debug
		temperatureNotOk := referenceTemperature < programmedInterval.TargetTemp
		// print for debug
		fmt.Printf("%f < %f: %t\n", referenceTemperature, programmedInterval.TargetTemp, temperatureNotOk)
		shouldHeat := ruleIsActive && temperatureNotOk
		if shouldHeat {
			return true
		}
	}
	return false
}

func TemperatureChangeController(ctx context.Context, boiler *model.Boiler, temperatureSensor *model.Sensor) {
	temperatureListener, err := temperatureSensor.Listen(ctx)
	if err != nil {
		panic(err)
	}
	for measure := range temperatureListener {
		averageTemperature, err := temperatureSensor.GetAverage(ctx, time.Now().Add(-20*time.Minute), time.Now())
		if err != nil {
			panic(err)
		}
		currentTemperature := measure.Value
		if averageTemperature != nil {
			currentTemperature = *averageTemperature
		}
		programmedIntervals, err := boiler.GetProgrammedIntervals(ctx)
		if err != nil {
			panic(err)
		}
		if ShouldHeat(programmedIntervals, currentTemperature) {
			boiler.Switch(ctx, model.StateOn)
		} else {
			boiler.Switch(ctx, model.StateOff)
		}
	}
}

func RuleChangeController(ctx context.Context, boiler *model.Boiler, temperatureSensor *model.Sensor) {
	boilerListener, err := boiler.Listen(ctx)
	if err != nil {
		panic(err)
	}
	boilerInfo, err := boiler.GetInfo(ctx)
	if err != nil {
		panic(err)
	}

	currentRules, err := json.Marshal(boilerInfo.ProgrammedIntervals)
	if err != nil {
		panic(err)
	}

	for boilerInfo := range boilerListener {
		newRules, err := json.Marshal(boilerInfo.ProgrammedIntervals)
		if err != nil {
			panic(err)
		}
		if !cmp.Equal(currentRules, newRules) {
			averageTemperature, err := temperatureSensor.GetAverage(ctx, time.Now().Add(-20*time.Minute), time.Now())
			if err != nil {
				panic(err)
			}
			programmedIntervals, err := boiler.GetProgrammedIntervals(ctx)
			if err != nil {
				panic(err)
			}
			if ShouldHeat(programmedIntervals, *averageTemperature) {
				boiler.Switch(ctx, model.StateOn)
			} else {
				boiler.Switch(ctx, model.StateOff)
			}
		}
		currentRules = newRules
	}
}
