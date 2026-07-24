import { useCallback, useEffect, useState } from "react";
import {
  api,
  CreateVehicleRequest,
  Vehicle
} from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { ArticleSearchDialog } from "./ArticleSearchDialog";
import { BarcodeSearchDialog } from "./BarcodeSearchDialog";
import { DeleteAttachmentDialog, DeleteVehicleDialog, ExhibitionAssignmentDialog, ImagePreviewDialog, QrDialog, ReportDialog } from "./VehicleDialogs";
import { VehicleInventoryPanel } from "./VehicleInventoryPanel";
import { VehicleEditorDialog } from "./VehicleEditorDialog";
import { maintenanceIsDue } from "./vehicleMaintenance";
import {
  PendingArticleImage,
  previewImageUrl,
  primaryImage,
  vehicleToForm
} from "./vehicleTransforms";
import {
  emptyOptions,
  fieldValue,
  hasArticleSearchCriteria,
  hasQrPayloadData,
  gattungenForCategory,
  missingVehicleModelFieldLabels,
  inferFunctionTypeFromSymbol,
} from "./vehicleViewModel";
import { useVehicleInventoryController } from "./useVehicleInventoryController";
import { useVehicleEditorController } from "./useVehicleEditorController";
import { useArticleSearchController } from "./useArticleSearchController";
import { useVehicleMediaController } from "./useVehicleMediaController";
import { useVehicleMaintenanceController } from "./useVehicleMaintenanceController";
import { useVehicleSparePartsController } from "./useVehicleSparePartsController";
import { useVehicleCVController } from "./useVehicleCVController";
import { useVehicleDecoderFilesController } from "./useVehicleDecoderFilesController";
import { useVehicleDocumentsController } from "./useVehicleDocumentsController";
import { useVehicleOutputController } from "./useVehicleOutputController";
import { useVehicleFunctionsController } from "./useVehicleFunctionsController";
import { useVehicleECoSDraftController } from "./useVehicleECoSDraftController";
import { createVehicleMutationCommands } from "./vehicleMutationCommands";
import { createVehicleFilterDefinitions, createVehicleInventoryRenderers } from "./vehicleInventoryRenderers";
import type { MasterDataOptions } from "./vehicleViewModel";

