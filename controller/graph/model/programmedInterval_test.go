package model_test

import (
	"fmt"
	"stupid-caldaia/controller/graph/model"
	"testing"
	"time"
)

func TestWindowStartTime(t *testing.T) {
	startTime := time.Date(2024, 1, 7, 12, 0, 0, 0, time.Local)
	duration := time.Duration(1) * time.Hour

	programmedInterval := &model.Rule{
		Start:      startTime,
		Duration:   duration,
		RepeatDays: []int{int(time.Monday), int(time.Thursday), int(time.Sunday)},
	}

	testCases := []struct {
		name string
		now  time.Time
		want time.Time
	}{
		{
			name: "Now is one hour ago",
			now:  startTime.Add(-time.Duration(1) * time.Hour),
			want: startTime,
		}, {
			name: "After 30 minutes",
			now:  startTime.Add(time.Duration(30) * time.Minute),
			want: startTime,
		},
		{
			name: "After first interval want next day",
			now:  startTime.Add(duration).Add(time.Millisecond),
			want: startTime.Add(time.Duration(24) * time.Hour),
		},
		{
			name: "After a week 30 min before",
			now:  startTime.Add(time.Duration(24) * time.Hour * 7).Add(-time.Duration(30) * time.Minute),
			want: startTime.Add(time.Duration(24) * time.Hour * 7),
		},
		{
			name: "After a week at start time",
			now:  startTime.Add(time.Duration(24) * time.Hour * 7),
			want: startTime.Add(time.Duration(24) * time.Hour * 7),
		},
		{
			name: "After 5 days",
			now:  startTime.Add(time.Duration(24) * time.Hour * 5),
			want: startTime.Add(time.Duration(24) * time.Hour * 7),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := programmedInterval.WindowStartTime(testCase.now)
			if testCase.want != got {
				t.Fatal("Window start time not matching expectation.",
					fmt.Sprintf("Programmed interval: %s", programmedInterval),
					fmt.Sprintf("Wanted: %s, but got %s", testCase.want, got),
					fmt.Sprintf("Now is: %s", testCase.now))
			}
		})
	}
}

func TestShouldBeActive(t *testing.T) {
	now := time.Now()
	oneHourAgo := now.Add(-time.Duration(1) * time.Hour)
	twoHours := time.Duration(2) * time.Hour

	rule := &model.Rule{
		Start:    oneHourAgo,
		Duration: twoHours,
	}

	if rule.ShouldBeStopped() {
		t.Fatal("Rule should not be stopped")
	}
	if !rule.ShouldBeActive() {
		t.Fatal("Rule should be active")
	}

	rule.StoppedTime = &now

	if !rule.ShouldBeStopped() {
		t.Fatal("Rule should be stopped")
	}
	if rule.ShouldBeActive() {
		t.Fatal("Rule should not be active")
	}
}
