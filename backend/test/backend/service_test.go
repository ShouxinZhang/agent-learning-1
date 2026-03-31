package backend_test

import (
	"testing"

	"agent-dou-dizhu/internal/tiandi/demo"
	"agent-dou-dizhu/internal/tiandi/fsm"
)

func TestServiceReturnsTestModeState(t *testing.T) {
	service, err := demo.NewServiceWithOptions(demo.ServiceOptions{Mode: fsm.ModeFixedP0PlayTest})
	if err != nil {
		t.Fatalf("new service with options: %v", err)
	}

	state, err := service.State()
	if err != nil {
		t.Fatalf("state failed: %v", err)
	}

	if state.Phase != string(fsm.PhasePlay) {
		t.Fatalf("expected PLAY phase, got %s", state.Phase)
	}
	if state.Landlord != "P0" {
		t.Fatalf("expected landlord P0, got %s", state.Landlord)
	}
	if state.TestMode == nil {
		t.Fatal("expected testMode payload")
	}
	if !state.TestMode.Enabled {
		t.Fatal("expected testMode to be enabled")
	}
	if state.TestMode.FixedLandlord != "P0" {
		t.Fatalf("expected fixed landlord P0, got %s", state.TestMode.FixedLandlord)
	}
	if !state.TestMode.DirectPlay {
		t.Fatal("expected directPlay to be true")
	}
	if state.CurrentActor != "" {
		t.Fatalf("expected no current actor, got %q", state.CurrentActor)
	}
	if len(state.AvailableActions) != 0 {
		t.Fatalf("expected no available actions, got %d", len(state.AvailableActions))
	}
	if !state.Bottom.Visible {
		t.Fatal("expected bottom cards to be visible")
	}
	if len(state.Players) != 3 {
		t.Fatalf("expected 3 players, got %d", len(state.Players))
	}
	if !state.Players[0].IsLandlord {
		t.Fatal("expected P0 player to be landlord")
	}
	if len(state.Players[0].Cards) != 20 {
		t.Fatalf("expected P0 to have 20 cards, got %d", len(state.Players[0].Cards))
	}
}

func TestDefaultServiceDoesNotExposeTestMode(t *testing.T) {
	service, err := demo.NewService()
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	state, err := service.State()
	if err != nil {
		t.Fatalf("state failed: %v", err)
	}

	if state.TestMode != nil {
		t.Fatal("expected normal service state to omit testMode")
	}
}
