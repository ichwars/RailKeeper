import { Check, Download, ExternalLink, Printer, X } from "lucide-react";
import { CreateVehicleRequest, ExhibitionEntry, ExhibitionList, Vehicle, VehicleAttachment } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { sourceDisplayName } from "./articleSearch";
import { formatDate } from "./vehicleFormat";
import { PendingArticleImage, vehicleExhibitionEligible } from "./vehicleTransforms";
import { AppSelect } from "../../shared/ui/AppSelect";

export function QrDialog({
  form,
  qrSvg,
  error,
  onClose,
  onDownloadPng,
  onDownloadSvg,
  onPrint
}: {
  form: CreateVehicleRequest;
  qrSvg: string;
  error: string;
  onClose: () => void;
  onDownloadPng: () => void;
  onDownloadSvg: () => void;
  onPrint: () => void;
}) {
  const { t } = useI18n();
  const decoderNumber = form.digitalDecoderNumber || form.dtDecoderNumber || "";
  return (
    <div className="confirm-layer qr-layer" role="dialog" aria-modal="true" aria-label="QR-Code">
      <section className="qr-dialog">
        <div className="panel-head form-head">
          <div>
            <h2>QR-Code</h2>
            <p>{form.inventoryNumber || t("vehicles.qr.noInventory")} - {form.name || t("vehicles.qr.noName")}</p>
            {decoderNumber && <p className="qr-dialog-meta">Decoder-Nr.: {decoderNumber}</p>}
          </div>
          <button type="button" className="icon-button" onClick={onClose} aria-label={t("vehicles.close")} title={t("vehicles.close")}>
            <X size={17} />
          </button>
        </div>
        {error && <p className="form-message">{error}</p>}
        <button type="button" className="qr-preview-button" onClick={onPrint} disabled={!qrSvg} title={t("vehicles.qr.printView")}>
          {qrSvg ? (
            <span dangerouslySetInnerHTML={{ __html: qrSvg }} />
          ) : (
            t("vehicles.qr.creating")
          )}
        </button>
        <div className="qr-dialog-actions">
          <button type="button" className="secondary-button" onClick={onDownloadPng} disabled={!qrSvg}>
            <Download size={16} aria-hidden="true" />
            PNG
          </button>
          <button type="button" className="secondary-button" onClick={onDownloadSvg} disabled={!qrSvg}>
            <Download size={16} aria-hidden="true" />
            SVG
          </button>
          <button type="button" className="primary-button" onClick={onPrint} disabled={!qrSvg}>
            <Printer size={16} aria-hidden="true" />
            {t("overview.print")}
          </button>
        </div>
      </section>
    </div>
  );
}

export function ImagePreviewDialog({
  image,
  onClose
}: {
  image: PendingArticleImage;
  onClose: () => void;
}) {
  const { t } = useI18n();
  return (
    <div className="confirm-layer image-preview-layer" role="dialog" aria-modal="true" aria-label={t("vehicles.imagePreview.title")}>
      <section className="image-preview-dialog">
        <div className="panel-head form-head">
          <div>
            <h2>{t("vehicles.imagePreview.title")}</h2>
            <p className="image-preview-source">
              {image.title || t("vehicles.imagePreview.defaultTitle")} - {sourceDisplayName(image.source)}
              <a className="icon-button image-title-link" href={image.source} target="_blank" rel="noreferrer" aria-label={t("vehicles.articleSearch.sourceOpen")} title={t("vehicles.articleSearch.sourceOpen")}>
                <ExternalLink size={15} />
              </a>
            </p>
          </div>
          <button type="button" className="icon-button" onClick={onClose} aria-label={t("vehicles.close")} title={t("vehicles.close")}>
            <X size={17} />
          </button>
        </div>
        <img src={image.url} alt="" />
      </section>
    </div>
  );
}

