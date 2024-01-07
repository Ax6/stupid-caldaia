package model

import (
	"context"
	"time"
)

type DayOfWeek = time.Weekday

// It should be active if now it's in the current time window and we shouldn't stop
func (p *ProgrammedInterval) ShouldBeActive() bool {
	wStart := p.CurrentWindowStart()
	if wStart.IsZero() {
		return false
	}
	now := time.Now()
	wEnd := wStart.Add(p.Duration)
	return now.After(wStart) && now.Before(wEnd) && !p.ShouldBeStopped()
}

// It should stop if the stop command was sent in the current window
func (p *ProgrammedInterval) ShouldBeStopped() bool {
	wStart := p.CurrentWindowStart()
	sTime := p.StoppedTime
	if wStart.IsZero() || sTime.IsZero() {
		return false
	}
	wEnd := wStart.Add(p.Duration)
	return sTime.After(wStart) && sTime.Before(wEnd)
}

func (p *ProgrammedInterval) CurrentWindowStart() time.Time {
	now := time.Now()
	if len(p.RepeatDays) == 0 {
		return p.Start
	}

	daysUntilTarget := 7
	for _, programmedWeekDay := range p.RepeatDays {
		// Calculate the difference in days between the current day and the target day
		daysAway := (programmedWeekDay - int(now.Weekday()) + 7) % 7
		if daysAway < daysUntilTarget {
			daysUntilTarget = daysAway
		}
	}

	currentWeekDay := now.Weekday()
	for _, programmedWeekDay := range p.RepeatDays {
		if currentWeekDay == time.Weekday(programmedWeekDay) {
			// We have to check if the repeated rule applies at this time of day
			todaysStart := time.Date(now.Year(), now.Month(), now.Day(), p.Start.Hour(), p.Start.Minute(), p.Start.Second(), 0, now.Location())
			return todaysStart
		}
	}
	return time.Time{}
}

func (p *ProgrammedInterval) WindowTimeout(ctx context.Context, alert chan *ProgrammedInterval) {
	select {
	case <-ctx.Done():
		return // Timeout was cancelled
	case <-time.After(p.Start.Sub(time.Now()) + p.Duration):
		alert <- p
		return
	}
}
