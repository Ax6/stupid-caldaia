package model

import (
	"context"
	"fmt"
	"time"
)

type DayOfWeek = time.Weekday

// It should be active if now it's in the current time window and we shouldn't stop
func (p *ProgrammedInterval) ShouldBeActive() bool {
	now := time.Now()
	wStart := p.WindowStartTime(now)
	wEnd := wStart.Add(p.Duration)
	return now.After(wStart) && now.Before(wEnd) && !p.ShouldBeStopped()
}

// It should stop if the stop command was sent in the current or upcoming window and we are past that time
func (p *ProgrammedInterval) ShouldBeStopped() bool {
	sTime := p.StoppedTime
	if sTime.IsZero() {
		return false
	}
	now := time.Now()
	wStart := p.WindowStartTime(now)
	wEnd := wStart.Add(p.Duration)
	return sTime.Before(now) && sTime.After(wStart) && sTime.Before(wEnd)
}

// Relative to the referenceTime, if we are in a window returns the start time
// of this window, otherwise returns the start of the upcoming window
func (p *ProgrammedInterval) WindowStartTime(referenceTime time.Time) time.Time {
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
			if todaysStart.Add(p.Duration).Before(now) {
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

func (p *ProgrammedInterval) WindowStopTimeout(ctx context.Context, alert chan<- *ProgrammedInterval) {
	now := time.Now()
	duration := p.WindowStartTime(now).Sub(now) + p.Duration
	fmt.Printf("⏰ Set stop timeout of %s for interval %s\n", duration, p)
	select {
	case <-ctx.Done():
		fmt.Printf("%s Context shut itself 😱 after %s\n", p.ID, time.Now().Sub(now))
		return // Timeout was cancelled
	case <-time.After(duration):
		fmt.Printf("✋ %s Alerting stop! (After %s)\n", p.ID, time.Now().Sub(now))
		alert <- p
		return
	}
}

func (p *ProgrammedInterval) WindowStartTimeout(ctx context.Context, alert chan<- *ProgrammedInterval) {
	now := time.Now()
	duration := p.WindowStartTime(now).Sub(now)
	fmt.Printf("⏰ Set start timeout of %s for interval %s\n", duration, p)
	select {
	case <-ctx.Done():
		fmt.Printf("%s Context shut itself 😱 after %s\n", p.ID, time.Now().Sub(now))
		return // Timeout was cancelled
	case <-time.After(duration):
		fmt.Printf("👉 %s Alerting start! (After %s)\n", p.ID, time.Now().Sub(now))
		alert <- p
		return
	}
}

func (p *ProgrammedInterval) String() string {
	startTime := p.Start.Format("15:04")
	stopTime := "Never"
	if !p.StoppedTime.IsZero() {
		stopTime = p.StoppedTime.Format("2006/01/02 15:04")
	}
	return fmt.Sprintf("\n\t- ID%s{Days %v at %s for %s target %0.f. Now active: %v, stopped: %s}", p.ID, p.RepeatDays, startTime, p.Duration, p.TargetTemp, p.IsActive, stopTime)
}
