package model

import (
	"fmt"
)

var Caldaia Switch = Switch{State: StateUnknown}

// Function to switch the relay on or off
// Accepts only two values: "on" or "off"
func (externalSwitch Switch) Set(targetState State) (State, error) {
	switch targetState {
	case StateOn:
		fmt.Println("Switching relay on")
	case StateOff:
		fmt.Println("Switching relay off")
	default:
		return targetState, fmt.Errorf("Invalid state to set")
	}
	externalSwitch.State = targetState
	return externalSwitch.State, nil;
}

// Function to get the current state of the relay
func (externalSwitch Switch) Get() (State, error) {
	return externalSwitch.State, nil
}