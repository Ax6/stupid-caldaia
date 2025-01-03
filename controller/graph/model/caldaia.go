package model

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/redis/go-redis/v9"
)

type BoilerConfig struct {
	Name                  string
	DefaultMinTemperature float64
	DefaultMaxTemperature float64
	SwitchPin             int
}

type Boiler struct {
	Config              BoilerConfig
	client              *redis.Client
	lock                sync.Mutex
	stateUpdateCancel   context.CancelFunc
	switchSeriesKey     string
	protectionSeriesKey string
}

const (
	stateUpdateBatchingTime = 1000 // In microseconds
)

func GetStateIndex(state State) int {
	for i, item := range AllState {
		if item == state {
			return i
		}
	}
	return -1
}

func NewBoiler(ctx context.Context, client *redis.Client, config BoilerConfig) (*Boiler, error) {
	boiler := Boiler{
		Config:              config,
		client:              client,
		switchSeriesKey:     "switch:" + config.Name,
		protectionSeriesKey: "overheating:" + config.Name,
	}
	_, err := boiler.GetInfo(ctx)

	// Check if switch state series already exists
	exists, err := client.Exists(ctx, boiler.switchSeriesKey).Result()
	if exists == 0 {
		_, err := client.TSCreateWithArgs(ctx, boiler.switchSeriesKey, &redis.TSOptions{}).Result()
		if err != nil {
			return &boiler, err
		}
	}

	exists, err = client.Exists(ctx, boiler.protectionSeriesKey).Result()
	if exists == 0 {
		_, err := client.TSCreateWithArgs(ctx, boiler.protectionSeriesKey, &redis.TSOptions{}).Result()
		if err != nil {
			return &boiler, err
		}
	}
	return &boiler, err
}

// Function to switch the relay on or off
// Accepts only two values: "on" or "off"
func (c *Boiler) Switch(ctx context.Context, targetState State) (*State, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	info, err := c.GetInfo(ctx)
	if err != nil {
		return nil, err
	}
	if targetState != StateOn && targetState != StateOff {
		return &targetState, fmt.Errorf("invalid state to set")
	}
	info.State = targetState
	err = c.save(ctx, info)
	return &info.State, err
}

func (c *Boiler) SetOverheating(ctx context.Context, isOverheating bool) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	info, err := c.GetInfo(ctx)
	if err != nil {
		return err
	}
	info.IsOverheatingProtectionActive = isOverheating
	err = c.save(ctx, info)
	return err
}

func (c *Boiler) SetMinTemp(ctx context.Context, temp float64) (*float64, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	info, err := c.GetInfo(ctx)
	if err != nil {
		return nil, err
	}
	if temp < info.MinTemp || temp > info.MaxTemp {
		return &info.MinTemp, fmt.Errorf("invalid min temperature")
	}
	info.MinTemp = temp
	err = c.save(ctx, info)
	return &info.MinTemp, err
}

func (c *Boiler) SetMaxTemp(ctx context.Context, temp float64) (*float64, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	info, err := c.GetInfo(ctx)
	if err != nil {
		return nil, err
	}
	if temp > info.MaxTemp || temp < info.MinTemp {
		return &info.MaxTemp, fmt.Errorf("invalid max temperature")
	}
	info.MaxTemp = temp
	err = c.save(ctx, info)
	return &info.MaxTemp, err
}

