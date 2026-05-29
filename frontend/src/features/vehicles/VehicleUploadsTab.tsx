import type { DragEvent, RefObject } from "react";
import { ChevronDown, ChevronUp, Download, ExternalLink, FileText, Image, Save, Search, Star, Trash2, Upload } from "lucide-react";
import { api, ArticleSearchDocument, Vehicle, VehicleAttachment, VehicleMaintenance } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { sourceDisplayName } from "./articleSearch";
import { attachmentAccept, attachmentCategories, imageAccept } from "./vehicleFiles";
import { formatFileSize } from "./vehicleFormat";
import { maintenanceOptionLabel } from "./vehicleMaintenance";
import { AttachmentEditState, PendingArticleImage } from "./vehicleTransforms";
import { AppSelect } from "../../shared/ui/AppSelect";

export function VehicleUploadsTab({
  selected,
  readonly,
  saving,
  imageInputRef,
  attachmentInputRef,
  maintenanceEntries,
  imageUploadMaintenanceID,
  pendingArticleImages,
  attachmentDragActive,
  attachmentUploadCategory,
  attachmentUploadMaintenanceID,
  attachmentUploadDescription,
  attachmentEdits,
  documentSearchLoading,
  documentSearchError,
  documentSearchRan,
  foundDocuments,
  onImageUploadMaintenanceIDChange,
  onUploadImages,
  onPreviewImage,
  onUpdatePendingImageTitle,
  onUpdatePendingImageMaintenance,
  onMovePendingImage,
  onSetPrimaryPendingImage,
  onRemovePendingImage,
  onUploadAttachment,
  onAttachmentDrag,
  onAttachmentDrop,
  onAttachmentUploadCategoryChange,
  onAttachmentUploadMaintenanceIDChange,
  onAttachmentUploadDescriptionChange,
  onUpdateAttachmentEdit,
  onSaveAttachment,
  onDeleteAttachment,
  onSearchDocuments,
  onImportDocument
}: {
  selected: Vehicle | null;
  readonly: boolean;
  saving: boolean;
  imageInputRef: RefObject<HTMLInputElement | null>;
  attachmentInputRef: RefObject<HTMLInputElement | null>;
  maintenanceEntries: VehicleMaintenance[];
  imageUploadMaintenanceID: string;
  pendingArticleImages: PendingArticleImage[];
  attachmentDragActive: boolean;
  attachmentUploadCategory: string;
  attachmentUploadMaintenanceID: string;
  attachmentUploadDescription: string;
  attachmentEdits: AttachmentEditState;
  documentSearchLoading: boolean;
  documentSearchError: string;
  documentSearchRan: boolean;
  foundDocuments: ArticleSearchDocument[];
  onImageUploadMaintenanceIDChange: (value: string) => void;
  onUploadImages: (files: FileList | null) => void;
  onPreviewImage: (image: PendingArticleImage) => void;
  onUpdatePendingImageTitle: (id: string, title: string) => void;
  onUpdatePendingImageMaintenance: (id: string, maintenanceId: string) => void;
  onMovePendingImage: (id: string, direction: -1 | 1) => void;
  onSetPrimaryPendingImage: (id: string) => void;
  onRemovePendingImage: (image: PendingArticleImage) => void;
  onUploadAttachment: (files: FileList | null) => void;
  onAttachmentDrag: (event: DragEvent<HTMLElement>) => void;
  onAttachmentDrop: (event: DragEvent<HTMLElement>) => void;
  onAttachmentUploadCategoryChange: (value: string) => void;
  onAttachmentUploadMaintenanceIDChange: (value: string) => void;
  onAttachmentUploadDescriptionChange: (value: string) => void;
  onUpdateAttachmentEdit: (id: string, patch: Partial<AttachmentEditState[string]>) => void;
  onSaveAttachment: (attachment: VehicleAttachment) => void;
  onDeleteAttachment: (attachment: VehicleAttachment) => void;
  onSearchDocuments: () => void;
  onImportDocument: (document: ArticleSearchDocument) => void;
}) {
  const { t } = useI18n();

  const importedAttachmentForDocument = (document: ArticleSearchDocument) => selected?.attachments?.find((attachment) => Boolean(document.url) && (attachment.description || "").includes(document.url));

  return (
    <section className="uploads-tab">
      <section className="upload-section">
        <div className="upload-head">
          <div>
            <h3>{t("vehicles.uploads.imagesTitle")}</h3>
            <p>{t("vehicles.uploads.imagesSubtitle")}</p>
          </div>
          <input
            ref={imageInputRef}
            type="file"
            multiple
            accept={imageAccept}
            className="visually-hidden"
            onChange={(event) => onUploadImages(event.target.files)}
            disabled={readonly || !selected}
          />
          {selected && maintenanceEntries.length > 0 && (
            <AppSelect
              className="upload-maintenance-select"
              value={imageUploadMaintenanceID}
              onChange={(event) => onImageUploadMaintenanceIDChange(event.target.value)}
              disabled={readonly || saving}
              aria-label={t("vehicles.uploads.noNewMaintenance")}
            >
              <option value="">{t("vehicles.uploads.noNewMaintenance")}</option>
              {maintenanceEntries.map((entry) => (
                <option key={entry.id} value={entry.id}>{maintenanceOptionLabel(entry)}</option>
              ))}
            </AppSelect>
          )}
          <button type="button" className="primary-button" onClick={() => imageInputRef.current?.click()} disabled={readonly || !selected || saving}>
            <Upload size={16} aria-hidden="true" />
            {t("vehicles.uploads.imageUpload")}
          </button>
        </div>
        {!selected && <p className="empty-state compact">{t("vehicles.uploads.noImagesUntilSave")}</p>}
        {pendingArticleImages.length === 0 ? (
          <div className="upload-list">
            <div className="image-placeholder large">
              <Image size={22} aria-hidden="true" />
              {t("vehicles.uploads.noPreview")}
            </div>
            <span>{t("vehicles.uploads.noImage")}</span>
          </div>
        ) : (
          <div className="pending-image-grid">
            {pendingArticleImages.map((image, imageIndex) => (
              <figure key={image.id} className={image.isPrimary ? "pending-image-card primary" : "pending-image-card"}>
                <button type="button" className="image-preview-button" onClick={() => onPreviewImage(image)} title={t("vehicles.uploads.openOriginal")} aria-label={t("vehicles.uploads.openOriginal")}>
                  <img src={image.url} alt="" />
                </button>
                <figcaption>
                  <input
                    value={image.title || ""}
                    onChange={(event) => onUpdatePendingImageTitle(image.id, event.target.value)}
                    disabled={readonly}
                    placeholder={t("vehicles.uploads.imageDescription")}
                    aria-label={t("vehicles.uploads.imageDescription")}
                  />
                  <span>{sourceDisplayName(image.source)}</span>
                  {maintenanceEntries.length > 0 && (
                    <AppSelect
                      className="image-maintenance-select"
                      value={image.maintenanceId || ""}
                      onChange={(event) => onUpdatePendingImageMaintenance(image.id, event.target.value)}
                      disabled={readonly || saving}
                      aria-label={t("vehicles.uploads.linkMaintenance")}
                    >
                      <option value="">{t("vehicles.uploads.noMaintenance")}</option>
                      {maintenanceEntries.map((entry) => (
                        <option key={entry.id} value={entry.id}>{maintenanceOptionLabel(entry)}</option>
                      ))}
                    </AppSelect>
                  )}
                  <div className="image-card-actions">
                    <a className="icon-button" href={image.source} target="_blank" rel="noreferrer" aria-label={t("vehicles.uploads.openSource")} title={t("vehicles.uploads.openSource")}>
                      <ExternalLink size={15} />
                    </a>
                    <button type="button" className="icon-button" onClick={() => onMovePendingImage(image.id, -1)} disabled={readonly || imageIndex === 0} aria-label={t("vehicles.uploads.moveUp")} title={t("vehicles.uploads.moveUp")}>
                      <ChevronUp size={15} />
                    </button>
                    <button type="button" className="icon-button" onClick={() => onMovePendingImage(image.id, 1)} disabled={readonly || imageIndex === pendingArticleImages.length - 1} aria-label={t("vehicles.uploads.moveDown")} title={t("vehicles.uploads.moveDown")}>
                      <ChevronDown size={15} />
                    </button>
                    <button type="button" className={image.isPrimary ? "icon-button active" : "icon-button"} onClick={() => onSetPrimaryPendingImage(image.id)} aria-label={t("vehicles.uploads.markPrimary")} title={image.isPrimary ? t("vehicles.uploads.primary") : t("vehicles.uploads.markPrimary")}>
                      <Star size={15} />
                    </button>
                    <button
                      type="button"
                      className="icon-button danger"
                      onClick={() => onRemovePendingImage(image)}
                      disabled={readonly || saving}
                      aria-label={t("vehicles.uploads.removeImage")}
                      title={t("vehicles.uploads.removeImage")}
                    >
                      <Trash2 size={15} />
                    </button>
                  </div>
                </figcaption>
              </figure>
            ))}
          </div>
        )}
      </section>

      <section className="upload-section">
        <div className="upload-head">
          <div>
            <h3>{t("vehicles.uploads.attachmentsTitle")}</h3>
            <p>{t("vehicles.uploads.attachmentsSubtitle")}</p>
          </div>
          <input
            ref={attachmentInputRef}
            type="file"
            multiple
            accept={attachmentAccept}
            className="visually-hidden"
            onChange={(event) => onUploadAttachment(event.target.files)}
            disabled={readonly || !selected}
          />
          <button type="button" className="primary-button" onClick={() => attachmentInputRef.current?.click()} disabled={readonly || !selected || saving}>
            <Upload size={16} aria-hidden="true" />
            {t("vehicles.uploads.attachmentUpload")}
          </button>
        </div>
        <section
          className={`attachment-upload-zone ${attachmentDragActive ? "active" : ""}`}
          onDragEnter={onAttachmentDrag}
          onDragOver={onAttachmentDrag}
          onDragLeave={onAttachmentDrag}
          onDrop={onAttachmentDrop}
          aria-label={t("vehicles.uploads.dropAria")}
        >
          <div>
            <strong>{t("vehicles.uploads.dropTitle")}</strong>
            <span>{t("vehicles.uploads.dropHelp")}</span>
          </div>
          <div className="attachment-upload-fields">
            <AppSelect value={attachmentUploadCategory} onChange={(event) => onAttachmentUploadCategoryChange(event.target.value)} disabled={readonly || !selected || saving}>
              <option value="">{t("vehicles.uploads.autoCategory")}</option>
              {attachmentCategories.map((category) => (
                <option key={category} value={category}>{category}</option>
              ))}
            </AppSelect>
            <AppSelect value={attachmentUploadMaintenanceID} onChange={(event) => onAttachmentUploadMaintenanceIDChange(event.target.value)} disabled={readonly || !selected || saving}>
              <option value="">{t("vehicles.uploads.noMaintenance")}</option>
              {maintenanceEntries.map((entry) => (
                <option key={entry.id} value={entry.id}>{maintenanceOptionLabel(entry)}</option>
              ))}
            </AppSelect>
            <input
              value={attachmentUploadDescription}
              onChange={(event) => onAttachmentUploadDescriptionChange(event.target.value)}
              disabled={readonly || !selected || saving}
              placeholder={t("vehicles.uploads.notePlaceholder")}
            />
          </div>
        </section>

        <section className="web-document-section">
          <div className="upload-head compact">
            <div>
              <h3>{t("vehicles.uploads.webDocumentsTitle")}</h3>
              <p>{t("vehicles.uploads.webDocumentsHelp")}</p>
            </div>
            <button type="button" className="secondary-button" onClick={onSearchDocuments} disabled={!selected || readonly || documentSearchLoading}>
              <Search size={15} aria-hidden="true" />
              {documentSearchLoading ? t("vehicles.uploads.webDocumentsSearching") : t("vehicles.uploads.webDocumentsSearch")}
            </button>
          </div>
          {documentSearchError && <p className="form-message">{documentSearchError}</p>}
          {documentSearchRan && !documentSearchLoading && foundDocuments.length === 0 && <p className="empty-state compact">{t("vehicles.uploads.webDocumentsEmpty")}</p>}
          {foundDocuments.length > 0 && (
            <div className="compact-table web-document-table">
              <table>
                <thead>
                  <tr>
                    <th>{t("vehicles.uploads.webDocument")}</th>
                    <th>{t("vehicles.spareParts.kind")}</th>
                    <th>{t("vehicles.articleSearch.source")}</th>
                    <th>{t("vehicles.actions")}</th>
                  </tr>
                </thead>
                <tbody>
                  {foundDocuments.map((document, index) => {
                    const importedAttachment = importedAttachmentForDocument(document);
                    const downloadUrl = importedAttachment && selected ? api.vehicleAttachmentDownloadUrl(selected.id, importedAttachment.id) : "";
                    return (
                      <tr key={`${document.url}-${index}`}>
                        <td><a href={document.url} target="_blank" rel="noreferrer"><ExternalLink size={14} /> {document.title}</a></td>
                        <td>{document.kind || "-"}</td>
                        <td>{document.source || "-"}</td>
                        <td>
                          {importedAttachment ? (
                            <div className="table-actions">
                              <a className="secondary-button" href={downloadUrl}>
                                <Download size={15} aria-hidden="true" />
                                {t("vehicles.uploads.webDocumentDownload")}
                              </a>
                              <a className="icon-button" href={`${downloadUrl}?inline=true`} target="_blank" rel="noreferrer" aria-label={t("vehicles.uploads.webDocumentOpen")} title={t("vehicles.uploads.webDocumentOpen")}>
                                <ExternalLink size={15} />
                              </a>
                            </div>
                          ) : (
                            <button type="button" className="secondary-button" onClick={() => onImportDocument(document)} disabled={readonly || saving}>
                              {t("vehicles.uploads.webDocumentImport")}
                            </button>
                          )}
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </section>

        {!selected && <p className="empty-state compact">{t("vehicles.uploads.attachmentsUntilSave")}</p>}
        {selected && (!selected.attachments || selected.attachments.length === 0) && (
          <p className="empty-state compact">{t("vehicles.uploads.attachmentsEmpty")}</p>
        )}
        {selected && selected.attachments && selected.attachments.length > 0 && (
          <div className="attachment-list">
            {selected.attachments.map((attachment) => {
              const edit = attachmentEdits[attachment.id] || {
                description: attachment.description || "",
                category: attachment.category || "",
                maintenanceId: attachment.maintenanceId || ""
              };
              const downloadUrl = api.vehicleAttachmentDownloadUrl(selected.id, attachment.id);
              return (
                <article key={attachment.id} className="attachment-row">
                  <div className="attachment-icon">
                    <FileText size={18} aria-hidden="true" />
                    <span>{attachment.originalName.split(".").pop()?.toUpperCase() || "DATEI"}</span>
                  </div>
                  <div className="attachment-main">
                    <strong>{attachment.originalName}</strong>
                    <span>{attachment.category || t("vehicles.uploads.noCategory")} - {attachment.mimeType || "Datei"} - {formatFileSize(attachment.sizeBytes)}</span>
                    <div className="attachment-edit-row">
                      <AppSelect value={edit.category} onChange={(event) => onUpdateAttachmentEdit(attachment.id, { category: event.target.value })} disabled={readonly}>
                        <option value="">{t("vehicles.uploads.category")}</option>
                        {attachmentCategories.map((category) => (
                          <option key={category} value={category}>{category}</option>
                        ))}
                      </AppSelect>
                      <AppSelect value={edit.maintenanceId} onChange={(event) => onUpdateAttachmentEdit(attachment.id, { maintenanceId: event.target.value })} disabled={readonly}>
                        <option value="">{t("vehicles.uploads.noMaintenance")}</option>
                        {maintenanceEntries.map((entry) => (
                          <option key={entry.id} value={entry.id}>{maintenanceOptionLabel(entry)}</option>
                        ))}
                      </AppSelect>
                      <input value={edit.description} onChange={(event) => onUpdateAttachmentEdit(attachment.id, { description: event.target.value })} disabled={readonly} placeholder={t("vehicles.uploads.note")} />
                    </div>
                  </div>
                  <div className="attachment-actions">
                    <a className="secondary-button" href={downloadUrl}>
                      <Download size={15} aria-hidden="true" />
                      Download
                    </a>
                    {attachment.mimeType?.includes("pdf") && (
                      <a className="icon-button" href={`${downloadUrl}?inline=true`} target="_blank" rel="noreferrer" aria-label={t("vehicles.uploads.openPdf")} title={t("vehicles.uploads.openPdf")}>
                        <ExternalLink size={15} />
                      </a>
                    )}
                    <button type="button" className="secondary-button" onClick={() => onSaveAttachment(attachment)} disabled={readonly || saving}>
                      <Save size={15} aria-hidden="true" />
                      {t("vehicles.save")}
                    </button>
                    <button type="button" className="danger-button" onClick={() => onDeleteAttachment(attachment)} disabled={readonly || saving}>
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
