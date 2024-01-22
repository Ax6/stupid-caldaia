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

func TestSetRuleAndUpdate(t *testing.T) {
	ctx := context.Background()
	boiler, err := internal.CreateTestBoiler(t, ctx)
	if err != nil {
		t.Fatal(err)
	}

	rulesListener, err := boiler.ListenRules(ctx)
	if err != nil {
		t.Fatal(err)
	}
	boilerListener, err := boiler.Listen(ctx)
	if err != nil {
		t.Fatal(err)
	}

	rule, err := boiler.SetRule(ctx, &model.Rule{
		Start:      time.Now(),
		Duration:   time.Second,
		TargetTemp: internal.MAX_TEMP,
	})
	if err != nil {
		t.Fatal("Could not set rule")
	}

	select {
	case msg := <-boilerListener:
		if msg.Rules[0].ID != rule.ID {
			t.Fatal("Expected same rule in return but ID was different")
		}
		break
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for message")
	}
	select {
	case msg := <-rulesListener:
		if msg[0].ID != rule.ID {
			t.Fatal("Expected same rule in return but ID was different")
		}
		break
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for message")
	}

	if rule.ID == "" {
		t.Fatal("Expected some sort of ID for this rule")
	}
	if rule.Duration != time.Second {
		t.Fatal("Want 1 second but duration is different")
	}

	updatedRule, err := boiler.SetRule(ctx, &model.Rule{
		ID:         rule.ID,
		Start:      time.Now(),
		Duration:   time.Hour,
		TargetTemp: internal.MAX_TEMP,
	})
	if err != nil {
		t.Fatal(err)
	}

	select {
	case msg := <-boilerListener:
		if msg.Rules[0].ID != rule.ID {
			t.Fatal("Expected same rule in return but ID was different")
		}
		break
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for message")
	}
	select {
	case msg := <-rulesListener:
		if msg[0].ID != rule.ID {
			t.Fatal("Expected same rule in return but ID was different")
		}
		break
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for message")
	}

	if updatedRule.ID != rule.ID {
		t.Fatal("Expected same rule in return but ID was different")
	}
	if updatedRule.Duration != time.Hour {
		t.Fatal("Want 1 hour but duration is different")
	}
}
