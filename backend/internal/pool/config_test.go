package pool

import "testing"

func TestConfigValidate(t *testing.T) {
	ok := DefaultConfig()
	if err := ok.Validate(); err != nil {
		t.Fatal(err)
	}
	bad := ok
	bad.MaxActive = 0
	if err := bad.Validate(); err == nil {
		t.Fatal("expected max active error")
	}
	bad = ok
	bad.MaxIdle = ok.MaxActive + 1
	if err := bad.Validate(); err == nil {
		t.Fatal("expected max idle error")
	}
	bad = ok
	bad.HealthInterval = 0
	if err := bad.Validate(); err == nil {
		t.Fatal("expected duration error")
	}
}

func TestStateTransitions(t *testing.T) {
	if !AllowedTransition(StateIdle, StateProbing) {
		t.Fatal("idle->probing")
	}
	if AllowedTransition(StateInUse, StateIdle) == false {
		t.Fatal("inuse->idle")
	}
	if AllowedTransition(StateClosed, StateIdle) {
		t.Fatal("closed cannot revive")
	}
	if !StateIdle.Borrowable() || StateProbing.Borrowable() {
		t.Fatal("borrowable flags")
	}
}
