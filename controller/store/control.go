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

	alertTimeoutStop := make(chan *model.ProgrammedInterval, 1)
	alertTimeoutStart := make(chan *model.ProgrammedInterval, 1)
	cancelContext, cancelTimeouts := context.WithCancel(ctx)
	defer cancelTimeouts()
	for _, programmedInterval := range info.ProgrammedIntervals {
		if programmedInterval.ShouldBeActive() {
			// Setup requires to fire start timeout too or the state won't be updated
			// Should be active -> Fire start/stop timeout
			go programmedInterval.WindowStartTimeout(cancelContext, alertTimeoutStart)
			go programmedInterval.WindowStopTimeout(cancelContext, alertTimeoutStop)
		} else {
			// Only waiting for the next window to start
			go programmedInterval.WindowStartTimeout(cancelContext, alertTimeoutStart)
		}
	}
	programmedIntervalsListener, err := boiler.ListenProgrammedIntervals(ctx)
	if err != nil {
		panic(err)
	}

	go func() { // Rules orchestration
		for {
			select {
			case programmedInterval := <-alertTimeoutStart:
				// Start requested programmed interval and trigger state update
				err := boiler.StartProgrammedInterval(ctx, programmedInterval.ID)
				if err != nil {
					fmt.Println(fmt.Errorf("Could not start rule %s after timeout: %w", programmedInterval, err))
				} else {
					fmt.Printf("🟢 Started programmed interval %s\n", programmedInterval)
				}
			case programmedInterval := <-alertTimeoutStop:
				// Stop programmed interval and trigger state update
				err := boiler.StopProgrammedInterval(ctx, programmedInterval.ID)
				if err != nil {
					fmt.Println(fmt.Errorf("Could not stop rule %s after timeout: %w", programmedInterval, err))
				} else {
					fmt.Printf("🛑 Timeout for programmed interval %s\n", programmedInterval)
				}
			case programmedIntervals := <-programmedIntervalsListener:
				// Listen to state updates (change of rules, stops and starts)
				cancelTimeouts()
				cancelContext, cancelTimeouts = context.WithCancel(ctx)
				now := time.Now()
				for _, programmedInterval := range programmedIntervals {
					if programmedInterval.ShouldBeActive() {
						fmt.Printf("🟢 Rule %s should be active\n", programmedInterval)
						// Restart programmed timeout to stop the rule
						go programmedInterval.WindowStopTimeout(cancelContext, alertTimeoutStop)
					} else {
						fmt.Printf("🛑 Rule %s should not be active\n", programmedInterval)
						// Restart programmed timeout to start the rule
						if programmedInterval.WindowStartTime(now).After(now) {
							// If we have an upcoming start time fire the timeout
							go programmedInterval.WindowStartTimeout(cancelContext, alertTimeoutStart)
						}
					}
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
