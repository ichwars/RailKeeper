import { ChevronRight, Info, TriangleAlert } from "lucide-react";
import { useI18n } from "../../shared/i18n";
import { railComPlusSource, railComPlusSymbolsSource } from "../../shared/railComPlus";
import { DecoderProfileStrip } from "./DecoderProfileStrip";
import { useDecoderProfile } from "./DecoderProfileSelect";
import { decoderMapping } from "./decoderAnalysis";

export function VehicleDecoderMapping({ state }: { state: ReturnType<typeof useDecoderProfile> }) {
  const { t } = useI18n(), mappings = decoderMapping(state.analysis, t);
  const mappingHelp = state.analysis.piko ? "vehicles.decoder.piko.mappingHelp" : state.analysis.roco ? "vehicles.decoder.roco.mappingHelp"
    : state.analysis.uhlenbrock ? "vehicles.decoder.uhlenbrock.mappingHelp"
    : state.analysis.kuehn ? "vehicles.decoder.kuehn.mappingHelp"
    : state.analysis.lenz ? "vehicles.decoder.lenz.mappingHelp" : "vehicles.decoder.mappingHelp";
  const flags = state.analysis.railComPlus?.flagsVerification;
  return <details className="function-mapping-details">
    <summary><ChevronRight size={16} aria-hidden="true" />{t("vehicles.decoder.mappingTitle")}
      <span>{t("vehicles.functions.technicalDetails")}</span></summary>
    <section className="decoder-mapping">
      <DecoderProfileStrip state={state} />
      <p className="decoder-hint">{t(mappingHelp)}</p>
      {!mappings.length ? <p className="empty-state compact">{t(state.analysis.piko || state.analysis.roco || state.analysis.uhlenbrock || state.analysis.esu5 || state.analysis.lenz || state.analysis.kuehn
        ? "vehicles.decoder.mappingEmpty" : "vehicles.decoder.unsupportedMapping")}</p>
        : <div className="table-wrap decoder-mapping-table"><table><thead><tr>
          <th>{t("vehicles.functions.title")}</th><th>{t("vehicles.decoder.condition")}</th>
          <th>{t("vehicles.decoder.action")}</th>
        </tr></thead><tbody>{mappings.map(row => <tr key={row.row}>
          <td><strong>{row.functionKey}</strong></td><td>{t(`vehicles.decoder.${row.direction}`)}</td>
          <td><strong>{row.actions.join(" · ")}</strong><small>{row.source}</small>
            {row.warning && <span className="decoder-mapping-warning"><TriangleAlert size={14} aria-hidden="true" />
              {t("vehicles.decoder.review")}</span>}</td>
        </tr>)}</tbody></table></div>}
      <p className="decoder-notice"><Info size={15} aria-hidden="true" />{t(state.analysis.piko || state.analysis.roco || state.analysis.uhlenbrock || state.analysis.kuehn || state.analysis.lenz ? mappingHelp : "vehicles.decoder.mappingLimits")}</p>
      {state.analysis.railComPlus && <>
        <p className="decoder-notice">{flags ? t("railcom.flags.verified", flags) : t("railcom.flags")}</p>
        <p className="decoder-hint"><a href={railComPlusSymbolsSource} target="_blank" rel="noreferrer">{t("railcom.symbolSource")}</a>
          {" · "}<a href={railComPlusSource} target="_blank" rel="noreferrer">{t("railcom.source")}</a></p>
      </>}
    </section>
  </details>;
}