export function ReportDialog({
  reportMode,
  reportTitle,
  reportSelection,
  reportIncludeQRCode,
  reportIncludeImages,
  selectedCount,
  canUseSelected,
  creating,
  onReportModeChange,
  onReportTitleChange,
  onReportSelectionChange,
  onReportIncludeQRCodeChange,
  onReportIncludeImagesChange,
  onClose,
  onSubmit
}: {
  reportMode: "summary" | "details";
  reportTitle: string;
  reportSelection: "all" | "selected";
  reportIncludeQRCode: boolean;
  reportIncludeImages: boolean;
  selectedCount: number;
  canUseSelected: boolean;
  creating: boolean;
  onReportModeChange: (value: "summary" | "details") => void;
  onReportTitleChange: (value: string) => void;
  onReportSelectionChange: (value: "all" | "selected") => void;
  onReportIncludeQRCodeChange: (value: boolean) => void;
  onReportIncludeImagesChange: (value: boolean) => void;
  onClose: () => void;
  onSubmit: () => void;
}) {
  const { t } = useI18n();
  return (
    <div className="confirm-layer report-layer" role="dialog" aria-modal="true" aria-label={t("vehicles.report.title")}>
      <section className="report-dialog">
        <header className="report-dialog-head">
          <div>
            <Printer size={18} aria-hidden="true" />
            <h2>{t("vehicles.report.title")}</h2>
          </div>
          <button type="button" className="icon-button" onClick={onClose} aria-label={t("vehicles.close")} title={t("vehicles.close")}>
            <X size={18} />
          </button>
        </header>

        <div className="report-form-grid">
          <label>
            {t("vehicles.report.type")}
            <AppSelect value={reportMode} onChange={(event) => onReportModeChange(event.target.value as "summary" | "details")}>
              <option value="summary">{t("vehicles.report.summary")}</option>
              <option value="details">{t("vehicles.report.details")}</option>
            </AppSelect>
          </label>

          <label>
            {t("vehicles.report.customTitle")}
            <input value={reportTitle} onChange={(event) => onReportTitleChange(event.target.value)} placeholder="Fahrzeugsammlung" />
          </label>

          <fieldset className="report-choice-group">
            <legend>{t("vehicles.report.scope")}</legend>
            <label>
              <input type="radio" checked={reportSelection === "all"} onChange={() => onReportSelectionChange("all")} />
              {t("vehicles.report.all")}
            </label>
            <label>
              <input type="radio" checked={reportSelection === "selected"} onChange={() => onReportSelectionChange("selected")} disabled={!canUseSelected} />
              {t("vehicles.report.selected", { count: selectedCount })}
            </label>
          </fieldset>

          <fieldset className="report-choice-group">
            <legend>{t("vehicles.report.options")}</legend>
            <label>
              <input type="checkbox" checked={reportIncludeQRCode} onChange={(event) => onReportIncludeQRCodeChange(event.target.checked)} />
              {t("vehicles.report.qrCode")}
            </label>
            <label>
              <input type="checkbox" checked={reportIncludeImages} onChange={(event) => onReportIncludeImagesChange(event.target.checked)} />
              {t("vehicles.report.image")}
            </label>
          </fieldset>
        </div>

        {creating && <p className="report-status">{t("vehicles.report.creating")}</p>}

        <footer className="report-dialog-actions">
          <button type="button" className="secondary-button" onClick={onClose}>
            {t("vehicles.cancel")}
          </button>
          <button type="button" className="primary-button" onClick={onSubmit} disabled={creating}>
            {t("vehicles.report.create")}
          </button>
        </footer>
      </section>
    </div>
  );
}

