import { useVehicleCollectionController } from "./useVehicleCollectionController";
import { useVehicleSetController } from "./useVehicleSetController";
import { useEffect, useState } from "react";
import { Vehicle } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { ArticleSearchDialog } from "../../shared/articleSearch/ArticleSearchDialog";
import { BarcodeSearchDialog } from "../../shared/articleSearch/BarcodeSearchDialog";
import { currentArticleValue, vehicleArticleSearchGroups, type ArticleFieldKey } from "./articleSearch";
import {
  DeleteVehicleDialog,
  ExhibitionAssignmentDialog,
  ImagePreviewDialog,
  QrDialog,
  ReportDialog
} from "./VehicleDialogs";
import { VehicleInventoryPanel } from "./VehicleInventoryPanel";
import { VehicleEditorDialog } from "./VehicleEditorDialog";
import { VehicleSetEditorDialog } from "./VehicleSetEditorDialog";
import { VehicleSetSummaryDialog } from "./VehicleSetSummaryDialog";

import { maintenanceIsDue } from "./vehicleMaintenance";

import {
  hasArticleSearchCriteria,
  hasQrPayloadData,
  gattungenForCategory,
  missingVehicleModelFieldLabels,
  inferFunctionTypeFromSymbol,

} from "./vehicleViewModel";
import { useVehicleInventoryController } from "./useVehicleInventoryController";
import { useVehicleColumnPreferences } from "./useVehicleColumnPreferences";
import { useVehicleEditorController } from "./useVehicleEditorController";
import { useArticleSearchController } from "./useArticleSearchController";
import { useVehicleMediaController } from "./useVehicleMediaController";
import { useVehicleMaintenanceController } from "./useVehicleMaintenanceController";
import { useVehicleSparePartsController } from "./useVehicleSparePartsController";
import { useVehicleCVController } from "./useVehicleCVController";
import { useVehicleDecoderFilesController } from "./useVehicleDecoderFilesController";
import { useVehicleUploads } from "./useVehicleUploads";
import { useVehicleMediaDeletion } from "./useVehicleMediaDeletion";
import { VehicleMediaDeleteDialog, VehicleMediaUploadDialog } from "./VehicleMediaDialogs";
import { useVehicleDocumentsController } from "./useVehicleDocumentsController";
import { useVehicleOutputController } from "./useVehicleOutputController";
import { useVehicleFunctionsController } from "./useVehicleFunctionsController";
import { useVehicleECoSDraftController } from "./useVehicleECoSDraftController";
import { createVehicleMutationCommands } from "./vehicleMutationCommands";
import { createVehicleFilterDefinitions, createVehicleInventoryRenderers } from "./vehicleInventoryRenderers";

