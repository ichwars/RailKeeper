import { useId, useRef } from "react";
import { createPortal } from "react-dom";
import { CircleHelp, Pencil, X } from "lucide-react";
import type { VehicleCVFile, VehicleCVValue } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { useModalDialogLayer } from "../../shared/ui/useModalDialogLayer";
import { cvIndexLabel } from "./CVPageControls";
import { decoderAnalysis } from "./decoderAnalysis";
import { cvValueInfo } from "./cvValueInfo";
import { formatCVExportTimestamp } from "./cvProvenance";
import { formatDateTime } from "./vehicleFormat";

export function VehicleCVDetails({ value, values, files, readonly, saving, onClose, onEdit }: {
  value: VehicleCVValue; values: VehicleCVValue[]; files: VehicleCVFile[];
  readonly: boolean; saving: boolean; onClose: () => void; onEdit: (value: VehicleCVValue) => void;
}) {
  const { language, t } = useI18n();
  const titleId = useId();
  const closeRef = useRef<HTMLButtonElement>(null);
  const { anchorRef, layerRef, onKeyDown } = useModalDialogLayer(onClose, closeRef);
  const analysis = decoderAnalysis(values, value.decoderProfile || "");
  const info = cvValueInfo(value, analysis, t);
  const source = files.find(file => file.id === value.sourceFileId);
  const isConfig = value.cvNumber === 29 && value.indexCV31 == null && value.indexCV32 == null;
  return <>
    <span ref={anchorRef} hidden aria-hidden="true" />
    {createPortal(<div ref={layerRef} className="cv-detail-layer" onKeyDown={onKeyDown}>
      <div className="cv-detail-backdrop" onClick={onClose} aria-hidden="true" />
      <aside className="cv-detail-drawer" role="dialog" aria-modal="true" aria-labelledby={titleId}>
        <header className="cv-detail-head"><div><span>{info.category}</span>
          <h2 id={titleId}>CV {value.cvNumber} · {info.name}</h2></div>
          <button ref={closeRef} type="button" className="icon-button" onClick={onClose}
            aria-label={t("vehicles.decoder.closeDetails")}><X size={18} aria-hidden="true" /></button>
        </header>
        <div className={info.known ? "cv-detail-notice" : "cv-detail-notice unassigned"}><CircleHelp size={17} aria-hidden="true" /><p>{info.help}</p></div>
        <dl className="cv-detail-values">
          <dt>{t("vehicles.decoder.readValue")}</dt><dd><strong className="cv-detail-reading">{value.value}</strong></dd>
          <dt>{t("vehicles.decoder.hexadecimal")}</dt><dd className="cv-detail-code">0x{value.value.toString(16).toUpperCase().padStart(2, "0")}</dd>
          <dt>{t("vehicles.decoder.binary")}</dt><dd className="cv-detail-code">{value.value.toString(2).padStart(8, "0")}</dd>
          <dt>{t("vehicles.cv.index")}</dt><dd>{cvIndexLabel(value, t("vehicles.cv.base"))}</dd>
          <dt>{t("vehicle.field.decoderProfile")}</dt><dd>{value.decoderProfile || t("vehicles.cv.noProfile")}</dd>
          <dt>{t("vehicles.decoder.source")}</dt><dd>{source?.originalName || t(value.sourceExportedAt || value.sourceLine != null ? "vehicles.decoder.importedSource" : "vehicles.decoder.sourceUnlinked")}
            <small>{t("vehicles.decoder.sourceLine")}: <span>{value.sourceLine ?? t("vehicles.decoder.notRecorded")}</span></small>
            <small>{t("vehicles.decoder.exportedAt")}: <span>{value.sourceExportedAt
              ? formatCVExportTimestamp(value.sourceExportedAt, language) : t("vehicles.decoder.notRecorded")}</span></small></dd>
        </dl>
        {isConfig && <section className="cv-detail-bits"><h3>{t("vehicles.decoder.individualSettings")}</h3>
          {[0, 1, 2, 3, 4, 5].map(bit => <div key={bit}><span className="cv-detail-code">Bit {bit}</span>
            <strong>{t(`vehicles.decoder.bit29.${bit}.${value.value & (1 << bit) ? "on" : "off"}`)}</strong></div>)}
        </section>}
        <footer className="cv-detail-footer"><p>{t("vehicles.decoder.detailProvenance")}</p>
          <button type="button" className="secondary-button" disabled={readonly || saving} onClick={() => onEdit(value)}>
            <Pencil size={15} aria-hidden="true" />{t("vehicles.decoder.editValue")}</button>
        </footer>
        <details className="cv-detail-description"><summary>{t("vehicles.decoder.recordDetails")}</summary>
          <dl><dt>{t("vehicles.decoder.recordedAt")}</dt><dd>{formatDateTime(value.createdAt)}</dd></dl>
          {value.description && <p>{value.description}</p>}
        </details>
      </aside>
    </div>, document.body)}
  </>;
}
