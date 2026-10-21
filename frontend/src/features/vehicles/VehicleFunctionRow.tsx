import { ChevronRight, FileText, Plus, ShieldCheck, Trash2, Undo2 } from "lucide-react";
import type { MasterDataEntry, VehicleFunctionInput } from "../../shared/api";
import { FunctionSymbolPicker } from "../../shared/functionSymbols";
import { useI18n } from "../../shared/i18n";
import type { RailComPlusFunction } from "../../shared/railComPlus";
import { AppSelect } from "../../shared/ui/AppSelect";
import { FunctionFlagsFields } from "./FunctionFlagsFields";
import { formatCVExportTimestamp } from "./cvProvenance";
import { functionNoteParts, withFunctionOwnNote } from "./functionDraft";
import type { FunctionEdit } from "./vehicleTransforms";

export function VehicleFunctionRow({ functionKey, edit, symbols, disabled, description, cv,
  update, remove, inferType }: {
  functionKey: string; edit: FunctionEdit; symbols: MasterDataEntry[]; disabled: boolean;
  description?: string; cv?: RailComPlusFunction;
  update: (patch: Partial<VehicleFunctionInput>) => void; remove: () => void;
  inferType: (key: string, symbols: MasterDataEntry[], fallback?: string) => string;
}) {
  const { t, language } = useI18n();
  const parts = functionNoteParts(edit.notes, functionKey), locked = disabled || Boolean(edit.deleted);
  const noteBytes = new TextEncoder().encode(edit.notes || "").length;
  const canAdd = !description && !edit.persisted && !edit.dirty;
  const action = canAdd ? "addKey" : edit.deleted ? "undoDelete" : "delete";
  const state = edit.deleted ? "pendingDelete" : edit.dirty ? "unsaved" : edit.persisted ? "saved" : description ? "recognized" : "notConfigured";
  return <article className={`function-row${edit.persisted ? " persisted" : ""}${edit.deleted ? " pending-delete" : ""}`}>
    <div className="function-editor-fields">
      <strong className="function-key">{functionKey}</strong>
      <input className="function-name-input" value={edit.name || ""} disabled={locked}
        placeholder={t("vehicles.functions.name")} aria-label={`${functionKey} ${t("vehicles.functions.name")}`}
        onChange={event => update({ name: event.target.value })} />
      <FunctionSymbolPicker value={edit.symbolKey || ""} functionType={edit.functionType} symbols={symbols}
        disabled={locked} markMissing label={`${functionKey} ${t("vehicles.functions.symbol")}`}
        onChange={(symbolKey, label) => update({ symbolKey,
          ...(!edit.name?.trim() && label ? { name: label } : {}),
          functionType: inferType(symbolKey, symbols, edit.functionType) })} />
      <AppSelect value={edit.mode || "dauer"} disabled={locked}
        aria-label={`${functionKey} ${t("vehicles.functions.switchingMode")}`}
        onChange={event => update({ mode: event.target.value })}>
        <option value="dauer">{t(!edit.persisted && !edit.modeAssigned
          ? "vehicles.functions.defaultMode" : "vehicles.functionMode.dauer")}</option>
        <option value="moment">{t("vehicles.functionMode.moment")}</option>
      </AppSelect>
      <button type="button" className="icon-button function-delete" onClick={() => canAdd ? update({ name: edit.name || functionKey }) : remove()}
        disabled={disabled || (!canAdd && !edit.persisted && !edit.dirty)}
        aria-label={t(`vehicles.functions.${action}`, { key: functionKey })}
        title={t(`vehicles.functions.${action}`, { key: functionKey })}>
        {canAdd ? <Plus size={16} /> : edit.deleted ? <Undo2 size={16} /> : <Trash2 size={16} />}
      </button>
    </div>
    <details className="function-row-details">
      <summary><span className="function-row-status"><FileText size={13} aria-hidden="true" />
        {t(`vehicles.functions.${state}`)}{description ? ` · ${description}` : ""}</span>
        <span className="function-details-label"><ChevronRight size={14} aria-hidden="true" />{t("vehicles.functions.details")}</span>
      </summary>
      <div className="function-row-detail-content">
        <FunctionFlagsFields functionKey={functionKey} edit={edit} disabled={locked} update={update} />
        <label className="function-own-note">{t("vehicles.functions.ownNote")}
          <textarea value={parts.own} disabled={locked} aria-label={`${functionKey} ${t("vehicles.functions.note")}`}
            onChange={event => update({ notes: withFunctionOwnNote(edit.notes, functionKey, event.target.value) })}
            placeholder={t("vehicles.functions.note")} rows={2} aria-invalid={noteBytes > 1000 || undefined} />
        </label>
        {noteBytes > 1000 && <p className="field-error">{t("vehicles.functions.noteTooLong")}</p>}
        <details className="function-source-details">
          <summary><ChevronRight size={14} aria-hidden="true" /><ShieldCheck size={14} aria-hidden="true" />
            {t("vehicles.functions.sourceDetails")}</summary>
          {parts.sources.length > 0 ? <div className="function-source-records">
            <strong>{t("vehicles.functions.savedSource")}</strong>
            {parts.sources.map((line, i) => <p key={i}>{line}</p>)}
          </div> : <p className="decoder-hint">{t("vehicles.functions.noSavedSource")}</p>}
          {cv && <div className="function-source-records">
            <strong>{t("vehicles.functions.currentCVs")}</strong>
            <p>CV31=1 / CV32=1 · CV{cv.firstCV}={cv.firstValue} · CV{cv.firstCV + 1}={cv.secondValue}</p>
            <p>{t(cv.momentary ? "vehicles.functionMode.moment" : "vehicles.functionMode.dauer")} · {t("railcom.inverted")}: {t(cv.inverted ? "common.yes" : "common.no")}</p>
            {cv.sourceLine != null && <p>{t("railcom.line", {
              line: [cv.sourceLine, cv.secondSourceLine].filter(line => line != null).join(" / ") })}</p>}
            {cv.sourceExportedAt && <p>{t("vehicles.decoder.exportedAt")}: {formatCVExportTimestamp(cv.sourceExportedAt, language)}</p>}
          </div>}
        </details>
      </div>
    </details>
  </article>;
}
