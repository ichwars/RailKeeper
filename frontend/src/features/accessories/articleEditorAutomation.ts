import type { MasterDataEntry } from "../../shared/api";
import type { ArticleEditorForm } from "./articleEditorModel";

function normalized(value: string) {
  return value.trim().toLocaleLowerCase("de-DE");
}

export function scaleForGauges(gauges: readonly string[], entries: readonly MasterDataEntry[]) {
  const firstGauge = gauges[0];
  if (!firstGauge) return "";
  const selected = normalized(firstGauge);
  const entry = entries.find((candidate) => candidate.active && (
    normalized(candidate.label) === selected || normalized(candidate.key) === selected
  ));
  const scale = entry?.metadata.scale;
  return typeof scale === "string" ? scale.trim() : "";
}

export function suggestedArticleKeywords(
  form: ArticleEditorForm,
  articleTypeLabel: string,
  subtypeLabel: string
) {
  if (!form.name.trim() && !form.manufacturer.trim()) return "";
  const seen = new Set<string>();
  return [form.name, form.manufacturer, articleTypeLabel, subtypeLabel]
    .map((value) => value.trim())
    .filter((value) => {
      const key = normalized(value);
      if (!key || seen.has(key)) return false;
      seen.add(key);
      return true;
    })
    .join(", ");
}
