package model

import (
	"math"
	"testing"
	"time"
)

var (
	bt0      = time.Now()
	bEndTime = bt0.Add(24 * time.Hour)
	// Benchmarking typical application. About 8 switches in 24 hours
	benchmarkSamples = []SwitchSample{
		{Time: bt0, State: StateOff},
		{Time: bt0.Add(time.Hour), State: StateOn},
		{Time: bt0.Add(2 * time.Hour), State: StateOff},
		{Time: bt0.Add(6 * time.Hour), State: StateOn},
		{Time: bt0.Add(7 * time.Hour), State: StateOff},
		{Time: bt0.Add(13 * time.Hour), State: StateOn},
		{Time: bt0.Add(14 * time.Hour), State: StateOff},
		{Time: bt0.Add(20 * time.Hour), State: StateOn},
		{Time: bt0.Add(23 * time.Hour), State: StateOff},
	}
)

func BenchmarkCalculateOverheatingIndex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		calculateOverheatingIndex(benchmarkSamples, bEndTime)
	}
}

func BenchmarkCalculateOverheatingIndex_TransferFunction(b *testing.B) {
	for i := 0; i < b.N; i++ {
		calculateOverheatingIndex_TransferFunction(benchmarkSamples, bEndTime)
	}
}

func TestCalculateOverheatingIndex(t *testing.T) {
	t0 := time.Now()

	testCases := []struct {
		name      string
		have      []SwitchSample
		endTime   time.Time
		want      float64
		precision float64
	}{
		{
			name: "Caught in the act",
			have: []SwitchSample{
				{
					Time:  t0,
					State: StateOn,
				},
			},
			endTime:   t0.Add(OH_TAU * time.Second),
			want:      0.632,
			precision: 0.001,
		},
		{
			name: "Normal OFF - ON",
			have: []SwitchSample{
				{
					Time:  t0,
					State: StateOff,
				},
				{
					Time:  t0.Add(10 * time.Second),
					State: StateOn,
				},
			},
			endTime:   t0.Add(24 * time.Hour),
			want:      1,
			precision: 0.001,
		},
		{
			name: "Normal ON - OFF",
			have: []SwitchSample{
				{
					Time:  t0,
					State: StateOn,
				},
				{
					Time:  t0.Add(5 * time.Hour),
					State: StateOff,
				},
			},
			endTime:   t0.Add(24 * time.Hour),
			want:      0,
			precision: 0.001,
		},
	}

	for _, ts := range testCases {
		t.Run(ts.name+"_calculate", func(t *testing.T) {
			got := calculateOverheatingIndex(ts.have, ts.endTime)
			if math.Abs(got-ts.want) > ts.precision {
				t.Fatalf("Failed to calculate overheating index. Wanted %.2f but got %.2f", ts.want, got)
			}
		})
	}
}
