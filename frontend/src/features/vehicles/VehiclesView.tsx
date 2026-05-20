import { DragEvent, FormEvent, useCallback, useEffect, useMemo, useRef, useState } from "react";
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
  ArticleSearchImage,
  ArticleSearchInput,
  ArticleSearchResponse,
  ArticleSearchResult,
  CreateVehicleRequest,
  ExhibitionEntry,
  ExhibitionEntryInput,
  ExhibitionList,
  MasterDataEntry,
  MasterDataRelation,
  VehicleAttachment,
  VehicleCVFile,
  VehicleCVValue,
  VehicleCVValueInput,
  VehicleExternalMappingInput,
  VehicleFunctionInput,
  VehicleMaintenance,
  VehicleMaintenanceInput,
  Vehicle
} from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { ArticleSearchDialog } from "./ArticleSearchDialog";
import { BarcodeSearchDialog } from "./BarcodeSearchDialog";
import { DeleteVehicleDialog, ExhibitionAssignmentDialog, ImagePreviewDialog, QrDialog, ReportDialog } from "./VehicleDialogs";
import { VehicleInventoryPanel } from "./VehicleInventoryPanel";
import { VehicleFunctionsTab } from "./VehicleFunctionsTab";
import { VehicleMaintenanceTab } from "./VehicleMaintenanceTab";
import { VehicleModelTab } from "./VehicleModelTab";
import { VehicleUploadsTab } from "./VehicleUploadsTab";
import { VehicleCVTab } from "./VehicleCVTab";
import { VehicleReadOnlyView } from "./VehicleReadOnlyView";
import {
  ArticleFieldKey,
  articleSelectionKey,
  articleValueForForm,
  currentArticleValue,
  imageSelectionKey,
  isArticleFieldKey,
} from "./articleSearch";
import {
  buildCVImportPreview,
  commonDecoderProfiles,
  cvValueKey,
  cvValuesFromImport,
  functionKeys,
  functionMappingsFromImport,
  isValidCVValueInput,
  isValidFunctionMapping
} from "./cvImport";
import type { CVFileUploadPreview, CVImportPreview } from "./cvImport";
import {
  attachmentCategoryForFile,
  isAllowedImageFile,
  isBlockedAttachmentFile,
  isBlockedCVFile
} from "./vehicleFiles";
import {
  maintenanceDaysUntilDue,
  maintenanceIsDue,
  todayISODate
} from "./vehicleMaintenance";
import { buildBrandedQrPngDataUrl, buildQrSvg, downloadQrPngFile, downloadQrSvgFile, printQrSvgLabel, qrPayload } from "./vehicleQr";
import { InventoryReportAssets, inventoryReportHtml, openPrintDocument } from "./vehicleReports";
import {
  attachmentsToEditState,
  functionsToEditState,
  normalizedText,
  PendingArticleImage,
  previewImageUrl,
  primaryImage,
  uploadedImageToPending,
  vehicleExhibitionEligible,
  vehicleImagesToPending,
  vehicleToExhibitionEntry,
  vehicleToForm
} from "./vehicleTransforms";
import type { AttachmentEditState, FunctionEditState } from "./vehicleTransforms";

import {
  articleSearchEnabled,
  articleSearchSources,
  compactValue,
  emptyCVForm,
  emptyFunctionEdit,
  emptyMaintenanceForm,
  emptyOptions,
  emptyVehicle,
  ecosRequiredFields,
  ecosVehicleDraftStorageKey,
  fieldValue,
  hasArticleIdentity,
  hasArticleSearchCriteria,
  hasQrPayloadData,
  inferFunctionTypeFromSymbol,
  inventoryViewMode,
  inventoryViewSettingKey,
  isBadArticleValue,
  optionValue,
  sanitizeArticleSearchResponse,
  sortLabels,
  valueForSort,
  vehicleFieldsForSearch
} from "./vehicleViewModel";
import type {
  ECoSRequiredField,
  ECoSVehicleDraftPayload,
  ExhibitionAssignment,
  InventoryFilter,
  InventoryReportMode,
  InventoryViewMode,
  InventoryReportSelection,
  MaintenanceFilter,
  MaintenanceReminder,
  MasterDataOptions,
  ModalMode,
  ModalTab,
  SortDirection,
  SortKey
} from "./vehicleViewModel";

