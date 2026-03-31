package agent

import (
	"context"
	"fmt"
	"strings"

	"agent-dou-dizhu/internal/tiandi/demo"
)

const (
	totalDeckCards       = 54
	defaultMatchMaxSteps = 256
)

type seatTracker struct {
	memory SeatRoundMemory
}

type matchTracker struct {
	roundIndex       int
	trickIndex       int
	playedBySeat     map[string][]string
	playedRankCounts map[string]int
	bombSignals      []string
	blackJokerPlayed bool
	redJokerPlayed   bool
	seats            map[string]*seatTracker
}

type turnExecution struct {
	model         string
	raw           string
	effectiveMode string
	decision      Decision
	nextState     demo.StateResponse
	applied       bool
	resultMessage string
	error         string
}

func newMatchTracker(state demo.StateResponse) *matchTracker {
	tracker := &matchTracker{
		roundIndex:       1,
		trickIndex:       0,
		playedBySeat:     map[string][]string{},
		playedRankCounts: map[string]int{},
		bombSignals:      []string{},
		seats:            map[string]*seatTracker{},
	}
	for _, player := range state.Players {
		tracker.ensureSeat(state, player.Seat)
	}
	return tracker
}

func (t *matchTracker) promptContext(state demo.StateResponse, seat string) PromptContext {
	t.ensureSeat(state, seat)
	return PromptContext{
		Seat:        seat,
		SeatRole:    seatRole(state, seat),
		CardCounter: t.cardCounter(state, seat),
		RoundMemory: t.roundMemory(state, seat),
	}
}

func (t *matchTracker) cardCounter(state demo.StateResponse, seat string) CardCounter {
	player := findPlayerBySeat(state, seat)
	counter := CardCounter{
		Seat:                 seat,
		SeatRole:             seatRole(state, seat),
		PlayedCardsBySeat:    cloneStringSliceMap(t.playedBySeat),
		PlayedRankCounts:     cloneIntMap(t.playedRankCounts),
		BlackJokerPlayed:     t.blackJokerPlayed,
		RedJokerPlayed:       t.redJokerPlayed,
		BombSignals:          append([]string{}, t.bombSignals...),
		TotalPlayedCardCount: totalPlayedCardCount(t.playedBySeat),
	}
	counter.RemainingUnknown = totalDeckCards - counter.TotalPlayedCardCount - len(player.Cards)
	if counter.RemainingUnknown < 0 {
		counter.RemainingUnknown = 0
	}
	return counter
}

func (t *matchTracker) roundMemory(state demo.StateResponse, seat string) SeatRoundMemory {
	t.ensureSeat(state, seat)
	memory := t.seats[seat].memory
	memory.Seat = seat
	memory.SeatRole = seatRole(state, seat)
	return cloneSeatRoundMemory(memory)
}

func (t *matchTracker) afterAction(before demo.StateResponse, decision Decision, after demo.StateResponse) {
	for _, player := range after.Players {
		t.ensureSeat(after, player.Seat)
	}

	if decision.Kind != "play" && decision.Kind != "pass" {
		return
	}

	t.trickIndex++
	for seat, tracker := range t.seats {
		tracker.memory.Seat = seat
		tracker.memory.SeatRole = seatRole(after, seat)
		tracker.memory.RoundIndex = t.roundIndex
		tracker.memory.TrickIndex = t.trickIndex
	}

	if decision.Kind == "play" && len(decision.Cards) > 0 {
		t.recordPlay(before, after, decision)
	}

	if before.CurrentTrick != nil && after.CurrentTrick == nil && after.Winner == "" {
		t.roundIndex++
		t.trickIndex = 0
		for seat, tracker := range t.seats {
			tracker.memory = zeroRoundMemory(seat, seatRole(after, seat))
			tracker.memory.RoundIndex = t.roundIndex
		}
	}
}

func (t *matchTracker) recordPlay(before demo.StateResponse, after demo.StateResponse, decision Decision) {
	action := PlayedAction{
		Seat:  decision.Seat,
		Kind:  decision.Kind,
		Cards: append([]string(nil), decision.Cards...),
	}
	if after.CurrentTrick != nil {
		action.ResolvedLabel = after.CurrentTrick.ResolvedHand.Label
		action.MainRank = after.CurrentTrick.ResolvedHand.MainRank
		if after.CurrentTrick.ResolvedHand.IsBomb {
			signal := fmt.Sprintf("%s:%s/%s", decision.Seat, after.CurrentTrick.ResolvedHand.Label, after.CurrentTrick.ResolvedHand.MainRank)
			if !contains(t.bombSignals, signal) {
				t.bombSignals = append(t.bombSignals, signal)
			}
		}
	}

	t.playedBySeat[decision.Seat] = append(t.playedBySeat[decision.Seat], decision.Cards...)
	for _, cardID := range decision.Cards {
		if card, ok := lookupCard(before, decision.Seat, cardID); ok {
			t.playedRankCounts[card.Rank]++
			if card.Rank == "BlackJoker" {
				t.blackJokerPlayed = true
			}
			if card.Rank == "RedJoker" {
				t.redJokerPlayed = true
			}
		}
	}

	for seat, tracker := range t.seats {
		relationship := relationshipForSeat(after, seat, decision.Seat)
		playCopy := action
		playCopy.Relationship = relationship
		switch relationship {
		case "self":
			tracker.memory.LastSelfPlay = clonePlayedActionPtr(&playCopy)
		case "teammate":
			tracker.memory.LastTeammatePlay = clonePlayedActionPtr(&playCopy)
		case "opponent":
			tracker.memory.LastOpponentPlay = clonePlayedActionPtr(&playCopy)
		}
	}
}

