package log

import (
	"testing"
)

func TestSetupLogger(t *testing.T) {
	err := SetupLogger()
	if err != nil {
		t.Fatal(err)
	}
	Logger.Info("hello world")
}
