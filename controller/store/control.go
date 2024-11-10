package store

import (
	"context"
	"fmt"
	"stupid-caldaia/controller/graph/model"
	"time"
)

// Given a set of rules and a reference temperature, the function tells us if
// the boiler should be heating
func ShouldHeat(rules []*model.Rule, referenceTemperature float64) bool {
	for _, rule := range rules {
		// Check if the programmed interval is active
		temperatureNotOk := referenceTemperature < rule.TargetTemp
		shouldHeat := rule.ShouldBeActive() && temperatureNotOk
		if shouldHeat {
			return true
		}
	}
	return false
}

func BoilerSwitchControl(ctx context.Context, boiler *model.Boiler, temperatureSensor *model.Sensor) error {
	temperatureListener, err := temperatureSensor.Listen(ctx)
	if err != nil {
		return err
	}
	ruleListener, err := boiler.ListenRules(ctx)
	if err != nil {
		return err
	}
	for {
		// Wait for updates to can affect control...
		var currentTemperature *float64 = nil
		select {
		case <-ruleListener:
		case measure := <-temperatureListener:
			currentTemperature = &measure.Value
		}
		// Actuate control strategy in case of new rules or a new temperature sample
		// First get average temperature of the last 10 minutes
		sensorAverageStart := time.Now().Add(-10 * time.Minute)
		sensorAverageEnd := time.Now()
		averageTemperature, err := temperatureSensor.GetAverage(ctx, sensorAverageStart, sensorAverageEnd)
		if err != nil {
			return fmt.Errorf("could not get average temperature for sensor '%s': %w", temperatureSensor.Name, err)
		}

		// Get latest boiler state
		boilerInfo, err := boiler.GetInfo(ctx)
		if err != nil {
			return fmt.Errorf("could not get Boiler info to set default reference temperature: %w", err)
		}

		var referenceTemperature *float64
		if averageTemperature != nil {
			referenceTemperature = averageTemperature
		} else if currentTemperature != nil {
			referenceTemperature = nil
		} else {
			// Default case will be to assume current temperature is boiler maximum
			// This way we can be safe that with no temperature the boiler is OFF
			referenceTemperature = &boilerInfo.MaxTemp
		}

		// And now, actually asses if we should do it or not
		if ShouldHeat(boilerInfo.Rules, *referenceTemperature) {
			_, err = boiler.Switch(ctx, model.StateOn)
		} else {
			_, err = boiler.Switch(ctx, model.StateOff)
		}
		if err != nil {
			return fmt.Errorf("failed to set boiler state: %w", err)
		}
	}
}

// Long running function to control start and finish of programmed intervals
func RuleTimingControl(ctx context.Context, boiler *model.Boiler) {
	alertTimeoutStop := make(chan *model.Rule)
	alertTimeoutStart := make(chan *model.Rule)
	manageChangeOfRules := func(rules []*model.Rule) context.CancelFunc {
		cancelContext, cancelTimeouts := context.WithCancel(ctx)
		now := time.Now()
		for _, rule := range rules {
			shouldBeActive := rule.ShouldBeActive()
			isActive := rule.IsActive
			willStartInFuture := rule.WindowStartTime(now).After(now)
			switch {
			case shouldBeActive && !isActive:
				// This has just been created and not active, should start immediately
				boiler.StartRule(ctx, rule.ID)
			case shouldBeActive && isActive:
				// Currently active, we need to set a stop timeout
				go rule.WindowStopTimeout(cancelContext, alertTimeoutStop)
			case !shouldBeActive && willStartInFuture:
				// Similar to the first case but the start is forecaste in the future
				go rule.WindowStartTimeout(cancelContext, alertTimeoutStart)
			}
		}
		return cancelTimeouts
	}

	info, err := boiler.GetInfo(ctx)
	if err != nil {
		panic(err)
	}
	cancelTimeouts := manageChangeOfRules(info.Rules)
	defer cancelTimeouts()

	ruleListener, err := boiler.ListenRules(ctx)
	if err != nil {
		panic(err)
	}
	go func() { // Rules orchestration
		for {
			select {
			case rule := <-alertTimeoutStart:
				fmt.Printf("🟢 Received alert. Starting programmed interval... %s\n", rule)
				// Start requested programmed interval and trigger state update
				rule, err := boiler.StartRule(ctx, rule.ID)
				if err != nil {
					fmt.Println(fmt.Errorf("could not start rule after timeout: %w %s", err, rule))
				}
			case rule := <-alertTimeoutStop:
				fmt.Printf("🛑 Received alert. Stopping programmed interval... %s\n", rule)
				rule, err := boiler.StopRule(ctx, rule.ID)
				if err != nil {
					fmt.Println(fmt.Errorf("could not stop rule after timeout: %w %s", err, rule))
				}
			case rule := <-ruleListener:
				// Listen to state updates (change of rules, stops and starts)
				fmt.Printf("🗽 Programmed intervals (count: %d) have changed, updating timeouts...\n", len(rule))
				cancelTimeouts()
				cancelTimeouts = manageChangeOfRules(rule)
			}
		}
	}()
}
