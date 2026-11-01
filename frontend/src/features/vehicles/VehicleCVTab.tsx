import { useEffect, useLayoutEffect, useMemo, useRef, useState, type RefObject } from "react";
import { Check, CircleHelp, Database, Download, FileCode, FileText, FolderOpen, Save, Trash2, Upload } from "lucide-react";
import { DecoderProfileStrip } from "./DecoderProfileStrip";
import { useDecoderProfile } from "./DecoderProfileSelect";
import { VehicleCVDetails } from "./VehicleCVDetails";
import { VehicleCVExplained } from "./VehicleCVExplained";
import { RailComPlusPreview } from "./RailComPlusPreview";
import { CVImportRowsTable } from "./CVImportRowsTable";
import { VehicleCVValuesTable } from "./VehicleCVValuesTable";
import { api, Vehicle, VehicleCVFile, VehicleCVValue, VehicleCVValueInput } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { CVFileUploadPreview, CVImportPreview, cvCategories, cvProtocols } from "./cvImport";
import { cvFileAccept } from "./vehicleFiles";
import { formatDateTime, formatFileSize } from "./vehicleFormat";
import { AppSelect } from "../../shared/ui/AppSelect";
import { ManufacturerCVProvider, ManufacturerCVStatus } from "./ManufacturerCVProvider";

type ECoSDraftPreview = {
  source?: "ecos" | "cs3";
  cvValues: VehicleCVValueInput[];
  sourceSummary: {
    name: string;
    objectId: number;
    profile?: string;
  };
};

