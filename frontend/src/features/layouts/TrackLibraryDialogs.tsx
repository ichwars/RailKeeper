import { KeyboardEvent, useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";

import type {
  TrackGeometryLibrary,
  TrackLibraryImportPreview,
  TrackLibraryPackage
} from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppFilePicker } from "../../shared/ui/AppFilePicker";
import { AppTextArea } from "../../shared/ui/AppTextArea";

const focusableSelector = "button:not([disabled]), input:not([disabled]), textarea:not([disabled]), " +
  "[tabindex]:not([tabindex='-1'])";

function trapDialogFocus(event: KeyboardEvent<HTMLDivElement>, layer: HTMLDivElement | null, close: () => void) {
  if (event.key === "Escape") {
    event.preventDefault();
    close();
    return;
  }
  if (event.key !== "Tab") return;
  const focusable = Array.from(layer?.querySelectorAll<HTMLElement>(focusableSelector) || []);
  if (focusable.length === 0) return;
  const first = focusable[0];
  const last = focusable[focusable.length - 1];
  if ((!event.shiftKey && event.target === last) || (event.shiftKey && event.target === first)) {
    event.preventDefault();
    (event.shiftKey ? last : first).focus();
  }
}

export function TrackLibraryImportDialog({ busy, message, onPreview, onImport, onClose }: {
  busy: boolean;
  message: string;
  onPreview: (doc: TrackLibraryPackage) => Promise<TrackLibraryImportPreview>;
  onImport: (doc: TrackLibraryPackage) => void | Promise<void>;
  onClose: () => void;
}) {
  const [file, setFile] = useState<File | null>(null);
  const [fileError, setFileError] = useState("");
  const [preview, setPreview] = useState<TrackLibraryImportPreview | null>(null);
  const [previewing, setPreviewing] = useState(false);
  const layerRef = useRef<HTMLDivElement | null>(null);
  const { t } = useI18n();

  useEffect(() => {
    layerRef.current?.querySelector<HTMLButtonElement>("button")?.focus();
  }, []);

  const chooseFile = async (next: File | null) => {
    setFile(next); setFileError(""); setPreview(null);
    if (!next) return;
    setPreviewing(true);
    try {
      const parsed: unknown = JSON.parse(await next.text());
      if (!isTrackLibraryPackage(parsed)) throw new Error(t("layouts.trackLibraries.invalidFile"));
      setPreview(await onPreview(parsed));
    } catch (reason) {
      setFileError(reason instanceof Error ? reason.message : t("layouts.trackLibraries.invalidFile"));
    } finally {
      setPreviewing(false);
    }
  };

  const close = () => { if (!busy && !previewing) onClose(); };
  return createPortal(<div ref={layerRef} className="confirm-layer track-library-dialog-layer" role="dialog"
    aria-modal="true" aria-label={t("layouts.trackLibraries.importTitle")}
    onKeyDown={(event) => trapDialogFocus(event, layerRef.current, close)}>
    <section className="panel track-library-dialog">
      <h2>{t("layouts.trackLibraries.importTitle")}</h2>
      <p>{t("layouts.trackLibraries.importHelp")}</p>
      <AppFilePicker label={t("layouts.trackLibraries.file")} accept=".json,application/json" file={file}
        disabled={busy || previewing} onFileChange={(next) => void chooseFile(next)}
        triggerLabel={t("layouts.trackLibraries.chooseFile")} clearLabel={t("common.delete")}
        emptyLabel={t("layouts.trackLibraries.noFile")} error={fileError || undefined} />
      {previewing ? <p className="layout-empty">{t("layouts.trackLibraries.previewing")}</p> : null}
      {preview ? <section className="track-library-preview" aria-label={t("layouts.trackLibraries.preview") }>
        <div><strong>{preview.package.library.manufacturer}</strong>
          <span>{preview.package.library.trackSystem} · {preview.package.library.gauge} ·
            {` ${preview.package.library.scale}`}</span></div>
        <dl><div><dt>{t("layouts.trackLibraries.version")}</dt><dd>{preview.package.library.version}</dd></div>
          <div><dt>{t("layouts.trackLibraries.definitions")}</dt><dd>{preview.definitionCount}</dd></div></dl>
        {preview.warnings.map((warning) => <p className="track-library-warning" key={warning}>
          {t(`layouts.trackLibraries.warning.${warning}`)}</p>)}
        {preview.conflict ? <p className="form-message">{t("layouts.trackLibraries.conflict")}</p> : null}
      </section> : null}
      {message ? <p className="form-message">{message}</p> : null}
      <div className="layout-form-actions">
        <button type="button" className="secondary-button" onClick={close} disabled={busy || previewing}>
          {t("common.cancel")}</button>
        <button type="button" className="primary-button" disabled={busy || !preview?.canImport}
          onClick={() => preview && void onImport(preview.package)}>
          {busy ? t("common.saving") : t("layouts.trackLibraries.importConfirm")}</button>
      </div>
    </section>
  </div>, document.body);
}

