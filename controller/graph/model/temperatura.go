package model

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	PrimaryRetentionTime = 7 * 24 * 60 * 60 * 1000 // 7 days
	CompactTime          = 5 * 60 * 1000           // 5 minutes
)

type SensorOptions struct {
	Name     string
	Position string
}

type Sensor struct {
	Name         string
	Position     string
	Client       *redis.Client
	Id           string
	compactedKey string
}

func NewSensor(ctx context.Context, client *redis.Client, opt *SensorOptions) (*Sensor, error) {
	key := opt.Name + ":" + opt.Position
	compactedKey := opt.Name + "_compacted" + ":" + opt.Position
	sensor := Sensor{opt.Name, opt.Position, client, key, compactedKey}

	// Check if sensor already exists
	exists, _ := sensor.Client.Exists(ctx, key).Result()
	if exists == 0 {
		// If not, create it
		_, err := sensor.Client.TSCreateWithArgs(ctx, key, &redis.TSOptions{
			Retention: PrimaryRetentionTime,
			Labels:    map[string]string{"position": sensor.Position},
		}).Result()
		if err != nil {
			return &sensor, err
		}
	}
	exists, _ = sensor.Client.Exists(ctx, compactedKey).Result()
	if exists == 0 {
		// Create compaction key
		_, err := sensor.Client.TSCreateWithArgs(ctx, compactedKey, &redis.TSOptions{Retention: 0}).Result()
		if err != nil {
			return &sensor, err
		}

		// Create compaction rule
		_, err = sensor.Client.TSCreateRule(ctx, key, compactedKey, redis.Avg, CompactTime).Result()
		if err != nil {
			return &sensor, err
		}
	}
	return &sensor, nil
}

func (s *Sensor) Get(ctx context.Context, from *time.Time, to *time.Time) ([]*Measure, error) {
	// Get data from Redis
	var fromTimestamp, toTimestamp int
	if to == nil {
		toTimestamp = int(time.Now().UnixMilli())
	} else {
		toTimestamp = int(to.UnixMilli())
	}
	if from == nil {
		fromTimestamp = toTimestamp - 24*60*60*1000
	} else {
		fromTimestamp = int(from.UnixMilli())
	}
	data, err := s.Client.TSRange(ctx, s.compactedKey, fromTimestamp, toTimestamp).Result()
	if err != nil {
		return nil, err
	}

	// Parse data
	measures := make([]*Measure, len(data))
	for index, sample := range data {
		measures[index] = &Measure{sample.Value, time.UnixMilli(int64(sample.Timestamp))}
	}
	return measures, nil
}

func (s *Sensor) Listen(ctx context.Context) (<-chan *Measure, error) {
	temperatureUpdates := make(chan *Measure)
	go func() {
		sub := s.Client.Subscribe(ctx, s.Id)
		for msg := range sub.Channel() {
			measure := Measure{}
			err := json.Unmarshal([]byte(msg.Payload), &measure)
			if err != nil {
				fmt.Println(err)
				break
			} else {
				temperatureUpdates <- &measure
			}
		}
		close(temperatureUpdates)
	}()
	return temperatureUpdates, nil
}

func (s *Sensor) Add(ctx context.Context, measure *Measure) error {
	_, err := s.Client.TSAdd(ctx, s.Id, int(measure.Timestamp.UnixMilli()), measure.Value).Result()
	if err != nil {
		return err
	}
	return nil
}

func (s *Sensor) Sample(ctx context.Context) (*Measure, error) {
	// Create a random number for testing
	value := rand.Float64() * 100
	sample := Measure{Timestamp: time.Now(), Value: value}

	// Publish the sample
	message, err := json.Marshal(sample)
	if err != nil {
		return &sample, err
	}
	err = s.Client.Publish(ctx, s.Id, message).Err()
	return &sample, err
}
