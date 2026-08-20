import { Plus } from "lucide-react";

import { useI18n } from "../../shared/i18n";
import { TransferArtifactPanel } from "./TransferArtifactPanel";
import { TransferHistoryTable } from "./TransferHistoryTable";
import { TransferJobDetails } from "./TransferJobDetails";
import { TransferJobList } from "./TransferJobList";
import { TransferProfilesTable } from "./TransferProfilesTable";
import { TransferSummaryStrip } from "./TransferSummaryStrip";
import { useDataTransferWorkspace } from "./useDataTransferWorkspace";

export function ImportExportView({ roles }: { roles: string[] }) {
  const { language, t } = useI18n();
  const workspace = useDataTransferWorkspace(roles);
  const detailJob = workspace.selectedJobDetails?.job ?? workspace.selectedJob;

  return (
    <div className="data-transfer-workspace">
      <header className="data-transfer-page-head">
        <div>
          <p className="eyebrow">{t("importExport.dashboard.eyebrow")}</p>
          <h1>{t("importExport.title")}</h1>
          <p>{t("importExport.dashboard.subtitle")}</p>
        </div>
        <div className="data-transfer-page-actions">
          <button
            type="button"
            className="primary-button"
            disabled={!workspace.capabilities.canImport || workspace.mutating}
            onClick={() => workspace.openDialog("import")}
          >
            <Plus size={17} aria-hidden="true" />
            {t("importExport.dashboard.newImport")}
          </button>
          <button
            type="button"
            className="secondary-button"
            disabled={!workspace.capabilities.canExport || workspace.mutating}
            onClick={() => workspace.openDialog("export")}
          >
            <Plus size={17} aria-hidden="true" />
            {t("importExport.dashboard.newExport")}
          </button>
        </div>
      </header>

      {(workspace.error || workspace.detailError) && (
        <p className="form-message error data-transfer-message" role="alert">
          {workspace.error || workspace.detailError}
        </p>
      )}

      <TransferSummaryStrip language={language} summary={workspace.summary} t={t} />

      <div className="data-transfer-dashboard">
        <TransferJobList
          allJobs={workspace.allJobs}
          filters={workspace.filters}
          jobs={workspace.jobs}
          language={language}
          loading={workspace.loading}
          onFilter={(filter) => workspace.setFilters(filter)}
          onSelect={workspace.selectJob}
          selectedJobId={workspace.selectedJobId}
          t={t}
        />

        <div className="data-transfer-center">
          <TransferProfilesTable
            canCreate={workspace.capabilities.canCreateProfiles}
            canEdit={workspace.capabilities.canUpdateProfiles}
            language={language}
            mutating={workspace.mutating}
            onCreate={() => workspace.openDialog("profile")}
            onEdit={(profileId) => workspace.openDialog("profile", profileId)}
            onRun={(profileId) => workspace.openDialog("export", profileId)}
            profiles={workspace.profiles}
            t={t}
          />
          <TransferHistoryTable
            jobs={workspace.allJobs}
            language={language}
            onSelect={workspace.selectJob}
            selectedJobId={workspace.selectedJobId}
            t={t}
          />
        </div>

        <div className="data-transfer-details">
          <TransferJobDetails
            canImport={workspace.capabilities.canImport}
            detailLoading={workspace.detailLoading}
            job={detailJob}
            language={language}
            mutating={workspace.mutating}
            onConfirm={workspace.confirmImport}
            onContinue={(profileId) => workspace.openDialog("import", profileId)}
            onRetry={workspace.retryJob}
            t={t}
          />
          <TransferArtifactPanel
            artifacts={workspace.selectedJobDetails?.artifacts ?? []}
            canOpenFolder={workspace.capabilities.canOpenFolder}
            downloadUrl={workspace.artifactDownloadUrl}
            language={language}
            mutating={workspace.mutating}
            onOpenFolder={workspace.openArtifactFolder}
            summary={workspace.summary}
            t={t}
          />
        </div>
      </div>
    </div>
  );
}
