import "@testing-library/jest-dom/vitest";
import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import App from "../../web/src/App";
import type { GameState } from "../../web/src/types";

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

function ok(data: unknown) {
  return Promise.resolve({
    ok: true,
    json: async () => data,
  } as Response);
}

describe("App", () => {
  it("loads game state and submits an action", async () => {
    const fetchMock = vi
      .fn()
      .mockImplementationOnce(() => ok(bidState))
      .mockImplementationOnce(() => ok(qiangState));

    vi.stubGlobal("fetch", fetchMock);

    const user = userEvent.setup();
    render(<App />);

    expect(await screen.findByText("P0 进行叫地主")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "叫地主" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "不叫" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "叫地主" }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/game/action", expect.any(Object));
    });
    expect(await screen.findByText("P1 进行抢地主")).toBeInTheDocument();
  });
});
