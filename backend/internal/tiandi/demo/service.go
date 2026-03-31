package demo

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"agent-dou-dizhu/internal/tiandi/domain"
	"agent-dou-dizhu/internal/tiandi/fsm"
	"agent-dou-dizhu/internal/tiandi/sortx"
)

type CardView struct {
	Label   string `json:"label"`
	Suit    string `json:"suit"`
	Rank    string `json:"rank"`
	IsLaizi bool   `json:"isLaizi"`
}

type PlayerView struct {
	Seat       string     `json:"seat"`
	IsLandlord bool       `json:"isLandlord"`
	IsCurrent  bool       `json:"isCurrent"`
	Cards      []CardView `json:"cards"`
}

type BottomView struct {
	Visible bool       `json:"visible"`
	Count   int        `json:"count"`
	Cards   []CardView `json:"cards"`
}

type StateResponse struct {
	Phase            string       `json:"phase"`
	CurrentActor     string       `json:"currentActor"`
	AvailableActions []string     `json:"availableActions"`
	Players          []PlayerView `json:"players"`
	Landlord         string       `json:"landlord"`
	Multiplier       int          `json:"multiplier"`
	Message          string       `json:"message"`
	Laizi            struct {
		Tian        string `json:"tian"`
		Di          string `json:"di"`
		TianVisible bool   `json:"tianVisible"`
		DiVisible   bool   `json:"diVisible"`
	} `json:"laizi"`
	Bottom BottomView `json:"bottom"`
}

type ActionRequest struct {
	Seat string `json:"seat"`
	Kind string `json:"kind"`
}

type Service struct {
	mu      sync.Mutex
	machine *fsm.Machine
}

func NewService() (*Service, error) {
	svc := &Service{}
	_, err := svc.Reset()
	return svc, err
}

func (s *Service) Reset() (StateResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	machine := fsm.NewMachine(rand.New(rand.NewSource(time.Now().UnixNano())))
	if err := machine.Start(); err != nil {
		return StateResponse{}, err
	}

	s.machine = machine
	return buildState(s.machine.Snapshot()), nil
}

func (s *Service) State() (StateResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.machine == nil {
		return StateResponse{}, fmt.Errorf("game service is not initialized")
	}
	return buildState(s.machine.Snapshot()), nil
}

func (s *Service) Apply(req ActionRequest) (StateResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.machine == nil {
		return StateResponse{}, fmt.Errorf("game service is not initialized")
	}

	seat, err := domain.ParseSeat(req.Seat)
	if err != nil {
		return StateResponse{}, err
	}

	if err := s.machine.Apply(fsm.Action{
		Seat: seat,
		Kind: fsm.ActionKind(req.Kind),
	}); err != nil {
		return StateResponse{}, err
	}

	return buildState(s.machine.Snapshot()), nil
}

type DemoResponse struct {
	Players    []PlayerView `json:"players"`
	Landlord   string       `json:"landlord"`
	Multiplier int          `json:"multiplier"`
	Laizi      struct {
		Tian string `json:"tian"`
		Di   string `json:"di"`
	} `json:"laizi"`
	Bottom []CardView `json:"bottom"`
}

func Build(seed int64) (DemoResponse, error) {
	machine := fsm.NewMachine(rand.New(rand.NewSource(seed)))
	if err := machine.Start(); err != nil {
		return DemoResponse{}, err
	}

	for {
		s := machine.Snapshot()
		if s.Phase == fsm.PhasePlay {
			return buildResponse(s), nil
		}

		action := nextAutoAction(s)
		if err := machine.Apply(action); err != nil {
			return DemoResponse{}, err
		}
	}
}

func nextAutoAction(s fsm.Snapshot) fsm.Action {
	switch s.Phase {
	case fsm.PhaseBid:
		return fsm.Action{Seat: s.CurrentActor, Kind: fsm.ActionJiaoDiZhu}
	case fsm.PhaseQiangDiZhu:
		if len(s.Robbers) == 0 {
			return fsm.Action{Seat: s.CurrentActor, Kind: fsm.ActionBuQiang}
		}
		return fsm.Action{Seat: s.CurrentActor, Kind: fsm.ActionQiangDiZhu}
	case fsm.PhaseWoQiang:
		return fsm.Action{Seat: s.CurrentActor, Kind: fsm.ActionWoQiang}
	default:
		return fsm.Action{}
	}
}

