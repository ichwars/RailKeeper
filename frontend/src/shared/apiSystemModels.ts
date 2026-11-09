import type { components } from "../generated/api";

export type BackupImportResult = {
  restoredTables: number;
  restoredRows: number;
  restoredFiles: number;
};

export type BackupValidationTable = {
  name: string;
  rows: number;
  missing: boolean;
  unknownColumns?: string[];
};

export type BackupValidationResult = {
  compatible: boolean;
  format?: string;
  version: number;
  createdAt?: string;
  tableCount: number;
  rowCount: number;
  fileCount: number;
  fileBytes: number;
  tables: BackupValidationTable[];
  warnings: string[];
  errors: string[];
};

export type VersionInfo = components["schemas"]["VersionInfo"];

export type StorageUsageCategory = components["schemas"]["StorageUsageCategory"];

export type StorageUsage = components["schemas"]["StorageUsage"];

export type StorageOptimizeResult = components["schemas"]["StorageOptimizeResult"];

export type StorageLocationInfo = {
  dataPath: string;
  mode: "windows_standalone" | "configured" | "server";
  openFolderAvailable: boolean;
  migrationReceipt?: {
    sourcePath: string;
    targetPath: string;
    migratedAt: string;
    version: string;
    filesVerified: number;
    acknowledged: boolean;
  };
};

export type SystemPrinter = {
  id: string;
  name: string;
  isDefault: boolean;
};

export type SystemPrinters = {
  status: "available" | "configured" | "unavailable";
  message: string;
  defaultPrinter?: string;
  printers: SystemPrinter[];
};

export type AuditLogEntry = {
  id: string;
  actorUserId?: string;
  actorUsername?: string;
  action: string;
  targetType?: string;
  targetId?: string;
  createdAt: string;
  detailsJson: string;
};

export type AuditLogResponse = {
  entries: AuditLogEntry[];
};
