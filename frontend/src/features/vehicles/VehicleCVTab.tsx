import { Fragment, RefObject } from "react";
import { Check, Download, FileText, Pencil, Save, Trash2, Upload } from "lucide-react";
import { api, Vehicle, VehicleCVFile, VehicleCVValue, VehicleCVValueInput } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { CVFileUploadPreview, CVImportPreview, cvCategories, cvProtocols } from "./cvImport";
import { cvFileAccept } from "./vehicleFiles";
import { formatDateTime, formatFileSize } from "./vehicleFormat";
import { AppSelect } from "../../shared/ui/AppSelect";

type ECoSDraftPreview = {
  cvValues: VehicleCVValueInput[];
  sourceSummary: {
    name: string;
    objectId: number;
  };
};

export function VehicleCVTab({
  selected,
  ecosDraft,
  readonly,
  saving,
  cvImportInputRef,
  cvFileInputRef,
  cvSummary,
  cvImportPreview,
  cvImportStats,
  cvForm,
  editingCVID,
  decoderProfileOptions,
  storedDecoderProfiles,
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
  ecosDraft: ECoSDraftPreview | null;
  readonly: boolean;
  saving: boolean;
  cvImportInputRef: RefObject<HTMLInputElement | null>;
  cvFileInputRef: RefObject<HTMLInputElement | null>;
  cvSummary: { values: number; profiles: number; files: number };
  cvImportPreview: CVImportPreview | null;
  cvImportStats: { selected: number; new: number; changed: number; same: number; invalid: number };
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
  return (                <section className="cv-tab">
                  <section className="cv-editor">
                    <div className="upload-head">
                      <div>
                        <h3>{t("vehicles.cv.title")}</h3>
                        <p>{t("vehicles.cv.subtitle")}</p>
                      </div>
                      <div className="cv-toolbar">
                        <input
                          ref={cvImportInputRef}
                          type="file"
                          accept=".json,.csv,.txt"
                          className="visually-hidden"
                          onChange={(event) => importCVValues(event.target.files)}
                          disabled={readonly || !selected || saving}
                        />
                        <button type="button" className="secondary-button" onClick={() => cvImportInputRef.current?.click()} disabled={readonly || !selected || saving}>
                          <Upload size={15} aria-hidden="true" />
                          Import
                        </button>
                        <button type="button" className="secondary-button" onClick={exportCVValues} disabled={!selected || !(selected.cvValues || []).length}>
                          <Download size={15} aria-hidden="true" />
                          Export
                        </button>
                      </div>
                    </div>
                    {!selected && !ecosDraft && <p className="empty-state compact">{t("vehicles.cv.emptyUntilSave")}</p>}
                    {!selected && ecosDraft && (
                      <section className="ecos-draft-cv-preview">
                        <div>
                          <h4>{t("vehicles.ecosDraft.cvPreviewTitle")}</h4>
                          <p>{t("vehicles.ecosDraft.cvPreviewSubtitle", { count: ecosDraft.cvValues.length })}</p>
                        </div>
                        <div className="ecos-draft-cv-grid">
                          {ecosDraft.cvValues.slice(0, 18).map((cvValue) => (
                            <span key={`${cvValue.cvNumber}-${cvValue.protocol || ""}-${cvValue.decoderProfile || ""}`}>
                              <strong>CV {cvValue.cvNumber}</strong>
                              {cvValue.value}
                            </span>
                          ))}
                          {ecosDraft.cvValues.length > 18 && <em>+{ecosDraft.cvValues.length - 18}</em>}
                        </div>
                        <p className="ecos-draft-mapping-preview">
                          {t("vehicles.ecosDraft.mappingPreviewTitle")}: {ecosDraft.sourceSummary.name} · #{ecosDraft.sourceSummary.objectId}
                        </p>
                      </section>
                    )}
                    {selected && (
                      <>
                        <div className="cv-summary">
                          <div>
                            <span>{t("vehicles.cv.values")}</span>
                            <strong>{cvSummary.values}</strong>
                          </div>
                          <div>
                            <span>{t("vehicles.cv.profiles")}</span>
                            <strong>{cvSummary.profiles}</strong>
                          </div>
                          <div>
                            <span>{t("vehicles.cv.files")}</span>
                            <strong>{cvSummary.files}</strong>
                          </div>
                        </div>
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
                              <button type="button" className="primary-button" onClick={applyCVImportPreview} disabled={saving || cvImportStats.selected === 0}>
                                <Check size={15} aria-hidden="true" />
                                {t("vehicles.articleSearch.applySelected")}
                              </button>
                              <button type="button" className="secondary-button" onClick={discardCVImportPreview} disabled={saving}>
                                {t("vehicles.cv.discard")}
                              </button>
                            </div>
                            <div className="table-wrap compact-table cv-import-table">
                              <table>
                                <thead>
                                  <tr>
                                    <th>{t("vehicles.articleSearch.apply")}</th>
                                    <th>CV</th>
                                    <th>{t("vehicles.articleSearch.current")}</th>
                                    <th>Import</th>
                                    <th>{t("vehicle.field.protocol")}</th>
                                    <th>{t("vehicles.cv.profiles")}</th>
                                    <th>{t("vehicles.articleSearch.status")}</th>
                                  </tr>
                                </thead>
                                <tbody>
                                  {cvImportPreview.rows.map((row) => (
                                    <tr key={row.id} className={`cv-import-${row.status}`}>
                                      <td>
                                        <input
                                          type="checkbox"
                                          checked={row.selected}
                                          onChange={(event) => toggleCVImportRow(row.id, event.target.checked)}
                                          disabled={row.status === "invalid" || saving}
                                          aria-label={t("vehicles.cv.applyCv", { cv: row.input.cvNumber })}
                                        />
                                      </td>
                                      <td>{row.input.cvNumber || "-"}</td>
                                      <td>{row.existing ? row.existing.value : "-"}</td>
                                      <td>{Number.isFinite(Number(row.input.value)) ? row.input.value : "-"}</td>
                                      <td>{row.input.protocol || row.existing?.protocol || "-"}</td>
                                      <td>{row.input.decoderProfile || row.existing?.decoderProfile || "-"}</td>
                                      <td>{row.message}</td>
                                    </tr>
                                  ))}
                                </tbody>
                              </table>
                            </div>
                          </section>
                        )}
                        {storedDecoderProfiles.length > 0 && (
                          <div className="decoder-profile-list" aria-label="Decoderprofile">
                            {storedDecoderProfiles.map((profile) => {
                              const valueCount = (selected.cvValues || []).filter((value) => value.decoderProfile === profile).length;
                              const fileCount = (selected.cvFiles || []).filter((file) => file.decoderProfile === profile).length;
                              return (
                                <button type="button" key={profile} onClick={() => updateCVForm({ decoderProfile: profile })} disabled={readonly || saving} title={t("vehicles.cv.useProfile", { profile })}>
                                  <strong>{profile}</strong>
                                  <span>{valueCount} CV · {fileCount} {t("vehicles.cv.files").toLocaleLowerCase()}</span>
                                </button>
                              );
                            })}
                          </div>
                        )}
                        <div className="cv-form">
                          <label>
                            CV-Nr.
                            <input type="number" min={1} max={1024} value={cvForm.cvNumber} onChange={(event) => updateCVForm({ cvNumber: Number(event.target.value) })} disabled={readonly || saving} />
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
                      </>
                    )}
                  </section>

                  <section className="cv-table-section">
                    {selected && (!selected.cvValues || selected.cvValues.length === 0) && (
                      <p className="empty-state compact">{t("vehicles.cv.empty")}</p>
                    )}
                    {selected && selected.cvValues && selected.cvValues.length > 0 && (
                      <div className="table-wrap compact-table">
                        <table>
                          <thead>
                            <tr>
                              <th>CV</th>
                              <th>{t("vehicle.field.value")}</th>
                              <th>{t("vehicle.field.category")}</th>
                              <th>{t("vehicle.field.protocol")}</th>
                              <th>{t("vehicle.field.decoderProfile")}</th>
                              <th>{t("vehicles.cv.description")}</th>
                              <th>{t("vehicles.actions")}</th>
                            </tr>
                          </thead>
                          <tbody>
                            {selected.cvValues.map((value) => (
                              <Fragment key={value.id}>
                                <tr>
                                  <td>{value.cvNumber}</td>
                                  <td>{value.value}</td>
                                  <td>{value.category || "-"}</td>
                                  <td>{value.protocol || "-"}</td>
                                  <td>{value.decoderProfile || "-"}</td>
                                  <td>{value.description || "-"}</td>
                                  <td>
                                    <div className="table-actions">
                                      <button type="button" className="icon-button" onClick={() => editCVValue(value)} disabled={readonly || saving} aria-label={t("vehicles.cv.edit")} title={t("vehicles.cv.edit")}>
                                        <Pencil size={15} />
                                      </button>
                                      <button type="button" className="icon-button danger" onClick={() => deleteCVValue(value)} disabled={readonly || saving} aria-label={t("vehicles.cv.delete")} title={t("vehicles.cv.delete")}>
                                        <Trash2 size={15} />
                                      </button>
                                    </div>
                                  </td>
                                </tr>
                                {value.history && value.history.length > 0 && (
                                  <tr className="cv-history-row">
                                    <td colSpan={7}>
                                      <details>
                                        <summary>{t("vehicles.cv.history", { count: value.history.length, suffix: value.history.length === 1 ? "" : language === "de" ? "en" : "s" })}</summary>
                                        <div className="cv-history-list">
                                          {value.history.slice(0, 5).map((entry) => (
                                            <span key={entry.id}>
                                              {formatDateTime(entry.changedAt)}: {entry.oldValue} -&gt; {entry.newValue}
                                            </span>
                                          ))}
                                        </div>
                                      </details>
                                    </td>
                                  </tr>
                                )}
                              </Fragment>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    )}
                  </section>

                  <section className="cv-files-section">
                    <div className="upload-head">
                      <div>
                        <h3>{t("vehicles.cv.filesTitle")}</h3>
                        <p>{t("vehicles.cv.filesSubtitle")}</p>
                      </div>
                      <input
                        ref={cvFileInputRef}
                        type="file"
                        multiple
                        accept={cvFileAccept}
                        className="visually-hidden"
                        onChange={(event) => uploadCVFiles(event.target.files)}
                        disabled={readonly || !selected || saving}
                      />
                      <button type="button" className="primary-button" onClick={() => cvFileInputRef.current?.click()} disabled={readonly || !selected || saving}>
                        <Upload size={16} aria-hidden="true" />
                        {t("vehicles.cv.uploadFile")}
                      </button>
                    </div>
                    {selected && (
                      <div className="cv-file-controls">
                      <input value={cvFileProfile} onChange={(event) => setCVFileProfile(event.target.value)} disabled={readonly || saving} placeholder={t("vehicles.cv.fileProfilePlaceholder")} />
                        <input value={cvFileDescription} onChange={(event) => setCVFileDescription(event.target.value)} disabled={readonly || saving} placeholder={t("vehicles.cv.fileNotePlaceholder")} />
                        <span>{t("vehicles.cv.autoMetadata")}</span>
                      </div>
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
                            <button type="button" className="secondary-button" onClick={previewCVFileValuesForImport} disabled={saving || readonly || cvFilePreviewStats.cvValues === 0}>
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
                </section>

  );
}
