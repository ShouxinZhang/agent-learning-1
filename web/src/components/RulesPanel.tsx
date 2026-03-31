import type { RulesCatalog } from "../types";

type RulesPanelProps = {
  catalog: RulesCatalog;
};

export function RulesPanel({ catalog }: RulesPanelProps) {
  return (
    <section className="rules-panel">
      <div className="rules-panel-header">
        <div>
          <h2>牌型规则</h2>
          <p>后端规则目录驱动，当前顺子最高支持到 {catalog.sequenceHigh}</p>
        </div>
        <span className="rules-version">v{catalog.version}</span>
      </div>

      <div className="rules-meta">
        <div className="rules-card">
          <strong>点数序</strong>
          <p>{catalog.rankOrder.join(" > ")}</p>
        </div>
        <div className="rules-card">
          <strong>默认比较</strong>
          {catalog.notes.map((note) => (
            <p key={note}>{note}</p>
          ))}
        </div>
      </div>

      <div className="rules-sections">
        {catalog.sections.map((section) => (
          <section key={section.key} className="rules-card">
            <div className="rules-section-heading">
              <strong>{section.title}</strong>
              <span>{section.items.length} 种</span>
            </div>
            <div className="rules-items">
              {section.items.map((item) => (
                <article key={item.key} className="rule-item">
                  <div className="rule-item-top">
                    <strong>{item.name}</strong>
                    <code>{item.pattern}</code>
                  </div>
                  <p>{item.description}</p>
                  {item.minCards ? <span className="rule-chip">最少 {item.minCards} 张</span> : null}
                  {item.notes?.map((note) => (
                    <p key={note} className="rule-note">
                      {note}
                    </p>
                  ))}
                </article>
              ))}
            </div>
          </section>
        ))}
      </div>

      <div className="rules-dual-grid">
        <section className="rules-card">
          <div className="rules-section-heading">
            <strong>炸弹优先级</strong>
            <span>{catalog.bombPriority.length} 档</span>
          </div>
          <div className="priority-list">
            {catalog.bombPriority.map((item) => (
              <article key={item.key} className="priority-item">
                <span className="priority-rank">#{item.rank}</span>
                <div>
                  <strong>{item.name}</strong>
                  <p>{item.description}</p>
                  {item.notes?.map((note) => (
                    <p key={note} className="rule-note">
                      {note}
                    </p>
                  ))}
                </div>
              </article>
            ))}
          </div>
        </section>

        <section className="rules-card">
          <div className="rules-section-heading">
            <strong>常规牌型优先级</strong>
            <span>{catalog.handPriority.length} 档</span>
          </div>
          <div className="priority-list">
            {catalog.handPriority.map((item) => (
              <article key={item.key} className="priority-item">
                <span className="priority-rank">#{item.rank}</span>
                <div>
                  <strong>{item.name}</strong>
                </div>
              </article>
            ))}
          </div>
        </section>
      </div>
    </section>
  );
}
