import { useEffect, useMemo, useState } from "react";
import { Vehicle } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import type { CVWorkingGroupOptions } from "./cvWorkingGroup";
import { useDecoderConfigurationController } from "./useDecoderConfigurationController";
import { useVehicleCollectionController } from "./useVehicleCollectionController";
import { useVehicleSetController } from "./useVehicleSetController";

import { maintenanceIsDue } from "./vehicleMaintenance";

import { useArticleSearchController } from "./useArticleSearchController";
import { useVehicleCVController } from "./useVehicleCVController";
import { useVehicleColumnPreferences } from "./useVehicleColumnPreferences";
import { useVehicleDecoderFilesController } from "./useVehicleDecoderFilesController";
import { useVehicleDocumentsController } from "./useVehicleDocumentsController";
import { useVehicleECoSDraftController } from "./useVehicleECoSDraftController";
import { useVehicleEditorController } from "./useVehicleEditorController";
import { useVehicleFunctionsController } from "./useVehicleFunctionsController";
import { useVehicleInventoryController } from "./useVehicleInventoryController";
import { useVehicleMaintenanceController } from "./useVehicleMaintenanceController";
import { useVehicleMediaController } from "./useVehicleMediaController";
import { useVehicleMediaDeletion } from "./useVehicleMediaDeletion";
import { useVehicleOutputController } from "./useVehicleOutputController";
import { useVehicleSparePartsController } from "./useVehicleSparePartsController";
import { useVehicleUploads } from "./useVehicleUploads";
import { createVehicleFilterDefinitions, createVehicleInventoryRenderers } from "./vehicleInventoryRenderers";
import { createVehicleMutationCommands } from "./vehicleMutationCommands";
import {
  gattungenForCategory,
  hasArticleSearchCriteria,
  hasQrPayloadData,
  missingVehicleModelFieldLabels
} from "./vehicleViewModel";

