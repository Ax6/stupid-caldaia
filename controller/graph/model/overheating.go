package model

import (
	"context"
	"math"
	"time"
)

const (
	OH_TAU     = 720.0
	OH_FS      = 1.0
	OH_HISTORY = 10 * OH_TAU
)

func GetCurrentOverheatingIndex(ctx context.Context, boiler *Boiler) (float64, error) {
	endTime := time.Now()
	startTime := endTime.Add(-OH_HISTORY * time.Second)
	samples, err := boiler.GetSwitchHistory(ctx, startTime, endTime)
	if err != nil {
		return 0, err
	}
	if len(samples) == 0 {
		// If we have no samples, we want at least one at the start so we can drive the index calculation
		info, err := boiler.GetInfo(ctx)
		if err != nil {
			return 0, err
		}
		samples = append(samples, &SwitchSample{
			Time:  startTime,
			State: info.State,
		})
	}
	return calculateOverheatingIndex(samples, endTime), nil
}

// As it turns out, with just a few On/Off samples over a long time period it is
// much more efficient to simply do the maths and precisely calculate the index
// rather than applying a transfer function.
func calculateOverheatingIndex(samples []*SwitchSample, endTime time.Time) float64 {
	if len(samples) == 0 {
		return 0
	}

	y := 0.0
	for sampleIndex, sample := range samples {
		fromTime := sample.Time
		toTime := endTime
		if sampleIndex < len(samples)-1 {
			toTime = samples[sampleIndex+1].Time
		}

		yn := 0.0
		if sample.State == StateOn {
			yn = 1.0
		}

		y = yn + (y-yn)*math.Exp(-(toTime.Sub(fromTime).Seconds()/OH_TAU))

	}
	return y
}

// Here just for reference and benchmark. This uses the discrete transfer
// function and is much less efficient
func calculateOverheatingIndex_TransferFunction(samples []*SwitchSample, endTime time.Time) float64 {
	if len(samples) == 0 {
		return 0
	}

	// Resample to FS
	sampleIndex := 0
	t := samples[sampleIndex].Time
	totalSteps := int(endTime.Sub(t).Seconds() * OH_FS)
	x := make([]float64, totalSteps)
	nextSampleTime := endTime
	if sampleIndex < len(samples)-1 {
		nextSampleTime = samples[sampleIndex+1].Time
	}
	for i := 0; i < totalSteps; i++ {
		switch samples[sampleIndex].State {
		case StateOn:
			x[i] = 1
		default:
			x[i] = 0
		}
		if t.After(nextSampleTime) {
			sampleIndex++
			nextSampleTime = endTime
			if sampleIndex < len(samples)-1 {
				nextSampleTime = samples[sampleIndex+1].Time
			}
		}
		t = t.Add(time.Second / OH_FS)
	}

	// Actually run transfer function
	yxn := 0.0 // Assume initial state is 0
	ux0 := 0.0
	alfa := (2*OH_FS*OH_TAU - 1)
	beta := (2*OH_FS*OH_TAU + 1)
	for _, uxn := range x {
		yxn = (1/beta)*(uxn+ux0) + yxn*(alfa/beta)
		ux0 = uxn
	}
	return yxn
}
