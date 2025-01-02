package store

import (
	"context"
	"fmt"
	"log"
	"stupid-caldaia/controller/graph/model"
	"time"
)

const (
	OVERHEATING_CHECK_PERIOD  = 15 * time.Second
	OVERHEATING_ON_THRESHOLD  = 0.9
	OVERHEATING_OFF_THRESHOLD = 0.2
)

// Long running function to enable/disable boiler based on overheating
func BoilerOverheatingControl(ctx context.Context, boiler *model.Boiler, checkInterval time.Duration) error {
	ticker := time.Tick(checkInterval)
	for {
		select {
		case <-ticker:
			currentIndex, err := model.GetCurrentOverheatingIndex(ctx, boiler)
			if err != nil {
				return err
			}
			info, err := boiler.GetInfo(ctx)
			if err != nil {
				return err
			}
			isProtected := info.IsOverheatingProtectionActive

			if currentIndex > OVERHEATING_ON_THRESHOLD && !isProtected {
				log.Printf("Enabling overheating protection. Recorded index above threshold.")
				boiler.SetOverheating(ctx, true)
			}

			if currentIndex < OVERHEATING_OFF_THRESHOLD && isProtected {
				log.Printf("Disabling overheating protection. Cooldown reached.")
				boiler.SetOverheating(ctx, false)
			}
		case <-ctx.Done():
			return nil
		}
	}
}

// Long running function to control the On/Off state
func BoilerSwitchControl(ctx context.Context, boiler *model.Boiler, temperatureSensor *model.Sensor) error {
	temperatureListener, err := temperatureSensor.Listen(ctx)
	if err != nil {
		return err
	}
	ruleListener, err := boiler.ListenRules(ctx)
	if err != nil {
		return err
	}
	overheatingListener, err := boiler.ListenOverheating(ctx)
	if err != nil {
		return err
	}
	for {
		// Wait for updates to can affect control...
		var currentTemperature *float64 = nil
		select {
		case <-ruleListener:
		case <-overheatingListener:
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

		// Get reference temperature
		var referenceTemperature *float64
		if averageTemperature != nil {
			// Good we have an average temperature, we'll use it as reference
			referenceTemperature = averageTemperature
		} else if currentTemperature != nil {
			// Alright, we'll fallback to the current temperature
			referenceTemperature = currentTemperature
		} else {
			// Default case will be to assume current temperature is boiler maximum
			// This way we can be safe that with no temperature the boiler is OFF
			referenceTemperature = &boilerInfo.MaxTemp
		}

		// Can heat if not protected from overheating
		canHeat := !boilerInfo.IsOverheatingProtectionActive

		// And now, actually asses if we should do it or not
		if shouldHeat(boilerInfo.Rules, *referenceTemperature) && canHeat {
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
func RuleTimingControl(ctx context.Context, boiler *model.Boiler) error {
	ruleListener, err := boiler.ListenRules(ctx)
	if err != nil {
		return err
	}
	for {
		info, err := boiler.GetInfo(ctx)
		if err != nil {
			return err
		}
		cancellableContext, cancelTimeouts := context.WithCancel(ctx)
		now := time.Now()
		for _, rule := range info.Rules {
			shouldBeStopped := rule.ShouldBeStopped()
			shouldBeActive := rule.ShouldBeActive()
			isActive := rule.IsActive
			willStartInFuture := rule.WindowStartTime(now).After(now)
			switch {
			case shouldBeActive && !isActive:
				// This has just been created and not active, should start immediately
				// May occour on cold starts
				fmt.Printf("🧊 🟢 Rule state different from expected state - enforcing 'StartRule': %s\n", rule)
				boiler.StartRule(ctx, rule.ID)
			case shouldBeStopped && isActive:
				// This may have been started in some previous lifetime and it wasn't stopped
				// May occour on cold starts
				fmt.Printf("🧊 🔴 Rule state different from expected state - enforcing 'StopRule': %s\n", rule)
				boiler.StopRule(ctx, rule.ID)
			case shouldBeActive && isActive:
				// Currently active, we need to set a stop timeout
				go waitAndStopRule(cancellableContext, boiler, rule)
			case !shouldBeActive && willStartInFuture:
				// Similar to the first case but the start is forecaste in the future
				go waitAndStartRule(cancellableContext, boiler, rule)
			}
		}

		// Listen to state updates (change of rules, stops and starts)
		newRules := <-ruleListener
		fmt.Printf("🗽 Programmed intervals (count: %d) have changed, updating timeouts...\n", len(newRules))
		cancelTimeouts()
	}
}

func waitAndStartRule(cancellableContext context.Context, boiler *model.Boiler, rule *model.Rule) {
	if rule.WindowStartTimeout(cancellableContext) {
		// When and if timeout occurred
		fmt.Printf("🟢 Received alert. Starting rule... %s\n", rule)
		rule, err := boiler.StartRule(cancellableContext, rule.ID)
		if err != nil {
			fmt.Println(fmt.Errorf("could not start rule after timeout: %w %s", err, rule))
		}
	}
}

func waitAndStopRule(cancellableContext context.Context, boiler *model.Boiler, rule *model.Rule) {
	if rule.WindowStopTimeout(cancellableContext) {
		// When and if timeout occurred
		fmt.Printf("🛑 Received alert. Stopping rule... %s\n", rule)
		rule, err := boiler.StopRule(cancellableContext, rule.ID)
		if err != nil {
			fmt.Println(fmt.Errorf("could not stop rule after timeout: %w %s", err, rule))
		}
	}
}

// Given a set of rules and a reference temperature, the function tells us if
// the boiler should be heating
func shouldHeat(rules []*model.Rule, referenceTemperature float64) bool {
	for _, rule := range rules {
		// Check if the programmed interval is active
		temperatureNotOk := referenceTemperature < rule.TargetTemp
		shouldHeat := rule.ShouldBeActive() && temperatureNotOk && !rule.IsBeingDelayed()
		if shouldHeat {
			return true
		}
	}
	return false
}
