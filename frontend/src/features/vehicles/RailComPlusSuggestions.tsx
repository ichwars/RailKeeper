import { Info } from "lucide-react";
import type { MasterDataEntry, VehicleFunctionInput } from "../../shared/api";
import { exactECoSSymbol } from "../../shared/ecosSymbols";
import { functionSymbolIcon } from "../../shared/functionSymbols";
import { useI18n } from "../../shared/i18n";
import { railComPlusSource, railComPlusSymbolsSource, type RailComPlusData, type RailComPlusFunction } from "../../shared/railComPlus";
import { formatCVExportTimestamp } from "./cvProvenance";
import { railComPlusFlagAvailability, railComPlusFlagPatch } from "./railComPlusFunctionFlags";
import type { FunctionEdit } from "./vehicleTransforms";

const reviewedSymbols: Record<number, string> = { 2: "function", 3: "headlight", 9: "shunting", 10: "abv" };

export function RailComPlusSuggestions({ data, symbols = [], readonly = true, edit, update }: {
  data?: RailComPlusData; symbols?: MasterDataEntry[]; readonly?: boolean;
  edit?: (key: string) => FunctionEdit;
  update?: (key: string, patch: Partial<VehicleFunctionInput>) => void;
}) {
  const { t, language } = useI18n();
  if (!data || (!data.name && !data.nameIssue && !data.functions.length && !data.incompleteFunctions && !data.unassignedFunctions)) return null;
  const source = (row: RailComPlusFunction) => `CV31=1 / CV32=1 · CV${row.firstCV}=${row.firstValue} · CV${row.firstCV + 1}=${row.secondValue}`;
  return <section className="railcom-plus" aria-label={t("railcom.title")}>
    <h3>{t("railcom.title")}</h3>
    {data.name && <p><strong>{t("railcom.name")}:</strong> {data.name}</p>}
    {data.nameIssue && <p className="decoder-notice">{t(`railcom.name.${data.nameIssue}`)}</p>}
    <p className="decoder-hint">{t("railcom.help")}</p>
    <div className="railcom-plus-list">{data.functions.map(row => {
      const symbol = exactECoSSymbol(row.iconCode, symbols);
      const label = symbol?.label || (reviewedSymbols[row.iconCode] ? t(`railcom.symbol.${reviewedSymbols[row.iconCode]}`)
        : t("railcom.symbol.unknown", { code: row.iconCode }));
      const current = edit?.(row.functionKey);
      const protectedSymbol = Boolean(current?.symbolKey);
      const available = current && railComPlusFlagAvailability(current);
      const flagPatch = current && railComPlusFlagPatch(data, row, current, t);
      const canApplyFlags = Boolean(flagPatch);
      const notesFull = Boolean(data.flagsVerification && available && (available.mode || available.inverted) && !flagPatch);
      const applyFlags = () => {
        if (readonly || !edit || !update) return;
        const patch = railComPlusFlagPatch(data, row, edit(row.functionKey), t);
        if (patch) update(row.functionKey, patch);
      };
      const apply = () => {
        if (readonly || !symbol || !edit || !update || edit(row.functionKey).symbolKey) return;
        const existing = edit(row.functionKey), note = `RailComPlus · ${source(row)} · ${t("railcom.symbol.code", { code: row.iconCode })}`;
        update(row.functionKey, {
          symbolKey: symbol.key,
          ...(!existing.name?.trim() ? { name: label } : {}),
          ...((!existing.functionType || existing.functionType === "standard") && symbol.metadata?.category === "Licht"
            ? { functionType: "licht" } : {}),
          notes: [existing.notes, note].filter(Boolean).join("\n")
        });
      };
      return <article key={row.functionKey}>
        <strong className="decoder-function-key">{row.functionKey}</strong>
        <div><strong>{symbol && functionSymbolIcon(symbol.key, undefined, symbol.metadata)} {label}</strong>
          <small>{t("railcom.symbol.code", { code: row.iconCode })} · {source(row)}</small>
          <small>{t(row.momentary ? "vehicles.functionMode.moment" : "vehicles.functionMode.dauer")} · {t("railcom.inverted")}: {t(row.inverted ? "common.yes" : "common.no")} · {t(data.flagsVerification ? "railcom.flags.checked" : "railcom.comparison")}</small>
          {(row.sourceLine != null || row.sourceExportedAt) && <small>
            {row.sourceLine != null && t("railcom.line", { line: row.sourceLine })}
            {row.sourceLine != null && row.sourceExportedAt && " · "}
            {row.sourceExportedAt && `${t("vehicles.decoder.exportedAt")}: ${formatCVExportTimestamp(row.sourceExportedAt, language)}`}
          </small>}
          {current?.name && <small>{t("railcom.current")}: {current.name}</small>}
          {data.flagsVerification && available && (!available.mode || !available.inverted) &&
            <small>{t(!available.mode && !available.inverted ? "railcom.flags.protectedBoth"
              : !available.mode ? "railcom.flags.protectedMode" : "railcom.flags.protectedInversion")}</small>}
          {notesFull && <small>{t("railcom.flags.notesLimit")}</small>}
        </div>
        {update && <div className="railcom-plus-actions"><button type="button" className="secondary-button" onClick={apply}
          disabled={readonly || !symbol || protectedSymbol} aria-label={t("railcom.apply", { key: row.functionKey })}>
          {t(protectedSymbol ? "railcom.protected" : symbol ? "railcom.apply" : "railcom.unavailable", { key: row.functionKey })}</button>
          {data.flagsVerification && <button type="button" className="secondary-button" onClick={applyFlags}
            disabled={readonly || !canApplyFlags} aria-label={t("railcom.flags.apply", { key: row.functionKey })}>
            {t(canApplyFlags ? "railcom.flags.apply" : notesFull ? "railcom.flags.noNoteSpace"
              : "railcom.flags.protectedBoth", { key: row.functionKey })}
          </button>}
        </div>}
      </article>;
    })}</div>
    {data.unassignedFunctions > 0 && <p className="decoder-hint">{t("railcom.unassigned", { count: data.unassignedFunctions })}</p>}
    {data.incompleteFunctions > 0 && <p className="decoder-notice">{t("railcom.incomplete", { count: data.incompleteFunctions })}</p>}
    <p className="decoder-notice"><Info size={15} aria-hidden="true" />{data.flagsVerification
      ? t("railcom.flags.verified", data.flagsVerification) : t("railcom.flags")}</p>
    <p className="decoder-hint"><a href={railComPlusSymbolsSource} target="_blank" rel="noreferrer">{t("railcom.symbolSource")}</a>{" · "}
      <a href={railComPlusSource} target="_blank" rel="noreferrer">{t("railcom.source")}</a></p>
    {update && <p className="decoder-hint">{t("railcom.save")}</p>}
  </section>;
}
