import { CheckCircle2, CirclePlus, CircleHelp, CircleMinus, ArrowRightLeft } from "lucide-react";
import { translate, useI18n, type Language } from "./i18n";

export type ImportReviewState = "new" | "changed" | "same" | "unclear" | "excluded";
const icons = { new: CirclePlus, changed: ArrowRightLeft, same: CheckCircle2, unclear: CircleHelp, excluded: CircleMinus };

export function ImportReviewStatus({ state, language }: { state: ImportReviewState; language?: Language }) {
  const { language: activeLanguage } = useI18n(), Icon = icons[state];
  return <span className={`import-review-status is-${state}`}>
    <Icon size={15} aria-hidden="true" />{translate(language ?? activeLanguage, `importReview.state.${state}`)}
  </span>;
}

export function ImportReviewContext({ source, target, effect, selected, total, language }: {
  source: string; target: string; effect: string; selected?: number; total?: number; language?: Language;
}) {
  const { language: activeLanguage } = useI18n();
  const t = (key: string, values?: Record<string, string | number>) => translate(language ?? activeLanguage, key, values);
  return <section className="import-review-context" aria-label={t("importReview.scope")}>
    <dl>{[{ key: "source", value: source }, { key: "target", value: target }, { key: "effect", value: effect }]
      .map(({ key, value }) => <div key={key}><dt>{t(`importReview.${key}`)}</dt>
        <dd>{value || (key === "source" ? t("importReview.unknownSource") : "–")}</dd></div>)}</dl>
    {selected !== undefined && total !== undefined && <p>{t("importReview.selection", { selected, total })}</p>}
  </section>;
}
