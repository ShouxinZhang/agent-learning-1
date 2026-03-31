package rules

type Kind string

const (
	KindSingle            Kind = "single"
	KindPair              Kind = "pair"
	KindTriple            Kind = "triple"
	KindTripleWithSingle  Kind = "triple_with_single"
	KindTripleWithPair    Kind = "triple_with_pair"
	KindFourWithTwoPair   Kind = "four_with_two_pair"
	KindStraight          Kind = "straight"
	KindConsecutivePairs  Kind = "consecutive_pairs"
	KindAirplane          Kind = "airplane"
	KindAirplaneWithSolo  Kind = "airplane_with_solo"
	KindAirplaneWithPairs Kind = "airplane_with_pairs"
	KindBomb              Kind = "bomb"
	KindFivePlusBomb      Kind = "five_plus_bomb"
	KindRocket            Kind = "rocket"
	KindPureLaiziBomb     Kind = "pure_laizi_bomb"
	KindMixedLaiziBomb    Kind = "mixed_laizi_bomb"
)

type Category string

const (
	CategoryBasic    Category = "basic"
	CategorySequence Category = "sequence"
	CategoryBomb     Category = "bomb"
)

type Attachment string

const (
	AttachmentNone  Attachment = "none"
	AttachmentSolo  Attachment = "solo"
	AttachmentPair  Attachment = "pair"
	AttachmentPairs Attachment = "pairs"
)

type Definition struct {
	Kind              Kind       `json:"kind"`
	Name              string     `json:"name"`
	Category          Category   `json:"category"`
	MinCards          int        `json:"minCards"`
	MaxCards          int        `json:"maxCards"`
	Attachment        Attachment `json:"attachment"`
	SequenceGroupSize int        `json:"sequenceGroupSize"`
	MinSequenceGroups int        `json:"minSequenceGroups"`
	AllowsRank2       bool       `json:"allowsRank2"`
	AllowsJokers      bool       `json:"allowsJokers"`
	SupportsLaizi     bool       `json:"supportsLaizi"`
	Priority          int        `json:"priority"`
	Notes             []string   `json:"notes"`
}

type RuleBook struct {
	Patterns     []Definition `json:"patterns"`
	Priority     []Kind       `json:"priority"`
	BombPriority []Kind       `json:"bombPriority"`
}

