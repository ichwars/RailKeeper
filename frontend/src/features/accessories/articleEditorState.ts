import {
  type AccessoryAsset,
  type AccessoryDocument,
  type AccessoryInstallation,
  type AccessoryPurchase,
  type AccessoryReservation,
  type AccessoryStockMovement,
  type AccessoryStockSummary,
  type AccessoryUsageHistory,
  type StorageLocation,
  type Vehicle
} from "../../shared/api";

export type ArticleEditorPermissions = {
  canEdit: boolean;
  canManageStock: boolean;
  canReserve: boolean;
  canInstall: boolean;
};

export type ArticleEditorResources = {
  locations: StorageLocation[];
  stock: AccessoryStockSummary | null;
  movements: AccessoryStockMovement[];
  assets: AccessoryAsset[];
  purchases: AccessoryPurchase[];
  documents: AccessoryDocument[];
  documentsLoaded?: boolean;
  reservations: AccessoryReservation[];
  installations: AccessoryInstallation[];
  usageHistory: AccessoryUsageHistory | null;
  vehicles: Vehicle[];
};

export const emptyResources = (): ArticleEditorResources => ({
  locations: [],
  stock: null,
  movements: [],
  assets: [],
  purchases: [],
  documents: [],
  documentsLoaded: false,
  reservations: [],
  installations: [],
  usageHistory: null,
  vehicles: [],
});

export function errorMessage(reason: unknown): string {
  return reason instanceof Error ? reason.message : "Die Artikelaktion konnte nicht abgeschlossen werden.";
}
