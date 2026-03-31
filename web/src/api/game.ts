import type { GameState } from "../types";

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "";

type ActionPayload = {
  seat: string;
  kind: string;
  cards?: string[];
  resolutionId?: string;
};

async function readJson(response: Response): Promise<GameState> {
  if (!response.ok) {
    throw new Error(`request failed: ${response.status}`);
  }
  return response.json() as Promise<GameState>;
}

export function fetchGameState(): Promise<GameState> {
  return fetch(`${API_BASE_URL}/api/game/state`).then(readJson);
}

export function resetGame(): Promise<GameState> {
  return fetch(`${API_BASE_URL}/api/game/reset`, {
    method: "POST",
  }).then(readJson);
}

export function applyGameAction(payload: ActionPayload): Promise<GameState> {
  return fetch(`${API_BASE_URL}/api/game/action`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  }).then(readJson);
}
