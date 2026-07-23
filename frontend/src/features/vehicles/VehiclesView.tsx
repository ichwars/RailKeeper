import { FormEvent, useCallback, useEffect, useMemo, useRef, useState } from "react";

import {

  ArrowUpDown,

  ChevronDown,

  ChevronUp,

  Circle,

  CircleOff,

  Cpu,

  Eye,

  Image,

  ImageOff,

  MoreVertical,

  PackageSearch,

  Pencil,

  Printer,

  QrCode,

  Trash2,

  Upload,

  Wrench,

  X

} from "lucide-react";

import {

  api,

  ArticleSearchDocument,

  CreateVehicleRequest,

  ExhibitionEntry,

  ExhibitionEntryInput,

  ExhibitionList,

  MasterDataEntry,

  MasterDataRelation,

  VehicleAttachment,

  VehicleExternalMappingInput,

  VehicleFunctionInput,

  Vehicle

} from "../../shared/api";

import { useI18n } from "../../shared/i18n";

import { ArticleSearchDialog } from "./ArticleSearchDialog";

import { BarcodeSearchDialog } from "./BarcodeSearchDialog";

import { DeleteAttachmentDialog, DeleteVehicleDialog, ExhibitionAssignmentDialog, ImagePreviewDialog, QrDialog, ReportDialog } from "./VehicleDialogs";

import { VehicleInventoryPanel } from "./VehicleInventoryPanel";

import { VehicleFunctionsTab } from "./VehicleFunctionsTab";

import { VehicleMaintenanceTab } from "./VehicleMaintenanceTab";

import { VehicleModelTab } from "./VehicleModelTab";

import { VehicleSparePartsTab } from "./VehicleSparePartsTab";

import { VehicleSpeedCurveTab } from "./VehicleSpeedCurveTab";

import { VehicleUploadsTab, webDocumentKey } from "./VehicleUploadsTab";

import { VehicleCVTab } from "./VehicleCVTab";

import { VehicleReadOnlyView } from "./VehicleReadOnlyView";

import {

  cvValueKey,

  functionKeys,

  functionMappingsFromImport,

  isValidFunctionMapping

} from "./cvImport";

import { maintenanceIsDue } from "./vehicleMaintenance";

import { buildBrandedQrPngDataUrl, buildQrSvg, downloadQrPngFile, downloadQrSvgFile, printQrSvgLabel, qrPayload } from "./vehicleQr";

import { InventoryReportAssets, inventoryReportHtml, openPrintDocument, reservePrintDocument } from "./vehicleReports";

import {

  functionsToEditState,

  normalizedText,

  PendingArticleImage,

  previewImageUrl,

  primaryImage,

  vehicleExhibitionEligible,

  vehicleToExhibitionEntry,

  vehicleToForm

} from "./vehicleTransforms";

import type { FunctionEditState } from "./vehicleTransforms";



import {

  articleSearchEnabled,

  articleSearchSources,

  compactValue,

  emptyFunctionEdit,

  emptyOptions,

  emptyVehicle,

  ecosImportSessionStorageKey,

  ecosRequiredFields,

  ecosVehicleDraftStorageKey,

  fieldValue,

  hasArticleSearchCriteria,

  hasQrPayloadData,

  inferFunctionTypeFromSymbol,

  optionValue,

  sanitizeArticleSearchResponse,

  vehicleFieldsForSearch

} from "./vehicleViewModel";
import { useVehicleInventoryController } from "./useVehicleInventoryController";
import { useVehicleEditorController } from "./useVehicleEditorController";
import { useArticleSearchController } from "./useArticleSearchController";
import { useVehicleMediaController } from "./useVehicleMediaController";
import { useVehicleMaintenanceController } from "./useVehicleMaintenanceController";
import { useVehicleSparePartsController } from "./useVehicleSparePartsController";
import { useVehicleCVController } from "./useVehicleCVController";
import { useVehicleDecoderFilesController } from "./useVehicleDecoderFilesController";

import type {

  ECoSRequiredField,

  ECoSVehicleDraftPayload,

  ExhibitionAssignment,

  InventoryFilter,

  InventoryReportMode,

  InventoryViewMode,

  InventoryReportSelection,

  MaintenanceFilter,

  MasterDataOptions,

  ModalTab,

  SortKey

} from "./vehicleViewModel";



