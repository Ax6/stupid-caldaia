package main

import (
	"context"
	"log"
	"stupid-caldaia/controller/graph/model"
	"stupid-caldaia/controller/store"
	"time"

	"github.com/parMaster/htu21"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/host/v3"
)

var (
	htu21Data   physic.Env
	htu21Device *htu21.Dev
)

func main() {

	ctx := context.Background()

	config, err := store.LoadConfig()
	if err != nil {
		panic(err)
	}

	_, sensors := config.CreateObjects(ctx)

	for {
		// Preparing to read sensor
		if _, err := host.Init(); err != nil {
			log.Fatal(err)
		}

		// Use i2creg I²C bus registry to find the first available I²C bus.
		b, err := i2creg.Open("1")
		if err != nil {
			log.Fatalf("failed to open I²C: %v", err)
		}

		htu21Device, err = htu21.NewI2C(b, 0x40)
		if err != nil {
			log.Fatalf("failed to initialize htu21: %v", err)
		}

		if err := htu21Device.Sense(&htu21Data); err != nil {
			log.Fatal(err)
		}
		// That's it!

		// Since it implements String() interface:
		log.Println(htu21Data.Temperature)
		log.Println(htu21Data.Humidity)

		// periph.io physic package allows do thing like:
		log.Println(htu21Data.Temperature.Celsius())
		log.Printf("Temperature: %.2f˚C", htu21Data.Temperature.Celsius())
		log.Printf("Temperature: %.2fF", htu21Data.Temperature.Fahrenheit())

		log.Printf("Humidity (milliRH): %s", htu21Data.Humidity/physic.MilliRH*physic.PercentRH)

		measure := model.Measure{
			Timestamp: time.Now(),
			Value:     htu21Data.Temperature.Celsius(),
		}
		sensors["temperatura:centrale"].AddSample(ctx, &measure)

		time.Sleep(1 * time.Second)
	}
}
