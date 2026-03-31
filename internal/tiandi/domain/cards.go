package domain

type Suit string

const (
	Spade   Suit = "spade"
	Heart   Suit = "heart"
	Club    Suit = "club"
	Diamond Suit = "diamond"
	Joker   Suit = "joker"
)

type Rank int

const (
	Rank3 Rank = 3 + iota
	Rank4
	Rank5
	Rank6
	Rank7
	Rank8
	Rank9
	Rank10
	RankJ
	RankQ
	RankK
	RankA
	Rank2
	RankBlackJoker
	RankRedJoker
)

type Card struct {
	Suit Suit
	Rank Rank
}

func FullDeck() []Card {
	deck := make([]Card, 0, 54)
	normalSuits := []Suit{Spade, Heart, Club, Diamond}
	for _, rank := range NormalRanks() {
		for _, suit := range normalSuits {
			deck = append(deck, Card{Suit: suit, Rank: rank})
		}
	}
	deck = append(deck,
		Card{Suit: Joker, Rank: RankBlackJoker},
		Card{Suit: Joker, Rank: RankRedJoker},
	)
	return deck
}

func NormalRanks() []Rank {
	return []Rank{
		Rank3, Rank4, Rank5, Rank6, Rank7, Rank8, Rank9,
		Rank10, RankJ, RankQ, RankK, RankA, Rank2,
	}
}

func (r Rank) IsNormal() bool {
	return r >= Rank3 && r <= Rank2
}

func (r Rank) IsJoker() bool {
	return r == RankBlackJoker || r == RankRedJoker
}

func (r Rank) String() string {
	switch r {
	case Rank3:
		return "3"
	case Rank4:
		return "4"
	case Rank5:
		return "5"
	case Rank6:
		return "6"
	case Rank7:
		return "7"
	case Rank8:
		return "8"
	case Rank9:
		return "9"
	case Rank10:
		return "10"
	case RankJ:
		return "J"
	case RankQ:
		return "Q"
	case RankK:
		return "K"
	case RankA:
		return "A"
	case Rank2:
		return "2"
	case RankBlackJoker:
		return "BlackJoker"
	case RankRedJoker:
		return "RedJoker"
	default:
		return "UnknownRank"
	}
}

func (c Card) String() string {
	if c.Suit == Joker {
		return c.Rank.String()
	}
	return c.Suit.String() + "-" + c.Rank.String()
}

func (s Suit) String() string {
	return string(s)
}
