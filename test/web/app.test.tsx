import "@testing-library/jest-dom/vitest";
import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import App from "../../web/src/App";
import type { GameState, RulesCatalog, ResolvedHandView } from "../../web/src/types";

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
  it("loads game state and submits a pre-play action", async () => {
    const fetchMock = vi
      .fn()
      .mockImplementationOnce(() => ok(bidState))
      .mockImplementationOnce(() => ok(rulesCatalog))
      .mockImplementationOnce(() => ok(qiangState));

    vi.stubGlobal("fetch", fetchMock);

    const user = userEvent.setup();
    render(<App />);

    expect(await screen.findByText("P0 进行叫地主")).toBeInTheDocument();
    expect(await screen.findByText("帮助说明")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "牌型规则" })).toBeInTheDocument();
    expect(screen.getByText("三带一")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "叫地主" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "不叫" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "比较说明" }));
    expect(screen.getByText("飞机比较先看三连组数，再比较最高三张点数。")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "赖子说明" }));
    expect(screen.getByText("若存在多个同优或等价解释，测试与展示阶段应返回全部可能。")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "叫地主" }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenNthCalledWith(3, "/api/game/action", expect.any(Object));
    });
    expect(await screen.findByText("P1 进行抢地主")).toBeInTheDocument();
  });

  it("keeps the game usable when rules loading fails", async () => {
    const fetchMock = vi
      .fn()
      .mockImplementationOnce(() => ok(bidState))
      .mockImplementationOnce(() => fail(500));

    vi.stubGlobal("fetch", fetchMock);

    render(<App />);

    expect(await screen.findByText("P0 进行叫地主")).toBeInTheDocument();
    expect(await screen.findByText("规则加载失败：request failed: 500")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "叫地主" })).toBeInTheDocument();
  });

  it("renders PLAY mode with selectable cards and play action", async () => {
    const fetchMock = vi
      .fn()
      .mockImplementationOnce(() => ok(playTestModeState))
      .mockImplementationOnce(() => ok(rulesCatalog));

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
      .mockImplementationOnce(() => ok(afterPlayState));

    vi.stubGlobal("fetch", fetchMock);

    const user = userEvent.setup();
    render(<App />);

    expect(await screen.findByRole("button", { name: "出牌" })).toBeDisabled();
    await user.click(screen.getByRole("button", { name: "p0-0" }));
    expect(screen.getByRole("button", { name: "出牌" })).toBeEnabled();

    await user.click(screen.getByRole("button", { name: "出牌" }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenNthCalledWith(3, "/api/game/action", expect.objectContaining({ method: "POST" }));
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
      .mockImplementationOnce(() => ok(invalidPlayState));

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
});
