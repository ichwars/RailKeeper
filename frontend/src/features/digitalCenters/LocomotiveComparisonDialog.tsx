import { useEffect, useRef, useState } from "react";
import { CheckCircle2, TriangleAlert, X } from "lucide-react";

import { useI18n, type Language } from "../../shared/i18n";
import { useModalDialogLayer } from "../../shared/ui/useModalDialogLayer";
import type {
  DigitalCenterWorkItem,
  DigitalCenterWriteConfirmation,
  DigitalCenterWriteField,
  DigitalCenterWritePreview
} from "./digitalCenterModel";

export function LocomotiveComparisonDialog({
  item, canWrite, loading, preview, confirmation, error, onPreview, onConfirm, onClose
}: {
  item: DigitalCenterWorkItem;
  canWrite: boolean;
  loading: boolean;
  preview: DigitalCenterWritePreview | null;
  confirmation: DigitalCenterWriteConfirmation | null;
  error: string;
  onPreview: (fields: DigitalCenterWriteField[]) => Promise<DigitalCenterWritePreview>;
  onConfirm: () => Promise<DigitalCenterWriteConfirmation>;
  onClose: () => void;
}) {
  const { t, language } = useI18n();
  const [confirmed, setConfirmed] = useState(false);
  const closeButtonRef = useRef<HTMLButtonElement>(null);
  const { anchorRef, layerRef, onKeyDown } = useModalDialogLayer(onClose, closeButtonRef);
  const rows = comparisonRows(item, t);
  const changedFields = rows.filter((row) => row.current !== row.desired).map((row) => row.field);
  const grantValid = Boolean(preview?.token.trim()) && validExpiry(preview?.expiresAt);
  const verified = confirmation?.result === "verified" && confirmation.applied && confirmation.verified;

  useEffect(() => setConfirmed(false), [preview?.token]);

  return (
    <>
      <span ref={anchorRef} aria-hidden="true" />
      <div ref={layerRef} className="digital-comparison-layer" role="dialog" aria-modal="true"
        aria-label={t("digitalCenters.dialog.label", { name: item.name })} onKeyDown={onKeyDown}>
        <section className="digital-comparison-dialog">
          <header><div><p className="eyebrow">{t("digitalCenters.dialog.eyebrow")}</p><h2>{item.name}</h2></div>
            <button ref={closeButtonRef} type="button" className="digital-center-icon-button"
              aria-label={t("digitalCenters.dialog.closeComparison")} onClick={onClose}>
              <X size={19} aria-hidden="true" />
            </button></header>

          {!preview && <ComparisonTable rows={rows} t={t} />}
          {preview && <WritePreview preview={preview} t={t} language={language} />}
          {error && <p className="digital-write-result error" role="alert"
            aria-label={t("digitalCenters.error.write")}>
            <TriangleAlert size={17} aria-hidden="true" />{error}
          </p>}
          {confirmation && <p className={`digital-write-result ${verified ? "verified" : "error"}`}>
            {verified ? <CheckCircle2 size={17} aria-hidden="true" /> :
              <TriangleAlert size={17} aria-hidden="true" />}
            <span><strong>{verified ? t("digitalCenters.write.verified") : resultLabel(confirmation, t)}</strong>
              <small>{confirmation.message}</small></span>
          </p>}

          {preview && !confirmation && <label className="digital-write-confirmation">
            <input type="checkbox" checked={confirmed}
              onChange={(event) => setConfirmed(event.target.checked)} />
            <span>{t("digitalCenters.write.consent")}</span>
          </label>}

          <footer>
            {!preview && <button type="button" className="digital-center-button"
              disabled={!canWrite || loading || changedFields.length === 0}
              onClick={() => void onPreview(changedFields).catch(() => undefined)}>
              {t("digitalCenters.write.createPreview")}
            </button>}
            {preview && !confirmation && <button type="button" className="digital-center-button"
              disabled={!confirmed || !grantValid || loading}
              title={!grantValid ? t("digitalCenters.write.grantInvalid") : undefined}
              onClick={() => void onConfirm().catch(() => undefined)}>
              {t("digitalCenters.write.confirm")}</button>}
            <button type="button" className="digital-center-button" onClick={onClose}>
              {t("digitalCenters.common.close")}</button>
          </footer>
          {!canWrite && <p className="digital-write-unavailable">
            {t("digitalCenters.write.unsupported")}
          </p>}
        </section>
      </div>
    </>
  );
}

type ComparisonRow = {
  label: string;
  field: DigitalCenterWriteField;
  current: string | number;
  desired: string | number;
};

function ComparisonTable({ rows, t }: { rows: ComparisonRow[]; t: Translate }) {
  return <table><thead><tr><th>{t("digitalCenters.dialog.field")}</th>
    <th>{t("digitalCenters.dialog.station")}</th><th>RailKeeper</th></tr></thead>
    <tbody>{rows.map((row) => <tr key={row.field}><th>{row.label}</th>
      <td title={String(row.current)}>{row.current}</td>
      <td title={String(row.desired)}>{row.desired}</td></tr>)}</tbody></table>;
}

function WritePreview({ preview, t, language }: {
  preview: DigitalCenterWritePreview; t: Translate; language: Language;
}) {
  return <section className="digital-write-preview" aria-label={t("digitalCenters.write.preview")}>
    <p><strong>{t("digitalCenters.write.direction")}</strong>
      <span>{t("digitalCenters.write.directionValue")}</span></p>
    <table><thead><tr><th>{t("digitalCenters.dialog.field")}</th>
      <th>{t("digitalCenters.write.current")}</th><th>{t("digitalCenters.write.desired")}</th></tr></thead>
      <tbody>{preview.changes.map((change) => <tr key={change.field}>
        <th>{fieldLabel(change.field, t)}</th><td>{change.current}</td><td>{change.desired}</td>
      </tr>)}</tbody></table>
    <small>{t("digitalCenters.write.grantUntil", { value: formatDateTime(preview.expiresAt, language) })}</small>
  </section>;
}

function comparisonRows(item: DigitalCenterWorkItem, t: Translate): ComparisonRow[] {
  return [
    { label: t("digitalCenters.field.name"), field: "name", current: item.center.name ?? "–",
      desired: item.railkeeper.name ?? "–" },
    { label: t("digitalCenters.field.address"), field: "address", current: item.center.decoderAddress ?? "–",
      desired: item.railkeeper.decoderAddress ?? "–" },
    { label: t("digitalCenters.field.protocol"), field: "protocol", current: item.center.protocol ?? "–",
      desired: item.railkeeper.protocol ?? "–" }
  ];
}

function validExpiry(value?: string) {
  if (!value) return false;
  const expiresAt = Date.parse(value);
  return Number.isFinite(expiresAt) && expiresAt > Date.now();
}

function fieldLabel(field: DigitalCenterWriteField, t: Translate) {
  return t(`digitalCenters.field.${field}`);
}

function resultLabel(confirmation: DigitalCenterWriteConfirmation, t: Translate) {
  return confirmation.result === "verification_failed" ? t("digitalCenters.write.verificationFailed") :
    t("digitalCenters.write.failed");
}

function formatDateTime(value: string, language: Language) {
  const parsed = new Date(value);
  return Number.isNaN(parsed.getTime()) ? value : parsed.toLocaleString(language === "de" ? "de-DE" : "en-GB");
}

type Translate = (key: string, values?: Record<string, string | number>) => string;
