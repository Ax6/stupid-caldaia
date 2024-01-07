package model

import (
	"context"
	"fmt"
	"time"
)

type DayOfWeek = time.Weekday

// It should be active if now it's in the current time window and we shouldn't stop
func (p *ProgrammedInterval) ShouldBeActive() bool {
	wStart := p.WindowStartTime()
	wEnd := wStart.Add(p.Duration)
	now := time.Now()
	return now.After(wStart) && now.Before(wEnd) && !p.ShouldBeStopped()
}

// It should stop if the stop command was sent in the current or upcoming window and we are past that time
func (p *ProgrammedInterval) ShouldBeStopped() bool {
	sTime := p.StoppedTime
	if sTime.IsZero() {
		return false
	}
	wStart := p.WindowStartTime()
	wEnd := wStart.Add(p.Duration)
	return sTime.Before(time.Now()) && sTime.After(wStart) && sTime.Before(wEnd)
}

// If we are in a window returns the start time of this window, otherwise returns the start of the upcoming window
func (p *ProgrammedInterval) WindowStartTime() time.Time {
	if len(p.RepeatDays) == 0 {
		return p.Start
	}

	now := time.Now()
	daysUntilTarget := 7
	todaysStart := time.Date(now.Year(), now.Month(), now.Day(), p.Start.Hour(), p.Start.Minute(), p.Start.Second(), 0, p.Start.Location())

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
	select {
	case <-ctx.Done():
		return // Timeout was cancelled
	case <-time.After(p.WindowStartTime().Sub(time.Now()) + p.Duration):
		alert <- p
		return
	}
}

func (p *ProgrammedInterval) WindowStartTimeout(ctx context.Context, alert chan<- *ProgrammedInterval) {
	now := time.Now()
	fmt.Printf("⏰ Set start timeout for programmed interval %s. Requested start in %s\n", p.ID, p.WindowStartTime().Sub(time.Now()))
	select {
	case <-ctx.Done():
		fmt.Printf("Context shut itself 😱 after %s\n", time.Now().Sub(now))
		return // Timeout was cancelled
	case <-time.After(p.WindowStartTime().Sub(time.Now())):
		fmt.Printf("😇 Alerting start! (After %s)\n", time.Now().Sub(now))
		alert <- p
		return
	}
}
