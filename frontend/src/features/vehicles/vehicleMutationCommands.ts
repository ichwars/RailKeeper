import type { Dispatch, FormEvent, SetStateAction } from "react";

import {
  api,
  type CreateVehicleRequest,
  type Vehicle,
  type VehicleSparePartInput
} from "../../shared/api";
import { cvValueKey } from "./cvImport";
import { clearVehicleCreateDraft } from "./vehicleCreateWizardState";
import type { FunctionEditState, PendingArticleImage } from "./vehicleTransforms";
import type { ECoSVehicleDraftPayload, ModalMode, ModalTab } from "./vehicleViewModel";

type Translator = (key: string, values?: Record<string, string | number>) => string;

type VehicleMutationCommandsOptions = {
  editor: {
    form: CreateVehicleRequest;
    selected: Vehicle | null;
    mode: ModalMode;
    setSaving: (saving: boolean) => void;
    setSelectedDetail: (vehicle: Vehicle) => void;
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
  };
  spareParts: {
    selectedInputs: () => VehicleSparePartInput[];
    clearSelected: () => void;
  };
  functions: {
    configuredKeys: string[];
    edit: (functionKey: string) => FunctionEditState[string];
  };
  ecos: {
    draft: ECoSVehicleDraftPayload | null;
    unclearFieldCount: number;
    markSaved: (draft: ECoSVehicleDraftPayload, vehicleId: string) => void;
    clear: () => void;
    returnToSession: () => void;
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
  const isRemotePendingImage = (image: PendingArticleImage) => !image.persisted && /^https?:\/\//i.test(image.url);

  const submit = async (event: FormEvent) => {
    event.preventDefault();
    editor.setSaveAttempted(true);

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

    editor.setSaving(true);
    onMessage("");
    try {
      const sparePartsToImport = spareParts.selectedInputs();
      const remoteImages = media.pendingImages.filter(isRemotePendingImage);
      const images = media.pendingImages.filter((image) => !isRemotePendingImage(image)).map((image, index) => ({
        id: image.persisted ? image.id : undefined,
        url: image.url,
        title: image.title,
        sourceUrl: image.source,
        maintenanceId: image.maintenanceId || "",
        isPrimary: Boolean(image.isPrimary),
        sortOrder: index
      }));
      const payload = { ...editor.form, images };
      let vehicle = editor.mode === "edit" && editor.selected
        ? await api.updateVehicle(editor.selected.id, payload)
        : await api.createVehicle(payload);
      if (editor.mode === "create") clearVehicleCreateDraft();

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
      if (remoteImages.length > 0) vehicle = await api.vehicle(vehicle.id);

      for (const part of sparePartsToImport) {
        await api.createVehicleSparePart(vehicle.id, part);
      }
      if (sparePartsToImport.length > 0) {
        spareParts.clearSelected();
        vehicle = await api.vehicle(vehicle.id);
      }

      if (ecos.draft && (editor.mode === "create" || editor.mode === "edit")) {
        await api.upsertVehicleExternalMapping(vehicle.id, ecos.draft.externalMapping);
        const detailBeforeECoSValues = await api.vehicle(vehicle.id);
        for (const cvValue of ecos.draft.cvValues) {
          const existing = (detailBeforeECoSValues.cvValues || []).find((entry) => (
            cvValueKey(entry) === cvValueKey(cvValue)
          ));
          if (existing) await api.updateVehicleCVValue(vehicle.id, existing.id, cvValue);
          else await api.createVehicleCVValue(vehicle.id, cvValue);
        }
        for (const functionKey of functions.configuredKeys) {
          const edit = functions.edit(functionKey);
          if (!edit.name?.trim() && !edit.symbolKey && !edit.notes?.trim()) continue;
          await api.updateVehicleFunction(vehicle.id, functionKey, {
            name: edit.name || "",
            symbolKey: edit.symbolKey || "",
            functionType: edit.functionType || "standard",
            mode: edit.mode || "dauer",
            directionDependent: Boolean(edit.directionDependent),
            notes: edit.notes || ""
          });
        }
        ecos.markSaved(ecos.draft, vehicle.id);
        vehicle = await api.vehicle(vehicle.id);
        const returnToEcos = Boolean(ecos.draft.returnToEcos);
        ecos.clear();
        if (returnToEcos) {
          reloadVehicles();
          editor.close();
          ecos.returnToSession();
          return;
        }
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
      onMessage(error instanceof Error ? error.message : String(error));
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
