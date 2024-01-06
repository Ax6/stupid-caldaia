package store

import (
	"context"
	"fmt"
	"stupid-caldaia/controller/graph/model"
	"time"
)

func ShouldHeat(programmedIntervals []*model.ProgrammedInterval, referenceTemperature float64) bool {
	for _, programmedInterval := range programmedIntervals {
		// Check if the programmed interval is active
		projectedStartTime := time.Now().Add(-programmedInterval.Duration)
		ruleIsActive := projectedStartTime.Before(programmedInterval.Start)
		temperatureNotOk := referenceTemperature < programmedInterval.TargetTemp
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
		boilerInfo, err := boiler.GetInfo(ctx)
		if err != nil {
			panic(err)
		}
		if ShouldHeat(boilerInfo.ProgrammedIntervals, currentTemperature) {
			boiler.Switch(ctx, model.StateOn)
		} else {
			boiler.Switch(ctx, model.StateOff)
		}
	}
}

func RuleTimeoutController(ctx context.Context, boiler *model.Boiler) {
	return
}

func RuleEnforceController(ctx context.Context, boiler *model.Boiler, temperatureSensor *model.Sensor) {
	programmedIntervalsListener, err := boiler.ListenProgrammedIntervals(ctx)
	if err != nil {
		panic(err)
	}
	for programmedIntervals := range programmedIntervalsListener {
		averageTemperature, err := temperatureSensor.GetAverage(ctx, time.Now().Add(-20*time.Minute), time.Now())
		if err != nil {
			panic(err)
		}
		referenceTemperature := *averageTemperature
		if averageTemperature == nil {
			boilerInfo, err := boiler.GetInfo(ctx)
			if err != nil {
				fmt.Println(fmt.Errorf("Could not get Boiler info to set default reference temperature: %w", err))
			}
			referenceTemperature = boilerInfo.MaxTemp
		}

		if ShouldHeat(programmedIntervals, referenceTemperature) {
			boiler.Switch(ctx, model.StateOn)
		} else {
			boiler.Switch(ctx, model.StateOff)
		}
	}
}
