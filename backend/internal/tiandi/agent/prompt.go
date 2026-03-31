package agent

import (
	"encoding/json"
	"fmt"
	"strings"

	"agent-dou-dizhu/internal/tiandi/demo"
	"agent-dou-dizhu/internal/tiandi/rules"
)

func BuildPrompt(state demo.StateResponse, catalog rules.Catalog) PromptResponse {
	return BuildPromptWithContext(state, catalog, PromptContext{
		Seat:        state.CurrentActor,
		SeatRole:    seatRole(state, state.CurrentActor),
		CardCounter: zeroCardCounter(state.CurrentActor, seatRole(state, state.CurrentActor)),
		RoundMemory: zeroRoundMemory(state.CurrentActor, seatRole(state, state.CurrentActor)),
	})
}

func BuildPromptWithContext(state demo.StateResponse, catalog rules.Catalog, ctx PromptContext) PromptResponse {
	seat := ctx.Seat
	if strings.TrimSpace(seat) == "" {
		seat = state.CurrentActor
	}
	if strings.TrimSpace(ctx.SeatRole) == "" {
		ctx.SeatRole = seatRole(state, seat)
	}
	if ctx.RoundMemory.Seat == "" {
		ctx.RoundMemory = zeroRoundMemory(seat, ctx.SeatRole)
	}
	if ctx.CardCounter.Seat == "" {
		ctx.CardCounter = zeroCardCounter(seat, ctx.SeatRole)
	}

	currentPlayer := findPlayerBySeat(state, seat)
	hand := make([]PromptCard, 0, len(currentPlayer.Cards))
	handIDs := make([]string, 0, len(currentPlayer.Cards))
	for _, card := range currentPlayer.Cards {
		hand = append(hand, PromptCard{
			ID:      card.ID,
			Label:   card.Label,
			Rank:    card.Rank,
			Suit:    card.Suit,
			IsLaizi: card.IsLaizi,
		})
		handIDs = append(handIDs, card.ID)
	}

	return PromptResponse{
		ModelHint:        DefaultModelHint,
		State:            state,
		Phase:            state.Phase,
		CurrentActor:     state.CurrentActor,
		AvailableActions: append([]string{}, state.AvailableActions...),
		PlayerSeat:       seat,
		PlayerRole:       ctx.SeatRole,
		PlayerHand:       hand,
		CardCounter:      ctx.CardCounter,
		RoundMemory:      ctx.RoundMemory,
		SystemPrompt:     buildSystemPrompt(),
		UserPrompt:       buildUserPrompt(state, catalog, handIDs, ctx),
		ActionSchema: ActionSchema{
			Format:   "json",
			Required: []string{"seat", "kind", "cards", "reason"},
			Properties: map[string]string{
				"seat":         "must equal currentActor",
				"kind":         "must be one of availableActions",
				"cards":        "card id array; required when kind=play, otherwise empty array",
				"resolutionId": "required when resolutionCandidates are present and you choose one concrete resolution",
				"reason":       "short Chinese explanation for trace log",
			},
			Example: Decision{
				Seat:         state.CurrentActor,
				Kind:         firstAction(state.AvailableActions),
				Cards:        exampleCards(state),
				ResolutionID: exampleResolutionID(state),
				Reason:       "选择当前局面中的保守合法动作。",
			},
		},
	}
}

func buildSystemPrompt() string {
	return strings.TrimSpace(`你是“天地癞子斗地主测试代理”。
你的唯一任务是根据服务端给出的局面选择一个合法动作。
你会同时拿到完整当前手牌、记牌器摘要、以及仅保留当前轮的队友/对手出牌记忆。
服务端规则是唯一事实源；你不能发明规则，也不能使用未在 availableActions 中出现的动作。
seat 必须等于 currentActor。
如果选择出牌，cards 只能填写当前玩家手牌中的牌 id。
如果存在 resolutionCandidates 且你的选择依赖某个候选解析，必须回填对应的 resolutionId。
如果桌面已有主牌且 availableActions 包含 pass，你可以选择 pass；否则禁止不出。
你必须只输出一个 JSON 对象，不要输出 Markdown、代码块或额外解释。`)
}

