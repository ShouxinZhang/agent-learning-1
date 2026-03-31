import type { RulesCatalog } from "../types";

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "";

async function readJson(response: Response): Promise<RulesCatalog> {
  if (!response.ok) {
    throw new Error(`request failed: ${response.status}`);
  }
  return response.json() as Promise<RulesCatalog>;
}

export function fetchRulesCatalog(): Promise<RulesCatalog> {
  return fetch(`${API_BASE_URL}/api/game/rules`).then(readJson);
}
