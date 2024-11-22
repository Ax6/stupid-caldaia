package model

import (
	"context"
	"fmt"
	"time"
)

type DayOfWeek = time.Weekday

// It should be active if now it's in the current time window and we shouldn't
// stop This includes a delay, hence, the rule "ShouldBeActive" also when 'now'
// is in the delay duration and not yet the actual start.
func (p *Rule) ShouldBeActive() bool {
	now := time.Now()
	wStart := p.WindowStartTime(now)
	wEnd := wStart.Add(p.DurationWithDelay())
	// If it shouldn't be in a stopped state and we're inside the window
	return now.After(wStart) && now.Before(wEnd) && !p.ShouldBeStopped()
}

// If now is in the delay window before the start
func (p *Rule) IsBeingDelayed() bool {
	now := time.Now()
	wStart := p.WindowStartTime(now)
	return now.After(wStart) && now.Before(wStart.Add(p.Delay))
}

// It should stop if the stop command was sent in the current or upcoming window and we are past that time
func (p *Rule) ShouldBeStopped() bool {
	stopTime := p.StoppedTime
	if stopTime == nil || stopTime.IsZero() {
		return false
	}
	now := time.Now()
	wStart := p.WindowStartTime(now)
	wEnd := wStart.Add(p.DurationWithDelay())
	// Basically if
	// 1. Last time we stopped the rule is before now
	// 2. Last time we stopped the rule is after the window start
	// 3. Last time we stopped the rule is before the window end
	// then -> Current rule should be in stopped state
	return stopTime.Before(now) && stopTime.After(wStart) && stopTime.Before(wEnd)
}

// Relative to the referenceTime, if we are in a window returns the start time
// of this window, otherwise returns the start of the upcoming window
func (p *Rule) WindowStartTime(referenceTime time.Time) time.Time {
	if len(p.RepeatDays) == 0 {
		return p.Start
	}
	now := referenceTime
	daysUntilTarget := 7
	todaysStart := time.Date(now.Year(), now.Month(), now.Day(), p.Start.Hour(), p.Start.Minute(), p.Start.Second(), p.Start.Nanosecond(), p.Start.Location())

	for _, programmedWeekDay := range p.RepeatDays {
		// Calculate the difference in days between the current day and the target day
		daysAway := (programmedWeekDay - int(now.Weekday()) + 7) % 7

		if daysAway == 0 {
			// If we are 0 days away we have to check if the repeated rule has finished already, in that case we are a week away
			if todaysStart.Add(p.DurationWithDelay()).Before(now) {
				daysAway = 7
			}
		}
		if daysAway < daysUntilTarget {
			daysUntilTarget = daysAway
		}
	}

	upcomingStart := todaysStart.Add(time.Duration(daysUntilTarget) * time.Hour * 24)
	return upcomingStart
}

// Sums delay and set duration
func (p *Rule) DurationWithDelay() time.Duration {
	return p.Delay + p.Duration
}

func (p *Rule) WindowStopTimeout(ctx context.Context) bool {
	now := time.Now()
	totalDuration := p.WindowStartTime(now).Sub(now) + p.DurationWithDelay()
	fmt.Printf("⏰ Set stop timeout of %s for interval %s\n", totalDuration, p)
	select {
	case <-ctx.Done():
		return false
	case <-time.After(totalDuration):
		fmt.Printf("✋ %s Alerting stop! (After %s)\n", p.ID, time.Since(now))
		return true
	}
}

func (p *Rule) WindowStartTimeout(ctx context.Context) bool {
	now := time.Now()
	duration := p.WindowStartTime(now).Sub(now)
	fmt.Printf("⏰ Set start timeout of %s for interval %s\n", duration, p)
	select {
	case <-ctx.Done():
		return false
	case <-time.After(duration):
		fmt.Printf("👉 %s Alerting start! (After %s)\n", p.ID, time.Since(now))
		return true
	}
}

func (p *Rule) String() string {
	startTime := p.Start.Format("15:04")
	stopTime := "Never"
	if p.StoppedTime != nil && !p.StoppedTime.IsZero() {
		stopTime = p.StoppedTime.Format("2006/01/02 15:04")
	}
	return fmt.Sprintf("\n\t- ID%s{Days %v at %s for %s target %0.f. Now active: %v, stopped: %s}", p.ID, p.RepeatDays, startTime, p.Duration, p.TargetTemp, p.IsActive, stopTime)
}
