import type { GameState } from "../types";
import { CardStrip } from "./CardStrip";

type CurrentTrickPanelProps = {
  trick?: GameState["currentTrick"];
  resolvedHand?: GameState["resolvedHand"];
  resolutionCandidates: GameState["resolutionCandidates"];
  playError: string;
  winner: string;
};

export function CurrentTrickPanel({ trick, resolvedHand, resolutionCandidates, playError, winner }: CurrentTrickPanelProps) {
  if (!trick && !resolvedHand && !playError && !winner) {
    return null;
  }

  return (
    <section className="status-inline trick-panel">
      {winner ? <strong>{winner} 已获胜</strong> : null}
      {playError ? <p className="error-copy">出牌失败：{playError}</p> : null}
      {trick ? (
        <>
          <strong>当前桌面：{trick.lastPlaySeat} 出了 {trick.resolvedHand.label}</strong>
          <CardStrip cards={trick.cards} />
        </>
      ) : (
        <strong>当前桌面为空，等待首出</strong>
      )}
      {resolvedHand ? <p className="rule-note">最近解析：{resolvedHand.label} / 主牌 {resolvedHand.mainRank}</p> : null}
      {resolutionCandidates.length > 1 ? (
        <p className="rule-note">
          候选解：{resolutionCandidates.map((item) => `${item.resolvedHand.label}${item.isPreferred ? "（优先）" : ""}`).join("、")}
        </p>
      ) : null}
    </section>
  );
}
