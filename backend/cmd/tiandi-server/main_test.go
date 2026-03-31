package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"agent-dou-dizhu/internal/tiandi/demo"
	"agent-dou-dizhu/internal/tiandi/fsm"
	"agent-dou-dizhu/internal/tiandi/rules"
)

func TestHandleRulesReturnsCatalog(t *testing.T) {
	service, err := demo.NewService()
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/game/rules", nil)
	rec := httptest.NewRecorder()

	newMux(service).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d want %d", rec.Code, http.StatusOK)
	}

	var payload rules.Catalog
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode rules payload: %v", err)
	}

	if payload.SequenceHigh != "A" {
		t.Fatalf("sequence high: got %q want %q", payload.SequenceHigh, "A")
	}
	if len(payload.Sections) == 0 {
		t.Fatal("expected rules sections")
	}
	if len(payload.BombPriority) == 0 {
		t.Fatal("expected bomb priority entries")
	}
}

func TestHandleRulesRejectsPost(t *testing.T) {
	service, err := demo.NewService()
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/game/rules", nil)
	rec := httptest.NewRecorder()

	newMux(service).ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status: got %d want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandleStateReturnsTestModePayload(t *testing.T) {
	service, err := demo.NewServiceWithOptions(demo.ServiceOptions{Mode: fsm.ModeFixedP0PlayTest})
	if err != nil {
		t.Fatalf("new service with options: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/game/state", nil)
	rec := httptest.NewRecorder()

	newMux(service).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d want %d", rec.Code, http.StatusOK)
	}

	var payload demo.StateResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode state payload: %v", err)
	}

	if payload.Phase != string(fsm.PhasePlay) {
		t.Fatalf("expected phase PLAY, got %s", payload.Phase)
	}
	if payload.Landlord != "P0" {
		t.Fatalf("expected landlord P0, got %s", payload.Landlord)
	}
	if payload.TestMode == nil || !payload.TestMode.Enabled {
		t.Fatal("expected enabled testMode payload")
	}
	if payload.TestMode.FixedLandlord != "P0" {
		t.Fatalf("expected fixed landlord P0, got %s", payload.TestMode.FixedLandlord)
	}
	if !payload.Bottom.Visible {
		t.Fatal("expected visible bottom cards")
	}
	if payload.CurrentActor != "P0" {
		t.Fatalf("expected current actor P0, got %q", payload.CurrentActor)
	}
	if len(payload.AvailableActions) != 1 || payload.AvailableActions[0] != "play" {
		t.Fatalf("expected available actions [play], got %#v", payload.AvailableActions)
	}
}

func TestHandleStateReturnsTestModeMetadata(t *testing.T) {
	service, err := demo.NewServiceWithOptions(demo.ServiceOptions{Mode: "fixed_p0_play_test"})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/game/state", nil)
	rec := httptest.NewRecorder()

	newMux(service).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d want %d", rec.Code, http.StatusOK)
	}

	var payload demo.StateResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode state payload: %v", err)
	}

	if payload.Phase != "PLAY" {
		t.Fatalf("phase: got %q want %q", payload.Phase, "PLAY")
	}
	if payload.Landlord != "P0" {
		t.Fatalf("landlord: got %q want %q", payload.Landlord, "P0")
	}
	if payload.TestMode == nil {
		t.Fatal("expected testMode metadata")
	}
	if !payload.TestMode.Enabled {
		t.Fatal("expected testMode enabled")
	}
	if payload.TestMode.FixedLandlord != "P0" {
		t.Fatalf("fixedLandlord: got %q want %q", payload.TestMode.FixedLandlord, "P0")
	}
	if !payload.Bottom.Visible {
		t.Fatal("expected bottom to be visible")
	}
	if payload.CurrentActor != "P0" {
		t.Fatalf("currentActor: got %q want %q", payload.CurrentActor, "P0")
	}
}
