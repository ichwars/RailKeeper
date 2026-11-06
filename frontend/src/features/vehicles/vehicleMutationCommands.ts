import type { Dispatch, FormEvent, SetStateAction } from "react";

import {
  api,
  ApiError,
  type CreateVehicleRequest,
  type Vehicle,
  type VehicleFunction,
  type VehicleFunctionChange,
  type VehicleImage,
  type VehicleSparePartInput
} from "../../shared/api";
import { adoptVehicleImportDraft } from "./vehicleImportAdoption";
import { importPendingVehicleImages, partitionVehicleImages } from "./vehicleImagePersistence";
import { clearVehicleCreateDraft } from "./vehicleCreateWizardState";
import { invalidFunctionChanges } from "./functionDraft";
import type { PendingArticleImage } from "./vehicleTransforms";
import type { ECoSVehicleDraftPayload, ModalMode, ModalTab } from "./vehicleViewModel";

type Translator = (key: string, values?: Record<string, string | number>) => string;

type VehicleMutationCommandsOptions = {
  draftOwner: string;
  editor: {
    form: CreateVehicleRequest;
    selected: Vehicle | null;
    mode: ModalMode;
    setSaving: (saving: boolean) => void;
    setSelectedDetail: (vehicle: Vehicle) => void;
    acknowledgeCoreSaved?: (vehicle: Vehicle) => void;
    hasRefreshConflicts?: boolean;
    reloadSelectedVehicle?: (vehicleId?: string) => Promise<void>;
    setMode: (mode: ModalMode) => void;
    setSaveAttempted: (attempted: boolean) => void;
    setActiveTab: (tab: ModalTab) => void;
    openModelSection: () => void;
    close: () => void;
  };
  validation: {
    missingRequiredLabels: string[];
  };
  media: {
    pendingImages: PendingArticleImage[];
    hasConflicts?: boolean;
    acknowledgeImportedImage?: (pending: PendingArticleImage, stored: VehicleImage) => void;
  };
  spareParts: {
    selectedInputs: () => VehicleSparePartInput[];
    clearSelected: () => void;
  };
  functions: {
    changes: () => VehicleFunctionChange[];
    acknowledgeSaved: (functions: VehicleFunction[] | undefined, vehicleID?: string) => void;
  };
  ecos: {
    draft: ECoSVehicleDraftPayload | null;
    unclearFieldCount: number;
    markSaved: (draft: ECoSVehicleDraftPayload, vehicleId: string) => void;
    clear: () => void;
    returnToSession: (draft: ECoSVehicleDraftPayload) => void;
  };
  deletion: {
    candidate: Vehicle | null;
    setCandidate: Dispatch<SetStateAction<Vehicle | null>>;
  };
  reloadVehicles: () => void;
  onMessage: (message: string) => void;
  t: Translator;
};

