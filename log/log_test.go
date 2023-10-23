package log

import (
	"testing"
)

func TestSetupLogger(t *testing.T) {
	err := SetupLogger(false)
	if err != nil {
		t.Fatal(err)
	}
	Logger.Debug("hello world")
}
