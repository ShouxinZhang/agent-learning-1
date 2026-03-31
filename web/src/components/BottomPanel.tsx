import type { DemoCard } from "../types";
import { CardStrip } from "./CardStrip";

type BottomPanelProps = {
  visible: boolean;
  count: number;
  cards: DemoCard[];
};

export function BottomPanel({ visible, count, cards }: BottomPanelProps) {
  return (
    <section className="bottom-panel">
      <div className="player-header">
        <strong>底牌</strong>
        <span>{visible ? "已公开" : `未公开 (${count} 张)`}</span>
      </div>
      {visible ? <CardStrip cards={cards} /> : <div className="bottom-hidden">???</div>}
    </section>
  );
}
