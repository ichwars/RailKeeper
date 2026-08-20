import { Plus } from "lucide-react";

import { useI18n } from "../../shared/i18n";
import { TransferArtifactPanel } from "./TransferArtifactPanel";
import { TransferExportDialog } from "./TransferExportDialog";
import { TransferHistoryTable } from "./TransferHistoryTable";
import { TransferImportDialog } from "./TransferImportDialog";
import { TransferJobDetails } from "./TransferJobDetails";
import { TransferJobList } from "./TransferJobList";
import { TransferProfileDialog } from "./TransferProfileDialog";
import { TransferProfilesTable } from "./TransferProfilesTable";
import { TransferSummaryStrip } from "./TransferSummaryStrip";
import { useDataTransferWorkspace } from "./useDataTransferWorkspace";

export function ImportExportView({ roles }: { roles: string[] }) {
  const { language, t } = useI18n();
  const workspace = useDataTransferWorkspace(roles);
  const detailJob = workspace.selectedJobDetails?.job ?? workspace.selectedJob;
  const dialogJobId = workspace.dialog?.kind !== "profile" ? workspace.dialog?.jobId : undefined;
  const dialogJob = dialogJobId
    ? [workspace.selectedJobDetails?.job, workspace.selectedJob, ...workspace.allJobs]
      .find((job) => job?.id === dialogJobId) || undefined
    : undefined;

  async function retryJob(jobId: string) {
    const retry = await workspace.retryJob(jobId);
    workspace.openDialog(retry.direction, retry.profileId || undefined, retry.id);
  }

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
            onContinue={(job) => workspace.openDialog("import", job.profileId || undefined, job.id)}
            onRetry={retryJob}
            t={t}
          />
          <TransferArtifactPanel
            artifacts={workspace.selectedJobDetails?.artifacts ?? []}
            canDelete={workspace.capabilities.canDeleteArtifacts}
            canOpenFolder={workspace.capabilities.canOpenFolder}
            downloadUrl={workspace.artifactDownloadUrl}
            language={language}
            mutating={workspace.mutating}
            onDelete={workspace.deleteArtifact}
            onOpenFolder={workspace.openArtifactFolder}
            summary={workspace.summary}
            t={t}
          />
        </div>
      </div>

      {workspace.dialog?.kind === "import" ? (
        <TransferImportDialog
          initialDetails={dialogJob?.id === workspace.selectedJobDetails?.job.id ? workspace.selectedJobDetails : null}
          initialJob={dialogJob}
          initialProfileId={workspace.dialog.profileId}
          language={language}
          onCancelJob={workspace.cancelImport}
          onClose={workspace.closeDialog}
          onConfirm={workspace.confirmImport}
          onCreateJob={workspace.createImportJob}
          onResolve={workspace.resolveIssue}
          onUpload={workspace.uploadImportFile}
          profiles={workspace.profiles}
        />
      ) : null}
      {workspace.dialog?.kind === "export" ? (
        <TransferExportDialog
          downloadUrl={workspace.artifactDownloadUrl}
          initialJob={dialogJob}
          initialProfileId={workspace.dialog.profileId}
          language={language}
          onClose={workspace.closeDialog}
          onCreateJob={workspace.createExportJob}
          onExecute={workspace.executeExportJob}
          profiles={workspace.profiles}
        />
      ) : null}
      {workspace.dialog?.kind === "profile" ? (
        <TransferProfileDialog
          availableAreas={workspace.availableAreas}
          canDisable={workspace.capabilities.canDisableProfiles}
          language={language}
          onClose={workspace.closeDialog}
          onCreate={workspace.createProfile}
          onDisable={workspace.disableProfile}
          onUpdate={workspace.updateProfile}
          profile={workspace.profiles.find((profile) => profile.id === workspace.dialog?.profileId)}
        />
      ) : null}
    </div>
  );
}
