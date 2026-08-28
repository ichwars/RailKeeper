import { useCallback, useRef, useState } from "react";

import { api, ApiError, type Vehicle } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import type { DigitalCenterWorkItem } from "./digitalCenterModel";
import {
  digitalCenterExternalMapping,
  type DigitalCenterVehicleAdoptionProvider
} from "./digitalCenterVehicleAdoption";

type DigitalCenterVehicleAdoptionOptions = {
  onAssigned: () => void | Promise<void>;
};

export function useDigitalCenterVehicleAdoption({ onAssigned }: DigitalCenterVehicleAdoptionOptions) {
  const { t } = useI18n();
  const loadRequestRef = useRef(0);
  const [vehicles, setVehicles] = useState<Vehicle[]>([]);
  const [selectedVehicleId, setSelectedVehicleId] = useState("");
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  const reset = useCallback(() => {
    loadRequestRef.current += 1;
    setVehicles([]);
    setSelectedVehicleId("");
    setLoading(false);
    setSaving(false);
    setError("");
  }, []);

  const load = useCallback(async (_item: DigitalCenterWorkItem) => {
    const requestId = ++loadRequestRef.current;
    setVehicles([]);
    setSelectedVehicleId("");
    setError("");
    setLoading(true);
    try {
      const result = await api.vehicles("");
      if (requestId === loadRequestRef.current) setVehicles(result);
    } catch {
      if (requestId === loadRequestRef.current) setError(t("digitalCenters.assignment.loadError"));
    } finally {
      if (requestId === loadRequestRef.current) setLoading(false);
    }
  }, [t]);

  const assign = useCallback(async (
    item: DigitalCenterWorkItem,
    provider: DigitalCenterVehicleAdoptionProvider,
    vehicleId: string
  ) => {
    if (!vehicleId.trim()) {
      setError(t("digitalCenters.assignment.selectionRequired"));
      return;
    }
    setSelectedVehicleId(vehicleId);
    setError("");
    setSaving(true);
    try {
      await api.upsertVehicleExternalMapping(vehicleId, digitalCenterExternalMapping(item, provider));
      await onAssigned();
    } catch (assignError) {
      setError(assignError instanceof ApiError && assignError.code === "external_mapping_conflict"
        ? t("digitalCenters.assignment.mappingConflict")
        : t("digitalCenters.assignment.saveError"));
    } finally {
      setSaving(false);
    }
  }, [onAssigned, t]);

  return {
    state: { vehicles, selectedVehicleId, loading, saving, error },
    setters: { setSelectedVehicleId },
    commands: { load, assign, reset }
  };
}
