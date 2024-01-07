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
		temperatureNotOk := referenceTemperature < programmedInterval.TargetTemp
		shouldHeat := programmedInterval.ShouldBeActive() && temperatureNotOk
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

func RuleTimingController(ctx context.Context, boiler *model.Boiler) {
	info, err := boiler.GetInfo(ctx)
	if err != nil {
		panic(err)
	}

	alertTimeout := make(chan *model.ProgrammedInterval)
	cancelContext, cancelTimeouts := context.WithCancel(ctx)
	defer cancelTimeouts()
	for _, programmedInterval := range info.ProgrammedIntervals {
		if programmedInterval.ShouldBeActive() {
			go programmedInterval.WindowTimeout(cancelContext, alertTimeout)
		}
	}
	programmedIntervalsListener, err := boiler.ListenProgrammedIntervals(ctx)
	if err != nil {
		panic(err)
	}
	go func() {
		for programmedInterval := range alertTimeout {
			err := boiler.StopProgrammedInterval(ctx, programmedInterval.ID)
			if err != nil {
				fmt.Println(fmt.Errorf("Could not stop rule %s after timeout: %w", programmedInterval.ID, err))
			} else {
				fmt.Printf("🛑 Timeout for programmed interval %s\n", programmedInterval.ID)
			}
		}
	}()
	go func() {
		for programmedIntervals := range programmedIntervalsListener {
			cancelTimeouts()
			cancelContext, cancelTimeouts = context.WithCancel(ctx)
			defer cancelTimeouts()
			for _, programmedInterval := range programmedIntervals {
				if programmedInterval.ShouldBeActive() {
					boiler.StartProgrammedInterval(ctx, programmedInterval.ID)
					go programmedInterval.WindowTimeout(cancelContext, alertTimeout)
				}
			}
		}
	}()
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
