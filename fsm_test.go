package fsm

import (
	"reflect"
	"testing"
)

func TestFSM(t *testing.T) {

	sm := New()
	sm.AddTransition("idle", "PLAY", "playing")
	sm.AddTransition("playing", "STOP", "idle")

	sm.SetOnEntry("idle", func() { t.Logf("entered idle") })
	sm.SetOnExit("idle", func() { t.Logf("left idle") })

	sm.SetOnEntry("playing", func() { t.Logf("entered playing") })
	sm.SetOnExit("playing", func() { t.Logf("left playing") })

	states := sm.States()
	wantedStates := States{
		"idle",
		"playing",
	}
	if !reflect.DeepEqual(wantedStates, states) {
		t.Errorf("want: %s, got: %s", wantedStates, states)
	}

	if err := sm.SetState("foo"); err == nil {
		t.Errorf("expected an error")
	}

	sm.SetState("idle")
	if sm.State() != "idle" {
		t.Errorf("want: %s, got: %s", "idle", sm.State())
	}
	validInputs := sm.ValidEvents(sm.State())
	wantedInputs := []Event{"PLAY"}
	if !reflect.DeepEqual(wantedInputs, validInputs) {
		t.Errorf("want: %v, got: %v", wantedInputs, validInputs)
	}

	err := sm.Process("PUSH")
	if err == nil {
		t.Errorf("expected an error")
	}
	if sm.State() != "idle" {
		t.Errorf("want: %s, got: %s", "idle", sm.State())
	}

	err = sm.Process("PLAY")
	if err != nil {
		t.Errorf("expected no error: %v", err)
	}
	if sm.State() != "playing" {
		t.Errorf("want: %s, got: %s", "playing", sm.State())
	}
	validInputs = sm.ValidEvents(sm.State())
	wantedInputs = []Event{"STOP"}
	if !reflect.DeepEqual(wantedInputs, validInputs) {
		t.Errorf("want: %v, got: %v", wantedInputs, validInputs)
	}

	err = sm.Process("STOP")
	if err != nil {
		t.Errorf("expected no error: %v", err)
	}
	if sm.State() != "idle" {
		t.Errorf("want: %s, got: %s", "idle", sm.State())
	}
	validInputs = sm.ValidEvents(sm.State())
	wantedInputs = []Event{"PLAY"}
	if !reflect.DeepEqual(wantedInputs, validInputs) {
		t.Errorf("want: %v, got: %v", wantedInputs, validInputs)
	}

}

func TestData(t *testing.T) {
	sm := New()

	sm.Data("a")["foo"] = "bar"

	if v := sm.Data("a").String("foo"); v != "bar" {
		t.Errorf("want: %s, got: %s", "bar", v)
	}
}
