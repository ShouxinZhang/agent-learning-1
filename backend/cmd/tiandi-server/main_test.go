package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"agent-dou-dizhu/internal/tiandi/agent"
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

	newTestMux(service).ServeHTTP(rec, req)

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

	newTestMux(service).ServeHTTP(rec, req)

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

	newTestMux(service).ServeHTTP(rec, req)

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

	newTestMux(service).ServeHTTP(rec, req)

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

func TestHandleAgentPromptReturnsPromptPayload(t *testing.T) {
	service, err := demo.NewServiceWithOptions(demo.ServiceOptions{Mode: fsm.ModeFixedP0PlayTest})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/game/agent/prompt", nil)
	rec := httptest.NewRecorder()

	newTestMux(service).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d want %d", rec.Code, http.StatusOK)
	}

	var payload agent.PromptResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode prompt payload: %v", err)
	}

	if payload.CurrentActor != "P0" {
		t.Fatalf("currentActor: got %q want %q", payload.CurrentActor, "P0")
	}
	if payload.PlayerSeat != "P0" {
		t.Fatalf("playerSeat: got %q want %q", payload.PlayerSeat, "P0")
	}
	if len(payload.PlayerHand) != 20 {
		t.Fatalf("expected 20 cards in prompt hand, got %d", len(payload.PlayerHand))
	}
	if !strings.Contains(payload.SystemPrompt, "JSON") {
		t.Fatalf("expected system prompt to mention JSON contract, got %q", payload.SystemPrompt)
	}
	if len(payload.State.Players) != 3 {
		t.Fatalf("expected embedded state in prompt payload, got %d players", len(payload.State.Players))
	}
}

func TestHandleAgentRunMockStoresTraceAndAdvancesState(t *testing.T) {
	service, err := demo.NewServiceWithOptions(demo.ServiceOptions{Mode: fsm.ModeFixedP0PlayTest})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	mux := newTestMux(service)
	runReq := httptest.NewRequest(http.MethodPost, "/api/game/agent/run", bytes.NewBufferString(`{"mode":"mock"}`))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, runReq)

	if rec.Code != http.StatusOK {
		t.Fatalf("run status: got %d want %d", rec.Code, http.StatusOK)
	}

	var payload agent.RunResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode run payload: %v", err)
	}

	if payload.Trace.Mode != agent.RunModeMock {
		t.Fatalf("trace mode: got %q want %q", payload.Trace.Mode, agent.RunModeMock)
	}
	if !payload.Trace.Applied {
		t.Fatalf("expected trace to be applied, got %#v", payload.Trace)
	}
	if payload.State.CurrentActor != "P1" {
		t.Fatalf("expected mock run to advance actor to P1, got %q", payload.State.CurrentActor)
	}
	if payload.State.CurrentTrick == nil {
		t.Fatal("expected current trick after mock play")
	}

	traceReq := httptest.NewRequest(http.MethodGet, "/api/game/agent/trace", nil)
	traceRec := httptest.NewRecorder()
	mux.ServeHTTP(traceRec, traceReq)

	if traceRec.Code != http.StatusOK {
		t.Fatalf("trace status: got %d want %d", traceRec.Code, http.StatusOK)
	}

	var tracePayload agent.TraceEnvelope
	if err := json.NewDecoder(traceRec.Body).Decode(&tracePayload); err != nil {
		t.Fatalf("decode trace payload: %v", err)
	}
	if tracePayload.Trace == nil {
		t.Fatal("expected stored trace")
	}
	if tracePayload.Trace.Decision.Kind != "play" {
		t.Fatalf("expected play trace, got %#v", tracePayload.Trace.Decision)
	}
}

