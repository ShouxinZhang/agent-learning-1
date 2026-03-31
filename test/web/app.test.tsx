import "@testing-library/jest-dom/vitest";
import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import App from "../../web/src/App.tsx";
import type {
  AgentMatchEnvelope,
  AgentMatchRunResponse,
  AgentMatchTrace,
  AgentPrompt,
  AgentRunResponse,
  AgentTrace,
  AgentTraceEnvelope,
  CardCounter,
  GameState,
  SeatRoundMemory,
  RulesCatalog,
  ResolvedHandView,
} from "../../web/src/types.ts";

afterEach(() => {
  cleanup();
});

function makeCards(count: number, prefix: string) {
  return Array.from({ length: count }, (_, index) => ({
    id: `${prefix}-${index}`,
    label: `${prefix}-${index}`,
    suit: "heart",
    rank: "A",
    isLaizi: false,
  }));
}

function makeResolvedHand(label: string): ResolvedHandView {
  return {
    kind: "single",
    label,
    pattern: "A",
    mainRank: "A",
    length: 1,
    groupCount: 1,
    attachmentType: "",
    compareKey: "A",
    usesLaizi: false,
    isBomb: false,
    bombTier: 0,
    cards: makeCards(1, `${label}-cards`),
    resolvedCards: makeCards(1, `${label}-resolved`),
  };
}

function makeAgentPrompt(currentActor: string, actions: string[], handCount: number): AgentPrompt {
  return {
    modelHint: "qwen/qwen3.6-plus-preview:free",
    phase: "PLAY",
    currentActor,
    availableActions: actions,
    playerSeat: currentActor,
    playerRole: currentActor === "P0" ? "landlord" : "farmer",
    playerHand: makeCards(handCount, `${currentActor}-prompt`),
    cardCounter: {
      seat: currentActor,
      seatRole: currentActor === "P0" ? "landlord" : "farmer",
      playedCardsBySeat: {},
      playedRankCounts: {},
      remainingUnknown: 54 - handCount,
      blackJokerPlayed: false,
      redJokerPlayed: false,
      bombSignals: [],
      totalPlayedCardCount: 0,
    },
    roundMemory: {
      seat: currentActor,
      seatRole: currentActor === "P0" ? "landlord" : "farmer",
      roundIndex: 1,
      trickIndex: 0,
    },
    systemPrompt: "你必须只输出一个 JSON 对象。",
    userPrompt: `currentActor=${currentActor}\navailableActions=${JSON.stringify(actions)}`,
    actionSchema: {
      format: "json",
      required: ["seat", "kind", "cards", "reason"],
      properties: {
        seat: "must equal currentActor",
        kind: "must be one of availableActions",
        cards: "card id array",
        reason: "short Chinese explanation",
      },
      example: {
        seat: currentActor,
        kind: actions[0] ?? "",
        cards: handCount > 0 ? [`${currentActor}-prompt-0`] : [],
        reason: "保守合法动作",
      },
    },
  };
}

function makeAgentTrace(mode: string, currentActor: string, applied: boolean, resultMessage: string): AgentTrace {
  return {
    mode,
    model: "qwen/qwen3.6-plus-preview:free",
    prompt: makeAgentPrompt(currentActor, ["play"], 20),
    rawResponse: '{"seat":"P0","kind":"play","cards":["p0-0"],"reason":"test"}',
    decision: {
      seat: currentActor,
      kind: "play",
      cards: ["p0-0"],
      reason: "test",
    },
    applied,
    error: "",
    resultMessage,
  };
}

const emptyTrace: AgentTraceEnvelope = {
  trace: null,
};

const emptyMatch: AgentMatchEnvelope = {
  match: null,
};

