package domain

import "math/rand"

type LaiziPair struct {
	Tian Rank
	Di   Rank
}

func SelectTianAndDiLaizi(rng *rand.Rand) LaiziPair {
	ranks := append([]Rank(nil), NormalRanks()...)
	first := rng.Intn(len(ranks))
	tian := ranks[first]
	ranks = append(ranks[:first], ranks[first+1:]...)
	di := ranks[rng.Intn(len(ranks))]
	return LaiziPair{
		Tian: tian,
		Di:   di,
	}
}

func (p LaiziPair) IsLaizi(card Card) bool {
	if card.Rank.IsJoker() {
		return false
	}
	return card.Rank == p.Tian || card.Rank == p.Di
}

func (p LaiziPair) CanRepresent(target Rank) bool {
	return target.IsNormal()
}
