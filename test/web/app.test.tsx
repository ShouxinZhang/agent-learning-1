import "@testing-library/jest-dom/vitest";
import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import App from "../../web/src/App";
import type { GameState, RulesCatalog } from "../../web/src/types";

afterEach(() => {
  cleanup();
});

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
    expect(await screen.findByText("牌型规则")).toBeInTheDocument();
    expect(screen.getByText("三带一")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "叫地主" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "不叫" })).toBeInTheDocument();

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
});
