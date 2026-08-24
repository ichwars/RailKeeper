import { Plus } from "lucide-react";
import { useState } from "react";

import { useI18n } from "../../shared/i18n";
import { TransferArtifactPanel } from "./TransferArtifactPanel";
import { TransferConfirmDialog, type TransferPendingAction } from "./TransferConfirmDialog";
import { TransferExportDialog } from "./TransferExportDialog";
import { TransferHistoryTable } from "./TransferHistoryTable";
import { TransferImportDialog } from "./TransferImportDialog";
import { TransferJobDetails } from "./TransferJobDetails";
import { TransferJobList } from "./TransferJobList";
import { TransferProfileDialog } from "./TransferProfileDialog";
import { TransferProfilesTable } from "./TransferProfilesTable";
import { TransferSummaryStrip } from "./TransferSummaryStrip";
import type { DataTransferJob } from "./dataTransferModel";
import { useDataTransferWorkspace } from "./useDataTransferWorkspace";

export function ImportExportView({ roles }: { roles: string[] }) {
  const { language, t } = useI18n();
  const workspace = useDataTransferWorkspace(roles);
  const [jobPendingDelete, setJobPendingDelete] = useState<DataTransferJob | null>(null);
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

  const deleteAction: TransferPendingAction | null = jobPendingDelete ? {
    title: t("importExport.dashboard.delete.title"),
    body: t("importExport.dashboard.delete.body", {
      name: jobPendingDelete.sourceName || jobPendingDelete.profileName
    }),
    confirmLabel: t("importExport.dashboard.delete.confirm"),
    dangerous: true,
    errorMessage: t("importExport.dashboard.delete.error"),
    run: () => workspace.deleteJob(jobPendingDelete.id)
  } : null;

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
          canDelete={workspace.capabilities.canDeleteJobs}
          filters={workspace.filters}
          jobs={workspace.jobs}
          language={language}
          loading={workspace.loading}
          mutating={workspace.mutating}
          onDeleteRequest={setJobPendingDelete}
          onFilter={(filter) => workspace.setFilters(filter)}
          onSelect={workspace.selectJob}
          selectedJobId={workspace.selectedJobId}
          t={t}
        />

        <div className="data-transfer-center">
          <TransferProfilesTable
            canCreate={workspace.capabilities.canCreateProfiles}
            canEdit={workspace.capabilities.canUpdateProfiles}
            canExport={workspace.capabilities.canExport}
            canImport={workspace.capabilities.canImport}
            language={language}
            mutating={workspace.mutating}
            onCreate={() => workspace.openDialog("profile")}
            onEdit={(profileId) => workspace.openDialog("profile", profileId)}
            onRun={(profile) => workspace.openDialog(profile.direction, profile.id)}
            profiles={workspace.profiles}
            t={t}
          />
          <TransferHistoryTable
            canDelete={workspace.capabilities.canDeleteJobs}
            jobs={workspace.allJobs}
            language={language}
            mutating={workspace.mutating}
            onDeleteRequest={setJobPendingDelete}
            onSelect={workspace.selectJob}
            selectedJobId={workspace.selectedJobId}
            t={t}
          />
        </div>

        <div className="data-transfer-details">
          <TransferJobDetails
            canExport={workspace.capabilities.canExport}
            canImport={workspace.capabilities.canImport}
            canRetry={(job) => job.direction === "import"
              ? workspace.capabilities.canRetryImport
              : workspace.capabilities.canRetryExport}
            detailLoading={workspace.detailLoading}
            job={detailJob}
            language={language}
            mutating={workspace.mutating}
            onContinue={(job) => workspace.openDialog(job.direction, job.profileId || undefined, job.id)}
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
          initialRequiresReupload={workspace.importRequiresReupload(dialogJob?.id)}
          language={language}
          onCancelJob={workspace.cancelImport}
          onClose={workspace.closeDialog}
          onConfirm={workspace.confirmImport}
          onCreateJob={workspace.createImportJob}
          onResolve={workspace.resolveIssue}
          onRefreshJob={workspace.refreshJobDetails}
          onUpload={workspace.uploadImportFile}
          profiles={workspace.profiles}
        />
      ) : null}
      {workspace.dialog?.kind === "export" ? (
        <TransferExportDialog
          canRetry={workspace.capabilities.canRetryExport}
          downloadUrl={workspace.artifactDownloadUrl}
          initialJob={dialogJob}
          initialProfileId={workspace.dialog.profileId}
          language={language}
          onClose={workspace.closeDialog}
          onCreateJob={workspace.createExportJob}
          onExecute={workspace.executeExportJob}
          onRefreshJob={workspace.refreshJobDetails}
          onRetry={workspace.retryJob}
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
          initialProfileId={workspace.dialog.profileId}
          profiles={workspace.profiles}
        />
      ) : null}
      <TransferConfirmDialog action={deleteAction} cancelLabel={t("importExport.dashboard.delete.cancel")}
        onClose={() => setJobPendingDelete(null)} />
    </div>
  );
}
