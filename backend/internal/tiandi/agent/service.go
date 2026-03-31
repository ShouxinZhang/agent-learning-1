package agent

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"agent-dou-dizhu/internal/tiandi/demo"
)

var ErrUnsupportedMode = errors.New("unsupported agent mode")

type completion struct {
	Model    string
	Decision Decision
	Raw      string
}

type Runner interface {
	Run(ctx context.Context, prompt PromptResponse, mode string, model string) (completion, error)
}

type DefaultRunner struct {
	httpClient *http.Client
}

type ServiceOptions struct {
	Runner Runner
	Now    func() time.Time
}

type Service struct {
	game   *demo.Service
	runner Runner
	now    func() time.Time

	mu        sync.RWMutex
	lastTrace *Trace
	lastMatch *MatchTrace
}

func NewDefaultRunner(httpClient *http.Client) *DefaultRunner {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &DefaultRunner{httpClient: httpClient}
}

func NewService(game *demo.Service, opts ServiceOptions) *Service {
	runner := opts.Runner
	if runner == nil {
		runner = NewDefaultRunner(nil)
	}

	now := opts.Now
	if now == nil {
		now = time.Now
	}

	return &Service{
		game:   game,
		runner: runner,
		now:    now,
	}
}

func (s *Service) Prompt() (PromptResponse, error) {
	state, err := s.game.State()
	if err != nil {
		return PromptResponse{}, err
	}
	return BuildPrompt(state, s.game.Rules()), nil
}

func (s *Service) MatchState() (MatchStateResponse, error) {
	state, err := s.game.State()
	if err != nil {
		return MatchStateResponse{}, err
	}
	return MatchStateResponse{
		State: state,
		Match: s.Match(),
	}, nil
}

func (s *Service) Trace() *Trace {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.lastTrace == nil {
		return nil
	}

	trace := *s.lastTrace
	return &trace
}

func (s *Service) Match() *MatchTrace {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.lastMatch == nil {
		return nil
	}

	match := *s.lastMatch
	return &match
}

func (s *Service) ResetMatch() (MatchStateResponse, error) {
	state, err := s.game.Reset()
	if err != nil {
		return MatchStateResponse{}, err
	}

	s.mu.Lock()
	s.lastMatch = nil
	s.lastTrace = nil
	s.mu.Unlock()

	return MatchStateResponse{State: state}, nil
}

func (s *Service) Run(ctx context.Context, req RunRequest) (RunResponse, error) {
	mode := normalizeMode(req.Mode)
	if mode != RunModeMock && mode != RunModeOpenRouter {
		return RunResponse{}, fmt.Errorf("%w %q", ErrUnsupportedMode, req.Mode)
	}

	prompt, err := s.Prompt()
	if err != nil {
		return RunResponse{}, err
	}

	now := s.now().UTC()
	trace := Trace{
		RunID:     now.Format("20060102T150405.000000000Z07:00"),
		CreatedAt: now.Format(time.RFC3339Nano),
		Mode:      mode,
		Model:     resolveModel(mode, req.Model, prompt.ModelHint),
		Prompt:    prompt,
	}

	state := prompt.State
	if state.CurrentActor == "" || len(state.AvailableActions) == 0 {
		trace.Error = "no agent action is available for the current state"
		trace.ResultMessage = state.Message
		trace.ResultState = &state
		s.storeTrace(trace)
		return RunResponse{State: state, Trace: trace}, nil
	}

	result, err := s.runner.Run(ctx, prompt, mode, req.Model)
	if result.Model != "" {
		trace.Model = result.Model
	}
	trace.RawResponse = result.Raw
	trace.Decision = result.Decision
	if err != nil {
		trace.Error = err.Error()
		trace.ResultMessage = state.Message
		trace.ResultState = &state
		s.storeTrace(trace)
		return RunResponse{State: state, Trace: trace}, nil
	}

	if err := ValidateDecision(state, result.Decision); err != nil {
		trace.Error = err.Error()
		trace.ResultMessage = state.Message
		trace.ResultState = &state
		s.storeTrace(trace)
		return RunResponse{State: state, Trace: trace}, nil
	}

	nextState, err := s.game.Apply(demo.ActionRequest{
		Seat:         result.Decision.Seat,
		Kind:         result.Decision.Kind,
		Cards:        append([]string(nil), result.Decision.Cards...),
		ResolutionID: result.Decision.ResolutionID,
	})
	if err != nil {
		currentState, stateErr := s.game.State()
		if stateErr == nil {
			nextState = currentState
		} else {
			nextState = state
		}
		trace.Error = err.Error()
		trace.ResultMessage = nextState.Message
		trace.ResultState = &nextState
		s.storeTrace(trace)
		return RunResponse{State: nextState, Trace: trace}, nil
	}

	if nextState.PlayError != "" {
		trace.Error = nextState.PlayError
		trace.ResultMessage = nextState.Message
		trace.ResultState = &nextState
		s.storeTrace(trace)
		return RunResponse{State: nextState, Trace: trace}, nil
	}

	trace.Applied = true
	trace.ResultMessage = nextState.Message
	trace.ResultState = &nextState
	s.storeTrace(trace)
	return RunResponse{State: nextState, Trace: trace}, nil
}

