import { useEffect, useMemo, useState, type FormEvent } from "react";
import { api, type MasterDataEntry, type MasterDataInput } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { useMasterDataBulkSelection } from "./MasterDataBulkSelection";
import {
  applyVisibleMetadata,
  cv8NameText,
  cv8DecimalText,
  cv8BinaryText,
  cv8HexText,
  cv8CountryText,
  emptyForm,
  entryToForm,
  loadableMasterDataTypes,
  masterDataTypes,
  manufacturerWebsite,
  manufacturerSearchDomainsText,
  manufacturerAliasesText,
  metadataString,
  normalizeCV8Decimal,
  type FormState,
  type SettingsConfirmDialog
} from "./settingsModel";
import type { SettingsNavigationController } from "./useSettingsNavigation";

type MasterDataSortDirection = "asc" | "desc";

type MasterDataSortState = {
  key: string;
  direction: MasterDataSortDirection;
};

type SettingsMasterDataOptions = {
  navigation: SettingsNavigationController;
  canEditMasterData: boolean;
  requestConfirmation: (dialog: SettingsConfirmDialog) => void;
};

export function useSettingsMasterData({
  navigation, canEditMasterData, requestConfirmation
}: SettingsMasterDataOptions) {
  const { t } = useI18n();
  const { activeSettingsTab, activeType, dataGroup } = navigation;
  const [itemsByType, setItemsByType] = useState<Record<string, MasterDataEntry[]>>({});
  const [loadedTypes, setLoadedTypes] = useState<Record<string, boolean>>({});
  const [loadingTypes, setLoadingTypes] = useState<Record<string, boolean>>({});
  const [editing, setEditing] = useState<MasterDataEntry | null>(null);
  const [form, setForm] = useState<FormState>(emptyForm);
  const [message, setMessage] = useState("");
  const [saving, setSaving] = useState(false);
  const [masterDataBatchBusy, setMasterDataBatchBusy] = useState(false);
  const [search, setSearch] = useState("");
  const [masterDataSort, setMasterDataSort] = useState<MasterDataSortState>({ key: "name", direction: "asc" });
  const activeDataType = useMemo(
    () => masterDataTypes.find((item) => item.type === activeType) || masterDataTypes[0],
    [activeType]
  );
  const masterLabel = (type: string) => t(`settings.master.${type}`);
  const masterDescription = (type: string) => t(`settings.master.${type}Desc`);
  const activeDataLabel = masterLabel(activeDataType.type);
  const activeDataDescription = masterDescription(activeDataType.type);
  const items = itemsByType[activeType] || [];
  const loading = Boolean(loadingTypes[activeType]);
  const isSymbolData = activeType === "symbols";
  const isManufacturerData = activeType === "manufacturer";
  const isCV8ManufacturerData = activeType === "cv8_manufacturer";

  const filteredItems = useMemo(() => {
    const needle = search.trim().toLocaleLowerCase("de-DE");
    if (!needle) return items;
    return items.filter((entry) =>
      `${entry.label} ${cv8NameText(entry)} ${entry.key} ${entry.sourceUrl || ""} ${manufacturerWebsite(entry)} ${manufacturerSearchDomainsText(entry)} ${manufacturerAliasesText(entry)} ${metadataString(entry, "description")} ${cv8DecimalText(entry)} ${cv8BinaryText(entry)} ${cv8HexText(entry)} ${cv8CountryText(entry)}`.toLocaleLowerCase("de-DE").includes(needle)
    );
  }, [items, search]);
  const masterDataSelection = useMasterDataBulkSelection(
    `${activeType}|${search}`,
    items,
    filteredItems
  );

  useEffect(() => {
    setEditing(null);
    setForm(emptyForm);
    setSearch("");
    setMessage("");
    setMasterDataSort({ key: "name", direction: "asc" });
  }, [activeType]);

  useEffect(() => {
    if (activeSettingsTab !== "data" || dataGroup !== "general" || loadedTypes[activeType]) return;

    let cancelled = false;
    const typesToLoad = loadableMasterDataTypes
      .map((item) => item.type)
      .filter((typeName) => !loadedTypes[typeName]);

    setMessage("");
    setLoadingTypes((current) => ({
      ...current,
      ...Object.fromEntries(typesToLoad.map((typeName) => [typeName, true]))
    }));

    api
      .managedMasterDataAll()
      .then((entriesByType) => {
        if (cancelled) return;

        const normalized = Object.fromEntries(
          loadableMasterDataTypes.map((item) => [item.type, entriesByType[item.type] || []])
        );
        const loaded = Object.fromEntries(loadableMasterDataTypes.map((item) => [item.type, true]));
        setItemsByType((current) => ({ ...current, ...normalized }));
        setLoadedTypes((current) => ({ ...current, ...loaded }));
      })
      .catch((error: Error) => {
        if (!cancelled) {
          setMessage(error.message);
        }
      })
      .finally(() => {
        if (!cancelled) {
          setLoadingTypes((current) => ({
            ...current,
            ...Object.fromEntries(typesToLoad.map((typeName) => [typeName, false]))
          }));
        }
      });

    return () => {
      cancelled = true;
    };
  }, [activeSettingsTab, activeType, dataGroup]);

  const reloadActiveType = () => {
    masterDataSelection.clear();
    setLoadingTypes((current) => ({ ...current, [activeType]: true }));
    setMessage("");
    api
      .managedMasterData(activeType)
      .then((entries) => {
        setItemsByType((current) => ({ ...current, [activeType]: entries }));
        setLoadedTypes((current) => ({ ...current, [activeType]: true }));
      })
      .catch((error: Error) => setMessage(error.message))
      .finally(() => setLoadingTypes((current) => ({ ...current, [activeType]: false })));
  };

  const update = (patch: Partial<FormState>) => {
    setForm((current) => ({ ...current, ...patch }));
  };

  const selectSymbolImage = (file: File | null) => {
    setMessage("");
    if (!file) return;

    const isAccepted = file.type.startsWith("image/") || file.name.toLocaleLowerCase("de-DE").endsWith(".svg");
    if (!isAccepted) {
      setMessage("Bitte SVG, PNG, JPG oder WebP als Symbolbild auswählen.");
      return;
    }
    if (file.size > 1024 * 1024) {
      setMessage("Symbolbild bitte unter 1 MB halten.");
      return;
    }

    const reader = new FileReader();
    reader.onload = () => {
      if (typeof reader.result === "string") {
        update({ imageData: reader.result });
      }
    };
    reader.onerror = () => setMessage("Symbolbild konnte nicht gelesen werden.");
    reader.readAsDataURL(file);
  };

  const startCreate = () => {
    setEditing(null);
    setForm(emptyForm);
    setMessage("");
  };

  const startEdit = (entry: MasterDataEntry) => {
    setEditing(entry);
    setForm(entryToForm(entry));
    setMessage("");
  };

  const submit = (event: FormEvent) => {
    event.preventDefault();

    setSaving(true);
    setMessage("");

    let metadata: Record<string, unknown>;
    try {
      metadata = JSON.parse(form.metadataText || "{}");
    } catch {
      setSaving(false);
      setMessage(t("settings.master.metadataInvalid"));
      return;
    }

    const cv8Decimal = isCV8ManufacturerData ? normalizeCV8Decimal(form.cvDecimal) : "";
    if (isCV8ManufacturerData && !cv8Decimal) {
      setSaving(false);
      setMessage(t("settings.master.cv8Invalid"));
      return;
    }

    metadata = applyVisibleMetadata(activeType, metadata, form);

    const input: MasterDataInput = {
      key: isCV8ManufacturerData && !editing ? `cv8-${cv8Decimal.padStart(3, "0")}` : form.key,
      label: isCV8ManufacturerData ? form.label.replace(/^\s*\d{1,3}\s*-\s*/, "").trim() : form.label,
      active: form.active,
      sortOrder: Number(form.sortOrder) || 0,
      sourceUrl: isSymbolData || isCV8ManufacturerData ? "" : form.sourceUrl,
      metadata
    };

    const action = editing
      ? api.updateMasterData(activeType, editing.key, input)
      : api.createMasterData(activeType, input);

    action
      .then((entry) => {
        setEditing(entry);
        setForm(entryToForm(entry));
        reloadActiveType();
      })
      .catch((error: Error) => setMessage(error.message))
      .finally(() => setSaving(false));
  };

  const setEntryActive = (entry: MasterDataEntry, active: boolean) => {
    setMessage("");
    api
      .setMasterDataActive(activeType, entry.key, active)
      .then((updated) => {
        setItemsByType((current) => ({
          ...current,
          [activeType]: (current[activeType] || []).map((item) => item.key === updated.key ? updated : item)
        }));
        if (editing?.key === updated.key) {
          setEditing(updated);
          setForm(entryToForm(updated));
        }
      })
      .catch((error: Error) => setMessage(error.message));
  };

  const deactivateEntry = (entry: MasterDataEntry) => {
    requestConfirmation({
      title: t("settings.master.deactivateTitle"),
      body: t("settings.master.deactivateBody"),
      confirmLabel: t("settings.master.deactivate"),
      onConfirm: () => setEntryActive(entry, false)
    });
  };

  const deactivateSelectedEntries = () => {
    const keys = [...masterDataSelection.selectedKeys];
    const count = keys.length;
    if (count === 0) return;
    requestConfirmation({
      title: t("settings.master.deactivateManyTitle"),
      body: t("settings.master.deactivateManyBody", { count, type: activeDataLabel }),
      confirmLabel: t(
        count === 1 ? "settings.master.deactivateOneConfirm" : "settings.master.deactivateManyConfirm",
        { count }
      ),
      onConfirm: () => {
        setMasterDataBatchBusy(true);
        setMessage("");
        api.setMasterDataActiveMany(activeType, keys, false)
          .then((updated) => {
            const byKey = new Map(updated.map((entry) => [entry.key, entry]));
            setItemsByType((current) => ({
              ...current,
              [activeType]: (current[activeType] || []).map((entry) => byKey.get(entry.key) || entry)
            }));
            if (editing) {
              const updatedEditing = byKey.get(editing.key);
              if (updatedEditing) {
                setEditing(updatedEditing);
                setForm(entryToForm(updatedEditing));
              }
            }
            masterDataSelection.clear();
          })
          .catch((error: Error) => setMessage(error.message))
          .finally(() => setMasterDataBatchBusy(false));
      }
    });
  };

  const deleteEntry = (entry: MasterDataEntry) => {
    requestConfirmation({
      title: t("settings.master.deleteTitle"),
      body: t("settings.master.deleteBody"),
      confirmLabel: t("settings.master.deletePermanently"),
      danger: true,
      onConfirm: () => {
        api
          .deleteMasterData(activeType, entry.key)
          .then(() => {
            if (editing?.key === entry.key) {
              startCreate();
            }
            reloadActiveType();
          })
          .catch((error: Error) => setMessage(error.message));
      }
    });
  };

  const toggleMasterDataSort = (key: string) => {
    setMasterDataSort((current) => ({
      key,
      direction: current.key === key && current.direction === "asc" ? "desc" : "asc"
    }));
  };

  const invalidateMasterData = () => {
    setLoadedTypes({});
    setItemsByType({});
  };
  const clearMasterDataSearch = () => setSearch("");

  return {
    activeType, canEditMasterData, editing, form, message, saving, masterDataBatchBusy,
    search, setSearch, masterDataSort, activeDataLabel, activeDataDescription, loading,
    isSymbolData, isManufacturerData, isCV8ManufacturerData, filteredItems, masterDataSelection,
    reloadActiveType, update, selectSymbolImage, startCreate, startEdit, submit, setEntryActive,
    deactivateEntry, deactivateSelectedEntries, deleteEntry, toggleMasterDataSort,
    invalidateMasterData, clearMasterDataSearch
  };
}

export type SettingsMasterDataController = ReturnType<typeof useSettingsMasterData>;
