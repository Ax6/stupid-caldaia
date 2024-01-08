package model_test

import (
	"context"
	"stupid-caldaia/controller/graph/model"
	"stupid-caldaia/controller/internal"
	"testing"
	"time"
)

func TestSetSwitchOn(t *testing.T) {
	ctx := context.Background()
	testBoiler, err := internal.CreateTestBoiler(t, ctx)
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
	testBoiler, err := internal.CreateTestBoiler(t, ctx)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Set min temp ok", func(t *testing.T) {
		t.Parallel()
		_, err = testBoiler.SetMinTemp(ctx, internal.MIN_TEMP+1)
		if err != nil {
			t.Fatal("Could not set min temp")
		}
	})
	t.Run("Set min temp not ok", func(t *testing.T) {
		t.Parallel()
		_, err = testBoiler.SetMinTemp(ctx, internal.MAX_TEMP+1)
		if err == nil {
			t.Fatal("Shouldn't be able to set temp > MAX_TEMP")
		}
	})
	t.Run("Set max temp ok", func(t *testing.T) {
		t.Parallel()
		_, err = testBoiler.SetMaxTemp(ctx, internal.MAX_TEMP-1)
		if err != nil {
			t.Fatal("Could not set  max temp")
		}
	})
	t.Run("Set max temp not ok", func(t *testing.T) {
		t.Parallel()
		_, err = testBoiler.SetMaxTemp(ctx, internal.MIN_TEMP-1)
		if err == nil {
			t.Fatal("Should not be able to set temp")
		}
	})
}

func TestSetAndDeleteRule(t *testing.T) {
	ctx := context.Background()
	boiler, err := internal.CreateTestBoiler(t, ctx)
	if err != nil {
		t.Fatal(err)
	}

	_, err = boiler.SetRule(ctx, &model.Rule{
		Start:      time.Now(),
		Duration:   time.Second,
		TargetTemp: internal.MAX_TEMP + 1,
	})
	if err == nil {
		t.Fatal("Souldn't be able to set target temperature above limit")
	}

	programmedIntervalUnderTest, err := boiler.SetRule(ctx, &model.Rule{
		Start:      time.Now(),
		Duration:   time.Second,
		TargetTemp: internal.MAX_TEMP,
	})
	if err != nil {
		t.Fatal(err)
	}

	boilerInfo, _ := boiler.GetInfo(ctx)
	if len(boilerInfo.Rules) != 1 || boilerInfo.Rules[0].ID != programmedIntervalUnderTest.ID {
		t.Fatal("Programmed interval was not added correctly")
	}

	err = boiler.DeleteRule(ctx, programmedIntervalUnderTest.ID)
	if err != nil {
		t.Fatal(err)
	}
	boilerInfo, _ = boiler.GetInfo(ctx)
	if len(boilerInfo.Rules) != 0 {
		t.Fatal("Programmed interval was note deleted")
	}
}