export function createVehicleMutationCommands({
  draftOwner,
  editor,
  validation,
  media,
  spareParts,
  functions,
  ecos,
  deletion,
  reloadVehicles,
  onMessage,
  t
}: VehicleMutationCommandsOptions) {
  const submit = async (event: FormEvent) => {
    event.preventDefault();
    editor.setSaveAttempted(true);

    if (media.hasConflicts) {
      editor.setActiveTab("uploads");
      onMessage(t("vehicles.mediaConflict.resolve"));
      return;
    }
    if (editor.mode === "edit" && editor.selected && (!editor.selected.revision || !editor.selected.imagesRevision)) {
      onMessage(t("vehicles.mediaConflict.missingRevision"));
      return;
    }

    if (editor.hasRefreshConflicts) {
      onMessage(t("vehicles.saveState.resolveConflicts"));
      return;
    }

    if (validation.missingRequiredLabels.length > 0) {
      editor.setActiveTab("model");
      editor.openModelSection();
      onMessage(t("vehicles.requiredMissing", { fields: validation.missingRequiredLabels.join(", ") }));
      return;
    }
    if (ecos.draft && ecos.unclearFieldCount > 0) {
      onMessage(t("vehicles.ecosDraft.unresolved", { count: ecos.unclearFieldCount }));
      return;
    }

    const functionChanges = functions.changes();
    const invalidKeys = invalidFunctionChanges(functionChanges);
    if (invalidKeys.length) {
      editor.setActiveTab("control");
      onMessage(t("vehicles.functions.invalidFields", { keys: invalidKeys.join(", ") }));
      return;
    }

    editor.setSaving(true);
    onMessage("");
    try {
      const sparePartsToImport = spareParts.selectedInputs();
      const { images, remoteImages } = partitionVehicleImages(media.pendingImages);
      const payload = { ...editor.form, images,
        ...(editor.mode === "edit" && editor.selected ? {
          expectedRevision: editor.selected.revision, expectedImagesRevision: editor.selected.imagesRevision
        } : {}), ...(functionChanges.length ? { functionChanges } : {}) };
      let vehicle = editor.mode === "edit" && editor.selected
        ? await api.updateVehicle(editor.selected.id, payload)
        : await api.createVehicle(payload);
      if (editor.mode === "create") clearVehicleCreateDraft(draftOwner);
      functions.acknowledgeSaved(vehicle.functions, vehicle.id);
      // The core transaction succeeded. Retrying an ancillary image/import failure must not create a duplicate.
      (editor.acknowledgeCoreSaved || editor.setSelectedDetail)(vehicle);
      editor.setMode("edit");

      vehicle = await importPendingVehicleImages(vehicle, remoteImages, images.length, media.acknowledgeImportedImage);

      for (const part of sparePartsToImport) {
        await api.createVehicleSparePart(vehicle.id, part);
      }
      if (sparePartsToImport.length > 0) {
        spareParts.clearSelected();
        vehicle = await api.vehicle(vehicle.id);
      }

      if (ecos.draft && (editor.mode === "create" || editor.mode === "edit")) {
        const result = await adoptVehicleImportDraft(vehicle, ecos.draft, {
          markSaved: ecos.markSaved,
          clear: ecos.clear,
          returnToSession: ecos.returnToSession,
          reloadVehicles,
          closeEditor: editor.close
        });
        vehicle = result.vehicle;
        if (result.returnedToSource) return;
      }

      vehicle = await api.vehicle(vehicle.id);
      editor.setSelectedDetail(vehicle);
      editor.setMode("edit");
      editor.setSaveAttempted(false);
      reloadVehicles();
      if (editor.mode === "create") onMessage(t("vehicles.createdContinue"));
      else if (sparePartsToImport.length > 0) {
        onMessage(t("vehicles.spareParts.importedCount", { count: sparePartsToImport.length }));
      }
    } catch (error) {
      if (error instanceof ApiError && ["vehicle_conflict", "vehicle_image_conflict"].includes(error.code)) {
        if (editor.selected) await editor.reloadSelectedVehicle?.(editor.selected.id);
        onMessage(t("vehicles.mediaConflict.changed"));
        return;
      }
      onMessage(error instanceof ApiError && error.code === "vehicle_function_conflict"
        ? t("vehicles.functions.conflict")
        : error instanceof ApiError && error.code === "vehicle_function_validation"
        ? t("vehicles.functions.invalidFields", { keys: functionChanges.map(change => change.functionKey).join(", ") })
        : error instanceof ApiError && error.code === "external_mapping_conflict"
        ? t("vehicles.ecosDraft.mappingConflict") : error instanceof Error ? error.message : String(error));
    } finally {
      editor.setSaving(false);
    }
  };

  const confirmDelete = () => {
    if (!deletion.candidate) return;
    api.deleteVehicle(deletion.candidate.id)
      .then(() => {
        if (editor.selected?.id === deletion.candidate?.id) editor.close();
        deletion.setCandidate(null);
        reloadVehicles();
      })
      .catch((error: Error) => onMessage(error.message));
  };

  return { submit, confirmDelete };
}
