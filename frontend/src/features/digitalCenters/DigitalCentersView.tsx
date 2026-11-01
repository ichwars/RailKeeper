import { useEffect, useRef, useState } from "react";

import { DigitalCenterList } from "./DigitalCenterList";
import { useI18n } from "../../shared/i18n";
import {
  isDigitalCenterVehicleAdoptionProvider,
  openDigitalCenterVehicleDraft
} from "./digitalCenterVehicleAdoption";
import { DigitalCenterToolbar } from "./DigitalCenterToolbar";
import { DigitalStatusPanel } from "./DigitalStatusPanel";
import { LocomotiveComparisonDialog } from "./LocomotiveComparisonDialog";
import { LocomotiveWorklist } from "./LocomotiveWorklist";
import { useDigitalCenterVehicleAdoption } from "./useDigitalCenterVehicleAdoption";
import { useDigitalCentersWorkspace } from "./useDigitalCentersWorkspace";
import { VehicleAssignmentDialog } from "./VehicleAssignmentDialog";
import { readECoSStoredCVs } from "./readECoSStoredCVs";
import type { DigitalCenterWorkItem } from "./digitalCenterModel";

function openDigitalSettings() {
  window.history.pushState(null, "", "/settings?tab=digital");
  window.dispatchEvent(new PopStateEvent("popstate"));
}

