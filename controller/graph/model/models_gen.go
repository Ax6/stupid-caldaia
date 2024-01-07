// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"fmt"
	"io"
	"strconv"
	"time"
)

type BoilerInfo struct {
	State               State                 `json:"state"`
	MinTemp             float64               `json:"minTemp"`
	MaxTemp             float64               `json:"maxTemp"`
	ProgrammedIntervals []*ProgrammedInterval `json:"programmedIntervals"`
}

type Measure struct {
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

type ProgrammedInterval struct {
	ID          string        `json:"id"`
	Start       time.Time     `json:"start"`
	Duration    time.Duration `json:"duration"`
	TargetTemp  float64       `json:"targetTemp"`
	RepeatDays  []int         `json:"repeatDays"`
	IsActive    bool          `json:"isActive"`
	StoppedTime time.Time     `json:"stoppedTime"`
}

type State string

const (
	StateOn      State = "ON"
	StateOff     State = "OFF"
	StateUnknown State = "UNKNOWN"
)

var AllState = []State{
	StateOn,
	StateOff,
	StateUnknown,
}

func (e State) IsValid() bool {
	switch e {
	case StateOn, StateOff, StateUnknown:
		return true
	}
	return false
}

func (e State) String() string {
	return string(e)
}

func (e *State) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = State(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid State", str)
	}
	return nil
}

func (e State) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
