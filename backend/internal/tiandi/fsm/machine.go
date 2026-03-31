package fsm

import (
	"fmt"
	"math/rand"

	"agent-dou-dizhu/internal/tiandi/domain"
	"agent-dou-dizhu/internal/tiandi/game"
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

type ActionKind string

const (
	ActionJiaoDiZhu  ActionKind = "jiaodizhu"
	ActionBuJiao     ActionKind = "bujiao"
	ActionQiangDiZhu ActionKind = "qiangdizhu"
	ActionBuQiang    ActionKind = "buqiang"
	ActionWoQiang    ActionKind = "woqiang"
)

type Action struct {
	Seat domain.Seat
	Kind ActionKind
}

type Snapshot struct {
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
}

type Machine struct {
	rng        *rand.Rand
	state      Snapshot
	bidActions int
	qiangQueue []domain.Seat
	qiangIndex int
}

func NewMachine(rng *rand.Rand) *Machine {
	return &Machine{
		rng: rng,
		state: Snapshot{
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
	return out
}

func (m *Machine) validateActor(seat domain.Seat) error {
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
			m.state.Phase = PhasePlay
			return nil
		default:
			return nil
		}
	}
}

func (p Phase) requiresInput() bool {
	return p == PhaseBid || p == PhaseQiangDiZhu || p == PhaseWoQiang
}
