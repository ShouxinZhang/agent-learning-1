package demo

import (
	"fmt"
	"math/rand"
	"slices"
	"sync"
	"time"

	"agent-dou-dizhu/internal/tiandi/domain"
	"agent-dou-dizhu/internal/tiandi/fsm"
	"agent-dou-dizhu/internal/tiandi/play"
	"agent-dou-dizhu/internal/tiandi/rules"
	"agent-dou-dizhu/internal/tiandi/sortx"
)

type CardView struct {
	ID      string `json:"id"`
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

type TestModeView struct {
	Enabled       bool   `json:"enabled"`
	Label         string `json:"label"`
	FixedLandlord string `json:"fixedLandlord"`
	DirectPlay    bool   `json:"directPlay"`
}

type ResolvedHandView struct {
	Kind           string     `json:"kind"`
	Label          string     `json:"label"`
	Pattern        string     `json:"pattern"`
	MainRank       string     `json:"mainRank"`
	Length         int        `json:"length"`
	GroupCount     int        `json:"groupCount"`
	AttachmentType string     `json:"attachmentType"`
	CompareKey     string     `json:"compareKey"`
	UsesLaizi      bool       `json:"usesLaizi"`
	IsBomb         bool       `json:"isBomb"`
	BombTier       int        `json:"bombTier"`
	Cards          []CardView `json:"cards"`
	ResolvedCards  []CardView `json:"resolvedCards"`
}

type ResolutionCandidateView struct {
	ID           string           `json:"id"`
	Priority     int              `json:"priority"`
	IsPreferred  bool             `json:"isPreferred"`
	ResolvedHand ResolvedHandView `json:"resolvedHand"`
}

type CurrentTrickView struct {
	LeadingSeat  string           `json:"leadingSeat"`
	LastPlaySeat string           `json:"lastPlaySeat"`
	PassCount    int              `json:"passCount"`
	Cards        []CardView       `json:"cards"`
	ResolvedHand ResolvedHandView `json:"resolvedHand"`
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
	Bottom               BottomView                `json:"bottom"`
	TestMode             *TestModeView             `json:"testMode,omitempty"`
	CurrentTrick         *CurrentTrickView         `json:"currentTrick,omitempty"`
	ResolvedHand         *ResolvedHandView         `json:"resolvedHand,omitempty"`
	ResolutionCandidates []ResolutionCandidateView `json:"resolutionCandidates"`
	PlayError            string                    `json:"playError"`
	Winner               string                    `json:"winner"`
}

type ActionRequest struct {
	Seat         string   `json:"seat"`
	Kind         string   `json:"kind"`
	Cards        []string `json:"cards,omitempty"`
	ResolutionID string   `json:"resolutionId,omitempty"`
}

type ServiceOptions struct {
	Mode fsm.Mode
}

type Service struct {
	mu      sync.Mutex
	machine *fsm.Machine
	options ServiceOptions
}

func NewService() (*Service, error) {
	return NewServiceWithOptions(ServiceOptions{Mode: fsm.ModeNormal})
}

func NewServiceWithOptions(opts ServiceOptions) (*Service, error) {
	if opts.Mode == "" {
		opts.Mode = fsm.ModeNormal
	}

	svc := &Service{options: opts}
	_, err := svc.Reset()
	return svc, err
}

func (s *Service) Reset() (StateResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	machine := fsm.NewMachineWithOptions(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		fsm.Options{Mode: s.options.Mode},
	)
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

	selectedCards, err := parseCardsForSeat(s.machine.Snapshot().Hands[seat], req.Cards)
	if err != nil {
		if isPlayPhaseAction(req.Kind) {
			state := buildState(s.machine.Snapshot())
			state.PlayError = err.Error()
			return state, nil
		}
		return StateResponse{}, err
	}

	if err := s.machine.Apply(fsm.Action{
		Seat:         seat,
		Kind:         fsm.ActionKind(req.Kind),
		Cards:        selectedCards,
		ResolutionID: req.ResolutionID,
	}); err != nil {
		if isPlayPhaseAction(req.Kind) {
			state := buildState(s.machine.Snapshot())
			if state.PlayError == "" {
				state.PlayError = err.Error()
			}
			return state, nil
		}
		return StateResponse{}, err
	}

	return buildState(s.machine.Snapshot()), nil
}

func (s *Service) Rules() rules.Catalog {
	return rules.CatalogData()
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
			IsCurrent:  seat == s.CurrentActor && len(availableActions(s)) > 0,
			Cards:      toCardViews(cards, s.Laizi),
		})
	}

	return resp
}

