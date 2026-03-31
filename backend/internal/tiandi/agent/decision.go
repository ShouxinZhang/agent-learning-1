package agent

import (
	"encoding/json"
	"fmt"
	"strings"

	"agent-dou-dizhu/internal/tiandi/demo"
)

func ChooseMockDecision(state demo.StateResponse) (Decision, string, error) {
	if state.CurrentActor == "" {
		return Decision{}, "", fmt.Errorf("current actor is empty")
	}
	if contains(state.AvailableActions, "pass") && state.CurrentTrick != nil {
		decision := Decision{
			Seat:   state.CurrentActor,
			Kind:   "pass",
			Cards:  []string{},
			Reason: "mock runner preserves server legality and passes against an active trick.",
		}
		raw, _ := json.Marshal(decision)
		return decision, string(raw), nil
	}
	if contains(state.AvailableActions, "play") {
		player := findCurrentPlayer(state)
		if len(player.Cards) == 0 {
			return Decision{}, "", fmt.Errorf("current player has no cards")
		}
		decision := Decision{
			Seat:   state.CurrentActor,
			Kind:   "play",
			Cards:  []string{player.Cards[0].ID},
			Reason: "mock runner selects the first card as a legal single.",
		}
		raw, _ := json.Marshal(decision)
		return decision, string(raw), nil
	}

	if len(state.AvailableActions) == 0 {
		return Decision{}, "", fmt.Errorf("no available actions")
	}

	decision := Decision{
		Seat:   state.CurrentActor,
		Kind:   state.AvailableActions[0],
		Cards:  []string{},
		Reason: "mock runner selects the first available non-play action.",
	}
	raw, _ := json.Marshal(decision)
	return decision, string(raw), nil
}

func ParseDecisionText(raw string) (Decision, error) {
	cleaned := strings.TrimSpace(raw)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	start := strings.Index(cleaned, "{")
	end := strings.LastIndex(cleaned, "}")
	if start >= 0 && end >= start {
		cleaned = cleaned[start : end+1]
	}

	var decision Decision
	decoder := json.NewDecoder(strings.NewReader(cleaned))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&decision); err != nil {
		return Decision{}, fmt.Errorf("parse decision json: %w", err)
	}
	if decision.Cards == nil {
		decision.Cards = []string{}
	}
	return decision, nil
}

func ValidateDecision(state demo.StateResponse, decision Decision) error {
	if decision.Seat != state.CurrentActor {
		return fmt.Errorf("decision seat %q does not match current actor %q", decision.Seat, state.CurrentActor)
	}
	if !contains(state.AvailableActions, decision.Kind) {
		return fmt.Errorf("decision kind %q is not in available actions %v", decision.Kind, state.AvailableActions)
	}
	if decision.Kind == "play" {
		if len(decision.Cards) == 0 {
			return fmt.Errorf("play decision requires cards")
		}
		player := findCurrentPlayer(state)
		allowed := map[string]struct{}{}
		for _, card := range player.Cards {
			allowed[card.ID] = struct{}{}
		}
		for _, id := range decision.Cards {
			if _, ok := allowed[id]; !ok {
				return fmt.Errorf("card %q is not available in current hand", id)
			}
		}
		return nil
	}
	if len(decision.Cards) > 0 {
		return fmt.Errorf("non-play decision must not include cards")
	}
	return nil
}
