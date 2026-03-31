import { CardStrip } from "./CardStrip";
import type { GamePlayer } from "../types";

type PlayerPanelProps = {
  player: GamePlayer;
  selectable?: boolean;
  selectedIds?: string[];
  onToggle?: (id: string) => void;
};

export function PlayerPanel({ player, selectable = false, selectedIds = [], onToggle }: PlayerPanelProps) {
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
      <CardStrip cards={player.cards} selectable={selectable} selectedIds={selectedIds} onToggle={onToggle} />
    </section>
  );
}
