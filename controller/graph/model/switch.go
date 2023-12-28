package model

import (
	"fmt"
)

var caldaia Boiler = Boiler{State: StateUnknown}

// Function to switch the relay on or off
// Accepts only two values: "on" or "off"
func (c Boiler) Set(targetState State) (State, error) {
	switch targetState {
	case StateOn:
		fmt.Println("Switching relay on")
	case StateOff:
		fmt.Println("Switching relay off")
	default:
		return targetState, fmt.Errorf("Invalid state to set")
	}
	c.State = targetState
	return c.State, nil
}

// Function to get the current state of the relay
func (c Boiler) Get() (State, error) {
	return c.State, nil
}
