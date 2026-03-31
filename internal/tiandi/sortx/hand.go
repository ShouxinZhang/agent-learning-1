package sortx

import (
	"slices"

	"agent-dou-dizhu/internal/tiandi/domain"
)

func SortedHand(cards []domain.Card, laizi domain.LaiziPair) []domain.Card {
	out := append([]domain.Card(nil), cards...)
	SortHand(out, laizi)
	return out
}

func SortHand(cards []domain.Card, laizi domain.LaiziPair) {
	slices.SortStableFunc(cards, func(a, b domain.Card) int {
		aLaizi := laizi.IsLaizi(a)
		bLaizi := laizi.IsLaizi(b)
		switch {
		case aLaizi && !bLaizi:
			return -1
		case !aLaizi && bLaizi:
			return 1
		}

		aRank := rankWeight(a.Rank)
		bRank := rankWeight(b.Rank)
		if aRank != bRank {
			return bRank - aRank
		}

		aSuit := suitWeight(a.Suit)
		bSuit := suitWeight(b.Suit)
		return aSuit - bSuit
	})
}

func rankWeight(rank domain.Rank) int {
	switch rank {
	case domain.RankRedJoker:
		return 17
	case domain.RankBlackJoker:
		return 16
	default:
		return int(rank)
	}
}

func suitWeight(suit domain.Suit) int {
	switch suit {
	case domain.Joker:
		return 0
	case domain.Spade:
		return 1
	case domain.Heart:
		return 2
	case domain.Club:
		return 3
	case domain.Diamond:
		return 4
	default:
		return 5
	}
}
