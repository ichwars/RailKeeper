import { DigitalCenterList } from "./DigitalCenterList";
import { DigitalCenterToolbar } from "./DigitalCenterToolbar";
import { DigitalStatusPanel } from "./DigitalStatusPanel";
import { LocomotiveComparisonDialog } from "./LocomotiveComparisonDialog";
import { LocomotiveWorklist } from "./LocomotiveWorklist";
import { useDigitalCentersWorkspace } from "./useDigitalCentersWorkspace";

function openDigitalSettings() {
  window.history.pushState(null, "", "/settings?tab=digital");
  window.dispatchEvent(new PopStateEvent("popstate"));
}

export function DigitalCentersView({ roles }: { roles: string[] }) {
  const workspace = useDigitalCentersWorkspace();

  return (
    <section className="digital-centers-workspace" data-can-administer={roles.includes("Admin")}>
      <header className="digital-centers-page-head">
        <p className="eyebrow">DIGITALBETRIEB</p>
        <h1>Digitalzentralen</h1>
        <p>Zentralen, Live-Daten und Synchronisation in einer Arbeitsansicht.</p>
      </header>

      <DigitalCenterToolbar
        centers={workspace.centers}
        selectedProvider={workspace.selectedProvider}
        selectedCenter={workspace.selectedCenter}
        liveStatus={workspace.liveStatus}
        actions={workspace.actions}
        loading={workspace.loading}
        onSelectCenter={workspace.selectCenter}
        onRead={workspace.readData}
        onStartLive={workspace.startLive}
        onStopLive={workspace.stopLive}
        onConfigure={openDigitalSettings}
      />

      <div className="digital-centers-layout" data-testid="digital-centers-layout">
        <DigitalCenterList
          centers={workspace.centers}
          selectedProvider={workspace.selectedProvider}
          total={workspace.workItems.total}
          loading={workspace.loading.workspace}
          onSelect={workspace.selectCenter}
          onConfigure={openDigitalSettings}
        />
        <LocomotiveWorklist
          page={workspace.workItems}
          search={workspace.search}
          compareStatus={workspace.compareStatus}
          loading={workspace.loading.worklist}
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
          liveStatus={workspace.liveStatus}
          messages={workspace.messages}
          actions={workspace.actions}
          loading={workspace.loading.live}
          errors={{ live: workspace.errors.live, messages: workspace.errors.messages }}
        />
      </div>

      {workspace.dialog?.kind === "comparison" && workspace.selectedItem && (
        <LocomotiveComparisonDialog item={workspace.selectedItem} onClose={workspace.closeDialog} />
      )}
    </section>
  );
}
