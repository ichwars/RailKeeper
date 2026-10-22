import { useState } from "react";
import { ChevronRight, FileText, Gauge, Info, Lightbulb, Radio, SlidersHorizontal } from "lucide-react";
import { useI18n } from "../../shared/i18n";
import type { VehicleCVValue } from "../../shared/api";
import type { DecoderAnalysis } from "./decoderAnalysis";
import { cvExplanationRows, type CVTopic } from "./cvExplanationRows";
import { isECoSReading } from "../../shared/decoderLibrary";

const topics = [
  { key: "address", icon: Radio }, { key: "motor", icon: Gauge },
  { key: "braking", icon: SlidersHorizontal }, { key: "outputs", icon: Lightbulb }
] as const;

export function VehicleCVExplained({ analysis: a, values, onOpenValue }: {
  analysis: DecoderAnalysis; values: VehicleCVValue[]; onOpenValue: (value: VehicleCVValue) => void;
}) {
  const { t, language } = useI18n();
  const [topic, setTopic] = useState<CVTopic>("address");
  const rows = cvExplanationRows(a, topic, t, language);
  const mainRows = topic === "address" ? rows.slice(0, 5) : topic === "braking" ? rows.slice(0, 2) : rows;
  const extraRows = topic === "address" ? [...rows.slice(5), ...cvExplanationRows(a, "communication", t, language)]
    : topic === "braking" ? rows.slice(2) : [];
  const renderRows = (items: typeof rows) => items.map(row => {
    const number = row.numbers[0];
    const value = values.find(value => (value.decoderProfile || "") === (a.rows[0]?.decoderProfile || "") &&
      value.cvNumber === number && value.indexCV31 == null && value.indexCV32 == null);
    const mapping = row.mappingSource;
    const source = mapping ? values.find(value => (value.decoderProfile || "") === (a.rows[0]?.decoderProfile || "") &&
      value.cvNumber === mapping.cvNumber && value.indexCV31 === mapping.indexCV31 && value.indexCV32 === mapping.indexCV32) : value;
    return <article className="cv-explain-row" key={row.key}>
      <div className="cv-explain-description"><h4>{row.name}</h4><p>{row.help}</p></div>
      <div className="cv-explain-reading"><strong>{row.reading}</strong><small>{row.note || row.source}</small></div>
      <button type="button" className="cv-source-link" aria-haspopup="dialog" disabled={!source}
        onClick={() => source && onOpenValue(source)}>
        {number ? `CV ${number}` : t("vehicles.decoder.reviewShort")}<ChevronRight size={15} aria-hidden="true" /></button>
    </article>;
  });
  return <section className="cv-explained">
    <div className="cv-explained-layout">
      <nav className="cv-topic-nav" aria-label={t("vehicles.decoder.topics")}>
        {topics.map(({ key, icon: Icon }) => <button type="button" key={key} aria-pressed={topic === key}
          className={topic === key ? "active" : ""} onClick={() => setTopic(key)}>
          <Icon size={16} aria-hidden="true" />{t(`vehicles.decoder.topic.${key}`)}</button>)}
      </nav>
      <div className="cv-explained-content">
        <div className="cv-section-heading"><div><h3>{t(`vehicles.decoder.topic.${topic}`)}</h3>
          <p>{t("vehicles.decoder.explainedIntro")}</p></div>
          <span className="cv-source-badge"><FileText size={13} aria-hidden="true" />{t(isECoSReading(a.rows[0] || {})
            ? "vehicles.decoder.sourceECoS" : "vehicles.decoder.sourceList")}</span>
        </div>
        <div className="cv-explain-list">{renderRows(mainRows)}</div>
        {!!extraRows.length && <details className="cv-additional-readings" key={topic}>
          <summary>{t("vehicles.decoder.additionalReadings")}</summary>
          <div className="cv-explain-list">{renderRows(extraRows)}</div>
        </details>}
        {!rows.length && <p className="empty-state compact">{t("vehicles.decoder.mappingEmpty")}</p>}
        <p className="decoder-notice"><Info size={16} aria-hidden="true" />{t(topic === "motor"
          ? "vehicles.decoder.topicHelp.motor" : "vehicles.decoder.savedSnapshot")}</p>
      </div>
    </div>
  </section>;
}
