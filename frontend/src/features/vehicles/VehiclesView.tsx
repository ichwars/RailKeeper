import { ArticleSearchDialog } from "../../shared/articleSearch/ArticleSearchDialog";
import { BarcodeSearchDialog } from "../../shared/articleSearch/BarcodeSearchDialog";
import { DecoderArchiveDialog } from "./DecoderCVArchive";
import {
  DeleteVehicleDialog,
  ExhibitionAssignmentDialog,
  ImagePreviewDialog,
  QrDialog,
  ReportDialog
} from "./VehicleDialogs";
import { VehicleInventoryPanel } from "./VehicleInventoryPanel";
import { VehicleSetEditorDialog } from "./VehicleSetEditorDialog";
import { VehicleSetSummaryDialog } from "./VehicleSetSummaryDialog";
import { currentArticleValue, vehicleArticleSearchGroups, type ArticleFieldKey } from "./articleSearch";

import { VehicleEditorHost } from "./VehicleEditorHost";
import { useVehicleWorkspace } from "./useVehicleWorkspace";

export function VehiclesView({ username, roles = ["Editor"] }: { username: string; roles?: string[] }) {
  const workspace = useVehicleWorkspace({ username, roles });
  const {
    archiveOwner,
    canReadDecoderArchive,
    setArchiveOwner,
    collection,
    inventory,
    columnPreferences,
    filters,
    sets,
    editor,
    output,
    setDeleteCandidate,
    canEditVehicleSets,
    renderers,
    articleSearch,
    t,
    media,
    deleteCandidate,
    mutations,
  } = workspace;
  return (
    <>
      {archiveOwner === username && canReadDecoderArchive && <DecoderArchiveDialog onClose={() => setArchiveOwner(null)} />}
      <VehicleInventoryPanel
        onOpenCVArchive={canReadDecoderArchive ? () => setArchiveOwner(username) : undefined}
        vehicles={collection.vehicles}
        sortedVehicles={inventory.sortedVehicles}
        loading={collection.loading}
        message={collection.message}
        messageTone={collection.messageTone}
        query={collection.query}
        columns={columnPreferences.columns}
        columnWidths={columnPreferences.widths}
        columnsLoading={columnPreferences.loading}
        sort={inventory.sort}
        inventoryView={inventory.inventoryView}
        inventoryFilter={inventory.inventoryFilter}
        maintenanceFilter={inventory.maintenanceFilter}
        qualityFilter={inventory.qualityFilter}
        manufacturerFilter={inventory.manufacturerFilter}
        categoryFilter={inventory.categoryFilter}
        gattungFilter={inventory.gattungFilter}
        railwayCompanyFilter={inventory.railwayCompanyFilter}
        epochFilter={inventory.epochFilter}
        adapterFilter={inventory.adapterFilter}
        exhibitionReadyFilter={inventory.exhibitionReadyFilter}
        inventorySummary={inventory.inventorySummary}
        maintenanceReminderSummary={inventory.maintenanceReminderSummary}
        nextMaintenanceReminder={inventory.nextMaintenanceReminder}
        inventoryFilters={filters.inventoryFilters}
        maintenanceFilters={filters.maintenanceFilters}
        inventoryFilterOptions={inventory.inventoryFilterOptions}
        hasActiveInventoryFilters={inventory.hasActiveInventoryFilters}
        allVisibleSelected={inventory.allVisibleSelected}
        selectedVehicleIDs={inventory.selectedVehicleIDs}
        onCreate={() => {
          sets.setCreatePrefill(null);
          editor.commands.openCreate();
        }}
        onReload={collection.load}
        onOpenReport={() => output.report.setters.setOpen(true)}
        onQueryChange={collection.setQuery}
        onToggleColumn={columnPreferences.toggleColumn}
        onMoveColumn={columnPreferences.moveColumn}
        onResetColumns={columnPreferences.resetColumns}
        onApplyCompactColumns={columnPreferences.applyCompactColumns}
        onPreviewColumnWidth={columnPreferences.loading ? undefined : columnPreferences.previewColumnWidth}
        onCommitColumnWidth={columnPreferences.loading ? undefined : columnPreferences.commitColumnWidth}
        onToggleSort={inventory.toggleSort}
        onInventoryViewChange={inventory.setInventoryViewMode}
        onInventoryFilterChange={inventory.setInventoryFilter}
        onMaintenanceFilterChange={inventory.setMaintenanceFilter}
        onQualityFilterChange={inventory.setQualityFilter}
        onManufacturerFilterChange={inventory.setManufacturerFilter}
        onCategoryFilterChange={inventory.setCategoryFilter}
        onGattungFilterChange={inventory.setGattungFilter}
        onRailwayCompanyFilterChange={inventory.setRailwayCompanyFilter}
        onEpochFilterChange={inventory.setEpochFilter}
        onAdapterFilterChange={inventory.setAdapterFilter}
        onExhibitionReadyFilterChange={inventory.setExhibitionReadyFilter}
        onResetFilters={inventory.resetInventoryFilters}
        onOpenDetail={editor.commands.openDetail}
        onOpenEdit={editor.commands.openEdit}
        onDelete={setDeleteCandidate}
        onToggleSelection={inventory.toggleVehicleSelection}
        onToggleSetSelection={inventory.toggleVehicleSetSelection}
        onToggleAllVisibleSelection={inventory.toggleAllVisibleSelection}
        onToggleExhibition={output.exhibition.commands.toggle}
        onOpenSet={(setID) => sets.openVehicleSet(setID, "view")}
        onEditSet={canEditVehicleSets ? (setID) => sets.openVehicleSet(setID, "edit") : undefined}
        onDuplicateSet={canEditVehicleSets ? sets.duplicateVehicleSetByID : undefined}
        renderQuickMenu={renderers.vehicleQuickMenu}
      />
      {output.report.state.open && (
        <ReportDialog
          reportMode={output.report.state.mode}
          reportTitle={output.report.state.title}
          reportSelection={output.report.state.selection}
          reportIncludeQRCode={output.report.state.includeQRCode}
          reportIncludeImages={output.report.state.includeImages}
          selectedCount={inventory.selectedVisibleVehicles.length}
          canUseSelected={inventory.someVisibleSelected}
          creating={output.report.state.creating}
          onReportModeChange={output.report.setters.setMode}
          onReportTitleChange={output.report.setters.setTitle}
          onReportSelectionChange={output.report.setters.setSelection}
          onReportIncludeQRCodeChange={output.report.setters.setIncludeQRCode}
          onReportIncludeImagesChange={output.report.setters.setIncludeImages}
          onClose={() => output.report.setters.setOpen(false)}
          onSubmit={output.report.commands.create}
        />
      )}
      <VehicleEditorHost workspace={workspace} />
      {sets.vehicleSetMode === "view" && sets.selectedVehicleSetID && (
        <VehicleSetSummaryDialog
          key={`${sets.vehicleSetSessionKey}:${sets.selectedVehicleSetID}`}
          setId={sets.selectedVehicleSetID}
          captureScope={() => sets.captureVehicleSetScope(sets.vehicleSetSessionKey)}
          canEdit={canEditVehicleSets}
          onClose={sets.closeVehicleSetDialog}
          onUpdated={() => undefined}
          onEdit={(setID) => sets.openVehicleSet(setID, "edit")}
          onDuplicate={sets.duplicateVehicleSet}
          onOpenVehicle={sets.openSetMember}
        />
      )}
      {sets.vehicleSetMode === "edit" && sets.selectedVehicleSetID && canEditVehicleSets && (
        <VehicleSetEditorDialog
          key={`${sets.vehicleSetSessionKey}:${sets.selectedVehicleSetID}`}
          setId={sets.selectedVehicleSetID}
          captureScope={() => sets.captureVehicleSetScope(sets.vehicleSetSessionKey)}
          options={collection.options}
          onClose={sets.closeVehicleSetDialog}
          onUpdated={sets.handleVehicleSetUpdated}
          onImageChanged={sets.handleVehicleSetImageChanged}
          onCatalogChanged={sets.handleVehicleSetImageChanged}
        />
      )}
      {articleSearch.state.open && editor.state.mode !== "create" && (
        <ArticleSearchDialog
          reviewTarget={t("importReview.target.vehicleDraft")}
          fieldGroups={vehicleArticleSearchGroups(t)}
          currentValue={(key) => currentArticleValue(editor.state.form, key as ArticleFieldKey)}
          loading={articleSearch.state.loading}
          response={articleSearch.state.response}
          error={articleSearch.state.error}
          selectedFields={articleSearch.state.selectedFields}
          selectedImages={articleSearch.state.selectedImages}
          onApply={articleSearch.commands.applyResult}
          onClose={() => articleSearch.setters.setOpen(false)}
          onToggleField={articleSearch.commands.toggleField}
          onToggleImage={articleSearch.commands.toggleImage}
        />
      )}
      {articleSearch.state.barcodeOpen && editor.state.mode !== "create" && (
        <BarcodeSearchDialog
          value={articleSearch.state.barcodeValue}
          onValueChange={articleSearch.setters.setBarcodeValue}
          onClose={() => articleSearch.setters.setBarcodeOpen(false)}
          onSubmit={articleSearch.commands.submitBarcode}
        />
      )}
      {output.qr.state.open && (
        <QrDialog
          form={editor.state.form}
          qrSvg={output.qr.state.svg}
          error={output.qr.state.error}
          onClose={output.qr.commands.close}
          onDownloadPng={output.qr.commands.downloadPng}
          onDownloadSvg={output.qr.commands.downloadSvg}
          onPrint={output.qr.commands.print}
        />
      )}
      {media.state.previewImage && (
        <ImagePreviewDialog
          image={media.state.previewImage}
          onClose={() => media.setters.setPreviewImage(null)}
        />
      )}
      {output.exhibition.state.assignment && (
        <ExhibitionAssignmentDialog
          assignment={output.exhibition.state.assignment}
          duplicateVehicle={output.exhibition.state.duplicateVehicle}
          duplicateDecoder={output.exhibition.state.duplicateDecoder}
          onClose={output.exhibition.commands.close}
          onListChange={output.exhibition.commands.loadEntries}
          onSave={output.exhibition.commands.save}
        />
      )}
      {deleteCandidate && (
        <DeleteVehicleDialog
          vehicle={deleteCandidate}
          onClose={() => setDeleteCandidate(null)}
          onConfirm={mutations.confirmDelete}
        />
      )}
    </>
  );
}
