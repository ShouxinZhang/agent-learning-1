import { useEffect, useState } from "react";
import {
  fetchAgentMatchTrace,
  fetchAgentPrompt,
  fetchAgentTrace,
  runMockAgent,
  runMockAgentMatch,
  runOpenRouterAgentMatch,
} from "./api/agent";
import { applyGameAction, fetchGameState, resetGame } from "./api/game";
import { fetchRulesCatalog } from "./api/rules";
import { ActionBar } from "./components/ActionBar";
import { AgentDebugPanel } from "./components/AgentDebugPanel";
import { BottomPanel } from "./components/BottomPanel";
import { CurrentTrickPanel } from "./components/CurrentTrickPanel";
import { PlayerPanel } from "./components/PlayerPanel";
import { RulesPanel } from "./components/RulesPanel";
import type { AgentMatchTrace, AgentPrompt, AgentTrace, GameState, RulesCatalog } from "./types";

type RulesState =
  | { status: "loading" }
  | { status: "error"; message: string }
  | { status: "ready"; data: RulesCatalog };

type LoadState =
  | { status: "loading" }
  | { status: "error"; message: string }
  | { status: "ready"; data: GameState; busy: boolean; rules: RulesState };

type AgentState = {
  prompt?: AgentPrompt;
  trace?: AgentTrace | null;
  match?: AgentMatchTrace | null;
  error?: string;
  busy: boolean;
};

function messageFromError(error: unknown) {
  return error instanceof Error ? error.message : "unknown error";
}

function combineAgentErrors(results: PromiseSettledResult<unknown>[]) {
  const messages = results
    .filter((result): result is PromiseRejectedResult => result.status === "rejected")
    .map((result) => messageFromError(result.reason));

  return messages.length > 0 ? [...new Set(messages)].join(" / ") : undefined;
}

