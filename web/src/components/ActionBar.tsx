type ActionBarProps = {
  currentActor: string;
  actions: string[];
  busy: boolean;
  onReset: () => void;
  onAction: (kind: string) => void;
};

const actionLabels: Record<string, string> = {
  jiaodizhu: "叫地主",
  bujiao: "不叫",
  qiangdizhu: "抢地主",
  buqiang: "不抢",
  woqiang: "我抢",
};

export function ActionBar({ currentActor, actions, busy, onReset, onAction }: ActionBarProps) {
  const safeActions = actions ?? [];

  return (
    <section className="action-bar">
      <div className="action-meta">
        <strong>当前操作人：{currentActor || "-"}</strong>
      </div>
      <div className="action-buttons">
        <button type="button" className="secondary-button" onClick={onReset} disabled={busy}>
          新开一局
        </button>
        {safeActions.map((action) => (
          <button
            key={action}
            type="button"
            className="primary-button"
            onClick={() => onAction(action)}
            disabled={busy}
          >
            {actionLabels[action] ?? action}
          </button>
        ))}
      </div>
    </section>
  );
}
