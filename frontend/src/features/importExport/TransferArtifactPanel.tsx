import { Download, Folder, FolderOpen, Trash2 } from "lucide-react";
import { useState } from "react";

import type { Language } from "../../shared/i18n";
import type { DataTransferArtifact, DataTransferSummary } from "./dataTransferModel";
import { TransferConfirmDialog, type TransferPendingAction } from "./TransferConfirmDialog";

type Translate = (key: string, values?: Record<string, string | number>) => string;

type TransferArtifactPanelProps = {
  artifacts: DataTransferArtifact[];
  canDelete: boolean;
  canOpenFolder: boolean;
  downloadUrl: (artifactId: string) => string;
  language: Language;
  mutating: boolean;
  onOpenFolder: () => Promise<unknown>;
  onDelete: (artifactId: string) => Promise<unknown>;
  summary: DataTransferSummary;
  t: Translate;
};

export function TransferArtifactPanel({
  artifacts,
  canDelete,
  canOpenFolder,
  downloadUrl,
  language,
  mutating,
  onOpenFolder,
  onDelete,
  summary,
  t
}: TransferArtifactPanelProps) {
  const [error, setError] = useState("");
  const [deletingId, setDeletingId] = useState("");
  const [pendingArtifact, setPendingArtifact] = useState<DataTransferArtifact | null>(null);
  const activeArtifacts = artifacts.filter((artifact) => !artifact.deletedAt);

  function openFolder() {
    setError("");
    void onOpenFolder().catch(() => setError(t("importExport.dashboard.storage.openError")));
  }

  async function deleteArtifact(artifact: DataTransferArtifact) {
    setDeletingId(artifact.id);
    setError("");
    try {
      await onDelete(artifact.id);
    } finally {
      setDeletingId("");
    }
  }

  const deleteAction: TransferPendingAction | null = pendingArtifact ? {
    title: language === "de" ? "Exportdatei löschen?" : "Delete export file?",
    body: language === "de"
      ? `Exportdatei „${pendingArtifact.displayName}“ dauerhaft löschen?`
      : `Permanently delete export file “${pendingArtifact.displayName}”?`,
    confirmLabel: language === "de" ? "Datei löschen" : "Delete file",
    dangerous: true,
    errorMessage: language === "de"
      ? "Die Exportdatei konnte nicht gelöscht werden."
      : "The export file could not be deleted.",
    run: () => deleteArtifact(pendingArtifact)
  } : null;

  return (
    <section className="panel data-transfer-panel transfer-storage-panel">
      <header className="data-transfer-panel-head">
        <h2><Folder size={20} aria-hidden="true" />{t("importExport.dashboard.storage.title")}</h2>
      </header>
      <div className="transfer-storage-location">
        <FolderOpen size={22} aria-hidden="true" />
        <span>
          <strong title={summary.artifactDirectory}>
            {summary.artifactDirectory || t("importExport.dashboard.summary.local")}
          </strong>
          <small>
            {t("importExport.dashboard.storage.files", { count: summary.artifactCount })}
            {" · "}{formatBytes(summary.artifactBytes, language)}
          </small>
        </span>
        {canOpenFolder && (
          <button type="button" className="secondary-button data-transfer-small-action" disabled={mutating}
            onClick={openFolder}>
            <Folder size={15} aria-hidden="true" />{t("importExport.dashboard.storage.openFolder")}
          </button>
        )}
      </div>
      {activeArtifacts.length > 0 && (
        <div className="transfer-artifact-links">
          {activeArtifacts.map((artifact) => (
            <span className="transfer-artifact-row" key={artifact.id}>
              <a
                href={downloadUrl(artifact.id)}
                aria-label={`${artifact.displayName} ${t("importExport.dashboard.storage.download")}`}
                title={artifact.displayName}
              >
                <Download size={14} aria-hidden="true" />
                <span>{artifact.displayName}</span>
                <small>{formatBytes(artifact.sizeBytes, language)}</small>
              </a>
              {canDelete ? (
                <button type="button" className="icon-button transfer-artifact-delete"
                  aria-label={`${artifact.displayName} ${language === "de" ? "löschen" : "delete"}`}
                  disabled={mutating || deletingId === artifact.id} onClick={() => setPendingArtifact(artifact)}>
                  <Trash2 size={14} aria-hidden="true" />
                </button>
              ) : null}
            </span>
          ))}
        </div>
      )}
      {error && <p className="form-message error transfer-storage-error" role="alert">{error}</p>}
      <TransferConfirmDialog action={deleteAction} cancelLabel={language === "de" ? "Abbrechen" : "Cancel"}
        onClose={() => setPendingArtifact(null)} />
    </section>
  );
}

function formatBytes(value: number, language: Language) {
  if (value <= 0) return "0 MB";
  const megabytes = value / (1024 * 1024);
  return `${new Intl.NumberFormat(language === "de" ? "de-DE" : "en-GB", {
    maximumFractionDigits: megabytes < 10 ? 1 : 0
  }).format(megabytes)} MB`;
}
