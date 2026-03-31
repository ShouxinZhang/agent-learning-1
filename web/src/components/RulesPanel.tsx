import { useState } from "react";
import type { RulesCatalog } from "../types";

type RulesPanelProps = {
  catalog: RulesCatalog;
};

type HelpView = "hands" | "compare" | "laizi" | "bombs" | "priority";

export function RulesPanel({ catalog }: RulesPanelProps) {
  const [activeView, setActiveView] = useState<HelpView>("hands");

  return (
    <aside className="rules-panel">
      <div className="rules-panel-header">
        <div>
          <h2>帮助说明</h2>
          <p>侧边栏参考资料，当前顺子最高支持到 {catalog.sequenceHigh}</p>
        </div>
        <span className="rules-version">v{catalog.version}</span>
      </div>

      <nav className="help-subnav" aria-label="帮助说明分区">
        <button
          type="button"
          className={`help-tab${activeView === "hands" ? " help-tab-active" : ""}`}
          onClick={() => setActiveView("hands")}
        >
          牌型规则
        </button>
        <button
          type="button"
          className={`help-tab${activeView === "compare" ? " help-tab-active" : ""}`}
          onClick={() => setActiveView("compare")}
        >
          比较说明
        </button>
        <button
          type="button"
          className={`help-tab${activeView === "laizi" ? " help-tab-active" : ""}`}
          onClick={() => setActiveView("laizi")}
        >
          赖子说明
        </button>
        <button
          type="button"
          className={`help-tab${activeView === "bombs" ? " help-tab-active" : ""}`}
          onClick={() => setActiveView("bombs")}
        >
          炸弹优先级
        </button>
        <button
          type="button"
          className={`help-tab${activeView === "priority" ? " help-tab-active" : ""}`}
          onClick={() => setActiveView("priority")}
        >
          常规优先级
        </button>
      </nav>

      {activeView === "hands" ? (
        <>
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
                      {item.compareBy ? <p className="rule-note">主比较键：{item.compareBy}</p> : null}
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
        </>
      ) : null}

      {activeView === "compare" ? (
        <section className="rules-card">
          <div className="rules-section-heading">
            <strong>比较说明</strong>
            <span>{catalog.comparisonNotes?.length ?? 0} 条</span>
          </div>
          <div className="priority-list">
            {(catalog.comparisonNotes ?? []).map((note) => (
              <article key={note} className="priority-item">
                <div>
                  <p>{note}</p>
                </div>
              </article>
            ))}
          </div>
        </section>
      ) : null}

      {activeView === "laizi" ? (
        <section className="rules-card">
          <div className="rules-section-heading">
            <strong>赖子说明</strong>
            <span>{catalog.laiziResolutionNotes?.length ?? 0} 条</span>
          </div>
          <div className="priority-list">
            {(catalog.laiziResolutionNotes ?? []).map((note) => (
              <article key={note} className="priority-item">
                <div>
                  <p>{note}</p>
                </div>
              </article>
            ))}
          </div>
        </section>
      ) : null}

      {activeView === "bombs" ? (
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
      ) : null}

      {activeView === "priority" ? (
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
      ) : null}
    </aside>
  );
}
