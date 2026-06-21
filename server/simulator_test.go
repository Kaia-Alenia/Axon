package main

import (
	"testing"
)

func TestNewSimulator(t *testing.T) {
	sim := NewSimulator()
	if sim == nil {
		t.Error("NewSimulator() returned nil, expected a non-nil InputSimulator")
	} else {
		// Clean up the simulator if it implements Close
		sim.Close()
	}
}
