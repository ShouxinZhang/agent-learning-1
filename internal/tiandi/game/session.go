package game

import (
	"fmt"
	"math/rand"

	"agent-dou-dizhu/internal/tiandi/domain"
)

type Session struct {
	Deck   []domain.Card
	Bottom []domain.Card
	Hands  map[domain.Seat][]domain.Card
}

func ShuffleDeck(rng *rand.Rand, deck []domain.Card) {
	rng.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})
}

func BuildSession(deck []domain.Card) (Session, error) {
	if len(deck) != 54 {
		return Session{}, fmt.Errorf("expected 54 cards, got %d", len(deck))
	}

	session := Session{
		Deck:   append([]domain.Card(nil), deck...),
		Bottom: append([]domain.Card(nil), deck[:3]...),
		Hands:  make(map[domain.Seat][]domain.Card, domain.PlayerCount),
	}

	rest := deck[3:]
	for idx, seat := range domain.AllSeats() {
		start := idx * 17
		end := start + 17
		session.Hands[seat] = append([]domain.Card(nil), rest[start:end]...)
	}

	return session, nil
}