func buildUserPrompt(state demo.StateResponse, catalog rules.Catalog, handIDs []string, ctx PromptContext) string {
	var lines []string
	lines = append(lines,
		fmt.Sprintf("phase=%s", state.Phase),
		fmt.Sprintf("currentActor=%s", state.CurrentActor),
		fmt.Sprintf("playerSeat=%s", ctx.Seat),
		fmt.Sprintf("playerRole=%s", ctx.SeatRole),
		fmt.Sprintf("availableActions=%s", mustJSON(state.AvailableActions)),
		fmt.Sprintf("message=%s", state.Message),
	)

	if state.CurrentTrick != nil {
		lines = append(lines,
			fmt.Sprintf("currentTrick.leadingSeat=%s", state.CurrentTrick.LeadingSeat),
			fmt.Sprintf("currentTrick.lastPlaySeat=%s", state.CurrentTrick.LastPlaySeat),
			fmt.Sprintf("currentTrick.passCount=%d", state.CurrentTrick.PassCount),
			fmt.Sprintf("currentTrick.cards=%s", mustJSON(cardIDs(state.CurrentTrick.Cards))),
			fmt.Sprintf("currentTrick.resolved=%s/%s", state.CurrentTrick.ResolvedHand.Label, state.CurrentTrick.ResolvedHand.MainRank),
		)
	}

	if state.ResolvedHand != nil {
		lines = append(lines, fmt.Sprintf("lastResolved=%s/%s", state.ResolvedHand.Label, state.ResolvedHand.MainRank))
	}
	if state.PlayError != "" {
		lines = append(lines, fmt.Sprintf("lastPlayError=%s", state.PlayError))
	}

	lines = append(lines, fmt.Sprintf("landlord=%s", state.Landlord))
	lines = append(lines, fmt.Sprintf("multiplier=%d", state.Multiplier))
	lines = append(lines, fmt.Sprintf("laizi=%s", mustJSON(state.Laizi)))
	lines = append(lines, fmt.Sprintf("playerHand=%s", mustJSON(handIDs)))
	lines = append(lines, fmt.Sprintf("playerHandCount=%d", len(handIDs)))
	lines = append(lines, fmt.Sprintf("players=%s", mustJSON(playerSummaries(state))))
	lines = append(lines, fmt.Sprintf("cardCounter=%s", mustJSON(ctx.CardCounter)))
	lines = append(lines, fmt.Sprintf("roundMemory=%s", mustJSON(ctx.RoundMemory)))
	lines = append(lines, "memoryContract=roundMemory 只保留当前轮有效信息；清轮后会重置为空结构。")

	if len(state.ResolutionCandidates) > 0 {
		lines = append(lines, fmt.Sprintf("resolutionCandidates=%s", mustJSON(state.ResolutionCandidates)))
	}

	ruleSummary := []string{}
	ruleSummary = append(ruleSummary, take(catalog.Notes, 2)...)
	ruleSummary = append(ruleSummary, take(catalog.ComparisonNotes, 2)...)
	ruleSummary = append(ruleSummary, take(catalog.LaiziResolutionNotes, 2)...)
	if len(ruleSummary) > 0 {
		lines = append(lines, fmt.Sprintf("ruleSummary=%s", mustJSON(ruleSummary)))
	}

	lines = append(lines,
		`outputContract={"seat":"currentActor","kind":"one of availableActions","cards":["card id"],"resolutionId":"optional candidate id","reason":"short chinese reason"}`,
		"仅返回 JSON。cards 在 kind=play 时必须非空，其他动作请返回空数组。",
	)

	return strings.Join(lines, "\n")
}

func findCurrentPlayer(state demo.StateResponse) demo.PlayerView {
	return findPlayerBySeat(state, state.CurrentActor)
}

func findPlayerBySeat(state demo.StateResponse, seat string) demo.PlayerView {
	for _, player := range state.Players {
		if player.Seat == seat {
			return player
		}
	}
	return demo.PlayerView{}
}

func cardIDs(cards []demo.CardView) []string {
	ids := make([]string, 0, len(cards))
	for _, card := range cards {
		ids = append(ids, card.ID)
	}
	return ids
}

func mustJSON(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "[]"
	}
	return string(data)
}

func take(values []string, limit int) []string {
	if limit >= len(values) {
		return append([]string(nil), values...)
	}
	return append([]string(nil), values[:limit]...)
}

func firstAction(actions []string) string {
	if len(actions) == 0 {
		return ""
	}
	return actions[0]
}

func exampleCards(state demo.StateResponse) []string {
	if !contains(state.AvailableActions, "play") {
		return []string{}
	}
	player := findCurrentPlayer(state)
	if len(player.Cards) == 0 {
		return []string{}
	}
	return []string{player.Cards[0].ID}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func playerSummaries(state demo.StateResponse) []map[string]any {
	summaries := make([]map[string]any, 0, len(state.Players))
	for _, player := range state.Players {
		summaries = append(summaries, map[string]any{
			"seat":       player.Seat,
			"isLandlord": player.IsLandlord,
			"isCurrent":  player.IsCurrent,
			"cardCount":  len(player.Cards),
		})
	}
	return summaries
}

func exampleResolutionID(state demo.StateResponse) string {
	if len(state.ResolutionCandidates) == 0 {
		return ""
	}
	return state.ResolutionCandidates[0].ID
}

func zeroRoundMemory(seat string, role string) SeatRoundMemory {
	return SeatRoundMemory{
		Seat:       seat,
		SeatRole:   role,
		RoundIndex: 1,
		TrickIndex: 0,
	}
}

func zeroCardCounter(seat string, role string) CardCounter {
	return CardCounter{
		PlayedCardsBySeat:    map[string][]string{},
		PlayedRankCounts:     map[string]int{},
		BombSignals:          []string{},
		Seat:                 seat,
		SeatRole:             role,
		RemainingUnknown:     54,
		TotalPlayedCardCount: 0,
	}
}

func seatRole(state demo.StateResponse, seat string) string {
	if seat == "" {
		return ""
	}
	if state.Landlord == "" {
		return "unknown"
	}
	if seat == state.Landlord {
		return "landlord"
	}
	return "farmer"
}