func buildState(s fsm.Snapshot) StateResponse {
	var resp StateResponse
	resp.Phase = string(s.Phase)
	resp.AvailableActions = availableActions(s)
	if resp.AvailableActions == nil {
		resp.AvailableActions = []string{}
	}
	resp.ResolutionCandidates = []ResolutionCandidateView{}
	resp.Multiplier = s.Multiplier
	resp.Message = phaseMessage(s)
	if s.HasLandlord {
		resp.Landlord = s.Landlord.String()
	}
	if s.HasWinner {
		resp.Winner = s.Winner.String()
	}
	if len(resp.AvailableActions) > 0 {
		resp.CurrentActor = s.CurrentActor.String()
	}
	if s.Mode == fsm.ModeFixedP0PlayTest {
		resp.TestMode = &TestModeView{
			Enabled:       true,
			Label:         "当前为测试模式 / 固定地主为 P0 / 直接进入 PLAY",
			FixedLandlord: domain.Seat0.String(),
			DirectPlay:    true,
		}
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
	if s.HasCurrentTrick {
		resp.CurrentTrick = &CurrentTrickView{
			LeadingSeat:  s.CurrentTrick.LeadingSeat.String(),
			LastPlaySeat: s.CurrentTrick.LastPlaySeat.String(),
			PassCount:    s.CurrentTrick.PassCount,
			Cards:        toCardViews(s.CurrentTrick.Cards, s.Laizi),
			ResolvedHand: toResolvedHandView(s.CurrentTrick.ResolvedHand, s.Laizi),
		}
	}
	if s.HasLastResolved {
		hand := toResolvedHandView(s.LastResolvedHand, s.Laizi)
		resp.ResolvedHand = &hand
	}
	if s.LastPlayError != "" {
		resp.PlayError = s.LastPlayError
	}
	if len(s.LastCandidates) > 0 {
		resp.ResolutionCandidates = toCandidateViews(s.LastCandidates, s.Laizi)
	}

	for _, seat := range domain.AllSeats() {
		cards := sortx.SortedHand(s.Hands[seat], s.Laizi)
		resp.Players = append(resp.Players, PlayerView{
			Seat:       seat.String(),
			IsLandlord: s.HasLandlord && seat == s.Landlord,
			IsCurrent:  len(resp.AvailableActions) > 0 && !s.HasWinner && seat == s.CurrentActor,
			Cards:      toCardViews(cards, s.Laizi),
		})
	}

	return resp
}

func availableActions(s fsm.Snapshot) []string {
	if s.HasWinner {
		return []string{}
	}
	switch s.Phase {
	case fsm.PhaseBid:
		return []string{string(fsm.ActionJiaoDiZhu), string(fsm.ActionBuJiao)}
	case fsm.PhaseQiangDiZhu:
		return []string{string(fsm.ActionQiangDiZhu), string(fsm.ActionBuQiang)}
	case fsm.PhaseWoQiang:
		return []string{string(fsm.ActionWoQiang), string(fsm.ActionBuQiang)}
	case fsm.PhasePlay:
		actions := []string{string(fsm.ActionPlay)}
		if s.HasCurrentTrick {
			actions = append(actions, string(fsm.ActionPass))
		}
		return actions
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
		if s.HasWinner {
			return fmt.Sprintf("%s 已获胜", s.Winner)
		}
		if s.HasCurrentTrick {
			return fmt.Sprintf("%s 跟牌或不出", s.CurrentActor)
		}
		if s.Mode == fsm.ModeFixedP0PlayTest {
			return "测试模式：已直接进入 PLAY，地主固定为 P0，可开始正式出牌"
		}
		return fmt.Sprintf("%s 首出", s.CurrentActor)
	default:
		return string(s.Phase)
	}
}

func toCardViews(cards []domain.Card, laizi domain.LaiziPair) []CardView {
	out := make([]CardView, 0, len(cards))
	for _, card := range cards {
		out = append(out, CardView{
			ID:      card.String(),
			Label:   card.String(),
			Suit:    card.Suit.String(),
			Rank:    card.Rank.String(),
			IsLaizi: laizi.IsLaizi(card),
		})
	}
	return out
}

func toResolvedHandView(hand play.ResolvedHand, laizi domain.LaiziPair) ResolvedHandView {
	view := ResolvedHandView{
		Kind:           string(hand.Kind),
		Label:          hand.Label,
		Pattern:        hand.Pattern,
		MainRank:       hand.MainRank.String(),
		Length:         hand.Length,
		GroupCount:     hand.GroupCount,
		AttachmentType: hand.AttachmentType,
		CompareKey:     hand.CompareKey,
		UsesLaizi:      hand.UsesLaizi,
		IsBomb:         hand.IsBomb,
		BombTier:       hand.BombTier,
		Cards:          toCardViews(hand.Cards, laizi),
	}
	for _, card := range hand.ResolvedCards {
		view.ResolvedCards = append(view.ResolvedCards, toCardViews([]domain.Card{card}, laizi)[0])
	}
	return view
}

func toCandidateViews(candidates []play.ResolutionCandidate, laizi domain.LaiziPair) []ResolutionCandidateView {
	out := make([]ResolutionCandidateView, 0, len(candidates))
	for _, candidate := range candidates {
		out = append(out, ResolutionCandidateView{
			ID:           candidate.ID,
			Priority:     candidate.Priority,
			IsPreferred:  candidate.IsPreferred,
			ResolvedHand: toResolvedHandView(candidate.ResolvedHand, laizi),
		})
	}
	return out
}

func parseCardsForSeat(hand []domain.Card, ids []string) ([]domain.Card, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	remaining := append([]domain.Card(nil), hand...)
	selected := make([]domain.Card, 0, len(ids))
	for _, id := range ids {
		index := slices.IndexFunc(remaining, func(card domain.Card) bool {
			return card.String() == id
		})
		if index < 0 {
			return nil, fmt.Errorf("card %q is not available in the current hand", id)
		}
		selected = append(selected, remaining[index])
		remaining = append(remaining[:index], remaining[index+1:]...)
	}
	return selected, nil
}

func isPlayPhaseAction(kind string) bool {
	return kind == string(fsm.ActionPlay) || kind == string(fsm.ActionPass)
}
