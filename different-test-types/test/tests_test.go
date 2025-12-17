package test

import (
	"testing"
)

func TestSuccess(t *testing.T) {
	t.Log("Test succeeded")
}

func TestLogFailure(t *testing.T) {
	t.Errorf("Test failed")
	t.Log("Reachable code")
}

func TestFails(t *testing.T) {
	t.Fail()
	t.Log("Failed but continued work")
	t.FailNow()
	t.Log("Unreachable code")
}

func TestIgnored(t *testing.T) {
	t.Skip("Skipped by some cause")
	t.Log("Unreachable code")
}

// TestIgnoredV2 works identically to TestIgnored
func TestIgnoredV2(t *testing.T) {
	t.Log("Skipped V2 by some cause")
	t.SkipNow()
	t.Log("Unreachable code")
}
