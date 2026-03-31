import type { DemoCard } from "../types";

type CardStripProps = {
  cards: DemoCard[];
};

export function CardStrip({ cards }: CardStripProps) {
  return (
    <div className="card-strip">
      {cards.map((card, index) => (
        <div
          key={`${card.label}-${index}`}
          className={`card-chip${card.isLaizi ? " card-chip-laizi" : ""}`}
        >
          {card.label}
        </div>
      ))}
    </div>
  );
}