export function VehiclesView({ username, roles = ["Editor"] }: { username: string; roles?: string[] }) {
  const { language, t } = useI18n();
  const { vehicles, setVehicles, options, query, setQuery, message, messageTone, setMessage, setStatusMessage, loading, load } =
    useVehicleCollectionController();
  const [deleteCandidate, setDeleteCandidate] = useState<Vehicle | null>(null);
  const [quickMenuVehicleID, setQuickMenuVehicleID] = useState("");
	const canEditVehicleSets = roles.includes("Admin") || roles.includes("Editor");

  const {
    state: {
      form,
      saving,
      selected,
      mode,
      modalOpen,
      activeTab,
      openSections,
      saveAttempted,
      refreshConflicts,
      lastSavedAction,
      readonly
    },
    setters: {
      setForm,
      setSaving,
      setSelected,
      setMode,
      setModalOpen,
      setActiveTab,
      setOpenSections,
      setSaveAttempted
    },
    commands: {
      update,
      setSelectedDetail,
      applyIndependentDetail,
      acknowledgeSavedVehicle,
      resolveRefreshConflicts,
      updateCategory,
      updateCouplingFront,
      updateCouplingSame,
      openCreate,
      closeModal,
      openDetail,
      openEdit,
      toggleSection,
      refreshSelectedVehicle
    }
  } = useVehicleEditorController({
    options,
    onCollectionChanged: load,
    onMessage: setMessage,
    onFormChange: (nextForm) => syncECoSUnclearFields(nextForm),
    onReset: (reason) => {
      clearECoSDraft();
      resetVehicleMedia(reason === "close");
      mediaUploads.commands.reset();
      mediaDeletion.commands.reset();
      resetFunctionEdits();
      resetMaintenanceForm();
      resetSparePartForm();
      resetSparePartSearch();
      resetUploadDocumentSearch();
      resetCVController();
      resetDecoderFiles();
    },
    onDetailLoaded: (detail) => {
      clearECoSDraft();
      loadVehicleMedia(detail);
      loadFunctionEdits(detail.functions, detail.id);
      resetMaintenanceForm();
      resetSparePartDetail();
      resetCVController();
      resetDecoderFiles();
    },
    onDetailRefreshed: (detail, previous) => {
      refreshVehicleMedia(detail, previous);
      loadFunctionEdits(detail.functions, detail.id);
    },
    onCoreSaved: (detail, previous, phase) => {
      acknowledgeMediaSaved(detail, previous, phase);
      loadFunctionEdits(detail.functions, detail.id);
    }
  });

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    if (params.get("create") === "1" && canEditVehicleSets) {
      params.delete("create");
      window.history.replaceState(null, "", `${window.location.pathname}${params.size ? `?${params}` : ""}`);
      openCreate();
      return;
    }

    const vehicleID = params.get("id");
    if (!vehicleID) return;
    const vehicle = vehicles.find((candidate) => candidate.id === vehicleID);
    if (!vehicle) return;
    const requestedTab = params.get("tab");
    params.delete("id");
    params.delete("tab");
    window.history.replaceState(null, "", `${window.location.pathname}${params.size ? `?${params}` : ""}`);
    openDetail(vehicle);
    if (requestedTab === "maintenance") setActiveTab("maintenance");
  }, [canEditVehicleSets, openCreate, openDetail, setActiveTab, vehicles]);

  const {
    state: {
      showConfiguredOnly: showConfiguredFunctionsOnly,
      configuredKeys: configuredFunctionKeys
    },
    refs: { importInputRef: functionImportInputRef },
    setters: {
      setEdits: setFunctionEdits,
      setShowConfiguredOnly: setShowConfiguredFunctionsOnly
    },
    commands: {
      functionEdit,
      updateFunctionEdit,
      reset: resetFunctionEdits,
      loadDetail: loadFunctionEdits,
      mergeDetail: mergeFunctionEdits,
      supplementFileSuggestions,
      changes: functionChanges,
      acknowledgeSaved: acknowledgeSavedFunctions,
      remove: deleteFunction,
      exportValues: exportFunctions,
      importValues: importFunctions
    }
  } = useVehicleFunctionsController({
    selected,
    form,
    setSaving,
    onMessage: setMessage,
    refreshSelectedVehicle: (vehicleId) => refreshSelectedVehicle(vehicleId)
  });

  const {
    state: {
      mediaConflicts,
      pendingImages: pendingArticleImages,
      previewImage,
      attachmentEdits,
      imageUploadMaintenanceId: imageUploadMaintenanceID,
      attachmentUploadCategory,
      attachmentUploadDescription,
      attachmentDragActive
    },
    refs: { imageInputRef, attachmentInputRef },
    setters: {
      setPendingImages: setPendingArticleImages,
      setPreviewImage,
      setAttachmentEdits,
      setImageUploadMaintenanceId: setImageUploadMaintenanceID,
      setAttachmentUploadCategory,
      setAttachmentUploadDescription,
      setAttachmentDragActive
    },
    commands: {
      reset: resetVehicleMedia,
      loadDetail: loadVehicleMedia,
      refreshDetail: refreshVehicleMedia,
      acknowledgeCoreSaved: acknowledgeMediaSaved,
      acknowledgeImportedImage,
      acknowledgeImage,
      removeDeletedImage,
      resolveConflict: resolveMediaConflict,
      addImages: addPendingImages,
      setPrimaryPendingImage,
      updatePendingImageTitle,
      updatePendingImageMaintenance,
      movePendingImage,
      removePendingImage,
      uploadImages,
      uploadAttachment,
      onAttachmentDrag,
      onAttachmentDrop,
      updateAttachmentEdit,
      saveAttachment,
      deleteAttachment
    }
  } = useVehicleMediaController({
    selected,
    readonly,
    saving,
    setSaving,
    onMessage: setMessage,
    refreshSelectedVehicle: (vehicleId) => refreshSelectedVehicle(vehicleId, "media"),
    reloadSelectedVehicle: (vehicleId) => refreshSelectedVehicle(vehicleId),
    prepareUpload: (...args) => mediaUploads.commands.prepare(...args),
    prepareDelete: request => mediaDeletion.commands.prepare(request)
  });

  const refreshSavedMedia = (vehicleId: string, kind: "image" | "attachment" | "decoder") =>
    refreshSelectedVehicle(vehicleId, kind === "decoder" ? "decoderFiles" : "media", true);
  const mediaUploads = useVehicleUploads({
    selected, readonly, saving, setSaving, onMessage: setMessage, onStatus: setStatusMessage,
    onImageStored: acknowledgeImage, refreshSaved: refreshSavedMedia,
    reload: vehicleId => refreshSelectedVehicle(vehicleId, undefined, true)
  });
  const mediaDeletion = useVehicleMediaDeletion({
    selected, readonly, saving, setSaving, onMessage: setMessage, onStatus: setStatusMessage,
    onImageDeleted: removeDeletedImage, refreshSaved: refreshSavedMedia,
    reload: vehicleId => refreshSelectedVehicle(vehicleId, undefined, true)
  });

  const {
    state: {
      form: maintenanceForm,
      editingId: editingMaintenanceID
    },
    commands: {
      updateForm: updateMaintenanceForm,
      resetForm: resetMaintenanceForm,
      edit: editMaintenance,
      save: saveMaintenance,
      complete: completeMaintenance,
      remove: deleteMaintenance
    }
  } = useVehicleMaintenanceController({
    selected,
    setSaving,
    onMessage: setMessage,
    refreshSelectedVehicle: (vehicleId) => refreshSelectedVehicle(vehicleId, "maintenance")
  });

  const {
    state: {
      form: sparePartForm,
      editingId: editingSparePartID,
      searchLoading: sparePartSearchLoading,
      searchError: sparePartSearchError,
      searchRan: sparePartSearchRan,
      foundParts: foundSpareParts,
      selectedFoundParts: selectedFoundSpareParts,
      sort: sparePartSort,
      lookupLoadingId: sparePartLookupLoadingID,
      lookupErrors: sparePartLookupErrors,
      lookupResults: sparePartLookupResults,
      statusLoading: sparePartStatusLoading,
      statuses: sparePartStatuses,
      importAllLoading: importAllSparePartsLoading,
      canImportAll: canImportAllAvailableSpareParts,
      importAllTitle: importAllAvailableSparePartsTitle
    },
    commands: {
      updateForm: updateSparePartForm,
      resetForm: resetSparePartForm,
      resetSearch: resetSparePartSearch,
      resetDetail: resetSparePartDetail,
      edit: editSparePart,
      save: saveSparePart,
      remove: deleteSparePart,
      toggleFound: toggleFoundSparePart,
      toggleAllFound: toggleAllFoundSpareParts,
      toggleSort: toggleSparePartSort,
      selectedInputs: selectedFoundSparePartInputs,
      clearSelectedFound: clearSelectedFoundSpareParts,
      searchSingle: searchSingleSparePart,
      applyLookup: applySparePartLookup,
      importAll: importAllAvailableSpareParts,
      extractAttachment: extractAttachmentSpareParts,
      runSearch: runSparePartSearch
    }
  } = useVehicleSparePartsController({
    selected,
    active: activeTab === "spareParts",
    attachmentEdits,
    setSaving,
    onMessage: setMessage,
    onOpenSpareParts: () => setActiveTab("spareParts"),
    refreshSelectedVehicle: (vehicleId) => refreshSelectedVehicle(vehicleId, "spareParts"),
    t
  });

  const {
    state: {
      form: cvForm,
      editingId: editingCVID,
      manualDraft: manualCVDraft,
      importPreview: cvImportPreview,
      summary: cvSummary,
      importStats: cvImportStats,
      storedDecoderProfiles,
      decoderProfileOptions
    },
    refs: { importInputRef: cvImportInputRef },
    commands: {
      updateForm: updateCVForm,
      resetForm: resetCVForm,
      reset: resetCVController,
      edit: editCVValue,
      save: saveCVValue,
      remove: deleteCVValue,
      applyManualDraft,
      discardManualDraft,
      discardManualChange,
      rebaseManualChange,
      exportValues: exportCVValues,
      importValues: importCVValues,
      toggleImportRow: toggleCVImportRow,
      selectImportRows: selectCVImportRows,
      applyImportPreview: applyCVImportPreview,
      setImportPreview: setCVImportPreview,
      discardImportPreview: discardCVImportPreview
    }
  } = useVehicleCVController({
    selected,
    decoderNumber: form.digitalDecoderNumber || form.dtDecoderNumber || "",
    setSaving,
    onMessage: setMessage,
    refreshSelectedVehicle: (vehicleId) => refreshSelectedVehicle(vehicleId, "cv"),
    reloadSelectedVehicle: (vehicleId) => refreshSelectedVehicle(vehicleId)
  });

  const {
    state: {
      fileProfile: cvFileProfile,
      fileDescription: cvFileDescription,
      uploadPreview: cvFileUploadPreview,
      previewStats: cvFilePreviewStats
    },
    refs: { fileInputRef: cvFileInputRef },
    commands: {
      reset: resetDecoderFiles,
      uploadFiles: uploadCVFiles,
      updateFileProfile: setCVFileProfile,
      updateFileDescription: setCVFileDescription,
      applyFirstSuggestion: applyFirstCVFileSuggestion,
      previewValuesForImport: previewCVFileValuesForImport,
      applyFunctionSuggestions: applyCVFileFunctionSuggestions,
      confirmUpload: confirmCVFileUpload,
      discardUploadPreview: discardCVFileUploadPreview,
      remove: deleteCVFile
    }
  } = useVehicleDecoderFilesController({
    selected,
    readonly,
    saving,
    setSaving,
    onMessage: setMessage,
    onStatus: setStatusMessage,
    onImportPreview: setCVImportPreview,
    onFunctionSuggestions: (previews) => {
      const result = supplementFileSuggestions(previews);
      if (result.applied > 0) setActiveTab("control");
      return result;
    },
    prepareUpload: mediaUploads.commands.prepare,
    prepareDelete: mediaDeletion.commands.prepare
  });

  const {
    state: {
      loading: documentSearchLoading,
      error: documentSearchError,
      ran: documentSearchRan,
      documents: foundUploadDocuments,
      selectedDocuments: selectedUploadDocuments
    },
    commands: {
      reset: resetUploadDocumentSearch,
      search: runUploadDocumentSearch,
      importOne: importFoundDocument,
      importSelected: importSelectedFoundDocuments,
      toggle: toggleFoundDocument,
      toggleAll: toggleAllFoundDocuments
    }
  } = useVehicleDocumentsController({
    selected,
    setSaving,
    onMessage: setMessage,
    refreshSelectedVehicle: (vehicleId) => refreshSelectedVehicle(vehicleId, "documents"),
    t
  });

  const {
    state: {
      open: articleSearchOpen,
      loading: articleSearchLoading,
      response: articleSearchResponse,
      error: articleSearchError,
      barcodeOpen: barcodeSearchOpen,
      barcodeValue: barcodeSearchValue,
      selectedFields: selectedArticleFields,
      selectedImages: selectedArticleImages
    },
    setters: {
      setOpen: setArticleSearchOpen,
      setBarcodeOpen: setBarcodeSearchOpen,
      setBarcodeValue: setBarcodeSearchValue
    },
    commands: {
      run: runArticleSearch,
      openBarcode: openBarcodeSearch,
      submitBarcode: submitBarcodeSearch,
      toggleField: toggleArticleField,
      toggleImage: toggleArticleImage,
      applyResult: applyArticleResult,
      restoreDraft: restoreArticleDraft
    }
  } = useArticleSearchController({
    form,
    pendingImageCount: pendingArticleImages.length,
    replaceForm: setForm,
    updateForm: update,
    onMessage: setMessage,
    t,
    addImages: addPendingImages
  });

  const columnPreferences = useVehicleColumnPreferences(setMessage);

  const {
    adapterFilter,
    allVisibleSelected,
    categoryFilter,
    epochFilter,
    exhibitionReadyFilter,
    gattungFilter,
    hasActiveInventoryFilters,
    inventoryFilter,
    inventoryFilterCounts,
    inventoryFilterOptions,
    inventorySummary,
    inventoryView,
    maintenanceFilter,
    maintenanceReminderSummary,
    manufacturerFilter,
    nextMaintenanceReminder,
    qualityFilter,
    railwayCompanyFilter,
    resetInventoryFilters,
    selectedVehicleIDs,
    selectedVisibleVehicles,
    setAdapterFilter,
    setCategoryFilter,
    setEpochFilter,
    setExhibitionReadyFilter,
    setGattungFilter,
    setInventoryFilter,
    setInventoryViewMode,
    setMaintenanceFilter,
    setManufacturerFilter,
    setQualityFilter,
    setRailwayCompanyFilter,
    someVisibleSelected,
    sort,
    sortedVehicles,
    toggleAllVisibleSelection,
    toggleSort,
		toggleVehicleSelection,
		toggleVehicleSetSelection
  } = useVehicleInventoryController(vehicles, columnPreferences.columns);
  const {
    qr: {
      state: { open: qrDialogOpen, svg: qrSvg, error: qrError },
      commands: {
        close: closeQrDialog,
        generate: generateQr,
        downloadPng: downloadQrPng,
        downloadSvg: downloadQrSvg,
        print: printQr,
        openForVehicle: openQrForVehicle
      }
    },
    report: {
      state: {
        open: reportDialogOpen,
        mode: reportMode,
        title: reportTitle,
        selection: reportSelection,
        includeQRCode: reportIncludeQRCode,
        includeImages: reportIncludeImages,
        creating: reportCreating
      },
      setters: {
        setOpen: setReportDialogOpen,
        setMode: setReportMode,
        setTitle: setReportTitle,
        setSelection: setReportSelection,
        setIncludeQRCode: setReportIncludeQRCode,
        setIncludeImages: setReportIncludeImages
      },
      commands: { create: createInventoryReport, printVehicle: printVehicleReport }
    },
    exhibition: {
      state: {
        assignment: exhibitionAssignment,
        duplicateVehicle: duplicateAssignmentVehicle,
        duplicateDecoder: duplicateAssignmentDecoder
      },
      commands: {
        close: closeExhibitionAssignment,
        loadEntries: loadAssignmentEntries,
        save: saveExhibitionAssignment,
        toggle: toggleVehicleExhibition
      }
    }
  } = useVehicleOutputController({
    username,
    form,
    selected,
    query,
    sort,
    sortedVehicles,
    selectedVisibleVehicles,
    setVehicles,
    setSelectedDetail: applyIndependentDetail,
    onMessage: setMessage,
    t
  });

  const {
    state: { draft: ecosDraft, unclearFields: ecosUnclearFields },
    setters: { setDraft: setEcosDraft, setUnclearFields: setEcosUnclearFields },
    commands: {
      clear: clearECoSDraft,
      syncUnclearFields: syncECoSUnclearFields,
      fieldClass: ecosFieldClass,
      markImportSessionSaved: markECoSImportSessionSaved,
      returnToImportSession: returnToECoSImportSession
    }
  } = useVehicleECoSDraftController({
    onOpenCreate: (prepared) => {
			setCreatePrefill(null);
      setSelected(null);
      setMode("create");
      setForm(prepared.form);
      setPendingArticleImages(prepared.mergeImages([]));
      setAttachmentEdits({});
      setImageUploadMaintenanceID("");
      setAttachmentUploadCategory("");
      setAttachmentUploadDescription("");
      setAttachmentDragActive(false);
      mediaUploads.commands.reset();
      mediaDeletion.commands.reset();
      setSaveAttempted(false);
      setFunctionEdits(Object.fromEntries(Object.entries(prepared.functionEdits)
        .map(([key, edit]) => [key, { ...edit, dirty: true, modeAssigned: true }])));
    },
    onOpenUpdate: (detail, prepared) => {
      setSelectedDetail(detail);
      setPendingArticleImages((current) => prepared.mergeImages(current));
      setMode("edit");
      setForm(prepared.form);
      mergeFunctionEdits(detail.functions, prepared.functionEdits);
    },
    onFinishOpen: (draft) => {
      resetMaintenanceForm();
      resetSparePartForm();
      resetSparePartSearch();
      resetUploadDocumentSearch();
      resetCVController();
      resetDecoderFiles();
      setShowConfiguredFunctionsOnly(true);
      setActiveTab("model");
      setOpenSections({ model: true, details: false, vehicle: false });
      setModalOpen(true);
      setMessage(draft.mode === "update" ? t("vehicles.ecosDraft.loadedUpdate") : t("vehicles.ecosDraft.loaded"));
    },
    onMessage: setMessage,
    t
  });

  useEffect(() => {
    if (!quickMenuVehicleID) return;
    const closeOnPointerDown = (event: PointerEvent) => {
      if (event.target instanceof Element && event.target.closest(".quick-menu-wrap, .quick-menu")) {
        return;
      }
      setQuickMenuVehicleID("");
    };
    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        setQuickMenuVehicleID("");
      }
    };
    window.addEventListener("pointerdown", closeOnPointerDown);
    window.addEventListener("keydown", closeOnEscape);
    return () => {
      window.removeEventListener("pointerdown", closeOnPointerDown);
      window.removeEventListener("keydown", closeOnEscape);
    };
  }, [quickMenuVehicleID]);

  const canRunArticleSearch = hasArticleSearchCriteria(form);
  const canGenerateQr = hasQrPayloadData(selected, form);
  const missingRequiredModelFields = missingVehicleModelFieldLabels(form, t);
  const showRequiredErrors = saveAttempted || Boolean(ecosDraft);
  const { inventoryFilters, maintenanceFilters } = createVehicleFilterDefinitions({
    vehicleCount: vehicles.length,
    counts: inventoryFilterCounts,
    t
  });
  const filteredGattungen = gattungenForCategory(options, form.category);

  const { submit, confirmDelete } = createVehicleMutationCommands({
    draftOwner: username,
    editor: {
      form,
      selected,
      mode,
      setSaving,
      setSelectedDetail: (detail) => acknowledgeSavedVehicle(detail, form, "complete"),
      acknowledgeCoreSaved: (detail) => acknowledgeSavedVehicle(detail, form, "core"),
      hasRefreshConflicts: refreshConflicts.length > 0,
      reloadSelectedVehicle: (vehicleId) => refreshSelectedVehicle(vehicleId),
      setMode,
      setSaveAttempted,
      setActiveTab,
      openModelSection: () => setOpenSections((current) => ({ ...current, model: true })),
      close: closeModal
    },
    validation: {
      missingRequiredLabels: missingRequiredModelFields
    },
    media: { pendingImages: pendingArticleImages, acknowledgeImportedImage,
      hasConflicts: mediaConflicts.some(conflict => conflict.kind === "image") },
    spareParts: {
      selectedInputs: selectedFoundSparePartInputs,
      clearSelected: clearSelectedFoundSpareParts
    },
    functions: {
      changes: functionChanges,
      acknowledgeSaved: acknowledgeSavedFunctions
    },
    ecos: {
      draft: ecosDraft,
      unclearFieldCount: ecosUnclearFields.size,
      markSaved: markECoSImportSessionSaved,
      clear: clearECoSDraft,
      returnToSession: returnToECoSImportSession
    },
    deletion: {
      candidate: deleteCandidate,
      setCandidate: setDeleteCandidate
    },
    reloadVehicles: load,
    onMessage: setMessage,
    t
  });
  const {
    selectedVehicleSetID, vehicleSetMode, createPrefill, setCreatePrefill,
    submitVehicleSet, closeVehicleSetDialog, openVehicleSet, duplicateVehicleSet,
    closeVehicleCreateDialog, duplicateVehicleSetByID, handleVehicleSetUpdated, handleVehicleSetImageChanged, openSetMember
  } = useVehicleSetController({
    username, form, pendingArticleImages, setSaving, setSelectedDetail, setMode,
    setActiveTab, closeModal, openCreate, openDetail, vehicles, load, onMessage: setMessage, t
  });
  const { vehicleQuickMenu, selectOptions } = createVehicleInventoryRenderers({
    sort,
    quickMenuVehicleID,
    setQuickMenuVehicleID,
    toggleSort,
    openDetail,
    openEdit,
    openQr: openQrForVehicle,
    printVehicle: printVehicleReport,
    setDeleteCandidate,
    t
  });
  const maintenanceEntries = selected?.maintenance || [];
  const maintenanceSummary = {
    due: maintenanceEntries.filter(maintenanceIsDue).length,
    planned: maintenanceEntries.filter((entry) => entry.status !== "erledigt").length,
    done: maintenanceEntries.filter((entry) => entry.status === "erledigt").length
  };

  return (
    <>
      <VehicleInventoryPanel
        vehicles={vehicles}
        sortedVehicles={sortedVehicles}
        loading={loading}
        message={message}
        messageTone={messageTone}
        query={query}
        columns={columnPreferences.columns}
        columnWidths={columnPreferences.widths}
        columnsLoading={columnPreferences.loading}
        sort={sort}
        inventoryView={inventoryView}
        inventoryFilter={inventoryFilter}
        maintenanceFilter={maintenanceFilter}
        qualityFilter={qualityFilter}
        manufacturerFilter={manufacturerFilter}
        categoryFilter={categoryFilter}
        gattungFilter={gattungFilter}
        railwayCompanyFilter={railwayCompanyFilter}
        epochFilter={epochFilter}
        adapterFilter={adapterFilter}
        exhibitionReadyFilter={exhibitionReadyFilter}
        inventorySummary={inventorySummary}
        maintenanceReminderSummary={maintenanceReminderSummary}
        nextMaintenanceReminder={nextMaintenanceReminder}
        inventoryFilters={inventoryFilters}
        maintenanceFilters={maintenanceFilters}
        inventoryFilterOptions={inventoryFilterOptions}
        hasActiveInventoryFilters={hasActiveInventoryFilters}
        allVisibleSelected={allVisibleSelected}
        selectedVehicleIDs={selectedVehicleIDs}
		onCreate={() => {
			setCreatePrefill(null);
			openCreate();
		}}
        onReload={load}
        onOpenReport={() => setReportDialogOpen(true)}
        onQueryChange={setQuery}
        onToggleColumn={columnPreferences.toggleColumn}
        onMoveColumn={columnPreferences.moveColumn}
        onResetColumns={columnPreferences.resetColumns}
        onPreviewColumnWidth={columnPreferences.loading ? undefined : columnPreferences.previewColumnWidth}
        onCommitColumnWidth={columnPreferences.loading ? undefined : columnPreferences.commitColumnWidth}
        onToggleSort={toggleSort}
        onInventoryViewChange={setInventoryViewMode}
        onInventoryFilterChange={setInventoryFilter}
        onMaintenanceFilterChange={setMaintenanceFilter}
        onQualityFilterChange={setQualityFilter}
        onManufacturerFilterChange={setManufacturerFilter}
        onCategoryFilterChange={setCategoryFilter}
        onGattungFilterChange={setGattungFilter}
        onRailwayCompanyFilterChange={setRailwayCompanyFilter}
        onEpochFilterChange={setEpochFilter}
        onAdapterFilterChange={setAdapterFilter}
        onExhibitionReadyFilterChange={setExhibitionReadyFilter}
        onResetFilters={resetInventoryFilters}
        onOpenDetail={openDetail}
        onOpenEdit={openEdit}
        onDelete={setDeleteCandidate}
        onToggleSelection={toggleVehicleSelection}
		onToggleSetSelection={toggleVehicleSetSelection}
        onToggleAllVisibleSelection={toggleAllVisibleSelection}
        onToggleExhibition={toggleVehicleExhibition}
		onOpenSet={(setID) => openVehicleSet(setID, "view")}
		onEditSet={canEditVehicleSets ? (setID) => openVehicleSet(setID, "edit") : undefined}
		onDuplicateSet={canEditVehicleSets ? duplicateVehicleSetByID : undefined}
        renderQuickMenu={vehicleQuickMenu}
      />
      {reportDialogOpen && (
        <ReportDialog
          reportMode={reportMode}
          reportTitle={reportTitle}
          reportSelection={reportSelection}
          reportIncludeQRCode={reportIncludeQRCode}
          reportIncludeImages={reportIncludeImages}
          selectedCount={selectedVisibleVehicles.length}
          canUseSelected={someVisibleSelected}
          creating={reportCreating}
          onReportModeChange={setReportMode}
          onReportTitleChange={setReportTitle}
          onReportSelectionChange={setReportSelection}
          onReportIncludeQRCodeChange={setReportIncludeQRCode}
          onReportIncludeImagesChange={setReportIncludeImages}
          onClose={() => setReportDialogOpen(false)}
          onSubmit={createInventoryReport}
        />
      )}
      {modalOpen && (
        <VehicleEditorDialog
          draftOwner={username}
          mode={mode}
          selected={selected}
          activeTab={activeTab}
          saving={saving}
          message={message}
          messageTone={messageTone}
          mediaDialogs={<>
            {mediaUploads.batch && <VehicleMediaUploadDialog batch={mediaUploads.batch}
              onClose={mediaUploads.commands.close} onConfirm={() => void mediaUploads.commands.confirm()}
              onSkip={mediaUploads.commands.skip} />}
            {mediaDeletion.candidate && <VehicleMediaDeleteDialog candidate={mediaDeletion.candidate}
              onClose={mediaDeletion.commands.close} onConfirm={() => void mediaDeletion.commands.confirm()} />}
          </>}
          refreshConflicts={refreshConflicts}
          mediaConflicts={mediaConflicts}
          onResolveMediaConflict={resolveMediaConflict}
          lastSavedAction={lastSavedAction}
          onResolveRefreshConflicts={resolveRefreshConflicts}
          onSubmit={submit}
          onSubmitSet={submitVehicleSet}
          onClose={closeVehicleCreateDialog}
          onTabChange={setActiveTab}
          onEdit={() => {
            if (selected) openEdit(selected);
          }}
          onPrint={() => {
            if (selected) printVehicleReport(selected);
          }}
          onQr={generateQr}
          onPreviewImage={setPreviewImage}
          setCreationDisabled={Boolean(ecosDraft)}
          createPrefill={createPrefill}
          articleSearchController={{
            state: {
              open: articleSearchOpen, loading: articleSearchLoading, response: articleSearchResponse,
              error: articleSearchError, barcodeOpen: barcodeSearchOpen, barcodeValue: barcodeSearchValue,
              selectedFields: selectedArticleFields, selectedImages: selectedArticleImages
            },
            setters: {
              setOpen: setArticleSearchOpen, setBarcodeOpen: setBarcodeSearchOpen,
              setBarcodeValue: setBarcodeSearchValue
            },
            commands: {
              run: runArticleSearch, openBarcode: openBarcodeSearch, submitBarcode: submitBarcodeSearch,
              toggleField: toggleArticleField, toggleImage: toggleArticleImage, applyResult: applyArticleResult,
              restoreDraft: restoreArticleDraft
            }
          }}
          tabs={{
            model: {
              form,
              cvValues: selected?.cvValues || [],
              ecosDecoderProfile: selected?.ecosDecoderProfile,
              onOpenCV: () => setActiveTab("cv"),
              externalMappings: selected?.externalMappings || [],
              readonly: readonly || saving,
              hideAcquisitionDetails: Boolean(selected?.vehicleSetId),
              articleSearchLoading,
              canRunArticleSearch,
              showRequiredErrors,
              options,
              filteredGattungen,
              openSections,
              selectOptions,
              ecosFieldClass,
              onToggleSection: toggleSection,
              onOpenBarcodeSearch: openBarcodeSearch,
              onRunArticleSearch: () => runArticleSearch(),
              onUpdate: update,
              onUpdateCategory: updateCategory,
              onOpenQr: generateQr,
              canOpenQr: canGenerateQr,
              onUpdateCouplingFront: updateCouplingFront,
              onUpdateCouplingSame: updateCouplingSame
            },
            functions: {
              selected,
              canReadECoS: roles.includes("Admin") && !readonly && !saving,
              decoderType: form.decoderType,
              draftMode: mode === "create" || Boolean(ecosDraft),
              draftCVValues: ecosDraft?.cvValues,
              draftDecoderProfile: ecosDraft?.source === "ecos" ? ecosDraft.sourceSummary.profile : undefined,
              readonly: readonly || saving,
              saving,
              functionImportInputRef,
              configuredFunctionKeys,
              showConfiguredFunctionsOnly,
              symbols: options.symbols,
              onImportFunctions: importFunctions,
              onExportFunctions: exportFunctions,
              onShowConfiguredFunctionsOnlyChange: setShowConfiguredFunctionsOnly,
              functionEdit,
              updateFunctionEdit,
              inferFunctionTypeFromSymbol,
              deleteFunction
            },
            speedCurve: { selected, ecosDraft },
            cv: {
              selected,
              decoderType: form.decoderType,
              ecosDraft,
              readonly: readonly || saving,
              saving,
              cvImportInputRef,
              cvFileInputRef,
              cvSummary,
              cvImportPreview,
              cvImportStats,
              cvForm,
              editingCVID,
              manualDraft: manualCVDraft,
              applyManualDraft,
              discardManualDraft,
              discardManualChange,
              rebaseManualChange,
              decoderProfileOptions,
              storedDecoderProfiles,
              cvFileProfile,
              cvFileDescription,
              cvFileUploadPreview,
              cvFilePreviewStats,
              importCVValues,
              exportCVValues,
              selectCVImportRows,
              applyCVImportPreview,
              discardCVImportPreview,
              toggleCVImportRow,
              updateCVForm,
              resetCVForm,
              saveCVValue,
              editCVValue,
              deleteCVValue,
              uploadCVFiles,
              setCVFileProfile,
              setCVFileDescription,
              applyFirstCVFileSuggestion,
              previewCVFileValuesForImport,
              applyCVFileFunctionSuggestions,
              confirmCVFileUpload,
              discardCVFileUploadPreview,
              deleteCVFile
            },
            uploads: {
              selected,
              canTransferToECoS: roles.includes("Admin") && !readonly && !saving,
              readonly: readonly || saving,
              saving,
              imageInputRef,
              attachmentInputRef,
              maintenanceEntries,
              imageUploadMaintenanceID,
              pendingArticleImages,
              attachmentDragActive,
              attachmentUploadCategory,
              attachmentUploadDescription,
              attachmentEdits,
              documentSearchLoading,
              documentSearchError,
              documentSearchRan,
              foundDocuments: foundUploadDocuments,
              selectedDocuments: selectedUploadDocuments,
              onImageUploadMaintenanceIDChange: setImageUploadMaintenanceID,
              onUploadImages: uploadImages,
              onPreviewImage: setPreviewImage,
              onUpdatePendingImageTitle: updatePendingImageTitle,
              onUpdatePendingImageMaintenance: updatePendingImageMaintenance,
              onMovePendingImage: movePendingImage,
              onSetPrimaryPendingImage: setPrimaryPendingImage,
              onRemovePendingImage: removePendingImage,
              onUploadAttachment: uploadAttachment,
              onAttachmentDrag,
              onAttachmentDrop,
              onAttachmentUploadCategoryChange: setAttachmentUploadCategory,
              onAttachmentUploadDescriptionChange: setAttachmentUploadDescription,
              onUpdateAttachmentEdit: updateAttachmentEdit,
              onSaveAttachment: saveAttachment,
              onExtractAttachmentSpareParts: extractAttachmentSpareParts,
              onDeleteAttachment: deleteAttachment,
              onSearchDocuments: runUploadDocumentSearch,
              onImportDocument: importFoundDocument,
              onImportSelectedDocuments: importSelectedFoundDocuments,
              onToggleDocument: toggleFoundDocument,
              onToggleAllDocuments: toggleAllFoundDocuments
            },
            maintenance: {
              selected,
              pendingArticleImages,
              readonly: readonly || saving,
              saving,
              maintenanceForm,
              editingMaintenanceID,
              maintenanceSummary,
              onUpdateMaintenanceForm: updateMaintenanceForm,
              onResetMaintenanceForm: resetMaintenanceForm,
              onSaveMaintenance: saveMaintenance,
              onCompleteMaintenance: completeMaintenance,
              onEditMaintenance: editMaintenance,
              onDeleteMaintenance: deleteMaintenance
            },
            spareParts: {
              selected,
              readonly: readonly || saving,
              saving,
              sparePartForm,
              editingSparePartID,
              sparePartLookupLoadingID,
              sparePartLookupErrors,
              sparePartLookupResults,
              sparePartStatusLoading,
              sparePartStatuses,
              importAllSparePartsLoading,
              canImportAllSpareParts: canImportAllAvailableSpareParts,
              importAllSparePartsTitle: importAllAvailableSparePartsTitle,
              onUpdateSparePartForm: updateSparePartForm,
              onResetSparePartForm: resetSparePartForm,
              onSaveSparePart: saveSparePart,
              onEditSparePart: editSparePart,
              onDeleteSparePart: deleteSparePart,
              onSearchSparePart: searchSingleSparePart,
              onApplySparePartLookup: applySparePartLookup,
              onImportAllSpareParts: importAllAvailableSpareParts,
              sparePartSort,
              onToggleSparePartSort: toggleSparePartSort
            }
          }}
        />
      )}
		{vehicleSetMode === "view" && selectedVehicleSetID && (
			<VehicleSetSummaryDialog
				setId={selectedVehicleSetID}
				canEdit={canEditVehicleSets}
				onClose={closeVehicleSetDialog}
				onUpdated={() => undefined}
				onEdit={(setID) => openVehicleSet(setID, "edit")}
				onDuplicate={duplicateVehicleSet}
				onOpenVehicle={openSetMember}
			/>
		)}
		{vehicleSetMode === "edit" && selectedVehicleSetID && canEditVehicleSets && (
			<VehicleSetEditorDialog
				setId={selectedVehicleSetID}
				options={options}
				onClose={closeVehicleSetDialog}
				onUpdated={handleVehicleSetUpdated}
				onImageChanged={handleVehicleSetImageChanged}
			/>
		)}
      {articleSearchOpen && mode !== "create" && (
        <ArticleSearchDialog
          fieldGroups={vehicleArticleSearchGroups(t)}
          currentValue={(key) => currentArticleValue(form, key as ArticleFieldKey)}
          loading={articleSearchLoading}
          response={articleSearchResponse}
          error={articleSearchError}
          selectedFields={selectedArticleFields}
          selectedImages={selectedArticleImages}
          onApply={applyArticleResult}
          onClose={() => setArticleSearchOpen(false)}
          onToggleField={toggleArticleField}
          onToggleImage={toggleArticleImage}
        />
      )}
      {barcodeSearchOpen && mode !== "create" && (
        <BarcodeSearchDialog
          value={barcodeSearchValue}
          onValueChange={setBarcodeSearchValue}
          onClose={() => setBarcodeSearchOpen(false)}
          onSubmit={submitBarcodeSearch}
        />
      )}
      {qrDialogOpen && (
        <QrDialog
          form={form}
          qrSvg={qrSvg}
          error={qrError}
          onClose={closeQrDialog}
          onDownloadPng={downloadQrPng}
          onDownloadSvg={downloadQrSvg}
          onPrint={printQr}
        />
      )}
      {previewImage && (
        <ImagePreviewDialog
          image={previewImage}
          onClose={() => setPreviewImage(null)}
        />
      )}
      {exhibitionAssignment && (
        <ExhibitionAssignmentDialog
          assignment={exhibitionAssignment}
          duplicateVehicle={duplicateAssignmentVehicle}
          duplicateDecoder={duplicateAssignmentDecoder}
          onClose={closeExhibitionAssignment}
          onListChange={loadAssignmentEntries}
          onSave={saveExhibitionAssignment}
        />
      )}
      {deleteCandidate && (
        <DeleteVehicleDialog
          vehicle={deleteCandidate}
          onClose={() => setDeleteCandidate(null)}
          onConfirm={confirmDelete}
        />
      )}
    </>
  );
}
