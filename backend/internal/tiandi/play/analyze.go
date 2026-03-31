package play

import (
	"fmt"
	"sort"

	"agent-dou-dizhu/internal/tiandi/domain"
	"agent-dou-dizhu/internal/tiandi/rules"
)

func AnalyzeSelection(cards []domain.Card, laizi domain.LaiziPair) ([]ResolutionCandidate, error) {
	if len(cards) == 0 {
		return nil, fmt.Errorf("at least one card is required")
	}

	sortedCards := append([]domain.Card(nil), cards...)
	sort.Slice(sortedCards, func(i, j int) bool {
		if sortedCards[i].Rank == sortedCards[j].Rank {
			return sortedCards[i].Suit < sortedCards[j].Suit
		}
		return sortedCards[i].Rank < sortedCards[j].Rank
	})

	candidates := make([]ResolutionCandidate, 0, 2)
	if hand, ok := analyzeStandard(sortedCards); ok {
		candidates = append(candidates, ResolutionCandidate{
			ID:           "candidate-standard",
			ResolvedHand: hand,
			Priority:     handPriority(hand),
		})
	}
	if hand, ok := analyzeLaizi(sortedCards, laizi); ok {
		candidates = append(candidates, ResolutionCandidate{
			ID:           "candidate-laizi",
			ResolvedHand: hand,
			Priority:     handPriority(hand),
		})
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("selected cards do not form a supported hand")
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Priority != candidates[j].Priority {
			return candidates[i].Priority < candidates[j].Priority
		}
		if candidates[i].ResolvedHand.MainRank != candidates[j].ResolvedHand.MainRank {
			return candidates[i].ResolvedHand.MainRank > candidates[j].ResolvedHand.MainRank
		}
		return candidates[i].ID < candidates[j].ID
	})
	candidates[0].IsPreferred = true
	return candidates, nil
}

func analyzeStandard(cards []domain.Card) (ResolvedHand, bool) {
	counts := rankCounts(cards)
	switch len(cards) {
	case 1:
		return newHand(rules.KindSingle, "单张", "A", cards, cards, cards[0].Rank, 1, "", false, 0), true
	case 2:
		if isRocket(cards) {
			return newHand(rules.KindRocket, "王炸", "BlackJoker + RedJoker", cards, cards, domain.RankRedJoker, 1, "", false, bombTier(rules.KindRocket)), true
		}
		if len(counts) == 1 && !cards[0].Rank.IsJoker() {
			return newHand(rules.KindPair, "对子", "AA", cards, cards, cards[0].Rank, 1, "", false, 0), true
		}
	case 3:
		if len(counts) == 1 && !cards[0].Rank.IsJoker() {
			return newHand(rules.KindTriple, "三张", "AAA", cards, cards, cards[0].Rank, 1, "", false, 0), true
		}
	case 4:
		if rank, ok := sameRank(counts); ok && !rank.IsJoker() {
			return newHand(rules.KindBombFour, "四张炸弹", "AAAA", cards, cards, rank, 1, "", false, bombTier(rules.KindBombFour)), true
		}
		if rank, ok := coreRank(counts, 3); ok {
			return newHand(rules.KindTripleWithSingle, "三带一", "AAA + B", cards, cards, rank, 1, "solo", false, 0), true
		}
	case 5:
		if rank, ok := sameRank(counts); ok && !rank.IsJoker() {
			return newHand(rules.KindBombFivePlus, "5+ 炸弹", "AAAAA...", cards, cards, rank, 1, "", false, bombTier(rules.KindBombFivePlus)), true
		}
		if rank, ok := coreAndPair(counts); ok {
			return newHand(rules.KindTripleWithPair, "三带二", "AAA + BB", cards, cards, rank, 1, "pair", false, 0), true
		}
		if isStraight(cards, counts) {
			return newHand(rules.KindStraight, "顺子", "ABCDE...", cards, cards, highestRank(cards), len(cards), "", false, 0), true
		}
	default:
		if rank, ok := sameRank(counts); ok && !rank.IsJoker() {
			return newHand(rules.KindBombFivePlus, "5+ 炸弹", "AAAAA...", cards, cards, rank, 1, "", false, bombTier(rules.KindBombFivePlus)), true
		}
		if isStraight(cards, counts) {
			return newHand(rules.KindStraight, "顺子", "ABCDE...", cards, cards, highestRank(cards), len(cards), "", false, 0), true
		}
		if isSerialPairs(cards, counts) {
			return newHand(rules.KindSerialPairs, "连对", "AA BB CC...", cards, cards, highestCountRank(counts, 2), len(cards)/2, "", false, 0), true
		}
		if hand, ok := analyzePlane(cards, counts); ok {
			return hand, true
		}
		if len(cards) == 6 {
			if rank, ok := coreRank(counts, 4); ok && len(counts) == 3 && hasExactCounts(counts, 4, 1, 1) {
				return newHand(rules.KindFourWithTwoSingles, "四带两单", "AAAA + B + C", cards, cards, rank, 1, "solo", false, 0), true
			}
		}
		if len(cards) == 8 {
			if rank, ok := coreRank(counts, 4); ok && len(counts) == 3 && hasExactCounts(counts, 4, 2, 2) {
				return newHand(rules.KindFourWithTwoPairs, "四带两对", "AAAA + BB + CC", cards, cards, rank, 1, "pairs", false, 0), true
			}
		}
	}

	return ResolvedHand{}, false
}