func (c *Boiler) SetRule(ctx context.Context, opt *Rule) (*Rule, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	info, err := c.GetInfo(ctx)
	if err != nil {
		return nil, err
	}
	if opt.TargetTemp < info.MinTemp || opt.TargetTemp > info.MaxTemp {
		return nil, fmt.Errorf("target temperature out of bounds")
	}

	// Map programmed intervals to a map for easier lookup
	lookupRules := make(map[string]*Rule)
	for _, interval := range info.Rules {
		lookupRules[interval.ID] = interval
	}

	// If ID is present in the opt, use that, otherwise generate a new one
	if opt.ID == "" {
		// Create ours if not present
		opt.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	lookupRules[opt.ID] = opt

	// Convert back to a slice
	rule := make([]*Rule, 0, len(lookupRules))
	for _, interval := range lookupRules {
		rule = append(rule, interval)
	}
	info.Rules = rule

	err = c.save(ctx, info)
	return opt, err
}

func (c *Boiler) StartRule(ctx context.Context, id string) (*Rule, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	info, err := c.GetInfo(ctx)
	if err != nil {
		return nil, err
	}

	alteredRule := &Rule{}
	for _, rule := range info.Rules {
		err = fmt.Errorf("could not find rule with id: %s", id)
		if rule.ID == id {
			rule.IsActive = true
			alteredRule = rule
			err = nil
			break
		}
	}
	if err != nil {
		return alteredRule, err
	}

	c.save(ctx, info)
	for _, rule := range info.Rules {
		if rule.ID == id {
			fmt.Printf("🔥 Started programmed interval %s\n", rule)
		}
	}
	return alteredRule, nil
}

func (c *Boiler) StopRule(ctx context.Context, id string) (*Rule, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	info, err := c.GetInfo(ctx)
	if err != nil {
		return nil, err
	}

	alteredInterval := &Rule{}
	for _, rule := range info.Rules {
		err = fmt.Errorf("could not find rule with id: %s", id)
		if rule.ID == id {
			stopTime := time.Now()
			rule.StoppedTime = &stopTime
			rule.IsActive = false
			alteredInterval = rule
			err = nil
			break
		}
	}
	if err != nil {
		return alteredInterval, err
	}
	c.save(ctx, info)
	fmt.Printf("💤 Stopped programmed interval %s\n", alteredInterval)
	return alteredInterval, nil
}

func (c *Boiler) DeleteRule(ctx context.Context, id string) error {
	fmt.Printf("Deleting rule %s\n", id)
	c.lock.Lock()
	defer c.lock.Unlock()
	info, err := c.GetInfo(ctx)
	if err != nil {
		return err
	}

	for index, rule := range info.Rules {
		err = fmt.Errorf("could not find programmed interval with id: %s", id)
		if rule.ID == id {
			info.Rules = append(info.Rules[:index], info.Rules[index+1:]...)
			err = nil
			break
		}
	}
	if err != nil {
		return err
	}

	err = c.save(ctx, info)
	return err
}

func (c *Boiler) ListenRules(ctx context.Context) (<-chan []*Rule, error) {
	ruleUpdates := make(chan []*Rule)
	boilerInfo, err := c.GetInfo(ctx)
	if err != nil {
		return nil, err
	}
	currentRules, err := json.Marshal(boilerInfo.Rules)
	if err != nil {
		return nil, err
	}
	boilerListener, err := c.Listen(ctx)
	if err != nil {
		return nil, err
	}
	go func() {
		defer close(ruleUpdates)
		for {
			select {
			case boilerInfo = <-boilerListener:
				newRules, err := json.Marshal(boilerInfo.Rules)
				if err != nil {
					fmt.Println(fmt.Errorf("error marshalling programmed intervals: %w", err))
					continue
				}
				if !cmp.Equal(currentRules, newRules) {
					select {
					case ruleUpdates <- boilerInfo.Rules:
					case <-ctx.Done():
						return
					}
				}
				currentRules = newRules
			case <-ctx.Done():
				return
			}
		}
	}()
	return ruleUpdates, nil
}

func (c *Boiler) ListenOverheating(ctx context.Context) (<-chan bool, error) {
	overheatingUpdates := make(chan bool)
	boilerInfo, err := c.GetInfo(ctx)
	if err != nil {
		return nil, err
	}
	currentFlag := boilerInfo.IsOverheatingProtectionActive
	boilerListener, err := c.Listen(ctx)
	if err != nil {
		return nil, err
	}
	go func() {
		defer close(overheatingUpdates)
		for {
			select {
			case boilerInfo = <-boilerListener:
				newFlag := boilerInfo.IsOverheatingProtectionActive
				if newFlag != currentFlag {
					select {
					case overheatingUpdates <- newFlag:
					case <-ctx.Done():
						return
					}
				}
				currentFlag = newFlag
			case <-ctx.Done():
				return
			}
		}
	}()
	return overheatingUpdates, nil
}

func (c *Boiler) Listen(ctx context.Context) (<-chan *BoilerInfo, error) {
	boilerUpdates := make(chan *BoilerInfo)
	go func() {
		defer close(boilerUpdates)
		sub := c.client.Subscribe(ctx, c.Config.Name)
		defer sub.Close()
		defer fmt.Println("👋 bye bye Mr American Pie...")

		if _, err := sub.Receive(ctx); err != nil {
			fmt.Printf("failed to receive from control PubSub: %s", err)
			return
		}

		redisChannel := sub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-redisChannel:
				boiler := BoilerInfo{}
				err := json.Unmarshal([]byte(msg.Payload), &boiler)
				if err != nil {
					fmt.Println(err)
					return
				} else {
					select {
					case boilerUpdates <- &boiler:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()
	return boilerUpdates, nil
}

func (c *Boiler) GetInfo(ctx context.Context) (*BoilerInfo, error) {
	data, err := c.client.Get(ctx, c.Config.Name).Result()
	switch err {
	case redis.Nil: // Data doesn't exist yet
		defaultInfo := &BoilerInfo{
			State:   StateUnknown,
			MinTemp: c.Config.DefaultMinTemperature,
			MaxTemp: c.Config.DefaultMaxTemperature,
			Rules:   nil,
		}
		data, err := json.Marshal(defaultInfo)
		if err != nil {
			return nil, err
		}
		err = c.client.Set(ctx, c.Config.Name, data, 0).Err()
		if err != nil {
			return nil, err
		}
		return defaultInfo, err
	case nil: // No error
		var info BoilerInfo
		err := json.Unmarshal([]byte(data), &info)
		return &info, err
	default:
		return nil, err
	}
}

func (c *Boiler) GetSwitchHistory(ctx context.Context, from time.Time, to time.Time) ([]*SwitchSample, error) {
	parseSwitchSample := func(sample redis.TSTimestampValue) SwitchSample {
		return SwitchSample{
			Time:  time.UnixMilli(sample.Timestamp),
			State: AllState[int(sample.Value)],
		}
	}
	defaultSwitchSample := SwitchSample{
		Time:  from,
		State: StateUnknown,
	}
	useDefault := true
	return readTimeSeries(ctx, c.client, c.switchSeriesKey, from, to, parseSwitchSample, useDefault, defaultSwitchSample)
}

func (c *Boiler) GetOverheatingProtectionHistory(ctx context.Context, from time.Time, to time.Time) ([]*OverheatingProtectionSample, error) {
	parseOverheatingSample := func(sample redis.TSTimestampValue) OverheatingProtectionSample {
		isActive := false
		if sample.Value == 1 {
			isActive = true
		}
		return OverheatingProtectionSample{
			Time:     time.UnixMilli(sample.Timestamp),
			IsActive: isActive,
		}
	}
	defaultOverheatingSample := OverheatingProtectionSample{
		Time:     from,
		IsActive: false,
	}
	useDefault := true
	return readTimeSeries(ctx, c.client, c.protectionSeriesKey, from, to, parseOverheatingSample, useDefault, defaultOverheatingSample)
}

func (c *Boiler) save(ctx context.Context, info *BoilerInfo) error {
	// Serialise data
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}

	// If there is diff set
	storedData, err := c.client.Get(ctx, c.Config.Name).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("cannot update database. Error when getting current state: %w", err)
	}
	diff := cmp.Diff([]byte(storedData), data)
	if diff != "" {
		// If we are about to save something different from what we have in the database then save
		err = c.client.Set(ctx, c.Config.Name, data, 0).Err()
		if err != nil {
			return err
		}
		// Add mapped switch sample
		stateIndex := GetStateIndex(info.State)
		timestampNow := int(time.Now().UnixMilli())
		err = c.client.TSAdd(ctx, c.switchSeriesKey, timestampNow, float64(stateIndex)).Err()
		if err != nil {
			return err
		}
		// Add overheating status sample
		overheating := 0.0
		if info.IsOverheatingProtectionActive {
			overheating = 1.0
		}
		err = c.client.TSAdd(ctx, c.protectionSeriesKey, timestampNow, overheating).Err()
		if err != nil {
			return err
		}

		// Schedule a new state message. This is delayed in case we make batch updates to the state
		go c.batchPublish(data)
	}
	return err
}

func (c *Boiler) batchPublish(data []byte) error {
	if c.stateUpdateCancel != nil {
		c.stateUpdateCancel()
	}
	cancelContext, cancelContextFunction := context.WithCancel(context.Background())
	c.stateUpdateCancel = cancelContextFunction
	select {
	case <-cancelContext.Done():
		return nil
	case <-time.After(time.Microsecond * stateUpdateBatchingTime):
		return c.client.Publish(cancelContext, c.Config.Name, data).Err()
	}
}

func readTimeSeries[T any](ctx context.Context, client *redis.Client, key string, from time.Time, to time.Time, parse func(v redis.TSTimestampValue) T, useInitSample bool, defaultInitSample T) ([]*T, error) {
	fromTimestamp := int(from.UnixMilli())
	toTimestamp := int(to.UnixMilli())
	data, err := client.TSRange(ctx, key, fromTimestamp, toTimestamp).Result()
	if err != nil {
		return nil, err
	}

	samplesCount := len(data)
	extraCount := 0
	var initSample T
	if useInitSample {
		// Establish initSample
		extraCount = 1
		optLimitToLast := &redis.TSRevRangeOptions{Count: 1}
		backData, err := client.TSRevRangeWithArgs(ctx, key, 0, fromTimestamp, optLimitToLast).Result()
		if err != nil {
			return nil, err
		}
		switch len(backData) {
		case 1:
			initSample = parse(backData[0])
		case 0:
			initSample = defaultInitSample
		default:
			return nil, fmt.Errorf("Unexpected redis behaviour, wanted 0 or 1 element got %d", len(backData))
		}
	}

	samples := make([]*T, samplesCount+extraCount)
	if useInitSample {
		samples[0] = &initSample
	}
	for i, redisSample := range data {
		sample := parse(redisSample)
		samples[i+extraCount] = &sample
	}

	return samples, nil
}
