package main

import (
	"reflect"
	"testing"
)

func TestConfigurationStructHasNoFieldClarifyTask(t *testing.T) {
	// Get the type of Configuration struct
	configType := reflect.TypeOf(Configuration{})

	// Check if the ClarifyTask field exists
	_, exists := configType.FieldByName("ClarifyTask")
	if exists {
		t.Error("ClarifyTask field still exists in Configuration struct but should have been removed")
	}
}
