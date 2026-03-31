package backend_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"agent-dou-dizhu/internal/tiandi/agent"
	"agent-dou-dizhu/internal/tiandi/demo"
	"agent-dou-dizhu/internal/tiandi/fsm"
)

func TestBuildPromptIncludesCoreStateAndContract(t *testing.T) {
	service, err := demo.NewServiceWithOptions(demo.ServiceOptions{Mode: fsm.ModeFixedP0PlayTest})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	state, err := service.State()
	if err != nil {
		t.Fatalf("state: %v", err)
	}

	prompt := agent.BuildPrompt(state, service.Rules())

	if prompt.CurrentActor != "P0" {
		t.Fatalf("current actor: got %q want %q", prompt.CurrentActor, "P0")
	}
	if len(prompt.AvailableActions) != 1 || prompt.AvailableActions[0] != "play" {
		t.Fatalf("available actions: got %#v", prompt.AvailableActions)
	}
	if !strings.Contains(prompt.SystemPrompt, "只输出一个 JSON 对象") {
		t.Fatalf("system prompt missing json constraint: %q", prompt.SystemPrompt)
	}
	if !strings.Contains(prompt.UserPrompt, "currentActor=P0") {
		t.Fatalf("user prompt missing actor: %q", prompt.UserPrompt)
	}
	if !strings.Contains(prompt.UserPrompt, `"play"`) {
		t.Fatalf("user prompt missing available action: %q", prompt.UserPrompt)
	}
	if len(prompt.PlayerHand) != 20 {
		t.Fatalf("expected 20 hand cards in test mode, got %d", len(prompt.PlayerHand))
	}
	if prompt.PlayerSeat != "P0" {
		t.Fatalf("player seat: got %q want %q", prompt.PlayerSeat, "P0")
	}
	if prompt.State.CurrentActor != "P0" {
		t.Fatalf("embedded state current actor: got %q want %q", prompt.State.CurrentActor, "P0")
	}
	if _, ok := prompt.ActionSchema.Properties["resolutionId"]; !ok {
		t.Fatalf("expected resolutionId in action schema, got %#v", prompt.ActionSchema.Properties)
	}
}

func TestBuildPromptWithContextIncludesCardCounterAndRoundMemory(t *testing.T) {
	service, err := demo.NewServiceWithOptions(demo.ServiceOptions{Mode: fsm.ModeFixedP0PlayTest})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	state, err := service.State()
	if err != nil {
		t.Fatalf("state: %v", err)
	}

	prompt := agent.BuildPromptWithContext(state, service.Rules(), agent.PromptContext{
		Seat:     "P0",
		SeatRole: "landlord",
		CardCounter: agent.CardCounter{
			Seat:                 "P0",
			SeatRole:             "landlord",
			PlayedCardsBySeat:    map[string][]string{"P1": []string{"S-4"}},
			PlayedRankCounts:     map[string]int{"4": 1},
			RemainingUnknown:     33,
			BlackJokerPlayed:     false,
			RedJokerPlayed:       false,
			BombSignals:          []string{"P1:四张炸弹/4"},
			TotalPlayedCardCount: 1,
		},
		RoundMemory: agent.SeatRoundMemory{
			Seat:       "P0",
			SeatRole:   "landlord",
			RoundIndex: 2,
			TrickIndex: 1,
			LastOpponentPlay: &agent.PlayedAction{
				Seat:          "P1",
				Kind:          "play",
				Cards:         []string{"S-4"},
				ResolvedLabel: "单张",
				MainRank:      "4",
				Relationship:  "opponent",
			},
		},
	})

	if prompt.PlayerRole != "landlord" {
		t.Fatalf("player role: got %q want landlord", prompt.PlayerRole)
	}
	if !strings.Contains(prompt.UserPrompt, `cardCounter={"seat":"P0"`) {
		t.Fatalf("user prompt missing card counter: %q", prompt.UserPrompt)
	}
	if !strings.Contains(prompt.UserPrompt, `"lastOpponentPlay":{"seat":"P1"`) {
		t.Fatalf("user prompt missing round memory payload: %q", prompt.UserPrompt)
	}
	if !strings.Contains(prompt.SystemPrompt, "记牌器摘要") {
		t.Fatalf("system prompt missing memory guidance: %q", prompt.SystemPrompt)
	}
}

