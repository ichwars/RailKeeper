import { useMemo, useRef, useState } from "react";

import {
  api,
  type Vehicle,
  type VehicleCVValue,
  type VehicleCVValueInput
} from "../../shared/api";
import {
  buildCVImportPreview,
  commonDecoderProfiles,
  cvValueKey,
  cvValuesFromImport,
  isValidCVValueInput,
  type CVImportPreview
} from "./cvImport";
import { emptyCVForm } from "./vehicleViewModel";

type UseVehicleCVControllerOptions = {
  selected: Vehicle | null;
  decoderNumber: string;
  setSaving: (saving: boolean) => void;
  onMessage: (message: string) => void;
  refreshSelectedVehicle: (vehicleId?: string) => Promise<void>;
};

export function useVehicleCVController({
  selected,
  decoderNumber,
  setSaving,
  onMessage,
  refreshSelectedVehicle
}: UseVehicleCVControllerOptions) {
  const [form, setForm] = useState<VehicleCVValueInput>(emptyCVForm);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [importPreview, setImportPreview] = useState<CVImportPreview | null>(null);
  const importInputRef = useRef<HTMLInputElement | null>(null);

  const storedDecoderProfiles = useMemo(() => Array.from(new Set([
    ...(selected?.cvValues || []).map((value) => value.decoderProfile)
      .filter((profile): profile is string => Boolean(profile)),
    ...(selected?.cvFiles || []).map((file) => file.decoderProfile)
      .filter((profile): profile is string => Boolean(profile))
  ])).sort((left, right) => left.localeCompare(right, "de-DE")), [selected]);

  const summary = useMemo(() => ({
    values: selected?.cvValues?.length || 0,
    files: selected?.cvFiles?.length || 0,
    profiles: storedDecoderProfiles.length
  }), [selected, storedDecoderProfiles]);

  const importStats = useMemo(() => ({
    selected: importPreview?.rows.filter((row) => row.selected && row.status !== "invalid").length || 0,
    new: importPreview?.rows.filter((row) => row.status === "new").length || 0,
    changed: importPreview?.rows.filter((row) => row.status === "changed").length || 0,
    same: importPreview?.rows.filter((row) => row.status === "same").length || 0,
    invalid: importPreview?.rows.filter((row) => row.status === "invalid").length || 0
  }), [importPreview]);

  const decoderProfileOptions = useMemo(
    () => Array.from(new Set([...commonDecoderProfiles, ...storedDecoderProfiles])),
    [storedDecoderProfiles]
  );

  const updateForm = (patch: Partial<VehicleCVValueInput>) => {
    setForm((current) => ({ ...current, ...patch }));
  };

  const resetForm = () => {
    setForm(emptyCVForm);
    setEditingId(null);
  };

  const reset = () => {
    resetForm();
    setImportPreview(null);
  };

  const edit = (value: VehicleCVValue) => {
    setForm({
      cvNumber: value.cvNumber,
      value: value.value,
      description: value.description || "",
      category: value.category || "",
      protocol: value.protocol || "",
      decoderProfile: value.decoderProfile || "",
      sourceFileId: value.sourceFileId || ""
    });
    setEditingId(value.id);
  };

  const save = () => {
    if (!selected) return;
    const payload = { ...form, cvNumber: Number(form.cvNumber), value: Number(form.value) };
    if (!isValidCVValueInput(payload)) {
      onMessage("CV-Nummer muss 1-1024 und Wert 0-255 sein.");
      return;
    }
    setSaving(true);
    onMessage("");
    const existing = editingId
      ? undefined
      : (selected.cvValues || []).find((entry) => cvValueKey(entry) === cvValueKey(payload));
    const action = editingId
      ? api.updateVehicleCVValue(selected.id, editingId, payload)
      : existing
        ? api.updateVehicleCVValue(selected.id, existing.id, payload)
        : api.createVehicleCVValue(selected.id, payload);
    action
      .then(() => refreshSelectedVehicle(selected.id))
      .then(resetForm)
      .catch((error: Error) => onMessage(error.message))
      .finally(() => setSaving(false));
  };

  const remove = (value: VehicleCVValue) => {
    if (!selected) return;
    setSaving(true);
    onMessage("");
    api.deleteVehicleCVValue(selected.id, value.id)
      .then(() => refreshSelectedVehicle(selected.id))
      .catch((error: Error) => onMessage(error.message))
      .finally(() => setSaving(false));
  };

  const exportValues = () => {
    if (!selected) return;
    const payload = {
      vehicle: { inventoryNumber: selected.inventoryNumber, name: selected.name, decoder: decoderNumber },
      cvValues: selected.cvValues || []
    };
    const url = URL.createObjectURL(new Blob([JSON.stringify(payload, null, 2)], { type: "application/json" }));
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = `${selected.inventoryNumber || "railkeeper"}-cv.json`;
    anchor.click();
    URL.revokeObjectURL(url);
  };

  const importValues = async (files: FileList | null) => {
    if (!selected || !files || files.length === 0) return;
    const [file] = Array.from(files);
    setSaving(true);
    onMessage("");
    try {
      const values = cvValuesFromImport(await file.text());
      const preview = buildCVImportPreview(file.name, values, selected.cvValues || []);
      if (!preview.rows.some((row) => row.status !== "invalid")) {
        throw new Error("Keine gültigen CV-Werte gefunden.");
      }
      setImportPreview(preview);
    } catch (error) {
      onMessage(error instanceof Error ? error.message : String(error));
    } finally {
      setSaving(false);
      if (importInputRef.current) importInputRef.current.value = "";
    }
  };

  const toggleImportRow = (id: string, selectedRow: boolean) => {
    setImportPreview((current) => current ? {
      ...current,
      rows: current.rows.map((row) => row.id === id ? { ...row, selected: selectedRow } : row)
    } : current);
  };

  const selectImportRows = (mode: "all" | "none" | "empty") => {
    setImportPreview((current) => current ? {
      ...current,
      rows: current.rows.map((row) => ({
        ...row,
        selected: row.status !== "invalid" && (mode === "all" || (mode === "empty" && row.status === "new"))
      }))
    } : current);
  };

  const applyImportPreview = () => {
    if (!selected || !importPreview) return;
    const rows = importPreview.rows.filter((row) => row.selected && row.status !== "invalid");
    if (rows.length === 0) {
      onMessage("Keine CV-Werte für den Import ausgewählt.");
      return;
    }
    setSaving(true);
    onMessage("");
    (async () => {
      for (const row of rows) {
        if (row.existing) await api.updateVehicleCVValue(selected.id, row.existing.id, row.input);
        else await api.createVehicleCVValue(selected.id, row.input);
      }
    })()
      .then(() => refreshSelectedVehicle(selected.id))
      .then(() => {
        setImportPreview(null);
        onMessage(`${rows.length} CV-Wert${rows.length === 1 ? "" : "e"} übernommen.`);
      })
      .catch((error: Error) => onMessage(error.message))
      .finally(() => setSaving(false));
  };

  return {
    state: {
      form,
      editingId,
      importPreview,
      summary,
      importStats,
      storedDecoderProfiles,
      decoderProfileOptions
    },
    refs: { importInputRef },
    commands: {
      updateForm,
      resetForm,
      reset,
      edit,
      save,
      remove,
      exportValues,
      importValues,
      toggleImportRow,
      selectImportRows,
      applyImportPreview,
      setImportPreview,
      discardImportPreview: () => setImportPreview(null)
    }
  };
}
