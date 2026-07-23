import { useMemo, useRef, useState } from "react";

import { api, type Vehicle, type VehicleCVFile } from "../../shared/api";
import {
  buildCVImportPreview,
  isValidFunctionMapping,
  type CVFileUploadPreview,
  type CVImportPreview
} from "./cvImport";
import { isBlockedCVFile } from "./vehicleFiles";

type UseVehicleDecoderFilesControllerOptions = {
  selected: Vehicle | null;
  setSaving: (saving: boolean) => void;
  onMessage: (message: string) => void;
  onImportPreview: (preview: CVImportPreview) => void;
  refreshSelectedVehicle: (vehicleId?: string) => Promise<void>;
};

export function useVehicleDecoderFilesController({
  selected,
  setSaving,
  onMessage,
  onImportPreview,
  refreshSelectedVehicle
}: UseVehicleDecoderFilesControllerOptions) {
  const [fileProfile, setFileProfile] = useState("");
  const [fileDescription, setFileDescription] = useState("");
  const [uploadPreview, setUploadPreview] = useState<CVFileUploadPreview | null>(null);
  const fileInputRef = useRef<HTMLInputElement | null>(null);

  const previewStats = useMemo(() => ({
    cvValues: uploadPreview?.previews.reduce(
      (sum, preview) => sum + (preview.suggestedCvValues?.length || 0),
      0
    ) || 0,
    functions: uploadPreview?.previews.reduce(
      (sum, preview) => sum + (preview.suggestedFunctions?.length || 0),
      0
    ) || 0
  }), [uploadPreview]);

  const reset = () => {
    setFileProfile("");
    setFileDescription("");
    setUploadPreview(null);
  };

  const uploadFiles = async (files: FileList | null) => {
    if (!selected || !files || files.length === 0) return;
    const uploadFiles = Array.from(files);
    const blocked = uploadFiles.find(isBlockedCVFile);
    if (blocked) {
      onMessage(
        `${blocked.name} ist als CV-Datei nicht erlaubt. ` +
        "Erlaubt sind JSON, CSV, TXT, XML, Z21, ESU, ESUX, LokProgrammer und ZIP."
      );
      return;
    }
    setSaving(true);
    onMessage("");
    try {
      setUploadPreview({ files: uploadFiles, previews: await Promise.all(uploadFiles.map(api.previewVehicleCVFile)) });
    } catch (error) {
      onMessage(error instanceof Error ? error.message : String(error));
    } finally {
      setSaving(false);
      if (fileInputRef.current) fileInputRef.current.value = "";
    }
  };

  const applyFirstSuggestion = () => {
    const suggestion = uploadPreview?.previews.find((preview) => preview.hasMetadata);
    if (!suggestion) return;
    if (suggestion.suggestedDecoderProfile) setFileProfile(suggestion.suggestedDecoderProfile);
    if (suggestion.suggestedDescription) setFileDescription(suggestion.suggestedDescription);
  };

  const previewValuesForImport = () => {
    if (!selected || !uploadPreview) return;
    const values = uploadPreview.previews.flatMap((preview) =>
      (preview.suggestedCvValues || []).map((value) => ({
        cvNumber: value.cvNumber,
        value: value.value,
        description: value.description || "",
        category: value.category || "",
        protocol: value.protocol || "",
        decoderProfile: preview.suggestedDecoderProfile || fileProfile || preview.decoder || preview.projectName || "",
        sourceFileId: ""
      }))
    );
    const preview = buildCVImportPreview("Decoder-Datei-Vorschau", values, selected.cvValues || []);
    if (!preview.rows.some((row) => row.status !== "invalid")) {
      onMessage("Keine gültigen CV-Werte in der Decoder-Vorschau gefunden.");
      return;
    }
    onImportPreview(preview);
    onMessage(`${values.length} erkannte CV-Werte für die Prüfung vorbereitet.`);
  };

  const applyFunctionSuggestions = () => {
    if (!selected || !uploadPreview) return;
    const mappings = uploadPreview.previews.flatMap((preview) =>
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
    const valid = Array.from(new Map(
      mappings.filter(isValidFunctionMapping).map((mapping) => [mapping.functionKey, mapping])
    ).values());
    if (valid.length === 0) {
      onMessage("Keine gültigen Funktionstasten in der Decoder-Vorschau gefunden.");
      return;
    }
    setSaving(true);
    onMessage("");
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
      .then(() => onMessage(
        `${valid.length} Funktionstaste${valid.length === 1 ? "" : "n"} aus der Decoder-Vorschau übernommen.`
      ))
      .catch((error: Error) => onMessage(error.message))
      .finally(() => setSaving(false));
  };

  const confirmUpload = () => {
    if (!selected || !uploadPreview) return;
    const uploadFiles = uploadPreview.files;
    setSaving(true);
    onMessage("");
    (async () => {
      for (const file of uploadFiles) {
        await api.uploadVehicleCVFile(selected.id, file, fileProfile, fileDescription);
      }
    })()
      .then(() => refreshSelectedVehicle(selected.id))
      .then(() => {
        setUploadPreview(null);
        onMessage(`${uploadFiles.length} CV-Datei${uploadFiles.length === 1 ? "" : "en"} gespeichert.`);
      })
      .catch((error: Error) => onMessage(error.message))
      .finally(() => setSaving(false));
  };

  const remove = (file: VehicleCVFile) => {
    if (!selected) return;
    setSaving(true);
    onMessage("");
    api.deleteVehicleCVFile(selected.id, file.id)
      .then(() => refreshSelectedVehicle(selected.id))
      .catch((error: Error) => onMessage(error.message))
      .finally(() => setSaving(false));
  };

  return {
    state: { fileProfile, fileDescription, uploadPreview, previewStats },
    refs: { fileInputRef },
    commands: {
      reset,
      uploadFiles,
      updateFileProfile: setFileProfile,
      updateFileDescription: setFileDescription,
      applyFirstSuggestion,
      previewValuesForImport,
      applyFunctionSuggestions,
      confirmUpload,
      discardUploadPreview: () => setUploadPreview(null),
      remove
    }
  };
}
