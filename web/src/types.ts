export type DemoCard = {
  label: string;
  suit: string;
  rank: string;
  isLaizi: boolean;
};

export type GamePlayer = {
  seat: string;
  isLandlord: boolean;
  isCurrent: boolean;
  cards: DemoCard[];
};

export type GameState = {
  phase: string;
  currentActor: string;
  availableActions: string[];
  message: string;
  players: GamePlayer[];
  landlord: string;
  multiplier: number;
  laizi: {
    tian: string;
    di: string;
    tianVisible: boolean;
    diVisible: boolean;
  };
  bottom: {
    visible: boolean;
    count: number;
    cards: DemoCard[];
  };
};

export type RuleHand = {
  key: string;
  name: string;
  pattern: string;
  description: string;
  minCards?: number;
  notes?: string[];
};

export type RuleSection = {
  key: string;
  title: string;
  items: RuleHand[];
};

export type BombPriority = {
  rank: number;
  key: string;
  name: string;
  description: string;
  notes?: string[];
};

export type HandPriorityRef = {
  rank: number;
  key: string;
  name: string;
};

export type RulesCatalog = {
  version: string;
  rankOrder: string[];
  sequenceHigh: string;
  notes: string[];
  sections: RuleSection[];
  bombPriority: BombPriority[];
  handPriority: HandPriorityRef[];
};
