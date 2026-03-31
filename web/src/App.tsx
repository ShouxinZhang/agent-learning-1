import { useEffect, useState } from "react";
import { applyGameAction, fetchGameState, resetGame } from "./api/game";
import { fetchRulesCatalog } from "./api/rules";
import { ActionBar } from "./components/ActionBar";
import { BottomPanel } from "./components/BottomPanel";
import { PlayerPanel } from "./components/PlayerPanel";
import { RulesPanel } from "./components/RulesPanel";
import type { GameState, RulesCatalog } from "./types";

type RulesState =
  | { status: "loading" }
  | { status: "error"; message: string }
  | { status: "ready"; data: RulesCatalog };

type LoadState =
  | { status: "loading" }
  | { status: "error"; message: string }
  | { status: "ready"; data: GameState; busy: boolean; rules: RulesState };

export default function App() {
  const [state, setState] = useState<LoadState>({ status: "loading" });

  useEffect(() => {
    let cancelled = false;

    Promise.allSettled([fetchGameState(), fetchRulesCatalog()]).then((results) => {
      if (cancelled) {
        return;
      }

      const [gameResult, rulesResult] = results;
      if (gameResult.status === "rejected") {
        setState({
          status: "error",
          message: gameResult.reason instanceof Error ? gameResult.reason.message : "unknown error",
        });
        return;
      }

      const rules: RulesState =
        rulesResult.status === "fulfilled"
          ? { status: "ready", data: rulesResult.value }
          : {
              status: "error",
              message: rulesResult.reason instanceof Error ? rulesResult.reason.message : "unknown error",
            };

      setState({ status: "ready", data: gameResult.value, busy: false, rules });
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
    } catch (error) {
      setState({
        status: "error",
        message: error instanceof Error ? error.message : "unknown error",
      });
    }
  };

  if (state.status === "loading") {
    return <main className="app-shell"><div className="status-card">加载 demo 中...</div></main>;
  }

  if (state.status === "error") {
    return <main className="app-shell"><div className="status-card">请求失败：{state.message}</div></main>;
  }

  const { data, busy, rules } = state;

  return (
    <main className="app-shell">
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

      <ActionBar
        currentActor={data.currentActor}
        actions={data.availableActions}
        busy={busy}
        onReset={() => {
          void run(resetGame);
        }}
        onAction={(kind) => {
          void run(() =>
            applyGameAction({
              seat: data.currentActor,
              kind,
            }),
          );
        }}
      />

      <section className="status-inline">
        <strong>{data.message}</strong>
      </section>

      {rules.status === "ready" ? (
        <RulesPanel catalog={rules.data} />
      ) : (
        <section className="rules-panel">
          <div className="rules-panel-header">
            <div>
              <h2>牌型规则</h2>
              <p>{rules.status === "loading" ? "规则加载中..." : `规则加载失败：${rules.message}`}</p>
            </div>
          </div>
        </section>
      )}

      <BottomPanel visible={data.bottom.visible} count={data.bottom.count} cards={data.bottom.cards} />

      <section className="players-grid">
        {data.players.map((player) => (
          <PlayerPanel key={player.seat} player={player} />
        ))}
      </section>
    </main>
  );
}
