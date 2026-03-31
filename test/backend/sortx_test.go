package backend_test

import (
	"testing"

	"agent-dou-dizhu/internal/tiandi/domain"
	"agent-dou-dizhu/internal/tiandi/sortx"
)

func TestSortHandPutsLaiziFirstThenDescending(t *testing.T) {
	laizi := domain.LaiziPair{Tian: domain.Rank4, Di: domain.RankQ}
	cards := []domain.Card{
		{Suit: domain.Heart, Rank: domain.RankA},
		{Suit: domain.Spade, Rank: domain.Rank4},
		{Suit: domain.Joker, Rank: domain.RankRedJoker},
		{Suit: domain.Club, Rank: domain.Rank2},
		{Suit: domain.Diamond, Rank: domain.RankQ},
	}

	sortx.SortHand(cards, laizi)

	want := []domain.Card{
		{Suit: domain.Diamond, Rank: domain.RankQ},
		{Suit: domain.Spade, Rank: domain.Rank4},
		{Suit: domain.Joker, Rank: domain.RankRedJoker},
		{Suit: domain.Club, Rank: domain.Rank2},
		{Suit: domain.Heart, Rank: domain.RankA},
	}

	for i := range want {
		if cards[i] != want[i] {
			t.Fatalf("index %d: got %+v want %+v", i, cards[i], want[i])
		}
	}
}