func TestParseDecisionTextAcceptsMarkdownFence(t *testing.T) {
	raw := "```json\n{\"seat\":\"P0\",\"kind\":\"play\",\"cards\":[\"heart-A\"],\"reason\":\"test\"}\n```"

	decision, err := agent.ParseDecisionText(raw)
	if err != nil {
		t.Fatalf("parse decision: %v", err)
	}

	if decision.Seat != "P0" || decision.Kind != "play" {
		t.Fatalf("unexpected decision: %#v", decision)
	}
	if len(decision.Cards) != 1 || decision.Cards[0] != "heart-A" {
		t.Fatalf("unexpected cards: %#v", decision.Cards)
	}
}

func TestValidateDecisionRejectsCardsOutsideCurrentHand(t *testing.T) {
	service, err := demo.NewServiceWithOptions(demo.ServiceOptions{Mode: fsm.ModeFixedP0PlayTest})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	state, err := service.State()
	if err != nil {
		t.Fatalf("state: %v", err)
	}

	err = agent.ValidateDecision(state, agent.Decision{
		Seat:   "P0",
		Kind:   "play",
		Cards:  []string{"not-a-real-card"},
		Reason: "invalid",
	})
	if err == nil {
		t.Fatal("expected invalid card to be rejected")
	}
}

func TestChooseMockDecisionPicksCurrentPlayersFirstCard(t *testing.T) {
	service, err := demo.NewServiceWithOptions(demo.ServiceOptions{Mode: fsm.ModeFixedP0PlayTest})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	state, err := service.State()
	if err != nil {
		t.Fatalf("state: %v", err)
	}

	decision, raw, err := agent.ChooseMockDecision(state)
	if err != nil {
		t.Fatalf("choose mock decision: %v", err)
	}

	if decision.Seat != "P0" || decision.Kind != "play" {
		t.Fatalf("unexpected decision: %#v", decision)
	}
	if len(decision.Cards) != 1 {
		t.Fatalf("expected one selected card, got %#v", decision.Cards)
	}
	if !strings.Contains(raw, `"kind":"play"`) {
		t.Fatalf("raw trace missing play kind: %q", raw)
	}
}

func TestChooseMockDecisionPassesAgainstActiveTrick(t *testing.T) {
	service, err := demo.NewServiceWithOptions(demo.ServiceOptions{Mode: fsm.ModeFixedP0PlayTest})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	state, err := service.State()
	if err != nil {
		t.Fatalf("state: %v", err)
	}

	playState, err := service.Apply(demo.ActionRequest{
		Seat:  "P0",
		Kind:  "play",
		Cards: []string{state.Players[0].Cards[0].ID},
	})
	if err != nil {
		t.Fatalf("apply opening play: %v", err)
	}

	decision, _, err := agent.ChooseMockDecision(playState)
	if err != nil {
		t.Fatalf("choose mock decision: %v", err)
	}
	if decision.Kind != "pass" {
		t.Fatalf("expected mock runner to pass against active trick, got %#v", decision)
	}
}

func TestAgentServiceRunMockAppliesAndStoresTrace(t *testing.T) {
	service, err := demo.NewServiceWithOptions(demo.ServiceOptions{Mode: fsm.ModeFixedP0PlayTest})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	agentService := agent.NewService(service, agent.ServiceOptions{})
	response, err := agentService.Run(context.Background(), agent.RunRequest{Mode: agent.RunModeMock})
	if err != nil {
		t.Fatalf("run mock: %v", err)
	}

	if !response.Trace.Applied {
		t.Fatalf("expected applied trace, got %#v", response.Trace)
	}
	if response.Trace.ResultState == nil {
		t.Fatal("expected trace result state")
	}
	if response.State.CurrentActor != "P1" {
		t.Fatalf("expected actor P1 after mock play, got %q", response.State.CurrentActor)
	}
	stored := agentService.Trace()
	if stored == nil {
		t.Fatal("expected stored trace")
	}
	if stored.Decision.Kind != "play" {
		t.Fatalf("expected stored play decision, got %#v", stored.Decision)
	}
}