func (s *Service) RunMatch(ctx context.Context, req MatchRunRequest) (MatchRunResponse, error) {
	mode := normalizeMode(req.Mode)
	if mode != RunModeMock && mode != RunModeOpenRouter {
		return MatchRunResponse{}, fmt.Errorf("%w %q", ErrUnsupportedMode, req.Mode)
	}

	fallbackMode := strings.TrimSpace(strings.ToLower(req.FallbackMode))
	if mode == RunModeOpenRouter && fallbackMode == "" {
		fallbackMode = RunModeMock
	}
	if fallbackMode != "" {
		fallbackMode = normalizeMode(fallbackMode)
		if fallbackMode != RunModeMock && fallbackMode != RunModeOpenRouter {
			return MatchRunResponse{}, fmt.Errorf("%w %q", ErrUnsupportedMode, req.FallbackMode)
		}
	}

	if req.ResetGame {
		if _, err := s.game.Reset(); err != nil {
			return MatchRunResponse{}, err
		}
	}

	state, err := s.game.State()
	if err != nil {
		return MatchRunResponse{}, err
	}

	now := s.now().UTC()
	trace := MatchTrace{
		MatchID:      now.Format("20060102T150405.000000000Z07:00"),
		StartedAt:    now.Format(time.RFC3339Nano),
		Status:       "running",
		Mode:         mode,
		FallbackMode: fallbackMode,
		Model:        resolveModel(mode, req.Model, DefaultModelHint),
		Steps:        []MatchStep{},
	}

	maxSteps := req.MaxSteps
	if maxSteps <= 0 {
		maxSteps = defaultMatchMaxSteps
	}

	tracker := newMatchTracker(state)
	for stepIndex := 1; stepIndex <= maxSteps; stepIndex++ {
		state, err = s.game.State()
		if err != nil {
			trace.Status = "failed"
			trace.Error = err.Error()
			break
		}
		if state.Winner != "" {
			trace.Status = "completed"
			trace.Winner = state.Winner
			trace.FinalState = &state
			break
		}
		if state.CurrentActor == "" || len(state.AvailableActions) == 0 {
			trace.Status = "failed"
			trace.Error = "match stopped before a winner because no current actor or actions were available"
			trace.FinalState = &state
			break
		}

		prompt := BuildPromptWithContext(state, s.game.Rules(), tracker.promptContext(state, state.CurrentActor))
		step := MatchStep{
			StepIndex:     stepIndex,
			Seat:          state.CurrentActor,
			AttemptMode:   mode,
			EffectiveMode: mode,
			Model:         resolveModel(mode, req.Model, prompt.ModelHint),
			Prompt:        prompt,
			RoundIndex:    prompt.RoundMemory.RoundIndex,
			TrickIndex:    prompt.RoundMemory.TrickIndex,
			CardCounter:   prompt.CardCounter,
			RoundMemory:   prompt.RoundMemory,
			StateBefore:   state,
		}

		execution := s.runTurn(ctx, state, prompt, mode, fallbackMode, req.Model)
		if execution.model != "" {
			step.Model = execution.model
		}
		if execution.effectiveMode != "" {
			step.EffectiveMode = execution.effectiveMode
		}
		step.Decision = execution.decision
		step.Applied = execution.applied
		step.Error = execution.error
		step.ResultMessage = execution.resultMessage
		step.StateAfter = execution.nextState
		trace.Steps = append(trace.Steps, step)
		trace.StepCount = len(trace.Steps)

		if !execution.applied {
			trace.Status = "failed"
			if trace.Error == "" {
				trace.Error = execution.error
			}
			finalState := execution.nextState
			trace.FinalState = &finalState
			break
		}

		tracker.afterAction(state, execution.decision, execution.nextState)

		if execution.nextState.Winner != "" {
			trace.Status = "completed"
			trace.Winner = execution.nextState.Winner
			finalState := execution.nextState
			trace.FinalState = &finalState
			break
		}
	}

	if trace.Status == "running" {
		currentState, stateErr := s.game.State()
		if stateErr == nil {
			trace.FinalState = &currentState
		}
		if trace.FinalState != nil && trace.FinalState.Winner != "" {
			trace.Status = "completed"
			trace.Winner = trace.FinalState.Winner
		} else {
			trace.Status = "max_steps_exceeded"
			trace.Error = fmt.Sprintf("match did not finish within %d steps", maxSteps)
		}
	}

	trace.FinishedAt = s.now().UTC().Format(time.RFC3339Nano)
	s.storeMatch(trace)

	finalState := state
	if trace.FinalState != nil {
		finalState = *trace.FinalState
	}
	return MatchRunResponse{
		State: finalState,
		Match: trace,
	}, nil
}

func (r *DefaultRunner) Run(ctx context.Context, prompt PromptResponse, mode string, model string) (completion, error) {
	switch normalizeMode(mode) {
	case RunModeMock:
		decision, raw, err := ChooseMockDecision(prompt.State)
		return completion{
			Model:    "mock/default",
			Decision: decision,
			Raw:      raw,
		}, err
	case RunModeOpenRouter:
		client, err := NewOpenRouterClient(r.httpClient)
		if err != nil {
			return completion{Model: resolveModel(RunModeOpenRouter, model, prompt.ModelHint)}, err
		}
		decision, raw, err := client.Run(ctx, prompt, model)
		return completion{
			Model:    resolveModel(RunModeOpenRouter, model, prompt.ModelHint),
			Decision: decision,
			Raw:      raw,
		}, err
	default:
		return completion{}, fmt.Errorf("%w %q", ErrUnsupportedMode, mode)
	}
}

func (s *Service) storeTrace(trace Trace) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stored := trace
	s.lastTrace = &stored
}

func (s *Service) storeMatch(match MatchTrace) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stored := match
	s.lastMatch = &stored
}

func normalizeMode(mode string) string {
	mode = strings.TrimSpace(strings.ToLower(mode))
	if mode == "" {
		return RunModeMock
	}
	return mode
}

func resolveModel(mode string, requested string, fallback string) string {
	if value := strings.TrimSpace(requested); value != "" {
		return value
	}
	if mode == RunModeMock {
		return "mock/default"
	}
	if value := strings.TrimSpace(fallback); value != "" {
		return value
	}
	return DefaultModelHint
}
