import type { AgentMatchTrace, AgentPrompt, AgentTrace, CardCounter, GamePlayer, GameState, PlayedAction, SeatRoundMemory } from "../types";

type AgentDebugPanelProps = {
  prompt?: AgentPrompt;
  trace?: AgentTrace | null;
  match?: AgentMatchTrace | null;
  error?: string;
  busy: boolean;
  onRefresh: () => void;
  onRunMock: () => void;
  onRunMockMatch: () => void;
  onRunLLMMatch: () => void;
};

function statusLabel(trace?: AgentTrace | null, match?: AgentMatchTrace | null, busy = false) {
  if (busy) {
    return { label: "执行中", className: "agent-badge agent-badge-running" };
  }
  if (match?.error) {
    return { label: "整局失败", className: "agent-badge agent-badge-failed" };
  }
  if (match?.winner || match?.status === "completed") {
    return { label: "整局完成", className: "agent-badge agent-badge-success" };
  }
  if (trace?.error) {
    return { label: "失败", className: "agent-badge agent-badge-failed" };
  }
  if (trace?.applied) {
    return { label: "成功", className: "agent-badge agent-badge-success" };
  }
  return { label: "空闲", className: "agent-badge agent-badge-idle" };
}

function latestMatchStep(match?: AgentMatchTrace | null) {
  const steps = Array.isArray(match?.steps) ? match.steps : [];
  if (steps.length === 0) {
    return undefined;
  }
  return steps[steps.length - 1];
}

function latestMatchState(match?: AgentMatchTrace | null): GameState | undefined {
  const lastStep = latestMatchStep(match);
  return match?.finalState ?? lastStep?.stateAfter ?? lastStep?.stateBefore;
}

function formatPlayedAction(action?: PlayedAction) {
  if (!action) {
    return "暂无";
  }
  const cards = (action.cards ?? []).length > 0 ? (action.cards ?? []).join(", ") : "无牌";
  const parts = [`${action.seat} ${action.kind}`, cards];
  if (action.resolvedLabel) {
    parts.push(action.resolvedLabel);
  }
  if (action.relationship) {
    parts.push(action.relationship);
  }
  return parts.join(" / ");
}

function topRankSummary(counter?: CardCounter) {
  if (!counter) {
    return "暂无";
  }
  const entries = Object.entries(counter.playedRankCounts).sort((left, right) => right[1] - left[1]);
  if (entries.length === 0) {
    return "暂无";
  }
  return entries
    .slice(0, 4)
    .map(([rank, count]) => `${rank}x${count}`)
    .join("、");
}

function seatRole(player: GamePlayer) {
  return player.isLandlord ? "地主" : "农民";
}

function winnerLabel(match?: AgentMatchTrace | null) {
  return match?.winner || match?.finalState?.winner || "未分出";
}