func TestMockAgentCanFinishFullMatchToWinner(t *testing.T) {
	service, err := demo.NewService()
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	agentService := agent.NewService(service, agent.ServiceOptions{})

	for step := 1; step <= 400; step++ {
		response, err := agentService.Run(context.Background(), agent.RunRequest{Mode: agent.RunModeMock})
		if err != nil {
			t.Fatalf("step %d run mock: %v", step, err)
		}
		if response.Trace.Error != "" {
			t.Fatalf("step %d unexpected trace error: %s", step, response.Trace.Error)
		}
		if response.Trace.Prompt.PlayerSeat == "" {
			t.Fatalf("step %d prompt player seat should not be empty", step)
		}
		if response.Trace.Prompt.CardCounter.PlayedCardsBySeat == nil {
			t.Fatalf("step %d card counter scaffold should be present", step)
		}
		if response.Trace.Prompt.RoundMemory.Seat == "" {
			t.Fatalf("step %d round memory scaffold should be present", step)
		}
		if response.State.Winner != "" {
			if response.Trace.Decision.Seat == "" {
				t.Fatalf("step %d winning trace should record a seat", step)
			}
			return
		}
	}

	t.Fatal("expected mock agent loop to finish a full match within 400 steps")
}

func TestRunMatchMockTracksWinnerCounterAndRoundMemory(t *testing.T) {
	service, err := demo.NewServiceWithOptions(demo.ServiceOptions{Mode: fsm.ModeFixedP0PlayTest})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	agentService := agent.NewService(service, agent.ServiceOptions{})
	response, err := agentService.RunMatch(context.Background(), agent.MatchRunRequest{
		Mode:      agent.RunModeMock,
		MaxSteps:  100,
		ResetGame: false,
	})
	if err != nil {
		t.Fatalf("run match: %v", err)
	}

	if response.Match.Status != "completed" {
		t.Fatalf("expected completed match, got %#v", response.Match)
	}
	if response.Match.Winner != "P0" {
		t.Fatalf("winner: got %q want P0", response.Match.Winner)
	}
	if len(response.Match.Steps) < 4 {
		t.Fatalf("expected at least four steps to inspect round transitions, got %d", len(response.Match.Steps))
	}

	first := response.Match.Steps[0]
	if first.Seat != "P0" || first.Prompt.PlayerSeat != "P0" {
		t.Fatalf("unexpected first step prompt seat: %#v", first)
	}
	if first.CardCounter.TotalPlayedCardCount != 0 {
		t.Fatalf("expected first step counter to start empty, got %#v", first.CardCounter)
	}

	second := response.Match.Steps[1]
	if second.Seat != "P1" {
		t.Fatalf("expected second step to belong to P1, got %#v", second)
	}
	if second.CardCounter.TotalPlayedCardCount != 1 {
		t.Fatalf("expected one played card before second step, got %#v", second.CardCounter)
	}
	if second.RoundMemory.LastOpponentPlay == nil || second.RoundMemory.LastOpponentPlay.Seat != "P0" {
		t.Fatalf("expected second step to remember landlord play as opponent, got %#v", second.RoundMemory)
	}
	if second.RoundMemory.TrickIndex != 1 {
		t.Fatalf("expected second step trick index 1, got %#v", second.RoundMemory)
	}

	fourth := response.Match.Steps[3]
	if fourth.Seat != "P0" {
		t.Fatalf("expected fourth step to start a fresh round for P0, got %#v", fourth)
	}
	if fourth.RoundMemory.RoundIndex != 2 || fourth.RoundMemory.TrickIndex != 0 {
		t.Fatalf("expected memory reset after trick clear, got %#v", fourth.RoundMemory)
	}
	if fourth.RoundMemory.LastOpponentPlay != nil || fourth.RoundMemory.LastTeammatePlay != nil || fourth.RoundMemory.LastSelfPlay != nil {
		t.Fatalf("expected round memory to reset after trick clear, got %#v", fourth.RoundMemory)
	}
}

