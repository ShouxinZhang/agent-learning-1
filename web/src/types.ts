export type DemoCard = {
  id: string;
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

export type ResolvedHandView = {
  kind: string;
  label: string;
  pattern: string;
  mainRank: string;
  length: number;
  groupCount: number;
  attachmentType: string;
  compareKey: string;
  usesLaizi: boolean;
  isBomb: boolean;
  bombTier: number;
  cards: DemoCard[];
  resolvedCards: DemoCard[];
};

export type ResolutionCandidateView = {
  id: string;
  priority: number;
  isPreferred: boolean;
  resolvedHand: ResolvedHandView;
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
  currentTrick?: {
    leadingSeat: string;
    lastPlaySeat: string;
    passCount: number;
    cards: DemoCard[];
    resolvedHand: ResolvedHandView;
  };
  resolvedHand?: ResolvedHandView;
  resolutionCandidates: ResolutionCandidateView[];
  playError: string;
  winner: string;
  testMode?: {
    enabled: boolean;
    label: string;
    fixedLandlord: string;
    directPlay: boolean;
  };
};

export type RuleHand = {
  key: string;
  name: string;
  pattern: string;
  description: string;
  compareBy?: string;
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
  comparisonNotes?: string[];
  laiziResolutionNotes?: string[];
  sections: RuleSection[];
  bombPriority: BombPriority[];
  handPriority: HandPriorityRef[];
};

export type AgentPromptCard = {
  id: string;
  label: string;
  rank: string;
  suit: string;
  isLaizi: boolean;
};

export type AgentDecision = {
  seat: string;
  kind: string;
  cards?: string[];
  resolutionId?: string;
  reason: string;
};

export type PlayedAction = {
  seat: string;
  kind: string;
  cards?: string[];
  resolvedLabel?: string;
  mainRank?: string;
  relationship?: string;
};

export type SeatRoundMemory = {
  seat?: string;
  seatRole?: string;
  roundIndex: number;
  trickIndex: number;
  lastSelfPlay?: PlayedAction;
  lastTeammatePlay?: PlayedAction;
  lastOpponentPlay?: PlayedAction;
};

export type CardCounter = {
  seat?: string;
  seatRole?: string;
  playedCardsBySeat: Record<string, string[]>;
  playedRankCounts: Record<string, number>;
  remainingUnknown: number;
  blackJokerPlayed: boolean;
  redJokerPlayed: boolean;
  bombSignals: string[];
  totalPlayedCardCount: number;
};

export type AgentPrompt = {
  modelHint: string;
  state?: GameState;
  phase: string;
  currentActor: string;
  availableActions: string[];
  playerSeat?: string;
  playerRole?: string;
  playerHand: AgentPromptCard[];
  cardCounter?: CardCounter;
  roundMemory?: SeatRoundMemory;
  systemPrompt: string;
  userPrompt: string;
  actionSchema: {
    format: string;
    required: string[];
    properties: Record<string, string>;
    example: AgentDecision;
  };
};

export type AgentTrace = {
  runId?: string;
  createdAt?: string;
  mode: string;
  model: string;
  prompt: AgentPrompt;
  rawResponse: string;
  decision: AgentDecision;
  applied: boolean;
  error: string;
  resultMessage: string;
};

export type AgentTraceEnvelope = {
  trace: AgentTrace | null;
};

export type AgentRunResponse = {
  state: GameState;
  trace: AgentTrace;
};

export type AgentMatchStep = {
  stepIndex: number;
  seat: string;
  attemptMode: string;
  effectiveMode: string;
  model: string;
  prompt: AgentPrompt;
  decision: AgentDecision;
  applied: boolean;
  error: string;
  resultMessage: string;
  roundIndex: number;
  trickIndex: number;
  cardCounter: CardCounter;
  roundMemory: SeatRoundMemory;
  stateBefore: GameState;
  stateAfter: GameState;
};

export type AgentMatchTrace = {
  matchId?: string;
  startedAt?: string;
  finishedAt?: string;
  status: string;
  mode: string;
  fallbackMode?: string;
  model: string;
  winner: string;
  steps?: AgentMatchStep[];
  finalState?: GameState;
  error?: string;
  stepCount: number;
};

export type AgentMatchEnvelope = {
  match: AgentMatchTrace | null;
};

export type AgentMatchRunResponse = {
  state: GameState;
  match: AgentMatchTrace;
};