export function TrackLibraryReviewDialog({ library, busy, message, onSubmit, onClose }: {
  library: TrackGeometryLibrary;
  busy: boolean;
  message: string;
  onSubmit: (note: string) => void | Promise<void>;
  onClose: () => void;
}) {
  const [note, setNote] = useState("");
  const layerRef = useRef<HTMLDivElement | null>(null);
  const { t } = useI18n();
  const close = () => { if (!busy) onClose(); };
  return createPortal(<div ref={layerRef} className="confirm-layer track-library-dialog-layer" role="dialog"
    aria-modal="true" aria-label={t("layouts.trackLibraries.verifyTitle")}
    onKeyDown={(event) => trapDialogFocus(event, layerRef.current, close)}>
    <section className="panel track-library-dialog">
      <h2>{t("layouts.trackLibraries.verifyTitle")}</h2>
      <p>{t("layouts.trackLibraries.verifyHelp", { name: `${library.manufacturer} ${library.trackSystem}` })}</p>
      <AppTextArea autoFocus required maxLength={500} label={t("layouts.trackLibraries.verificationNote")}
        value={note} onChange={(event) => setNote(event.target.value)} />
      {message ? <p className="form-message">{message}</p> : null}
      <div className="layout-form-actions">
        <button type="button" className="secondary-button" onClick={close} disabled={busy}>
          {t("common.cancel")}</button>
        <button type="button" className="primary-button" disabled={busy || !note.trim()}
          onClick={() => void onSubmit(note.trim())}>
          {busy ? t("common.saving") : t("layouts.trackLibraries.verify")}</button>
      </div>
    </section>
  </div>, document.body);
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function isStatus(value: unknown): value is "draft" | "verified" | "retired" {
  return value === "draft" || value === "verified" || value === "retired";
}

function isTrackLibraryPackage(value: unknown): value is TrackLibraryPackage {
  if (!isRecord(value) || value.format !== "railkeeper.track-library" || value.schemaVersion !== 1 ||
    !isRecord(value.library) || !Array.isArray(value.definitions)) return false;
  const library = value.library;
  const texts = [library.manufacturer, library.trackSystem, library.gauge, library.scale,
    library.version, library.sourceUrl];
  return texts.every((entry) => typeof entry === "string") && isStatus(library.status) &&
    value.definitions.every((definition) => isRecord(definition) &&
      typeof definition.articleNumber === "string" && typeof definition.name === "string" &&
      typeof definition.kind === "string" && typeof definition.lengthMm === "number" &&
      typeof definition.sourceUrl === "string" && isStatus(definition.status) &&
      isRecord(definition.geometry) && definition.geometry.schemaVersion === 1 &&
      Array.isArray(definition.geometry.ports) && Array.isArray(definition.geometry.routes));
}