export function VehicleCVTab({
  selected,
  decoderType,
  ecosDraft,
  readonly,
  saving,
  cvImportInputRef,
  cvFileInputRef,
  cvImportPreview,
  cvImportStats,
  cvForm,
  editingCVID,
  cvFileProfile,
  cvFileDescription,
  cvFileUploadPreview,
  cvFilePreviewStats,
  importCVValues,
  exportCVValues,
  selectCVImportRows,
  applyCVImportPreview,
  discardCVImportPreview,
  toggleCVImportRow,
  updateCVForm,
  resetCVForm,
  saveCVValue,
  editCVValue,
  deleteCVValue,
  uploadCVFiles,
  setCVFileProfile,
  setCVFileDescription,
  applyFirstCVFileSuggestion,
  previewCVFileValuesForImport,
  applyCVFileFunctionSuggestions,
  confirmCVFileUpload,
  discardCVFileUploadPreview,
  deleteCVFile
}: {
  selected: Vehicle | null;
  decoderType?: string;
  ecosDraft: ECoSDraftPreview | null;
  readonly: boolean;
  saving: boolean;
  cvImportInputRef: RefObject<HTMLInputElement | null>;
  cvFileInputRef: RefObject<HTMLInputElement | null>;
  cvSummary: { values: number; profiles: number; files: number };
  cvImportPreview: CVImportPreview | null;
  cvImportStats: { imported?: number; selected: number; new: number; changed: number; same: number; invalid: number };
  cvForm: VehicleCVValueInput;
  editingCVID: string | null;
  decoderProfileOptions: string[];
  storedDecoderProfiles: string[];
  cvFileProfile: string;
  cvFileDescription: string;
  cvFileUploadPreview: CVFileUploadPreview | null;
  cvFilePreviewStats: { cvValues: number; functions: number };
  importCVValues: (files: FileList | null) => void;
  exportCVValues: () => void;
  selectCVImportRows: (mode: "all" | "none" | "empty") => void;
  applyCVImportPreview: () => void;
  discardCVImportPreview: () => void;
  toggleCVImportRow: (id: string, selected: boolean) => void;
  updateCVForm: (patch: Partial<VehicleCVValueInput>) => void;
  resetCVForm: () => void;
  saveCVValue: () => void;
  editCVValue: (value: VehicleCVValue) => void;
  deleteCVValue: (value: VehicleCVValue) => void;
  uploadCVFiles: (files: FileList | null) => void;
  setCVFileProfile: (value: string) => void;
  setCVFileDescription: (value: string) => void;
  applyFirstCVFileSuggestion: () => void;
  previewCVFileValuesForImport: () => void;
  applyCVFileFunctionSuggestions: () => void;
  confirmCVFileUpload: () => void;
  discardCVFileUploadPreview: () => void;
  deleteCVFile: (file: VehicleCVFile) => void;
}) {
  const { language, t } = useI18n();
  const [view, setView] = useState(cvFileUploadPreview ? "files" : "explained");
  const sectionRef = useRef<HTMLElement>(null);
  const editorRef = useRef<HTMLDetailsElement>(null);
  const [detailId, setDetailId] = useState<string | null>(null);
  const values = useMemo<VehicleCVValue[]>(() => {
    if (selected) return selected.cvValues || [];
    if (!ecosDraft) return [];
    return ecosDraft.cvValues.map((value, index) => ({ ...value,
      id: `draft-cv:${ecosDraft.sourceSummary.objectId}:${index}`,
      vehicleId: "", createdAt: "", updatedAt: ""
    }));
  }, [selected, ecosDraft]);
  const detailValue = values.find(value => value.id === detailId);
  const draftSource = !selected && ecosDraft
    ? `${ecosDraft.sourceSummary.name} · #${ecosDraft.sourceSummary.objectId}` : undefined;
  useEffect(() => {
    if (view === "raw" && editingCVID && editorRef.current) {
      editorRef.current.scrollIntoView({ block: "start" });
      editorRef.current.querySelector("input")?.focus();
    }
  }, [view, editingCVID]);
  useLayoutEffect(() => {
    const body = sectionRef.current?.closest(".modal-body");
    if (body) body.scrollTop = 0;
  }, [view]);
  const activeView = view;
  const ecosProfile = useMemo(() => selected?.ecosDecoderProfile ||
    (ecosDraft?.source === "ecos" && ecosDraft.sourceSummary.profile
      ? { name: ecosDraft.sourceSummary.profile, objectId: String(ecosDraft.sourceSummary.objectId), host: "", readAt: "" }
      : undefined), [selected?.ecosDecoderProfile, ecosDraft]);
  const manualType = decoderType ?? selected?.decoderType ?? "";
  const decoder = useDecoderProfile(values, manualType, ecosProfile);
  const viewIcons = { explained: CircleHelp, raw: Database, files: FolderOpen };
  return <ManufacturerCVProvider values={values} profile={detailValue?.decoderProfile ?? decoder.profile}
    manualType={manualType} ecosProfile={ecosProfile}><section className="cv-tab" ref={sectionRef}>
    <div className="cv-entry-actions"><h3>{t("vehicles.decoder.dataTitle")}</h3><div className="cv-toolbar">
      <input ref={cvImportInputRef} type="file" accept=".json,.csv,.txt" className="visually-hidden"
        onChange={(event) => importCVValues(event.target.files)} disabled={readonly || !selected || saving} />
      <input ref={cvFileInputRef} type="file" multiple accept={cvFileAccept} className="visually-hidden"
        onChange={(event) => uploadCVFiles(event.target.files)} disabled={readonly || !selected || saving} />
      <button type="button" className="primary-button" onClick={() => cvImportInputRef.current?.click()} disabled={readonly || !selected || saving}>
        <Upload size={15} />{t("vehicles.decoder.importList")}</button>
      <button type="button" className="secondary-button" onClick={() => { setView("files"); cvFileInputRef.current?.click(); }} disabled={readonly || !selected || saving}>
        <FileText size={15} />{t("vehicles.decoder.uploadProject")}</button>

    </div></div>
    {!selected && !ecosDraft && <p className="empty-state compact">{t("vehicles.cv.emptyUntilSave")}</p>}
    {!selected && ecosDraft && <div className="ecos-draft-cv-preview">
      <p>{t("vehicles.ecosDraft.cvPreviewSubtitle", { count: ecosDraft.cvValues.length })}</p>
      <p>{t("vehicles.ecosDraft.mappingPreviewTitle")}: {draftSource}</p>
    </div>}
    {!!cvImportStats.imported && <p className="cv-import-success" role="status">{t("vehicles.cv.imported", { count: cvImportStats.imported })}</p>}
                        {cvImportPreview && (
                          <section className="cv-import-preview" aria-label={t("vehicles.cv.importPreview")}>
                            <div className="cv-import-head">
                              <div>
                                <h4>{t("vehicles.cv.importCheck")}</h4>
                                <p>{cvImportPreview.fileName}</p>
                              </div>
                              <div className="cv-import-badges" aria-label={t("vehicles.cv.importSummary")}>
                                <span>{t("vehicles.cv.new", { count: cvImportStats.new })}</span>
                                <span>{t("vehicles.cv.changed", { count: cvImportStats.changed })}</span>
                                <span>{t("vehicles.cv.same", { count: cvImportStats.same })}</span>
                                {cvImportStats.invalid > 0 && <span className="danger">{t("vehicles.cv.invalid", { count: cvImportStats.invalid })}</span>}
                              </div>
                            </div>
                            <div className="cv-import-actions">
                              <button type="button" className="secondary-button" onClick={() => selectCVImportRows("empty")} disabled={saving}>
                                {t("vehicles.cv.onlyNew")}
                              </button>
                              <button type="button" className="secondary-button" onClick={() => selectCVImportRows("all")} disabled={saving}>
                                {t("vehicles.articleSearch.selectAll")}
                              </button>
                              <button type="button" className="secondary-button" onClick={() => selectCVImportRows("none")} disabled={saving}>
                                {t("vehicles.articleSearch.selectNone")}
                              </button>
                              <button type="button" className="primary-button" onClick={applyCVImportPreview} disabled={readonly || saving || cvImportStats.selected === 0}>
                                <Check size={15} aria-hidden="true" />
                                {t("vehicles.articleSearch.applySelected")}
                              </button>
                              <button type="button" className="secondary-button" onClick={discardCVImportPreview} disabled={saving}>
                                {t("vehicles.cv.discard")}
                              </button>
                            </div>
                            <CVImportRowsTable key={cvImportPreview.fileName} preview={cvImportPreview} saving={saving} readonly={readonly} toggle={toggleCVImportRow} />
                          </section>
                        )}

    <DecoderProfileStrip state={decoder} />
    <ManufacturerCVStatus />
    <nav className="cv-subtabs" aria-label={t("vehicles.decoder.cvViews")}>
      {(["explained", "raw", "files"] as const).map((tab) => { const Icon = viewIcons[tab]; return <button type="button" key={tab}
        className={activeView === tab ? "active" : ""} aria-pressed={activeView === tab} onClick={() => setView(tab)}>
        <Icon size={16} aria-hidden="true" />{t(`vehicles.decoder.view.${tab}`)}{tab === "raw" && <span className="cv-tab-count" aria-hidden="true">{values.length}</span>}</button>; })}
    </nav>
    {activeView === "explained" && <VehicleCVExplained key={decoder.profile} analysis={decoder.analysis} values={values} onOpenValue={value => setDetailId(value.id)} />}
    {activeView === "raw" && (selected || ecosDraft) && <section className="cv-table-section">
      <div className="cv-section-heading"><div><h3>{t("vehicles.decoder.view.raw")}</h3><p>{t("vehicles.decoder.rawIntro")}</p></div>
        <button type="button" className="secondary-button" onClick={exportCVValues} disabled={!selected?.cvValues?.length}>
          <Download size={15} aria-hidden="true" />{t("vehicles.decoder.exportValues")}</button></div>
      {selected && <details ref={editorRef} className="cv-manual-editor" open={editingCVID ? true : undefined}><summary>{t("vehicles.decoder.manualCV")}</summary>
                        <div className="cv-form">
                          <label>
                            CV-Nr.
                            <input type="number" min={1} max={1024} value={cvForm.cvNumber} onChange={(event) => updateCVForm({ cvNumber: Number(event.target.value) })} disabled={readonly || saving} />
                          </label>
                          <label>CV31
                            <input aria-label="CV31" type="number" min={0} max={255} value={cvForm.indexCV31 ?? ""}
                              onChange={(event) => updateCVForm({ indexCV31: event.target.value === "" ? null : Number(event.target.value) })} disabled={readonly || saving} />
                          </label>
                          <label>CV32
                            <input aria-label="CV32" type="number" min={0} max={255} value={cvForm.indexCV32 ?? ""}
                              onChange={(event) => updateCVForm({ indexCV32: event.target.value === "" ? null : Number(event.target.value) })} disabled={readonly || saving} />
                          </label>
                          <label>
                            {t("vehicle.field.value")}
                            <input type="number" min={0} max={255} value={cvForm.value} onChange={(event) => updateCVForm({ value: Number(event.target.value) })} disabled={readonly || saving} />
                          </label>
                          <label>
                            {t("vehicle.field.category")}
                            <AppSelect value={cvForm.category || ""} onChange={(event) => updateCVForm({ category: event.target.value })} disabled={readonly || saving}>
                              <option value="">{t("vehicle.field.category")}</option>
                              {cvCategories.map((category) => (
                                <option key={category} value={category}>{category}</option>
                              ))}
                            </AppSelect>
                          </label>
                          <label>
                            {t("vehicle.field.protocol")}
                            <AppSelect value={cvForm.protocol || ""} onChange={(event) => updateCVForm({ protocol: event.target.value })} disabled={readonly || saving}>
                              <option value="">{t("vehicles.cv.noProtocol")}</option>
                              {cvProtocols.map((protocol) => (
                                <option key={protocol} value={protocol}>{protocol}</option>
                              ))}
                            </AppSelect>
                          </label>
                          <label>
                            {t("vehicle.field.decoderProfile")}
                            <input value={cvForm.decoderProfile || ""} onChange={(event) => updateCVForm({ decoderProfile: event.target.value })} disabled={readonly || saving} placeholder="z. B. ESU LokPilot 5" />
                          </label>
                          {(selected.cvFiles || []).length > 0 && (
                            <label>
                              {t("vehicles.cv.sourceFile")}
                              <AppSelect value={cvForm.sourceFileId || ""} onChange={(event) => updateCVForm({ sourceFileId: event.target.value })} disabled={readonly || saving}>
                                <option value="">{t("vehicles.cv.noFile")}</option>
                                {(selected.cvFiles || []).map((file) => (
                                  <option key={file.id} value={file.id}>{file.originalName}</option>
                                ))}
                              </AppSelect>
                            </label>
                          )}
                          <label className="cv-description">
                            {t("vehicles.cv.description")}
                            <input value={cvForm.description || ""} onChange={(event) => updateCVForm({ description: event.target.value })} disabled={readonly || saving} />
                          </label>
                        </div>
                        <p className="cv-import-help">{t("vehicles.cv.indexHelp")}</p>
                        <div className="cv-actions">
                          {editingCVID && (
                            <button type="button" className="secondary-button" onClick={resetCVForm} disabled={readonly || saving}>
                              {t("vehicles.cancel")}
                            </button>
                          )}
                          <button type="button" className="primary-button" onClick={saveCVValue} disabled={readonly || saving}>
                            <Save size={15} aria-hidden="true" />
                            {editingCVID ? t("vehicles.cv.saveCv") : t("vehicles.cv.addCv")}
                          </button>
                        </div>

      </details>}
      <VehicleCVValuesTable values={values} decoderType={selected?.decoderType} readonly={readonly || !selected} saving={saving} edit={editCVValue} remove={deleteCVValue} onExplain={value => setDetailId(value.id)} />
    </section>}
    <div hidden={activeView !== "files"} className="cv-files-view">
      <div className="cv-section-heading"><div><h3>{t("vehicles.decoder.view.files")}</h3><p>{t("vehicles.decoder.filesIntro")}</p></div></div>
      <div className="cv-file-options"><article><Database className="cv-option-icon" size={36} aria-hidden="true" />
        <span className="cv-file-types">TXT · CSV · JSON</span><h4>{t("vehicles.decoder.importList")}</h4>
        <p>{t("vehicles.decoder.listHelp")}</p><ul>{["values", "digital", "functions"].map((key) => <li key={key}>{t(`vehicles.decoder.listFeature.${key}`)}</li>)}</ul>
      </article><article><FileCode className="cv-option-icon project" size={36} aria-hidden="true" />
        <span className="cv-file-types">ESUX · ESU · LokProgrammer</span><h4>{t("vehicles.decoder.uploadProject")}</h4>
        <p>{t("vehicles.decoder.projectHelp")}</p><ul>{["original", "metadata", "values"].map((key) => <li key={key}>{t(`vehicles.decoder.projectFeature.${key}`)}</li>)}</ul>
      </article></div>
                  <section className="cv-files-section">
                    <div className="upload-head">
                      <div>
                        <h3>{t("vehicles.decoder.sourcesTitle")}</h3>
                        <p>{t("vehicles.cv.filesSubtitle")}</p>
                      </div>
                    </div>
                    {selected && (
                      <details className="cv-upload-options"><summary>{t("vehicles.decoder.fileOptions")}</summary><div className="cv-file-controls">
                      <input value={cvFileProfile} onChange={(event) => setCVFileProfile(event.target.value)} disabled={readonly || saving} placeholder={t("vehicles.cv.fileProfilePlaceholder")} />
                        <input value={cvFileDescription} onChange={(event) => setCVFileDescription(event.target.value)} disabled={readonly || saving} placeholder={t("vehicles.cv.fileNotePlaceholder")} />
                        <span>{t("vehicles.cv.autoMetadata")}</span>
                      </div></details>
                    )}
                    {selected && cvFileUploadPreview && (
                      <section className="cv-file-preview">
                        <div className="upload-head compact">
                          <div>
                            <h3>{t("vehicles.cv.uploadPreview")}</h3>
                            <p>{t("vehicles.cv.previewHelp")}</p>
                          </div>
                          <div className="inline-actions">
                            <button type="button" className="secondary-button" onClick={applyFirstCVFileSuggestion} disabled={saving || !cvFileUploadPreview.previews.some((preview) => preview.hasMetadata)}>
                              {t("vehicles.cv.applySuggestion")}
                            </button>
                            <button type="button" className="secondary-button" onClick={previewCVFileValuesForImport} disabled={saving || readonly || cvFilePreviewStats.cvValues === 0 || cvFileUploadPreview.previews.some((preview) => Boolean(preview.cvImportError))}>
                              {t("vehicles.cv.checkCvs")}
                            </button>
                            <button type="button" className="secondary-button" onClick={applyCVFileFunctionSuggestions} disabled={saving || readonly || cvFilePreviewStats.functions === 0}>
                              {t("vehicles.cv.applyFunctions")}
                            </button>
                            <button type="button" className="primary-button" onClick={confirmCVFileUpload} disabled={saving || readonly}>
                              <Upload size={15} aria-hidden="true" />
                              {t("vehicles.cv.saveFiles")}
                            </button>
                            <button type="button" className="secondary-button" onClick={discardCVFileUploadPreview} disabled={saving}>
                              {t("vehicles.cancel")}
                            </button>
                          </div>
                        </div>
                        <div className="cv-file-preview-list">
                          {cvFileUploadPreview.previews.map((preview) => (
                            <article key={preview.fileName} className={preview.hasMetadata ? "" : "no-metadata"}>
                              <div>
                                <strong>{preview.fileName}</strong>
                                <span>{preview.mimeType || "Datei"} - {formatFileSize(preview.sizeBytes)}</span>
                              </div>
                              {preview.suggestedPreviewImage && (
                                <figure className="decoder-preview-image">
                                  <img src={preview.suggestedPreviewImage.dataUrl} alt="" />
                                  <figcaption>{preview.suggestedPreviewImage.width} × {preview.suggestedPreviewImage.height}</figcaption>
                                </figure>
                              )}
                              {preview.cvImportError && <p role="alert">{t("vehicles.cv.invalidFile")} {preview.cvImportError}</p>}
                              {preview.hasMetadata ? (
                                <dl>
                                  <div><dt>{t("vehicles.cv.project")}</dt><dd>{preview.projectName || "-"}</dd></div>
                                  <div><dt>{t("vehicles.cv.decoder")}</dt><dd>{preview.decoder || "-"}</dd></div>
                                  <div><dt>{t("vehicles.cv.address")}</dt><dd>{preview.address || "-"}</dd></div>
                                  <div><dt>{t("vehicles.cv.type")}</dt><dd>{preview.type || "-"}</dd></div>
                                  <div><dt>{t("vehicles.cv.manufacturer")}</dt><dd>{preview.manufacturer || "-"}</dd></div>
                                  <div><dt>LokProgrammer</dt><dd>{preview.lokProgrammer || "-"}</dd></div>
                                </dl>
                              ) : (
                                <p>{t("vehicles.cv.noMetadata")}</p>
                              )}
                              {!preview.cvImportError && <RailComPlusPreview values={preview.suggestedCvValues || []} defaultProfile={preview.suggestedDecoderProfile} />}
                              {((preview.suggestedCvValues?.length || 0) > 0 || (preview.suggestedFunctions?.length || 0) > 0) && (
                                <div className="decoder-preview-summary">
                                  {(preview.suggestedCvValues?.length || 0) > 0 && (
                                    <span>{t("vehicles.cv.detectedValues", { count: preview.suggestedCvValues?.length || 0 })}</span>
                                  )}
                                  {(preview.suggestedFunctions?.length || 0) > 0 && (
                                    <span>{t("vehicles.cv.detectedFunctions", { count: preview.suggestedFunctions?.length || 0 })}</span>
                                  )}
                                </div>
                              )}
                            </article>
                          ))}
                        </div>
                      </section>
                    )}
                    {!!selected?.cvValues?.length && <div className="cv-imported-source"><FileText size={18} aria-hidden="true" /><div><strong>{t("vehicles.decoder.importedValues", { count: selected.cvValues.length })}</strong><small>{t("vehicles.decoder.importedValuesHelp")}</small></div><button type="button" className="cv-source-link" onClick={() => setView("raw")}>{t("vehicles.decoder.showValues")}</button></div>}
                    {selected && (!selected.cvFiles || selected.cvFiles.length === 0) && (
                      <p className="empty-state compact">{t("vehicles.cv.filesEmpty")}</p>
                    )}
                    {selected && selected.cvFiles && selected.cvFiles.length > 0 && (
                      <div className="attachment-list">
                        {selected.cvFiles.map((file) => {
                          const downloadUrl = api.vehicleCVFileDownloadUrl(selected.id, file.id);
                          return (
                            <article key={file.id} className="attachment-row">
                              <div className="attachment-icon">
                                <FileText size={18} aria-hidden="true" />
                                <span>{file.originalName.split(".").pop()?.toUpperCase() || "CV"}</span>
                              </div>
                              <div className="attachment-main">
                                <strong>{file.originalName}</strong>
                                <span>{file.decoderProfile || t("vehicles.cv.noProfile")} - {file.mimeType || "Datei"} - {formatFileSize(file.sizeBytes)}</span>
                                {file.description && <span>{file.description}</span>}
                              </div>
                              <div className="attachment-actions">
                                <a className="secondary-button" href={downloadUrl}>
                                  <Download size={15} aria-hidden="true" />
                                  Download
                                </a>
                                <button type="button" className="danger-button" onClick={() => deleteCVFile(file)} disabled={readonly || saving}>
                                  <Trash2 size={15} aria-hidden="true" />
                                  {t("vehicles.delete")}
                                </button>
                              </div>
                            </article>
                          );
                        })}
                      </div>
                    )}
                  </section>

    </div>
    {detailValue && <VehicleCVDetails value={detailValue} values={values} files={selected?.cvFiles || []}
      decoderType={selected?.decoderType}
      sourceLabel={draftSource} readonly={readonly || !selected} saving={saving} onClose={() => setDetailId(null)}
      onEdit={value => { if (!selected) return; setDetailId(null); setView("raw"); editCVValue(value); }} />}
  </section></ManufacturerCVProvider>;
}
