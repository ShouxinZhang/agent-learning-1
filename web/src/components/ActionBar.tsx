type ActionBarProps = {
  currentActor: string;
  actions: string[];
  busy: boolean;
  selectedCount?: number;
  idleMessage?: string;
  onReset: () => void;
  onClearSelection?: () => void;
  onAction: (kind: string) => void;
};

const actionLabels: Record<string, string> = {
  jiaodizhu: "叫地主",
  bujiao: "不叫",
  qiangdizhu: "抢地主",
  buqiang: "不抢",
  woqiang: "我抢",
  play: "出牌",
  pass: "不出",
};

export function ActionBar({
  currentActor,
  actions,
  busy,
  selectedCount = 0,
  idleMessage = "当前无可执行动作",
  onReset,
  onClearSelection,
  onAction,
}: ActionBarProps) {
  const safeActions = actions ?? [];
  const hasActions = safeActions.length > 0;
  const showClear = safeActions.includes("play");

  return (
    <section className="action-bar">
      <div className="action-meta">
        <strong>{hasActions ? `当前操作人：${currentActor}` : idleMessage}</strong>
        {showClear ? <div>已选 {selectedCount} 张</div> : null}
      </div>
      <div className="action-buttons">
        <button type="button" className="secondary-button" onClick={onReset} disabled={busy}>
          新开一局
        </button>
        {showClear ? (
          <button type="button" className="secondary-button" onClick={onClearSelection} disabled={busy || selectedCount === 0}>
            清空选择
          </button>
        ) : null}
        {safeActions.map((action) => (
          <button
            key={action}
            type="button"
            className="primary-button"
            onClick={() => onAction(action)}
            disabled={busy || (action === "play" && selectedCount === 0)}
          >
            {actionLabels[action] ?? action}
          </button>
        ))}
      </div>
    </section>
  );
}
