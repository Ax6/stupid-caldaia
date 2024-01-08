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

func getRule(ctx context.Context, c *model.Boiler, id string) *model.Rule {
	info, err := c.GetInfo(ctx)
	if err != nil {
		panic(err)
	}
	for _, programmedInterval := range info.Rules {
		err = fmt.Errorf("Could not find programmedInterval with id: %s", id)
		if programmedInterval.ID == id {
			return programmedInterval
		}
	}
	panic(err)
}

func TestRuleTimingControllerBasic(t *testing.T) {
	ctx := context.Background()
	testBoiler, _ := internal.CreateTestBoiler(t, ctx)

	go store.RuleTimingController(ctx, testBoiler)
	time.Sleep(SMALL_TIME)

	programmedInterval, err := testBoiler.SetRule(ctx, &model.Rule{
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
	programmedInterval = info.Rules[0]
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
	programmedInterval = info.Rules[0]
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

	originalRule, err := testBoiler.SetRule(ctx, &model.Rule{
		Start:      time.Now(),
		Duration:   FULL_TIME,
		TargetTemp: internal.MAX_TEMP - 1,
	})
	if err != nil {
		t.Fatal(fmt.Errorf("Could not create programmed interval %w", err))
	} else {
		fmt.Printf("Added %s\n", originalRule)
	}

	time.Sleep(SMALL_TIME)
	go store.RuleTimingController(ctx, testBoiler)
	time.Sleep(SMALL_TIME)

	info, _ := testBoiler.GetInfo(ctx)
	programmedInterval := info.Rules[0]
	if !programmedInterval.ShouldBeActive() {
		t.Fatalf("Set programmed interval: expecting ShouldBeActive true but it's not %s\nOriginal:%s\n", programmedInterval, originalRule)
	}
	if !programmedInterval.IsActive {
		t.Fatalf("Set programmed interval: expecting IsActive true but it's not %s\nOriginal:%s\n", programmedInterval, originalRule)
	}

	testBoiler.StopRule(ctx, programmedInterval.ID)
	time.Sleep(SMALL_TIME)

	info, _ = testBoiler.GetInfo(ctx)
	programmedInterval = info.Rules[0]
	if programmedInterval.ShouldBeActive() {
		t.Fatalf("Set programmed interval: expecting ShouldBeActive false but it's not %s\nOriginal:%s\n", programmedInterval, originalRule)
	}
	if programmedInterval.IsActive {
		t.Fatalf("Set programmed interval: expecting IsActive false but it's not %s\nOriginal:%s\n", programmedInterval, originalRule)
	}
	if !programmedInterval.ShouldBeStopped() {
		t.Fatalf("Set programmed interval: expecting ShouldBeStopped true but it's  not %s\nOriginal:%s\n", programmedInterval, originalRule)
	}
}

func TestRuleTimingControllerMultipleRules(t *testing.T) {
	ctx := context.Background()
	testBoiler, err := internal.CreateTestBoiler(t, ctx)
	if err != nil {
		t.Fatal("Could not create test boiler %w", err)
	}
	go store.RuleTimingController(ctx, testBoiler)
	now := time.Now()

	p_first, _ := testBoiler.SetRule(ctx, &model.Rule{
		Start:      now,
		Duration:   FULL_TIME,
		TargetTemp: internal.MAX_TEMP - 3,
		RepeatDays: []int{0, 1, 2, 3, 4, 5, 6, 7},
	})

	p_during_first, _ := testBoiler.SetRule(ctx, &model.Rule{
		Start:      now.Add(HALF_TIME),
		Duration:   FULL_TIME,
		TargetTemp: internal.MAX_TEMP - 2,
	})

	p_after_the_others, _ := testBoiler.SetRule(ctx, &model.Rule{
		Start:      now.Add(2 * FULL_TIME),
		Duration:   HALF_TIME,
		TargetTemp: internal.MAX_TEMP - 1,
		RepeatDays: []int{0, 1, 2, 3, 4, 5, 6, 7},
	})

	time.Sleep(SMALL_TIME)
	time.Sleep(SMALL_TIME)
	p_first = getRule(ctx, testBoiler, p_first.ID)
	p_during_first = getRule(ctx, testBoiler, p_during_first.ID)
	p_after_the_others = getRule(ctx, testBoiler, p_after_the_others.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !p_first.IsActive {
		t.Fatalf("%d Expecting p_first to be active but is not %s", time.Now().UnixMilli(), p_first)
	}
	if p_during_first.IsActive {
		t.Fatal("Expecting p_during_first to not be active but it is")
	}
	if p_after_the_others.IsActive {
		t.Fatal("Expecting p_after_the_others to not be active but it is")
	}

	time.Sleep(HALF_TIME)
	p_first = getRule(ctx, testBoiler, p_first.ID)
	p_during_first = getRule(ctx, testBoiler, p_during_first.ID)
	p_after_the_others = getRule(ctx, testBoiler, p_after_the_others.ID)
	if !p_first.IsActive {
		t.Fatal("Expecting p_first to be active but is not")
	}
	if !p_during_first.IsActive {
		t.Fatal("Expecting p_during_first to be active but it is not")
	}
	if p_after_the_others.IsActive {
		t.Fatal("Expecting p_after_the_others to not be active but it is")
	}

	testBoiler.StopRule(ctx, p_during_first.ID)
	time.Sleep(SMALL_TIME)
	p_first = getRule(ctx, testBoiler, p_first.ID)
	p_during_first = getRule(ctx, testBoiler, p_during_first.ID)
	p_after_the_others = getRule(ctx, testBoiler, p_after_the_others.ID)
	if !p_first.IsActive {
		t.Fatal("Expecting p_first to be active but is not")
	}
	if p_during_first.IsActive {
		t.Fatal("Expecting p_during_first to not be active but it is")
	}
	if p_after_the_others.IsActive {
		t.Fatal("Expecting p_after_the_others to not be active but it is")
	}

	time.Sleep(HALF_TIME)
	p_first = getRule(ctx, testBoiler, p_first.ID)
	p_during_first = getRule(ctx, testBoiler, p_during_first.ID)
	p_after_the_others = getRule(ctx, testBoiler, p_after_the_others.ID)
	if p_first.IsActive {
		t.Fatal("Expecting p_first to not be active but it is")
	}
	if p_during_first.IsActive {
		t.Fatal("Expecting p_during_first to not be active but it is")
	}
	if p_after_the_others.IsActive {
		t.Fatal("Expecting p_after_the_others to not be active but it is")
	}

	time.Sleep(FULL_TIME)
	p_first = getRule(ctx, testBoiler, p_first.ID)
	p_during_first = getRule(ctx, testBoiler, p_during_first.ID)
	p_after_the_others = getRule(ctx, testBoiler, p_after_the_others.ID)
	if p_first.IsActive {
		t.Fatal("Expecting p_first to not be active but it is")
	}
	if p_during_first.IsActive {
		t.Fatal("Expecting p_during_first to not be active but it is")
	}
	if !p_after_the_others.IsActive {
		t.Fatal("Expecting p_after_the_others to be active but it is not")
	}
}
