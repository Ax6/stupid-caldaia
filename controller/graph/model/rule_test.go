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
	testCases := []struct {
		name      string
		have      *model.Rule
		firstWant bool
		thenWant  bool
	}{
		{
			name: "rule current active",
			have: &model.Rule{
				Start:    now.Add(-time.Minute),
				Duration: time.Hour,
			},
			firstWant: true,
			thenWant:  false,
		},
		{
			name: "rule future active",
			have: &model.Rule{
				Start:    now.Add(time.Hour),
				Duration: time.Hour,
			},
			firstWant: false,
			thenWant:  false,
		},
		{
			name: "rule past active",
			have: &model.Rule{
				Start:    now.Add(-time.Hour),
				Duration: time.Minute,
			},
			firstWant: false,
			thenWant:  false,
		},
		{
			name: "rule current active current delay",
			have: &model.Rule{
				Start:    now.Add(-time.Minute),
				Duration: time.Hour,
				Delay:    time.Hour,
			},
			firstWant: true,
			thenWant:  false,
		},
		{
			name: "rule current active past delay",
			have: &model.Rule{
				Start:    now.Add(-time.Hour),
				Duration: time.Hour,
				Delay:    time.Minute,
			},
			firstWant: true,
			thenWant:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			shouldBe := tc.have.ShouldBeActive()
			if shouldBe != tc.firstWant {
				t.Fatalf("Rule ShouldBeActive is %v while expected is %v", shouldBe, tc.firstWant)
			}
			tc.have.StoppedTime = &now
			shouldBe = tc.have.ShouldBeActive()
			if shouldBe != tc.thenWant {
				t.Fatalf("Rule ShouldBeActive is %v while expected is %v", shouldBe, tc.firstWant)
			}
		})
	}
}

func TestShouldBeStopped(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name      string
		have      *model.Rule
		firstWant bool
		thenWant  bool
	}{
		{
			name: "rule current active",
			have: &model.Rule{
				Start:    now.Add(-time.Minute),
				Duration: time.Hour,
			},
			firstWant: false,
			thenWant:  true,
		},
		{
			name: "rule future active",
			have: &model.Rule{
				Start:    now.Add(time.Hour),
				Duration: time.Hour,
			},
			firstWant: false,
			thenWant:  false,
		},
		{
			name: "rule past active",
			have: &model.Rule{
				Start:    now.Add(-time.Hour),
				Duration: time.Minute,
			},
			firstWant: false,
			thenWant:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			shouldBe := tc.have.ShouldBeStopped()
			if shouldBe != tc.firstWant {
				t.Fatalf("Rule ShouldBeStopped is %v while expected is %v", shouldBe, tc.firstWant)
			}
			tc.have.StoppedTime = &now
			shouldBe = tc.have.ShouldBeStopped()
			if shouldBe != tc.thenWant {
				t.Fatalf("Rule ShouldBeStopped is %v while expected is %v", shouldBe, tc.firstWant)
			}
		})
	}
}

func TestIsBeingDelayed(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name string
		have *model.Rule
		want bool
	}{
		{
			name: "rule current active no delay",
			have: &model.Rule{
				Start:    now.Add(-time.Minute),
				Duration: time.Hour,
			},
			want: false,
		},
		{
			name: "rule current active current delayed",
			have: &model.Rule{
				Start:    now.Add(-time.Minute),
				Duration: time.Hour,
				Delay:    time.Hour,
			},
			want: true,
		},
		{
			name: "rule current active past delayed",
			have: &model.Rule{
				Start:    now.Add(-time.Hour),
				Duration: time.Hour,
				Delay:    time.Minute,
			},
			want: false,
		},
		{
			name: "rule inactive future delayed",
			have: &model.Rule{
				Start:    now.Add(time.Minute),
				Duration: time.Hour,
				Delay:    time.Hour,
			},
			want: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			beingDelayed := tc.have.IsBeingDelayed()
			if beingDelayed != tc.want {
				t.Fatalf("Rule IsBeingDelayed is %v while expected is %v", beingDelayed, tc.want)
			}
		})
	}
}
