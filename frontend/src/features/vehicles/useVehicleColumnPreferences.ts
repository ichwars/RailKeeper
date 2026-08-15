import { useCallback, useEffect, useRef, useState } from "react";

import { api } from "../../shared/api";
import { translate, useI18n } from "../../shared/i18n";
import {
  defaultVehicleTableColumns,
  moveVehicleTableColumn,
  parseVehicleTableColumns,
  serializeVehicleTableColumns,
  toggleVehicleTableColumn,
  type VehicleColumnMove,
  type VehicleTableColumn
} from "./vehicleTableColumns";

export const vehicleTableColumnSettingKey = "railkeeper.vehicles.tableColumns";

export function useVehicleColumnPreferences(onMessage: (message: string) => void) {
  const [columns, setColumns] = useState<VehicleTableColumn[]>(() => [
    ...defaultVehicleTableColumns
  ]);
  const [loading, setLoading] = useState(true);
  const saveQueue = useRef<Promise<void>>(Promise.resolve());
  const onMessageRef = useRef(onMessage);
  const { language } = useI18n();

  useEffect(() => {
    onMessageRef.current = onMessage;
  }, [onMessage]);

  useEffect(() => {
    let cancelled = false;

    api.profileSettings()
      .then(({ settings }) => {
        if (!cancelled) {
          setColumns(parseVehicleTableColumns(settings[vehicleTableColumnSettingKey]));
        }
      })
      .catch(() => {
        if (!cancelled) {
          onMessageRef.current(translate(language, "vehicles.columns.loadError"));
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [language]);

  const queueSave = useCallback((next: VehicleTableColumn[]) => {
    saveQueue.current = saveQueue.current
      .catch(() => undefined)
      .then(() => api.updateProfileSettings({
        [vehicleTableColumnSettingKey]: serializeVehicleTableColumns(next)
      }))
      .then(() => undefined)
      .catch(() => {
        onMessageRef.current(translate(language, "vehicles.columns.saveError"));
      });
  }, [language]);

  const applyColumns = useCallback((
    update: (current: VehicleTableColumn[]) => VehicleTableColumn[]
  ) => {
    setColumns((current) => {
      const next = update(current);
      queueSave(next);
      return next;
    });
  }, [queueSave]);

  const toggleColumn = useCallback((column: VehicleTableColumn) => {
    applyColumns((current) => toggleVehicleTableColumn(current, column));
  }, [applyColumns]);

  const moveColumn = useCallback((column: VehicleTableColumn, direction: VehicleColumnMove) => {
    applyColumns((current) => moveVehicleTableColumn(current, column, direction));
  }, [applyColumns]);

  const resetColumns = useCallback(() => {
    applyColumns(() => [...defaultVehicleTableColumns]);
  }, [applyColumns]);

  return {
    columns,
    loading,
    moveColumn,
    resetColumns,
    toggleColumn
  };
}