func TestHandleAgentRunRejectsBadJSON(t *testing.T) {
	service, err := demo.NewServiceWithOptions(demo.ServiceOptions{Mode: fsm.ModeFixedP0PlayTest})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/game/agent/run", bytes.NewBufferString(`{`))
	rec := httptest.NewRecorder()

	newTestMux(service).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandleAgentRunRejectsUnsupportedMode(t *testing.T) {
	service, err := demo.NewServiceWithOptions(demo.ServiceOptions{Mode: fsm.ModeFixedP0PlayTest})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/game/agent/run", bytes.NewBufferString(`{"mode":"unknown"}`))
	rec := httptest.NewRecorder()

	newTestMux(service).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandleAgentMatchRunAndStateEndpoints(t *testing.T) {
	service, err := demo.NewServiceWithOptions(demo.ServiceOptions{Mode: fsm.ModeFixedP0PlayTest})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	mux := newTestMux(service)

	runReq := httptest.NewRequest(http.MethodPost, "/api/game/agent/match/run", bytes.NewBufferString(`{"mode":"mock","maxSteps":100,"resetGame":false}`))
	runRec := httptest.NewRecorder()
	mux.ServeHTTP(runRec, runReq)

	if runRec.Code != http.StatusOK {
		t.Fatalf("run status: got %d want %d", runRec.Code, http.StatusOK)
	}

	var runPayload agent.MatchRunResponse
	if err := json.NewDecoder(runRec.Body).Decode(&runPayload); err != nil {
		t.Fatalf("decode run payload: %v", err)
	}
	if runPayload.Match.Status != "completed" || runPayload.Match.Winner != "P0" {
		t.Fatalf("unexpected match payload: %#v", runPayload.Match)
	}

	stateReq := httptest.NewRequest(http.MethodGet, "/api/game/agent/match/state", nil)
	stateRec := httptest.NewRecorder()
	mux.ServeHTTP(stateRec, stateReq)

	if stateRec.Code != http.StatusOK {
		t.Fatalf("state status: got %d want %d", stateRec.Code, http.StatusOK)
	}

	var statePayload agent.MatchStateResponse
	if err := json.NewDecoder(stateRec.Body).Decode(&statePayload); err != nil {
		t.Fatalf("decode match state payload: %v", err)
	}
	if statePayload.Match == nil || statePayload.Match.Winner != "P0" {
		t.Fatalf("expected stored match in state payload, got %#v", statePayload)
	}
	if statePayload.State.Winner != "P0" {
		t.Fatalf("expected state payload to expose winner, got %#v", statePayload.State)
	}

	traceReq := httptest.NewRequest(http.MethodGet, "/api/game/agent/match/trace", nil)
	traceRec := httptest.NewRecorder()
	mux.ServeHTTP(traceRec, traceReq)

	if traceRec.Code != http.StatusOK {
		t.Fatalf("trace status: got %d want %d", traceRec.Code, http.StatusOK)
	}

	var tracePayload agent.MatchEnvelope
	if err := json.NewDecoder(traceRec.Body).Decode(&tracePayload); err != nil {
		t.Fatalf("decode trace payload: %v", err)
	}
	if tracePayload.Match == nil || tracePayload.Match.StepCount == 0 {
		t.Fatalf("expected stored match trace, got %#v", tracePayload)
	}
}

func TestHandleAgentMatchResetClearsStoredTrace(t *testing.T) {
	service, err := demo.NewServiceWithOptions(demo.ServiceOptions{Mode: fsm.ModeFixedP0PlayTest})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	mux := newTestMux(service)
	runReq := httptest.NewRequest(http.MethodPost, "/api/game/agent/match/run", bytes.NewBufferString(`{"mode":"mock","maxSteps":10,"resetGame":false}`))
	runRec := httptest.NewRecorder()
	mux.ServeHTTP(runRec, runReq)

	resetReq := httptest.NewRequest(http.MethodPost, "/api/game/agent/match/reset", nil)
	resetRec := httptest.NewRecorder()
	mux.ServeHTTP(resetRec, resetReq)

	if resetRec.Code != http.StatusOK {
		t.Fatalf("reset status: got %d want %d", resetRec.Code, http.StatusOK)
	}

	traceReq := httptest.NewRequest(http.MethodGet, "/api/game/agent/match/trace", nil)
	traceRec := httptest.NewRecorder()
	mux.ServeHTTP(traceRec, traceReq)

	var tracePayload agent.MatchEnvelope
	if err := json.NewDecoder(traceRec.Body).Decode(&tracePayload); err != nil {
		t.Fatalf("decode trace payload: %v", err)
	}
	if tracePayload.Match != nil {
		t.Fatalf("expected trace to be cleared after reset, got %#v", tracePayload.Match)
	}
}

func TestServerAddrFromEnv(t *testing.T) {
	t.Run("defaults to 8080", func(t *testing.T) {
		t.Setenv("TIANDI_SERVER_ADDR", "")
		t.Setenv("PORT", "")
		if got := serverAddrFromEnv(); got != ":8080" {
			t.Fatalf("addr: got %q want %q", got, ":8080")
		}
	})

	t.Run("prefers explicit addr", func(t *testing.T) {
		t.Setenv("TIANDI_SERVER_ADDR", ":18080")
		t.Setenv("PORT", "19090")
		if got := serverAddrFromEnv(); got != ":18080" {
			t.Fatalf("addr: got %q want %q", got, ":18080")
		}
	})

	t.Run("uses port env", func(t *testing.T) {
		t.Setenv("TIANDI_SERVER_ADDR", "")
		t.Setenv("PORT", "19090")
		if got := serverAddrFromEnv(); got != ":19090" {
			t.Fatalf("addr: got %q want %q", got, ":19090")
		}
	})
}

func newTestMux(service *demo.Service) *http.ServeMux {
	return newMux(service, agent.NewService(service, agent.ServiceOptions{}))
}
