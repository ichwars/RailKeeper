import { Info, TriangleAlert } from "lucide-react";
import type { VehicleCVValue, VehicleFunctionInput } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { useDecoderProfile } from "./DecoderProfileSelect";
import { DecoderProfileStrip } from "./DecoderProfileStrip";
import { esuMapping } from "./decoderAnalysis";

export function VehicleDecoderMapping({ values, readonly, update, edit }: {
  values: VehicleCVValue[]; readonly: boolean;
  update: (key: string, patch: Partial<VehicleFunctionInput>) => void;
  edit: (key: string) => VehicleFunctionInput;
}) {
  const { t } = useI18n();
  const state = useDecoderProfile(values), mappings = esuMapping(state.analysis, t);
  const groups = Array.from(new Set(mappings.map((row) => row.functionKey)));
  const suggest = (key: string) => {
    const rows = mappings.filter((row) => row.functionKey === key);
    update(key, {
      name: Array.from(new Set(rows.flatMap((row) => row.actions))).join(" / "),
      functionType: "standard", directionDependent: rows.some((row) => row.direction !== "both"),
      notes: rows.map((row) => `${t(`vehicles.decoder.${row.direction}`)}: ${row.actions.join(" · ")} (${row.source})`).join("\n")
    });
  };
  return <section className="decoder-mapping">
    <DecoderProfileStrip state={state} />
    <div className="cv-section-heading"><div><h3>{t("vehicles.decoder.mappingTitle")}</h3>
      <p>{t("vehicles.decoder.mappingHelp")}</p></div></div>
    {!mappings.length && <p className="empty-state compact">{t(state.analysis.esu5 ? "vehicles.decoder.mappingEmpty" : "vehicles.decoder.unsupportedMapping")}</p>}
    {!!mappings.length && <>
      <div className="decoder-mapping-stats"><span>{t("vehicles.decoder.mappingKeys", { count: groups.length })}</span>
        <span>{t("vehicles.decoder.mappingRows", { count: mappings.length })}</span></div>
      <div className="table-wrap decoder-mapping-table"><table><thead><tr>
        <th>{t("vehicles.functions.title")}</th><th>{t("vehicles.decoder.condition")}</th>
        <th>{t("vehicles.decoder.action")}</th><th>{t("vehicles.decoder.currentName")}</th><th>{t("vehicles.actions")}</th>
      </tr></thead><tbody>{mappings.map((row) => {
        const existing = edit(row.functionKey);
        const conflict = Boolean(existing.name || existing.symbolKey || existing.notes);
        return <tr key={row.row}><td><strong className="decoder-function-key">{row.functionKey}</strong></td>
          <td>{t(`vehicles.decoder.${row.direction}`)}</td>
          <td><strong>{row.actions.join(" · ")}</strong><small>{row.source}</small>
            {row.warning && <span className="decoder-mapping-warning"><TriangleAlert size={14} aria-hidden="true" />{t("vehicles.decoder.review")}</span>}</td>
          <td>{existing.name || <span className="decoder-hint">{t("vehicles.decoder.unnamed")}</span>}</td>
          <td><button type="button" className="cv-source-link" disabled={readonly || conflict} onClick={() => suggest(row.functionKey)}>
            {t(conflict ? "vehicles.decoder.alreadyNamed" : "vehicles.decoder.useSuggestion", { key: row.functionKey })}</button></td>
        </tr>;
      })}</tbody></table></div>
      <p className="decoder-notice"><Info size={16} aria-hidden="true" />{t("vehicles.decoder.mappingLimits")}</p>
      <p className="decoder-hint">{t("vehicles.decoder.saveSuggestion")}</p>
    </>}
  </section>;
}
