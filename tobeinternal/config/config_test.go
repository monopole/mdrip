package config

import (
	"testing"
)

func TestGetConfig(t *testing.T) {
	_, err := GetConfig()
	if err == nil {
		t.Errorf("expected complaints about no data source args")
		return
	}
	if err.Error() != "must specify a data source - files, directory, or github clone url" {
		t.Errorf("expected data source complaint, not: " + err.Error())
	}
}
