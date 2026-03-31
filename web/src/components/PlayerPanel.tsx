import { CardStrip } from "./CardStrip";
import type { GamePlayer } from "../types";

type PlayerPanelProps = {
  player: GamePlayer;
};

export function PlayerPanel({ player }: PlayerPanelProps) {
  return (
    <section
      className={[
        "player-panel",
        player.isLandlord ? "player-panel-landlord" : "",
        player.isCurrent ? "player-panel-current" : "",
      ]
        .filter(Boolean)
        .join(" ")}
    >
      <div className="player-header">
        <strong>{player.seat}</strong>
        <span>{player.isLandlord ? "地主" : "农民"}</span>
        <span>{player.isCurrent ? "当前操作" : "等待中"}</span>
        <span>{player.cards.length} 张</span>
      </div>
      <CardStrip cards={player.cards} />
    </section>
  );
}