func TestRunMatchOpenRouterFallsBackToMockAndStillFinishes(t *testing.T) {
	t.Setenv("OPENROUTER_API_KEY", "")

	service, err := demo.NewServiceWithOptions(demo.ServiceOptions{Mode: fsm.ModeFixedP0PlayTest})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	agentService := agent.NewService(service, agent.ServiceOptions{})
	response, err := agentService.RunMatch(context.Background(), agent.MatchRunRequest{
		Mode:      agent.RunModeOpenRouter,
		MaxSteps:  100,
		ResetGame: false,
	})
	if err != nil {
		t.Fatalf("run match: %v", err)
	}

	if response.Match.Status != "completed" || response.Match.Winner != "P0" {
		t.Fatalf("expected fallback match to finish with P0 winner, got %#v", response.Match)
	}
	if len(response.Match.Steps) == 0 {
		t.Fatal("expected match steps")
	}
	first := response.Match.Steps[0]
	if first.EffectiveMode != agent.RunModeMock {
		t.Fatalf("expected fallback to mock, got %#v", first)
	}
	if !strings.Contains(first.Error, "OPENROUTER_API_KEY") {
		t.Fatalf("expected step error to retain openrouter failure, got %#v", first)
	}
}

func TestAgentServiceRunOpenRouterWithoutAPIKeyReturnsStructuredTraceError(t *testing.T) {
	t.Setenv("OPENROUTER_API_KEY", "")

	service, err := demo.NewServiceWithOptions(demo.ServiceOptions{Mode: fsm.ModeFixedP0PlayTest})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	agentService := agent.NewService(service, agent.ServiceOptions{})
	response, err := agentService.Run(context.Background(), agent.RunRequest{Mode: agent.RunModeOpenRouter})
	if err != nil {
		t.Fatalf("run openrouter: %v", err)
	}

	if response.Trace.Error == "" {
		t.Fatal("expected trace error when openrouter key is missing")
	}
	if response.Trace.Applied {
		t.Fatalf("expected run to remain unapplied, got %#v", response.Trace)
	}
	if response.State.CurrentActor != "P0" {
		t.Fatalf("expected unchanged state, got actor %q", response.State.CurrentActor)
	}
}

func TestOpenRouterClientRunUsesChatCompletions(t *testing.T) {
	t.Setenv("OPENROUTER_API_KEY", "test-key")
	t.Setenv("OPENROUTER_BASE_URL", "http://example.test")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("authorization header: got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"seat\":\"P0\",\"kind\":\"play\",\"cards\":[\"card-1\"],\"reason\":\"ok\"}"}}]}`))
	}))
	defer server.Close()

	t.Setenv("OPENROUTER_BASE_URL", server.URL)

	client, err := agent.NewOpenRouterClient(server.Client())
	if err != nil {
		t.Fatalf("new openrouter client: %v", err)
	}

	decision, raw, err := client.Run(context.Background(), agent.PromptResponse{
		ModelHint:    agent.DefaultModelHint,
		SystemPrompt: "system",
		UserPrompt:   "user",
	}, "")
	if err != nil {
		t.Fatalf("run openrouter client: %v", err)
	}

	if decision.Kind != "play" || decision.Seat != "P0" {
		t.Fatalf("unexpected decision: %#v", decision)
	}
	if !strings.Contains(raw, `"kind":"play"`) {
		t.Fatalf("unexpected raw content: %q", raw)
	}
}

func TestNewOpenRouterClientRequiresAPIKey(t *testing.T) {
	t.Setenv("OPENROUTER_API_KEY", "")
	_, err := agent.NewOpenRouterClient(nil)
	if err == nil {
		t.Fatal("expected missing api key to fail")
	}
}