func (t *matchTracker) ensureSeat(state demo.StateResponse, seat string) {
	if strings.TrimSpace(seat) == "" {
		return
	}
	if _, ok := t.seats[seat]; ok {
		return
	}
	t.seats[seat] = &seatTracker{
		memory: zeroRoundMemory(seat, seatRole(state, seat)),
	}
	t.seats[seat].memory.RoundIndex = t.roundIndex
	t.seats[seat].memory.TrickIndex = t.trickIndex
}

func totalPlayedCardCount(playedBySeat map[string][]string) int {
	total := 0
	for _, cards := range playedBySeat {
		total += len(cards)
	}
	return total
}

func cloneStringSliceMap(in map[string][]string) map[string][]string {
	out := make(map[string][]string, len(in))
	for key, values := range in {
		out[key] = append([]string(nil), values...)
	}
	return out
}

func cloneIntMap(in map[string]int) map[string]int {
	out := make(map[string]int, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func clonePlayedActionPtr(in *PlayedAction) *PlayedAction {
	if in == nil {
		return nil
	}
	copyAction := *in
	copyAction.Cards = append([]string(nil), in.Cards...)
	return &copyAction
}

func cloneSeatRoundMemory(in SeatRoundMemory) SeatRoundMemory {
	in.LastSelfPlay = clonePlayedActionPtr(in.LastSelfPlay)
	in.LastTeammatePlay = clonePlayedActionPtr(in.LastTeammatePlay)
	in.LastOpponentPlay = clonePlayedActionPtr(in.LastOpponentPlay)
	return in
}

func relationshipForSeat(state demo.StateResponse, viewer string, actor string) string {
	if viewer == "" || actor == "" {
		return "unknown"
	}
	if viewer == actor {
		return "self"
	}
	if state.Landlord == "" {
		return "unknown"
	}
	if viewer == state.Landlord {
		return "opponent"
	}
	if actor == state.Landlord {
		return "opponent"
	}
	return "teammate"
}

func lookupCard(state demo.StateResponse, seat string, cardID string) (demo.CardView, bool) {
	player := findPlayerBySeat(state, seat)
	for _, card := range player.Cards {
		if card.ID == cardID {
			return card, true
		}
	}
	return demo.CardView{}, false
}

func attemptModes(mode string, fallback string) []string {
	primary := normalizeMode(mode)
	modes := []string{primary}
	if strings.TrimSpace(fallback) != "" {
		normalizedFallback := normalizeMode(fallback)
		if normalizedFallback != primary {
			modes = append(modes, normalizedFallback)
		}
	}
	return modes
}

func (s *Service) runTurn(ctx context.Context, state demo.StateResponse, prompt PromptResponse, mode string, fallbackMode string, model string) turnExecution {
	issues := []string{}
	currentState := state

	for _, candidateMode := range attemptModes(mode, fallbackMode) {
		result, err := s.runner.Run(ctx, prompt, candidateMode, model)
		execution := turnExecution{
			model:         result.Model,
			raw:           result.Raw,
			effectiveMode: candidateMode,
			decision:      result.Decision,
		}
		if err != nil {
			issues = append(issues, fmt.Sprintf("%s: %v", candidateMode, err))
			continue
		}
		if err := ValidateDecision(currentState, result.Decision); err != nil {
			issues = append(issues, fmt.Sprintf("%s: %v", candidateMode, err))
			continue
		}

		nextState, applyErr := s.game.Apply(demo.ActionRequest{
			Seat:         result.Decision.Seat,
			Kind:         result.Decision.Kind,
			Cards:        append([]string(nil), result.Decision.Cards...),
			ResolutionID: result.Decision.ResolutionID,
		})
		if applyErr != nil {
			issues = append(issues, fmt.Sprintf("%s: %v", candidateMode, applyErr))
			continue
		}
		if nextState.PlayError != "" {
			issues = append(issues, fmt.Sprintf("%s: %s", candidateMode, nextState.PlayError))
			continue
		}

		execution.nextState = nextState
		execution.applied = true
		execution.resultMessage = nextState.Message
		execution.error = strings.Join(issues, " | ")
		return execution
	}

	latestState, err := s.game.State()
	if err == nil {
		currentState = latestState
	}
	return turnExecution{
		effectiveMode: normalizeMode(mode),
		nextState:     currentState,
		resultMessage: currentState.Message,
		error:         strings.Join(issues, " | "),
	}
}