function makeCardCounter(seat: string): CardCounter {
  return {
    seat,
    seatRole: seat === "P0" ? "landlord" : "farmer",
    playedCardsBySeat: { P0: ["p0-0"] },
    playedRankCounts: { A: 1 },
    remainingUnknown: 33,
    blackJokerPlayed: false,
    redJokerPlayed: false,
    bombSignals: [],
    totalPlayedCardCount: 1,
  };
}

function makeRoundMemory(seat: string): SeatRoundMemory {
  return {
    seat,
    seatRole: seat === "P0" ? "landlord" : "farmer",
    roundIndex: 1,
    trickIndex: 1,
    lastOpponentPlay: {
      seat: "P0",
      kind: "play",
      cards: ["p0-0"],
      resolvedLabel: "单张",
      relationship: "opponent",
    },
  };
}

function makeAgentMatchTrace(): AgentMatchTrace {
  return {
    matchId: "match-1",
    startedAt: "2026-04-01T00:00:00Z",
    finishedAt: "2026-04-01T00:01:00Z",
    status: "completed",
    mode: "mock",
    model: "qwen/qwen3.6-plus-preview:free",
    winner: "P0",
    stepCount: 3,
    finalState: winnerState,
    steps: [
      {
        stepIndex: 3,
        seat: "P2",
        attemptMode: "mock",
        effectiveMode: "mock",
        model: "mock/default",
        prompt: {
          ...makeAgentPrompt("P2", ["play", "pass"], 4),
          cardCounter: makeCardCounter("P2"),
          roundMemory: makeRoundMemory("P2"),
        },
        decision: {
          seat: "P2",
          kind: "pass",
          reason: "test",
        },
        applied: true,
        error: "",
        resultMessage: "P0 获胜，牌局结束",
        roundIndex: 1,
        trickIndex: 3,
        cardCounter: makeCardCounter("P2"),
        roundMemory: makeRoundMemory("P2"),
        stateBefore: afterPlayState,
        stateAfter: winnerState,
      },
    ],
  };
}

const bidState: GameState = {
  phase: "BID",
  currentActor: "P0",
  availableActions: ["jiaodizhu", "bujiao"],
  landlord: "",
  multiplier: 1,
  message: "P0 进行叫地主",
  laizi: {
    tian: "?",
    di: "Q",
    tianVisible: false,
    diVisible: true,
  },
  bottom: {
    visible: false,
    count: 3,
    cards: [],
  },
  resolutionCandidates: [],
  playError: "",
  winner: "",
  players: [
    {
      seat: "P0",
      isLandlord: false,
      isCurrent: true,
      cards: [{ id: "heart-A", label: "heart-A", suit: "heart", rank: "A", isLaizi: false }],
    },
    {
      seat: "P1",
      isLandlord: false,
      isCurrent: false,
      cards: [{ id: "club-10", label: "club-10", suit: "club", rank: "10", isLaizi: false }],
    },
    {
      seat: "P2",
      isLandlord: false,
      isCurrent: false,
      cards: [{ id: "spade-3", label: "spade-3", suit: "spade", rank: "3", isLaizi: false }],
    },
  ],
};

const playTestModeState: GameState = {
  phase: "PLAY",
  currentActor: "P0",
  availableActions: ["play"],
  landlord: "P0",
  multiplier: 1,
  message: "测试模式：已直接进入 PLAY，地主固定为 P0，可开始正式出牌",
  testMode: {
    enabled: true,
    label: "当前为测试模式 / 固定地主为 P0 / 直接进入 PLAY",
    fixedLandlord: "P0",
    directPlay: true,
  },
  laizi: {
    tian: "A",
    di: "Q",
    tianVisible: true,
    diVisible: true,
  },
  bottom: {
    visible: true,
    count: 3,
    cards: makeCards(3, "bottom"),
  },
  resolutionCandidates: [],
  playError: "",
  winner: "",
  players: [
    {
      seat: "P0",
      isLandlord: true,
      isCurrent: true,
      cards: makeCards(20, "p0"),
    },
    {
      seat: "P1",
      isLandlord: false,
      isCurrent: false,
      cards: makeCards(17, "p1"),
    },
    {
      seat: "P2",
      isLandlord: false,
      isCurrent: false,
      cards: makeCards(17, "p2"),
    },
  ],
};