func analyzeLaizi(cards []domain.Card, laizi domain.LaiziPair) (ResolvedHand, bool) {
	laiziCards := make([]domain.Card, 0, len(cards))
	normalCards := make([]domain.Card, 0, len(cards))
	for _, card := range cards {
		if laizi.IsLaizi(card) {
			laiziCards = append(laiziCards, card)
		} else {
			normalCards = append(normalCards, card)
		}
	}
	if len(laiziCards) == 0 {
		return ResolvedHand{}, false
	}
	if len(cards) == 1 {
		return newHand(rules.KindSingle, "单张", "A", cards, cards, cards[0].Rank, 1, "", false, 0), true
	}

	counts := rankCounts(normalCards)
	if len(cards) == 4 {
		if len(normalCards) == 0 && sameRankCards(cards) {
			return newHand(rules.KindPureLaiziBomb, "纯赖子炸弹", "LLLL", cards, cards, cards[0].Rank, 1, "", true, bombTier(rules.KindPureLaiziBomb)), true
		}
		for target := range counts {
			if counts[target]+len(laiziCards) == 4 && !target.IsJoker() {
				resolved := append([]domain.Card(nil), normalCards...)
				for i := 0; i < len(laiziCards); i++ {
					resolved = append(resolved, domain.Card{Suit: laiziCards[i].Suit, Rank: target})
				}
				sortResolved(resolved)
				return newHand(rules.KindLaiziSubstituteBomb, "赖子替代炸弹", "AAAL", cards, resolved, target, 1, "", true, bombTier(rules.KindLaiziSubstituteBomb)), true
			}
		}
		if len(counts) == 1 && len(laiziCards) == 1 {
			for target, count := range counts {
				if count == 3 {
					resolved := append([]domain.Card(nil), normalCards...)
					resolved = append(resolved, domain.Card{Suit: laiziCards[0].Suit, Rank: target})
					sortResolved(resolved)
					return newHand(rules.KindLaiziSubstituteBomb, "赖子替代炸弹", "AAAL", cards, resolved, target, 1, "", true, bombTier(rules.KindLaiziSubstituteBomb)), true
				}
			}
		}
	}

	return ResolvedHand{}, false
}

func newHand(kind rules.Kind, label, pattern string, cards, resolved []domain.Card, mainRank domain.Rank, groupCount int, attachmentType string, usesLaizi bool, tier int) ResolvedHand {
	hand := ResolvedHand{
		Kind:           kind,
		Label:          label,
		Pattern:        pattern,
		Cards:          append([]domain.Card(nil), cards...),
		ResolvedCards:  append([]domain.Card(nil), resolved...),
		MainRank:       mainRank,
		GroupCount:     groupCount,
		Length:         len(cards),
		AttachmentType: attachmentType,
		UsesLaizi:      usesLaizi,
		IsBomb:         isBombKind(kind),
		BombTier:       tier,
		CompareKey:     mainRank.String(),
	}
	return hand
}

func handPriority(hand ResolvedHand) int {
	if hand.IsBomb {
		return hand.BombTier
	}
	return 100
}

func bombTier(kind rules.Kind) int {
	switch kind {
	case rules.KindRocket:
		return 1
	case rules.KindBombFivePlus:
		return 2
	case rules.KindPureLaiziBomb:
		return 3
	case rules.KindBombFour:
		return 4
	case rules.KindLaiziSubstituteBomb:
		return 5
	default:
		return 100
	}
}

func isBombKind(kind rules.Kind) bool {
	switch kind {
	case rules.KindRocket, rules.KindBombFivePlus, rules.KindPureLaiziBomb, rules.KindBombFour, rules.KindLaiziSubstituteBomb:
		return true
	default:
		return false
	}
}

func rankCounts(cards []domain.Card) map[domain.Rank]int {
	out := make(map[domain.Rank]int, len(cards))
	for _, card := range cards {
		out[card.Rank]++
	}
	return out
}

func sameRank(counts map[domain.Rank]int) (domain.Rank, bool) {
	if len(counts) != 1 {
		return 0, false
	}
	for rank := range counts {
		return rank, true
	}
	return 0, false
}

func sameRankCards(cards []domain.Card) bool {
	if len(cards) == 0 {
		return false
	}
	first := cards[0].Rank
	for _, card := range cards[1:] {
		if card.Rank != first {
			return false
		}
	}
	return true
}

func coreRank(counts map[domain.Rank]int, target int) (domain.Rank, bool) {
	for rank, count := range counts {
		if count == target {
			return rank, true
		}
	}
	return 0, false
}

