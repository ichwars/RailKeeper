import { observeAction } from "../../shared/asyncAction";
import { VehicleCatalogPanel } from "./VehicleCatalogPanel";
import { VehicleEditorDialog } from "./VehicleEditorDialog";
import { catalogFieldsFrom } from "./vehicleCatalogModel";

import { VehicleMediaDeleteDialog, VehicleMediaUploadDialog } from "./VehicleMediaDialogs";
import type { VehicleWorkspace } from "./useVehicleWorkspace";
import {
  inferFunctionTypeFromSymbol
} from "./vehicleViewModel";

export function VehicleEditorHost({ workspace }: { workspace: VehicleWorkspace }) {
  const {
    username,
    roles,
    editor,
    collection,
    canEditVehicleSets,
    mediaUploads,
    mediaDeletion,
    media,
    mutations,
    sets,
    output,
    ecos,
    articleSearch,
    canRunArticleSearch,
    showRequiredErrors,
    filteredGattungen,
    renderers,
    canGenerateQr,
    functions,
    canReadDecoderArchive,
    decoderConfigurations,
    cvWorkingGroup,
    cvs,
    decoderFiles,
    maintenanceEntries,
    documents,
    spareParts,
    maintenance,
    maintenanceSummary,
  } = workspace;
  const selected = editor.state.selected;
  return <>{editor.state.modalOpen && (
    <VehicleEditorDialog
      draftOwner={username}
      mode={editor.state.mode}
      selected={selected}
      activeTab={editor.state.activeTab}
      saving={editor.state.saving}
      message={collection.message}
      messageTone={collection.messageTone}
      catalog={selected && editor.state.mode !== "create" ? <VehicleCatalogPanel key={`${editor.state.sessionKey}:${selected.id}`}
        targetType="vehicle" targetId={selected.id} targetName={selected.name}
        savedFields={catalogFieldsFrom(selected)} readonly={editor.state.readonly || !canEditVehicleSets} busy={editor.state.saving}
        captureScope={() => editor.commands.captureScope(editor.state.sessionKey)}
        onSaved={confirmed => editor.commands.refreshSelectedVehicle(selected.id, confirmed ? "catalog" : undefined, true)} /> : undefined}
      mediaDialogs={<>
        {mediaUploads.batch && <VehicleMediaUploadDialog batch={mediaUploads.batch}
          onClose={mediaUploads.commands.close} onConfirm={() => observeAction(mediaUploads.commands.confirm())}
          onReview={() => observeAction(mediaUploads.commands.reviewWorkingGroup())}
          onSkip={mediaUploads.commands.skip} />}
        {mediaDeletion.candidate && <VehicleMediaDeleteDialog candidate={mediaDeletion.candidate}
          onClose={mediaDeletion.commands.close} onConfirm={() => observeAction(mediaDeletion.commands.confirm())} />}
      </>}
      refreshConflicts={editor.state.refreshConflicts}
      mediaConflicts={media.state.mediaConflicts}
      onResolveMediaConflict={media.commands.resolveConflict}
      lastSavedAction={editor.state.lastSavedAction}
      onResolveRefreshConflicts={editor.commands.resolveRefreshConflicts}
      onSubmit={mutations.submit}
      onSubmitSet={sets.submitVehicleSet}
      onClose={sets.closeVehicleCreateDialog}
      onTabChange={editor.setters.setActiveTab}
      onEdit={() => {
        if (selected) editor.commands.openEdit(selected);
      }}
      onPrint={() => {
        if (selected) observeAction(output.report.commands.printVehicle(selected));
      }}
      onQr={output.qr.commands.generate}
      onPreviewImage={media.setters.setPreviewImage}
      setCreationDisabled={Boolean(ecos.state.draft)}
      createPrefill={sets.createPrefill}
      articleSearchController={{
        state: {
          open: articleSearch.state.open, loading: articleSearch.state.loading, response: articleSearch.state.response,
          error: articleSearch.state.error, barcodeOpen: articleSearch.state.barcodeOpen, barcodeValue: articleSearch.state.barcodeValue,
          selectedFields: articleSearch.state.selectedFields, selectedImages: articleSearch.state.selectedImages
        },
        setters: {
          setOpen: articleSearch.setters.setOpen, setBarcodeOpen: articleSearch.setters.setBarcodeOpen,
          setBarcodeValue: articleSearch.setters.setBarcodeValue
        },
        commands: {
          run: articleSearch.commands.run, openBarcode: articleSearch.commands.openBarcode, submitBarcode: articleSearch.commands.submitBarcode,
          toggleField: articleSearch.commands.toggleField, toggleImage: articleSearch.commands.toggleImage, applyResult: articleSearch.commands.applyResult,
          restoreDraft: articleSearch.commands.restoreDraft
        }
      }}
      tabs={{
        model: {
          form: editor.state.form,
          cvValues: selected?.cvValues || [],
          ecosDecoderProfile: selected?.ecosDecoderProfile,
          onOpenCV: () => editor.setters.setActiveTab("cv"),
          externalMappings: selected?.externalMappings || [],
          readonly: editor.state.readonly || editor.state.saving,
          hideAcquisitionDetails: Boolean(selected?.vehicleSetId),
          articleSearchLoading: articleSearch.state.loading,
          canRunArticleSearch,
          showRequiredErrors,
          options: collection.options,
          filteredGattungen,
          openSections: editor.state.openSections,
          selectOptions: renderers.selectOptions,
          ecosFieldClass: ecos.commands.fieldClass,
          onToggleSection: editor.commands.toggleSection,
          onOpenBarcodeSearch: articleSearch.commands.openBarcode,
          onRunArticleSearch: () => articleSearch.commands.run(),
          onUpdate: editor.commands.update,
          onUpdateCategory: editor.commands.updateCategory,
          onOpenQr: output.qr.commands.generate,
          canOpenQr: canGenerateQr,
          onUpdateCouplingFront: editor.commands.updateCouplingFront,
          onUpdateCouplingSame: editor.commands.updateCouplingSame
        },
        functions: {
          selected: selected,
          canReadECoS: roles.includes("Admin") && !editor.state.readonly && !editor.state.saving,
          decoderType: editor.state.form.decoderType,
          draftMode: editor.state.mode === "create" || Boolean(ecos.state.draft),
          draftCVValues: ecos.state.draft?.cvValues,
          draftDecoderProfile: ecos.state.draft?.source === "ecos" ? ecos.state.draft.sourceSummary.profile : undefined,
          readonly: editor.state.readonly || editor.state.saving,
          saving: editor.state.saving,
          functionImportInputRef: functions.refs.importInputRef,
          configuredFunctionKeys: functions.state.configuredKeys,
          showConfiguredFunctionsOnly: functions.state.showConfiguredOnly,
          symbols: collection.options.symbols,
          onImportFunctions: functions.commands.importValues,
          onExportFunctions: functions.commands.exportValues,
          onShowConfiguredFunctionsOnlyChange: functions.setters.setShowConfiguredOnly,
          functionEdit: functions.commands.functionEdit,
          updateFunctionEdit: functions.commands.updateFunctionEdit,
          inferFunctionTypeFromSymbol,
          deleteFunction: functions.commands.remove
        },
        speedCurve: { selected: selected, ecosDraft: ecos.state.draft },
        cv: {
          selected: selected,
          configurations: canReadDecoderArchive ? decoderConfigurations : undefined,
          canReadArchive: canReadDecoderArchive,
          archiveOwnerKey: username,
          decoderType: editor.state.form.decoderType,
          ecosDraft: ecos.state.draft,
          readonly: editor.state.readonly || editor.state.saving || Boolean(selected && !cvWorkingGroup.ready),
          saving: editor.state.saving,
          cvImportInputRef: cvs.refs.importInputRef,
          cvFileInputRef: decoderFiles.refs.fileInputRef,
          cvSummary: cvs.state.summary,
          cvImportPreview: cvs.state.importPreview,
          cvImportNeedsReview: cvs.state.importNeedsReview,
          recompareCVImportPreview: cvs.commands.recompareImportPreview,
          cvImportStats: cvs.state.importStats,
          cvForm: cvs.state.form,
          editingCVID: cvs.state.editingId,
          manualDraft: cvs.state.manualDraft,
          applyManualDraft: cvs.commands.applyManualDraft,
          discardManualDraft: cvs.commands.discardManualDraft,
          discardManualChange: cvs.commands.discardManualChange,
          rebaseManualChange: cvs.commands.rebaseManualChange,
          decoderProfileOptions: cvs.state.decoderProfileOptions,
          storedDecoderProfiles: cvs.state.storedDecoderProfiles,
          cvFileProfile: decoderFiles.state.fileProfile,
          cvFileDescription: decoderFiles.state.fileDescription,
          cvFileUploadPreview: decoderFiles.state.uploadPreview,
          cvFilePreviewStats: decoderFiles.state.previewStats,
          importCVValues: cvs.commands.importValues,
          exportCVValues: cvs.commands.exportValues,
          selectCVImportRows: cvs.commands.selectImportRows,
          applyCVImportPreview: cvs.commands.applyImportPreview,
          discardCVImportPreview: cvs.commands.discardImportPreview,
          toggleCVImportRow: cvs.commands.toggleImportRow,
          updateCVForm: cvs.commands.updateForm,
          resetCVForm: cvs.commands.resetForm,
          saveCVValue: cvs.commands.save,
          editCVValue: cvs.commands.edit,
          deleteCVValue: cvs.commands.remove,
          uploadCVFiles: decoderFiles.commands.uploadFiles,
          setCVFileProfile: decoderFiles.commands.updateFileProfile,
          setCVFileDescription: decoderFiles.commands.updateFileDescription,
          applyFirstCVFileSuggestion: decoderFiles.commands.applyFirstSuggestion,
          previewCVFileValuesForImport: decoderFiles.commands.previewValuesForImport,
          applyCVFileFunctionSuggestions: decoderFiles.commands.applyFunctionSuggestions,
          confirmCVFileUpload: decoderFiles.commands.confirmUpload,
          discardCVFileUploadPreview: decoderFiles.commands.discardUploadPreview,
          deleteCVFile: decoderFiles.commands.remove
        },
        uploads: {
          selected: selected,
          canTransferToECoS: roles.includes("Admin") && !editor.state.readonly && !editor.state.saving,
          readonly: editor.state.readonly || editor.state.saving,
          saving: editor.state.saving,
          imageInputRef: media.refs.imageInputRef,
          attachmentInputRef: media.refs.attachmentInputRef,
          maintenanceEntries,
          imageUploadMaintenanceID: media.state.imageUploadMaintenanceId,
          pendingArticleImages: media.state.pendingImages,
          attachmentDragActive: media.state.attachmentDragActive,
          attachmentUploadCategory: media.state.attachmentUploadCategory,
          attachmentUploadDescription: media.state.attachmentUploadDescription,
          attachmentEdits: media.state.attachmentEdits,
          documentSearchLoading: documents.state.loading,
          documentSearchError: documents.state.error,
          documentSearchRan: documents.state.ran,
          foundDocuments: documents.state.documents,
          selectedDocuments: documents.state.selectedDocuments,
          onImageUploadMaintenanceIDChange: media.setters.setImageUploadMaintenanceId,
          onUploadImages: media.commands.uploadImages,
          onPreviewImage: media.setters.setPreviewImage,
          onUpdatePendingImageTitle: media.commands.updatePendingImageTitle,
          onUpdatePendingImageMaintenance: media.commands.updatePendingImageMaintenance,
          onMovePendingImage: media.commands.movePendingImage,
          onSetPrimaryPendingImage: media.commands.setPrimaryPendingImage,
          onRemovePendingImage: media.commands.removePendingImage,
          onUploadAttachment: media.commands.uploadAttachment,
          onAttachmentDrag: media.commands.onAttachmentDrag,
          onAttachmentDrop: media.commands.onAttachmentDrop,
          onAttachmentUploadCategoryChange: media.setters.setAttachmentUploadCategory,
          onAttachmentUploadDescriptionChange: media.setters.setAttachmentUploadDescription,
          onUpdateAttachmentEdit: media.commands.updateAttachmentEdit,
          onSaveAttachment: media.commands.saveAttachment,
          onExtractAttachmentSpareParts: spareParts.commands.extractAttachment,
          onDeleteAttachment: media.commands.deleteAttachment,
          onSearchDocuments: documents.commands.search,
          onImportDocument: documents.commands.importOne,
          onImportSelectedDocuments: documents.commands.importSelected,
          onToggleDocument: documents.commands.toggle,
          onToggleAllDocuments: documents.commands.toggleAll
        },
        maintenance: {
          selected: selected,
          pendingArticleImages: media.state.pendingImages,
          readonly: editor.state.readonly || editor.state.saving,
          saving: editor.state.saving,
          maintenanceForm: maintenance.state.form,
          editingMaintenanceID: maintenance.state.editingId,
          maintenanceSummary,
          onUpdateMaintenanceForm: maintenance.commands.updateForm,
          onResetMaintenanceForm: maintenance.commands.resetForm,
          onSaveMaintenance: maintenance.commands.save,
          onCompleteMaintenance: maintenance.commands.complete,
          onEditMaintenance: maintenance.commands.edit,
          onDeleteMaintenance: maintenance.commands.remove
        },
        spareParts: {
          selected: selected,
          readonly: editor.state.readonly || editor.state.saving,
          saving: editor.state.saving,
          sparePartForm: spareParts.state.form,
          editingSparePartID: spareParts.state.editingId,
          sparePartLookupLoadingID: spareParts.state.lookupLoadingId,
          sparePartLookupErrors: spareParts.state.lookupErrors,
          sparePartLookupResults: spareParts.state.lookupResults,
          sparePartStatusLoading: spareParts.state.statusLoading,
          sparePartStatuses: spareParts.state.statuses,
          importAllSparePartsLoading: spareParts.state.importAllLoading,
          canImportAllSpareParts: spareParts.state.canImportAll,
          importAllSparePartsTitle: spareParts.state.importAllTitle,
          onUpdateSparePartForm: spareParts.commands.updateForm,
          onResetSparePartForm: spareParts.commands.resetForm,
          onSaveSparePart: spareParts.commands.save,
          onEditSparePart: spareParts.commands.edit,
          onDeleteSparePart: spareParts.commands.remove,
          onSearchSparePart: spareParts.commands.searchSingle,
          onApplySparePartLookup: spareParts.commands.applyLookup,
          onImportAllSpareParts: spareParts.commands.importAll,
          sparePartSort: spareParts.state.sort,
          onToggleSparePartSort: spareParts.commands.toggleSort
        }
      }}
    />
  )}</>;
}