export function useVehicleWorkspace({ username, roles = ["Editor"] }: { username: string; roles?: string[] }) {
  const { language, t } = useI18n();
  const collection =
    useVehicleCollectionController();
  const [deleteCandidate, setDeleteCandidate] = useState<Vehicle | null>(null);
  const [quickMenuVehicleID, setQuickMenuVehicleID] = useState("");
  const [archiveOwner, setArchiveOwner] = useState<string | null>(null);
  const canReadDecoderArchive = roles.some(role => ["Admin", "Editor", "Viewer"].includes(role));
  const canEditVehicleSets = roles.includes("Admin") || roles.includes("Editor");

  const editor = useVehicleEditorController({
    options: collection.options,
    onCollectionChanged: collection.load,
    onMessage: collection.setMessage,
    onFormChange: (nextForm) => ecos.commands.syncUnclearFields(nextForm),
    onReset: (reason) => {
      ecos.commands.clear();
      media.commands.reset(reason === "close");
      mediaUploads.commands.reset();
      mediaDeletion.commands.reset();
      functions.commands.reset();
      maintenance.commands.resetForm();
      spareParts.commands.resetForm();
      spareParts.commands.resetSearch();
      documents.commands.reset();
      cvs.commands.reset();
      decoderFiles.commands.reset();
    },
    onDetailLoaded: (detail) => {
      ecos.commands.clear();
      media.commands.loadDetail(detail);
      functions.commands.loadDetail(detail.functions, detail.id);
      maintenance.commands.resetForm();
      spareParts.commands.resetDetail();
      cvs.commands.reset();
      decoderFiles.commands.reset();
    },
    onDetailRefreshed: (detail, previous) => {
      media.commands.refreshDetail(detail, previous);
      functions.commands.loadDetail(detail.functions, detail.id);
    },
    onCoreSaved: (detail, previous, phase) => {
      media.commands.acknowledgeCoreSaved(detail, previous, phase);
      functions.commands.loadDetail(detail.functions, detail.id);
    }
  });

  const decoderConfigurations = useDecoderConfigurationController({
    ownerKey: `${username}:${editor.state.sessionKey}`, selected: editor.state.selected,
    active: editor.state.modalOpen && editor.state.activeTab === "cv" && canReadDecoderArchive,
    readonly: editor.state.readonly || !canEditVehicleSets, busy: editor.state.saving, setBusy: editor.setters.setSaving,
    captureScope: () => editor.commands.captureScope(editor.state.sessionKey)
  });
  const cvGroupsCovered = useMemo(() => {
    const profiles = new Set(decoderConfigurations.context?.workspaces.map(group => group.profile));
    return [...(editor.state.selected?.cvValues || []), ...(editor.state.selected?.cvFiles || [])].every(value => profiles.has(value.decoderProfile || ""));
  }, [decoderConfigurations.context?.workspaces, editor.state.selected?.cvValues, editor.state.selected?.cvFiles]);
  const cvWorkingGroup: CVWorkingGroupOptions = {
    ownerKey: `${username}:${editor.state.sessionKey}`, context: decoderConfigurations.context,
    groups: decoderConfigurations.context?.workspaces, target: decoderConfigurations.activeTarget,
    ready: Boolean(decoderConfigurations.context && !decoderConfigurations.loading && decoderConfigurations.targetCurrent && cvGroupsCovered),
    reload: () => decoderConfigurations.load(false, true)
  };

  const { openCreate, openDetail } = editor.commands;
  const { setActiveTab } = editor.setters;
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
    const vehicle = collection.vehicles.find((candidate) => candidate.id === vehicleID);
    if (!vehicle) return;
    const requestedTab = params.get("tab");
    params.delete("id");
    params.delete("tab");
    window.history.replaceState(null, "", `${window.location.pathname}${params.size ? `?${params}` : ""}`);
    openDetail(vehicle);
    if (requestedTab === "maintenance") setActiveTab("maintenance");
  }, [canEditVehicleSets, openCreate, openDetail, setActiveTab, collection.vehicles]);

  const functions = useVehicleFunctionsController({
    selected: editor.state.selected,
    form: editor.state.form,
    setSaving: editor.setters.setSaving,
    onMessage: collection.setMessage,
    refreshSelectedVehicle: (vehicleId) => editor.commands.refreshSelectedVehicle(vehicleId)
  });

  const media = useVehicleMediaController({
    selected: editor.state.selected,
    readonly: editor.state.readonly,
    saving: editor.state.saving,
    setSaving: editor.setters.setSaving,
    onMessage: collection.setMessage,
    refreshSelectedVehicle: (vehicleId) => editor.commands.refreshSelectedVehicle(vehicleId, "media"),
    reloadSelectedVehicle: (vehicleId) => editor.commands.refreshSelectedVehicle(vehicleId),
    prepareUpload: (...args) => mediaUploads.commands.prepare(...args),
    prepareDelete: request => mediaDeletion.commands.prepare(request)
  });

  const refreshSavedMedia = (vehicleId: string, kind: "image" | "attachment" | "decoder") =>
    editor.commands.refreshSelectedVehicle(vehicleId, kind === "decoder" ? "decoderFiles" : "media", true);
  const mediaUploads = useVehicleUploads({
    selected: editor.state.selected, readonly: editor.state.readonly, saving: editor.state.saving, workingGroup: cvWorkingGroup, setSaving: editor.setters.setSaving, onMessage: collection.setMessage, onStatus: collection.setStatusMessage,
    onImageStored: media.commands.acknowledgeImage, refreshSaved: refreshSavedMedia,
    reload: vehicleId => editor.commands.refreshSelectedVehicle(vehicleId, undefined, true)
  });
  const mediaDeletion = useVehicleMediaDeletion({
    selected: editor.state.selected, readonly: editor.state.readonly, saving: editor.state.saving, setSaving: editor.setters.setSaving, onMessage: collection.setMessage, onStatus: collection.setStatusMessage,
    onImageDeleted: media.commands.removeDeletedImage, refreshSaved: refreshSavedMedia,
    reload: vehicleId => editor.commands.refreshSelectedVehicle(vehicleId, undefined, true)
  });

  const maintenance = useVehicleMaintenanceController({
    selected: editor.state.selected,
    setSaving: editor.setters.setSaving,
    onMessage: collection.setMessage,
    refreshSelectedVehicle: (vehicleId) => editor.commands.refreshSelectedVehicle(vehicleId, "maintenance")
  });

  const spareParts = useVehicleSparePartsController({
    selected: editor.state.selected,
    active: editor.state.activeTab === "spareParts",
    attachmentEdits: media.state.attachmentEdits,
    setSaving: editor.setters.setSaving,
    onMessage: collection.setMessage,
    onOpenSpareParts: () => editor.setters.setActiveTab("spareParts"),
    refreshSelectedVehicle: (vehicleId) => editor.commands.refreshSelectedVehicle(vehicleId, "spareParts"),
    t
  });

  const cvs = useVehicleCVController({
    selected: editor.state.selected,
    workingGroup: cvWorkingGroup,
    readonly: editor.state.readonly,
    saving: editor.state.saving,
    decoderNumber: editor.state.form.digitalDecoderNumber || editor.state.form.dtDecoderNumber || "",
    setSaving: editor.setters.setSaving,
    onMessage: collection.setMessage,
    refreshSelectedVehicle: (vehicleId) => editor.commands.refreshSelectedVehicle(vehicleId, "cv"),
    refreshImportedValues: (vehicleId) => editor.commands.refreshSelectedVehicle(vehicleId, "cv", true),
    reloadSelectedVehicle: (vehicleId) => editor.commands.refreshSelectedVehicle(vehicleId)
  });

  const decoderFiles = useVehicleDecoderFilesController({
    selected: editor.state.selected,
    workingGroup: cvWorkingGroup,
    readonly: editor.state.readonly,
    saving: editor.state.saving,
    setSaving: editor.setters.setSaving,
    onMessage: collection.setMessage,
    onStatus: collection.setStatusMessage,
    onImportPreview: cvs.commands.setImportPreview,
    onFunctionSuggestions: (previews) => {
      const result = functions.commands.supplementFileSuggestions(previews);
      if (result.applied > 0) editor.setters.setActiveTab("control");
      return result;
    },
    prepareUpload: mediaUploads.commands.prepare,
    prepareDelete: mediaDeletion.commands.prepare
  });

  const documents = useVehicleDocumentsController({
    selected: editor.state.selected,
    setSaving: editor.setters.setSaving,
    onMessage: collection.setMessage,
    refreshSelectedVehicle: (vehicleId) => editor.commands.refreshSelectedVehicle(vehicleId, "documents"),
    t
  });

  const articleSearch = useArticleSearchController({
    targetKey: editor.state.modalOpen ? `${editor.state.mode}:${editor.state.selected?.id || "new"}` : "closed",
    form: editor.state.form,
    pendingImageCount: media.state.pendingImages.length,
    replaceForm: editor.setters.setForm,
    updateForm: editor.commands.update,
    onMessage: collection.setMessage,
    t,
    addImages: media.commands.addImages
  });

  const columnPreferences = useVehicleColumnPreferences(collection.setMessage);

  const inventory = useVehicleInventoryController(collection.vehicles, columnPreferences.columns);
  const output = useVehicleOutputController({
    username,
    form: editor.state.form,
    selected: editor.state.selected,
    query: collection.query,
    sort: inventory.sort,
    sortedVehicles: inventory.sortedVehicles,
    selectedVisibleVehicles: inventory.selectedVisibleVehicles,
    setVehicles: collection.setVehicles,
    setSelectedDetail: editor.commands.applyIndependentDetail,
    onMessage: collection.setMessage,
    t
  });

  const ecos = useVehicleECoSDraftController({
    onOpenCreate: (prepared) => {
      sets.setCreatePrefill(null);
      editor.setters.setSelected(null);
      editor.setters.setMode("create");
      editor.setters.setForm(prepared.form);
      media.setters.setPendingImages(prepared.mergeImages([]));
      media.setters.setAttachmentEdits({});
      media.setters.setImageUploadMaintenanceId("");
      media.setters.setAttachmentUploadCategory("");
      media.setters.setAttachmentUploadDescription("");
      media.setters.setAttachmentDragActive(false);
      mediaUploads.commands.reset();
      mediaDeletion.commands.reset();
      editor.setters.setSaveAttempted(false);
      functions.setters.setEdits(Object.fromEntries(Object.entries(prepared.functionEdits)
        .map(([key, edit]) => [key, { ...edit, dirty: true, modeAssigned: true }])));
    },
    onOpenUpdate: (detail, prepared) => {
      editor.commands.setSelectedDetail(detail);
      media.setters.setPendingImages((current) => prepared.mergeImages(current));
      editor.setters.setMode("edit");
      editor.setters.setForm(prepared.form);
      functions.commands.mergeDetail(detail.functions, prepared.functionEdits);
    },
    onFinishOpen: (draft) => {
      maintenance.commands.resetForm();
      spareParts.commands.resetForm();
      spareParts.commands.resetSearch();
      documents.commands.reset();
      cvs.commands.reset();
      decoderFiles.commands.reset();
      functions.setters.setShowConfiguredOnly(true);
      editor.setters.setActiveTab("model");
      editor.setters.setOpenSections({ model: true, details: false, vehicle: false });
      editor.setters.setModalOpen(true);
      collection.setMessage(draft.mode === "update" ? t("vehicles.ecosDraft.loadedUpdate") : t("vehicles.ecosDraft.loaded"));
    },
    onMessage: collection.setMessage,
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

  const canRunArticleSearch = hasArticleSearchCriteria(editor.state.form);
  const canGenerateQr = hasQrPayloadData(editor.state.selected, editor.state.form);
  const missingRequiredModelFields = missingVehicleModelFieldLabels(editor.state.form, t);
  const showRequiredErrors = editor.state.saveAttempted || Boolean(ecos.state.draft);
  const filters = createVehicleFilterDefinitions({
    vehicleCount: collection.vehicles.length,
    counts: inventory.inventoryFilterCounts,
    t
  });
  const filteredGattungen = gattungenForCategory(collection.options, editor.state.form.category);

  const mutations = createVehicleMutationCommands({
    draftOwner: username,
    editor: {
      captureScope: editor.commands.captureScope,
      form: editor.state.form,
      selected: editor.state.selected,
      mode: editor.state.mode,
      setSaving: editor.setters.setSaving,
      setSelectedDetail: (detail) => editor.commands.acknowledgeSavedVehicle(detail, editor.state.form, "complete"),
      acknowledgeCoreSaved: (detail) => editor.commands.acknowledgeSavedVehicle(detail, editor.state.form, "core"),
      hasRefreshConflicts: editor.state.refreshConflicts.length > 0,
      reloadSelectedVehicle: (vehicleId) => editor.commands.refreshSelectedVehicle(vehicleId),
      setMode: editor.setters.setMode,
      setSaveAttempted: editor.setters.setSaveAttempted,
      setActiveTab: editor.setters.setActiveTab,
      openModelSection: () => editor.setters.setOpenSections((current) => ({ ...current, model: true })),
      close: editor.commands.closeModal
    },
    validation: {
      missingRequiredLabels: missingRequiredModelFields
    },
    media: {
      pendingImages: media.state.pendingImages, acknowledgeImportedImage: media.commands.acknowledgeImportedImage,
      hasConflicts: media.state.mediaConflicts.some(conflict => conflict.kind === "image")
    },
    spareParts: {
      selectedInputs: spareParts.commands.selectedInputs,
      clearSelected: spareParts.commands.clearSelectedFound
    },
    functions: {
      changes: functions.commands.changes,
      acknowledgeSaved: functions.commands.acknowledgeSaved
    },
    ecos: {
      draft: ecos.state.draft,
      unclearFieldCount: ecos.state.unclearFields.size,
      markSaved: ecos.commands.markImportSessionSaved,
      clear: ecos.commands.clear,
      returnToSession: ecos.commands.returnToImportSession
    },
    deletion: {
      candidate: deleteCandidate,
      setCandidate: setDeleteCandidate
    },
    reloadVehicles: collection.load,
    onMessage: collection.setMessage,
    t
  });
  const sets = useVehicleSetController({
    captureScope: editor.commands.captureScope,
    username, form: editor.state.form, pendingArticleImages: media.state.pendingImages, setSaving: editor.setters.setSaving, setSelectedDetail: editor.commands.setSelectedDetail, setMode: editor.setters.setMode,
    setActiveTab: editor.setters.setActiveTab, closeModal: editor.commands.closeModal, openCreate: editor.commands.openCreate, openDetail: editor.commands.openDetail, vehicles: collection.vehicles, load: collection.load, onMessage: collection.setMessage, t
  });
  const renderers = createVehicleInventoryRenderers({
    sort: inventory.sort,
    quickMenuVehicleID,
    setQuickMenuVehicleID,
    toggleSort: inventory.toggleSort,
    openDetail: editor.commands.openDetail,
    openEdit: editor.commands.openEdit,
    openQr: output.qr.commands.openForVehicle,
    printVehicle: output.report.commands.printVehicle,
    setDeleteCandidate,
    t
  });
  const maintenanceEntries = editor.state.selected?.maintenance || [];
  const maintenanceSummary = {
    due: maintenanceEntries.filter(maintenanceIsDue).length,
    planned: maintenanceEntries.filter((entry) => entry.status !== "erledigt").length,
    done: maintenanceEntries.filter((entry) => entry.status === "erledigt").length
  };

  return {
    username,
    roles,
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
    mediaUploads,
    mediaDeletion,
    media,
    mutations,
    ecos,
    articleSearch,
    canRunArticleSearch,
    showRequiredErrors,
    filteredGattungen,
    canGenerateQr,
    functions,
    decoderConfigurations,
    cvWorkingGroup,
    cvs,
    decoderFiles,
    maintenanceEntries,
    documents,
    spareParts,
    maintenance,
    maintenanceSummary,
    t,
    deleteCandidate,
  };
}

export type VehicleWorkspace = ReturnType<typeof useVehicleWorkspace>;