func coreAndPair(counts map[domain.Rank]int) (domain.Rank, bool) {
	var triple domain.Rank
	hasTriple := false
	hasPair := false
	for rank, count := range counts {
		switch count {
		case 3:
			triple = rank
			hasTriple = true
		case 2:
			hasPair = true
		default:
			return 0, false
		}
	}
	return triple, hasTriple && hasPair
}

func hasExactCounts(counts map[domain.Rank]int, expected ...int) bool {
	values := make([]int, 0, len(counts))
	for _, count := range counts {
		values = append(values, count)
	}
	sort.Ints(values)
	sort.Ints(expected)
	if len(values) != len(expected) {
		return false
	}
	for i := range values {
		if values[i] != expected[i] {
			return false
		}
	}
	return true
}

func highestRank(cards []domain.Card) domain.Rank {
	maxRank := cards[0].Rank
	for _, card := range cards[1:] {
		if card.Rank > maxRank {
			maxRank = card.Rank
		}
	}
	return maxRank
}

func highestCountRank(counts map[domain.Rank]int, expected int) domain.Rank {
	var best domain.Rank
	for rank, count := range counts {
		if count == expected && rank > best {
			best = rank
		}
	}
	return best
}

func isStraight(cards []domain.Card, counts map[domain.Rank]int) bool {
	if len(cards) < 5 || len(counts) != len(cards) {
		return false
	}
	ranks := sortedRanks(counts)
	for _, rank := range ranks {
		if rank >= domain.Rank2 || rank.IsJoker() {
			return false
		}
	}
	return ranksConsecutive(ranks)
}

func isSerialPairs(cards []domain.Card, counts map[domain.Rank]int) bool {
	if len(cards) < 6 || len(cards)%2 != 0 {
		return false
	}
	ranks := sortedRanks(counts)
	if len(ranks) != len(cards)/2 {
		return false
	}
	for _, rank := range ranks {
		if rank >= domain.Rank2 || rank.IsJoker() || counts[rank] != 2 {
			return false
		}
	}
	return ranksConsecutive(ranks)
}

func analyzePlane(cards []domain.Card, counts map[domain.Rank]int) (ResolvedHand, bool) {
	if len(cards) < 6 {
		return ResolvedHand{}, false
	}
	tripRanks := make([]domain.Rank, 0, len(counts))
	for rank, count := range counts {
		if rank >= domain.Rank2 || rank.IsJoker() {
			continue
		}
		if count == 3 {
			tripRanks = append(tripRanks, rank)
		}
	}
	sort.Slice(tripRanks, func(i, j int) bool { return tripRanks[i] < tripRanks[j] })
	if len(tripRanks) < 2 || !ranksConsecutive(tripRanks) {
		return ResolvedHand{}, false
	}
	switch {
	case len(cards)%3 == 0 && len(tripRanks)*3 == len(cards):
		return newHand(rules.KindPlaneBase, "飞机不带", "AAA BBB...", cards, cards, tripRanks[len(tripRanks)-1], len(tripRanks), "", false, 0), true
	case len(cards)%4 == 0 && len(tripRanks) == len(cards)/4:
		for _, count := range counts {
			if count != 3 && count != 1 {
				return ResolvedHand{}, false
			}
		}
		return newHand(rules.KindPlaneWithSingles, "飞机带单", "AAA BBB + X + Y", cards, cards, tripRanks[len(tripRanks)-1], len(tripRanks), "solo", false, 0), true
	case len(cards)%5 == 0 && len(tripRanks) == len(cards)/5:
		pairCount := 0
		for _, count := range counts {
			if count == 2 {
				pairCount++
				continue
			}
			if count != 3 {
				return ResolvedHand{}, false
			}
		}
		if pairCount == len(tripRanks) {
			return newHand(rules.KindPlaneWithPairs, "飞机带对", "AAA BBB + BB + CC", cards, cards, tripRanks[len(tripRanks)-1], len(tripRanks), "pair", false, 0), true
		}
	}
	return ResolvedHand{}, false
}

func sortedRanks(counts map[domain.Rank]int) []domain.Rank {
	out := make([]domain.Rank, 0, len(counts))
	for rank := range counts {
		out = append(out, rank)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func ranksConsecutive(ranks []domain.Rank) bool {
	if len(ranks) == 0 {
		return false
	}
	for i := 1; i < len(ranks); i++ {
		if ranks[i] != ranks[i-1]+1 {
			return false
		}
	}
	return true
}

func isRocket(cards []domain.Card) bool {
	if len(cards) != 2 {
		return false
	}
	hasBlack := false
	hasRed := false
	for _, card := range cards {
		if card.Rank == domain.RankBlackJoker {
			hasBlack = true
		}
		if card.Rank == domain.RankRedJoker {
			hasRed = true
		}
	}
	return hasBlack && hasRed
}

func sortResolved(cards []domain.Card) {
	sort.Slice(cards, func(i, j int) bool {
		if cards[i].Rank == cards[j].Rank {
			return cards[i].Suit < cards[j].Suit
		}
		return cards[i].Rank < cards[j].Rank
	})
}
