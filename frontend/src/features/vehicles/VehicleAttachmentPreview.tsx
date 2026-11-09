import { Download, ExternalLink, FileText, X } from "lucide-react";
import { VehicleAttachment, api } from "../../shared/api";
import { useI18n } from "../../shared/i18n";

export function attachmentPreviewKind(attachment: VehicleAttachment) {
  const mimeType = (attachment.mimeType || "").toLowerCase();
  const name = (attachment.originalName || "").toLowerCase();
  if (mimeType.includes("pdf") || name.endsWith(".pdf")) return "pdf";
  if (mimeType.startsWith("image/") || /\.(jpe?g|png|webp)$/i.test(name)) return "image";
  if (
    mimeType.startsWith("text/") ||
    mimeType.includes("json") ||
    mimeType.includes("xml") ||
    /\.(txt|csv|json|xml|html?)$/i.test(name)
  ) {
    return "text";
  }
  return "unsupported";
}

export function AttachmentPreviewFrame({ attachment, downloadUrl }: { attachment: VehicleAttachment; downloadUrl: string }) {
  const { t } = useI18n();
  const inlineUrl = `${downloadUrl}?inline=true`;
  const kind = attachmentPreviewKind(attachment);

  if (kind === "pdf") {
    return (
      <div className="attachment-preview pdf-preview">
        <iframe src={inlineUrl} title={t("vehicles.uploads.previewTitle", { name: attachment.originalName })} />
      </div>
    );
  }
  if (kind === "image") {
    return (
      <div className="attachment-preview image-file-preview">
        <img src={inlineUrl} alt={attachment.originalName} />
      </div>
    );
  }
  if (kind === "text") {
    return (
      <div className="attachment-preview text-file-preview">
        <iframe src={inlineUrl} title={t("vehicles.uploads.previewTitle", { name: attachment.originalName })} sandbox="" />
      </div>
    );
  }
  return (
    <div className="attachment-preview attachment-preview-empty">
      <FileText size={18} aria-hidden="true" />
      <span>{t("vehicles.uploads.previewUnavailable")}</span>
    </div>
  );
}

export function AttachmentPreviewDialog({
  vehicleId,
  attachment,
  onClose
}: {
  vehicleId: string;
  attachment: VehicleAttachment;
  onClose: () => void;
}) {
  const { t } = useI18n();
  const downloadUrl = api.vehicleAttachmentDownloadUrl(vehicleId, attachment.id);

  return (
    <div className="confirm-layer attachment-preview-layer" role="dialog" aria-modal="true" aria-label={t("vehicles.uploads.previewTitle", { name: attachment.originalName })}>
      <section className="attachment-preview-dialog">
        <div className="panel-head form-head">
          <div>
            <h2>{t("vehicles.uploads.preview")}</h2>
            <p>{attachment.originalName}</p>
          </div>
          <button type="button" className="icon-button" onClick={onClose} aria-label={t("vehicles.close")} title={t("vehicles.close")}>
            <X size={17} />
          </button>
        </div>
        <AttachmentPreviewFrame attachment={attachment} downloadUrl={downloadUrl} />
        <footer className="attachment-preview-actions">
          <a className="secondary-button" href={downloadUrl}>
            <Download size={15} aria-hidden="true" />
            Download
          </a>
          <a className="secondary-button" href={`${downloadUrl}?inline=true`} target="_blank" rel="noreferrer">
            <ExternalLink size={15} aria-hidden="true" />
            {t("vehicles.uploads.openAttachment")}
          </a>
        </footer>
      </section>
    </div>
  );
}
