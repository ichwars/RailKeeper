import { useEffect, useRef, useState } from "react";
import { CheckCircle2, TriangleAlert, X } from "lucide-react";

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
  const [confirmed, setConfirmed] = useState(false);
  const closeButtonRef = useRef<HTMLButtonElement>(null);
  const { anchorRef, layerRef, onKeyDown } = useModalDialogLayer(onClose, closeButtonRef);
  const rows = comparisonRows(item);
  const changedFields = rows.filter((row) => row.current !== row.desired).map((row) => row.field);
  const grantValid = Boolean(preview?.token.trim()) && validExpiry(preview?.expiresAt);
  const verified = confirmation?.result === "verified" && confirmation.applied && confirmation.verified;

  useEffect(() => setConfirmed(false), [preview?.token]);

  return (
    <>
      <span ref={anchorRef} aria-hidden="true" />
      <div ref={layerRef} className="digital-comparison-layer" role="dialog" aria-modal="true"
        aria-label={`Lok-Vergleich ${item.name}`} onKeyDown={onKeyDown}>
        <section className="digital-comparison-dialog">
          <header><div><p className="eyebrow">LOK-ABGLEICH</p><h2>{item.name}</h2></div>
            <button ref={closeButtonRef} type="button" className="digital-center-icon-button"
              aria-label="Vergleich schließen" onClick={onClose}>
              <X size={19} aria-hidden="true" />
            </button></header>

          {!preview && <ComparisonTable rows={rows} />}
          {preview && <WritePreview preview={preview} />}
          {error && <p className="digital-write-result error" role="alert" aria-label="Schreibfehler">
            <TriangleAlert size={17} aria-hidden="true" />{error}
          </p>}
          {confirmation && <p className={`digital-write-result ${verified ? "verified" : "error"}`}>
            {verified ? <CheckCircle2 size={17} aria-hidden="true" /> :
              <TriangleAlert size={17} aria-hidden="true" />}
            <span><strong>{verified ? "Schreiben verifiziert" : resultLabel(confirmation)}</strong>
              <small>{confirmation.message}</small></span>
          </p>}

          {preview && !confirmation && <label className="digital-write-confirmation">
            <input type="checkbox" checked={confirmed}
              onChange={(event) => setConfirmed(event.target.checked)} />
            <span>Ich bestätige, dass die angezeigten Werte in die Digitalzentrale geschrieben werden.</span>
          </label>}

          <footer>
            {!preview && <button type="button" className="digital-center-button"
              disabled={!canWrite || loading || changedFields.length === 0}
              onClick={() => void onPreview(changedFields).catch(() => undefined)}>
              Schreibvorschau erstellen
            </button>}
            {preview && !confirmation && <button type="button" className="digital-center-button"
              disabled={!confirmed || !grantValid || loading}
              title={!grantValid ? "Die Schreibfreigabe ist nicht mehr gültig." : undefined}
              onClick={() => void onConfirm().catch(() => undefined)}>Änderungen schreiben</button>}
            <button type="button" className="digital-center-button" onClick={onClose}>Schließen</button>
          </footer>
          {!canWrite && <p className="digital-write-unavailable">
            Diese Digitalzentrale unterstützt keine Schreibbefehle.
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

function ComparisonTable({ rows }: { rows: ComparisonRow[] }) {
  return <table><thead><tr><th>Feld</th><th>Digitalzentrale</th><th>RailKeeper</th></tr></thead>
    <tbody>{rows.map((row) => <tr key={row.field}><th>{row.label}</th>
      <td title={String(row.current)}>{row.current}</td>
      <td title={String(row.desired)}>{row.desired}</td></tr>)}</tbody></table>;
}

function WritePreview({ preview }: { preview: DigitalCenterWritePreview }) {
  return <section className="digital-write-preview" aria-label="Schreibvorschau">
    <p><strong>Richtung</strong><span>RailKeeper → Digitalzentrale</span></p>
    <table><thead><tr><th>Feld</th><th>Aktuell</th><th>Gewünscht</th></tr></thead>
      <tbody>{preview.changes.map((change) => <tr key={change.field}>
        <th>{fieldLabel(change.field)}</th><td>{change.current}</td><td>{change.desired}</td>
      </tr>)}</tbody></table>
    <small>Freigabe gültig bis {formatDateTime(preview.expiresAt)}</small>
  </section>;
}

function comparisonRows(item: DigitalCenterWorkItem): ComparisonRow[] {
  return [
    { label: "Name", field: "name", current: item.center.name ?? "–", desired: item.railkeeper.name ?? "–" },
    { label: "Decoder-Adresse", field: "address", current: item.center.decoderAddress ?? "–",
      desired: item.railkeeper.decoderAddress ?? "–" },
    { label: "Protokoll", field: "protocol", current: item.center.protocol ?? "–",
      desired: item.railkeeper.protocol ?? "–" }
  ];
}

function validExpiry(value?: string) {
  if (!value) return false;
  const expiresAt = Date.parse(value);
  return Number.isFinite(expiresAt) && expiresAt > Date.now();
}

function fieldLabel(field: DigitalCenterWriteField) {
  if (field === "address") return "Decoder-Adresse";
  if (field === "protocol") return "Protokoll";
  return "Name";
}

function resultLabel(confirmation: DigitalCenterWriteConfirmation) {
  return confirmation.result === "verification_failed" ? "Schreibprüfung fehlgeschlagen" :
    "Schreiben fehlgeschlagen";
}

function formatDateTime(value: string) {
  const parsed = new Date(value);
  return Number.isNaN(parsed.getTime()) ? value : parsed.toLocaleString("de-DE");
}
