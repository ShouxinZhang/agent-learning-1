package agent

import "agent-dou-dizhu/internal/tiandi/demo"

const (
	DefaultModelHint  = "qwen/qwen3.6-plus-preview:free"
	RunModeMock       = "mock"
	RunModeOpenRouter = "openrouter"
)

type PromptCard struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	Rank    string `json:"rank"`
	Suit    string `json:"suit"`
	IsLaizi bool   `json:"isLaizi"`
}

type Decision struct {
	Seat         string   `json:"seat"`
	Kind         string   `json:"kind"`
	Cards        []string `json:"cards,omitempty"`
	ResolutionID string   `json:"resolutionId,omitempty"`
	Reason       string   `json:"reason,omitempty"`
}

type ActionSchema struct {
	Format     string            `json:"format"`
	Required   []string          `json:"required"`
	Properties map[string]string `json:"properties"`
	Example    Decision          `json:"example"`
}

type PromptContext struct {
	Seat        string
	SeatRole    string
	CardCounter CardCounter
	RoundMemory SeatRoundMemory
}

type PromptResponse struct {
	ModelHint        string             `json:"modelHint"`
	State            demo.StateResponse `json:"state"`
	Phase            string             `json:"phase"`
	CurrentActor     string             `json:"currentActor"`
	AvailableActions []string           `json:"availableActions"`
	PlayerSeat       string             `json:"playerSeat"`
	PlayerRole       string             `json:"playerRole"`
	PlayerHand       []PromptCard       `json:"playerHand"`
	CardCounter      CardCounter        `json:"cardCounter"`
	RoundMemory      SeatRoundMemory    `json:"roundMemory"`
	SystemPrompt     string             `json:"systemPrompt"`
	UserPrompt       string             `json:"userPrompt"`
	ActionSchema     ActionSchema       `json:"actionSchema"`
}

type Trace struct {
	RunID         string              `json:"runId"`
	CreatedAt     string              `json:"createdAt"`
	Mode          string              `json:"mode"`
	Model         string              `json:"model"`
	Prompt        PromptResponse      `json:"prompt"`
	RawResponse   string              `json:"rawResponse"`
	Decision      Decision            `json:"decision"`
	Applied       bool                `json:"applied"`
	Error         string              `json:"error"`
	ResultMessage string              `json:"resultMessage"`
	ResultState   *demo.StateResponse `json:"resultState,omitempty"`
}

type TraceEnvelope struct {
	Trace *Trace `json:"trace"`
}

type RunRequest struct {
	Mode  string `json:"mode"`
	Model string `json:"model,omitempty"`
}

type RunResponse struct {
	State demo.StateResponse `json:"state"`
	Trace Trace              `json:"trace"`
}

type PlayedAction struct {
	Seat          string   `json:"seat"`
	Kind          string   `json:"kind"`
	Cards         []string `json:"cards"`
	ResolvedLabel string   `json:"resolvedLabel,omitempty"`
	MainRank      string   `json:"mainRank,omitempty"`
	Relationship  string   `json:"relationship,omitempty"`
}

type SeatRoundMemory struct {
	Seat             string        `json:"seat,omitempty"`
	SeatRole         string        `json:"seatRole,omitempty"`
	RoundIndex       int           `json:"roundIndex"`
	TrickIndex       int           `json:"trickIndex"`
	LastSelfPlay     *PlayedAction `json:"lastSelfPlay,omitempty"`
	LastTeammatePlay *PlayedAction `json:"lastTeammatePlay,omitempty"`
	LastOpponentPlay *PlayedAction `json:"lastOpponentPlay,omitempty"`
}

type CardCounter struct {
	Seat                 string              `json:"seat,omitempty"`
	SeatRole             string              `json:"seatRole,omitempty"`
	PlayedCardsBySeat    map[string][]string `json:"playedCardsBySeat"`
	PlayedRankCounts     map[string]int      `json:"playedRankCounts"`
	RemainingUnknown     int                 `json:"remainingUnknown"`
	BlackJokerPlayed     bool                `json:"blackJokerPlayed"`
	RedJokerPlayed       bool                `json:"redJokerPlayed"`
	BombSignals          []string            `json:"bombSignals"`
	TotalPlayedCardCount int                 `json:"totalPlayedCardCount"`
}

type MatchStep struct {
	StepIndex     int                `json:"stepIndex"`
	Seat          string             `json:"seat"`
	AttemptMode   string             `json:"attemptMode"`
	EffectiveMode string             `json:"effectiveMode"`
	Model         string             `json:"model"`
	Prompt        PromptResponse     `json:"prompt"`
	Decision      Decision           `json:"decision"`
	Applied       bool               `json:"applied"`
	Error         string             `json:"error"`
	ResultMessage string             `json:"resultMessage"`
	RoundIndex    int                `json:"roundIndex"`
	TrickIndex    int                `json:"trickIndex"`
	CardCounter   CardCounter        `json:"cardCounter"`
	RoundMemory   SeatRoundMemory    `json:"roundMemory"`
	StateBefore   demo.StateResponse `json:"stateBefore"`
	StateAfter    demo.StateResponse `json:"stateAfter"`
}

type MatchTrace struct {
	MatchID      string              `json:"matchId"`
	StartedAt    string              `json:"startedAt"`
	FinishedAt   string              `json:"finishedAt,omitempty"`
	Status       string              `json:"status"`
	Mode         string              `json:"mode"`
	FallbackMode string              `json:"fallbackMode,omitempty"`
	Model        string              `json:"model"`
	Winner       string              `json:"winner"`
	Steps        []MatchStep         `json:"steps"`
	FinalState   *demo.StateResponse `json:"finalState,omitempty"`
	Error        string              `json:"error,omitempty"`
	StepCount    int                 `json:"stepCount"`
}

type MatchEnvelope struct {
	Match *MatchTrace `json:"match"`
}

type MatchRunRequest struct {
	Mode         string `json:"mode"`
	Model        string `json:"model,omitempty"`
	FallbackMode string `json:"fallbackMode,omitempty"`
	MaxSteps     int    `json:"maxSteps,omitempty"`
	ResetGame    bool   `json:"resetGame,omitempty"`
}

type MatchRunResponse struct {
	State demo.StateResponse `json:"state"`
	Match MatchTrace         `json:"match"`
}

type MatchStateResponse struct {
	State demo.StateResponse `json:"state"`
	Match *MatchTrace        `json:"match"`
}