export function VehiclesView({ username }: { username: string }) {

  const { language, t } = useI18n();

  const [vehicles, setVehicles] = useState<Vehicle[]>([]);

  const [options, setOptions] = useState<MasterDataOptions>(emptyOptions);

  const [query, setQuery] = useState("");

  const [message, setMessage] = useState("");

  const [loading, setLoading] = useState(false);

  const [deleteCandidate, setDeleteCandidate] = useState<Vehicle | null>(null);
  const [documentSearchLoading, setDocumentSearchLoading] = useState(false);

  const [documentSearchError, setDocumentSearchError] = useState("");

  const [documentSearchRan, setDocumentSearchRan] = useState(false);

  const [foundUploadDocuments, setFoundUploadDocuments] = useState<ArticleSearchDocument[]>([]);
  const [selectedUploadDocuments, setSelectedUploadDocuments] = useState<Record<string, boolean>>({});

  const [functionEdits, setFunctionEdits] = useState<FunctionEditState>({});

  const [showConfiguredFunctionsOnly, setShowConfiguredFunctionsOnly] = useState(false);

  const [ecosDraft, setEcosDraft] = useState<ECoSVehicleDraftPayload | null>(null);

  const [ecosUnclearFields, setEcosUnclearFields] = useState<Set<ECoSRequiredField>>(() => new Set());

  const functionImportInputRef = useRef<HTMLInputElement | null>(null);

  const [qrDialogOpen, setQrDialogOpen] = useState(false);

  const [qrSvg, setQrSvg] = useState("");

  const [qrError, setQrError] = useState("");

  const [reportDialogOpen, setReportDialogOpen] = useState(false);

  const [reportMode, setReportMode] = useState<InventoryReportMode>("summary");

  const [reportTitle, setReportTitle] = useState("Fahrzeugsammlung");

  const [reportSelection, setReportSelection] = useState<InventoryReportSelection>("all");

  const [reportIncludeQRCode, setReportIncludeQRCode] = useState(true);

  const [reportIncludeImages, setReportIncludeImages] = useState(true);

  const [reportCreating, setReportCreating] = useState(false);

  const [exhibitionAssignment, setExhibitionAssignment] = useState<ExhibitionAssignment | null>(null);

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
      setEcosDraft(null);
      setEcosUnclearFields(new Set());
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
      setEcosDraft(null);
      setEcosUnclearFields(new Set());
      loadVehicleMedia(detail);
      setFunctionEdits(functionsToEditState(detail.functions));
      resetMaintenanceForm();
      resetSparePartDetail();
      resetCVController();
      resetDecoderFiles();
    }
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



  const openECoSDraft = useCallback((draft: ECoSVehicleDraftPayload) => {

    const draftImages: PendingArticleImage[] = (draft.imageSuggestions || []).map((image, index) => ({

      ...image,

      id: image.id || `ecos-${draft.sourceSummary.objectId}-${index}`,

      isPrimary: index === 0,

      maintenanceId: ""

    }));

    const mergeDraftImages = (current: PendingArticleImage[]) => {

      if (draftImages.length === 0) return current;

      const existing = new Set(current.map((image) => image.url));

      const next = [...current, ...draftImages.filter((image) => !existing.has(image.url))];

      if (!next.some((image) => image.isPrimary) && next.length > 0) {

        next[0] = { ...next[0], isPrimary: true };

      }

      return next;

    };

    const draftFunctionEdits = Object.fromEntries((draft.functionValues || []).map((item) => [

      item.functionKey,

      {

        name: item.name || "",

        symbolKey: item.symbolKey || "",

        functionType: item.functionType || "standard",

        mode: item.mode || "dauer",

        directionDependent: Boolean(item.directionDependent),

        notes: item.notes || "",

        persisted: false

      }

    ]));

    const applyDraftValues = (base: CreateVehicleRequest) => {

      const next = { ...base };

      const keys = draft.importedKeys?.length

        ? draft.importedKeys

        : Object.keys(draft.vehicle) as (keyof CreateVehicleRequest)[];

      keys.forEach((key) => {

        const value = draft.vehicle[key];

        if (typeof value === "boolean" || (typeof value === "string" && value.trim() !== "")) {

          (next as Record<string, unknown>)[key] = value;

        }

      });

      return next;

    };

    const finishOpen = () => {

      resetMaintenanceForm();

      resetSparePartForm();

      resetSparePartSearch();

      resetUploadDocumentSearch();

      resetCVController();

      resetDecoderFiles();

      setEcosDraft(draft);

      setEcosUnclearFields(new Set(draft.unclearFields));

      setShowConfiguredFunctionsOnly((draft.functionValues || []).length > 0);

      setActiveTab("model");

      setOpenSections({ model: true, details: false, vehicle: false });

      setModalOpen(true);

      setMessage(draft.mode === "update" ? t("vehicles.ecosDraft.loadedUpdate") : t("vehicles.ecosDraft.loaded"));

    };

    if (draft.mode === "update" && draft.targetVehicleId) {

      api.vehicle(draft.targetVehicleId)

        .then((detail) => {

          setSelectedDetail(detail);

          setPendingArticleImages((current) => mergeDraftImages(current));

          setMode("edit");

          setForm(applyDraftValues(vehicleToForm(detail)));

          setFunctionEdits({

            ...functionsToEditState(detail.functions),

            ...draftFunctionEdits

          });

          finishOpen();

        })

        .catch((error: Error) => setMessage(error.message));

      return;

    }

    const nextForm = applyDraftValues(emptyVehicle);

    setSelected(null);

    setMode("create");

    setForm(nextForm);

    setPendingArticleImages(mergeDraftImages([]));

    setAttachmentEdits({});

    setImageUploadMaintenanceID("");

    setAttachmentUploadCategory("");

    setAttachmentUploadDescription("");

    setAttachmentDragActive(false);
    setAttachmentDeleteCandidate(null);
    setSaveAttempted(false);

    setFunctionEdits(draftFunctionEdits);

    finishOpen();

  }, [t]);



  useEffect(() => {

    const rawDraft = window.sessionStorage.getItem(ecosVehicleDraftStorageKey);

    if (!rawDraft) return;



    try {

      const draft = JSON.parse(rawDraft) as ECoSVehicleDraftPayload;

      if (draft?.source === "ecos") {

        openECoSDraft(draft);

      }

    } catch {

      setMessage(t("vehicles.ecosDraft.invalid"));

    } finally {

      window.sessionStorage.removeItem(ecosVehicleDraftStorageKey);

      if (window.location.search.includes("source=ecos")) {

        window.history.replaceState(null, "", "/vehicles");

      }

    }

  }, [openECoSDraft, t]);



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

  const requiredModelFields = [
    { key: "manufacturer" as const, label: t("vehicle.field.manufacturer"), value: form.manufacturer },
    { key: "name" as const, label: t("vehicle.field.name"), value: form.name },
    { key: "gauge" as const, label: t("vehicle.field.gauge"), value: form.gauge },
    { key: "category" as const, label: t("vehicle.field.category"), value: form.category },
    { key: "gattung" as const, label: t("vehicle.field.gattung"), value: form.gattung }
  ];

  const missingRequiredModelFields = requiredModelFields.filter((field) => !compactValue(field.value));

  const showRequiredErrors = saveAttempted || Boolean(ecosDraft);



  const inventoryFilters = [

    { key: "all" as const, label: t("vehicles.filter.all"), count: inventoryFilterCounts.all },

    { key: "digital" as const, label: t("vehicles.filter.digital"), count: inventoryFilterCounts.digital, icon: <Cpu size={15} aria-hidden="true" /> },

    { key: "analog" as const, label: t("vehicles.filter.analog"), count: inventoryFilterCounts.analog, icon: <Circle size={15} aria-hidden="true" /> },

    { key: "withImages" as const, label: t("vehicles.filter.withImages"), count: inventoryFilterCounts.withImages, icon: <Image size={15} aria-hidden="true" /> },

    { key: "withoutImages" as const, label: t("vehicles.filter.withoutImages"), count: inventoryFilterCounts.withoutImages, icon: <ImageOff size={15} aria-hidden="true" /> }

  ];



  const maintenanceFilters = [

    { key: "all" as const, label: t("vehicles.filter.all"), count: vehicles.length },

    { key: "due" as const, label: t("vehicles.filter.maintenanceDue"), count: inventoryFilterCounts.maintenanceDue, icon: <Wrench size={15} aria-hidden="true" /> },

    { key: "none" as const, label: t("vehicles.filter.withoutMaintenance"), count: inventoryFilterCounts.withoutMaintenance, icon: <CircleOff size={15} aria-hidden="true" /> }

  ];



  const filteredGattungen = useMemo(() => {

    const categoryKey = options.categories.find((entry) => optionValue(entry) === form.category)?.key;

    if (!categoryKey) {

      return options.gattungen;

    }

    const allowed = new Set(

      options.categoryRelations

        .filter((relation) => relation.parentKey === categoryKey)

        .map((relation) => relation.childKey)

    );

    return options.gattungen.filter((entry) => allowed.has(entry.key));

  }, [form.category, options]);



  const syncECoSUnclearFields = (nextForm: CreateVehicleRequest) => {

    setEcosUnclearFields((current) => {

      if (!ecosDraft && current.size === 0) return current;

      const next = new Set(current);

      ecosRequiredFields.forEach((field) => {

        if (compactValue(nextForm[field])) {

          next.delete(field);

        } else if (ecosDraft?.unclearFields.includes(field)) {

          next.add(field);

        }

      });

      return next;

    });

  };



  const ecosFieldClass = (field: ECoSRequiredField) => (ecosDraft && ecosUnclearFields.has(field) ? "ecos-unclear-field" : "");



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



  const resetUploadDocumentSearch = () => {

    setDocumentSearchLoading(false);

    setDocumentSearchError("");

    setDocumentSearchRan(false);

    setFoundUploadDocuments([]);
    setSelectedUploadDocuments({});

  };



  const runUploadDocumentSearch = () => {

    if (!selected) return;

    if (!articleSearchEnabled()) {

      setDocumentSearchError("Die Artikeldaten-Websuche ist in den Einstellungen deaktiviert.");

      setDocumentSearchRan(true);

      return;

    }

    const searchForm = vehicleToForm(selected);

    if (!hasArticleSearchCriteria(searchForm)) {

      setDocumentSearchError(t("vehicles.articleSearch.missingInput"));

      setDocumentSearchRan(true);

      return;

    }

    setDocumentSearchLoading(true);

    setDocumentSearchError("");

    setDocumentSearchRan(true);

    setFoundUploadDocuments([]);
    setSelectedUploadDocuments({});

    api

      .articleSearch({

        manufacturer: searchForm.manufacturer,

        articleNumber: searchForm.articleNumber,

        name: searchForm.name,

        gauge: searchForm.gauge,

        searchSources: articleSearchSources(),

        fields: vehicleFieldsForSearch(searchForm)

      })

      .then((response) => {

        const sanitized = sanitizeArticleSearchResponse(response);

        const documents = new Map<string, ArticleSearchDocument>();

        sanitized.results.forEach((result) => {

          (result.documents || []).forEach((document) => {

            const key = (document.url || document.title || "").toLocaleLowerCase();

            if (key && !documents.has(key)) documents.set(key, document);

          });

        });

        setFoundUploadDocuments(Array.from(documents.values()));

      })

      .catch((error: Error) => setDocumentSearchError(error.message))

      .finally(() => setDocumentSearchLoading(false));

  };



  const categoryForFoundDocument = (document: ArticleSearchDocument) => {

    const signal = `${document.kind || ""} ${document.title || ""}`.toLocaleLowerCase("de-DE");

    if (signal.includes("spare") || signal.includes("ersatzteil") || signal.includes("et-blatt")) return "Ersatzteilliste";

    if (signal.includes("manual") || signal.includes("anleitung") || signal.includes("bedienung")) return "Anleitung";

    return "Dokumentation";

  };

  const importFoundDocuments = (documents: ArticleSearchDocument[]) => {

    if (!selected) return;

    const importableDocuments = documents.filter((document) => document.url);

    if (importableDocuments.length === 0) return;

    setSaving(true);

    setMessage("");

    (async () => {

      for (const document of importableDocuments) {

        await api.importVehicleAttachmentFromUrl(selected.id, {

          url: document.url,

          title: document.title || "Dokument",

          description: `Quelle: ${document.source || document.url}\n${document.url}`,

          category: categoryForFoundDocument(document),

          maintenanceId: ""

        });

      }

    })()

      .then(() => refreshSelectedVehicle(selected.id))

      .then(() => {

        setSelectedUploadDocuments({});

        setMessage(t(importableDocuments.length === 1 ? "vehicles.uploads.webDocumentImported" : "vehicles.uploads.webDocumentsImported", { count: importableDocuments.length }));

      })

      .catch((error: Error) => setMessage(error.message))

      .finally(() => setSaving(false));

  };



  const importFoundDocument = (document: ArticleSearchDocument) => {

    importFoundDocuments([document]);

  };

  const toggleFoundDocument = (document: ArticleSearchDocument, index: number, checked: boolean) => {

    const key = webDocumentKey(document, index);

    setSelectedUploadDocuments((current) => {

      const next = { ...current };

      if (checked) {

        next[key] = true;

      } else {

        delete next[key];

      }

      return next;

    });

  };

  const toggleAllFoundDocuments = (checked: boolean) => {

    if (!checked) {

      setSelectedUploadDocuments({});

      return;

    }

    const existingDocumentUrls = new Set((selected?.attachments || []).map((attachment) => attachment.description || ""));

    setSelectedUploadDocuments(Object.fromEntries(foundUploadDocuments.flatMap((document, index) => {

      if (!document.url || Array.from(existingDocumentUrls).some((description) => description.includes(document.url))) {

        return [];

      }

      return [[webDocumentKey(document, index), true]];

    })));

  };

  const importSelectedFoundDocuments = () => {

    const documents = foundUploadDocuments.filter((document, index) => selectedUploadDocuments[webDocumentKey(document, index)]);

    importFoundDocuments(documents);

  };



  const functionEdit = (functionKey: string) => functionEdits[functionKey] || emptyFunctionEdit(functionKey);



  const updateFunctionEdit = (functionKey: string, patch: Partial<VehicleFunctionInput>) => {

    setFunctionEdits((current) => ({

      ...current,

      [functionKey]: {

        ...emptyFunctionEdit(functionKey),

        ...current[functionKey],

        ...patch

      }

    }));

  };



  const saveFunction = (functionKey: string) => {

    if (!selected) return;

    const edit = functionEdit(functionKey);

    if (!edit.persisted && !edit.name?.trim() && !edit.symbolKey && !edit.notes?.trim()) {

      setMessage(`${functionKey}: Bitte Funktionsname, Symbol oder Notiz eintragen.`);

      return;

    }

    setSaving(true);

    setMessage("");

    api

      .updateVehicleFunction(selected.id, functionKey, {

        name: edit.name || "",

        symbolKey: edit.symbolKey || "",

        functionType: edit.functionType || "standard",

        mode: edit.mode || "dauer",

        directionDependent: Boolean(edit.directionDependent),

        notes: edit.notes || ""

      })

      .then(() => refreshSelectedVehicle(selected.id))

      .catch((error: Error) => setMessage(error.message))

      .finally(() => setSaving(false));

  };



  const deleteFunction = (functionKey: string) => {

    if (!selected) return;

    setSaving(true);

    setMessage("");

    api

      .deleteVehicleFunction(selected.id, functionKey)

      .then(() => refreshSelectedVehicle(selected.id))

      .catch((error: Error) => setMessage(error.message))

      .finally(() => setSaving(false));

  };



  const exportFunctions = () => {

    if (!selected) return;

    const functionMappings = configuredFunctionKeys.map((functionKey) => {

      const edit = functionEdit(functionKey);

      return {

        functionKey,

        name: edit.name || "",

        symbolKey: edit.symbolKey || "",

        functionType: edit.functionType || "standard",

        mode: edit.mode || "dauer",

        directionDependent: Boolean(edit.directionDependent),

        notes: edit.notes || ""

      };

    });

    const payload = {

      vehicle: {

        inventoryNumber: selected.inventoryNumber,

        name: selected.name,

        decoder: form.digitalDecoderNumber || form.dtDecoderNumber || ""

      },

      functions: functionMappings

    };

    const blob = new Blob([JSON.stringify(payload, null, 2)], { type: "application/json" });

    const url = URL.createObjectURL(blob);

    const anchor = document.createElement("a");

    anchor.href = url;

    anchor.download = `${selected.inventoryNumber || "railkeeper"}-funktionen.json`;

    anchor.click();

    URL.revokeObjectURL(url);

  };



  const importFunctions = (files: FileList | null) => {

    if (!selected || !files || files.length === 0) return;

    const [file] = Array.from(files);

    setSaving(true);

    setMessage("");

    file

      .text()

      .then(functionMappingsFromImport)

      .then(async (rows) => {

        const valid = rows.filter(isValidFunctionMapping);

        if (valid.length === 0) {

          throw new Error("Keine g?ltigen Funktionszuordnungen gefunden.");

        }

        for (const row of valid) {

          await api.updateVehicleFunction(selected.id, row.functionKey, {

            name: row.name || "",

            symbolKey: row.symbolKey || "",

            functionType: row.functionType || "standard",

            mode: row.mode || "dauer",

            directionDependent: Boolean(row.directionDependent),

            notes: row.notes || ""

          });

        }

      })

      .then(() => refreshSelectedVehicle(selected.id))

      .catch((error: Error) => setMessage(error.message))

      .finally(() => {

        setSaving(false);

        if (functionImportInputRef.current) {

          functionImportInputRef.current.value = "";

        }

      });

  };



  const generateQr = async () => {

    if (!hasQrPayloadData(selected, form)) {

      setQrDialogOpen(false);

      setQrSvg("");

      setQrError("");

      setMessage(t("vehicles.qr.missingInput"));

      return;

    }



    setQrDialogOpen(true);

    setQrSvg("");

    setQrError("");

    try {

      setQrSvg(await buildQrSvg(selected, form));

    } catch (error) {

      setQrError(error instanceof Error ? error.message : "QR-Code konnte nicht erstellt werden.");

    }

  };



  const downloadQrSvg = () => {

    downloadQrSvgFile(qrSvg, form.inventoryNumber || "railkeeper");

  };



  const downloadQrPng = async () => {

    await downloadQrPngFile(qrPayload(selected, form), form.inventoryNumber || "railkeeper");

  };



  const printQr = () => {

    try {

      printQrSvgLabel(qrSvg, form);

    } catch (error) {

      setQrError(error instanceof Error ? error.message : "Druckfenster konnte nicht geöffnet werden.");

    }

  };



  const buildInventoryReportAssets = async (reportVehicles: Vehicle[], includeQRCode = reportIncludeQRCode) => {

    const assets: InventoryReportAssets = {};

    if (!includeQRCode) return assets;

    await Promise.all(

      reportVehicles.map(async (vehicle) => {

        assets[vehicle.id] = {

          qrCode: await buildBrandedQrPngDataUrl(qrPayload(vehicle, vehicleToForm(vehicle)), 192)

        };

      })

    );

    return assets;

  };



  const loadCompleteReportVehicles = async (reportVehicles: Vehicle[]) => {

    return Promise.all(reportVehicles.map((vehicle) => api.vehicle(vehicle.id)));

  };



  const createInventoryReport = async (event?: FormEvent) => {

    event?.preventDefault();

    const reportVehicles = reportSelection === "selected" ? selectedVisibleVehicles : sortedVehicles;
    const reportName = `railkeeper-inventory-${reportMode}`;

    if (reportVehicles.length === 0) {

      setMessage("Es gibt keine Fahrzeuge f?r den PDF-Report.");

      return;

    }

    setReportCreating(true);
    const reportWindow = reservePrintDocument(reportName, reportTitle.trim() || "Fahrzeugsammlung");

    try {

      const completeReportVehicles = await loadCompleteReportVehicles(reportVehicles);

      const assets = await buildInventoryReportAssets(completeReportVehicles);

      const html = inventoryReportHtml(completeReportVehicles, query, sort, {

        mode: reportMode,

        title: reportTitle.trim() || "Fahrzeugsammlung",

        includeQRCode: reportIncludeQRCode,

        includeImages: reportIncludeImages

      }, assets);

      openPrintDocument(html, reportName, reportWindow);

      setReportDialogOpen(false);

    } catch (error) {

      setMessage(error instanceof Error ? error.message : "Report konnte nicht erstellt werden.");

    } finally {

      setReportCreating(false);

    }

  };



  const printVehicleReport = async (vehicle: Vehicle) => {

    const reportName = `railkeeper-vehicle-${vehicle.id}`;
    const reportWindow = reservePrintDocument(reportName, vehicle.name || vehicle.inventoryNumber || "Fahrzeugsammlung");

    try {

      const completeVehicle = await api.vehicle(vehicle.id);

      const assets = await buildInventoryReportAssets([completeVehicle], true);

      const html = inventoryReportHtml([completeVehicle], completeVehicle.inventoryNumber || completeVehicle.name, sort, {

        mode: "details",

        title: completeVehicle.name || completeVehicle.inventoryNumber || "Fahrzeugsammlung",

        includeQRCode: true,

        includeImages: true

      }, assets);

      openPrintDocument(html, reportName, reportWindow);

    } catch (error) {

      setMessage(error instanceof Error ? error.message : "Report konnte nicht erstellt werden.");

    }

  };



  const updateVehicleExhibitionFlag = (vehicle: Vehicle, exhibition: boolean) => {

    setMessage("");

    return api

      .updateVehicle(vehicle.id, { ...vehicleToForm(vehicle), exhibition })

      .then((updated) => {

        const nextVehicle = {

          ...updated,

          images: updated.images ?? vehicle.images,

          attachments: updated.attachments ?? vehicle.attachments

        };

        setVehicles((current) => current.map((item) => (item.id === nextVehicle.id ? nextVehicle : item)));

        if (selected?.id === nextVehicle.id) {

          setSelectedDetail(nextVehicle);

        }

        return nextVehicle;

      });

  };



  const loadAssignmentEntries = (listID: string) => {

    if (!listID) {

      setExhibitionAssignment((current) => current ? { ...current, selectedListID: "", entries: [], loadingEntries: false, error: "" } : current);

      return;

    }

    setExhibitionAssignment((current) => current ? { ...current, selectedListID: listID, loadingEntries: true, error: "" } : current);

    api

      .exhibitionEntries(listID)

      .then((entries) => {

        setExhibitionAssignment((current) => current && current.selectedListID === listID ? { ...current, entries, loadingEntries: false, error: "" } : current);

      })

      .catch((error: Error) => {

        setExhibitionAssignment((current) => current ? { ...current, entries: [], loadingEntries: false, error: error.message } : current);

      });

  };



  const openExhibitionAssignment = (vehicle: Vehicle) => {

    if (!vehicleExhibitionEligible(vehicle)) {

      setMessage(t("vehicles.exhibition.requiresDecoder"));

      return;

    }

    setMessage("");

    setExhibitionAssignment({

      vehicle,

      lists: [],

      selectedListID: "",

      entries: [],

      loadingEntries: true,

      saving: false,

      error: ""

    });

    api

      .exhibitionLists()

      .then((lists) => {

        const availableLists = lists.filter((list) => !list.locked);

        const firstListID = availableLists[0]?.id || "";

        setExhibitionAssignment((current) => current ? {

          ...current,

          lists: availableLists,

          selectedListID: firstListID,

          loadingEntries: Boolean(firstListID),

          error: firstListID ? "" : t("vehicles.exhibition.noOpenLists")

        } : current);

        if (firstListID) {

          return api.exhibitionEntries(firstListID).then((entries) => {

            setExhibitionAssignment((current) => current && current.selectedListID === firstListID ? { ...current, entries, loadingEntries: false, error: "" } : current);

          });

        }

        return undefined;

      })

      .catch((error: Error) => {

        setExhibitionAssignment((current) => current ? { ...current, loadingEntries: false, error: error.message } : current);

      });

  };



  const duplicateAssignmentVehicle = exhibitionAssignment

    ? exhibitionAssignment.entries.find((entry) =>

      normalizedText(entry.owner) === normalizedText(username) &&

      normalizedText(entry.locomotiveName) === normalizedText(exhibitionAssignment.vehicle.name)

    )

    : undefined;

  const duplicateAssignmentDecoder = exhibitionAssignment?.vehicle.digitalDecoderNumber

    ? exhibitionAssignment.entries.find((entry) => normalizedText(entry.decoderNumber) === normalizedText(exhibitionAssignment.vehicle.digitalDecoderNumber))

    : undefined;



  const saveExhibitionAssignment = () => {

    if (!exhibitionAssignment || !exhibitionAssignment.selectedListID || duplicateAssignmentVehicle || duplicateAssignmentDecoder) return;

    setExhibitionAssignment((current) => current ? { ...current, saving: true, error: "" } : current);

    api

      .vehicle(exhibitionAssignment.vehicle.id)

      .then((detail) => api.createExhibitionEntry(exhibitionAssignment.selectedListID, vehicleToExhibitionEntry(detail, username)).then(() => detail))

      .then((detail) => updateVehicleExhibitionFlag(detail, true))

      .then(() => {

        setExhibitionAssignment(null);

        setMessage(t("vehicles.exhibition.assigned"));

      })

      .catch((error: Error) => {

        setExhibitionAssignment((current) => current ? { ...current, saving: false, error: error.message } : current);

      });

  };



  const toggleVehicleExhibition = (vehicle: Vehicle, checked: boolean) => {

    if (checked) {

      openExhibitionAssignment(vehicle);

      return;

    }

    updateVehicleExhibitionFlag(vehicle, false)

      .then(() => setMessage(t("vehicles.exhibition.disabled")))

      .catch((error: Error) => setMessage(error.message));

  };



  const openQrForVehicle = (vehicle: Vehicle) => {

    setQrSvg("");

    setQrError("");

    api

      .vehicle(vehicle.id)

      .then(async (detail) => {

        setSelectedDetail(detail);

        setQrDialogOpen(true);

        setQrSvg(await buildQrSvg(detail, vehicleToForm(detail)));

      })

      .catch((error: Error) => setQrError(error.message));

  };



  const markECoSImportSessionSaved = (draft: ECoSVehicleDraftPayload, vehicleId: string) => {

    if (!draft.returnToEcos) return;

    try {

      const rawSession = window.sessionStorage.getItem(ecosImportSessionStorageKey);

      if (!rawSession) return;

      const session = JSON.parse(rawSession) as {

        id?: string;

        statuses?: Record<string, { status: string; vehicleId?: string; updatedAt?: string }>;

        updatedAt?: string;

      };

      if (session.id !== draft.returnToEcos.sessionId) return;

      const key = String(draft.returnToEcos.objectId);

      const now = new Date().toISOString();

      const nextSession = {

        ...session,

        updatedAt: now,

        statuses: {

          ...(session.statuses || {}),

          [key]: {

            ...(session.statuses || {})[key],

            status: "saved",

            vehicleId,

            updatedAt: now

          }

        }

      };

      window.sessionStorage.setItem(ecosImportSessionStorageKey, JSON.stringify(nextSession));

    } catch {

      window.sessionStorage.removeItem(ecosImportSessionStorageKey);

    }

  };



  const returnToECoSImportSession = () => {

    window.history.pushState(null, "", "/import-export?source=ecos");

    window.dispatchEvent(new PopStateEvent("popstate"));

  };

  const isRemotePendingImage = (image: PendingArticleImage) => !image.persisted && /^https?:\/\//i.test(image.url);



  const submit = async (event: FormEvent) => {

    event.preventDefault();
    setSaveAttempted(true);

    if (missingRequiredModelFields.length > 0) {

      setActiveTab("model");

      setOpenSections((current) => ({ ...current, model: true }));

      setMessage(t("vehicles.requiredMissing", { fields: missingRequiredModelFields.map((field) => field.label).join(", ") }));

      return;

    }

    if (ecosDraft && ecosUnclearFields.size > 0) {

      setMessage(t("vehicles.ecosDraft.unresolved", { count: ecosUnclearFields.size }));

      return;

    }



    setSaving(true);

    setMessage("");



    try {

      const sparePartsToImport = selectedFoundSparePartInputs();

      const remoteImages = pendingArticleImages.filter(isRemotePendingImage);

      const images = pendingArticleImages.filter((image) => !isRemotePendingImage(image)).map((image, index) => ({

        id: image.persisted ? image.id : undefined,

        url: image.url,

        title: image.title,

        sourceUrl: image.source,

        maintenanceId: image.maintenanceId || "",

        isPrimary: Boolean(image.isPrimary),

        sortOrder: index

      }));

      const payload = { ...form, images };

      let vehicle = mode === "edit" && selected

        ? await api.updateVehicle(selected.id, payload)

        : await api.createVehicle(payload);

      if (remoteImages.length > 0) {

        for (const [imageIndex, image] of remoteImages.entries()) {

          await api.importVehicleImageFromUrl(vehicle.id, {

            url: image.url,

            title: image.title || "",

            sourceUrl: image.source || image.url,

            maintenanceId: image.maintenanceId || "",

            isPrimary: Boolean(image.isPrimary),

            sortOrder: images.length + imageIndex

          });

        }

        vehicle = await api.vehicle(vehicle.id);

      }

      if (sparePartsToImport.length > 0) {

        for (const part of sparePartsToImport) {

          await api.createVehicleSparePart(vehicle.id, part);

        }

        clearSelectedFoundSpareParts();

        vehicle = await api.vehicle(vehicle.id);

      }



      if (ecosDraft && (mode === "create" || mode === "edit")) {

        await api.upsertVehicleExternalMapping(vehicle.id, ecosDraft.externalMapping);

        const detailBeforeECoSValues = await api.vehicle(vehicle.id);

        for (const cvValue of ecosDraft.cvValues) {

          const existing = (detailBeforeECoSValues.cvValues || []).find((entry) => cvValueKey(entry) === cvValueKey(cvValue));

          if (existing) {

            await api.updateVehicleCVValue(vehicle.id, existing.id, cvValue);

          } else {

            await api.createVehicleCVValue(vehicle.id, cvValue);

          }

        }

        for (const functionKey of configuredFunctionKeys) {

          const edit = functionEdit(functionKey);

          if (!edit.name?.trim() && !edit.symbolKey && !edit.notes?.trim()) {

            continue;

          }

          await api.updateVehicleFunction(vehicle.id, functionKey, {

            name: edit.name || "",

            symbolKey: edit.symbolKey || "",

            functionType: edit.functionType || "standard",

            mode: edit.mode || "dauer",

            directionDependent: Boolean(edit.directionDependent),

            notes: edit.notes || ""

          });

        }

        markECoSImportSessionSaved(ecosDraft, vehicle.id);

        vehicle = await api.vehicle(vehicle.id);

        setEcosDraft(null);

        setEcosUnclearFields(new Set());

        if (ecosDraft.returnToEcos) {

          load();

          closeModal();

          returnToECoSImportSession();

          return;

        }

      }



      vehicle = await api.vehicle(vehicle.id);

      setSelectedDetail(vehicle);

      setMode("edit");

      setSaveAttempted(false);

      load();

      if (mode === "create") {

        setMessage(t("vehicles.createdContinue"));

      } else if (sparePartsToImport.length > 0) {

        setMessage(t("vehicles.spareParts.importedCount", { count: sparePartsToImport.length }));

      }

    } catch (error) {

      setMessage(error instanceof Error ? error.message : String(error));

    } finally {

      setSaving(false);

    }

  };



  const confirmDelete = () => {

    if (!deleteCandidate) return;



    api

      .deleteVehicle(deleteCandidate.id)

      .then(() => {

        if (selected?.id === deleteCandidate.id) {

          closeModal();

        }

        setDeleteCandidate(null);

        load();

      })

      .catch((error: Error) => setMessage(error.message));

  };



  const sortHeader = (key: SortKey) => (

    <button

      type="button"

      className={`sort-button ${sort.key === key ? "active" : ""}`}

      onClick={() => toggleSort(key)}

      title={t("common.sort", { label: t(`vehicle.field.${key}`) })}

    >

      {t(`vehicle.field.${key}`)}

      {sort.key === key

        ? sort.direction === "asc"

          ? <ChevronUp size={14} />

          : <ChevronDown size={14} />

        : <ArrowUpDown size={13} />}

    </button>

  );



  const vehicleQuickMenu = (vehicle: Vehicle) => (

    <div className="quick-menu-wrap">

      <button

        type="button"

        className={quickMenuVehicleID === vehicle.id ? "icon-button active" : "icon-button"}

        onClick={() => setQuickMenuVehicleID((current) => current === vehicle.id ? "" : vehicle.id)}

        aria-label={t("vehicles.quickMenu")}

        title={t("vehicles.quickMenu")}

      >

        <MoreVertical size={16} />

      </button>

      {quickMenuVehicleID === vehicle.id && (

        <div className="quick-menu" role="menu">

          <button type="button" role="menuitem" onClick={() => { setQuickMenuVehicleID(""); openDetail(vehicle); }}><Eye size={14} />{t("vehicles.view")}</button>

          <button type="button" role="menuitem" onClick={() => { setQuickMenuVehicleID(""); openEdit(vehicle); }}><Pencil size={14} />{t("vehicles.edit")}</button>

          <span className="quick-menu-separator" role="separator" />

          <button type="button" role="menuitem" onClick={() => { setQuickMenuVehicleID(""); openQrForVehicle(vehicle); }}><QrCode size={14} />QR-Code</button>

          <button type="button" role="menuitem" onClick={() => { setQuickMenuVehicleID(""); printVehicleReport(vehicle); }}><Printer size={14} />{t("overview.print")}</button>

          <button type="button" role="menuitem" onClick={() => { setQuickMenuVehicleID(""); openEdit(vehicle, "uploads"); }}><Upload size={14} />Uploads</button>

          <button type="button" role="menuitem" onClick={() => { setQuickMenuVehicleID(""); openEdit(vehicle, "maintenance"); }}><Wrench size={14} />{t("vehicles.maintenance")}</button>

          <button type="button" role="menuitem" onClick={() => { setQuickMenuVehicleID(""); openEdit(vehicle, "spareParts"); }}><PackageSearch size={14} />{t("vehicles.tab.spareParts")}</button>

          <span className="quick-menu-separator" role="separator" />

          <button type="button" role="menuitem" className="danger" onClick={() => { setQuickMenuVehicleID(""); setDeleteCandidate(vehicle); }}><Trash2 size={14} />{t("vehicles.delete")}</button>

        </div>

      )}

    </div>

  );



  const selectOptions = (items: MasterDataEntry[], emptyLabel = "Keine Auswahl") => (

    <>

      <option value="">{emptyLabel}</option>

      {items.map((entry) => (

        <option key={entry.key} value={optionValue(entry)}>

          {entry.label}

        </option>

      ))}

    </>

  );



  const maintenanceEntries = selected?.maintenance || [];

  const maintenanceSummary = {

    due: maintenanceEntries.filter(maintenanceIsDue).length,

    planned: maintenanceEntries.filter((entry) => entry.status !== "erledigt").length,

    done: maintenanceEntries.filter((entry) => entry.status === "erledigt").length

  };

  const configuredFunctionKeys = functionKeys.filter((functionKey) => {

    const edit = functionEdit(functionKey);

    return Boolean(edit.persisted || edit.name || edit.symbolKey || edit.notes);

  });

  const visibleFunctionKeys = showConfiguredFunctionsOnly ? configuredFunctionKeys : functionKeys;

  const functionSummary = {

    configured: configuredFunctionKeys.length,

    sound: configuredFunctionKeys.filter((functionKey) => functionEdit(functionKey).functionType === "sound").length,

    light: configuredFunctionKeys.filter((functionKey) => functionEdit(functionKey).functionType === "licht").length

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

        <div className="modal-layer" role="dialog" aria-modal="true" aria-label={t("vehicles.modal.aria")}>

          <form key={`${mode}-${selected?.id || "new"}`} className={mode === "view" ? "vehicle-modal vehicle-read-modal-shell" : "vehicle-modal"} onSubmit={submit}>

            <header className="modal-head">

              <h2>{mode === "create" ? t("vehicles.modal.create") : mode === "edit" ? t("vehicles.modal.edit") : t("vehicles.modal.view")}</h2>

              <button type="button" className="icon-button" onClick={closeModal} aria-label={t("vehicles.close")} title={t("vehicles.close")}>

                <X size={18} />

              </button>

            </header>



            {mode === "view" && selected ? (

              <VehicleReadOnlyView

                vehicle={selected}

                onEdit={() => openEdit(selected)}

                onPrint={() => printVehicleReport(selected)}

                onQr={generateQr}

                onPreviewImage={setPreviewImage}

              />

            ) : (

              <>

            <nav className="modal-tabs" aria-label={t("vehicles.modal.aria")}>

              <button type="button" className={activeTab === "model" ? "active" : ""} onClick={() => setActiveTab("model")}>

                {t("vehicles.tab.model")}

              </button>

              <button type="button" className={activeTab === "control" ? "active" : ""} onClick={() => setActiveTab("control")}>

                {t("vehicles.tab.control")}

              </button>

              <button type="button" className={activeTab === "speedCurve" ? "active" : ""} onClick={() => setActiveTab("speedCurve")}>

                {t("vehicles.tab.speedCurve")}

              </button>

              <button type="button" className={activeTab === "cv" ? "active" : ""} onClick={() => setActiveTab("cv")}>

                CV

              </button>

              <button type="button" className={activeTab === "uploads" ? "active" : ""} onClick={() => setActiveTab("uploads")}>

                {t("vehicles.tab.uploads")}

              </button>

              <button type="button" className={activeTab === "maintenance" ? "active" : ""} onClick={() => setActiveTab("maintenance")}>

                {t("vehicles.tab.maintenance")}

              </button>

              <button type="button" className={activeTab === "spareParts" ? "active" : ""} onClick={() => setActiveTab("spareParts")}>

                {t("vehicles.tab.spareParts")}

              </button>

            </nav>



            <div className="modal-body">

              {activeTab === "model" && (

                <VehicleModelTab

                  form={form}

                  externalMappings={selected?.externalMappings || []}

                  readonly={readonly}

                  articleSearchLoading={articleSearchLoading}

                  canRunArticleSearch={canRunArticleSearch}

                  showRequiredErrors={showRequiredErrors}

                  options={options}

                  filteredGattungen={filteredGattungen}

                  openSections={openSections}

                  selectOptions={selectOptions}

                  ecosFieldClass={ecosFieldClass}

                  onToggleSection={toggleSection}

                  onOpenBarcodeSearch={openBarcodeSearch}

                  onRunArticleSearch={() => runArticleSearch()}

                  onUpdate={update}

                  onUpdateCategory={updateCategory}

                  onOpenQr={generateQr}

                  canOpenQr={canGenerateQr}

                  onUpdateCouplingFront={updateCouplingFront}

                  onUpdateCouplingSame={updateCouplingSame}

                />

              )}

              {activeTab === "control" && (

                <VehicleFunctionsTab

                  selected={selected}

                  draftMode={Boolean(ecosDraft)}

                  readonly={readonly}

                  saving={saving}

                  functionImportInputRef={functionImportInputRef}

                  configuredFunctionKeys={configuredFunctionKeys}

                  functionSummary={functionSummary}

                  showConfiguredFunctionsOnly={showConfiguredFunctionsOnly}

                  visibleFunctionKeys={visibleFunctionKeys}

                  symbols={options.symbols}

                  onImportFunctions={importFunctions}

                  onExportFunctions={exportFunctions}

                  onShowConfiguredFunctionsOnlyChange={setShowConfiguredFunctionsOnly}

                  functionEdit={functionEdit}

                  updateFunctionEdit={updateFunctionEdit}

                  inferFunctionTypeFromSymbol={inferFunctionTypeFromSymbol}

                  saveFunction={saveFunction}

                  deleteFunction={deleteFunction}

                />

              )}

              {activeTab === "speedCurve" && (

                <VehicleSpeedCurveTab

                  selected={selected}

                  ecosDraft={ecosDraft}

                />

              )}

              {activeTab === "cv" && (

                <VehicleCVTab

                  selected={selected}

                  ecosDraft={ecosDraft}

                  readonly={readonly}

                  saving={saving}

                  cvImportInputRef={cvImportInputRef}

                  cvFileInputRef={cvFileInputRef}

                  cvSummary={cvSummary}

                  cvImportPreview={cvImportPreview}

                  cvImportStats={cvImportStats}

                  cvForm={cvForm}

                  editingCVID={editingCVID}

                  decoderProfileOptions={decoderProfileOptions}

                  storedDecoderProfiles={storedDecoderProfiles}

                  cvFileProfile={cvFileProfile}

                  cvFileDescription={cvFileDescription}

                  cvFileUploadPreview={cvFileUploadPreview}

                  cvFilePreviewStats={cvFilePreviewStats}

                  importCVValues={importCVValues}

                  exportCVValues={exportCVValues}

                  selectCVImportRows={selectCVImportRows}

                  applyCVImportPreview={applyCVImportPreview}

                  discardCVImportPreview={discardCVImportPreview}

                  toggleCVImportRow={toggleCVImportRow}

                  updateCVForm={updateCVForm}

                  resetCVForm={resetCVForm}

                  saveCVValue={saveCVValue}

                  editCVValue={editCVValue}

                  deleteCVValue={deleteCVValue}

                  uploadCVFiles={uploadCVFiles}

                  setCVFileProfile={setCVFileProfile}

                  setCVFileDescription={setCVFileDescription}

                  applyFirstCVFileSuggestion={applyFirstCVFileSuggestion}

                  previewCVFileValuesForImport={previewCVFileValuesForImport}

                  applyCVFileFunctionSuggestions={applyCVFileFunctionSuggestions}

                  confirmCVFileUpload={confirmCVFileUpload}

                  discardCVFileUploadPreview={discardCVFileUploadPreview}

                  deleteCVFile={deleteCVFile}

                />

              )}

              {activeTab === "uploads" && (

                <VehicleUploadsTab

                  selected={selected}

                  readonly={readonly}

                  saving={saving}

                  imageInputRef={imageInputRef}

                  attachmentInputRef={attachmentInputRef}

                  maintenanceEntries={maintenanceEntries}

                  imageUploadMaintenanceID={imageUploadMaintenanceID}

                  pendingArticleImages={pendingArticleImages}

                  attachmentDragActive={attachmentDragActive}

                  attachmentUploadCategory={attachmentUploadCategory}

                  attachmentUploadDescription={attachmentUploadDescription}

                  attachmentEdits={attachmentEdits}

                  documentSearchLoading={documentSearchLoading}

                  documentSearchError={documentSearchError}

                  documentSearchRan={documentSearchRan}

                  foundDocuments={foundUploadDocuments}

                  selectedDocuments={selectedUploadDocuments}

                  onImageUploadMaintenanceIDChange={setImageUploadMaintenanceID}

                  onUploadImages={uploadImages}

                  onPreviewImage={setPreviewImage}

                  onUpdatePendingImageTitle={updatePendingImageTitle}

                  onUpdatePendingImageMaintenance={updatePendingImageMaintenance}

                  onMovePendingImage={movePendingImage}

                  onSetPrimaryPendingImage={setPrimaryPendingImage}

                  onRemovePendingImage={removePendingImage}

                  onUploadAttachment={uploadAttachment}

                  onAttachmentDrag={onAttachmentDrag}

                  onAttachmentDrop={onAttachmentDrop}

                  onAttachmentUploadCategoryChange={setAttachmentUploadCategory}

                  onAttachmentUploadDescriptionChange={setAttachmentUploadDescription}

                  onUpdateAttachmentEdit={updateAttachmentEdit}

                  onSaveAttachment={saveAttachment}

                  onExtractAttachmentSpareParts={extractAttachmentSpareParts}

                  onDeleteAttachment={setAttachmentDeleteCandidate}

                  onSearchDocuments={runUploadDocumentSearch}

                  onImportDocument={importFoundDocument}

                  onImportSelectedDocuments={importSelectedFoundDocuments}

                  onToggleDocument={toggleFoundDocument}

                  onToggleAllDocuments={toggleAllFoundDocuments}

                />

              )}

              {activeTab === "maintenance" && (

                <VehicleMaintenanceTab

                  selected={selected}

                  pendingArticleImages={pendingArticleImages}

                  readonly={readonly}

                  saving={saving}

                  maintenanceForm={maintenanceForm}

                  editingMaintenanceID={editingMaintenanceID}

                  maintenanceSummary={maintenanceSummary}

                  onUpdateMaintenanceForm={updateMaintenanceForm}

                  onResetMaintenanceForm={resetMaintenanceForm}

                  onSaveMaintenance={saveMaintenance}

                  onCompleteMaintenance={completeMaintenance}

                  onEditMaintenance={editMaintenance}

                  onDeleteMaintenance={deleteMaintenance}

                />

              )}

              {activeTab === "spareParts" && (

                <VehicleSparePartsTab

                  selected={selected}

                  readonly={readonly}

                  saving={saving}

                  sparePartForm={sparePartForm}

                  editingSparePartID={editingSparePartID}

                  sparePartLookupLoadingID={sparePartLookupLoadingID}

                  sparePartLookupErrors={sparePartLookupErrors}

                  sparePartLookupResults={sparePartLookupResults}

                  sparePartStatusLoading={sparePartStatusLoading}

                  sparePartStatuses={sparePartStatuses}

                  importAllSparePartsLoading={importAllSparePartsLoading}

                  canImportAllSpareParts={canImportAllAvailableSpareParts}

                  importAllSparePartsTitle={importAllAvailableSparePartsTitle}

                  onUpdateSparePartForm={updateSparePartForm}

                  onResetSparePartForm={resetSparePartForm}

                  onSaveSparePart={saveSparePart}

                  onEditSparePart={editSparePart}

                  onDeleteSparePart={deleteSparePart}

                  onSearchSparePart={searchSingleSparePart}

                  onApplySparePartLookup={applySparePartLookup}

                  onImportAllSpareParts={importAllAvailableSpareParts}

                  sparePartSort={sparePartSort}

                  onToggleSparePartSort={toggleSparePartSort}

                />

              )}            </div>



            <footer className="modal-actions">

              {message && <p className="form-message">{message}</p>}

              <button type="button" className="secondary-button" onClick={closeModal}>

                {t("vehicles.cancel")}

              </button>

              <button className="primary-button" disabled={saving}>

                {saving ? t("vehicles.saving") : mode === "create" ? t("vehicles.createAndContinue") : t("vehicles.saveChanges")}

              </button>

            </footer>

              </>

            )}

          </form>

        </div>

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

          onClose={() => setQrDialogOpen(false)}

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

          onClose={() => setExhibitionAssignment(null)}

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
