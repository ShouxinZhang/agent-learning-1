package fsm

import (
	"fmt"
	"math/rand"

	"agent-dou-dizhu/internal/tiandi/domain"
	"agent-dou-dizhu/internal/tiandi/game"
	"agent-dou-dizhu/internal/tiandi/play"
)

type Phase string

const (
	PhaseShuffle       Phase = "SHUFFLE"
	PhaseLockBottom    Phase = "LOCK_BOTTOM"
	PhaseRollLaizi     Phase = "ROLL_LAIZI"
	PhaseShowDiLaizi   Phase = "SHOW_DILAIZI"
	PhaseDeal          Phase = "DEAL"
	PhaseBid           Phase = "BID"
	PhaseQiangDiZhu    Phase = "QIANGDIZHU"
	PhaseWoQiang       Phase = "WOQIANG"
	PhaseShowBottom    Phase = "SHOW_BOTTOM"
	PhaseShowTianLaizi Phase = "SHOW_TIANLAIZI"
	PhasePlay          Phase = "PLAY"
)

type Mode string

const (
	ModeNormal          Mode = "normal"
	ModeFixedP0PlayTest Mode = "fixed_p0_play_test"
)

type Options struct {
	Mode Mode
}

type ActionKind string

const (
	ActionJiaoDiZhu  ActionKind = "jiaodizhu"
	ActionBuJiao     ActionKind = "bujiao"
	ActionQiangDiZhu ActionKind = "qiangdizhu"
	ActionBuQiang    ActionKind = "buqiang"
	ActionWoQiang    ActionKind = "woqiang"
	ActionPlay       ActionKind = "play"
	ActionPass       ActionKind = "pass"
)

type Action struct {
	Seat         domain.Seat
	Kind         ActionKind
	Cards        []domain.Card
	ResolutionID string
}

type Snapshot struct {
	Mode              Mode
	Phase             Phase
	Deck              []domain.Card
	Bottom            []domain.Card
	Hands             map[domain.Seat][]domain.Card
	Laizi             domain.LaiziPair
	DiLaiziRevealed   bool
	TianLaiziRevealed bool
	StartingBidder    domain.Seat
	CurrentActor      domain.Seat
	Candidate         domain.Seat
	HasCandidate      bool
	Landlord          domain.Seat
	HasLandlord       bool
	Multiplier        int
	Robbers           []domain.Seat
	LeadingSeat       domain.Seat
	LastPlaySeat      domain.Seat
	HasCurrentTrick   bool
	CurrentTrick      play.CurrentTrick
	PassCount         int
	Winner            domain.Seat
	HasWinner         bool
	LastResolvedHand  play.ResolvedHand
	HasLastResolved   bool
	LastCandidates    []play.ResolutionCandidate
	LastPlayError     string
}

type Machine struct {
	rng        *rand.Rand
	options    Options
	state      Snapshot
	bidActions int
	qiangQueue []domain.Seat
	qiangIndex int
}

func NewMachine(rng *rand.Rand) *Machine {
	return NewMachineWithOptions(rng, Options{Mode: ModeNormal})
}

func NewMachineWithOptions(rng *rand.Rand, opts Options) *Machine {
	if opts.Mode == "" {
		opts.Mode = ModeNormal
	}

	return &Machine{
		rng:     rng,
		options: opts,
		state: Snapshot{
			Mode:       opts.Mode,
			Phase:      PhaseShuffle,
			Hands:      make(map[domain.Seat][]domain.Card, domain.PlayerCount),
			Multiplier: 1,
		},
	}
}

func (m *Machine) Start() error {
	return m.advanceAuto()
}

func (m *Machine) Apply(action Action) error {
	if err := m.validateActor(action.Seat); err != nil {
		return err
	}

	switch m.state.Phase {
	case PhaseBid:
		return m.applyBid(action)
	case PhaseQiangDiZhu:
		return m.applyQiang(action)
	case PhaseWoQiang:
		return m.applyWoQiang(action)
	case PhasePlay:
		return m.applyPlay(action)
	default:
		return fmt.Errorf("phase %s does not accept player input", m.state.Phase)
	}
}

func (m *Machine) Snapshot() Snapshot {
	hands := make(map[domain.Seat][]domain.Card, len(m.state.Hands))
	for seat, cards := range m.state.Hands {
		hands[seat] = append([]domain.Card(nil), cards...)
	}

	out := m.state
	out.Deck = append([]domain.Card(nil), m.state.Deck...)
	out.Bottom = append([]domain.Card(nil), m.state.Bottom...)
	out.Hands = hands
	out.Robbers = append([]domain.Seat(nil), m.state.Robbers...)
	if m.state.HasCurrentTrick {
		out.CurrentTrick.Cards = append([]domain.Card(nil), m.state.CurrentTrick.Cards...)
		out.CurrentTrick.ResolvedHand = play.CloneResolvedHand(m.state.CurrentTrick.ResolvedHand)
	}
	out.LastCandidates = play.CloneCandidates(m.state.LastCandidates)
	if m.state.HasLastResolved {
		out.LastResolvedHand = play.CloneResolvedHand(m.state.LastResolvedHand)
	}
	return out
}

