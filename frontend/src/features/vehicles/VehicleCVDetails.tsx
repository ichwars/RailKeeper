import { useId, useRef } from "react";
import { createPortal } from "react-dom";
import { CircleHelp, Pencil, X } from "lucide-react";
import type { VehicleCVFile, VehicleCVValue } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { useModalDialogLayer } from "../../shared/ui/useModalDialogLayer";
import { cvIndexLabel } from "./CVPageControls";
import { decoderAnalysis } from "./decoderAnalysis";
import { useDecoderLibrary } from "../../shared/DecoderLibraryProvider";
import { cvValueInfo } from "./cvValueInfo";
import { useManufacturerCVHints } from "./ManufacturerCVProvider";
import { formatCVExportTimestamp } from "./cvProvenance";
import { formatDateTime } from "./vehicleFormat";
import { isECoSReading } from "../../shared/decoderLibrary";

export function VehicleCVDetails({ value, values, files, sourceLabel, decoderType = "", readonly, saving, onClose, onEdit }: {
  value: VehicleCVValue; values: VehicleCVValue[]; files: VehicleCVFile[];
  sourceLabel?: string; decoderType?: string;
  readonly: boolean; saving: boolean; onClose: () => void; onEdit: (value: VehicleCVValue) => void;
}) {
  const { language, t } = useI18n();
  const titleId = useId();
  const closeRef = useRef<HTMLButtonElement>(null);
  const { anchorRef, layerRef, onKeyDown } = useModalDialogLayer(onClose, closeRef);
  const library = useDecoderLibrary();
  const analysis = decoderAnalysis(values, value.decoderProfile || "", library.profiles, library.inventory ?? null, decoderType);
  const hints = useManufacturerCVHints();
  const info = cvValueInfo(value, analysis, t, language, hints.lookup(value));
  const hint = info.manufacturerHint;
  const hintSource = hint?.reference.sources[hint.row.sourceIndex];
  const source = files.find(file => file.id === value.sourceFileId);
  const reference = info.definition?.reference;
  const documentationSources = reference && info.profile
    ? [info.profile.sources[reference.sourceIndex]] : info.profile?.sources || [];
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
        {isECoSReading(value) && <p className="decoder-hint">{t("vehicles.decoder.ecosCache")}</p>}
        <dl className="cv-detail-values">
          <dt>{t("vehicles.decoder.readValue")}</dt><dd><strong className="cv-detail-reading">{value.value}</strong></dd>
          <dt>{t("vehicles.decoder.hexadecimal")}</dt><dd className="cv-detail-code">0x{value.value.toString(16).toUpperCase().padStart(2, "0")}</dd>
          <dt>{t("vehicles.decoder.binary")}</dt><dd className="cv-detail-code">{value.value.toString(2).padStart(8, "0")}</dd>
          <dt>{t("vehicles.cv.index")}</dt><dd>{cvIndexLabel(value, t("vehicles.cv.base"),
            info.indexUnknown ? t("vehicles.decoder.indexUnknown") : undefined)}</dd>
          {info.definition?.unit && <><dt>{t("decoderLibrary.unit")}</dt><dd>{info.definition.unit[language]}</dd></>}
          {info.definition?.factoryDefault != null && <>
            <dt>{t("decoderLibrary.documentedDefault")}</dt><dd>{info.definition.factoryDefault}
              <small>{t("decoderLibrary.defaultIsReference")}</small></dd>
          </>}
          <dt>{t("vehicle.field.decoderProfile")}</dt><dd>{value.decoderProfile || t("vehicles.cv.noProfile")}</dd>
          <dt>{t("vehicles.decoder.source")}</dt><dd>{source?.originalName || sourceLabel ||
            (isECoSReading(value) ? value.description || "ECoS" : t(value.sourceExportedAt || value.sourceLine != null ? "vehicles.decoder.importedSource" : "vehicles.decoder.sourceUnlinked"))}
            <small>{t("vehicles.decoder.sourceLine")}: <span>{value.sourceLine ?? t("vehicles.decoder.notRecorded")}</span></small>
            <small>{t("vehicles.decoder.exportedAt")}: <span>{value.sourceExportedAt
              ? formatCVExportTimestamp(value.sourceExportedAt, language) : t("vehicles.decoder.notRecorded")}</span></small></dd>
        </dl>
        {info.definition?.bits && <section className="cv-detail-bits"><h3>{t("vehicles.decoder.individualSettings")}</h3>
          {info.definition.bits.map(bit => <div key={bit.bit}><span className="cv-detail-code">Bit {bit.bit}</span>
            <div><small>{bit.name[language]}</small>
              <strong>{(value.value & (1 << bit.bit) ? bit.on : bit.off)[language]}</strong></div></div>)}
        </section>}
        {info.profile && <section className="decoder-documentation"><h3>{t("decoderLibrary.definitionSource")}</h3>
          <p>{language === "en" ? info.profile.nameEn : info.profile.name} · {info.profile.revision}</p>
          {analysis.vehicleDecoderType && <p>{t("vehicles.decoder.vehicleEvidence", {
            manufacturer: analysis.libraryMatch.manufacturerId ?? "?"
          })}</p>}
          {(info.definition?.firmwareMin || info.profile.firmwareMin) && <p>{t("decoderLibrary.minimumFirmware", {
            minimum: info.definition?.firmwareMin || info.profile.firmwareMin || ""
          })}</p>}
          {documentationSources.map(document => <div key={document.url}>
            <a href={document.url} target="_blank" rel="noreferrer">{document.title} · {document.version}</a>
            {reference && <p>{reference.location}</p>}
            <small>{document.scope}</small>
          </div>)}
        </section>}
        {hint && hintSource && <section className="decoder-documentation">
          <h3>{t("decoderLibrary.hintTitle")} · {hint.model.name}</h3>
          <p>{t(hint.row.applicability === "model" ? "decoderLibrary.modelApplicability" : "decoderLibrary.manualApplicability")}</p>
          <p>{hint.reference.note[language]}</p>
          {(hint.row.valueRange || hint.row.factoryValue) && <dl className="cv-detail-values">
            {hint.row.valueRange && <><dt>{t("decoderLibrary.valueRange")}</dt><dd>{hint.row.valueRange}</dd></>}
            {hint.row.factoryValue && <><dt>{t("decoderLibrary.documentedDefault")}</dt><dd>{hint.row.factoryValue}
              <small>{t("decoderLibrary.defaultIsReference")}</small></dd></>}
          </dl>}
          <a href={`${hintSource.url}${hint.row.page > 0 ? `#page=${hint.row.page}` : ""}`} target="_blank" rel="noreferrer">{hintSource.title} · {hintSource.version}</a>
          <p>{hint.row.page > 0 ? t("decoderLibrary.sourcePage", { page: hint.row.page }) : hintSource.scope}</p>
          <small>{hint.reference.revision}</small>
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
