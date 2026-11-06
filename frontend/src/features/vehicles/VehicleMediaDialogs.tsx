import { useRef, type ReactNode } from "react";
import { createPortal } from "react-dom";
import { X } from "lucide-react";
import { useI18n } from "../../shared/i18n";
import { useModalDialogLayer } from "../../shared/ui/useModalDialogLayer";
import { retryableUploadItem, vehicleUploadCounts, type VehicleUploadBatch } from "./vehicleUploadBatch";
import type { VehicleMediaDeletion } from "./useVehicleMediaDeletion";

function MediaDialog({ title, busy, onClose, children, actions }: {
  title: string; busy: boolean; onClose: () => void; children: ReactNode; actions: ReactNode;
}) {
  const { t } = useI18n();
  const contentRef = useRef<HTMLDivElement>(null);
  const { anchorRef, layerRef, onKeyDown } = useModalDialogLayer(onClose, contentRef);
  return <><span ref={anchorRef} hidden />{createPortal(
    <div ref={layerRef} className="confirm-layer vehicle-media-action-layer" role="dialog" aria-modal="true"
      aria-label={title} onKeyDown={onKeyDown}>
      <section className="vehicle-media-action-dialog" aria-busy={busy}>
        <header className="panel-head form-head"><h2>{title}</h2>
          <button type="button" className="icon-button" disabled={busy} onClick={onClose} aria-label={t("vehicles.close")}>
            <X size={18} aria-hidden="true" />
          </button>
        </header>
        <div ref={contentRef} className="vehicle-media-action-content" tabIndex={0} role="region" aria-label={title}>
          {children}
        </div>
        <footer className="vehicle-media-action-buttons">{actions}</footer>
      </section>
    </div>, document.body)}</>;
}

export function VehicleMediaUploadDialog({ batch, onClose, onConfirm, onSkip }: {
  batch: VehicleUploadBatch; onClose: () => void; onConfirm: () => void; onSkip: (key: string) => void;
}) {
  const { t } = useI18n();
  const busy = batch.phase === "uploading", confirmation = batch.phase === "confirm";
  const counts = vehicleUploadCounts(batch), metadata = batch.metadata;
  const title = t(confirmation ? "vehicles.mediaUpload.confirmTitle" : "vehicles.mediaUpload.resultTitle");
  return <MediaDialog title={title} busy={busy} onClose={onClose} actions={<>
    <button type="button" className="secondary-button" disabled={busy} onClick={onClose}>
      {t(confirmation ? "vehicles.cancel" : "vehicles.close")}
    </button>
    {batch.items.some(retryableUploadItem) && <button type="button" className="primary-button"
      disabled={busy || batch.needsReview} onClick={onConfirm}>
      {t(confirmation ? "vehicles.mediaUpload.confirm" : "vehicles.mediaUpload.resume")}
    </button>}
  </>}>
    <p><strong>{batch.vehicleLabel}</strong></p>
    <p>{t(`vehicles.mediaUpload.scope.${batch.kind}`, { count: counts.total })}</p>
    <p>{t("vehicles.mediaUpload.independent")}</p>
    {metadata.category && <p>{t("vehicles.mediaUpload.category")}: {metadata.category}</p>}
    {metadata.description && <p>{t("vehicles.mediaUpload.description")}: {metadata.description}</p>}
    {metadata.decoderProfile && <p>{t("vehicles.mediaUpload.profile")}: {metadata.decoderProfile}</p>}
    {metadata.maintenanceId && <p>{t("vehicles.mediaUpload.maintenance")}</p>}
    {metadata.isPrimary && <p>{t("vehicles.mediaUpload.primary")}</p>}
    {!confirmation && <p role="status" aria-live="polite">{t("vehicles.mediaUpload.progress", counts)}</p>}
    {(counts.unknown > 0) && <p className="vehicle-media-action-warning">{t("vehicles.mediaUpload.unknownHint")}</p>}
    {batch.notice && <p role="alert" className="vehicle-media-action-warning">{batch.notice}</p>}
    <ul className="vehicle-media-upload-items">{batch.items.map(item => <li key={item.key}>
      <div><strong>{item.file.name}</strong><span>{Math.ceil(item.file.size / 1024)} KB</span></div>
      <p>{t(`vehicles.mediaUpload.state.${item.status}`)}</p>
      {item.error && <p className="form-message">{item.error}</p>}
      {!confirmation && retryableUploadItem(item) && <button type="button" className="secondary-button"
        disabled={busy} onClick={() => onSkip(item.key)}>{t("vehicles.mediaUpload.skip", { name: item.file.name })}</button>}
    </li>)}</ul>
  </MediaDialog>;
}

export function VehicleMediaDeleteDialog({ candidate, onClose, onConfirm }: {
  candidate: VehicleMediaDeletion; onClose: () => void; onConfirm: () => void;
}) {
  const { t } = useI18n();
  return <MediaDialog title={t("vehicles.mediaDelete.title")} busy={candidate.busy} onClose={onClose} actions={<>
    <button type="button" className="secondary-button" disabled={candidate.busy} onClick={onClose}>{t("vehicles.cancel")}</button>
    <button type="button" className="danger-button" disabled={candidate.busy || candidate.stale} onClick={onConfirm}>
      {t("vehicles.delete")}
    </button>
  </>}>
    <p><strong>{candidate.vehicleLabel}</strong></p><p>{candidate.label}</p>
    <p>{t(`vehicles.mediaDelete.scope.${candidate.kind}`)}</p>
    {candidate.kind === "decoder" && <p>{t("vehicles.mediaDelete.references", { count: candidate.references })}</p>}
    <p>{t("vehicles.mediaUpload.independent")}</p>
    {candidate.error && <p role="alert" className="form-message">{candidate.error}</p>}
  </MediaDialog>;
}