export function VehiclesView({ username }: { username: string }) {
  const { language, t } = useI18n();
  const [vehicles, setVehicles] = useState<Vehicle[]>([]);
  const [form, setForm] = useState<CreateVehicleRequest>(emptyVehicle);
  const [options, setOptions] = useState<MasterDataOptions>(emptyOptions);
  const [query, setQuery] = useState("");
  const [message, setMessage] = useState("");
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [selected, setSelected] = useState<Vehicle | null>(null);
  const [mode, setMode] = useState<ModalMode>("create");
  const [modalOpen, setModalOpen] = useState(false);
  const [activeTab, setActiveTab] = useState<ModalTab>("model");
  const [openSections, setOpenSections] = useState({
    model: true,
    details: false,
    vehicle: false
  });
  const [deleteCandidate, setDeleteCandidate] = useState<Vehicle | null>(null);
  const [articleSearchOpen, setArticleSearchOpen] = useState(false);
  const [articleSearchLoading, setArticleSearchLoading] = useState(false);
  const [articleSearchResponse, setArticleSearchResponse] = useState<ArticleSearchResponse | null>(null);
  const [articleSearchError, setArticleSearchError] = useState("");
  const [barcodeSearchOpen, setBarcodeSearchOpen] = useState(false);
  const [barcodeSearchValue, setBarcodeSearchValue] = useState("");
  const [selectedArticleFields, setSelectedArticleFields] = useState<Record<string, boolean>>({});
  const [selectedArticleImages, setSelectedArticleImages] = useState<Record<string, boolean>>({});
  const [pendingArticleImages, setPendingArticleImages] = useState<PendingArticleImage[]>([]);
  const [previewImage, setPreviewImage] = useState<PendingArticleImage | null>(null);
  const [attachmentEdits, setAttachmentEdits] = useState<AttachmentEditState>({});
  const [imageUploadMaintenanceID, setImageUploadMaintenanceID] = useState("");
  const [attachmentUploadCategory, setAttachmentUploadCategory] = useState("");
  const [attachmentUploadDescription, setAttachmentUploadDescription] = useState("");
  const [attachmentUploadMaintenanceID, setAttachmentUploadMaintenanceID] = useState("");
  const [attachmentDragActive, setAttachmentDragActive] = useState(false);
  const [maintenanceForm, setMaintenanceForm] = useState<VehicleMaintenanceInput>(emptyMaintenanceForm);
  const [editingMaintenanceID, setEditingMaintenanceID] = useState<string | null>(null);
  const [functionEdits, setFunctionEdits] = useState<FunctionEditState>({});
  const [showConfiguredFunctionsOnly, setShowConfiguredFunctionsOnly] = useState(false);
  const [cvForm, setCVForm] = useState<VehicleCVValueInput>(emptyCVForm);
  const [editingCVID, setEditingCVID] = useState<string | null>(null);
  const [cvFileProfile, setCVFileProfile] = useState("");
  const [cvFileDescription, setCVFileDescription] = useState("");
  const [cvImportPreview, setCVImportPreview] = useState<CVImportPreview | null>(null);
  const [cvFileUploadPreview, setCVFileUploadPreview] = useState<CVFileUploadPreview | null>(null);
  const [ecosDraft, setEcosDraft] = useState<ECoSVehicleDraftPayload | null>(null);
  const [ecosUnclearFields, setEcosUnclearFields] = useState<Set<ECoSRequiredField>>(() => new Set());
  const imageInputRef = useRef<HTMLInputElement | null>(null);
  const attachmentInputRef = useRef<HTMLInputElement | null>(null);
  const cvFileInputRef = useRef<HTMLInputElement | null>(null);
  const cvImportInputRef = useRef<HTMLInputElement | null>(null);
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
  const [selectedVehicleIDs, setSelectedVehicleIDs] = useState<Set<string>>(() => new Set());
  const [inventoryView, setInventoryView] = useState<InventoryViewMode>(inventoryViewMode);
  const [inventoryFilter, setInventoryFilter] = useState<InventoryFilter>("all");
  const [maintenanceFilter, setMaintenanceFilter] = useState<MaintenanceFilter>("all");
  const [manufacturerFilter, setManufacturerFilter] = useState("");
  const [categoryFilter, setCategoryFilter] = useState("");
  const [gattungFilter, setGattungFilter] = useState("");
  const [exhibitionReadyFilter, setExhibitionReadyFilter] = useState(false);
  const [exhibitionAssignment, setExhibitionAssignment] = useState<ExhibitionAssignment | null>(null);
  const [quickMenuVehicleID, setQuickMenuVehicleID] = useState("");
  const [sort, setSort] = useState<{ key: SortKey; direction: SortDirection }>({
    key: "inventoryNumber",
    direction: "asc"
  });

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
    const availableIDs = new Set(vehicles.map((vehicle) => vehicle.id));
    setSelectedVehicleIDs((current) => {
      const next = new Set(Array.from(current).filter((id) => availableIDs.has(id)));
      return next.size === current.size ? current : next;
    });
  }, [vehicles]);

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
    const nextForm = { ...emptyVehicle, ...draft.vehicle };
    setSelected(null);
    setMode("create");
    setForm(nextForm);
    setPendingArticleImages([]);
    setAttachmentEdits({});
    setImageUploadMaintenanceID("");
    setAttachmentUploadCategory("");
    setAttachmentUploadDescription("");
    setAttachmentUploadMaintenanceID("");
    setAttachmentDragActive(false);
    setFunctionEdits({});
    resetMaintenanceForm();
    resetCVForm();
    setCVImportPreview(null);
    setCVFileUploadPreview(null);
    setEcosDraft(draft);
    setEcosUnclearFields(new Set(draft.unclearFields));
    setActiveTab("model");
    setOpenSections({ model: true, details: false, vehicle: false });
    setModalOpen(true);
    setMessage(t("vehicles.ecosDraft.loaded"));
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

  const inventoryFilterCounts = useMemo(() => {
    const withImages = vehicles.filter((vehicle) => (vehicle.images || []).length > 0).length;
    const digital = vehicles.filter((vehicle) => vehicle.digital).length;
    const maintenanceDue = vehicles.filter((vehicle) => (vehicle.maintenance || []).some(maintenanceIsDue)).length;
    const exhibitionReady = vehicles.filter((vehicle) => vehicle.exhibitionReady).length;

    return {
      all: vehicles.length,
      digital,
      analog: vehicles.length - digital,
      withImages,
      withoutImages: vehicles.length - withImages,
      maintenanceDue,
      withoutMaintenance: vehicles.length - maintenanceDue,
      exhibitionReady
    };
  }, [vehicles]);

  const inventoryFilterOptions = useMemo(() => {
    const unique = (values: string[]) =>
      Array.from(new Set(values.map((value) => value.trim()).filter(Boolean))).sort((left, right) =>
        left.localeCompare(right, "de-DE", { sensitivity: "base" })
      );

    const gattungSource = categoryFilter ? vehicles.filter((vehicle) => vehicle.category === categoryFilter) : vehicles;

    return {
      manufacturers: unique(vehicles.map((vehicle) => vehicle.manufacturer)),
      categories: unique(vehicles.map((vehicle) => vehicle.category || "")),
      gattungen: unique(gattungSource.map((vehicle) => vehicle.gattung || ""))
    };
  }, [categoryFilter, vehicles]);

  const filteredVehicles = useMemo(() => {
    return vehicles.filter((vehicle) => {
      const maintenanceDue = (vehicle.maintenance || []).some(maintenanceIsDue);
      if (inventoryFilter === "digital" && !vehicle.digital) return false;
      if (inventoryFilter === "analog" && vehicle.digital) return false;
      if (inventoryFilter === "withImages" && (vehicle.images || []).length === 0) return false;
      if (inventoryFilter === "withoutImages" && (vehicle.images || []).length > 0) return false;
      if (maintenanceFilter === "due" && !maintenanceDue) return false;
      if (maintenanceFilter === "none" && maintenanceDue) return false;
      if (manufacturerFilter && vehicle.manufacturer !== manufacturerFilter) return false;
      if (categoryFilter && vehicle.category !== categoryFilter) return false;
      if (gattungFilter && vehicle.gattung !== gattungFilter) return false;
      if (exhibitionReadyFilter && !vehicle.exhibitionReady) return false;
      return true;
    });
  }, [categoryFilter, exhibitionReadyFilter, gattungFilter, inventoryFilter, maintenanceFilter, manufacturerFilter, vehicles]);

  const sortedVehicles = useMemo(() => {
    return [...filteredVehicles].sort((left, right) => {
      const result = valueForSort(left, sort.key).localeCompare(valueForSort(right, sort.key), "de-DE", {
        numeric: true,
        sensitivity: "base"
      });
      return sort.direction === "asc" ? result : -result;
    });
  }, [filteredVehicles, sort]);

  const selectedVisibleVehicles = useMemo(
    () => sortedVehicles.filter((vehicle) => selectedVehicleIDs.has(vehicle.id)),
    [selectedVehicleIDs, sortedVehicles]
  );
  const allVisibleSelected = sortedVehicles.length > 0 && sortedVehicles.every((vehicle) => selectedVehicleIDs.has(vehicle.id));
  const someVisibleSelected = sortedVehicles.some((vehicle) => selectedVehicleIDs.has(vehicle.id));
  const articleIdentityFilled = hasArticleIdentity(form);
  const canRunArticleSearch = hasArticleSearchCriteria(form);
  const canGenerateQr = hasQrPayloadData(selected, form);

  const maintenanceReminders = useMemo<MaintenanceReminder[]>(() => {
    return vehicles
      .flatMap((vehicle) =>
        (vehicle.maintenance || []).flatMap((entry) => {
          const daysUntilDue = maintenanceDaysUntilDue(entry);
          if (daysUntilDue === null || daysUntilDue > 14) return [];
          return [{ vehicle, entry, daysUntilDue }];
        })
      )
      .sort((left, right) => left.daysUntilDue - right.daysUntilDue || left.vehicle.inventoryNumber.localeCompare(right.vehicle.inventoryNumber, "de-DE"));
  }, [vehicles]);

  const maintenanceReminderSummary = {
    due: maintenanceReminders.filter((item) => item.daysUntilDue <= 0).length,
    upcoming: maintenanceReminders.filter((item) => item.daysUntilDue > 0).length
  };
  const nextMaintenanceReminder = maintenanceReminders[0];
  const inventorySummary = useMemo(() => {
    const categories = new Set(vehicles.map((vehicle) => vehicle.category).filter(Boolean));
    const digital = vehicles.filter((vehicle) => vehicle.digital).length;
    const withImages = vehicles.filter((vehicle) => (vehicle.images || []).length > 0).length;
    return {
      categories: categories.size,
      digital,
      analog: vehicles.length - digital,
      withImages
    };
  }, [vehicles]);

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

  const hasActiveInventoryFilters =
    inventoryFilter !== "all" ||
    maintenanceFilter !== "all" ||
    Boolean(manufacturerFilter || categoryFilter || gattungFilter || exhibitionReadyFilter);

  const resetInventoryFilters = () => {
    setInventoryFilter("all");
    setMaintenanceFilter("all");
    setManufacturerFilter("");
    setCategoryFilter("");
    setGattungFilter("");
    setExhibitionReadyFilter(false);
  };

  const toggleVehicleSelection = (vehicleID: string) => {
    setSelectedVehicleIDs((current) => {
      const next = new Set(current);
      if (next.has(vehicleID)) {
        next.delete(vehicleID);
      } else {
        next.add(vehicleID);
      }
      return next;
    });
  };

  const toggleAllVisibleSelection = () => {
    setSelectedVehicleIDs((current) => {
      const next = new Set(current);
      if (allVisibleSelected) {
        sortedVehicles.forEach((vehicle) => next.delete(vehicle.id));
      } else {
        sortedVehicles.forEach((vehicle) => next.add(vehicle.id));
      }
      return next;
    });
  };

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

  const readonly = mode === "view";

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

  const update = (patch: Partial<CreateVehicleRequest>) => {
    setForm((current) => {
      const next = { ...current, ...patch };
      syncECoSUnclearFields(next);
      return next;
    });
  };

  const setSelectedDetail = (detail: Vehicle) => {
    setEcosDraft(null);
    setEcosUnclearFields(new Set());
    setSelected(detail);
    setForm(vehicleToForm(detail));
    setPendingArticleImages(vehicleImagesToPending(detail));
    setAttachmentEdits(attachmentsToEditState(detail.attachments));
    setFunctionEdits(functionsToEditState(detail.functions));
    setEditingMaintenanceID(null);
    setMaintenanceForm(emptyMaintenanceForm);
    setEditingCVID(null);
    setCVForm(emptyCVForm);
    setCVImportPreview(null);
    setCVFileUploadPreview(null);
  };

  const updateCategory = (category: string) => {
    const categoryKey = options.categories.find((entry) => optionValue(entry) === category)?.key;
    const allowed = new Set(
      options.categoryRelations
        .filter((relation) => relation.parentKey === categoryKey)
        .map((relation) => relation.childKey)
    );
    const currentGattung = options.gattungen.find((entry) => optionValue(entry) === form.gattung);
    update({
      category,
      gattung: currentGattung && allowed.has(currentGattung.key) ? form.gattung : ""
    });
  };

  const updateCouplingFront = (couplingFront: string) => {
    update({
      couplingFront,
      couplingRear: form.couplingSame ? couplingFront : form.couplingRear
    });
  };

  const updateCouplingSame = (couplingSame: boolean) => {
    update({
      couplingSame,
      couplingRear: couplingSame ? form.couplingFront : form.couplingRear
    });
  };

  const runArticleSearch = (searchForm = form, searchInput?: ArticleSearchInput) => {
    if (!articleSearchEnabled()) {
      setArticleSearchError("Die Artikeldaten-Websuche ist in den Einstellungen deaktiviert.");
      setArticleSearchOpen(true);
      setArticleSearchResponse(null);
      return;
    }

    if (!hasArticleSearchCriteria(searchForm, searchInput)) {
      setArticleSearchOpen(false);
      setArticleSearchLoading(false);
      setArticleSearchResponse(null);
      setArticleSearchError(t("vehicles.articleSearch.missingInput"));
      setMessage(t("vehicles.articleSearch.missingInput"));
      return;
    }

    setArticleSearchOpen(true);
    setArticleSearchLoading(true);
    setArticleSearchError("");
    setArticleSearchResponse(null);
    setSelectedArticleFields({});
    setSelectedArticleImages({});

    api
      .articleSearch(searchInput ?? {
        manufacturer: searchForm.manufacturer,
        articleNumber: searchForm.articleNumber,
        name: searchForm.name,
        gauge: searchForm.gauge,
        searchSources: articleSearchSources(),
        fields: vehicleFieldsForSearch(searchForm)
      })
      .then((response) => {
        const sanitized = sanitizeArticleSearchResponse(response);
        setArticleSearchResponse(sanitized);
        const initialSelection: Record<string, boolean> = {};
        sanitized.results.forEach((result, index) => {
          Object.keys(result.fields).filter(isArticleFieldKey).forEach((key) => {
            initialSelection[articleSelectionKey(result, key, index)] = !currentArticleValue(searchForm, key);
          });
        });
        setSelectedArticleFields(initialSelection);
      })
      .catch((error: Error) => setArticleSearchError(error.message))
      .finally(() => setArticleSearchLoading(false));
  };

  const openBarcodeSearch = () => {
    setBarcodeSearchValue(form.ean || "");
    setBarcodeSearchOpen(true);
  };

  const submitBarcodeSearch = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const code = barcodeSearchValue.trim();
    if (!code) {
      setMessage("Bitte einen Barcode oder eine EAN eingeben.");
      return;
    }
    const nextForm = { ...form, ean: code };
    setForm(nextForm);
    setBarcodeSearchOpen(false);
    runArticleSearch(nextForm, {
      searchSources: articleSearchSources(),
      fields: {
        ean: code
      }
    });
  };

  const toggleArticleField = (result: ArticleSearchResult, index: number, key: string, checked: boolean) => {
    setSelectedArticleFields((current) => ({ ...current, [articleSelectionKey(result, key, index)]: checked }));
  };

  const toggleArticleImage = (result: ArticleSearchResult, index: number, image: ArticleSearchImage, checked: boolean) => {
    setSelectedArticleImages((current) => ({ ...current, [imageSelectionKey(result, image, index)]: checked }));
  };

  const setArticleFieldSelection = (modeName: "empty" | "all" | "none") => {
    if (!articleSearchResponse) return;
    const next: Record<string, boolean> = {};
    articleSearchResponse.results.forEach((result, index) => {
      Object.keys(result.fields).filter(isArticleFieldKey).forEach((key) => {
        const selectionKey = articleSelectionKey(result, key, index);
        next[selectionKey] = modeName === "all" || (modeName === "empty" && !currentArticleValue(form, key));
      });
    });
    setSelectedArticleFields(next);
  };

  const applyArticleResult = (result: ArticleSearchResult) => {
    const patch: Partial<CreateVehicleRequest> = {};
    const foundResultIndex = articleSearchResponse?.results.findIndex((entry) => entry.url === result.url) ?? 0;
    const resultIndex = foundResultIndex >= 0 ? foundResultIndex : 0;
    Object.entries(result.fields).forEach(([key, field]) => {
      if (!isArticleFieldKey(key) || !selectedArticleFields[articleSelectionKey(result, key, resultIndex)]) return;
      if (isBadArticleValue(key, field.value)) return;
      Object.assign(patch, { [key]: articleValueForForm(key, field.value) });
    });
    const selectedImages = (result.images || [])
      .filter((image) => selectedArticleImages[imageSelectionKey(result, image, resultIndex)])
      .map((image, imageIndex) => ({ ...image, id: `${result.url}-${image.url}`, isPrimary: pendingArticleImages.length === 0 && imageIndex === 0 }));
    if (selectedImages.length > 0) {
      setPendingArticleImages((current) => {
        const existing = new Set(current.map((image) => image.url));
        const next = [...current, ...selectedImages.filter((image) => !existing.has(image.url))];
        if (!next.some((image) => image.isPrimary) && next.length > 0) {
          next[0] = { ...next[0], isPrimary: true };
        }
        return next;
      });
    }
    update(patch);
    setArticleSearchOpen(false);
  };

  const setPrimaryPendingImage = (id: string) => {
    setPendingArticleImages((current) => current.map((image) => ({ ...image, isPrimary: image.id === id })));
  };

  const updatePendingImageTitle = (id: string, title: string) => {
    setPendingArticleImages((current) => current.map((image) => (image.id === id ? { ...image, title } : image)));
  };

  const updatePendingImageMaintenance = (id: string, maintenanceId: string) => {
    setPendingArticleImages((current) => current.map((image) => (image.id === id ? { ...image, maintenanceId } : image)));
  };

  const movePendingImage = (id: string, direction: -1 | 1) => {
    setPendingArticleImages((current) => {
      const index = current.findIndex((image) => image.id === id);
      const target = index + direction;
      if (index < 0 || target < 0 || target >= current.length) return current;
      const next = [...current];
      [next[index], next[target]] = [next[target], next[index]];
      return next;
    });
  };

  const removePendingImage = (image: PendingArticleImage) => {
    if (image.maintenanceId) {
      setMessage("Bild ist mit einer Wartung verknüpft. Bitte zuerst die Verknüpfung entfernen und speichern.");
      return;
    }
    const removeFromState = () => {
      setPendingArticleImages((current) => {
        const next = current.filter((entry) => entry.id !== image.id);
        if (next.length > 0 && !next.some((entry) => entry.isPrimary)) {
          next[0] = { ...next[0], isPrimary: true };
        }
        return next;
      });
    };

    if (selected && image.persisted) {
      setSaving(true);
      api
        .deleteVehicleImage(selected.id, image.id)
        .then(() => {
          removeFromState();
          refreshSelectedVehicle(selected.id);
        })
        .catch((error: Error) => setMessage(error.message))
        .finally(() => setSaving(false));
      return;
    }
    removeFromState();
  };

  const refreshSelectedVehicle = (vehicleID = selected?.id) => {
    if (!vehicleID) return;
    api
      .vehicle(vehicleID)
      .then((detail) => {
        setSelectedDetail(detail);
        load();
      })
      .catch((error: Error) => setMessage(error.message));
  };

  const uploadImages = (files: FileList | null) => {
    if (!selected || !files || files.length === 0) return;
    const uploadFiles = Array.from(files);
    const invalid = uploadFiles.find((file) => !isAllowedImageFile(file));
    if (invalid) {
      setMessage(`${invalid.name} ist kein erlaubtes Bildformat.`);
      if (imageInputRef.current) {
        imageInputRef.current.value = "";
      }
      return;
    }
    setSaving(true);
    setMessage("");
    (async () => {
      for (const file of uploadFiles) {
        const image = await api.uploadVehicleImage(selected.id, file, file.name, pendingArticleImages.length === 0, imageUploadMaintenanceID);
        setPendingArticleImages((current) => {
          const next = [...current, uploadedImageToPending(image)];
          if (!next.some((entry) => entry.isPrimary) && next.length > 0) {
            next[0] = { ...next[0], isPrimary: true };
          }
          return next;
        });
      }
    })()
      .then(() => refreshSelectedVehicle(selected.id))
      .then(() => {
        setCVFileProfile("");
        setCVFileDescription("");
      })
      .catch((error: Error) => setMessage(error.message))
      .finally(() => {
        setSaving(false);
        if (imageInputRef.current) {
          imageInputRef.current.value = "";
        }
      });
  };

  const uploadAttachment = (files: FileList | null) => {
    if (!selected || !files || files.length === 0) return;
    const uploadFiles = Array.from(files);
    const blocked = uploadFiles.find(isBlockedAttachmentFile);
    if (blocked) {
      setMessage(`${blocked.name} ist als Beilage nicht erlaubt. Erlaubt sind PDF, TXT, CSV, JSON, XML, ZIP sowie JPG, PNG und WebP.`);
      if (attachmentInputRef.current) {
        attachmentInputRef.current.value = "";
      }
      return;
    }
    setSaving(true);
    setMessage("");
    (async () => {
      for (const file of uploadFiles) {
        await api.uploadVehicleAttachment(
          selected.id,
          file,
          attachmentUploadCategory || attachmentCategoryForFile(file),
          attachmentUploadDescription,
          attachmentUploadMaintenanceID
        );
      }
    })()
      .then(() => refreshSelectedVehicle(selected.id))
      .catch((error: Error) => setMessage(error.message))
      .finally(() => {
        setSaving(false);
        if (attachmentInputRef.current) {
          attachmentInputRef.current.value = "";
        }
      });
  };

  const onAttachmentDrag = (event: DragEvent<HTMLElement>) => {
    event.preventDefault();
    event.stopPropagation();
    if (readonly || !selected || saving) return;
    setAttachmentDragActive(event.type === "dragenter" || event.type === "dragover");
  };

  const onAttachmentDrop = (event: DragEvent<HTMLElement>) => {
    event.preventDefault();
    event.stopPropagation();
    setAttachmentDragActive(false);
    if (readonly || !selected || saving) return;
    uploadAttachment(event.dataTransfer.files);
  };

  const updateAttachmentEdit = (attachmentID: string, patch: Partial<{ description: string; category: string; maintenanceId: string }>) => {
    setAttachmentEdits((current) => ({
      ...current,
      [attachmentID]: {
        description: current[attachmentID]?.description || "",
        category: current[attachmentID]?.category || "",
        maintenanceId: current[attachmentID]?.maintenanceId || "",
        ...patch
      }
    }));
  };

  const saveAttachment = (attachment: VehicleAttachment) => {
    if (!selected) return;
    const edit = attachmentEdits[attachment.id] || { description: "", category: "", maintenanceId: "" };
    setSaving(true);
    api
      .updateVehicleAttachment(selected.id, attachment.id, edit)
      .then(() => refreshSelectedVehicle(selected.id))
      .catch((error: Error) => setMessage(error.message))
      .finally(() => setSaving(false));
  };

  const deleteAttachment = (attachment: VehicleAttachment) => {
    if (!selected) return;
    setSaving(true);
    api
      .deleteVehicleAttachment(selected.id, attachment.id)
      .then(() => refreshSelectedVehicle(selected.id))
      .catch((error: Error) => setMessage(error.message))
      .finally(() => setSaving(false));
  };

  const updateMaintenanceForm = (patch: Partial<VehicleMaintenanceInput>) => {
    setMaintenanceForm((current) => ({ ...current, ...patch }));
  };

  const resetMaintenanceForm = () => {
    setMaintenanceForm(emptyMaintenanceForm);
    setEditingMaintenanceID(null);
  };

  const editMaintenance = (entry: VehicleMaintenance) => {
    setMaintenanceForm({
      kind: entry.kind || "Wartung",
      status: entry.status || "geplant",
      conditionRating: entry.conditionRating || "",
      dueDate: entry.dueDate || "",
      completedAt: entry.completedAt || "",
      cost: entry.cost || "",
      notes: entry.notes || ""
    });
    setEditingMaintenanceID(entry.id);
  };

  const saveMaintenance = () => {
    if (!selected) return;
    setSaving(true);
    setMessage("");
    const payload: VehicleMaintenanceInput = {
      ...maintenanceForm,
      status: maintenanceForm.status === "fällig" ? "faellig" : maintenanceForm.status,
      cost: maintenanceForm.cost?.trim().replace(/\s*€$/, "") || "",
      completedAt: maintenanceForm.status === "erledigt" && !maintenanceForm.completedAt ? todayISODate() : maintenanceForm.completedAt
    };
    const action = editingMaintenanceID
      ? api.updateVehicleMaintenance(selected.id, editingMaintenanceID, payload)
      : api.createVehicleMaintenance(selected.id, payload);
    action
      .then(() => refreshSelectedVehicle(selected.id))
      .then(() => resetMaintenanceForm())
      .catch((error: Error) => setMessage(error.message))
      .finally(() => setSaving(false));
  };

  const completeMaintenance = (entry: VehicleMaintenance) => {
    if (!selected) return;
    setSaving(true);
    setMessage("");
    api
      .updateVehicleMaintenance(selected.id, entry.id, {
        kind: entry.kind,
        status: "erledigt",
        conditionRating: entry.conditionRating || "",
        dueDate: entry.dueDate || "",
        completedAt: entry.completedAt || todayISODate(),
        cost: entry.cost || "",
        notes: entry.notes || ""
      })
      .then(() => refreshSelectedVehicle(selected.id))
      .catch((error: Error) => setMessage(error.message))
      .finally(() => setSaving(false));
  };

  const deleteMaintenance = (entry: VehicleMaintenance) => {
    if (!selected) return;
    setSaving(true);
    setMessage("");
    api
      .deleteVehicleMaintenance(selected.id, entry.id)
      .then(() => refreshSelectedVehicle(selected.id))
      .catch((error: Error) => setMessage(error.message))
      .finally(() => setSaving(false));
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
          throw new Error("Keine gültigen Funktionszuordnungen gefunden.");
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

  const updateCVForm = (patch: Partial<VehicleCVValueInput>) => {
    setCVForm((current) => ({ ...current, ...patch }));
  };

  const resetCVForm = () => {
    setCVForm(emptyCVForm);
    setEditingCVID(null);
  };

  const editCVValue = (value: VehicleCVValue) => {
    setCVForm({
      cvNumber: value.cvNumber,
      value: value.value,
      description: value.description || "",
      category: value.category || "",
      protocol: value.protocol || "",
      decoderProfile: value.decoderProfile || "",
      sourceFileId: value.sourceFileId || ""
    });
    setEditingCVID(value.id);
  };

  const saveCVValue = () => {
    if (!selected) return;
    const payload = {
      ...cvForm,
      cvNumber: Number(cvForm.cvNumber),
      value: Number(cvForm.value)
    };
    if (!isValidCVValueInput(payload)) {
      setMessage("CV-Nummer muss 1-1024 und Wert 0-255 sein.");
      return;
    }
    setSaving(true);
    setMessage("");
    const existing = !editingCVID
      ? (selected.cvValues || []).find((entry) => cvValueKey(entry) === cvValueKey(payload))
      : undefined;
    const action = editingCVID
      ? api.updateVehicleCVValue(selected.id, editingCVID, payload)
      : existing
        ? api.updateVehicleCVValue(selected.id, existing.id, payload)
        : api.createVehicleCVValue(selected.id, payload);
    action
      .then(() => refreshSelectedVehicle(selected.id))
      .then(() => resetCVForm())
      .catch((error: Error) => setMessage(error.message))
      .finally(() => setSaving(false));
  };

  const deleteCVValue = (value: VehicleCVValue) => {
    if (!selected) return;
    setSaving(true);
    setMessage("");
    api
      .deleteVehicleCVValue(selected.id, value.id)
      .then(() => refreshSelectedVehicle(selected.id))
      .catch((error: Error) => setMessage(error.message))
      .finally(() => setSaving(false));
  };

  const exportCVValues = () => {
    if (!selected) return;
    const payload = {
      vehicle: {
        inventoryNumber: selected.inventoryNumber,
        name: selected.name,
        decoder: form.digitalDecoderNumber || form.dtDecoderNumber || ""
      },
      cvValues: selected.cvValues || []
    };
    const blob = new Blob([JSON.stringify(payload, null, 2)], { type: "application/json" });
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = `${selected.inventoryNumber || "railkeeper"}-cv.json`;
    anchor.click();
    URL.revokeObjectURL(url);
  };

  const importCVValues = (files: FileList | null) => {
    if (!selected || !files || files.length === 0) return;
    const [file] = Array.from(files);
    setSaving(true);
    setMessage("");
    file
      .text()
      .then(cvValuesFromImport)
      .then((values) => {
        const preview = buildCVImportPreview(file.name, values, selected.cvValues || []);
        if (!preview.rows.some((row) => row.status !== "invalid")) {
          throw new Error("Keine gültigen CV-Werte gefunden.");
        }
        setCVImportPreview(preview);
      })
      .catch((error: Error) => setMessage(error.message))
      .finally(() => {
        setSaving(false);
        if (cvImportInputRef.current) {
          cvImportInputRef.current.value = "";
        }
      });
  };

  const toggleCVImportRow = (id: string, selectedRow: boolean) => {
    setCVImportPreview((current) => current ? {
      ...current,
      rows: current.rows.map((row) => row.id === id ? { ...row, selected: selectedRow } : row)
    } : current);
  };

  const selectCVImportRows = (modeName: "all" | "none" | "empty") => {
    setCVImportPreview((current) => current ? {
      ...current,
      rows: current.rows.map((row) => ({
        ...row,
        selected: row.status !== "invalid" && (
          modeName === "all" ||
          (modeName === "empty" && row.status === "new")
        )
      }))
    } : current);
  };

  const applyCVImportPreview = () => {
    if (!selected || !cvImportPreview) return;
    const rows = cvImportPreview.rows.filter((row) => row.selected && row.status !== "invalid");
    if (rows.length === 0) {
      setMessage("Keine CV-Werte für den Import ausgewählt.");
      return;
    }
    setSaving(true);
    setMessage("");
    (async () => {
      for (const row of rows) {
        if (row.existing) {
          await api.updateVehicleCVValue(selected.id, row.existing.id, row.input);
        } else {
          await api.createVehicleCVValue(selected.id, row.input);
        }
      }
    })()
      .then(() => refreshSelectedVehicle(selected.id))
      .then(() => {
        setCVImportPreview(null);
        setMessage(`${rows.length} CV-Wert${rows.length === 1 ? "" : "e"} übernommen.`);
      })
      .catch((error: Error) => setMessage(error.message))
      .finally(() => setSaving(false));
  };

  const uploadCVFiles = (files: FileList | null) => {
    if (!selected || !files || files.length === 0) return;
    const uploadFiles = Array.from(files);
    const blocked = uploadFiles.find(isBlockedCVFile);
    if (blocked) {
      setMessage(`${blocked.name} ist als CV-Datei nicht erlaubt. Erlaubt sind JSON, CSV, TXT, XML, Z21, ESU, ESUX, LokProgrammer und ZIP.`);
      return;
    }
    setSaving(true);
    setMessage("");
    Promise.all(uploadFiles.map((file) => api.previewVehicleCVFile(file)))
      .then((previews) => {
        setCVFileUploadPreview({ files: uploadFiles, previews });
      })
      .catch((error: Error) => setMessage(error.message))
      .finally(() => {
        setSaving(false);
        if (cvFileInputRef.current) {
          cvFileInputRef.current.value = "";
        }
      });
  };

  const applyFirstCVFileSuggestion = () => {
    const suggestion = cvFileUploadPreview?.previews.find((preview) => preview.hasMetadata);
    if (!suggestion) return;
    if (suggestion.suggestedDecoderProfile) {
      setCVFileProfile(suggestion.suggestedDecoderProfile);
    }
    if (suggestion.suggestedDescription) {
      setCVFileDescription(suggestion.suggestedDescription);
    }
  };

  const previewCVFileValuesForImport = () => {
    if (!selected || !cvFileUploadPreview) return;
    const values = cvFileUploadPreview.previews.flatMap((preview) =>
      (preview.suggestedCvValues || []).map((value) => ({
        cvNumber: value.cvNumber,
        value: value.value,
        description: value.description || "",
        category: value.category || "",
        protocol: value.protocol || "",
        decoderProfile: preview.suggestedDecoderProfile || cvFileProfile || preview.decoder || preview.projectName || "",
        sourceFileId: ""
      }))
    );
    const preview = buildCVImportPreview("Decoder-Datei-Vorschau", values, selected.cvValues || []);
    if (!preview.rows.some((row) => row.status !== "invalid")) {
      setMessage("Keine gültigen CV-Werte in der Decoder-Vorschau gefunden.");
      return;
    }
    setCVImportPreview(preview);
    setMessage(`${values.length} erkannte CV-Werte für die Prüfung vorbereitet.`);
  };

  const applyCVFileFunctionSuggestions = () => {
    if (!selected || !cvFileUploadPreview) return;
    const mappings = cvFileUploadPreview.previews.flatMap((preview) =>
      (preview.suggestedFunctions || []).map((mapping) => ({
        functionKey: mapping.functionKey,
        name: mapping.name || "",
        symbolKey: "",
        functionType: mapping.functionType || "standard",
        mode: "dauer",
        directionDependent: false,
        notes: preview.fileName
      }))
    );
    const valid = Array.from(new Map(mappings.filter(isValidFunctionMapping).map((mapping) => [mapping.functionKey, mapping])).values());
    if (valid.length === 0) {
      setMessage("Keine gültigen Funktionstasten in der Decoder-Vorschau gefunden.");
      return;
    }
    setSaving(true);
    setMessage("");
    (async () => {
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
    })()
      .then(() => refreshSelectedVehicle(selected.id))
      .then(() => setMessage(`${valid.length} Funktionstaste${valid.length === 1 ? "" : "n"} aus der Decoder-Vorschau übernommen.`))
      .catch((error: Error) => setMessage(error.message))
      .finally(() => setSaving(false));
  };

  const confirmCVFileUpload = () => {
    if (!selected || !cvFileUploadPreview) return;
    const uploadFiles = cvFileUploadPreview.files;
    setSaving(true);
    setMessage("");
    (async () => {
      for (const file of uploadFiles) {
        await api.uploadVehicleCVFile(selected.id, file, cvFileProfile, cvFileDescription);
      }
    })()
      .then(() => refreshSelectedVehicle(selected.id))
      .then(() => {
        setCVFileUploadPreview(null);
        setMessage(`${uploadFiles.length} CV-Datei${uploadFiles.length === 1 ? "" : "en"} gespeichert.`);
      })
      .catch((error: Error) => setMessage(error.message))
      .finally(() => setSaving(false));
  };

  const deleteCVFile = (file: VehicleCVFile) => {
    if (!selected) return;
    setSaving(true);
    setMessage("");
    api
      .deleteVehicleCVFile(selected.id, file.id)
      .then(() => refreshSelectedVehicle(selected.id))
      .catch((error: Error) => setMessage(error.message))
      .finally(() => setSaving(false));
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
      setQrError(error instanceof Error ? error.message : "Druckfenster konnte nicht ge?ffnet werden.");
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
    if (reportVehicles.length === 0) {
      setMessage("Es gibt keine Fahrzeuge für den PDF-Report.");
      return;
    }
    try {
      const completeReportVehicles = await loadCompleteReportVehicles(reportVehicles);
      const assets = await buildInventoryReportAssets(completeReportVehicles);
      const html = inventoryReportHtml(completeReportVehicles, query, sort, {
        mode: reportMode,
        title: reportTitle.trim() || "Fahrzeugsammlung",
        includeQRCode: reportIncludeQRCode,
        includeImages: reportIncludeImages
      }, assets);
      openPrintDocument(html, `railkeeper-inventory-${reportMode}`);
      setReportDialogOpen(false);
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Report konnte nicht erstellt werden.");
    }
  };

  const printVehicleReport = async (vehicle: Vehicle) => {
    try {
      const completeVehicle = await api.vehicle(vehicle.id);
      const assets = await buildInventoryReportAssets([completeVehicle], true);
      const html = inventoryReportHtml([completeVehicle], completeVehicle.inventoryNumber || completeVehicle.name, sort, {
        mode: "details",
        title: completeVehicle.name || completeVehicle.inventoryNumber || "Fahrzeugsammlung",
        includeQRCode: true,
        includeImages: true
      }, assets);
      openPrintDocument(html, `railkeeper-vehicle-${completeVehicle.id}`);
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Report konnte nicht erstellt werden.");
    }
  };

  const toggleSort = (key: SortKey) => {
    setSort((current) => ({
      key,
      direction: current.key === key && current.direction === "asc" ? "desc" : "asc"
    }));
  };

  const setInventoryViewMode = (modeName: InventoryViewMode) => {
    setInventoryView(modeName);
    window.localStorage.setItem(inventoryViewSettingKey, modeName);
  };

  const openCreate = () => {
    setEcosDraft(null);
    setEcosUnclearFields(new Set());
    setSelected(null);
    setMode("create");
    setForm(emptyVehicle);
    setPendingArticleImages([]);
    setAttachmentEdits({});
    setImageUploadMaintenanceID("");
    setAttachmentUploadCategory("");
    setAttachmentUploadDescription("");
    setAttachmentUploadMaintenanceID("");
    setAttachmentDragActive(false);
    setFunctionEdits({});
    resetMaintenanceForm();
    resetCVForm();
    setCVImportPreview(null);
    setCVFileUploadPreview(null);
    setActiveTab("model");
    setOpenSections({ model: true, details: false, vehicle: false });
    setModalOpen(true);
    setMessage("");
  };

  const closeModal = () => {
    setEcosDraft(null);
    setEcosUnclearFields(new Set());
    setModalOpen(false);
    setSelected(null);
    setMode("create");
    setForm(emptyVehicle);
    setPendingArticleImages([]);
    setAttachmentEdits({});
    setImageUploadMaintenanceID("");
    setAttachmentUploadCategory("");
    setAttachmentUploadDescription("");
    setAttachmentUploadMaintenanceID("");
    setAttachmentDragActive(false);
    setFunctionEdits({});
    resetMaintenanceForm();
    resetCVForm();
    setCVImportPreview(null);
    setCVFileUploadPreview(null);
    setPreviewImage(null);
    setMessage("");
  };

  const updateVehicleExhibitionFlag = (vehicle: Vehicle, exhibition: boolean) => {
    setMessage("");
    return api
      .updateVehicle(vehicle.id, { ...vehicleToForm(vehicle), exhibition })
      .then((updated) => {
        setVehicles((current) => current.map((item) => (item.id === updated.id ? updated : item)));
        if (selected?.id === updated.id) {
          setSelectedDetail(updated);
        }
        return updated;
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

  const openDetail = (vehicle: Vehicle, tab: ModalTab = "model") => {
    api
      .vehicle(vehicle.id)
      .then((detail) => {
        setSelectedDetail(detail);
        setMode("view");
        setActiveTab(tab);
        setOpenSections({ model: true, details: false, vehicle: false });
        setModalOpen(true);
        setMessage("");
      })
      .catch((error: Error) => setMessage(error.message));
  };

  const openEdit = (vehicle: Vehicle) => {
    api
      .vehicle(vehicle.id)
      .then((detail) => {
        setSelectedDetail(detail);
        setMode("edit");
        setActiveTab("model");
        setOpenSections({ model: true, details: false, vehicle: false });
        setModalOpen(true);
        setMessage("");
      })
      .catch((error: Error) => setMessage(error.message));
  };

  const openQrForVehicle = (vehicle: Vehicle) => {
    setQrDialogOpen(true);
    setQrSvg("");
    setQrError("");
    api
      .vehicle(vehicle.id)
      .then(async (detail) => {
        setSelectedDetail(detail);
        setMode("view");
        setActiveTab("model");
        setOpenSections({ model: true, details: false, vehicle: false });
        setModalOpen(true);
        setQrSvg(await buildQrSvg(detail, vehicleToForm(detail)));
      })
      .catch((error: Error) => setQrError(error.message));
  };

  const toggleSection = (section: keyof typeof openSections) => {
    setOpenSections((current) => ({ ...current, [section]: !current[section] }));
  };

  const submit = async (event: FormEvent) => {
    event.preventDefault();
    if (ecosDraft && ecosUnclearFields.size > 0) {
      setMessage(t("vehicles.ecosDraft.unresolved", { count: ecosUnclearFields.size }));
      return;
    }

    setSaving(true);
    setMessage("");

    try {
      const images = pendingArticleImages.map((image, index) => ({
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

      if (ecosDraft && mode === "create") {
        await api.upsertVehicleExternalMapping(vehicle.id, ecosDraft.externalMapping);
        for (const cvValue of ecosDraft.cvValues) {
          await api.createVehicleCVValue(vehicle.id, cvValue);
        }
        vehicle = await api.vehicle(vehicle.id);
        setEcosDraft(null);
        setEcosUnclearFields(new Set());
      }

      setSelectedDetail(vehicle);
      setMode("view");
      load();
      if (mode === "create") {
        closeModal();
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
          <button type="button" role="menuitem" onClick={() => { setQuickMenuVehicleID(""); openDetail(vehicle, "uploads"); }}><Upload size={14} />Uploads</button>
          <button type="button" role="menuitem" onClick={() => { setQuickMenuVehicleID(""); openDetail(vehicle, "maintenance"); }}><Wrench size={14} />{t("vehicles.maintenance")}</button>
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
  const cvSummary = {
    values: selected?.cvValues?.length || 0,
    files: selected?.cvFiles?.length || 0,
    profiles: new Set([
      ...(selected?.cvValues || []).map((value) => value.decoderProfile).filter((profile): profile is string => Boolean(profile)),
      ...(selected?.cvFiles || []).map((file) => file.decoderProfile).filter((profile): profile is string => Boolean(profile))
    ]).size
  };
  const cvImportStats = {
    selected: cvImportPreview?.rows.filter((row) => row.selected && row.status !== "invalid").length || 0,
    new: cvImportPreview?.rows.filter((row) => row.status === "new").length || 0,
    changed: cvImportPreview?.rows.filter((row) => row.status === "changed").length || 0,
    same: cvImportPreview?.rows.filter((row) => row.status === "same").length || 0,
    invalid: cvImportPreview?.rows.filter((row) => row.status === "invalid").length || 0
  };
  const cvFilePreviewStats = {
    cvValues: cvFileUploadPreview?.previews.reduce((sum, preview) => sum + (preview.suggestedCvValues?.length || 0), 0) || 0,
    functions: cvFileUploadPreview?.previews.reduce((sum, preview) => sum + (preview.suggestedFunctions?.length || 0), 0) || 0
  };
  const storedDecoderProfiles = Array.from(new Set([
    ...(selected?.cvValues || []).map((value) => value.decoderProfile).filter((profile): profile is string => Boolean(profile)),
    ...(selected?.cvFiles || []).map((file) => file.decoderProfile).filter((profile): profile is string => Boolean(profile))
  ])).sort((a, b) => a.localeCompare(b, "de-DE"));
  const decoderProfileOptions = Array.from(new Set([...commonDecoderProfiles, ...storedDecoderProfiles]));

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
              <button type="button" className={activeTab === "cv" ? "active" : ""} onClick={() => setActiveTab("cv")}>
                CV
              </button>
              <button type="button" className={activeTab === "uploads" ? "active" : ""} onClick={() => setActiveTab("uploads")}>
                {t("vehicles.tab.uploads")}
              </button>
              <button type="button" className={activeTab === "maintenance" ? "active" : ""} onClick={() => setActiveTab("maintenance")}>
                {t("vehicles.tab.maintenance")}
              </button>
            </nav>

            <div className="modal-body">
              {activeTab === "model" && (
                <VehicleModelTab
                  form={form}
                  readonly={readonly}
                  articleSearchLoading={articleSearchLoading}
                  canRunArticleSearch={canRunArticleSearch}
                  articleIdentityFilled={articleIdentityFilled}
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
                  discardCVImportPreview={() => setCVImportPreview(null)}
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
                  discardCVFileUploadPreview={() => setCVFileUploadPreview(null)}
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
                  attachmentUploadMaintenanceID={attachmentUploadMaintenanceID}
                  attachmentUploadDescription={attachmentUploadDescription}
                  attachmentEdits={attachmentEdits}
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
                  onAttachmentUploadMaintenanceIDChange={setAttachmentUploadMaintenanceID}
                  onAttachmentUploadDescriptionChange={setAttachmentUploadDescription}
                  onUpdateAttachmentEdit={updateAttachmentEdit}
                  onSaveAttachment={saveAttachment}
                  onDeleteAttachment={deleteAttachment}
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
              )}            </div>

            <footer className="modal-actions">
              {message && <p className="form-message">{message}</p>}
              <button type="button" className="secondary-button" onClick={closeModal}>
                {t("vehicles.cancel")}
              </button>
              <button className="primary-button" disabled={saving}>
                {saving ? t("vehicles.saving") : t("vehicles.save")}
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
          onSelectEmptyFields={() => setArticleFieldSelection("empty")}
          onSelectAllFields={() => setArticleFieldSelection("all")}
          onClearFields={() => setArticleFieldSelection("none")}
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
    </>
  );
}