export function DigitalCentersView({ roles }: { roles: string[] }) {
  const { t } = useI18n();
  const workspace = useDigitalCentersWorkspace();
  const returnReadStartedRef = useRef(false);
  const cvCancelled = useRef(false);
  const cvPreparing = useRef(false);
  const [preparingCVs, setPreparingCVs] = useState(false);
  const [cvReadError, setCVReadError] = useState("");
  const adoption = useDigitalCenterVehicleAdoption({
    onAssigned: async () => {
      workspace.closeDialog();
      workspace.closeDetail();
      await workspace.readData();
    }
  });
  const adoptionProvider = isDigitalCenterVehicleAdoptionProvider(workspace.readSession?.provider)
    ? workspace.readSession.provider
    : null;

  useEffect(() => () => { cvCancelled.current = true; }, [workspace.selectedItem?.id, workspace.readSession?.id]);

  const prepareAdoption = async (item: DigitalCenterWorkItem) => {
    if (adoptionProvider !== "ecos") return item;
    if (cvPreparing.current) return null;
    cvPreparing.current = true;
    cvCancelled.current = false;
    setPreparingCVs(true);
    setCVReadError("");
    try {
      return await readECoSStoredCVs(item, workspace.acceptCVRead,
        { useSaved: true, isCancelled: () => cvCancelled.current });
    } catch {
      if (!cvCancelled.current) setCVReadError(t("digitalCenters.cvs.readError"));
      return null;
    } finally {
      cvPreparing.current = false;
      setPreparingCVs(false);
    }
  };

  useEffect(() => {
    if (returnReadStartedRef.current || workspace.loading.workspace) return;
    const search = new URLSearchParams(window.location.search);
    const sessionID = search.get("sessionId")?.trim();
    if (!sessionID) return;
    if (search.get("review") === "backup" || !search.get("objectId")?.trim()) {
      returnReadStartedRef.current = true;
      void workspace.loadSession(sessionID).catch(() => undefined);
      return;
    }
    if (!workspace.actions.canRead) return;
    returnReadStartedRef.current = true;
    window.history.replaceState(null, "", "/digital-centers");
    void workspace.readData().catch(() => undefined);
  }, [workspace.actions.canRead, workspace.loadSession, workspace.loading.workspace, workspace.readData]);

  return (
    <section className="digital-centers-workspace" data-can-administer={roles.includes("Admin")}>
      <header className="digital-centers-page-head">
        <p className="eyebrow">{t("digitalCenters.eyebrow")}</p>
        <h1>{t("digitalCenters.title")}</h1>
        <p>{t("digitalCenters.subtitle")}</p>
      </header>

      {workspace.reviewingSnapshot && <p className="digital-workspace-snapshot-note" role="status">
        {t("digitalCenters.backup.reviewHint")} {" "}
        <a href="/digital-centers">{t("digitalCenters.backup.closeReview")}</a>
      </p>}

      <DigitalCenterToolbar
        centers={workspace.centers}
        selectedProvider={workspace.selectedProvider}
        selectedCenter={workspace.selectedCenter}
        liveStatus={workspace.liveStatus}
        actions={workspace.actions}
        loading={workspace.loading}
        readError={workspace.errors.read}
        onSelectCenter={workspace.selectCenter}
        onRead={workspace.readData}
        onStartLive={workspace.startLive}
        onStopLive={workspace.stopLive}
        onConfigure={openDigitalSettings}
      />

      {workspace.errors.write && <p className="digital-workspace-operation-error" role="alert"
        aria-label={workspace.errors.write.includes(t("digitalCenters.error.readAgainMarker"))
          ? t("digitalCenters.error.writeConflict") : t("digitalCenters.error.write")}>
        {workspace.errors.write}
      </p>}

      <div className="digital-centers-layout" data-testid="digital-centers-layout">
        <DigitalCenterList
          centers={workspace.centers}
          selectedProvider={workspace.selectedProvider}
          total={workspace.sessionTotal}
          loading={workspace.loading.workspace}
          error={workspace.errors.workspace}
          onSelect={workspace.selectCenter}
          onConfigure={openDigitalSettings}
          onRetry={workspace.refresh}
        />
        <LocomotiveWorklist
          page={workspace.workItems}
          search={workspace.search}
          compareStatus={workspace.compareStatus}
          loading={workspace.loading.read || workspace.loading.worklist}
          error={workspace.errors.worklist}
          onSearch={workspace.setSearch}
          onCompareStatus={workspace.setCompareStatus}
          onPage={workspace.setPage}
          onPageSize={workspace.setPageSize}
          onRefresh={workspace.reviewingSnapshot && workspace.readSession
            ? () => workspace.loadSession(workspace.readSession!.id) : workspace.readData}
          onCompare={(itemID) => {
            workspace.selectItem(itemID);
            workspace.openDialog("comparison", itemID);
          }}
        />
        <DigitalStatusPanel
          tab={workspace.tab}
          onTab={workspace.setTab}
          selectedCenter={workspace.selectedCenter}
          liveStatus={workspace.liveStatus}
          messages={workspace.messages}
          actions={workspace.actions}
          loading={workspace.loading.live}
          errors={{ live: workspace.errors.live, messages: workspace.errors.messages }}
        />
      </div>

      {workspace.dialog?.kind === "comparison" && workspace.selectedItem && (
        <LocomotiveComparisonDialog item={workspace.selectedItem}
          canReadCVs={roles.includes("Admin") && adoptionProvider === "ecos"}
          onCVRead={(item) => { workspace.acceptCVRead(item); setCVReadError(""); }}
          preparingCVs={preparingCVs}
          cvReadError={cvReadError}
          onCancelCVRead={() => { cvCancelled.current = true; }}
          canWrite={workspace.actions.canWrite}
          canAdopt={roles.includes("Admin") && adoptionProvider !== null && !workspace.reviewingSnapshot}
          loading={workspace.loading.write}
          preview={workspace.writePreview}
          confirmation={workspace.writeConfirmation}
          error={workspace.errors.write}
          onPreview={workspace.previewWrite}
          onConfirm={workspace.confirmWrite}
          onCreateVehicle={async () => {
            if (adoptionProvider) {
              const ready = await prepareAdoption(workspace.selectedItem!);
              if (ready) openDigitalCenterVehicleDraft(ready, adoptionProvider);
            }
          }}
          onAssignVehicle={async () => {
            const ready = await prepareAdoption(workspace.selectedItem!);
            if (!ready) return;
            workspace.openDialog("assignment", ready.id);
            void adoption.commands.load(ready);
          }}
          onClose={() => {
            cvCancelled.current = true;
            setCVReadError("");
            workspace.closeDialog();
            workspace.closeDetail();
          }} />
      )}

      {workspace.dialog?.kind === "assignment" && workspace.selectedItem && roles.includes("Admin") &&
        adoptionProvider && (
        <VehicleAssignmentDialog
          item={workspace.selectedItem}
          provider={adoptionProvider}
          vehicles={adoption.state.vehicles}
          selectedVehicleId={adoption.state.selectedVehicleId}
          loading={adoption.state.loading}
          saving={adoption.state.saving}
          error={adoption.state.error}
          onSelect={adoption.setters.setSelectedVehicleId}
          onAssign={(vehicleId) => adoption.commands.assign(
            workspace.selectedItem!,
            adoptionProvider,
            vehicleId
          )}
          onClose={() => {
            adoption.commands.reset();
            workspace.closeDialog();
            workspace.closeDetail();
          }}
        />
      )}
    </section>
  );
}
