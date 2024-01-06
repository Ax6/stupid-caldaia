package model_test

import (
	"context"
	"stupid-caldaia/controller/graph/model"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

const (
	MIN_TEMP = 10
	MAX_TEMP = 20
)

func createTestBoiler(t *testing.T, ctx context.Context) (*model.Boiler, error) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: server.Addr(),
	})
	boiler, err := model.NewBoiler(ctx, client, model.BoilerConfig{
		Name:                  "test",
		DefaultMinTemperature: MIN_TEMP,
		DefaultMaxTemperature: MAX_TEMP,
		SwitchPin:             1,
	})
	if err != nil {
		return nil, err
	}
	return boiler, nil
}

func TestSetSwitchOn(t *testing.T) {
	ctx := context.Background()
	testBoiler, err := createTestBoiler(t, ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = testBoiler.Switch(ctx, model.StateOn)
	if err != nil {
		t.Fatal(err)
	}
	info, err := testBoiler.GetInfo(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if info.State != model.StateOn {
		t.Fatal("Boiler should be on")
	}
}

func TestSetMinTemp(t *testing.T) {
	ctx := context.Background()
	testBoiler, err := createTestBoiler(t, ctx)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Set min temp ok", func(t *testing.T) {
		t.Parallel()
		_, err = testBoiler.SetMinTemp(ctx, MIN_TEMP+1)
		if err != nil {
			t.Fatal("Could not set min temp")
		}
	})
	t.Run("Set min temp not ok", func(t *testing.T) {
		t.Parallel()
		_, err = testBoiler.SetMinTemp(ctx, MAX_TEMP+1)
		if err == nil {
			t.Fatal("Shouldn't be able to set temp > MAX_TEMP")
		}
	})
	t.Run("Set max temp ok", func(t *testing.T) {
		t.Parallel()
		_, err = testBoiler.SetMaxTemp(ctx, MAX_TEMP-1)
		if err != nil {
			t.Fatal("Could not set  max temp")
		}
	})
	t.Run("Set max temp not ok", func(t *testing.T) {
		t.Parallel()
		_, err = testBoiler.SetMaxTemp(ctx, MIN_TEMP-1)
		if err == nil {
			t.Fatal("Should not be able to set temp")
		}
	})
}

func TestSetAndDeleteProgrammedInterval(t *testing.T) {
	ctx := context.Background()
	boiler, err := createTestBoiler(t, ctx)
	if err != nil {
		t.Fatal(err)
	}

	_, err = boiler.SetProgrammedInterval(ctx, &model.ProgrammedInterval{
		Start:      time.Now(),
		Duration:   time.Second,
		TargetTemp: MAX_TEMP + 1,
	})
	if err == nil {
		t.Fatal("Souldn't be able to set target temperature above limit")
	}

	programmedInterval, err := boiler.SetProgrammedInterval(ctx, &model.ProgrammedInterval{
		Start:      time.Now(),
		Duration:   time.Second,
		TargetTemp: MAX_TEMP,
	})
	if err != nil {
		t.Fatal(err)
	}

	programmedIntervals, err := boiler.GetProgrammedIntervals(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, ok := programmedIntervals[programmedInterval.ID]
	if !ok {
		t.Fatal("Programmed interval was not created")
	}

	_, err = boiler.DeleteProgrammedInterval(ctx, programmedInterval.ID)
	if err != nil {
		t.Fatal(err)
	}

	programmedIntervals, err = boiler.GetProgrammedIntervals(ctx)
	_, ok = programmedIntervals[programmedInterval.ID]
	if ok {
		t.Fatal("Did not delete programmed interval")
	}
}