export function AgentDebugPanel({
  prompt,
  trace,
  match,
  error,
  busy,
  onRefresh,
  onRunMock,
  onRunMockMatch,
  onRunLLMMatch,
}: AgentDebugPanelProps) {
  const badge = statusLabel(trace, match, busy);
  const matchStep = latestMatchStep(match);
  const matchState = latestMatchState(match);
  const matchCounter = matchStep?.cardCounter;
  const matchMemory = matchStep?.roundMemory;
  const recentDecisionCards = matchStep?.decision.cards ?? [];
  const promptActions = prompt?.availableActions ?? [];
  const matchBombSignals = matchCounter?.bombSignals ?? [];
  const recentDecision = matchStep
    ? `${matchStep.seat} / ${matchStep.decision.kind}${recentDecisionCards.length > 0 ? ` / ${recentDecisionCards.join(", ")}` : ""}`
    : "暂无";

  return (
    <section className="agent-panel" aria-label="Agent 调试面板">
      <div className="agent-panel-header">
        <div>
          <h2>Agent 调试</h2>
          <p>展示后端 prompt、最近一次决策与执行结果。</p>
        </div>
        <span className={badge.className}>{badge.label}</span>
      </div>

      <div className="agent-toolbar">
        <button type="button" className="secondary-button" onClick={onRefresh} disabled={busy}>
          刷新调试数据
        </button>
        <button type="button" className="primary-button" onClick={onRunMock} disabled={busy || !prompt}>
          运行 Mock Agent
        </button>
        <button type="button" className="primary-button" onClick={onRunMockMatch} disabled={busy}>
          运行整局 Mock 对局
        </button>
        <button type="button" className="secondary-button" onClick={onRunLLMMatch} disabled={busy || !prompt}>
          运行整局 免费 LLM 对局
        </button>
      </div>

      {error ? <p className="error-copy">Agent 调试数据加载失败：{error}</p> : null}

      <section className="agent-section">
        <div className="agent-section-heading">
          <strong>整局观测</strong>
          <span>{match ? `${match.stepCount} steps` : "暂无记录"}</span>
        </div>
        {match ? (
          <>
            <p className="rule-note">整局状态：{match.status}</p>
            <p className="rule-note">赢家：{winnerLabel(match)}</p>
            <p className="rule-note">当前轮次：第 {matchStep?.roundIndex ?? 1} 轮 / 第 {matchStep?.trickIndex ?? 1} 手</p>
            <p className="rule-note">最近动作：{recentDecision}</p>
            <p className="rule-note">最近结果：{matchStep?.resultMessage || match.error || "暂无"}</p>
            {match.error ? <p className="error-copy">整局错误：{match.error}</p> : null}

            {matchState ? (
              <div className="agent-seat-grid">
                {matchState.players.map((player) => (
                  <article key={player.seat} className="agent-seat-card">
                    <strong>{player.seat}</strong>
                    <p>{seatRole(player)}</p>
                    <p>剩余手牌：{player.cards.length} 张</p>
                    <p>{player.isCurrent ? "当前行动位" : "等待中"}</p>
                  </article>
                ))}
              </div>
            ) : null}

            <div className="agent-meta-grid">
              <article className="agent-step">
                <strong>记牌器摘要</strong>
                <p>观察座位：{matchCounter?.seat ?? "N/A"}</p>
                <p>已出牌总数：{matchCounter?.totalPlayedCardCount ?? 0}</p>
                <p>未知剩余：{matchCounter?.remainingUnknown ?? 0}</p>
                <p>高频点数：{topRankSummary(matchCounter)}</p>
                <p>
                  王信息：
                  {matchCounter?.blackJokerPlayed ? " 黑王已出" : " 黑王未出"}
                  {" / "}
                  {matchCounter?.redJokerPlayed ? "红王已出" : "红王未出"}
                </p>
                <p>炸弹信号：{matchBombSignals.join(" / ") || "暂无"}</p>
              </article>

              <article className="agent-step">
                <strong>上一轮记忆</strong>
                <p>观察座位：{matchMemory?.seat ?? "N/A"} / {matchMemory?.seatRole ?? "未知角色"}</p>
                <p>self：{formatPlayedAction(matchMemory?.lastSelfPlay)}</p>
                <p>teammate：{formatPlayedAction(matchMemory?.lastTeammatePlay)}</p>
                <p>opponent：{formatPlayedAction(matchMemory?.lastOpponentPlay)}</p>
              </article>
            </div>

            <details className="agent-details">
              <summary>查看最近整局步骤</summary>
              <div className="agent-steps">
                {(match.steps ?? []).slice(-5).reverse().map((step) => {
                  const stepCards = step.decision.cards ?? [];
                  return (
                  <article key={`${step.stepIndex}-${step.seat}`} className="agent-step">
                    <strong>
                      Step {step.stepIndex} · {step.seat} · 第 {step.roundIndex} 轮 / 第 {step.trickIndex} 手
                    </strong>
                    <p>模式：{step.effectiveMode || step.attemptMode}</p>
                    <p>动作：{step.decision.kind}{stepCards.length > 0 ? ` / ${stepCards.join(", ")}` : ""}</p>
                    <p>结果：{step.resultMessage || step.error || "暂无"}</p>
                  </article>
                  );
                })}
              </div>
            </details>
          </>
        ) : (
          <p className="rule-note">当前还没有整局执行记录。</p>
        )}
      </section>

      <section className="agent-section">
        <div className="agent-section-heading">
          <strong>Prompt 摘要</strong>
          <span>{prompt?.currentActor ?? "N/A"}</span>
        </div>
        {prompt ? (
          <>
            <p className="rule-note">模型提示：{prompt.modelHint}</p>
            <p className="rule-note">阶段：{prompt.phase}</p>
            <p className="rule-note">动作：{promptActions.join(" / ") || "无"}</p>
            <p className="rule-note">当前手牌：{prompt.playerHand.length} 张</p>
            <p className="rule-note">记牌器已出：{prompt.cardCounter?.totalPlayedCardCount ?? 0} 张</p>
            <p className="rule-note">上一轮对手动作：{formatPlayedAction(prompt.roundMemory?.lastOpponentPlay)}</p>
            <details className="agent-details">
              <summary>查看完整 Prompt</summary>
              <pre>{[prompt.systemPrompt, prompt.userPrompt].filter(Boolean).join("\n\n")}</pre>
            </details>
          </>
        ) : (
          <p className="rule-note">当前暂无可展示的 prompt。</p>
        )}
      </section>

      <section className="agent-section">
        <div className="agent-section-heading">
          <strong>最近一次 Trace</strong>
          <span>{trace?.mode ?? "空"}</span>
        </div>
        {trace ? (
          <>
            <p className="rule-note">模型：{trace.model}</p>
            <p className="rule-note">执行状态：{trace.applied ? "已执行" : "未执行"}</p>
            <p className="rule-note">结果：{trace.resultMessage || "暂无"}</p>
            {trace.error ? <p className="error-copy">执行错误：{trace.error}</p> : null}
            <pre className="agent-code">{JSON.stringify(trace.decision, null, 2)}</pre>
            {trace.rawResponse ? <pre className="agent-code">{trace.rawResponse}</pre> : null}
          </>
        ) : (
          <p className="rule-note">当前还没有 agent 执行记录。</p>
        )}
      </section>
    </section>
  );
}