func (m *Machine) validateActor(seat domain.Seat) error {
	if m.state.HasWinner {
		return fmt.Errorf("game is already finished")
	}
	if !m.state.Phase.requiresInput() {
		return fmt.Errorf("phase %s does not require player input", m.state.Phase)
	}
	if seat != m.state.CurrentActor {
		return fmt.Errorf("expected actor %s, got %s", m.state.CurrentActor, seat)
	}
	return nil
}

func (m *Machine) applyBid(action Action) error {
	switch action.Kind {
	case ActionJiaoDiZhu:
		m.state.Candidate = action.Seat
		m.state.HasCandidate = true
		m.prepareQiangPhase()
		return nil
	case ActionBuJiao:
		m.bidActions++
		if m.bidActions == domain.PlayerCount {
			m.setLandlord(m.state.StartingBidder)
			return m.advanceAuto()
		}
		m.state.CurrentActor = m.state.CurrentActor.Next()
		return nil
	default:
		return fmt.Errorf("invalid bid action %s", action.Kind)
	}
}

func (m *Machine) applyQiang(action Action) error {
	switch action.Kind {
	case ActionQiangDiZhu:
		m.state.Robbers = append(m.state.Robbers, action.Seat)
		m.state.Multiplier *= 2
	case ActionBuQiang:
	default:
		return fmt.Errorf("invalid qiang action %s", action.Kind)
	}

	m.qiangIndex++
	if m.qiangIndex < len(m.qiangQueue) {
		m.state.CurrentActor = m.qiangQueue[m.qiangIndex]
		return nil
	}

	if len(m.state.Robbers) == 0 {
		m.setLandlord(m.state.Candidate)
		return m.advanceAuto()
	}

	m.state.CurrentActor = m.state.Candidate
	m.state.Phase = PhaseWoQiang
	return nil
}

func (m *Machine) applyWoQiang(action Action) error {
	switch action.Kind {
	case ActionWoQiang:
		m.state.Multiplier *= 2
		m.setLandlord(m.state.Candidate)
	case ActionBuQiang:
		m.setLandlord(m.state.Robbers[len(m.state.Robbers)-1])
	default:
		return fmt.Errorf("invalid woqiang action %s", action.Kind)
	}

	return m.advanceAuto()
}

func (m *Machine) applyPlay(action Action) error {
	switch action.Kind {
	case ActionPass:
		return m.applyPass(action)
	case ActionPlay:
		return m.applyPlayCards(action)
	default:
		return fmt.Errorf("invalid play action %s", action.Kind)
	}
}

func (m *Machine) applyPass(action Action) error {
	if !m.state.HasCurrentTrick {
		return fmt.Errorf("pass is only allowed when there is an active trick")
	}
	m.state.LastPlayError = ""
	m.state.PassCount++
	m.state.CurrentTrick.PassCount = m.state.PassCount
	if m.state.PassCount >= 2 {
		m.state.HasCurrentTrick = false
		m.state.CurrentTrick = play.CurrentTrick{}
		m.state.PassCount = 0
		m.state.LeadingSeat = m.state.LastPlaySeat
		m.state.CurrentActor = m.state.LastPlaySeat
		return nil
	}

	m.state.CurrentActor = action.Seat.Next()
	return nil
}

func (m *Machine) applyPlayCards(action Action) error {
	if len(action.Cards) == 0 {
		return fmt.Errorf("play action requires selected cards")
	}
	if err := ensureCardsOwned(m.state.Hands[action.Seat], action.Cards); err != nil {
		return err
	}

	candidates, err := play.AnalyzeSelection(action.Cards, m.state.Laizi)
	if err != nil {
		m.state.LastPlayError = err.Error()
		return err
	}
	selected := candidates[0]
	if action.ResolutionID != "" {
		for _, candidate := range candidates {
			if candidate.ID == action.ResolutionID {
				selected = candidate
				break
			}
		}
	}

	if m.state.HasCurrentTrick {
		ok, compareErr := play.Beats(selected.ResolvedHand, m.state.CurrentTrick.ResolvedHand)
		if compareErr != nil {
			m.state.LastPlayError = compareErr.Error()
			return compareErr
		}
		if !ok {
			m.state.LastPlayError = "selected cards do not beat the current trick"
			return fmt.Errorf(m.state.LastPlayError)
		}
	} else {
		m.state.LeadingSeat = action.Seat
	}

	nextHand, err := removeCards(m.state.Hands[action.Seat], action.Cards)
	if err != nil {
		m.state.LastPlayError = err.Error()
		return err
	}
	m.state.Hands[action.Seat] = nextHand
	m.state.LastPlaySeat = action.Seat
	m.state.PassCount = 0
	m.state.HasCurrentTrick = true
	m.state.CurrentTrick = play.CurrentTrick{
		LeadingSeat:  m.state.LeadingSeat,
		LastPlaySeat: action.Seat,
		Cards:        append([]domain.Card(nil), action.Cards...),
		ResolvedHand: play.CloneResolvedHand(selected.ResolvedHand),
		PassCount:    0,
	}
	m.state.LastResolvedHand = play.CloneResolvedHand(selected.ResolvedHand)
	m.state.HasLastResolved = true
	m.state.LastCandidates = play.CloneCandidates(candidates)
	m.state.LastPlayError = ""

	if len(nextHand) == 0 {
		m.state.HasWinner = true
		m.state.Winner = action.Seat
		return nil
	}

	m.state.CurrentActor = action.Seat.Next()
	return nil
}

