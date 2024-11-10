// To be used when running locally, sends some random values and listens to updates
package main

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"stupid-caldaia/controller/graph/model"
	"stupid-caldaia/controller/store"
	"time"
)

const (
	nSamples            = 10000
	samplesInterval     = 10 * time.Second
	prepopulateDuration = 24 * time.Hour
)

var (
	temperatureSeries     = make([]float64, nSamples)
	humiditySeries        = make([]float64, nSamples)
	seriesIndex       int = 0
)

func generateSamples(count int, min float64, max float64, std float64) []float64 {
	series := make([]float64, count)
	value := 0.0
	for i := 0; i < count; i++ {
		if i == 0 {
			value = min + rand.Float64()*(max-min)
		} else {
			value = value + std*(rand.Float64()-0.5)
		}
		if value > max {
			value = max
		} else if value < min {
			value = min
		}
		series[i] = value
	}
	return series
}

func sendNextSampleForTime(ctx context.Context, sensors map[string]*model.Sensor, time time.Time) {
	if seriesIndex == nSamples {
		seriesIndex = 0
	}

	temperatureSample := temperatureSeries[seriesIndex]
	humiditySample := humiditySeries[seriesIndex]

	sensors["umidita:centrale"].AddSample(ctx, &model.Measure{
		Value:     humiditySample,
		Timestamp: time,
	})
	sensors["temperatura:centrale"].AddSample(ctx, &model.Measure{
		Value:     temperatureSample,
		Timestamp: time,
	})

	seriesIndex++
}

func main() {
	ctx := context.Background()

	config, err := store.LoadConfig()
	if err != nil {
		log.Panic(err)
	}

	_, sensors, boiler := config.CreateObjects(ctx)
	boilerListener, err := boiler.Listen(ctx)
	if err != nil {
		log.Panic(err)
	}

	fmt.Println("🤫 Mock worker started. Prepopulating data...")

	temperatureSeries = generateSamples(nSamples, 15, 25, 0.3)
	humiditySeries = generateSamples(nSamples, 0, 100, 2)
	startTime := time.Now().Add(-prepopulateDuration)
	for t := startTime; t.Before(time.Now()); t = t.Add(samplesInterval) {
		sendNextSampleForTime(ctx, sensors, t)
	}

	fmt.Println("🤫 Sensing...")
	timeout := time.After(10 * time.Second)
	for {
		select {
		case boilerInfo := <-boilerListener:
			switch boilerInfo.State {
			case model.StateOff:
				fmt.Println("🤫 Switching boiler OFF")
			case model.StateOn:
				fmt.Println("🤫 Switching boiler ON")
			}
		case <-timeout:
			timeout = time.After(10 * time.Second)
			sendNextSampleForTime(ctx, sensors, time.Now())
		case <-ctx.Done():
			fmt.Println("Right guys... 👱‍♀️👋")
			break
		}
	}
}
