import { Archive, BookOpen, Download, ShieldCheck, Upload } from "lucide-react";
import { useCallback, useEffect, useState } from "react";

import { api, type TrackGeometryLibrary, type TrackLibraryPackage } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { LayoutConfirmDialog, type LayoutPendingAction } from "./LayoutConfirmDialog";
import { TrackLibraryImportDialog, TrackLibraryReviewDialog } from "./TrackLibraryDialogs";

export function TrackLibraryPanel({ canManage = false }: { canManage?: boolean }) {
  const [libraries, setLibraries] = useState<TrackGeometryLibrary[]>([]);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");
  const [dialogMessage, setDialogMessage] = useState("");
  const [importOpen, setImportOpen] = useState(false);
  const [reviewLibrary, setReviewLibrary] = useState<TrackGeometryLibrary | null>(null);
  const [pending, setPending] = useState<LayoutPendingAction | null>(null);
  const { t } = useI18n();
  const genericError = t("layouts.error.generic");

  const load = useCallback(async () => {
    setLoading(true); setMessage("");
    try { setLibraries(await api.trackLibraries()); }
    catch (reason) { setMessage(reason instanceof Error ? reason.message : genericError); }
    finally { setLoading(false); }
  }, [genericError]);

  useEffect(() => { void load(); }, [load]);

  const importLibrary = async (doc: TrackLibraryPackage) => {
    setBusy(true); setDialogMessage("");
    try {
      await api.importTrackLibrary({ confirmed: true, package: doc });
      setImportOpen(false); await load();
    } catch (reason) {
      setDialogMessage(reason instanceof Error ? reason.message : t("layouts.error.generic"));
    } finally { setBusy(false); }
  };

  const verify = async (note: string) => {
    if (!reviewLibrary) return;
    setBusy(true); setDialogMessage("");
    try {
      await api.updateTrackLibraryStatus(reviewLibrary.id, {
        confirmed: true, status: "verified", verificationNote: note
      });
      setReviewLibrary(null); await load();
    } catch (reason) {
      setDialogMessage(reason instanceof Error ? reason.message : t("layouts.error.generic"));
    } finally { setBusy(false); }
  };

  const askRetire = (library: TrackGeometryLibrary) => setPending({
    title: t("layouts.trackLibraries.retireTitle"), dangerous: true,
    body: t("layouts.trackLibraries.retireBody", { name: `${library.manufacturer} ${library.trackSystem}` }),
    confirmLabel: t("layouts.trackLibraries.retire"),
    run: async () => {
      await api.updateTrackLibraryStatus(library.id, {
        confirmed: true, status: "retired", verificationNote: ""
      });
      await load();
    }
  });

  const download = async (library: TrackGeometryLibrary) => {
    setMessage("");
    try {
      const doc = await api.exportTrackLibrary(library.id);
      const blob = new Blob([JSON.stringify(doc, null, 2)], { type: "application/json" });
      const url = URL.createObjectURL(blob);
      const anchor = document.createElement("a");
      anchor.href = url;
      anchor.download = trackLibraryFileName(library);
      anchor.click();
      URL.revokeObjectURL(url);
    } catch (reason) { setMessage(reason instanceof Error ? reason.message : t("layouts.error.generic")); }
  };

  return <section className="panel track-library-panel">
    <div className="layout-panel-head"><div className="panel-title"><BookOpen size={17} />
      <div><h3>{t("layouts.trackLibraries.title")}</h3><p>{t("layouts.trackLibraries.subtitle")}</p></div>
    </div>{canManage ? <button type="button" className="secondary-button compact-action"
      onClick={() => { setDialogMessage(""); setImportOpen(true); }}>
      <Upload size={14} />{t("layouts.trackLibraries.import")}</button> : null}</div>
    {message ? <p className="form-message">{message}</p> : null}
    {loading ? <p className="layout-empty">{t("layouts.trackLibraries.loading")}</p> :
      libraries.length === 0 ? <p className="layout-empty">{t("layouts.trackLibraries.empty")}</p> :
        <div className="track-library-list">{libraries.map((library) => <article key={library.id}
          className="track-library-card">
          <div className="track-library-main"><div><strong>{library.manufacturer} · {library.trackSystem}</strong>
            <span>{library.gauge} · {library.scale} · {t("layouts.trackLibraries.version")} {library.version}</span></div>
            <span className={`status-pill revision-${library.status}`}>
              {t(`layouts.trackLibraries.status.${library.status}`)}</span></div>
          <dl><div><dt>{t("layouts.trackLibraries.definitions")}</dt><dd>{library.definitionCount}</dd></div>
            <div><dt>{t("layouts.trackLibraries.source")}</dt><dd><a href={library.sourceUrl}
              target="_blank" rel="noreferrer">{t("layouts.trackLibraries.openSource")}</a></dd></div></dl>
          {library.verificationNote ? <p className="track-library-note">{library.verificationNote}</p> : null}
          <div className="layout-plan-actions">
            <button type="button" className="secondary-button compact-action" onClick={() => void download(library)}>
              <Download size={14} />{t("layouts.trackLibraries.export")}</button>
            {canManage && library.status === "draft" ? <button type="button"
              className="primary-button compact-action" onClick={() => { setDialogMessage(""); setReviewLibrary(library); }}>
              <ShieldCheck size={14} />{t("layouts.trackLibraries.verify")}</button> : null}
            {canManage && library.status === "verified" ? <button type="button"
              className="secondary-button compact-action" onClick={() => askRetire(library)}>
              <Archive size={14} />{t("layouts.trackLibraries.retire")}</button> : null}
          </div>
        </article>)}</div>}
    {importOpen ? <TrackLibraryImportDialog busy={busy} message={dialogMessage}
      onPreview={(doc) => api.previewTrackLibraryImport(doc)} onImport={importLibrary}
      onClose={() => setImportOpen(false)} /> : null}
    {reviewLibrary ? <TrackLibraryReviewDialog library={reviewLibrary} busy={busy} message={dialogMessage}
      onSubmit={verify} onClose={() => setReviewLibrary(null)} /> : null}
    <LayoutConfirmDialog action={pending} onClose={() => setPending(null)} />
  </section>;
}

function trackLibraryFileName(library: TrackGeometryLibrary) {
  const stem = `${library.manufacturer}-${library.trackSystem}-${library.version}`
    .toLocaleLowerCase("de-DE").replace(/[^a-z0-9]+/g, "-").replace(/^-|-$/g, "");
  return `${stem || "track-library"}.railkeeper.json`;
}
