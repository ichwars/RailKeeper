import { useEffect, useRef } from "react";

import { DigitalCenterList } from "./DigitalCenterList";
import { useI18n } from "../../shared/i18n";
import { openDigitalCenterVehicleDraft } from "./digitalCenterVehicleAdoption";
import { DigitalCenterToolbar } from "./DigitalCenterToolbar";
import { DigitalStatusPanel } from "./DigitalStatusPanel";
import { LocomotiveComparisonDialog } from "./LocomotiveComparisonDialog";
import { LocomotiveWorklist } from "./LocomotiveWorklist";
import { useDigitalCenterVehicleAdoption } from "./useDigitalCenterVehicleAdoption";
import { useDigitalCentersWorkspace } from "./useDigitalCentersWorkspace";
import { VehicleAssignmentDialog } from "./VehicleAssignmentDialog";

function openDigitalSettings() {
  window.history.pushState(null, "", "/settings?tab=digital");
  window.dispatchEvent(new PopStateEvent("popstate"));
}

export function DigitalCentersView({ roles }: { roles: string[] }) {
  const { t } = useI18n();
  const workspace = useDigitalCentersWorkspace();
  const returnReadStartedRef = useRef(false);
  const adoption = useDigitalCenterVehicleAdoption({
    onAssigned: async () => {
      workspace.closeDialog();
      workspace.closeDetail();
      await workspace.readData();
    }
  });

  useEffect(() => {
    if (returnReadStartedRef.current || !workspace.actions.canRead) return;
    const search = new URLSearchParams(window.location.search);
    if (!search.get("sessionId")?.trim() || !search.get("objectId")?.trim()) return;
    returnReadStartedRef.current = true;
    window.history.replaceState(null, "", "/digital-centers");
    void workspace.readData().catch(() => undefined);
  }, [workspace.actions.canRead, workspace.readData]);

  return (
    <section className="digital-centers-workspace" data-can-administer={roles.includes("Admin")}>
      <header className="digital-centers-page-head">
        <p className="eyebrow">{t("digitalCenters.eyebrow")}</p>
        <h1>{t("digitalCenters.title")}</h1>
        <p>{t("digitalCenters.subtitle")}</p>
      </header>

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
          onRefresh={workspace.readData}
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
          canWrite={workspace.actions.canWrite}
          canAdopt={roles.includes("Admin")}
          loading={workspace.loading.write}
          preview={workspace.writePreview}
          confirmation={workspace.writeConfirmation}
          error={workspace.errors.write}
          onPreview={workspace.previewWrite}
          onConfirm={workspace.confirmWrite}
          onCreateVehicle={() => openDigitalCenterVehicleDraft(workspace.selectedItem!)}
          onAssignVehicle={() => {
            workspace.openDialog("assignment", workspace.selectedItem!.id);
            void adoption.commands.load(workspace.selectedItem!);
          }}
          onClose={() => {
            workspace.closeDialog();
            workspace.closeDetail();
          }} />
      )}

      {workspace.dialog?.kind === "assignment" && workspace.selectedItem && roles.includes("Admin") && (
        <VehicleAssignmentDialog
          item={workspace.selectedItem}
          vehicles={adoption.state.vehicles}
          selectedVehicleId={adoption.state.selectedVehicleId}
          loading={adoption.state.loading}
          saving={adoption.state.saving}
          error={adoption.state.error}
          onSelect={adoption.setters.setSelectedVehicleId}
          onAssign={(vehicleId) => adoption.commands.assign(workspace.selectedItem!, vehicleId)}
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
