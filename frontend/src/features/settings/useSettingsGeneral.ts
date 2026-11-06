import { useEffect, useState } from "react";
import { api, type InventoryNumberScheme, type VersionInfo } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import type { SettingsTab } from "./settingsModel";
import type { SettingsPreferencesController } from "./useSettingsPreferences";

const updateStatusChangedEvent = "railkeeper-update-status-changed";

const emptyInventorySchemeDraft = {
  category: "",
  prefix: "RK",
  nextNumber: 1,
  padding: 6,
  active: true
};

export function useSettingsGeneral(
  activeSettingsTab: SettingsTab,
  preferences: SettingsPreferencesController,
  loadStorageUsage: () => void
) {
  const { t } = useI18n();
  const { betaUpdates, loadSystemPrinters } = preferences;
  const [inventorySchemes, setInventorySchemes] = useState<InventoryNumberScheme[]>([]);
  const [inventorySchemesLoading, setInventorySchemesLoading] = useState(false);
  const [inventorySchemesMessage, setInventorySchemesMessage] = useState("");
  const [newInventoryScheme, setNewInventoryScheme] = useState(emptyInventorySchemeDraft);
  const [versionInfo, setVersionInfo] = useState<VersionInfo | null>(null);
  const [versionMessage, setVersionMessage] = useState("");
  const [versionLoading, setVersionLoading] = useState(false);

  useEffect(() => {
    if (activeSettingsTab !== "general" || inventorySchemes.length > 0 || inventorySchemesLoading) return;
    loadInventorySchemes();
  }, [activeSettingsTab, inventorySchemes.length, inventorySchemesLoading]);

  useEffect(() => {
    if (activeSettingsTab !== "general") return;
    loadVersionInfo();
    loadStorageUsage();
    loadSystemPrinters();
  }, [activeSettingsTab]);

  const loadVersionInfo = (forceCheck = false, includeBeta = betaUpdates) => {
    setVersionLoading(true);
    setVersionMessage("");
    api
      .version(forceCheck, includeBeta)
      .then((info) => {
        setVersionInfo(info);
        setVersionMessage(info.message || t("settings.updates.message.reachable", { version: info.version || t("settings.unknown") }));
        window.dispatchEvent(new CustomEvent(updateStatusChangedEvent, { detail: { forceCheck } }));
      })
      .catch((error: Error) => setVersionMessage(error.message))
      .finally(() => setVersionLoading(false));
  };

  const loadInventorySchemes = () => {
    setInventorySchemesLoading(true);
    setInventorySchemesMessage("");
    api
      .inventoryNumberSchemes()
      .then(setInventorySchemes)
      .catch((error: Error) => setInventorySchemesMessage(error.message))
      .finally(() => setInventorySchemesLoading(false));
  };

  const updateInventoryScheme = (category: string, patch: Partial<InventoryNumberScheme>) => {
    setInventorySchemes((current) =>
      current.map((scheme) => (scheme.category === category ? { ...scheme, ...patch } : scheme))
    );
  };

  const updateNewInventoryScheme = (patch: Partial<typeof emptyInventorySchemeDraft>) => {
    setNewInventoryScheme((current) => ({ ...current, ...patch }));
  };

  const inventorySchemePreview = (prefix: string, nextNumber: number, padding: number) =>
    `${prefix.trim() || "RK"}-${String(Math.max(1, Number(nextNumber) || 1)).padStart(Math.max(1, Number(padding) || 6), "0")}`;

  const createInventoryScheme = () => {
    setInventorySchemesMessage("");
    api
      .createInventoryNumberScheme({
        category: newInventoryScheme.category,
        prefix: newInventoryScheme.prefix,
        nextNumber: Number(newInventoryScheme.nextNumber) || 1,
        padding: Number(newInventoryScheme.padding) || 6,
        active: newInventoryScheme.active
      })
      .then((created) => {
        setInventorySchemes((current) => [...current, created]);
        setNewInventoryScheme(emptyInventorySchemeDraft);
      })
      .catch((error: Error) => setInventorySchemesMessage(error.message));
  };

  const saveInventoryScheme = (scheme: InventoryNumberScheme) => {
    setInventorySchemesMessage("");
    api
      .updateInventoryNumberScheme(scheme.category, {
        prefix: scheme.prefix,
        nextNumber: Number(scheme.nextNumber) || 1,
        padding: Number(scheme.padding) || 6,
        active: scheme.active
      })
      .then((updated) => updateInventoryScheme(updated.category, updated))
      .catch((error: Error) => setInventorySchemesMessage(error.message));
  };

  return {
    inventorySchemes,
    setInventorySchemes,
    inventorySchemesLoading,
    setInventorySchemesLoading,
    inventorySchemesMessage,
    setInventorySchemesMessage,
    newInventoryScheme,
    setNewInventoryScheme,
    versionInfo,
    setVersionInfo,
    versionMessage,
    setVersionMessage,
    versionLoading,
    setVersionLoading,
    loadVersionInfo,
    loadInventorySchemes,
    updateInventoryScheme,
    updateNewInventoryScheme,
    inventorySchemePreview,
    createInventoryScheme,
    saveInventoryScheme
  };
}

export type SettingsGeneralController = ReturnType<typeof useSettingsGeneral>;
