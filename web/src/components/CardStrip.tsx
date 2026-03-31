import type { DemoCard } from "../types";

type CardStripProps = {
  cards: DemoCard[];
  selectable?: boolean;
  selectedIds?: string[];
  onToggle?: (id: string) => void;
};

export function CardStrip({ cards, selectable = false, selectedIds = [], onToggle }: CardStripProps) {
  return (
    <div className="card-strip">
      {cards.map((card) => (
        <button
          key={card.id}
          type="button"
          className={[
            "card-chip",
            card.isLaizi ? "card-chip-laizi" : "",
            selectedIds.includes(card.id) ? "card-chip-selected" : "",
            selectable ? "card-chip-selectable" : "",
          ]
            .filter(Boolean)
            .join(" ")}
          onClick={() => onToggle?.(card.id)}
          disabled={!selectable}
        >
          {card.label}
        </button>
      ))}
    </div>
  );
}