export function ExhibitionAssignmentDialog({
  assignment,
  duplicateVehicle,
  duplicateDecoder,
  onClose,
  onListChange,
  onSave
}: {
  assignment: {
    vehicle: Vehicle;
    lists: ExhibitionList[];
    selectedListID: string;
    loadingEntries: boolean;
    saving: boolean;
    error: string;
  };
  duplicateVehicle?: ExhibitionEntry;
  duplicateDecoder?: ExhibitionEntry;
  onClose: () => void;
  onListChange: (listID: string) => void;
  onSave: () => void;
}) {
  const { t } = useI18n();
  return (
    <div className="confirm-layer" role="dialog" aria-modal="true" aria-label={t("vehicles.exhibition.dialogTitle")}>
      <section className="confirm-card exhibition-assign-dialog">
        <div className="panel-head form-head">
          <div>
            <h2>{t("vehicles.exhibition.dialogTitle")}</h2>
          </div>
          <button type="button" className="icon-button" onClick={onClose} aria-label={t("vehicles.close")}>
            <X size={17} />
          </button>
        </div>
        <p className="exhibition-assign-vehicle">
          {assignment.vehicle.inventoryNumber} · {assignment.vehicle.name}
        </p>
        <label className="exhibition-assign-field">
          <span>{t("vehicles.exhibition.list")}</span>
          <AppSelect
            value={assignment.selectedListID}
            onChange={(event) => onListChange(event.target.value)}
            disabled={assignment.loadingEntries || assignment.saving || assignment.lists.length === 0}
          >
            {assignment.lists.length === 0 && <option value="">{t("vehicles.exhibition.noOpenLists")}</option>}
            {assignment.lists.map((list) => (
              <option key={list.id} value={list.id}>
                {list.designation} · {formatDate(list.date)}
              </option>
            ))}
          </AppSelect>
        </label>
        <div className="exhibition-assign-checks" aria-live="polite">
          <span className={vehicleExhibitionEligible(assignment.vehicle) ? "check-ok" : "check-error"}>
            {vehicleExhibitionEligible(assignment.vehicle) ? <Check size={14} /> : <X size={14} />}
            {t("vehicles.exhibition.checkDecoder")}
          </span>
          <span className={!duplicateVehicle ? "check-ok" : "check-error"}>
            {!duplicateVehicle ? <Check size={14} /> : <X size={14} />}
            {duplicateVehicle ? t("vehicles.exhibition.duplicateVehicle", { name: duplicateVehicle.locomotiveName }) : t("vehicles.exhibition.checkVehicle")}
          </span>
          <span className={!duplicateDecoder ? "check-ok" : "check-error"}>
            {!duplicateDecoder ? <Check size={14} /> : <X size={14} />}
            {duplicateDecoder ? t("vehicles.exhibition.duplicateDecoder", { name: duplicateDecoder.locomotiveName }) : t("vehicles.exhibition.checkUniqueDecoder")}
          </span>
        </div>
        {assignment.error && <p className="form-message error">{assignment.error}</p>}
        <div className="confirm-actions">
          <button type="button" className="secondary-button" onClick={onClose}>
            {t("vehicles.cancel")}
          </button>
          <button
            type="button"
            className="primary-button"
            onClick={onSave}
            disabled={
              assignment.saving ||
              assignment.loadingEntries ||
              !assignment.selectedListID ||
              Boolean(duplicateVehicle) ||
              Boolean(duplicateDecoder)
            }
          >
            {assignment.saving ? t("vehicles.saving") : t("vehicles.exhibition.assign")}
          </button>
        </div>
      </section>
    </div>
  );
}

export function DeleteVehicleDialog({
  vehicle,
  onClose,
  onConfirm
}: {
  vehicle: Vehicle;
  onClose: () => void;
  onConfirm: () => void;
}) {
  const { t } = useI18n();
  return (
    <div className="confirm-layer" role="dialog" aria-modal="true" aria-label={t("vehicles.delete.aria")}>
      <section className="confirm-card">
        <div className="panel-head form-head">
          <h2>{t("vehicles.delete.title")}</h2>
          <button type="button" className="icon-button" onClick={onClose} aria-label={t("vehicles.close")}>
            <X size={17} />
          </button>
        </div>
        <p>
          {vehicle.inventoryNumber} - {vehicle.name}
        </p>
        <div className="confirm-actions">
          <button type="button" className="secondary-button" onClick={onClose}>
            {t("vehicles.cancel")}
          </button>
          <button type="button" className="danger-button" onClick={onConfirm}>
            {t("vehicles.delete")}
          </button>
        </div>
      </section>
    </div>
  );
}

export function DeleteAttachmentDialog({
  attachment,
  onClose,
  onConfirm
}: {
  attachment: VehicleAttachment;
  onClose: () => void;
  onConfirm: () => void;
}) {
  const { t } = useI18n();
  return (
    <div className="confirm-layer" role="dialog" aria-modal="true" aria-label={t("vehicles.uploads.deleteAttachmentAria")}>
      <section className="confirm-card">
        <div className="panel-head form-head">
          <h2>{t("vehicles.uploads.deleteAttachmentTitle")}</h2>
          <button type="button" className="icon-button" onClick={onClose} aria-label={t("vehicles.close")}>
            <X size={17} />
          </button>
        </div>
        <p>{attachment.originalName}</p>
        <div className="confirm-actions">
          <button type="button" className="secondary-button" onClick={onClose}>
            {t("vehicles.cancel")}
          </button>
          <button type="button" className="danger-button" onClick={onConfirm}>
            {t("vehicles.delete")}
          </button>
        </div>
      </section>
    </div>
  );
}
