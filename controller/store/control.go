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
	alertTimeoutStop := make(chan *model.ProgrammedInterval, 1)
	alertTimeoutStart := make(chan *model.ProgrammedInterval, 1)
	manageChangeOfRules := func(programmedIntervals []*model.ProgrammedInterval) context.CancelFunc {
		cancelContext, cancelTimeouts := context.WithCancel(ctx)
		now := time.Now()
		for _, programmedInterval := range programmedIntervals {
			shouldBeActive := programmedInterval.ShouldBeActive()
			isActive := programmedInterval.IsActive
			willStartInFuture := programmedInterval.WindowStartTime(now).After(now)
			switch {
			case shouldBeActive && !isActive:
				// This has just been created and not active, should start immediately
				boiler.StartProgrammedInterval(ctx, programmedInterval.ID)
			case shouldBeActive && isActive:
				// Currently active, we need to set a stop timeout
				go programmedInterval.WindowStopTimeout(cancelContext, alertTimeoutStop)
			case !shouldBeActive && willStartInFuture:
				// Similar to the first case but the start is forecaste in the future
				go programmedInterval.WindowStartTimeout(cancelContext, alertTimeoutStart)
			}
		}
		return cancelTimeouts
	}

	info, err := boiler.GetInfo(ctx)
	if err != nil {
		panic(err)
	}
	cancelTimeouts := manageChangeOfRules(info.ProgrammedIntervals)
	defer cancelTimeouts()

	programmedIntervalsListener, err := boiler.ListenProgrammedIntervals(ctx)
	if err != nil {
		panic(err)
	}
	go func() { // Rules orchestration
		for {
			select {
			case programmedInterval := <-alertTimeoutStart:
				fmt.Printf("🟢 Received alert. Starting programmed interval... %s\n", programmedInterval)
				// Start requested programmed interval and trigger state update
				programmedInterval, err := boiler.StartProgrammedInterval(ctx, programmedInterval.ID)
				if err != nil {
					fmt.Println(fmt.Errorf("Could not start rule after timeout: %w %s\n", err, programmedInterval))
				}
			case programmedInterval := <-alertTimeoutStop:
				fmt.Printf("🛑 Received alert. Stopping programmed interval... %s\n", programmedInterval)
				programmedInterval, err := boiler.StopProgrammedInterval(ctx, programmedInterval.ID)
				if err != nil {
					fmt.Println(fmt.Errorf("Could not stop rule after timeout: %w %s\n", err, programmedInterval))
				}
			case programmedIntervals := <-programmedIntervalsListener:
				// Listen to state updates (change of rules, stops and starts)
				fmt.Printf("🗽 Programmed intervals (count: %d) have changed, updating timeouts...\n", len(programmedIntervals))
				cancelTimeouts()
				cancelTimeouts = manageChangeOfRules(programmedIntervals)
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