export default function App() {
  const [state, setState] = useState<LoadState>({ status: "loading" });
  const [selectedCardIds, setSelectedCardIds] = useState<string[]>([]);
  const [agentState, setAgentState] = useState<AgentState>({ busy: false });

  const refreshAgentDebug = async () => {
    setAgentState((current) => ({ ...current, busy: true }));

    const [promptResult, traceResult, matchResult] = await Promise.allSettled([
      fetchAgentPrompt(),
      fetchAgentTrace(),
      fetchAgentMatchTrace(),
    ]);
    setAgentState({
      prompt: promptResult.status === "fulfilled" ? promptResult.value : undefined,
      trace: traceResult.status === "fulfilled" ? traceResult.value : undefined,
      match: matchResult.status === "fulfilled" ? matchResult.value : undefined,
      error: combineAgentErrors([promptResult, traceResult, matchResult]),
      busy: false,
    });
  };

  useEffect(() => {
    let cancelled = false;

    Promise.allSettled([fetchGameState(), fetchRulesCatalog(), fetchAgentPrompt(), fetchAgentTrace(), fetchAgentMatchTrace()]).then((results) => {
      if (cancelled) {
        return;
      }

      const [gameResult, rulesResult, promptResult, traceResult, matchResult] = results;
      if (gameResult.status === "rejected") {
        setState({
          status: "error",
          message: messageFromError(gameResult.reason),
        });
        return;
      }

      const rules: RulesState =
        rulesResult.status === "fulfilled"
          ? { status: "ready", data: rulesResult.value }
          : { status: "error", message: messageFromError(rulesResult.reason) };

      setState({ status: "ready", data: gameResult.value, busy: false, rules });
      setSelectedCardIds([]);
      setAgentState({
        prompt: promptResult.status === "fulfilled" ? promptResult.value : undefined,
        trace: traceResult.status === "fulfilled" ? traceResult.value : undefined,
        match: matchResult.status === "fulfilled" ? matchResult.value : undefined,
        error: combineAgentErrors([promptResult, traceResult, matchResult]),
        busy: false,
      });
    });

    return () => {
      cancelled = true;
    };
  }, []);

  const run = async (job: () => Promise<GameState>) => {
    if (state.status !== "ready") {
      return;
    }

    setState({ ...state, busy: true });
    try {
      const data = await job();
      setState({ status: "ready", data, busy: false, rules: state.rules });
      setSelectedCardIds([]);
      void refreshAgentDebug();
    } catch (error) {
      setState({
        status: "error",
        message: messageFromError(error),
      });
    }
  };

  const handleRunMockAgent = async () => {
    if (state.status !== "ready") {
      return;
    }

    const nextRules = state.rules;
    setAgentState((current) => ({ ...current, busy: true, error: undefined }));
    try {
      const result = await runMockAgent();
      setState({ status: "ready", data: result.state, busy: false, rules: nextRules });
      setSelectedCardIds([]);
      setAgentState((current) => ({
        ...current,
        trace: result.trace,
        busy: false,
      }));
      void refreshAgentDebug();
    } catch (error) {
      setAgentState((current) => ({
        ...current,
        busy: false,
        error: messageFromError(error),
      }));
    }
  };

  const handleRunMockMatch = async () => {
    if (state.status !== "ready") {
      return;
    }

    const nextRules = state.rules;
    setAgentState((current) => ({ ...current, busy: true, error: undefined }));
    try {
      const result = await runMockAgentMatch();
      setState({ status: "ready", data: result.state, busy: false, rules: nextRules });
      setSelectedCardIds([]);
      setAgentState((current) => ({
        ...current,
        match: result.match,
        busy: false,
      }));
      void refreshAgentDebug();
    } catch (error) {
      setAgentState((current) => ({
        ...current,
        busy: false,
        error: messageFromError(error),
      }));
    }
  };

  const handleRunLLMMatch = async () => {
    if (state.status !== "ready") {
      return;
    }

    const nextRules = state.rules;
    setAgentState((current) => ({ ...current, busy: true, error: undefined }));
    try {
      const result = await runOpenRouterAgentMatch();
      setState({ status: "ready", data: result.state, busy: false, rules: nextRules });
      setSelectedCardIds([]);
      setAgentState((current) => ({
        ...current,
        match: result.match,
        busy: false,
      }));
      void refreshAgentDebug();
    } catch (error) {
      setAgentState((current) => ({
        ...current,
        busy: false,
        error: messageFromError(error),
      }));
    }
  };

  if (state.status === "loading") {
    return <main className="app-shell"><div className="status-card">加载 demo 中...</div></main>;
  }

  if (state.status === "error") {
    return <main className="app-shell"><div className="status-card">请求失败：{state.message}</div></main>;
  }

  const { data, busy, rules } = state;
  const currentPlayer = data.players.find((player) => player.seat === data.currentActor);
  const selectedCards = currentPlayer?.cards.filter((card) => selectedCardIds.includes(card.id)) ?? [];
  const toggleCard = (id: string) => {
    if (busy) {
      return;
    }
    setSelectedCardIds((current) =>
      current.includes(id) ? current.filter((item) => item !== id) : [...current, id],
    );
  };

  return (
    <main className="app-shell">
      <div className="app-layout">
        <section className="app-main">
          <header className="topbar">
            <div className="metric">
              <span>阶段</span>
              <strong>{data.phase}</strong>
            </div>
            <div className="metric">
              <span>天赖子</span>
              <strong>{data.laizi.tian}</strong>
            </div>
            <div className="metric">
              <span>地赖子</span>
              <strong>{data.laizi.di}</strong>
            </div>
            <div className="metric">
              <span>地主</span>
              <strong>{data.landlord}</strong>
            </div>
            <div className="metric">
              <span>倍数</span>
              <strong>x{data.multiplier}</strong>
            </div>
          </header>

          {data.testMode?.enabled ? (
            <section className="status-inline">
              <strong>{data.testMode.label}</strong>
            </section>
          ) : null}

          <ActionBar
            currentActor={data.currentActor}
            actions={data.availableActions}
            busy={busy}
            selectedCount={selectedCardIds.length}
            idleMessage={data.testMode?.enabled ? "测试模式下当前无可执行动作" : "当前无可执行动作"}
            onReset={() => {
              void run(resetGame);
            }}
            onClearSelection={() => {
              setSelectedCardIds([]);
            }}
            onAction={(kind) => {
              const payload =
                kind === "play"
                  ? {
                      seat: data.currentActor,
                      kind,
                      cards: selectedCards.map((card) => card.id),
                    }
                  : {
                      seat: data.currentActor,
                      kind,
                    };
              void run(() => applyGameAction(payload));
            }}
          />

          <section className="status-inline">
            <strong>{data.message}</strong>
          </section>

          <CurrentTrickPanel
            trick={data.currentTrick}
            resolvedHand={data.resolvedHand}
            resolutionCandidates={data.resolutionCandidates}
            playError={data.playError}
            winner={data.winner}
          />

          <BottomPanel visible={data.bottom.visible} count={data.bottom.count} cards={data.bottom.cards} />

          <section className="players-grid">
            {data.players.map((player) => (
              <PlayerPanel
                key={player.seat}
                player={player}
                selectable={player.seat === data.currentActor && data.availableActions.includes("play")}
                selectedIds={selectedCardIds}
                onToggle={toggleCard}
              />
            ))}
          </section>
        </section>

        <aside className="app-help">
          <AgentDebugPanel
            prompt={agentState.prompt}
            trace={agentState.trace ?? null}
            match={agentState.match ?? null}
            error={agentState.error}
            busy={busy || agentState.busy}
            onRefresh={() => {
              void refreshAgentDebug();
            }}
            onRunMock={() => {
              void handleRunMockAgent();
            }}
            onRunMockMatch={() => {
              void handleRunMockMatch();
            }}
            onRunLLMMatch={() => {
              void handleRunLLMMatch();
            }}
          />
          {rules.status === "ready" ? (
            <RulesPanel catalog={rules.data} />
          ) : (
            <section className="rules-panel">
              <div className="rules-panel-header">
                <div>
                  <h2>帮助说明</h2>
                  <p>{rules.status === "loading" ? "规则加载中..." : `规则加载失败：${rules.message}`}</p>
                </div>
              </div>
            </section>
          )}
        </aside>
      </div>
    </main>
  );
}
