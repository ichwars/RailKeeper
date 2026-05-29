import type { RefObject } from "react";
import { Download, Save, Trash2, Upload } from "lucide-react";
import { MasterDataEntry, Vehicle, VehicleFunctionInput } from "../../shared/api";
import { FunctionSymbolPicker, functionSymbolIcon, functionSymbolMetadata } from "../../shared/functionSymbols";
import { useI18n } from "../../shared/i18n";
import { functionModes } from "./cvImport";
import { AppSelect } from "../../shared/ui/AppSelect";

type FunctionEdit = VehicleFunctionInput & { persisted?: boolean };

export function VehicleFunctionsTab({
  selected,
  draftMode = false,
  readonly,
  saving,
  functionImportInputRef,
  configuredFunctionKeys,
  functionSummary,
  showConfiguredFunctionsOnly,
  visibleFunctionKeys,
  symbols,
  onImportFunctions,
  onExportFunctions,
  onShowConfiguredFunctionsOnlyChange,
  functionEdit,
  updateFunctionEdit,
  inferFunctionTypeFromSymbol,
  saveFunction,
  deleteFunction
}: {
  selected: Vehicle | null;
  draftMode?: boolean;
  readonly: boolean;
  saving: boolean;
  functionImportInputRef: RefObject<HTMLInputElement | null>;
  configuredFunctionKeys: string[];
  functionSummary: {
    configured: number;
    sound: number;
    light: number;
  };
  showConfiguredFunctionsOnly: boolean;
  visibleFunctionKeys: string[];
  symbols: MasterDataEntry[];
  onImportFunctions: (files: FileList | null) => void;
  onExportFunctions: () => void;
  onShowConfiguredFunctionsOnlyChange: (value: boolean) => void;
  functionEdit: (functionKey: string) => FunctionEdit;
  updateFunctionEdit: (functionKey: string, patch: Partial<VehicleFunctionInput>) => void;
  inferFunctionTypeFromSymbol: (symbolKey: string, symbols: MasterDataEntry[], fallback?: string) => string;
  saveFunction: (functionKey: string) => void;
  deleteFunction: (functionKey: string) => void;
}) {
  const { t } = useI18n();
  const canEditFunctions = Boolean(selected || draftMode);

  return (
    <section className="functions-tab">
      <div className="upload-head">
        <div>
          <h3>{t("vehicles.functions.title")}</h3>
          <p>{t("vehicles.functions.subtitle")}</p>
        </div>
        <div className="cv-toolbar">
          <input
            ref={functionImportInputRef}
            type="file"
            accept="application/json,.json"
            className="visually-hidden"
            onChange={(event) => onImportFunctions(event.target.files)}
            disabled={readonly || !selected || saving}
          />
          <button type="button" className="secondary-button" onClick={() => functionImportInputRef.current?.click()} disabled={readonly || !selected || saving}>
            <Upload size={15} aria-hidden="true" />
            Import
          </button>
          <button type="button" className="secondary-button" onClick={onExportFunctions} disabled={!selected || configuredFunctionKeys.length === 0}>
            <Download size={15} aria-hidden="true" />
            Export
          </button>
        </div>
      </div>
      {!canEditFunctions && <p className="empty-state compact">{t("vehicles.functions.emptyUntilSave")}</p>}
      {canEditFunctions && (
        <div className="function-list">
          <div className="function-toolbar">
            <div className="function-summary">
              <span><strong>{functionSummary.configured}</strong> {t("vehicles.functions.configured")}</span>
              <span><strong>{functionSummary.sound}</strong> {t("vehicles.functions.sound")}</span>
              <span><strong>{functionSummary.light}</strong> {t("vehicles.functions.light")}</span>
            </div>
            <label className="switch-label compact-switch">
              <span>{t("vehicles.functions.onlyConfigured")}</span>
              <span className="switch-field">
                <input
                  type="checkbox"
                  checked={showConfiguredFunctionsOnly}
                  onChange={(event) => onShowConfiguredFunctionsOnlyChange(event.target.checked)}
                  disabled={saving}
                />
                <span />
              </span>
            </label>
          </div>
          {visibleFunctionKeys.length === 0 && (
            <p className="empty-state compact">{t("vehicles.functions.empty")}</p>
          )}
          {visibleFunctionKeys.map((functionKey) => {
            const edit = functionEdit(functionKey);
            return (
              <article key={functionKey} className={edit.persisted ? "function-row persisted" : "function-row"}>
                <strong className="function-key">
                  {functionSymbolIcon(edit.symbolKey, edit.functionType, functionSymbolMetadata(symbols, edit.symbolKey))}
                  {functionKey}
                </strong>
                <input
                  className="function-name-input"
                  value={edit.name || ""}
                  onChange={(event) => updateFunctionEdit(functionKey, { name: event.target.value })}
                  disabled={readonly || saving}
                  placeholder={t("vehicles.functions.name")}
                  aria-label={`${functionKey} ${t("vehicles.functions.name")}`}
                />
                <FunctionSymbolPicker
                  value={edit.symbolKey || ""}
                  functionType={edit.functionType}
                  symbols={symbols}
                  disabled={readonly || saving}
                  label={`${functionKey} ${t("vehicles.functions.symbol")}`}
                  onChange={(symbolKey, symbolLabel) => updateFunctionEdit(functionKey, {
                    symbolKey,
                    name: symbolLabel || edit.name || "",
                    functionType: inferFunctionTypeFromSymbol(symbolKey, symbols, edit.functionType)
                  })}
                />
                <AppSelect
                  value={edit.mode || "dauer"}
                  onChange={(event) => updateFunctionEdit(functionKey, { mode: event.target.value })}
                  disabled={readonly || saving}
                  aria-label={`${functionKey} ${t("vehicles.functions.mode")}`}
                >
                  {functionModes.map((modeName) => (
                    <option key={modeName} value={modeName}>{t(`vehicles.functionMode.${modeName}`)}</option>
                  ))}
                </AppSelect>
                <label className="switch-card function-direction">
                  <span>{t("vehicles.functions.direction")}</span>
                  <span className="switch-field">
                    <input
                      type="checkbox"
                      checked={Boolean(edit.directionDependent)}
                      onChange={(event) => updateFunctionEdit(functionKey, { directionDependent: event.target.checked })}
                      disabled={readonly || saving}
                    />
                    <span />
                  </span>
                </label>
                <input
                  className="function-note-input"
                  value={edit.notes || ""}
                  onChange={(event) => updateFunctionEdit(functionKey, { notes: event.target.value })}
                  disabled={readonly || saving}
                  placeholder={t("vehicles.functions.note")}
                  aria-label={`${functionKey} ${t("vehicles.functions.note")}`}
                />
                <div className="function-actions">
                  <button type="button" className="icon-button" onClick={() => saveFunction(functionKey)} disabled={readonly || saving || !selected} aria-label={t("vehicles.functions.save", { key: functionKey })} title={t("vehicles.save")}>
                    <Save size={15} />
                  </button>
                  <button type="button" className="icon-button danger" onClick={() => deleteFunction(functionKey)} disabled={readonly || saving || !selected || !edit.persisted} aria-label={t("vehicles.functions.delete", { key: functionKey })} title={t("vehicles.delete")}>
                    <Trash2 size={15} />
                  </button>
                </div>
              </article>
            );
          })}
        </div>
      )}
    </section>
  );
}
