package rules

type Catalog struct {
	Version      string            `json:"version"`
	RankOrder    []string          `json:"rankOrder"`
	SequenceHigh string            `json:"sequenceHigh"`
	Notes        []string          `json:"notes"`
	Sections     []Section         `json:"sections"`
	BombPriority []BombPriority    `json:"bombPriority"`
	HandPriority []HandPriorityRef `json:"handPriority"`
}

type Section struct {
	Key   string     `json:"key"`
	Title string     `json:"title"`
	Items []HandRule `json:"items"`
}

type HandRule struct {
	Key         string   `json:"key"`
	Name        string   `json:"name"`
	Pattern     string   `json:"pattern"`
	Description string   `json:"description"`
	MinCards    int      `json:"minCards,omitempty"`
	Notes       []string `json:"notes,omitempty"`
}

type BombPriority struct {
	Rank        int      `json:"rank"`
	Key         string   `json:"key"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Notes       []string `json:"notes,omitempty"`
}

type HandPriorityRef struct {
	Rank int    `json:"rank"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

func CatalogData() Catalog {
	return Catalog{
		Version:      "2026-04-01",
		RankOrder:    []string{"3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A", "2", "BlackJoker", "RedJoker"},
		SequenceHigh: "A",
		Notes: []string{
			"牌型比较默认只比较主序列或关键牌。",
			"王不可参与顺子、连对、飞机的点数延展。",
			"当前规则目录用于前后端展示与后续判型实现，不直接等同于完整出牌引擎。",
		},
		Sections: []Section{
			{
				Key:   "basic",
				Title: "基础牌型",
				Items: []HandRule{
					{Key: "single", Name: "单张", Pattern: "A", Description: "1 张任意非王或王牌。", MinCards: 1},
					{Key: "pair", Name: "对子", Pattern: "AA", Description: "2 张同点数牌。", MinCards: 2},
					{Key: "triple", Name: "三张", Pattern: "AAA", Description: "3 张同点数牌。", MinCards: 3},
				},
			},
			{
				Key:   "combo",
				Title: "组合牌型",
				Items: []HandRule{
					{Key: "triple_with_single", Name: "三带一", Pattern: "AAA + B", Description: "三张同点数带 1 张单牌。", MinCards: 4},
					{Key: "triple_with_pair", Name: "三带二", Pattern: "AAA + BB", Description: "三张同点数带 1 对。", MinCards: 5},
					{Key: "four_with_two_pairs", Name: "四带二/两对", Pattern: "AAAA + BB + CC", Description: "四张同点数带两对。", MinCards: 8},
				},
			},
			{
				Key:   "sequence",
				Title: "连续牌型",
				Items: []HandRule{
					{
						Key:         "straight",
						Name:        "顺子",
						Pattern:     "ABCDE...",
						Description: "至少 5 张连续单牌，最高到 A。",
						MinCards:    5,
						Notes:       []string{"不允许包含大小王。"},
					},
					{
						Key:         "serial_pairs",
						Name:        "连对",
						Pattern:     "AA BB CC...",
						Description: "至少 3 组连续对子。",
						MinCards:    6,
						Notes:       []string{"不允许包含大小王。"},
					},
					{
						Key:         "plane_base",
						Name:        "飞机不带",
						Pattern:     "AAA BBB...",
						Description: "至少 2 组连续三张。",
						MinCards:    6,
						Notes:       []string{"可作为飞机带牌的主干。"},
					},
				},
			},
			{
				Key:   "plane",
				Title: "飞机扩展",
				Items: []HandRule{
					{
						Key:         "plane_with_singles",
						Name:        "飞机带两张单牌",
						Pattern:     "AAA BBB + X + Y",
						Description: "每组连续三张对应带 1 张单牌。",
						MinCards:    8,
					},
					{
						Key:         "plane_with_pairs",
						Name:        "飞机带一对",
						Pattern:     "AAA BBB + BB + CC",
						Description: "每组连续三张对应带 1 对。",
						MinCards:    10,
					},
				},
			},
			{
				Key:   "bomb",
				Title: "炸弹与特殊压制",
				Items: []HandRule{
					{Key: "bomb_four", Name: "四张炸弹", Pattern: "AAAA", Description: "4 张同点数牌。", MinCards: 4},
					{Key: "bomb_five_plus", Name: "5+ 炸弹", Pattern: "AAAAA...", Description: "5 张及以上同点数组合。", MinCards: 5},
					{
						Key:         "rocket",
						Name:        "王炸",
						Pattern:     "BlackJoker + RedJoker",
						Description: "大小王组合，最高压制牌型。",
						MinCards:    2,
					},
					{
						Key:         "pure_laizi_bomb",
						Name:        "纯赖子炸弹",
						Pattern:     "LLLL",
						Description: "4 张来自同一赖子点数的赖子牌，不允许混合天赖与地赖。",
						MinCards:    4,
					},
					{
						Key:         "laizi_substitute_bomb",
						Name:        "赖子替代炸弹",
						Pattern:     "AAAL",
						Description: "4 张炸弹中存在赖子替代。",
						MinCards:    4,
					},
				},
			},
		},
		BombPriority: []BombPriority{
			{
				Rank:        1,
				Key:         "rocket",
				Name:        "王炸",
				Description: "最高牌型，压过全部普通炸弹。",
			},
			{
				Rank:        2,
				Key:         "bomb_five_plus",
				Name:        "5+ 炸弹",
				Description: "五张及以上同点数组合，高于全部四张炸弹。",
			},
			{
				Rank:        3,
				Key:         "pure_laizi_bomb",
				Name:        "纯赖子炸弹",
				Description: "4 张同一赖子点数构成，不允许混合不同赖子。",
				Notes:       []string{"例如四张天赖或四张地赖。"},
			},
			{
				Rank:        4,
				Key:         "bomb_four",
				Name:        "实心无赖子炸弹",
				Description: "4 张无赖子的同点数炸弹。",
			},
			{
				Rank:        5,
				Key:         "laizi_substitute_bomb",
				Name:        "赖子替代炸弹",
				Description: "4 张炸弹中至少一张由赖子替代形成。",
			},
		},
		HandPriority: []HandPriorityRef{
			{Rank: 1, Key: "rocket", Name: "王炸"},
			{Rank: 2, Key: "bomb", Name: "炸弹"},
			{Rank: 3, Key: "plane", Name: "飞机"},
			{Rank: 4, Key: "straight", Name: "顺子"},
			{Rank: 5, Key: "serial_pairs", Name: "连对"},
			{Rank: 6, Key: "triple_with_pair", Name: "三带二"},
			{Rank: 7, Key: "triple_with_single", Name: "三带一"},
			{Rank: 8, Key: "triple", Name: "三张"},
			{Rank: 9, Key: "pair", Name: "对子"},
			{Rank: 10, Key: "single", Name: "单张"},
		},
	}
}
