package store_test

import (
	"fmt"
	"testing"
	"time"

	"stupid-caldaia/controller/graph/model"
	"stupid-caldaia/controller/internal"
	"stupid-caldaia/controller/store"

	"golang.org/x/net/context"
)

var (
	LAG_FACTOR = 100 // Time in Ms
	SMALL_TIME = time.Duration(LAG_FACTOR/10) * time.Millisecond
	HALF_TIME  = time.Duration(LAG_FACTOR/2) * time.Millisecond
	FULL_TIME  = time.Duration(LAG_FACTOR) * time.Millisecond
)

func TestRuleTimingController(t *testing.T) {
	ctx := context.Background()
	testBoiler, _ := internal.CreateTestBoiler(t, ctx)

	go store.RuleTimingController(ctx, testBoiler)
	time.Sleep(SMALL_TIME)

	programmedInterval, err := testBoiler.SetProgrammedInterval(ctx, &model.ProgrammedInterval{
		Start:      time.Now().Add(SMALL_TIME),
		Duration:   FULL_TIME,
		TargetTemp: internal.MAX_TEMP - 1,
	})
	if err != nil {
		t.Fatal(fmt.Errorf("Could not create programmed interval %w", err))
	}
	if programmedInterval.IsActive {
		t.Fatal("Programmed interval was just created and it's already active, this is unexpected")
	}

	// Sleep sleep...
	time.Sleep(HALF_TIME)
	// Now we expect programmed interval to be active
	info, _ := testBoiler.GetInfo(ctx)
	programmedInterval = info.ProgrammedIntervals[0]
	if !programmedInterval.ShouldBeActive() {
		t.Fatal("Set programmed interval: expecting ShouldBeActive true but it's not")
	}
	if !programmedInterval.IsActive {
		t.Fatal("Set programmed interval: expecting IsActive true but it's not")
	}
	if programmedInterval.ShouldBeStopped() {
		t.Fatal("Set programmed interval: expecting ShouldBeStopped false but it's not")
	}

	// Sleep sleep...
	time.Sleep(FULL_TIME)
	// Now we expect programmed interval to not be active
	info, _ = testBoiler.GetInfo(ctx)
	programmedInterval = info.ProgrammedIntervals[0]
	if programmedInterval.ShouldBeActive() {
		t.Fatal("Set programmed interval: expecting ShouldBeActive false but it's not")
	}
	if programmedInterval.IsActive {
		t.Fatal("Set programmed interval: expecting IsActive false but it's not")
	}
}

func TestRuleTimingControllerEdgeCases(t *testing.T) {
	ctx := context.Background()
	testBoiler, _ := internal.CreateTestBoiler(t, ctx)

	programmedInterval, err := testBoiler.SetProgrammedInterval(ctx, &model.ProgrammedInterval{
		Start:      time.Now(),
		Duration:   FULL_TIME,
		TargetTemp: internal.MAX_TEMP - 1,
	})
	if err != nil {
		t.Fatal(fmt.Errorf("Could not create programmed interval %w", err))
	} else {
		fmt.Printf("Added %s\n", programmedInterval)
	}

	time.Sleep(SMALL_TIME)
	go store.RuleTimingController(ctx, testBoiler)
	time.Sleep(SMALL_TIME)

	info, _ := testBoiler.GetInfo(ctx)
	programmedInterval = info.ProgrammedIntervals[0]
	if !programmedInterval.ShouldBeActive() {
		t.Fatal("Set programmed interval: expecting ShouldBeActive true but it's not")
	}
	if !programmedInterval.IsActive {
		t.Fatal("Set programmed interval: expecting IsActive true but it's not")
	}

	testBoiler.StopProgrammedInterval(ctx, programmedInterval.ID)
	time.Sleep(SMALL_TIME)

	info, _ = testBoiler.GetInfo(ctx)
	programmedInterval = info.ProgrammedIntervals[0]
	if programmedInterval.ShouldBeActive() {
		t.Fatal("Set programmed interval: expecting ShouldBeActive false but it's not")
	}
	if programmedInterval.IsActive {
		t.Fatal("Set programmed interval: expecting IsActive false but it's not")
	}
	if !programmedInterval.ShouldBeStopped() {
		t.Fatal("Set programmed interval: expecting ShouldBeStopped true but it's not")
	}
}
