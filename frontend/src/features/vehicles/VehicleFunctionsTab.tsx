import { useEffect, useMemo, useState, type RefObject } from "react";
import { Download, Plus, TriangleAlert, Upload } from "lucide-react";
import type { MasterDataEntry, Vehicle, VehicleCVValueInput, VehicleFunctionInput } from "../../shared/api";
import { translate, useI18n } from "../../shared/i18n";
import { functionKeys } from "./cvImport";
import { DecoderProfileSelect, useDecoderProfile } from "./DecoderProfileSelect";
import { esuMapping } from "./decoderAnalysis";
import { functionCVSuggestion } from "./functionCVSuggestions";
import { VehicleDecoderMapping } from "./VehicleDecoderMapping";
import { VehicleFunctionRow } from "./VehicleFunctionRow";
import type { FunctionEdit } from "./vehicleTransforms";

export function VehicleFunctionsTab({ selected, draftMode = false, draftCVValues, readonly, saving,
  functionImportInputRef, configuredFunctionKeys, showConfiguredFunctionsOnly, symbols,
  onImportFunctions, onExportFunctions, onShowConfiguredFunctionsOnlyChange, functionEdit,
  updateFunctionEdit, inferFunctionTypeFromSymbol, deleteFunction }: {
  selected: Vehicle | null; draftMode?: boolean; draftCVValues?: VehicleCVValueInput[];
  readonly: boolean; saving: boolean; functionImportInputRef: RefObject<HTMLInputElement | null>;
  configuredFunctionKeys: string[]; showConfiguredFunctionsOnly: boolean; symbols: MasterDataEntry[];
  onImportFunctions: (files: FileList | null) => Promise<string | undefined> | void; onExportFunctions: () => void;
  onShowConfiguredFunctionsOnlyChange: (value: boolean) => void;
  functionEdit: (key: string) => FunctionEdit;
  updateFunctionEdit: (key: string, patch: Partial<VehicleFunctionInput>) => void;
  inferFunctionTypeFromSymbol: (key: string, symbols: MasterDataEntry[], fallback?: string) => string;
  deleteFunction: (key: string) => void;
}) {
  const { t, language } = useI18n(), [message, setMessage] = useState("");
  useEffect(() => setMessage(""), [selected?.updatedAt, selected?.functions]);
  const state = useDecoderProfile(selected?.cvValues || draftCVValues || [], selected?.decoderType);
  const mappings = useMemo(() => esuMapping(state.analysis, (key, args) => translate(language, key, args)), [state.analysis, language]);
  const recognized = new Set(mappings.map(row => row.functionKey));
  const relevant = functionKeys.filter(key => recognized.has(key) || configuredFunctionKeys.includes(key));
  const visible = functionKeys.filter(key => !showConfiguredFunctionsOnly || relevant.includes(key) || functionEdit(key).deleted);
  const suggestions = relevant.map(key => ({ key,
    ...functionCVSuggestion(state.analysis, key, functionEdit(key), symbols, t, mappings) }));
  const canEdit = Boolean(selected || draftMode), locked = readonly || saving;
  const blocked = suggestions.filter(row => row.notesFull).map(row => row.key);
  const apply = () => {
    const patches = suggestions.filter(row => row.patch);
    for (const row of patches) updateFunctionEdit(row.key, row.patch!);
    setMessage(t("vehicles.functions.supplemented", { count: patches.length }));
  };
  return <section className="functions-tab function-editor">
    <div className="upload-head"><div><h3>{t("vehicles.functions.editorTitle")}</h3>
      <p>{t("vehicles.functions.editorHelp")}</p></div>
      <div className="cv-toolbar">
        <input ref={functionImportInputRef} type="file" accept="application/json,.json" className="visually-hidden"
          onChange={async event => { setMessage(""); const result = await onImportFunctions(event.target.files);
            if (result) setMessage(result); }} disabled={locked || !canEdit} />
        <button type="button" className="secondary-button" disabled={locked || !canEdit}
          onClick={() => functionImportInputRef.current?.click()}><Upload size={15} />{t("vehicles.functions.import")}</button>
        <button type="button" className="secondary-button" disabled={!selected || configuredFunctionKeys.length === 0}
          onClick={onExportFunctions}><Download size={15} />{t("vehicles.functions.export")}</button>
      </div>
    </div>
    {state.failed && <p className="decoder-notice" role="status">{t("decoderLibrary.loadFailed")}</p>}
    {state.loading && <p className="decoder-hint" role="status">{t("vehicles.functions.loadingSources")}</p>}
    {!canEdit ? <p className="empty-state compact">{t("vehicles.functions.emptyUntilSave")}</p> : <>
      {state.profiles.length > 1 && <DecoderProfileSelect {...state} />}
      <div className="function-toolbar">
        <div className="function-editor-counts"><span>{t("vehicles.functions.recognizedCount", { count: recognized.size })}</span>
          <span>{t("vehicles.functions.savedCount", { count: selected?.functions?.length || 0 })}</span>
          {functionKeys.some(key => functionEdit(key).dirty) && <span>{t("vehicles.functions.unsaved")}</span>}</div>
        <label className="switch-label compact-switch"><span>{t("vehicles.functions.onlyConfigured")}</span>
          <span className="switch-field"><input type="checkbox" checked={showConfiguredFunctionsOnly}
            onChange={event => onShowConfiguredFunctionsOnlyChange(event.target.checked)} disabled={saving} /><span /></span>
        </label>
      </div>
      <div className="function-supplement">
        <button type="button" className="secondary-button" disabled={locked || !suggestions.some(row => row.patch)} onClick={apply}>
          <Plus size={15} aria-hidden="true" />{t("vehicles.functions.supplement")}</button>
        <p>{t("vehicles.functions.supplementHelp")}</p>
      </div>
      {message && <p className="decoder-notice" role="status">{message}</p>}
      {blocked.length > 0 && <p className="decoder-notice"><TriangleAlert size={15} aria-hidden="true" />
        {t("vehicles.functions.sourceBlocked", { keys: blocked.join(", ") })}</p>}
      <div className="function-editor-columns" aria-hidden="true"><span>{t("vehicles.functions.key")}</span>
        <span>{t("vehicles.functions.name")}</span><span>{t("vehicles.functions.symbol")}</span>
        <span>{t("vehicles.functions.switchingMode")}</span><span /></div>
      <div className="function-list">
        {visible.length === 0 && <p className="empty-state compact">{t("vehicles.functions.emptyHelp")}</p>}
        {visible.map(key => <VehicleFunctionRow key={key} functionKey={key} edit={functionEdit(key)}
          symbols={symbols} disabled={locked} inferType={inferFunctionTypeFromSymbol}
          description={[...new Set(mappings.filter(row => row.functionKey === key).flatMap(row => row.actions))].join(" / ")}
          cv={state.analysis.railComPlus?.functions.find(row => row.functionKey === key)}
          update={patch => updateFunctionEdit(key, patch)} remove={() => deleteFunction(key)} />)}
      </div>
      <p className="function-missing-help"><TriangleAlert size={14} aria-hidden="true" />{t("vehicles.functions.missingHelp")}</p>
    </>}
    {state.analysis.rows.length > 0 && <VehicleDecoderMapping state={state} />}
  </section>;
}
