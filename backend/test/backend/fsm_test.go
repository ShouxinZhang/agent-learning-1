package backend_test

import (
	"math/rand"
	"testing"

	"agent-dou-dizhu/internal/tiandi/domain"
	"agent-dou-dizhu/internal/tiandi/fsm"
)

func TestAllPassFallsBackToStartingBidder(t *testing.T) {
	m := fsm.NewMachine(rand.New(rand.NewSource(1)))
	if err := m.Start(); err != nil {
		t.Fatalf("start failed: %v", err)
	}

	start := m.Snapshot().StartingBidder
	for i := 0; i < domain.PlayerCount; i++ {
		s := m.Snapshot()
		if err := m.Apply(fsm.Action{Seat: s.CurrentActor, Kind: fsm.ActionBuJiao}); err != nil {
			t.Fatalf("bid %d failed: %v", i, err)
		}
	}

	s := m.Snapshot()
	if s.Phase != fsm.PhasePlay {
		t.Fatalf("expected PLAY phase, got %s", s.Phase)
	}
	if s.Landlord != start {
		t.Fatalf("expected landlord %s, got %s", start, s.Landlord)
	}
	if s.Multiplier != 1 {
		t.Fatalf("expected multiplier 1, got %d", s.Multiplier)
	}
	if len(s.Hands[s.Landlord]) != 20 {
		t.Fatalf("expected landlord to have 20 cards, got %d", len(s.Hands[s.Landlord]))
	}
}

func TestCandidateWinsWhenNobodyQiang(t *testing.T) {
	m := fsm.NewMachine(rand.New(rand.NewSource(2)))
	if err := m.Start(); err != nil {
		t.Fatalf("start failed: %v", err)
	}

	start := m.Snapshot()
	if err := m.Apply(fsm.Action{Seat: start.CurrentActor, Kind: fsm.ActionJiaoDiZhu}); err != nil {
		t.Fatalf("jiao failed: %v", err)
	}

	for i := 0; i < 2; i++ {
		s := m.Snapshot()
		if err := m.Apply(fsm.Action{Seat: s.CurrentActor, Kind: fsm.ActionBuQiang}); err != nil {
			t.Fatalf("qiang %d failed: %v", i, err)
		}
	}

	s := m.Snapshot()
	if s.Phase != fsm.PhasePlay {
		t.Fatalf("expected PLAY phase, got %s", s.Phase)
	}
	if s.Landlord != start.CurrentActor {
		t.Fatalf("expected landlord %s, got %s", start.CurrentActor, s.Landlord)
	}
	if s.Multiplier != 1 {
		t.Fatalf("expected multiplier 1, got %d", s.Multiplier)
	}
}

func TestLastRobberWinsWhenCandidateDeclines(t *testing.T) {
	m := fsm.NewMachine(rand.New(rand.NewSource(3)))
	if err := m.Start(); err != nil {
		t.Fatalf("start failed: %v", err)
	}

	start := m.Snapshot()
	candidate := start.CurrentActor
	if err := m.Apply(fsm.Action{Seat: candidate, Kind: fsm.ActionJiaoDiZhu}); err != nil {
		t.Fatalf("jiao failed: %v", err)
	}

	var lastRobber domain.Seat
	for i := 0; i < 2; i++ {
		s := m.Snapshot()
		lastRobber = s.CurrentActor
		if err := m.Apply(fsm.Action{Seat: s.CurrentActor, Kind: fsm.ActionQiangDiZhu}); err != nil {
			t.Fatalf("qiang %d failed: %v", i, err)
		}
	}

	s := m.Snapshot()
	if s.Phase != fsm.PhaseWoQiang {
		t.Fatalf("expected WOQIANG phase, got %s", s.Phase)
	}
	if err := m.Apply(fsm.Action{Seat: s.CurrentActor, Kind: fsm.ActionBuQiang}); err != nil {
		t.Fatalf("buqiang failed: %v", err)
	}

	s = m.Snapshot()
	if s.Landlord != lastRobber {
		t.Fatalf("expected landlord %s, got %s", lastRobber, s.Landlord)
	}
	if s.Multiplier != 4 {
		t.Fatalf("expected multiplier 4, got %d", s.Multiplier)
	}
}

func TestCandidateRetakesWithWoQiang(t *testing.T) {
	m := fsm.NewMachine(rand.New(rand.NewSource(4)))
	if err := m.Start(); err != nil {
		t.Fatalf("start failed: %v", err)
	}

	candidate := m.Snapshot().CurrentActor
	if err := m.Apply(fsm.Action{Seat: candidate, Kind: fsm.ActionJiaoDiZhu}); err != nil {
		t.Fatalf("jiao failed: %v", err)
	}

	s := m.Snapshot()
	if err := m.Apply(fsm.Action{Seat: s.CurrentActor, Kind: fsm.ActionQiangDiZhu}); err != nil {
		t.Fatalf("first qiang failed: %v", err)
	}
	s = m.Snapshot()
	if err := m.Apply(fsm.Action{Seat: s.CurrentActor, Kind: fsm.ActionBuQiang}); err != nil {
		t.Fatalf("second qiang failed: %v", err)
	}
	s = m.Snapshot()
	if err := m.Apply(fsm.Action{Seat: s.CurrentActor, Kind: fsm.ActionWoQiang}); err != nil {
		t.Fatalf("woqiang failed: %v", err)
	}

	s = m.Snapshot()
	if s.Landlord != candidate {
		t.Fatalf("expected landlord %s, got %s", candidate, s.Landlord)
	}
	if s.Multiplier != 4 {
		t.Fatalf("expected multiplier 4, got %d", s.Multiplier)
	}
}