func (m *Machine) prepareQiangPhase() {
	m.qiangQueue = m.qiangQueue[:0]
	seat := m.state.StartingBidder.Next()
	for len(m.qiangQueue) < 2 {
		if seat != m.state.Candidate {
			m.qiangQueue = append(m.qiangQueue, seat)
		}
		seat = seat.Next()
	}
	m.qiangIndex = 0
	m.state.Phase = PhaseQiangDiZhu
	m.state.CurrentActor = m.qiangQueue[0]
}

func (m *Machine) setLandlord(seat domain.Seat) {
	m.state.Landlord = seat
	m.state.HasLandlord = true
	m.state.Phase = PhaseShowBottom
}

func (m *Machine) advanceAuto() error {
	for {
		switch m.state.Phase {
		case PhaseShuffle:
			deck := domain.FullDeck()
			game.ShuffleDeck(m.rng, deck)
			m.state.Deck = deck
			m.state.Phase = PhaseLockBottom
		case PhaseLockBottom:
			session, err := game.BuildSession(m.state.Deck)
			if err != nil {
				return err
			}
			m.state.Bottom = session.Bottom
			m.state.Hands = session.Hands
			m.state.Phase = PhaseRollLaizi
		case PhaseRollLaizi:
			m.state.Laizi = domain.SelectTianAndDiLaizi(m.rng)
			m.state.Phase = PhaseShowDiLaizi
		case PhaseShowDiLaizi:
			m.state.DiLaiziRevealed = true
			m.state.Phase = PhaseDeal
		case PhaseDeal:
			if m.options.Mode == ModeFixedP0PlayTest {
				m.enterFixedP0PlayTestMode()
				return nil
			}
			m.state.StartingBidder = domain.Seat(m.rng.Intn(domain.PlayerCount))
			m.state.CurrentActor = m.state.StartingBidder
			m.state.Phase = PhaseBid
			return nil
		case PhaseShowBottom:
			if !m.state.HasLandlord {
				return fmt.Errorf("landlord must be decided before showing bottom")
			}
			m.state.Hands[m.state.Landlord] = append(m.state.Hands[m.state.Landlord], m.state.Bottom...)
			m.state.Phase = PhaseShowTianLaizi
		case PhaseShowTianLaizi:
			m.state.TianLaiziRevealed = true
			m.state.CurrentActor = m.state.Landlord
			m.state.LeadingSeat = m.state.Landlord
			m.state.Phase = PhasePlay
			return nil
		default:
			return nil
		}
	}
}

func (m *Machine) enterFixedP0PlayTestMode() {
	m.state.StartingBidder = domain.Seat0
	m.state.CurrentActor = domain.Seat0
	m.state.Landlord = domain.Seat0
	m.state.HasLandlord = true
	m.state.Hands[domain.Seat0] = append(m.state.Hands[domain.Seat0], m.state.Bottom...)
	m.state.TianLaiziRevealed = true
	m.state.Multiplier = 1
	m.state.HasCandidate = false
	m.state.Candidate = 0
	m.state.Robbers = nil
	m.qiangQueue = nil
	m.qiangIndex = 0
	m.bidActions = 0
	m.state.Phase = PhasePlay
	m.state.LeadingSeat = domain.Seat0
}

func (p Phase) requiresInput() bool {
	return p == PhaseBid || p == PhaseQiangDiZhu || p == PhaseWoQiang || p == PhasePlay
}

func ensureCardsOwned(hand []domain.Card, selected []domain.Card) error {
	remaining := append([]domain.Card(nil), hand...)
	for _, card := range selected {
		found := false
		for index, owned := range remaining {
			if owned == card {
				remaining = append(remaining[:index], remaining[index+1:]...)
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("card %s is not in the current hand", card)
		}
	}
	return nil
}

func removeCards(hand []domain.Card, selected []domain.Card) ([]domain.Card, error) {
	out := append([]domain.Card(nil), hand...)
	for _, card := range selected {
		found := false
		for index, owned := range out {
			if owned == card {
				out = append(out[:index], out[index+1:]...)
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("card %s is not in the current hand", card)
		}
	}
	return out, nil
}