func DefaultRuleBook() RuleBook {
	patterns := []Definition{
		{
			Kind:          KindSingle,
			Name:          "单张",
			Category:      CategoryBasic,
			MinCards:      1,
			MaxCards:      1,
			Attachment:    AttachmentNone,
			AllowsRank2:   true,
			AllowsJokers:  true,
			SupportsLaizi: true,
			Priority:      10,
		},
		{
			Kind:          KindPair,
			Name:          "对子",
			Category:      CategoryBasic,
			MinCards:      2,
			MaxCards:      2,
			Attachment:    AttachmentNone,
			AllowsRank2:   true,
			AllowsJokers:  false,
			SupportsLaizi: true,
			Priority:      20,
		},
		{
			Kind:          KindTriple,
			Name:          "三张",
			Category:      CategoryBasic,
			MinCards:      3,
			MaxCards:      3,
			Attachment:    AttachmentNone,
			AllowsRank2:   true,
			AllowsJokers:  false,
			SupportsLaizi: true,
			Priority:      30,
		},
		{
			Kind:          KindTripleWithSingle,
			Name:          "三带一",
			Category:      CategoryBasic,
			MinCards:      4,
			MaxCards:      4,
			Attachment:    AttachmentSolo,
			AllowsRank2:   true,
			AllowsJokers:  true,
			SupportsLaizi: true,
			Priority:      40,
		},
		{
			Kind:          KindTripleWithPair,
			Name:          "三带二",
			Category:      CategoryBasic,
			MinCards:      5,
			MaxCards:      5,
			Attachment:    AttachmentPair,
			AllowsRank2:   true,
			AllowsJokers:  false,
			SupportsLaizi: true,
			Priority:      50,
		},
		{
			Kind:          KindFourWithTwoPair,
			Name:          "四带二/两对",
			Category:      CategoryBasic,
			MinCards:      8,
			MaxCards:      8,
			Attachment:    AttachmentPairs,
			AllowsRank2:   true,
			AllowsJokers:  false,
			SupportsLaizi: true,
			Priority:      55,
		},
		{
			Kind:              KindStraight,
			Name:              "顺子",
			Category:          CategorySequence,
			MinCards:          5,
			MaxCards:          10,
			Attachment:        AttachmentNone,
			SequenceGroupSize: 1,
			MinSequenceGroups: 5,
			AllowsRank2:       false,
			AllowsJokers:      false,
			SupportsLaizi:     true,
			Priority:          60,
			Notes: []string{
				"顺子最少 5 张",
				"顺子最高到 A",
			},
		},
		{
			Kind:              KindConsecutivePairs,
			Name:              "连对",
			Category:          CategorySequence,
			MinCards:          6,
			MaxCards:          24,
			Attachment:        AttachmentNone,
			SequenceGroupSize: 2,
			MinSequenceGroups: 3,
			AllowsRank2:       false,
			AllowsJokers:      false,
			SupportsLaizi:     true,
			Priority:          70,
		},
		{
			Kind:              KindAirplane,
			Name:              "飞机不带",
			Category:          CategorySequence,
			MinCards:          6,
			MaxCards:          18,
			Attachment:        AttachmentNone,
			SequenceGroupSize: 3,
			MinSequenceGroups: 2,
			AllowsRank2:       false,
			AllowsJokers:      false,
			SupportsLaizi:     true,
			Priority:          80,
		},
		{
			Kind:              KindAirplaneWithSolo,
			Name:              "飞机带两单",
			Category:          CategorySequence,
			MinCards:          8,
			MaxCards:          20,
			Attachment:        AttachmentSolo,
			SequenceGroupSize: 3,
			MinSequenceGroups: 2,
			AllowsRank2:       false,
			AllowsJokers:      true,
			SupportsLaizi:     true,
			Priority:          85,
			Notes: []string{
				"每个三连带 1 张单牌",
			},
		},
		{
			Kind:              KindAirplaneWithPairs,
			Name:              "飞机带一对",
			Category:          CategorySequence,
			MinCards:          10,
			MaxCards:          24,
			Attachment:        AttachmentPair,
			SequenceGroupSize: 3,
			MinSequenceGroups: 2,
			AllowsRank2:       false,
			AllowsJokers:      false,
			SupportsLaizi:     true,
			Priority:          90,
			Notes: []string{
				"每个三连带 1 对",
			},
		},
		{
			Kind:          KindMixedLaiziBomb,
			Name:          "赖子替代炸弹",
			Category:      CategoryBomb,
			MinCards:      4,
			MaxCards:      4,
			Attachment:    AttachmentNone,
			AllowsRank2:   true,
			AllowsJokers:  false,
			SupportsLaizi: true,
			Priority:      110,
			Notes: []string{
				"四张同点数",
				"至少 1 张通过赖子替代形成",
			},
		},
		{
			Kind:          KindBomb,
			Name:          "实心无赖子炸弹",
			Category:      CategoryBomb,
			MinCards:      4,
			MaxCards:      4,
			Attachment:    AttachmentNone,
			AllowsRank2:   true,
			AllowsJokers:  false,
			SupportsLaizi: false,
			Priority:      120,
			Notes: []string{
				"四张同点数",
				"不含赖子",
			},
		},
		{
			Kind:          KindPureLaiziBomb,
			Name:          "纯赖子炸弹",
			Category:      CategoryBomb,
			MinCards:      4,
			MaxCards:      4,
			Attachment:    AttachmentNone,
			AllowsRank2:   true,
			AllowsJokers:  false,
			SupportsLaizi: true,
			Priority:      130,
			Notes: []string{
				"四张来自同一赖子点数",
				"不允许混合天赖与地赖",
			},
		},
		{
			Kind:          KindFivePlusBomb,
			Name:          "5+ 炸弹",
			Category:      CategoryBomb,
			MinCards:      5,
			MaxCards:      20,
			Attachment:    AttachmentNone,
			AllowsRank2:   true,
			AllowsJokers:  false,
			SupportsLaizi: true,
			Priority:      140,
			Notes: []string{
				"五张及以上同点数",
			},
		},
		{
			Kind:          KindRocket,
			Name:          "王炸",
			Category:      CategoryBomb,
			MinCards:      2,
			MaxCards:      2,
			Attachment:    AttachmentNone,
			AllowsRank2:   false,
			AllowsJokers:  true,
			SupportsLaizi: false,
			Priority:      150,
			Notes: []string{
				"BlackJoker + RedJoker",
			},
		},
	}

	priority := []Kind{
		KindRocket,
		KindFivePlusBomb,
		KindPureLaiziBomb,
		KindBomb,
		KindMixedLaiziBomb,
		KindAirplaneWithPairs,
		KindAirplaneWithSolo,
		KindAirplane,
		KindStraight,
		KindConsecutivePairs,
		KindFourWithTwoPair,
		KindTripleWithPair,
		KindTripleWithSingle,
		KindTriple,
		KindPair,
		KindSingle,
	}

	bombPriority := []Kind{
		KindRocket,
		KindFivePlusBomb,
		KindPureLaiziBomb,
		KindBomb,
		KindMixedLaiziBomb,
	}

	return RuleBook{
		Patterns:     cloneDefinitions(patterns),
		Priority:     append([]Kind(nil), priority...),
		BombPriority: append([]Kind(nil), bombPriority...),
	}
}

func Lookup(kind Kind) (Definition, bool) {
	for _, def := range DefaultRuleBook().Patterns {
		if def.Kind == kind {
			return def, true
		}
	}
	return Definition{}, false
}

func cloneDefinitions(in []Definition) []Definition {
	out := make([]Definition, 0, len(in))
	for _, def := range in {
		clone := def
		clone.Notes = append([]string(nil), def.Notes...)
		out = append(out, clone)
	}
	return out
}
