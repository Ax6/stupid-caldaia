package model_test

import (
	"context"
	"stupid-caldaia/controller/graph/model"
	"stupid-caldaia/controller/testutils"
	"testing"
	"time"
)

func TestSetSwitchOn(t *testing.T) {
	ctx := context.Background()
	testBoiler, err := testutils.CreateTestBoiler(ctx, t)
	if err != nil {
		t.Fatal(err)
	}
	_, err = testBoiler.Switch(ctx, model.StateOn)
	if err != nil {
		t.Fatalf("Could not switch boiler on: %v", err)
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
	testBoiler, err := testutils.CreateTestBoiler(ctx, t)
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name    string
		temp    float64
		wantErr bool
	}{
		{
			name:    "Set min temp ok",
			temp:    testutils.MIN_TEMP + 1,
			wantErr: false,
		},
		{
			name:    "Set min temp not ok",
			temp:    testutils.MAX_TEMP + 1,
			wantErr: true,
		},
		{
			name:    "Set max temp ok",
			temp:    testutils.MAX_TEMP - 1,
			wantErr: false,
		},
		{
			name:    "Set max temp not ok",
			temp:    testutils.MIN_TEMP - 1,
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err = testBoiler.SetMinTemp(ctx, tc.temp)
			if tc.wantErr && err == nil {
				t.Fatal("Shouldn't be able to set given temp")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("Could not set temp: %v", err)
			}
		})
	}
}

func TestSetAndDeleteRule(t *testing.T) {
	ctx := context.Background()
	boiler, err := testutils.CreateTestBoiler(ctx, t)
	if err != nil {
		t.Fatal(err)
	}

	_, err = boiler.SetRule(ctx, &model.Rule{
		Start:      time.Now(),
		Duration:   time.Second,
		TargetTemp: testutils.MAX_TEMP + 1,
	})
	if err == nil {
		t.Fatal("Souldn't be able to set target temperature above limit")
	}

	ruleUnderTest, err := boiler.SetRule(ctx, &model.Rule{
		Start:      time.Now(),
		Duration:   time.Second,
		TargetTemp: testutils.MAX_TEMP,
	})
	if err != nil {
		t.Fatal(err)
	}

	boilerInfo, _ := boiler.GetInfo(ctx)
	if len(boilerInfo.Rules) != 1 || boilerInfo.Rules[0].ID != ruleUnderTest.ID {
		t.Fatal("Programmed interval was not added correctly")
	}

	err = boiler.DeleteRule(ctx, ruleUnderTest.ID)
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
	boiler, err := testutils.CreateTestBoiler(ctx, t)
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
		TargetTemp: testutils.MAX_TEMP,
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
		TargetTemp: testutils.MAX_TEMP,
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

func TestGetSwitchHistory(t *testing.T) {
	ctx := context.Background()
	boiler, err := testutils.CreateTestBoiler(ctx, t)
	if err != nil {
		t.Fatal(err)
	}
	iWant := []model.State{model.StateOff, model.StateOn, model.StateOff}
	for _, state := range iWant {
		boiler.Switch(ctx, state)
		time.Sleep(time.Millisecond)
	}

	aSecondAgo := time.Now().Add(-time.Hour)
	samples, err := boiler.GetSwitchHistory(ctx, aSecondAgo, time.Now())
	if len(samples) != len(iWant) {
		t.Fatalf("Expected %d states but got %d", len(iWant), len(samples))
	}
	previousTime := aSecondAgo
	for i, sample := range samples {
		if sample.State != iWant[i] {
			t.Fatalf("Expected sample %d to be %v but got %v", i, iWant[i], sample.State)
		}
		if sample.Time.After(previousTime) {
			previousTime = sample.Time
		} else {
			t.Fatalf("Expected time increase but got less or same")
		}
	}
}
