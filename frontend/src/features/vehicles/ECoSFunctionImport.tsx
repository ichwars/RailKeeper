import { useEffect, useRef, useState } from "react";
import { api, type ECoSFunctionSnapshot, type MasterDataEntry, type VehicleFunctionInput } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppCheckbox } from "../../shared/ui/AppCheckbox";
import { functionSymbolIcon } from "../../shared/functionSymbols";
import { DecoderLibraryNotice } from "./DecoderLibraryNotice";
import type { DecoderAnalysis, MappingSuggestion } from "./decoderAnalysis";
import type { FunctionEdit } from "./vehicleTransforms";
import { ecosFunctionSuggestion } from "./ecosFunctionSuggestions";

export function ECoSFunctionImport({ vehicleID, objectID, analysis, mappings, symbols, locked, libraryUnavailable,
  functionEdit, updateFunctionEdit, onKeysRead }: {
  vehicleID?: string; objectID?: string; analysis: DecoderAnalysis; mappings: MappingSuggestion[];
  symbols: MasterDataEntry[];
  locked: boolean; libraryUnavailable: boolean; functionEdit: (key: string) => FunctionEdit;
  updateFunctionEdit: (key: string, patch: Partial<VehicleFunctionInput>) => void;
  onKeysRead?: (keys: string[]) => void;
}) {
  const { t } = useI18n();
  const [snapshot, setSnapshot] = useState<ECoSFunctionSnapshot | null>(null);
  const [selected, setSelected] = useState<string[]>([]);
  const [busy, setBusy] = useState(false), [error, setError] = useState(false), [applied, setApplied] = useState(false);
  const pending = useRef<AbortController | null>(null);
  useEffect(() => () => { pending.current?.abort(); }, []);
  useEffect(() => setSelected([]), [analysis, symbols, mappings]);
  const rows = snapshot?.functions.map(value => ecosFunctionSuggestion(value, analysis,
    functionEdit(`F${value.index}`), symbols, snapshot.readAt, snapshot.objectId, t, snapshot.host, mappings)) || [];
  const blocked = locked || busy || libraryUnavailable;
  const read = async () => {
    if (!vehicleID || !objectID || blocked) return;
    pending.current?.abort();
    const controller = new AbortController(); pending.current = controller;
    const timeout = window.setTimeout(() => controller.abort(), 42000);
    setBusy(true); setError(false); setSnapshot(null); setSelected([]); setApplied(false);
    onKeysRead?.([]);
    try {
      const result = await api.readVehicleECoSFunctions(vehicleID, controller.signal);
      if (controller.signal.aborted) return;
      if (String(result.objectId) !== objectID) throw new Error("target_changed");
      setSnapshot(result);
      onKeysRead?.(result.functions.map(value => `F${value.index}`));
    } catch {
      if (pending.current === controller) setError(true);
    } finally {
      window.clearTimeout(timeout);
      if (pending.current === controller) setBusy(false);
    }
  };
  const apply = () => {
    if (blocked) return;
    for (const row of rows) if (selected.includes(row.key) && row.status === "ready" && row.patch) {
      updateFunctionEdit(row.key, row.patch);
    }
    setSelected([]); setApplied(true);
  };
  return <details className="ecos-function-import">
    <summary>{t("ecosFunctions.title")}</summary>
    <p>{t("ecosFunctions.help")}</p>
    <button type="button" className="secondary-button" disabled={blocked || !vehicleID || !objectID}
      onClick={() => void read()}>{t(busy ? "ecosFunctions.reading" : "ecosFunctions.read")}</button>
    {(!vehicleID || !objectID) && <p className="decoder-hint">{t("ecosFunctions.mappingRequired")}</p>}
    {error && <p role="alert">{t("ecosFunctions.error")}</p>}
    {snapshot && <>
      <DecoderLibraryNotice values={analysis.rows} profile={analysis.rows[0]?.decoderProfile || ""}
        match={analysis.libraryMatch} vehicleDecoderType={analysis.vehicleDecoderType} />
      <p>{snapshot.host} · {t("ecosFunctions.source", { id: snapshot.objectId, at: snapshot.readAt })}</p>
      {rows.length === 0 ? <p role="status">{t("ecosFunctions.empty")}</p> : <>
        <div className="ecos-function-table" role="region" aria-label={t("ecosFunctions.title")} tabIndex={0}>
          <table><thead><tr><th>{t("vehicles.functions.key")}</th><th>RailKeeper</th><th>ECoS</th>
            <th>{t("decoderLibrary.title")}</th><th>{t("ecosFunctions.proposal")}</th></tr></thead>
            <tbody>{rows.map(row => <tr key={row.key}>
              <th><AppCheckbox label={row.key} checked={selected.includes(row.key)}
                disabled={blocked || row.status !== "ready"}
                onChange={event => setSelected(keys => event.target.checked
                  ? [...keys, row.key] : keys.filter(k => k !== row.key))} /></th>
              <td>{functionEdit(row.key).name || "–"}
                <small>{symbols.find(symbol => symbol.key === functionEdit(row.key).symbolKey)?.label || "–"}</small>
                {functionEdit(row.key).persisted && <small>{t(`vehicles.functionMode.${functionEdit(row.key).mode}`)}</small>}</td>
              <td>{row.stationSymbol && functionSymbolIcon(row.stationSymbol.key, undefined, row.stationSymbol.metadata)}
                <span>{row.stationSymbol?.label || t("ecosFunctions.unknownSymbol")}</span>
                <small>{t("ecosFunctions.code", { code: snapshot.functions.find(v => `F${v.index}` === row.key)!.description })}
                  {" · "}{t(`vehicles.functionMode.${row.mode}`)}</small></td>
              <td>{row.library?.name || "–"}{row.library?.symbolKey &&
                <small>{symbols.find(symbol => symbol.key === row.library?.symbolKey)?.label || "–"}</small>}
                {row.library?.mode && <small>{t(`vehicles.functionMode.${row.library.mode}`)}</small>}</td>
              <td>{row.status === "ready" ? <>
                <span>{row.patch?.name || functionEdit(row.key).name || row.key}</span>
                <small>{t("ecosFunctions.fields", { fields: Object.keys(row.patch || {})
                  .filter(key => key !== "notes" && key !== "flagsVersion")
                  .map(key => t(key === "inverted" ? "railcom.inverted" :
                    `vehicles.functions.${key === "symbolKey" ? "symbol" : key}`)).join(", ") })}</small>
              </> : t(`ecosFunctions.${row.status}`)}</td>
            </tr>)}</tbody></table>
        </div>
        <div className="cv-toolbar">
          <button type="button" className="secondary-button" disabled={blocked}
            onClick={() => setSelected(rows.filter(row => row.status === "ready").map(row => row.key))}>
            {t("ecosFunctions.selectFree")}</button>
          <button type="button" className="secondary-button" disabled={blocked ||
            !rows.some(row => row.status === "ready" && selected.includes(row.key))} onClick={apply}>
            {t("ecosFunctions.apply")}</button>
        </div>
      </>}
    </>}
    {applied && <p role="status">{t("ecosFunctions.applied")}</p>}
  </details>;
}
