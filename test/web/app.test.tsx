import "@testing-library/jest-dom/vitest";
import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import App from "../../web/src/App";
import type { GameState, RulesCatalog } from "../../web/src/types";

afterEach(() => {
  cleanup();
});

function makeCards(count: number, prefix: string) {
  return Array.from({ length: count }, (_, index) => ({
    label: `${prefix}-${index}`,
    suit: "heart",
    rank: "A",
    isLaizi: false,
  }));
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
  players: [
    {
      seat: "P0",
      isLandlord: false,
      isCurrent: true,
      cards: [{ label: "heart-A", suit: "heart", rank: "A", isLaizi: false }],
    },
    {
      seat: "P1",
      isLandlord: false,
      isCurrent: false,
      cards: [{ label: "club-10", suit: "club", rank: "10", isLaizi: false }],
    },
    {
      seat: "P2",
      isLandlord: false,
      isCurrent: false,
      cards: [{ label: "spade-3", suit: "spade", rank: "3", isLaizi: false }],
    },
  ],
};

const playTestModeState: GameState = {
  phase: "PLAY",
  currentActor: "",
  availableActions: [],
  landlord: "P0",
  multiplier: 1,
  message: "测试模式：已直接进入 PLAY，地主固定为 P0",
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
  players: [
    {
      seat: "P0",
      isLandlord: true,
      isCurrent: false,
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
  it("loads game state and submits an action", async () => {
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

  it("renders PLAY test mode without action buttons", async () => {
    const fetchMock = vi
      .fn()
      .mockImplementationOnce(() => ok(playTestModeState))
      .mockImplementationOnce(() => ok(rulesCatalog));

    vi.stubGlobal("fetch", fetchMock);

    render(<App />);

    expect(await screen.findByText("当前为测试模式 / 固定地主为 P0 / 直接进入 PLAY")).toBeInTheDocument();
    expect(await screen.findByText("测试模式下当前无可执行动作")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "牌型规则" })).toBeInTheDocument();
    expect(screen.getByText("测试模式：已直接进入 PLAY，地主固定为 P0")).toBeInTheDocument();
    expect(screen.getByText("20 张")).toBeInTheDocument();
    expect(screen.getAllByText("地主")[0]).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "新开一局" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "叫地主" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "抢地主" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "不叫" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "不抢" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "我抢" })).not.toBeInTheDocument();
  });
});
