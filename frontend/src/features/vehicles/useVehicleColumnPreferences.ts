import { useCallback } from "react";

import { translate, useI18n } from "../../shared/i18n";
import { setTableColumnWidth } from "../../shared/tableColumnLayout";
import { useProfileTableLayout } from "../../shared/useProfileTableLayout";
import {
  defaultVehicleTableColumns,
  defaultVehicleTableLayout,
  moveVehicleTableColumn,
  parseVehicleTableLayout,
  serializeVehicleTableLayout,
  toggleVehicleTableColumn,
  vehicleTableColumnWidthDefinitions,
  type VehicleColumnMove,
  type VehicleTableColumn
} from "./vehicleTableColumns";

export const vehicleTableColumnSettingKey = "railkeeper.vehicles.tableColumns";

export function useVehicleColumnPreferences(onMessage: (message: string) => void) {
  const { language } = useI18n();
  const preferences = useProfileTableLayout({
    settingKey: vehicleTableColumnSettingKey,
    defaultLayout: defaultVehicleTableLayout,
    parse: parseVehicleTableLayout,
    serialize: serializeVehicleTableLayout,
    onLoadError: () => onMessage(translate(language, "vehicles.columns.loadError")),
    onSaveError: () => onMessage(translate(language, "vehicles.columns.saveError"))
  });

  const applyColumns = useCallback((
    update: (current: VehicleTableColumn[]) => VehicleTableColumn[]
  ) => {
    preferences.commit((current) => ({
      ...current,
      columns: update(current.columns)
    }));
  }, [preferences]);

  const toggleColumn = useCallback((column: VehicleTableColumn) => {
    applyColumns((current) => toggleVehicleTableColumn(current, column));
  }, [applyColumns]);

  const moveColumn = useCallback((column: VehicleTableColumn, direction: VehicleColumnMove) => {
    applyColumns((current) => moveVehicleTableColumn(current, column, direction));
  }, [applyColumns]);

  const resetColumns = useCallback(() => {
    preferences.commit(() => ({ columns: [...defaultVehicleTableColumns], widths: {} }));
  }, [preferences]);

  const previewColumnWidth = useCallback((column: VehicleTableColumn, width: number) => {
    preferences.preview((current) => setTableColumnWidth(
      current,
      column,
      width,
      vehicleTableColumnWidthDefinitions
    ));
  }, [preferences]);

  const commitColumnWidth = useCallback((column: VehicleTableColumn, width: number) => {
    preferences.commit((current) => setTableColumnWidth(
      current,
      column,
      width,
      vehicleTableColumnWidthDefinitions
    ));
  }, [preferences]);

  return {
    columns: preferences.layout.columns,
    widths: preferences.layout.widths,
    loading: preferences.loading,
    commitColumnWidth,
    moveColumn,
    previewColumnWidth,
    resetColumns,
    toggleColumn
  };
}