func buildResponse(s fsm.Snapshot) DemoResponse {
	var resp DemoResponse
	resp.Landlord = s.Landlord.String()
	resp.Multiplier = s.Multiplier
	resp.Laizi.Tian = s.Laizi.Tian.String()
	resp.Laizi.Di = s.Laizi.Di.String()
	resp.Bottom = toCardViews(s.Bottom, s.Laizi)

	for _, seat := range domain.AllSeats() {
		cards := sortx.SortedHand(s.Hands[seat], s.Laizi)
		resp.Players = append(resp.Players, PlayerView{
			Seat:       seat.String(),
			IsLandlord: seat == s.Landlord,
			IsCurrent:  seat == s.CurrentActor && len(availableActions(s.Phase)) > 0,
			Cards:      toCardViews(cards, s.Laizi),
		})
	}

	return resp
}

func buildState(s fsm.Snapshot) StateResponse {
	var resp StateResponse
	resp.Phase = string(s.Phase)
	resp.AvailableActions = availableActions(s.Phase)
	if resp.AvailableActions == nil {
		resp.AvailableActions = []string{}
	}
	resp.Multiplier = s.Multiplier
	resp.Message = phaseMessage(s)
	if s.HasLandlord {
		resp.Landlord = s.Landlord.String()
	}
	if len(resp.AvailableActions) > 0 {
		resp.CurrentActor = s.CurrentActor.String()
	}

	resp.Laizi.DiVisible = s.DiLaiziRevealed
	resp.Laizi.TianVisible = s.TianLaiziRevealed
	if s.DiLaiziRevealed {
		resp.Laizi.Di = s.Laizi.Di.String()
	} else {
		resp.Laizi.Di = "?"
	}
	if s.TianLaiziRevealed {
		resp.Laizi.Tian = s.Laizi.Tian.String()
	} else {
		resp.Laizi.Tian = "?"
	}

	resp.Bottom.Visible = s.HasLandlord
	resp.Bottom.Count = len(s.Bottom)
	if resp.Bottom.Visible {
		resp.Bottom.Cards = toCardViews(s.Bottom, s.Laizi)
	}

	for _, seat := range domain.AllSeats() {
		cards := sortx.SortedHand(s.Hands[seat], s.Laizi)
		resp.Players = append(resp.Players, PlayerView{
			Seat:       seat.String(),
			IsLandlord: s.HasLandlord && seat == s.Landlord,
			IsCurrent:  len(resp.AvailableActions) > 0 && seat == s.CurrentActor,
			Cards:      toCardViews(cards, s.Laizi),
		})
	}

	return resp
}

func availableActions(phase fsm.Phase) []string {
	switch phase {
	case fsm.PhaseBid:
		return []string{string(fsm.ActionJiaoDiZhu), string(fsm.ActionBuJiao)}
	case fsm.PhaseQiangDiZhu:
		return []string{string(fsm.ActionQiangDiZhu), string(fsm.ActionBuQiang)}
	case fsm.PhaseWoQiang:
		return []string{string(fsm.ActionWoQiang), string(fsm.ActionBuQiang)}
	default:
		return []string{}
	}
}

func phaseMessage(s fsm.Snapshot) string {
	switch s.Phase {
	case fsm.PhaseBid:
		return fmt.Sprintf("%s 进行叫地主", s.CurrentActor)
	case fsm.PhaseQiangDiZhu:
		return fmt.Sprintf("%s 进行抢地主", s.CurrentActor)
	case fsm.PhaseWoQiang:
		return fmt.Sprintf("%s 进行我抢确认", s.CurrentActor)
	case fsm.PhasePlay:
		return "已完成出牌前流程，等待正式出牌"
	default:
		return string(s.Phase)
	}
}

func toCardViews(cards []domain.Card, laizi domain.LaiziPair) []CardView {
	out := make([]CardView, 0, len(cards))
	for _, card := range cards {
		out = append(out, CardView{
			Label:   card.String(),
			Suit:    card.Suit.String(),
			Rank:    card.Rank.String(),
			IsLaizi: laizi.IsLaizi(card),
		})
	}
	return out
}
