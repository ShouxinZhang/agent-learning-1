package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"agent-dou-dizhu/internal/tiandi/demo"
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