export function VehiclesView({ username }: { username: string }) {
  const { language, t } = useI18n();
  const [vehicles, setVehicles] = useState<Vehicle[]>([]);
  const [options, setOptions] = useState<MasterDataOptions>(emptyOptions);
  const [query, setQuery] = useState("");
  const [message, setMessage] = useState("");
  const [loading, setLoading] = useState(false);
  const [deleteCandidate, setDeleteCandidate] = useState<Vehicle | null>(null);
  const [quickMenuVehicleID, setQuickMenuVehicleID] = useState("");

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
      updateCategory,
      updateCouplingFront,
      updateCouplingSame,
      openCreate,
      closeModal,
      openDetail,
      openEdit,
      toggleSection
    }
  } = useVehicleEditorController({
    options,
    onMessage: setMessage,
    onFormChange: (nextForm) => syncECoSUnclearFields(nextForm),
    onReset: (reason) => {
      clearECoSDraft();
      resetVehicleMedia(reason === "close");
      setFunctionEdits({});
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
      loadFunctionEdits(detail.functions);
      resetMaintenanceForm();
      resetSparePartDetail();
      resetCVController();
      resetDecoderFiles();
    }
  });

  const {
    state: {
      showConfiguredOnly: showConfiguredFunctionsOnly,
      configuredKeys: configuredFunctionKeys,
      visibleKeys: visibleFunctionKeys,
      summary: functionSummary
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
      save: saveFunction,
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
      attachmentDeleteCandidate,
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
      setAttachmentDeleteCandidate,
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
    refreshSelectedVehicle: (vehicleId) => refreshSelectedVehicle(vehicleId),
    onImageUploadComplete: () => {
      setCVFileProfile("");
      setCVFileDescription("");
    }
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
    refreshSelectedVehicle: (vehicleId) => refreshSelectedVehicle(vehicleId)
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
    refreshSelectedVehicle: (vehicleId) => refreshSelectedVehicle(vehicleId),
    t
  });

  const {
    state: {
      form: cvForm,
      editingId: editingCVID,
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
    refreshSelectedVehicle: (vehicleId) => refreshSelectedVehicle(vehicleId)
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
    setSaving,
    onMessage: setMessage,
    onImportPreview: setCVImportPreview,
    refreshSelectedVehicle: (vehicleId) => refreshSelectedVehicle(vehicleId)
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
    refreshSelectedVehicle: (vehicleId) => refreshSelectedVehicle(vehicleId),
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
      applyResult: applyArticleResult
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

  const {
    allVisibleSelected,
    categoryFilter,
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
    resetInventoryFilters,
    selectedVehicleIDs,
    selectedVisibleVehicles,
    setCategoryFilter,
    setExhibitionReadyFilter,
    setGattungFilter,
    setInventoryFilter,
    setInventoryViewMode,
    setMaintenanceFilter,
    setManufacturerFilter,
    setQualityFilter,
    someVisibleSelected,
    sort,
    sortedVehicles,
    toggleAllVisibleSelection,
    toggleSort,
    toggleVehicleSelection
  } = useVehicleInventoryController(vehicles);
  const load = useCallback(() => {
    setLoading(true);
    setMessage("");
    api
      .vehicles(query)
      .then(setVehicles)
      .catch((error: Error) => setMessage(error.message))
      .finally(() => setLoading(false));
  }, [query]);

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
    setSelectedDetail,
    onMessage: setMessage,
    t
  });

  useEffect(() => {
    load();
  }, [load]);

  useEffect(() => {
    const reloadVisible = () => {
      if (!document.hidden) {
        load();
      }
    };
    window.addEventListener("focus", reloadVisible);
    window.addEventListener("online", reloadVisible);
    document.addEventListener("visibilitychange", reloadVisible);
    return () => {
      window.removeEventListener("focus", reloadVisible);
      window.removeEventListener("online", reloadVisible);
      document.removeEventListener("visibilitychange", reloadVisible);
    };
  }, [load]);

  useEffect(() => {
    Promise.all([
      api.masterDataAll(true),
      api.masterDataRelations("vehicle_category", "vehicle_gattung")
    ])
      .then(([entriesByType, categoryRelations]) => {
        setOptions({
          manufacturers: entriesByType.manufacturer || [],
          gauges: entriesByType.gauge || [],
          epochs: entriesByType.epoch || [],
          railwayCompanies: entriesByType.railway_company || [],
          categories: entriesByType.vehicle_category || [],
          gattungen: entriesByType.vehicle_gattung || [],
          symbols: entriesByType.symbols || [],
          categoryRelations
        });
      })
      .catch((error: Error) => setMessage(error.message));
  }, []);

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
      setSelected(null);
      setMode("create");
      setForm(prepared.form);
      setPendingArticleImages(prepared.mergeImages([]));
      setAttachmentEdits({});
      setImageUploadMaintenanceID("");
      setAttachmentUploadCategory("");
      setAttachmentUploadDescription("");
      setAttachmentDragActive(false);
      setAttachmentDeleteCandidate(null);
      setSaveAttempted(false);
      setFunctionEdits(prepared.functionEdits);
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
      setShowConfiguredFunctionsOnly((draft.functionValues || []).length > 0);
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
      if (event.target instanceof Element && event.target.closest(".quick-menu-wrap")) {
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

  const refreshSelectedVehicle = (vehicleID = selected?.id) => {
    if (!vehicleID) return Promise.resolve();
    return api
      .vehicle(vehicleID)
      .then((detail) => {
        setSelectedDetail(detail);
        load();
      })
      .catch((error: Error) => setMessage(error.message));
  };
  const { submit, confirmDelete } = createVehicleMutationCommands({
    editor: {
      form,
      selected,
      mode,
      setSaving,
      setSelectedDetail,
      setMode,
      setSaveAttempted,
      setActiveTab,
      openModelSection: () => setOpenSections((current) => ({ ...current, model: true })),
      close: closeModal
    },
    validation: {
      missingRequiredLabels: missingRequiredModelFields
    },
    media: { pendingImages: pendingArticleImages },
    spareParts: {
      selectedInputs: selectedFoundSparePartInputs,
      clearSelected: clearSelectedFoundSpareParts
    },
    functions: {
      configuredKeys: configuredFunctionKeys,
      edit: functionEdit
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
  const { sortHeader, vehicleQuickMenu, selectOptions } = createVehicleInventoryRenderers({
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
        query={query}
        inventoryView={inventoryView}
        inventoryFilter={inventoryFilter}
        maintenanceFilter={maintenanceFilter}
        qualityFilter={qualityFilter}
        manufacturerFilter={manufacturerFilter}
        categoryFilter={categoryFilter}
        gattungFilter={gattungFilter}
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
        onCreate={openCreate}
        onReload={load}
        onOpenReport={() => setReportDialogOpen(true)}
        onQueryChange={setQuery}
        onInventoryViewChange={setInventoryViewMode}
        onInventoryFilterChange={setInventoryFilter}
        onMaintenanceFilterChange={setMaintenanceFilter}
        onQualityFilterChange={setQualityFilter}
        onManufacturerFilterChange={setManufacturerFilter}
        onCategoryFilterChange={setCategoryFilter}
        onGattungFilterChange={setGattungFilter}
        onExhibitionReadyFilterChange={setExhibitionReadyFilter}
        onResetFilters={resetInventoryFilters}
        onOpenDetail={openDetail}
        onOpenEdit={openEdit}
        onDelete={setDeleteCandidate}
        onToggleSelection={toggleVehicleSelection}
        onToggleAllVisibleSelection={toggleAllVisibleSelection}
        onToggleExhibition={toggleVehicleExhibition}
        renderSortHeader={sortHeader}
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
          mode={mode}
          selected={selected}
          activeTab={activeTab}
          saving={saving}
          message={message}
          onSubmit={submit}
          onClose={closeModal}
          onTabChange={setActiveTab}
          onEdit={() => {
            if (selected) openEdit(selected);
          }}
          onPrint={() => {
            if (selected) printVehicleReport(selected);
          }}
          onQr={generateQr}
          onPreviewImage={setPreviewImage}
          tabs={{
            model: {
              form,
              externalMappings: selected?.externalMappings || [],
              readonly,
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
              draftMode: Boolean(ecosDraft),
              readonly,
              saving,
              functionImportInputRef,
              configuredFunctionKeys,
              functionSummary,
              showConfiguredFunctionsOnly,
              visibleFunctionKeys,
              symbols: options.symbols,
              onImportFunctions: importFunctions,
              onExportFunctions: exportFunctions,
              onShowConfiguredFunctionsOnlyChange: setShowConfiguredFunctionsOnly,
              functionEdit,
              updateFunctionEdit,
              inferFunctionTypeFromSymbol,
              saveFunction,
              deleteFunction
            },
            speedCurve: { selected, ecosDraft },
            cv: {
              selected,
              ecosDraft,
              readonly,
              saving,
              cvImportInputRef,
              cvFileInputRef,
              cvSummary,
              cvImportPreview,
              cvImportStats,
              cvForm,
              editingCVID,
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
              readonly,
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
              onDeleteAttachment: setAttachmentDeleteCandidate,
              onSearchDocuments: runUploadDocumentSearch,
              onImportDocument: importFoundDocument,
              onImportSelectedDocuments: importSelectedFoundDocuments,
              onToggleDocument: toggleFoundDocument,
              onToggleAllDocuments: toggleAllFoundDocuments
            },
            maintenance: {
              selected,
              pendingArticleImages,
              readonly,
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
              readonly,
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
      {articleSearchOpen && (
        <ArticleSearchDialog
          form={form}
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
      {barcodeSearchOpen && (
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
      {attachmentDeleteCandidate && (
        <DeleteAttachmentDialog
          attachment={attachmentDeleteCandidate}
          onClose={() => setAttachmentDeleteCandidate(null)}
          onConfirm={() => deleteAttachment(attachmentDeleteCandidate)}
        />
      )}
    </>
  );
}
