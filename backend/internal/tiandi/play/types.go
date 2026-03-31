package play

import (
	"agent-dou-dizhu/internal/tiandi/domain"
	"agent-dou-dizhu/internal/tiandi/rules"
)

type ResolvedHand struct {
	Kind           rules.Kind    `json:"kind"`
	Label          string        `json:"label"`
	Pattern        string        `json:"pattern"`
	Cards          []domain.Card `json:"cards"`
	ResolvedCards  []domain.Card `json:"resolvedCards"`
	MainRank       domain.Rank   `json:"mainRank"`
	GroupCount     int           `json:"groupCount,omitempty"`
	Length         int           `json:"length"`
	AttachmentType string        `json:"attachmentType,omitempty"`
	UsesLaizi      bool          `json:"usesLaizi"`
	IsBomb         bool          `json:"isBomb"`
	BombTier       int           `json:"bombTier,omitempty"`
	CompareKey     string        `json:"compareKey,omitempty"`
}

type ResolvedCard = domain.Card

type CurrentTrick struct {
	LeadingSeat  domain.Seat   `json:"leadingSeat"`
	LastPlaySeat domain.Seat   `json:"lastPlaySeat"`
	Cards        []domain.Card `json:"cards"`
	ResolvedHand ResolvedHand  `json:"resolvedHand"`
	PassCount    int           `json:"passCount"`
}

type ResolutionCandidate struct {
	ID           string       `json:"id"`
	ResolvedHand ResolvedHand `json:"resolvedHand"`
	Priority     int          `json:"priority"`
	IsPreferred  bool         `json:"isPreferred"`
}

func CloneResolvedHand(in ResolvedHand) ResolvedHand {
	out := in
	out.Cards = append([]domain.Card(nil), in.Cards...)
	out.ResolvedCards = append([]domain.Card(nil), in.ResolvedCards...)
	return out
}

func CloneCandidates(in []ResolutionCandidate) []ResolutionCandidate {
	out := make([]ResolutionCandidate, 0, len(in))
	for _, candidate := range in {
		copyCandidate := candidate
		copyCandidate.ResolvedHand = CloneResolvedHand(candidate.ResolvedHand)
		out = append(out, copyCandidate)
	}
	return out
}
