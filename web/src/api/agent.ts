import type {
  AgentMatchEnvelope,
  AgentMatchRunResponse,
  AgentMatchTrace,
  AgentPrompt,
  AgentRunResponse,
  AgentTrace,
  AgentTraceEnvelope,
} from "../types";

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "";
const DEFAULT_FREE_MODEL = "qwen/qwen3.6-plus-preview:free";

async function readJson<T>(response: Response): Promise<T> {
  if (!response.ok) {
    throw new Error(`request failed: ${response.status}`);
  }
  return response.json() as Promise<T>;
}

export function fetchAgentPrompt(): Promise<AgentPrompt> {
  return fetch(`${API_BASE_URL}/api/game/agent/prompt`).then(readJson<AgentPrompt>);
}

export async function fetchAgentTrace(): Promise<AgentTrace | null> {
  const payload = await fetch(`${API_BASE_URL}/api/game/agent/trace`).then(readJson<AgentTraceEnvelope | AgentTrace>);
  if ("trace" in payload) {
    return payload.trace ?? null;
  }
  return payload;
}

async function readMatchEnvelope(url: string) {
  const payload = await fetch(url).then(readJson<AgentMatchEnvelope | AgentMatchTrace>);
  if ("match" in payload) {
    return payload.match ?? null;
  }
  return payload;
}

export async function fetchAgentMatchTrace(): Promise<AgentMatchTrace | null> {
  try {
    return await readMatchEnvelope(`${API_BASE_URL}/api/game/agent/match/trace`);
  } catch (error) {
    if (error instanceof Error && /404/.test(error.message)) {
      return readMatchEnvelope(`${API_BASE_URL}/api/game/agent/match`);
    }
    throw error;
  }
}

export function runMockAgent(): Promise<AgentRunResponse> {
  return fetch(`${API_BASE_URL}/api/game/agent/run`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ mode: "mock" }),
  }).then(readJson<AgentRunResponse>);
}

export function runMockAgentMatch(): Promise<AgentMatchRunResponse> {
  return fetch(`${API_BASE_URL}/api/game/agent/match`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ mode: "mock", resetGame: true }),
  }).then(readJson<AgentMatchRunResponse>);
}

export function runOpenRouterAgentMatch(): Promise<AgentMatchRunResponse> {
  return fetch(`${API_BASE_URL}/api/game/agent/match`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      mode: "openrouter",
      model: DEFAULT_FREE_MODEL,
      fallbackMode: "mock",
      resetGame: true,
    }),
  }).then(readJson<AgentMatchRunResponse>);
}
