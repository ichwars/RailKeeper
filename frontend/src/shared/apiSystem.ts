import type { DigitalCenterSettings, SMTPSettings, SMTPSettingsInput, SMTPTestRequest } from "./apiAuthModels";
import type {
  AuditLogResponse, BackupImportResult, BackupValidationResult, StorageLocationInfo, StorageOptimizeResult,
  StorageUsage, SystemPrinters, VersionInfo
} from "./apiSystemModels";
import type { APIRequest } from "./apiTransport";

export function createSystemAPI(request: APIRequest) {
  return {
    version: (check = false, includePrerelease = false) =>
      request<VersionInfo>(
        `/version${check ? `?check=true&prerelease=${includePrerelease ? "true" : "false"}` : ""}`,
        {},
        { timeoutMs: 10000 }
      ),
    storageUsage: () => request<StorageUsage>("/system/storage", {}, { timeoutMs: 30000 }),
    optimizeStorage: () => request<StorageOptimizeResult>("/system/storage/optimize", { method: "POST" }, { timeoutMs: 120000 }),
    storageLocationInfo: () => request<StorageLocationInfo>("/system/storage/info"),
    openStorageFolder: () => request<void>("/system/storage/open-folder", { method: "POST" }),
    acknowledgeStorageMigration: () =>
      request<void>("/system/storage/migration-receipt/acknowledge", { method: "POST" }),
    systemPrinters: () => request<SystemPrinters>("/system/printers", {}, { timeoutMs: 10000 }),
    auditLog: (limit = 50) => request<AuditLogResponse>(`/system/audit-log?limit=${encodeURIComponent(String(limit))}`, {}, { timeoutMs: 10000 }),
    smtpSettings: () => request<SMTPSettings>("/system/smtp", {}, { timeoutMs: 10000 }),
    updateSMTPSettings: (input: SMTPSettingsInput) =>
      request<SMTPSettings>("/system/smtp", {
        method: "PUT",
        body: JSON.stringify(input)
      }),
    testSMTPSettings: (input: SMTPTestRequest) =>
      request<{ status: string }>("/system/smtp/test", {
        method: "POST",
        body: JSON.stringify(input)
      }, { timeoutMs: 30000 }),
    digitalSettings: () => request<DigitalCenterSettings>("/system/digital-settings", {}, { timeoutMs: 10000 }),
    updateDigitalSettings: (input: DigitalCenterSettings) =>
      request<DigitalCenterSettings>("/system/digital-settings", {
        method: "PUT",
        body: JSON.stringify(input)
      }),
    backupExportUrl: () => "/api/v1/backup/export",
    validateBackup: (file: File) => {
      const form = new FormData();
      form.append("file", file);
      return request<BackupValidationResult>(
        "/backup/validate",
        {
          method: "POST",
          body: form
        },
        { timeoutMs: 120000 }
      );
    },
    restoreBackup: (file: File) => {
      const form = new FormData();
      form.append("file", file);
      return request<BackupImportResult>(
        "/backup/restore",
        {
          method: "POST",
          body: form
        },
        { timeoutMs: 120000 }
      );
    }
  };
}
