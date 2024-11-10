package model

import (
	"context"
	"encoding/json"
	"fmt"
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

// Get returns the measures of the sensor in the given time interval.
// If from is nil, it will be set to 24 hours before to.
// If to is nil, it will be set to the current time.
func (s *Sensor) Get(ctx context.Context, from time.Time, to time.Time) ([]*Measure, error) {
	// Get data from Redis
	fromTimestamp := int(from.UnixMilli())
	toTimestamp := int(to.UnixMilli())
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
	fmt.Println("Listening for updates on sensor", s.Id)
	temperatureUpdates := make(chan *Measure)
	go func() {
		defer close(temperatureUpdates)
		sub := s.Client.Subscribe(ctx, s.Id)
		defer sub.Close()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-sub.Channel():
				if !ok {
					return
				}

				// Attempt to unmarshal the payload into a Measure
				measure := &Measure{}
				err := json.Unmarshal([]byte(msg.Payload), measure)
				if err != nil {
					fmt.Println("Error unmarshalling payload:", err)
					continue
				}

				// Send the measure to the updates channel
				select {
				case temperatureUpdates <- measure:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return temperatureUpdates, nil
}

func (s *Sensor) AddSample(ctx context.Context, sample *Measure) error {
	// Add sample to Redis
	_, err := s.Client.TSAdd(ctx, s.Id, int(sample.Timestamp.UnixMilli()), sample.Value).Result()
	if err != nil {
		return err
	}

	// Publish measure
	message, err := json.Marshal(sample)
	if err != nil {
		return err
	}
	return s.Client.Publish(ctx, s.Id, message).Err()
}

func (s *Sensor) GetAverage(ctx context.Context, from time.Time, to time.Time) (*float64, error) {
	measureRange, err := s.Get(ctx, from, to)
	if err != nil {
		return nil, err
	}
	if len(measureRange) == 0 {
		return nil, nil
	}
	sum := 0.0
	for _, measure := range measureRange {
		sum += measure.Value
	}
	average := sum / float64(len(measureRange))
	return &average, nil
}