const qiangState: GameState = {
  ...bidState,
  phase: "QIANGDIZHU",
  currentActor: "P1",
  availableActions: ["qiangdizhu", "buqiang"],
  message: "P1 进行抢地主",
  players: bidState.players.map((player) => ({
    ...player,
    isCurrent: player.seat === "P1",
  })),
};

const afterPlayState: GameState = {
  ...playTestModeState,
  currentActor: "P1",
  availableActions: ["play", "pass"],
  message: "P1 跟牌或不出",
  currentTrick: {
    leadingSeat: "P0",
    lastPlaySeat: "P0",
    passCount: 0,
    cards: [{ id: "p0-0", label: "p0-0", suit: "heart", rank: "A", isLaizi: false }],
    resolvedHand: makeResolvedHand("单张"),
  },
  resolvedHand: makeResolvedHand("单张"),
  resolutionCandidates: [
    {
      id: "candidate-standard",
      priority: 100,
      isPreferred: true,
      resolvedHand: makeResolvedHand("单张"),
    },
  ],
  players: [
    {
      seat: "P0",
      isLandlord: true,
      isCurrent: false,
      cards: makeCards(19, "p0"),
    },
    {
      seat: "P1",
      isLandlord: false,
      isCurrent: true,
      cards: makeCards(17, "p1"),
    },
    {
      seat: "P2",
      isLandlord: false,
      isCurrent: false,
      cards: makeCards(17, "p2"),
    },
  ],
};

const invalidPlayState: GameState = {
  ...playTestModeState,
  playError: "selected cards do not form a supported hand",
  resolutionCandidates: [],
  currentTrick: undefined,
  resolvedHand: undefined,
};

const winnerState: GameState = {
  ...afterPlayState,
  currentActor: "",
  availableActions: [],
  message: "P0 获胜，牌局结束",
  winner: "P0",
  players: [
    {
      seat: "P0",
      isLandlord: true,
      isCurrent: false,
      cards: [],
    },
    {
      seat: "P1",
      isLandlord: false,
      isCurrent: false,
      cards: makeCards(3, "p1"),
    },
    {
      seat: "P2",
      isLandlord: false,
      isCurrent: false,
      cards: makeCards(4, "p2"),
    },
  ],
};

const rulesCatalog: RulesCatalog = {
  version: "2026-04-01",
  rankOrder: ["3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A", "2", "BlackJoker", "RedJoker"],
  sequenceHigh: "A",
  notes: ["牌型比较默认只比较主序列或关键牌。"],
  comparisonNotes: ["飞机比较先看三连组数，再比较最高三张点数。"],
  laiziResolutionNotes: ["若存在多个同优或等价解释，测试与展示阶段应返回全部可能。"],
  sections: [
    {
      key: "combo",
      title: "组合牌型",
      items: [
        {
          key: "triple_with_single",
          name: "三带一",
          pattern: "AAA + B",
          description: "三张同点数带 1 张单牌。",
          minCards: 4,
        },
      ],
    },
  ],
  bombPriority: [
    {
      rank: 1,
      key: "rocket",
      name: "王炸",
      description: "最高牌型，压过全部普通炸弹。",
    },
  ],
  handPriority: [
    {
      rank: 1,
      key: "rocket",
      name: "王炸",
    },
  ],
};

function ok(data: unknown) {
  return Promise.resolve({
    ok: true,
    json: async () => data,
  } as Response);
}

function fail(status: number) {
  return Promise.resolve({
    ok: false,
    status,
    json: async () => ({}),
  } as Response);
}

