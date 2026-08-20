import { Download, Folder, FolderOpen } from "lucide-react";
import { useState } from "react";

import type { Language } from "../../shared/i18n";
import type { DataTransferArtifact, DataTransferSummary } from "./dataTransferModel";

type Translate = (key: string, values?: Record<string, string | number>) => string;

type TransferArtifactPanelProps = {
  artifacts: DataTransferArtifact[];
  canOpenFolder: boolean;
  downloadUrl: (artifactId: string) => string;
  language: Language;
  mutating: boolean;
  onOpenFolder: () => Promise<unknown>;
  summary: DataTransferSummary;
  t: Translate;
};

export function TransferArtifactPanel({
  artifacts,
  canOpenFolder,
  downloadUrl,
  language,
  mutating,
  onOpenFolder,
  summary,
  t
}: TransferArtifactPanelProps) {
  const [error, setError] = useState("");

  function openFolder() {
    setError("");
    void onOpenFolder().catch(() => setError(t("importExport.dashboard.storage.openError")));
  }

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
      {artifacts.length > 0 && (
        <div className="transfer-artifact-links">
          {artifacts.map((artifact) => (
            <a
              href={downloadUrl(artifact.id)}
              key={artifact.id}
              aria-label={`${artifact.displayName} ${t("importExport.dashboard.storage.download")}`}
              title={artifact.displayName}
            >
              <Download size={14} aria-hidden="true" />
              <span>{artifact.displayName}</span>
              <small>{formatBytes(artifact.sizeBytes, language)}</small>
            </a>
          ))}
        </div>
      )}
      {error && <p className="form-message error transfer-storage-error" role="alert">{error}</p>}
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
