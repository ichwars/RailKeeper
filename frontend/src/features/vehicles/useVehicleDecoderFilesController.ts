import { useMemo, useRef, useState } from "react";

import { api, type Vehicle, type VehicleCVFile, type VehicleCVFilePreview, type VehicleCVValueInput } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import {
  buildCVImportPreview,
  type CVFileUploadPreview,
  type CVImportPreview
} from "./cvImport";
import { readCVImportFile } from "./cvFileImport";
import { isBlockedCVFile } from "./vehicleFiles";
import { maxVehicleUploadFiles, type VehicleUploadKind, type VehicleUploadMetadata } from "./vehicleUploadBatch";

type UseVehicleDecoderFilesControllerOptions = {
  selected: Vehicle | null;
  readonly: boolean;
  saving: boolean;
  setSaving: (saving: boolean) => void;
  onMessage: (message: string) => void;
  onStatus: (message: string) => void;
  onImportPreview: (preview: CVImportPreview) => void;
  onFunctionSuggestions: (previews: VehicleCVFilePreview[]) => { applied: number; blocked: string[] };
  prepareUpload: (kind: VehicleUploadKind, files: File[], metadata: VehicleUploadMetadata, confirmed?: boolean) => boolean;
  prepareDelete: (request: { kind: "decoder"; record: VehicleCVFile }) => void;
};

export function useVehicleDecoderFilesController({
  selected,
  readonly,
  saving,
  setSaving,
  onMessage,
  onStatus,
  onImportPreview,
  onFunctionSuggestions,
  prepareUpload,
  prepareDelete
}: UseVehicleDecoderFilesControllerOptions) {
  const { t } = useI18n();
  const [fileProfile, setFileProfile] = useState("");
  const [fileDescription, setFileDescription] = useState("");
  const [uploadPreview, setUploadPreview] = useState<CVFileUploadPreview | null>(null);
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const scope = useRef(selected?.id); scope.current = selected?.id;
  const version = useRef(0), previewOwner = useRef("");
  const canUsePreview = () => !readonly && !saving && selected && selected.id === previewOwner.current;

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
    version.current++;
    previewOwner.current = "";
    setFileProfile("");
    setFileDescription("");
    setUploadPreview(null);
  };

  const uploadFiles = async (files: FileList | null) => {
    if (!selected || readonly || saving || !files || files.length === 0) return;
    const uploadFiles = Array.from(files);
    if (uploadFiles.length > maxVehicleUploadFiles) {
      onMessage(t("vehicles.mediaUpload.limit", { count: maxVehicleUploadFiles })); return;
    }
    const blocked = uploadFiles.find(isBlockedCVFile);
    if (blocked) {
      onMessage(t("vehicles.mediaUpload.invalid.decoder", { name: blocked.name }));
      return;
    }
    setSaving(true);
    onMessage("");
    const request = ++version.current, vehicleId = selected.id;
    try {
      const previews = await Promise.all(uploadFiles.map(api.previewVehicleCVFile));
      if (version.current !== request || scope.current !== vehicleId) return;
      previewOwner.current = vehicleId;
      setUploadPreview({ files: uploadFiles, previews });
    } catch (error) {
      if (version.current === request && scope.current === vehicleId) onMessage(error instanceof Error ? error.message : String(error));
    } finally {
      if (version.current === request) {
        setSaving(false);
        if (fileInputRef.current) fileInputRef.current.value = "";
      }
    }
  };

  const applyFirstSuggestion = () => {
    if (!canUsePreview()) return;
    const suggestion = uploadPreview?.previews.find((preview) => preview.hasMetadata);
    if (!suggestion) return;
    if (suggestion.suggestedDecoderProfile) setFileProfile(suggestion.suggestedDecoderProfile);
    if (suggestion.suggestedDescription) setFileDescription(suggestion.suggestedDescription);
  };

  const previewValuesForImport = async () => {
    if (!canUsePreview() || !selected || !uploadPreview) return;
    const request = ++version.current, vehicleId = selected.id;
    setSaving(true);
    onMessage("");
    try {
      const batches = await Promise.all(uploadPreview.files.map(async (file, index) => {
        const preview = uploadPreview.previews[index];
        if (preview.cvImportError) throw new Error(preview.cvImportError);
        const rows: VehicleCVValueInput[] = /\.(txt|csv|json)$/i.test(file.name) ? await readCVImportFile(file) : preview.suggestedCvValues || [];
        return rows.map((value) => ({
          ...value,
          decoderProfile: value.decoderProfile || preview.suggestedDecoderProfile || preview.decoder || fileProfile || preview.projectName || "",
          sourceFileId: ""
        }));
      }));
      const values = batches.flat();
      if (version.current !== request || scope.current !== vehicleId) return;
      const preview = buildCVImportPreview("Decoder-Datei-Vorschau", values, selected.cvValues || []);
      if (!preview.rows.some((row) => row.status !== "invalid")) {
        onMessage("Keine gültigen CV-Werte in der Decoder-Vorschau gefunden.");
        return;
      }
      onImportPreview({ ...preview, vehicleId: selected.id });
      onStatus(t("vehicles.mediaUpload.cvPrepared", { count: values.length }));
    } catch (error) {
      if (version.current === request && scope.current === vehicleId) onMessage(error instanceof Error ? error.message : String(error));
    } finally {
      if (version.current === request) setSaving(false);
    }
  };

  const applyFunctionSuggestions = () => {
    if (!canUsePreview() || !uploadPreview) return;
    const result = onFunctionSuggestions(uploadPreview.previews);
    onStatus(t("vehicles.functions.fileDraft", { count: result.applied }) + (result.blocked.length
      ? " " + t("vehicles.functions.fileBlocked", { keys: result.blocked.join(", ") }) : ""));
  };

  const confirmUpload = () => {
    if (!canUsePreview() || !uploadPreview) return;
    if (prepareUpload("decoder", uploadPreview.files, { decoderProfile: fileProfile, description: fileDescription }, true)) {
      setUploadPreview(null);
    }
  };

  const remove = (file: VehicleCVFile) => {
    if (!selected || readonly || saving) return;
    prepareDelete({ kind: "decoder", record: file });
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
