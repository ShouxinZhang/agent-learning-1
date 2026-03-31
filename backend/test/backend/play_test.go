package backend_test

import (
	"testing"

	"agent-dou-dizhu/internal/tiandi/demo"
	"agent-dou-dizhu/internal/tiandi/domain"
	"agent-dou-dizhu/internal/tiandi/fsm"
	"agent-dou-dizhu/internal/tiandi/play"
	"agent-dou-dizhu/internal/tiandi/rules"
)

func TestAnalyzeSelectionPrefersLaiziBombCandidate(t *testing.T) {
	laizi := domain.LaiziPair{Tian: domain.RankQ, Di: domain.RankK}
	cards := []domain.Card{
		{Suit: domain.Spade, Rank: domain.RankA},
		{Suit: domain.Heart, Rank: domain.RankA},
		{Suit: domain.Club, Rank: domain.RankA},
		{Suit: domain.Diamond, Rank: domain.RankQ},
	}

	candidates, err := play.AnalyzeSelection(cards, laizi)
	if err != nil {
		t.Fatalf("analyze selection: %v", err)
	}
	if len(candidates) < 2 {
		t.Fatalf("expected multiple candidates, got %d", len(candidates))
	}
	if candidates[0].ResolvedHand.Kind != rules.KindLaiziSubstituteBomb {
		t.Fatalf("expected preferred candidate to be laizi bomb, got %s", candidates[0].ResolvedHand.Kind)
	}
	if !candidates[0].IsPreferred {
		t.Fatal("expected first candidate to be preferred")
	}
}

func TestAttachmentCoreComparisonUsesMainRankOnly(t *testing.T) {
	left := play.ResolvedHand{Kind: rules.KindFourWithTwoSingles, MainRank: domain.RankA}
	right := play.ResolvedHand{Kind: rules.KindTripleWithPair, MainRank: domain.RankK}

	ok, err := play.Beats(left, right)
	if err != nil {
		t.Fatalf("beats: %v", err)
	}
	if !ok {
		t.Fatal("expected four-with-two-singles to beat triple-with-pair by main rank")
	}
}

func TestServicePlayAndPassFlow(t *testing.T) {
	service, err := demo.NewServiceWithOptions(demo.ServiceOptions{Mode: fsm.ModeFixedP0PlayTest})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	state, err := service.State()
	if err != nil {
		t.Fatalf("state: %v", err)
	}
	if state.CurrentActor != "P0" {
		t.Fatalf("expected P0 to act first, got %q", state.CurrentActor)
	}

	firstCard := state.Players[0].Cards[0].ID
	state, err = service.Apply(demo.ActionRequest{
		Seat:  "P0",
		Kind:  "play",
		Cards: []string{firstCard},
	})
	if err != nil {
		t.Fatalf("play first card: %v", err)
	}
	if state.CurrentTrick == nil {
		t.Fatal("expected current trick after first play")
	}
	if state.CurrentTrick.LastPlaySeat != "P0" {
		t.Fatalf("expected last play seat P0, got %q", state.CurrentTrick.LastPlaySeat)
	}
	if state.CurrentActor != "P1" {
		t.Fatalf("expected next actor P1, got %q", state.CurrentActor)
	}
	if len(state.Players[0].Cards) != 19 {
		t.Fatalf("expected P0 to have 19 cards left, got %d", len(state.Players[0].Cards))
	}

	state, err = service.Apply(demo.ActionRequest{Seat: "P1", Kind: "pass"})
	if err != nil {
		t.Fatalf("pass for P1: %v", err)
	}
	if state.CurrentActor != "P2" {
		t.Fatalf("expected current actor P2 after first pass, got %q", state.CurrentActor)
	}
	if state.CurrentTrick == nil || state.CurrentTrick.PassCount != 1 {
		t.Fatalf("expected pass count 1, got %#v", state.CurrentTrick)
	}

	state, err = service.Apply(demo.ActionRequest{Seat: "P2", Kind: "pass"})
	if err != nil {
		t.Fatalf("pass for P2: %v", err)
	}
	if state.CurrentActor != "P0" {
		t.Fatalf("expected turn to reset to P0, got %q", state.CurrentActor)
	}
	if state.CurrentTrick != nil {
		t.Fatal("expected trick to clear after two passes")
	}
	if len(state.AvailableActions) != 1 || state.AvailableActions[0] != "play" {
		t.Fatalf("expected actions [play] after trick reset, got %#v", state.AvailableActions)
	}
}

func TestServiceReturnsInlineErrorForPassWithoutCurrentTrick(t *testing.T) {
	service, err := demo.NewServiceWithOptions(demo.ServiceOptions{Mode: fsm.ModeFixedP0PlayTest})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	state, err := service.Apply(demo.ActionRequest{Seat: "P0", Kind: "pass"})
	if err != nil {
		t.Fatalf("expected inline error state, got err: %v", err)
	}
	if state.PlayError == "" {
		t.Fatal("expected playError to be populated")
	}
	if state.CurrentActor != "P0" {
		t.Fatalf("expected current actor to remain P0, got %q", state.CurrentActor)
	}
}