describe("App", () => {
  it("loads game state, rules, and agent debug panel", async () => {
    const fetchMock = vi
      .fn()
      .mockImplementationOnce(() => ok(playTestModeState))
      .mockImplementationOnce(() => ok(rulesCatalog))
      .mockImplementationOnce(() => ok(makeAgentPrompt("P0", ["play"], 20)))
      .mockImplementationOnce(() => ok(emptyTrace))
      .mockImplementationOnce(() => ok(emptyMatch));

    vi.stubGlobal("fetch", fetchMock);

    render(<App />);

    expect(await screen.findByText("Agent 调试")).toBeInTheDocument();
    expect(screen.getByText("模型提示：qwen/qwen3.6-plus-preview:free")).toBeInTheDocument();
    expect(screen.getByText("当前还没有 agent 执行记录。")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "运行 Mock Agent" })).toBeInTheDocument();
  });

  it("loads game state and submits a pre-play action", async () => {
    const fetchMock = vi
      .fn()
      .mockImplementationOnce(() => ok(bidState))
      .mockImplementationOnce(() => ok(rulesCatalog))
      .mockImplementationOnce(() => ok(makeAgentPrompt("P0", ["jiaodizhu", "bujiao"], 1)))
      .mockImplementationOnce(() => ok(emptyTrace))
      .mockImplementationOnce(() => ok(emptyMatch))
      .mockImplementationOnce(() => ok(qiangState))
      .mockImplementationOnce(() => ok(makeAgentPrompt("P1", ["qiangdizhu", "buqiang"], 1)))
      .mockImplementationOnce(() => ok(emptyTrace))
      .mockImplementationOnce(() => ok(emptyMatch));

    vi.stubGlobal("fetch", fetchMock);

    const user = userEvent.setup();
    render(<App />);

    expect(await screen.findByText("P0 进行叫地主")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "叫地主" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "不叫" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "叫地主" }));

    await waitFor(() => {
      expect(fetchMock.mock.calls.some(([url]) => url === "/api/game/action")).toBe(true);
    });
    expect(await screen.findByText("P1 进行抢地主")).toBeInTheDocument();
  });

  it("keeps the game usable when rules and agent debug loading fail", async () => {
    const fetchMock = vi
      .fn()
      .mockImplementationOnce(() => ok(bidState))
      .mockImplementationOnce(() => fail(500))
      .mockImplementationOnce(() => fail(503))
      .mockImplementationOnce(() => fail(504))
      .mockImplementationOnce(() => fail(505));

    vi.stubGlobal("fetch", fetchMock);

    render(<App />);

    expect(await screen.findByText("P0 进行叫地主")).toBeInTheDocument();
    expect(await screen.findByText("规则加载失败：request failed: 500")).toBeInTheDocument();
    expect(screen.getByText(/Agent 调试数据加载失败：request failed: 503 \/ request failed: 504 \/ request failed: 505/)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "叫地主" })).toBeInTheDocument();
  });

  it("renders PLAY mode with selectable cards and play action", async () => {
    const fetchMock = vi
      .fn()
      .mockImplementationOnce(() => ok(playTestModeState))
      .mockImplementationOnce(() => ok(rulesCatalog))
      .mockImplementationOnce(() => ok(makeAgentPrompt("P0", ["play"], 20)))
      .mockImplementationOnce(() => ok(emptyTrace))
      .mockImplementationOnce(() => ok(emptyMatch));

    vi.stubGlobal("fetch", fetchMock);

    render(<App />);

    expect(await screen.findByText("当前为测试模式 / 固定地主为 P0 / 直接进入 PLAY")).toBeInTheDocument();
    expect(screen.getByText("测试模式：已直接进入 PLAY，地主固定为 P0，可开始正式出牌")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "出牌" })).toBeDisabled();
    expect(screen.queryByRole("button", { name: "不出" })).not.toBeInTheDocument();
    expect(screen.getByText("20 张")).toBeInTheDocument();
  });

  it("submits selected cards during PLAY and renders current trick", async () => {
    const fetchMock = vi
      .fn()
      .mockImplementationOnce(() => ok(playTestModeState))
      .mockImplementationOnce(() => ok(rulesCatalog))
      .mockImplementationOnce(() => ok(makeAgentPrompt("P0", ["play"], 20)))
      .mockImplementationOnce(() => ok(emptyTrace))
      .mockImplementationOnce(() => ok(emptyMatch))
      .mockImplementationOnce(() => ok(afterPlayState))
      .mockImplementationOnce(() => ok(makeAgentPrompt("P1", ["play", "pass"], 17)))
      .mockImplementationOnce(() => ok({ trace: makeAgentTrace("mock", "P0", true, "P1 跟牌或不出") }))
      .mockImplementationOnce(() => ok(emptyMatch));

    vi.stubGlobal("fetch", fetchMock);

    const user = userEvent.setup();
    render(<App />);

    expect(await screen.findByRole("button", { name: "出牌" })).toBeDisabled();
    await user.click(screen.getByRole("button", { name: "p0-0" }));
    expect(screen.getByRole("button", { name: "出牌" })).toBeEnabled();

    await user.click(screen.getByRole("button", { name: "出牌" }));

    await waitFor(() => {
      expect(fetchMock.mock.calls.some(([url]) => url === "/api/game/action")).toBe(true);
    });

    expect(await screen.findByText("当前桌面：P0 出了 单张")).toBeInTheDocument();
    expect(screen.getByText("最近解析：单张 / 主牌 A")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "不出" })).toBeInTheDocument();
  });

  it("keeps the table rendered and shows inline play error for invalid PLAY action", async () => {
    const fetchMock = vi
      .fn()
      .mockImplementationOnce(() => ok(playTestModeState))
      .mockImplementationOnce(() => ok(rulesCatalog))
      .mockImplementationOnce(() => ok(makeAgentPrompt("P0", ["play"], 20)))
      .mockImplementationOnce(() => ok(emptyTrace))
      .mockImplementationOnce(() => ok(emptyMatch))
      .mockImplementationOnce(() => ok(invalidPlayState))
      .mockImplementationOnce(() => ok(makeAgentPrompt("P0", ["play"], 20)))
      .mockImplementationOnce(() => ok(emptyTrace))
      .mockImplementationOnce(() => ok(emptyMatch));

    vi.stubGlobal("fetch", fetchMock);

    const user = userEvent.setup();
    render(<App />);

    await user.click(await screen.findByRole("button", { name: "p0-0" }));
    await user.click(screen.getByRole("button", { name: "出牌" }));

    expect(await screen.findByText("出牌失败：selected cards do not form a supported hand")).toBeInTheDocument();
    expect(screen.getByText("当前操作人：P0")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "出牌" })).toBeDisabled();
    expect(screen.queryByText(/^请求失败：/)).not.toBeInTheDocument();
  });

  it("runs mock agent and renders the latest trace", async () => {
    const runResponse: AgentRunResponse = {
      state: afterPlayState,
      trace: makeAgentTrace("mock", "P0", true, "P1 跟牌或不出"),
    };

    const fetchMock = vi
      .fn()
      .mockImplementationOnce(() => ok(playTestModeState))
      .mockImplementationOnce(() => ok(rulesCatalog))
      .mockImplementationOnce(() => ok(makeAgentPrompt("P0", ["play"], 20)))
      .mockImplementationOnce(() => ok(emptyTrace))
      .mockImplementationOnce(() => ok(emptyMatch))
      .mockImplementationOnce(() => ok(runResponse))
      .mockImplementationOnce(() => ok(makeAgentPrompt("P1", ["play", "pass"], 17)))
      .mockImplementationOnce(() => ok({ trace: runResponse.trace }))
      .mockImplementationOnce(() => ok(emptyMatch));

    vi.stubGlobal("fetch", fetchMock);

    const user = userEvent.setup();
    render(<App />);

    await user.click(await screen.findByRole("button", { name: "运行 Mock Agent" }));

    await waitFor(() => {
      expect(fetchMock.mock.calls.some(([url]) => url === "/api/game/agent/run")).toBe(true);
    });

    expect(await screen.findByText("结果：P1 跟牌或不出")).toBeInTheDocument();
    expect(screen.getByText("执行状态：已执行")).toBeInTheDocument();
    expect(screen.getByText("当前桌面：P0 出了 单张")).toBeInTheDocument();
  });

  it("runs mock match and renders winner plus card counter summary", async () => {
    const matchTrace = makeAgentMatchTrace();
    const matchResponse: AgentMatchRunResponse = {
      state: winnerState,
      match: matchTrace,
    };

    const fetchMock = vi
      .fn()
      .mockImplementationOnce(() => ok(playTestModeState))
      .mockImplementationOnce(() => ok(rulesCatalog))
      .mockImplementationOnce(() => ok(makeAgentPrompt("P0", ["play"], 20)))
      .mockImplementationOnce(() => ok(emptyTrace))
      .mockImplementationOnce(() => ok(emptyMatch))
      .mockImplementationOnce(() => ok(matchResponse))
      .mockImplementationOnce(() => ok(makeAgentPrompt("P0", [], 0)))
      .mockImplementationOnce(() => ok(emptyTrace))
      .mockImplementationOnce(() => ok({ match: matchTrace }));

    vi.stubGlobal("fetch", fetchMock);

    const user = userEvent.setup();
    render(<App />);

    await user.click(await screen.findByRole("button", { name: "运行整局 Mock 对局" }));

    await waitFor(() => {
      expect(fetchMock.mock.calls.some(([url]) => url === "/api/game/agent/match")).toBe(true);
    });

    expect(await screen.findByText("赢家：P0")).toBeInTheDocument();
    expect(screen.getByText("已出牌总数：1")).toBeInTheDocument();
    expect(screen.getByText("上一轮记忆")).toBeInTheDocument();
  });

  it("renders winner banner when the backend state is already finished", async () => {
    const fetchMock = vi
      .fn()
      .mockImplementationOnce(() => ok(winnerState))
      .mockImplementationOnce(() => ok(rulesCatalog))
      .mockImplementationOnce(() => ok(makeAgentPrompt("P0", [], 0)))
      .mockImplementationOnce(() => ok(emptyTrace))
      .mockImplementationOnce(() => ok(emptyMatch));

    vi.stubGlobal("fetch", fetchMock);

    render(<App />);

    expect(await screen.findByText("P0 已获胜")).toBeInTheDocument();
    expect(screen.getByText("P0 获胜，牌局结束")).toBeInTheDocument();
  });

  it("renders openrouter trace error without breaking the page", async () => {
    const failedTrace: AgentTraceEnvelope = {
      trace: {
        ...makeAgentTrace("openrouter", "P0", false, "测试模式：已直接进入 PLAY，地主固定为 P0，可开始正式出牌"),
        error: "OPENROUTER_API_KEY is not set",
      },
    };

    const fetchMock = vi
      .fn()
      .mockImplementationOnce(() => ok(playTestModeState))
      .mockImplementationOnce(() => ok(rulesCatalog))
      .mockImplementationOnce(() => ok(makeAgentPrompt("P0", ["play"], 20)))
      .mockImplementationOnce(() => ok(failedTrace))
      .mockImplementationOnce(() => ok(emptyMatch));

    vi.stubGlobal("fetch", fetchMock);

    render(<App />);

    expect(await screen.findByText("执行错误：OPENROUTER_API_KEY is not set")).toBeInTheDocument();
    expect(screen.getByText("测试模式：已直接进入 PLAY，地主固定为 P0，可开始正式出牌")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "出牌" })).toBeDisabled();
  });
});
