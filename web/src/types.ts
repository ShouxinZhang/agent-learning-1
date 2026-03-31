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
