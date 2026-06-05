import { FormEvent, useEffect, useMemo, useState } from "react";
import type { ReactNode } from "react";
import {
  ChevronDown,
  ChevronUp,
  Database,
  Download,
  Eye,
  EyeOff,
  ExternalLink,
  HardDrive,
  History,
  KeyRound,
  Mail,
  Palette,
  Pencil,
  Plus,
  RefreshCw,
  Shield,
  ShieldAlert,
  Trash2,
  Upload,
  UserCog,
  Users,
  Wrench,
  X
} from "lucide-react";
import type { AppView } from "../../app/App";
import {
  api,
  AuditLogEntry,
  BackupValidationResult,
  InventoryNumberScheme,
  MasterDataEntry,
  MasterDataInput,
  Role,
  Session,
  SessionRecord,
  SMTPSettings,
  SMTPSettingsInput,
  SystemPrinters,
  StorageUsage,
  TwoFactorStatus,
  UserAccount,
  VersionInfo
} from "../../shared/api";
import { Language, useI18n } from "../../shared/i18n";
import { applyStoredThemeOptions, applyThemePreference, readThemePreference, themePreferenceKey, ThemePreference } from "../../shared/theme";
import { SettingsAuthTab } from "./SettingsAuthTab";
import { SettingsDigitalTab } from "./SettingsDigitalTab";

import {
  applyVisibleMetadata,
  articleSearchSettingKey,
  articleSearchSourceOptions,
  articleSearchSourcesSettingKey,
  auditActor,
  auditTarget,
  cv8BinaryText,
  cv8CountryText,
  cv8DecimalText,
  cv8HexText,
  cv8NameText,
  defaultArticleSearchSources,
  defaultSidebarOrder,
  emptyForm,
  emptyPasswordForm,
  emptyUserForm,
  entryToForm,
  externalLink,
  formatBytes,
  formatDateTime,
  loadableMasterDataTypes,
  localSettingKeys,
  manufacturerAliasesText,
  manufacturerNeedsWebsiteReview,
  manufacturerSearchDomainsText,
  manufacturerWebsite,
  masterDataImage,
  masterDataTypes,
  metadataString,
  nominalScalesText,
  normalizeCV8Binary,
  normalizeCV8Decimal,
  normalizeCV8Hex,
  readArticleSearchSources,
  readLocalBool,
  readLocalSetting,
  readSidebarPrefs,
  settingsTabs,
  sidebarOrderChangedEvent,
  sidebarPrefsKey
} from "./settingsModel";
import type { FormState, PasswordFormState, SettingsTab, UserFormState } from "./settingsModel";
import { AppSelect } from "../../shared/ui/AppSelect";

const updateStatusChangedEvent = "railkeeper-update-status-changed";

const emptyInventorySchemeDraft = {
  category: "",
  prefix: "RK",
  nextNumber: 1,
  padding: 6,
  active: true
};

type SettingsConfirmDialog = {
  title: string;
  body?: string;
  confirmLabel: string;
  danger?: boolean;
  onConfirm: () => void;
};

type MasterDataSortDirection = "asc" | "desc";

type MasterDataSortState = {
  key: string;
  direction: MasterDataSortDirection;
};

type MasterDataTableColumn = {
  key: string;
  label: string;
  align?: "left" | "center" | "right";
  wrap?: boolean;
  sortable?: boolean;
  sortValue?: (entry: MasterDataEntry) => string | number;
  render: (entry: MasterDataEntry) => ReactNode;
};

const compareMasterDataValues = (left: string | number | undefined, right: string | number | undefined) => {
  if (typeof left === "number" && typeof right === "number") {
    return left - right;
  }
  return String(left ?? "").localeCompare(String(right ?? ""), "de-DE", {
    numeric: true,
    sensitivity: "base"
  });
};

export function SettingsView({ username }: { username: string }) {
  const { language, setLanguage, t } = useI18n();
  const [activeSettingsTab, setActiveSettingsTab] = useState<SettingsTab>(() => {
    const tab = new URLSearchParams(window.location.search).get("tab");
    return settingsTabs.some((item) => item.id === tab) ? tab as SettingsTab : "general";
  });
  const [activeType, setActiveType] = useState(masterDataTypes[0].type);
  const [itemsByType, setItemsByType] = useState<Record<string, MasterDataEntry[]>>({});
  const [loadedTypes, setLoadedTypes] = useState<Record<string, boolean>>({});
  const [loadingTypes, setLoadingTypes] = useState<Record<string, boolean>>({});
  const [editing, setEditing] = useState<MasterDataEntry | null>(null);
  const [form, setForm] = useState<FormState>(emptyForm);
  const [message, setMessage] = useState("");
  const [saving, setSaving] = useState(false);
  const [search, setSearch] = useState("");
  const [masterDataSort, setMasterDataSort] = useState<MasterDataSortState>({ key: "name", direction: "asc" });
  const [articleSearchEnabled, setArticleSearchEnabled] = useState(
    () => window.localStorage.getItem(articleSearchSettingKey) !== "false"
  );
  const [articleSearchSources, setArticleSearchSources] = useState<string[]>(readArticleSearchSources);
  const [design, setDesign] = useState<ThemePreference>(readThemePreference);
  const [inventorySchemes, setInventorySchemes] = useState<InventoryNumberScheme[]>([]);
  const [inventorySchemesLoading, setInventorySchemesLoading] = useState(false);
  const [inventorySchemesMessage, setInventorySchemesMessage] = useState("");
  const [newInventoryScheme, setNewInventoryScheme] = useState(emptyInventorySchemeDraft);
  const [backupFile, setBackupFile] = useState<File | null>(null);
  const [backupValidation, setBackupValidation] = useState<BackupValidationResult | null>(null);
  const [backupMessage, setBackupMessage] = useState("");
  const [backupRestoreConfirm, setBackupRestoreConfirm] = useState("");
  const [backupSaving, setBackupSaving] = useState(false);
  const [backupValidating, setBackupValidating] = useState(false);
  const [backupFileInputKey, setBackupFileInputKey] = useState(0);
  const [masterDataFile, setMasterDataFile] = useState<File | null>(null);
  const [masterDataMessage, setMasterDataMessage] = useState("");
  const [masterDataSaving, setMasterDataSaving] = useState(false);
  const [masterDataFileInputKey, setMasterDataFileInputKey] = useState(0);
  const [defaultView, setDefaultView] = useState(() => {
    const storedDefaultView = readLocalSetting(localSettingKeys.defaultView, "overview");
    return storedDefaultView === "inventory" ? "vehicles" : storedDefaultView;
  });
  const [sidebarOrder, setSidebarOrder] = useState<AppView[]>(() => readSidebarPrefs(username).order);
  const [sidebarHidden, setSidebarHidden] = useState<AppView[]>(() => readSidebarPrefs(username).hidden);
  const [dateFormat, setDateFormat] = useState(() => readLocalSetting(localSettingKeys.dateFormat, "system"));
  const [timeFormat, setTimeFormat] = useState(() => readLocalSetting(localSettingKeys.timeFormat, "system"));
  const [defaultPrinter, setDefaultPrinter] = useState(() => readLocalSetting(localSettingKeys.defaultPrinter, "system-dialog"));
  const [betaUpdates, setBetaUpdates] = useState(() => readLocalBool(localSettingKeys.betaUpdates, false));
  const [darkBackground, setDarkBackground] = useState(() => readLocalSetting(localSettingKeys.darkBackground, "neutral"));
  const [darkAccent, setDarkAccent] = useState(() => readLocalSetting(localSettingKeys.darkAccent, "green"));
  const [darkStyle, setDarkStyle] = useState(() => readLocalSetting(localSettingKeys.darkStyle, "classic"));
  const [lightBackground, setLightBackground] = useState(() => readLocalSetting(localSettingKeys.lightBackground, "neutral"));
  const [lightAccent, setLightAccent] = useState(() => readLocalSetting(localSettingKeys.lightAccent, "green"));
  const [lightStyle, setLightStyle] = useState(() => readLocalSetting(localSettingKeys.lightStyle, "classic"));
  const [versionInfo, setVersionInfo] = useState<VersionInfo | null>(null);
  const [versionMessage, setVersionMessage] = useState("");
  const [versionLoading, setVersionLoading] = useState(false);
  const [storageUsage, setStorageUsage] = useState<StorageUsage | null>(null);
  const [storageMessage, setStorageMessage] = useState("");
  const [storageLoading, setStorageLoading] = useState(false);
  const [storageOptimizing, setStorageOptimizing] = useState(false);
  const [systemPrinters, setSystemPrinters] = useState<SystemPrinters | null>(null);
  const [currentSession, setCurrentSession] = useState<Session | null>(null);
  const [authMessage, setAuthMessage] = useState("");
  const [availableRoles, setAvailableRoles] = useState<Role[]>([]);
  const [users, setUsers] = useState<UserAccount[]>([]);
  const [usersLoading, setUsersLoading] = useState(false);
  const [auditLog, setAuditLog] = useState<AuditLogEntry[]>([]);
  const [auditLogLoading, setAuditLogLoading] = useState(false);
  const [auditLogMessage, setAuditLogMessage] = useState("");
  const [sessions, setSessions] = useState<SessionRecord[]>([]);
  const [sessionsLoading, setSessionsLoading] = useState(false);
  const [sessionsMessage, setSessionsMessage] = useState("");
  const [editingUser, setEditingUser] = useState<UserAccount | null>(null);
  const [userForm, setUserForm] = useState<UserFormState>(emptyUserForm);
  const [userSaving, setUserSaving] = useState(false);
  const [passwordForm, setPasswordForm] = useState<PasswordFormState>(emptyPasswordForm);
  const [passwordSaving, setPasswordSaving] = useState(false);
  const [passwordMessage, setPasswordMessage] = useState("");
  const [smtpSettings, setSmtpSettings] = useState<SMTPSettings | null>(null);
  const [smtpForm, setSmtpForm] = useState<SMTPSettingsInput & { password: string; testRecipient: string }>({
    enabled: false,
    publicUrl: "",
    host: "",
    port: "587",
    username: "",
    password: "",
    from: "",
    tlsMode: "starttls",
    clearPassword: false,
    testRecipient: ""
  });
  const [smtpLoading, setSmtpLoading] = useState(false);
  const [smtpSaving, setSmtpSaving] = useState(false);
  const [smtpTesting, setSmtpTesting] = useState(false);
  const [smtpMessage, setSmtpMessage] = useState("");
  const [twoFactorStatus, setTwoFactorStatus] = useState<TwoFactorStatus | null>(null);
  const [twoFactorCode, setTwoFactorCode] = useState("");
  const [twoFactorPassword, setTwoFactorPassword] = useState("");
  const [twoFactorLoading, setTwoFactorLoading] = useState(false);
  const [twoFactorSaving, setTwoFactorSaving] = useState(false);
  const [twoFactorMessage, setTwoFactorMessage] = useState("");
  const [confirmDialog, setConfirmDialog] = useState<SettingsConfirmDialog | null>(null);
  const canManageUsers = Boolean(currentSession?.roles.includes("Admin"));
  const languageLocale = language === "de" ? "de-DE" : "en-US";
  const backupRestorePhrase = t("settings.backup.confirmPhrase");
  const backupRestoreConfirmed =
    backupRestoreConfirm.trim().toLocaleUpperCase(languageLocale) === backupRestorePhrase.toLocaleUpperCase(languageLocale);

  const activeDataType = useMemo(
    () => masterDataTypes.find((item) => item.type === activeType) || masterDataTypes[0],
    [activeType]
  );
  const masterLabel = (type: string) => t(`settings.master.${type}`);
  const masterDescription = (type: string) => t(`settings.master.${type}Desc`);
  const auditLabel = (action: string) => {
    const label = t(`settings.auditAction.${action}`);
    return label === `settings.auditAction.${action}` ? action : label;
  };
  const roleDescription = (role: string) => {
    const description = t(`settings.role.${role}`);
    return description === `settings.role.${role}` ? t("settings.roles.custom") : description;
  };
  const inventoryCategoryLabel = (category: string) => {
    const label = t(`settings.category.${category}`);
    return label === `settings.category.${category}` ? category : label;
  };
  const storageCategoryLabel = (key: string, label: string) => {
    const normalized = `${key} ${label}`.toLocaleLowerCase("de-DE");
    const categoryKey =
      normalized.includes("database") || normalized.includes("datenbank") ? "database" :
      normalized.includes("thumbnail") || normalized.includes("vorschau") ? "thumbnails" :
      normalized.includes("image") || normalized.includes("bild") ? "images" :
      normalized.includes("attachment") || normalized.includes("beilage") ? "attachments" :
      normalized.includes("other") || normalized.includes("sonstig") ? "other" : "";
    if (!categoryKey) return label;
    return t(`settings.storage.category.${categoryKey}`);
  };
  const fileCountLabel = (count: number, bytes: number) =>
    t("settings.storage.fileCount", {
      count,
      suffix: count === 1 ? "" : language === "de" ? "en" : "s",
      size: formatBytes(bytes)
    });
  const versionStatusLabel = (info: VersionInfo) => {
    if (info.updateAvailable) return t("settings.updates.status.updateAvailable");
    if (info.status === "current") return betaUpdates ? t("settings.updates.status.currentBeta") : t("settings.updates.status.current");
    if (info.status === "no_release") return t("settings.updates.status.noRelease");
    if (info.status === "unavailable") return t("settings.updates.status.offline");
    return t("settings.updates.status.local");
  };
  const localizedStatusMessage = (message: string) => {
    if (message.includes("Windows-Systemdrucker")) return t("settings.printer.messageLoaded");
    if (message.includes("Keine Release-Information")) return t("settings.updates.message.noRelease");
    if (message.includes("RailKeeper ist aktuell")) return t("settings.updates.message.current");
    return message;
  };
  const updateVersionLabel = (version?: string) => version?.replace(/^v/i, "") || t("settings.unknown");
  const updateCheckedAtLabel = versionInfo?.checkedAt && versionInfo.status !== "local"
    ? t("settings.updates.checkedAt", { date: formatDateTime(versionInfo.checkedAt) })
    : t("settings.updates.notChecked");
  const activeDataLabel = masterLabel(activeDataType.type);
  const activeDataDescription = masterDescription(activeDataType.type);
  const storageFileCount = useMemo(
    () => (storageUsage?.categories || []).reduce((total, category) => total + category.files, 0),
    [storageUsage]
  );
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

  useEffect(() => {
    setEditing(null);
    setForm(emptyForm);
    setSearch("");
    setMessage("");
    setMasterDataSort({ key: "name", direction: "asc" });
  }, [activeType]);

  useEffect(() => {
    if (activeSettingsTab !== "general" || inventorySchemes.length > 0 || inventorySchemesLoading) return;
    loadInventorySchemes();
  }, [activeSettingsTab, inventorySchemes.length, inventorySchemesLoading]);

  useEffect(() => {
    const prefs = readSidebarPrefs(username);
    setSidebarOrder(prefs.order);
    setSidebarHidden(prefs.hidden);
  }, [username]);

  useEffect(() => {
    let cancelled = false;
    api
      .profileSettings()
      .then(({ settings }) => {
        if (cancelled) return;
        const store = (key: string, setter?: (value: string) => void) => {
          const value = settings[key];
          if (value === undefined) return;
          window.localStorage.setItem(key, value);
          setter?.(value);
        };
        store(localSettingKeys.defaultView, (value) => setDefaultView(value === "inventory" ? "vehicles" : value));
        store(localSettingKeys.dateFormat, setDateFormat);
        store(localSettingKeys.timeFormat, setTimeFormat);
        store(localSettingKeys.defaultPrinter, setDefaultPrinter);
        store(localSettingKeys.betaUpdates, (value) => setBetaUpdates(value === "true"));
        store(themePreferenceKey, (value) => {
          if (value === "system" || value === "light" || value === "dark") {
            setDesign(value);
            applyThemePreference(value);
          }
        });
        store(localSettingKeys.darkBackground, setDarkBackground);
        store(localSettingKeys.darkAccent, setDarkAccent);
        store(localSettingKeys.darkStyle, setDarkStyle);
        store(localSettingKeys.lightBackground, setLightBackground);
        store(localSettingKeys.lightAccent, setLightAccent);
        store(localSettingKeys.lightStyle, setLightStyle);
        applyStoredThemeOptions();
        store(articleSearchSettingKey, (value) => setArticleSearchEnabled(value !== "false"));
        const sourceValue = settings[articleSearchSourcesSettingKey];
        if (sourceValue) {
          try {
            const sources = JSON.parse(sourceValue) as string[];
            if (Array.isArray(sources)) {
              window.localStorage.setItem(articleSearchSourcesSettingKey, sourceValue);
              setArticleSearchSources(sources);
            }
          } catch {
            // Ignore malformed profile settings and keep the local fallback.
          }
        }
        const sidebarValue = settings[sidebarPrefsKey(username)];
        if (sidebarValue) {
          try {
            const prefs = JSON.parse(sidebarValue) as { order?: AppView[]; hidden?: AppView[] };
            const nextOrder = Array.isArray(prefs.order) ? prefs.order : [];
            const nextHidden = Array.isArray(prefs.hidden) ? prefs.hidden : [];
            window.localStorage.setItem(sidebarPrefsKey(username), sidebarValue);
            setSidebarOrder(nextOrder);
            setSidebarHidden(nextHidden);
            window.dispatchEvent(new Event(sidebarOrderChangedEvent));
          } catch {
            // Ignore malformed profile settings and keep the local fallback.
          }
        }
      })
      .catch(() => undefined);
    return () => {
      cancelled = true;
    };
  }, [username]);

  useEffect(() => {
    if (activeSettingsTab !== "general") return;
    loadVersionInfo();
    loadStorageUsage();
    loadSystemPrinters();
  }, [activeSettingsTab]);

  useEffect(() => {
    if (activeSettingsTab !== "data" || storageUsage || storageLoading) return;
    loadStorageUsage();
  }, [activeSettingsTab, storageUsage, storageLoading]);

  useEffect(() => {
    loadCurrentSession();
  }, [username]);

  useEffect(() => {
    if (activeSettingsTab !== "auth" || !canManageUsers) return;
    loadUsersAndRoles();
    loadAuditLog();
    loadSessions();
    loadSMTPSettings();
  }, [activeSettingsTab, canManageUsers]);

  useEffect(() => {
    if (activeSettingsTab !== "auth") return;
    loadTwoFactorStatus();
  }, [activeSettingsTab]);

  useEffect(() => {
    if (activeSettingsTab !== "data" || loadedTypes[activeType]) return;

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
      .masterDataAll()
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
  }, [activeSettingsTab, activeType]);

  const reloadActiveType = () => {
    setLoadingTypes((current) => ({ ...current, [activeType]: true }));
    setMessage("");
    api
      .masterData(activeType)
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

  const saveProfileSetting = (key: string, value: string) => {
    void api.updateProfileSettings({ [key]: value }).catch(() => undefined);
  };

  const setLocalSetting = (key: string, value: string, setter: (value: string) => void) => {
    setter(value);
    window.localStorage.setItem(key, value);
    saveProfileSetting(key, value);
    if (key.startsWith("railkeeper.settings.dark") || key.startsWith("railkeeper.settings.light")) {
      applyStoredThemeOptions();
    }
  };

  const setLocalBool = (key: string, value: boolean, setter: (value: boolean) => void) => {
    setter(value);
    window.localStorage.setItem(key, String(value));
    saveProfileSetting(key, String(value));
  };

  const saveSidebarPrefs = (nextOrder: AppView[], nextHidden: AppView[]) => {
    setSidebarOrder(nextOrder);
    setSidebarHidden(nextHidden);
    const value = JSON.stringify({ order: nextOrder, hidden: nextHidden });
    window.localStorage.setItem(sidebarPrefsKey(username), value);
    saveProfileSetting(sidebarPrefsKey(username), value);
    window.dispatchEvent(new Event(sidebarOrderChangedEvent));
  };

  const saveSidebarOrder = (nextOrder: AppView[]) => {
    saveSidebarPrefs(nextOrder, sidebarHidden);
  };

  const moveSidebarItem = (view: AppView, direction: -1 | 1) => {
    const index = sidebarOrder.indexOf(view);
    const nextIndex = index + direction;
    if (index < 0 || nextIndex < 0 || nextIndex >= sidebarOrder.length) return;
    const nextOrder = [...sidebarOrder];
    [nextOrder[index], nextOrder[nextIndex]] = [nextOrder[nextIndex], nextOrder[index]];
    saveSidebarOrder(nextOrder);
  };

  const resetSidebarOrder = () => {
    saveSidebarPrefs(defaultSidebarOrder, []);
  };

  const toggleSidebarVisibility = (view: AppView) => {
    if (view === "settings") return;
    const nextHidden = sidebarHidden.includes(view)
      ? sidebarHidden.filter((item) => item !== view)
      : [...sidebarHidden, view];
    saveSidebarPrefs(sidebarOrder, nextHidden);
  };

  const updateArticleSearchEnabled = (enabled: boolean) => {
    setArticleSearchEnabled(enabled);
    window.localStorage.setItem(articleSearchSettingKey, String(enabled));
    saveProfileSetting(articleSearchSettingKey, String(enabled));
  };
  const updateArticleSearchSource = (source: string, enabled: boolean) => {
    setArticleSearchSources((current) => {
      const next = enabled ? [...new Set([...current, source])] : current.filter((item) => item !== source);
      const normalized = next.length > 0 ? next : ["web"];
      const value = JSON.stringify(normalized);
      window.localStorage.setItem(articleSearchSourcesSettingKey, value);
      saveProfileSetting(articleSearchSourcesSettingKey, value);
      return normalized;
    });
  };

  const updateDesign = (preference: ThemePreference) => {
    setDesign(preference);
    applyThemePreference(preference);
    saveProfileSetting(themePreferenceKey, preference);
  };

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

  const updateBetaUpdates = (enabled: boolean) => {
    setLocalBool(localSettingKeys.betaUpdates, enabled, setBetaUpdates);
  };

  const loadStorageUsage = () => {
    setStorageLoading(true);
    setStorageMessage("");
    api
      .storageUsage()
      .then(setStorageUsage)
      .catch((error: Error) => setStorageMessage(error.message))
      .finally(() => setStorageLoading(false));
  };

  const optimizeStorage = () => {
    setStorageOptimizing(true);
    setStorageMessage("");
    api
      .optimizeStorage()
      .then((result) => {
        setStorageMessage(t("settings.storage.optimizeDone", { bytes: formatBytes(result.reclaimedBytes) }));
        loadStorageUsage();
      })
      .catch((error: Error) => setStorageMessage(error.message))
      .finally(() => setStorageOptimizing(false));
  };

  const loadSystemPrinters = () => {
    api
      .systemPrinters()
      .then((result) => {
        setSystemPrinters(result);
        if (defaultPrinter === "system-dialog" && result.defaultPrinter) {
          const value = `printer:${result.defaultPrinter}`;
          setDefaultPrinter(value);
          window.localStorage.setItem(localSettingKeys.defaultPrinter, value);
          saveProfileSetting(localSettingKeys.defaultPrinter, value);
        }
      })
      .catch(() => undefined);
  };

  const loadCurrentSession = () => {
    setAuthMessage("");
    api
      .session()
      .then(setCurrentSession)
      .catch((error: Error) => setAuthMessage(error.message));
  };

  const loadTwoFactorStatus = () => {
    setTwoFactorLoading(true);
    setTwoFactorMessage("");
    api
      .twoFactorStatus()
      .then(setTwoFactorStatus)
      .catch((error: Error) => setTwoFactorMessage(error.message))
      .finally(() => setTwoFactorLoading(false));
  };

  const loadUsersAndRoles = () => {
    setUsersLoading(true);
    setAuthMessage("");
    Promise.all([api.roles(), api.users()])
      .then(([roles, accounts]) => {
        setAvailableRoles(roles);
        setUsers(accounts);
      })
      .catch((error: Error) => setAuthMessage(error.message))
      .finally(() => setUsersLoading(false));
  };

  const loadAuditLog = () => {
    setAuditLogLoading(true);
    setAuditLogMessage("");
    api
      .auditLog(10)
      .then((result) => setAuditLog(result.entries))
      .catch((error: Error) => setAuditLogMessage(error.message))
      .finally(() => setAuditLogLoading(false));
  };

  const loadSessions = () => {
    setSessionsLoading(true);
    setSessionsMessage("");
    api
      .sessions(5)
      .then(setSessions)
      .catch((error: Error) => setSessionsMessage(error.message))
      .finally(() => setSessionsLoading(false));
  };

  const loadSMTPSettings = () => {
    setSmtpLoading(true);
    setSmtpMessage("");
    api
      .smtpSettings()
      .then((settings) => {
        setSmtpSettings(settings);
        setSmtpForm((current) => ({
          ...current,
          enabled: settings.enabled,
          publicUrl: settings.publicUrl || "",
          host: settings.host || "",
          port: settings.port || "587",
          username: settings.username || "",
          password: "",
          from: settings.from || "",
          tlsMode: settings.tlsMode || "starttls",
          clearPassword: false
        }));
      })
      .catch((error: Error) => setSmtpMessage(error.message))
      .finally(() => setSmtpLoading(false));
  };

  const saveSMTPSettings = (event: FormEvent) => {
    event.preventDefault();
    setSmtpSaving(true);
    setSmtpMessage("");
    api
      .updateSMTPSettings({
        enabled: smtpForm.enabled,
        publicUrl: smtpForm.publicUrl,
        host: smtpForm.host,
        port: smtpForm.port,
        username: smtpForm.username,
        password: smtpForm.password || undefined,
        from: smtpForm.from,
        tlsMode: smtpForm.tlsMode,
        clearPassword: smtpForm.clearPassword
      })
      .then((settings) => {
        setSmtpSettings(settings);
        setSmtpForm((current) => ({
          ...current,
          enabled: settings.enabled,
          publicUrl: settings.publicUrl || "",
          host: settings.host || "",
          port: settings.port || "587",
          username: settings.username || "",
          password: "",
          from: settings.from || "",
          tlsMode: settings.tlsMode || "starttls",
          clearPassword: false
        }));
        setSmtpMessage(t("settings.smtp.saved"));
      })
      .catch((error: Error) => setSmtpMessage(error.message))
      .finally(() => setSmtpSaving(false));
  };

  const testSMTPSettings = () => {
    const recipient = smtpForm.testRecipient.trim();
    if (!recipient) {
      setSmtpMessage(t("settings.smtp.recipientRequired"));
      return;
    }
    setSmtpTesting(true);
    setSmtpMessage("");
    api
      .testSMTPSettings({ recipient })
      .then(() => setSmtpMessage(t("settings.smtp.testSent", { recipient })))
      .catch((error: Error) => setSmtpMessage(error.message))
      .finally(() => setSmtpTesting(false));
  };

  const setupTwoFactor = (onPrepared?: (status: TwoFactorStatus) => void) => {
    setTwoFactorSaving(true);
    setTwoFactorMessage("");
    api
      .setupTwoFactor()
      .then((status) => {
        setTwoFactorStatus(status);
        setTwoFactorMessage(t("settings.auth.twoFactorSetupDone"));
        if (canManageUsers) loadAuditLog();
        onPrepared?.(status);
      })
      .catch((error: Error) => setTwoFactorMessage(error.message))
      .finally(() => setTwoFactorSaving(false));
  };

  const enableTwoFactor = (onEnabled?: (status: TwoFactorStatus) => void) => {
    const code = twoFactorCode.trim();
    if (!code) {
      setTwoFactorMessage(t("settings.auth.twoFactorCodeRequired"));
      return;
    }
    setTwoFactorSaving(true);
    setTwoFactorMessage("");
    api
      .enableTwoFactor({ code })
      .then((status) => {
        setTwoFactorStatus({ ...status, enabled: true, prepared: true });
        setTwoFactorCode("");
        setTwoFactorMessage(t("settings.auth.twoFactorEnabled"));
        loadCurrentSession();
        loadTwoFactorStatus();
        if (canManageUsers) loadUsersAndRoles();
        if (canManageUsers) loadAuditLog();
        onEnabled?.({ ...status, enabled: true, prepared: true });
      })
      .catch((error: Error) => setTwoFactorMessage(error.message))
      .finally(() => setTwoFactorSaving(false));
  };

  const disableTwoFactor = (onDisabled?: () => void) => {
    if (!twoFactorPassword.trim()) {
      setTwoFactorMessage(t("settings.auth.twoFactorPasswordRequired"));
      return;
    }
    setTwoFactorSaving(true);
    setTwoFactorMessage("");
    api
      .disableTwoFactor({ currentPassword: twoFactorPassword, code: twoFactorCode })
      .then(() => {
        setTwoFactorStatus({ enabled: false, prepared: false });
        setTwoFactorCode("");
        setTwoFactorPassword("");
        setTwoFactorMessage(t("settings.auth.twoFactorDisabled"));
        loadCurrentSession();
        if (canManageUsers) {
          loadUsersAndRoles();
          loadSessions();
          loadAuditLog();
        }
        onDisabled?.();
      })
      .catch((error: Error) => setTwoFactorMessage(error.message))
      .finally(() => setTwoFactorSaving(false));
  };

  const confirmSettingsAction = () => {
    const action = confirmDialog?.onConfirm;
    setConfirmDialog(null);
    action?.();
  };

  const revokeSession = (session: SessionRecord) => {
    setConfirmDialog({
      title: t("settings.sessions.revoke"),
      body: t("settings.sessions.revokeConfirm", { username: session.username }),
      confirmLabel: t("settings.sessions.revoke"),
      danger: true,
      onConfirm: () => {
        setSessionsMessage("");
        api
          .revokeSession(session.id)
          .then(() => {
            loadSessions();
            loadAuditLog();
          })
          .catch((error: Error) => setSessionsMessage(error.message));
      }
    });
  };

  const changePassword = (event: FormEvent) => {
    event.preventDefault();
    setPasswordMessage("");
    if (passwordForm.newPassword.length < 12) {
      setPasswordMessage(t("settings.password.error.tooShort"));
      return;
    }
    if (passwordForm.newPassword !== passwordForm.confirmPassword) {
      setPasswordMessage(t("settings.password.error.mismatch"));
      return;
    }
    setPasswordSaving(true);
    api
      .changePassword({
        currentPassword: passwordForm.currentPassword,
        newPassword: passwordForm.newPassword
      })
      .then(() => {
        setPasswordForm(emptyPasswordForm);
        setPasswordMessage(t("settings.password.changed"));
        loadSessions();
        loadAuditLog();
      })
      .catch((error: Error) => setPasswordMessage(error.message))
      .finally(() => setPasswordSaving(false));
  };

  const startUserCreate = () => {
    setEditingUser(null);
    setUserForm(emptyUserForm);
    setAuthMessage("");
  };

  const startUserEdit = (user: UserAccount) => {
    setEditingUser(user);
    setUserForm({
      username: user.username,
      email: user.email || "",
      password: "",
      confirmPassword: "",
      roles: user.roles.length > 0 ? user.roles : ["Viewer"]
    });
    setAuthMessage("");
  };

  const toggleUserRole = (role: string, checked: boolean) => {
    setUserForm((current) => {
      const nextRoles = checked
        ? [...current.roles, role]
        : current.roles.filter((currentRole) => currentRole !== role);
      return { ...current, roles: Array.from(new Set(nextRoles)) };
    });
  };

  const saveUser = (event: FormEvent, onSaved?: () => void) => {
    event.preventDefault();
    setUserSaving(true);
    setAuthMessage("");

    const passwordTouched = userForm.password.length > 0 || userForm.confirmPassword.length > 0;
    if ((!editingUser || passwordTouched) && userForm.password.length < 12) {
      setUserSaving(false);
      setAuthMessage(t("settings.users.passwordTooShort"));
      return;
    }

    if (userForm.password !== userForm.confirmPassword) {
      setUserSaving(false);
      setAuthMessage(t("setup.passwordMismatch"));
      return;
    }

    const input = {
      username: userForm.username,
      email: userForm.email,
      password: userForm.password || undefined,
      roles: userForm.roles
    };
    const action = editingUser ? api.updateUser(editingUser.id, input) : api.createUser(input);

    action
      .then((user) => {
        setEditingUser(user);
        setUserForm({ username: user.username, email: user.email || "", password: "", confirmPassword: "", roles: user.roles });
        loadUsersAndRoles();
        loadAuditLog();
        loadCurrentSession();
        onSaved?.();
      })
      .catch((error: Error) => setAuthMessage(error.message))
      .finally(() => setUserSaving(false));
  };

  const deleteUser = (user: UserAccount) => {
    setConfirmDialog({
      title: t("vehicles.delete"),
      body: t("settings.users.deleteConfirm", { username: user.username }),
      confirmLabel: t("vehicles.delete"),
      danger: true,
      onConfirm: () => {
        setAuthMessage("");
        api
          .deleteUser(user.id)
          .then(() => {
            if (editingUser?.id === user.id) {
              startUserCreate();
            }
            loadUsersAndRoles();
            loadAuditLog();
          })
          .catch((error: Error) => setAuthMessage(error.message));
      }
    });
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

  const deleteEntry = (entry: MasterDataEntry) => {
    setConfirmDialog({
      title: t("vehicles.delete"),
      body: t("settings.master.deleteConfirm", { label: entry.label }),
      confirmLabel: t("vehicles.delete"),
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

  const selectBackupFile = (file: File | null) => {
    setBackupFile(file);
    setBackupValidation(null);
    setBackupMessage("");
    setBackupRestoreConfirm("");
    if (!file) return;

    setBackupValidating(true);
    api
      .validateBackup(file)
      .then((result) => setBackupValidation(result))
      .catch((error: Error) => setBackupMessage(error.message))
      .finally(() => setBackupValidating(false));
  };

  const restoreBackup = () => {
    if (!backupFile) {
      setBackupMessage(t("settings.backup.error.noFile"));
      return;
    }
    if (!backupValidation?.compatible) {
      setBackupMessage(t("settings.backup.error.notValidated"));
      return;
    }
    if (!backupRestoreConfirmed) {
      setBackupMessage(t("settings.backup.error.confirmRequired", { phrase: backupRestorePhrase }));
      return;
    }
    setConfirmDialog({
      title: t("settings.backup.restoreButton"),
      body: t("settings.backup.confirmRestore"),
      confirmLabel: t("settings.backup.restoreButton"),
      danger: true,
      onConfirm: () => {
        setBackupSaving(true);
        setBackupMessage("");
        api
          .restoreBackup(backupFile)
          .then((result) => {
            setBackupMessage(t("settings.backup.restored", { rows: result.restoredRows, files: result.restoredFiles }));
            setLoadedTypes({});
            setItemsByType({});
            setBackupFile(null);
            setBackupValidation(null);
            setBackupRestoreConfirm("");
            setBackupFileInputKey((current) => current + 1);
            loadStorageUsage();
          })
          .catch((error: Error) => setBackupMessage(error.message))
          .finally(() => setBackupSaving(false));
      }
    });
  };

  const importMasterData = () => {
    if (!masterDataFile) {
      setMasterDataMessage(t("settings.masterTransfer.fileMissing"));
      return;
    }
    setConfirmDialog({
      title: t("settings.masterTransfer.upload"),
      body: t("settings.masterTransfer.confirm"),
      confirmLabel: t("settings.masterTransfer.upload"),
      danger: true,
      onConfirm: () => {
        setMasterDataSaving(true);
        setMasterDataMessage("");
        api
          .importMasterData(masterDataFile)
          .then((result) => {
            setMasterDataMessage(t("settings.masterTransfer.imported", { entries: result.importedEntries, relations: result.importedRelations }));
            setLoadedTypes({});
            setItemsByType({});
            setMasterDataFile(null);
            setMasterDataFileInputKey((current) => current + 1);
            setSearch("");
          })
          .catch((error: Error) => setMasterDataMessage(error.message))
          .finally(() => setMasterDataSaving(false));
      }
    });
  };

  const toggleMasterDataSort = (key: string) => {
    setMasterDataSort((current) => ({
      key,
      direction: current.key === key && current.direction === "asc" ? "desc" : "asc"
    }));
  };

  const masterDataColumns: MasterDataTableColumn[] = [];
  if (isSymbolData) {
    masterDataColumns.push({
      key: "symbol",
      label: t("settings.data.symbol"),
      align: "center",
      sortable: false,
      render: (entry) => {
        const symbolImage = masterDataImage(entry);
        return symbolImage ? <img className="master-data-symbol-preview" src={symbolImage} alt="" /> : "-";
      }
    });
  }
  masterDataColumns.push({
    key: "name",
    label: isCV8ManufacturerData ? t("settings.data.cv8Manufacturer") : t("settings.data.name"),
    sortValue: (entry) => isCV8ManufacturerData ? cv8NameText(entry) : entry.label,
    render: (entry) => (
      <>
        <strong>{isCV8ManufacturerData ? cv8NameText(entry) : entry.label}</strong>
        {isManufacturerData && manufacturerNeedsWebsiteReview(entry) && <span className="master-data-warning">{t("settings.data.wikiSourceWarning")}</span>}
      </>
    )
  });
  if (isManufacturerData) {
    masterDataColumns.push(
      {
        key: "nominal-scales",
        label: t("settings.data.nominalScales"),
        align: "center",
        sortValue: (entry) => nominalScalesText(entry),
        render: (entry) => nominalScalesText(entry) || "-"
      },
      {
        key: "manufacturer-website",
        label: t("settings.data.manufacturerWebsite"),
        wrap: true,
        sortValue: (entry) => manufacturerWebsite(entry),
        render: (entry) => manufacturerWebsite(entry) || "-"
      },
      {
        key: "search-domains",
        label: t("settings.data.searchDomains"),
        wrap: true,
        sortValue: (entry) => manufacturerSearchDomainsText(entry),
        render: (entry) => manufacturerSearchDomainsText(entry) || "-"
      }
    );
  }
  if (isCV8ManufacturerData) {
    masterDataColumns.push(
      {
        key: "cv8-decimal",
        label: t("settings.data.cv8Decimal"),
        align: "center",
        sortValue: (entry) => Number(cv8DecimalText(entry)) || 0,
        render: (entry) => <code>{cv8DecimalText(entry) || "-"}</code>
      },
      {
        key: "cv8-binary",
        label: t("settings.data.cv8Binary"),
        align: "center",
        sortValue: (entry) => cv8BinaryText(entry),
        render: (entry) => <code>{cv8BinaryText(entry) || "-"}</code>
      },
      {
        key: "cv8-hex",
        label: t("settings.data.cv8Hex"),
        align: "center",
        sortValue: (entry) => cv8HexText(entry),
        render: (entry) => <code>{cv8HexText(entry) || "-"}</code>
      },
      {
        key: "cv8-country",
        label: t("settings.data.cv8Country"),
        align: "center",
        sortValue: (entry) => cv8CountryText(entry),
        render: (entry) => cv8CountryText(entry) || "-"
      }
    );
  }
  if (!isSymbolData && !isCV8ManufacturerData) {
    masterDataColumns.push({
      key: "link",
      label: t("settings.data.link"),
      align: "center",
      sortValue: (entry) => isManufacturerData ? manufacturerWebsite(entry) : entry.sourceUrl || "",
      render: (entry) => {
        const website = manufacturerWebsite(entry);
        const link = isManufacturerData ? (website ? { href: website, title: t("settings.data.openManufacturerWebsite") } : null) : externalLink(entry);
        return link ? (
          <a className="table-icon-link" href={link.href} target="_blank" rel="noreferrer" aria-label={link.title} title={link.title}>
            <ExternalLink size={16} />
          </a>
        ) : "-";
      }
    });
  }
  if (isSymbolData) {
    masterDataColumns.push({
      key: "description",
      label: t("settings.data.description"),
      wrap: true,
      sortValue: (entry) => metadataString(entry, "description"),
      render: (entry) => metadataString(entry, "description") || "-"
    });
  }
  masterDataColumns.push({
    key: "actions",
    label: t("settings.data.actions"),
    align: "right",
    sortable: false,
    render: (entry) => (
      <div className="table-actions">
        <button type="button" className="icon-button" onClick={() => startEdit(entry)} aria-label={t("vehicles.edit")} title={t("vehicles.edit")}>
          <Pencil size={16} />
        </button>
        <button type="button" className="icon-button danger" onClick={() => deleteEntry(entry)} aria-label={t("vehicles.delete")} title={t("vehicles.delete")}>
          <Trash2 size={16} />
        </button>
      </div>
    )
  });
  const activeSortColumn =
    masterDataColumns.find((column) => column.key === masterDataSort.key && column.sortable !== false && column.sortValue) ||
    masterDataColumns.find((column) => column.key === "name" && column.sortValue);
  const sortedItems = activeSortColumn?.sortValue
    ? [...filteredItems].sort((left, right) => {
      const result = compareMasterDataValues(activeSortColumn.sortValue?.(left), activeSortColumn.sortValue?.(right));
      return masterDataSort.direction === "asc" ? result : -result;
    })
    : filteredItems;

  return (
    <>
      <section className="settings-head">
        <h1>
          {t("settings.title")} {versionInfo?.version && <span>{versionInfo.version}</span>}
        </h1>
        <p>{t("settings.subtitle")}</p>
      </section>

      <nav className="settings-primary-tabs" aria-label={t("settings.title")}>
        {settingsTabs.map((tab) => (
          <button
            key={tab.id}
            type="button"
            className={activeSettingsTab === tab.id ? "active" : ""}
            onClick={() => setActiveSettingsTab(tab.id)}
          >
            {t(tab.labelKey)}
          </button>
        ))}
      </nav>

      {activeSettingsTab === "general" && (
        <section className="settings-dashboard-grid">
          <div className="settings-card-stack">
            <section className="panel settings-card settings-tool-card">
              <div className="settings-section-head">
                <div>
                  <h2>{t("settings.general.title")}</h2>
                  <p>{t("settings.general.subtitle")}</p>
                </div>
              </div>
              <div className="settings-field-grid">
                <label>
                  {t("settings.language")}
                  <AppSelect value={language} onChange={(event) => setLanguage(event.target.value as Language)}>
                    <option value="de">{t("settings.language.de")}</option>
                    <option value="en">{t("settings.language.en")}</option>
                  </AppSelect>
                </label>
                <label>
                  {t("settings.defaultView")}
                  <AppSelect value={defaultView} onChange={(event) => setLocalSetting(localSettingKeys.defaultView, event.target.value, setDefaultView)}>
                    <option value="overview">{t("nav.overview")}</option>
                    <option value="vehicles">{t("nav.vehicles")}</option>
                    <option value="exhibition">{t("nav.exhibition")}</option>
                    <option value="importExport">{t("nav.importExport")}</option>
                    <option value="settings">{t("nav.settings")}</option>
                  </AppSelect>
                </label>
                <label>
                  {t("settings.dateFormat")}
                  <AppSelect value={dateFormat} onChange={(event) => setLocalSetting(localSettingKeys.dateFormat, event.target.value, setDateFormat)}>
                    <option value="system">{t("settings.option.systemDefault")}</option>
                    <option value="de">{t("settings.option.dateDe")}</option>
                    <option value="iso">{t("settings.option.dateIso")}</option>
                  </AppSelect>
                </label>
                <label>
                  {t("settings.timeFormat")}
                  <AppSelect value={timeFormat} onChange={(event) => setLocalSetting(localSettingKeys.timeFormat, event.target.value, setTimeFormat)}>
                    <option value="system">{t("settings.option.systemDefault")}</option>
                    <option value="24h">{t("settings.option.hours24")}</option>
                    <option value="12h">{t("settings.option.hours12")}</option>
                  </AppSelect>
                </label>
                <label className="settings-field-wide">
                  {t("settings.defaultPrinter")}
                  <AppSelect value={defaultPrinter} onChange={(event) => setLocalSetting(localSettingKeys.defaultPrinter, event.target.value, setDefaultPrinter)}>
                    <option value="system-dialog">{t("settings.printer.system")}</option>
                    {(systemPrinters?.printers || []).map((printer) => (
                      <option key={printer.id} value={`printer:${printer.name}`}>
                        {printer.name}{printer.isDefault ? ` (${t("settings.defaultPrinter.default")})` : ""}
                      </option>
                    ))}
                    <option value="ask">{t("settings.printer.ask")}</option>
                    <option value="pdf">{t("settings.printer.pdf")}</option>
                  </AppSelect>
                </label>
              </div>
              <section className="sidebar-order-box" aria-label={t("settings.sidebarOrder.title")}>
                <div>
                  <h3>{t("settings.sidebarOrder.title")}</h3>
                  <p>{t("settings.sidebarOrder.subtitle")}</p>
                </div>
                <div className="sidebar-order-list">
                  {sidebarOrder.map((view, index) => (
                    <div key={view} className={sidebarHidden.includes(view) ? "muted-row" : ""}>
                      <span>{index + 1}</span>
                      <strong>{t(`nav.${view}`)}</strong>
                      <button
                        type="button"
                        className="icon-button"
                        onClick={() => toggleSidebarVisibility(view)}
                        disabled={view === "settings"}
                        aria-label={sidebarHidden.includes(view) ? t("settings.sidebarOrder.show", { label: t(`nav.${view}`) }) : t("settings.sidebarOrder.hide", { label: t(`nav.${view}`) })}
                        title={view === "settings" ? t("settings.sidebarOrder.alwaysActive") : sidebarHidden.includes(view) ? t("settings.sidebarOrder.showShort") : t("settings.sidebarOrder.hideShort")}
                      >
                        {sidebarHidden.includes(view) ? <EyeOff size={15} /> : <Eye size={15} />}
                      </button>
                      <button type="button" className="icon-button" onClick={() => moveSidebarItem(view, -1)} disabled={index === 0} aria-label={t("overview.widget.forward", { label: t(`nav.${view}`) })} title={t("overview.moveForward")}>
                        <ChevronUp size={15} />
                      </button>
                      <button type="button" className="icon-button" onClick={() => moveSidebarItem(view, 1)} disabled={index === sidebarOrder.length - 1} aria-label={t("overview.widget.backward", { label: t(`nav.${view}`) })} title={t("overview.moveBackward")}>
                        <ChevronDown size={15} />
                      </button>
                    </div>
                  ))}
                </div>
                <button type="button" className="secondary-button compact-action" onClick={resetSidebarOrder}>
                  {t("settings.sidebarOrder.reset")}
                </button>
              </section>
            </section>

            <section className="panel settings-card settings-tool-card">
              <div className="settings-section-head">
                <div>
                  <h2>{t("settings.inventoryNumbers.title")}</h2>
                  <p>{t("settings.inventoryNumbers.subtitle")}</p>
                </div>
                <button type="button" className="icon-button" onClick={loadInventorySchemes} aria-label={t("settings.inventoryNumbers.refresh")} title={t("settings.inventoryNumbers.refresh")} disabled={inventorySchemesLoading}>
                  <RefreshCw size={16} />
                </button>
              </div>

              <div className="table-wrap settings-inline-table">
                <table>
                  <thead>
                    <tr>
                      <th>{t("settings.inventoryNumbers.category")}</th>
                      <th>{t("settings.inventoryNumbers.prefix")}</th>
                      <th>{t("settings.inventoryNumbers.next")}</th>
                      <th>{t("settings.inventoryNumbers.padding")}</th>
                      <th>{t("common.active")}</th>
                      <th>{t("settings.inventoryNumbers.preview")}</th>
                      <th>{t("exhibition.actions")}</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr className="inventory-scheme-create-row">
                      <td>
                        <input
                          value={newInventoryScheme.category}
                          onChange={(event) => updateNewInventoryScheme({ category: event.target.value })}
                          placeholder={t("settings.inventoryNumbers.newCategory")}
                        />
                      </td>
                      <td>
                        <input
                          value={newInventoryScheme.prefix}
                          onChange={(event) => updateNewInventoryScheme({ prefix: event.target.value })}
                          placeholder={t("settings.inventoryNumbers.newPrefix")}
                        />
                      </td>
                      <td>
                        <input type="number" min={1} value={newInventoryScheme.nextNumber} onChange={(event) => updateNewInventoryScheme({ nextNumber: Number(event.target.value) })} />
                      </td>
                      <td>
                        <input type="number" min={1} max={12} value={newInventoryScheme.padding} onChange={(event) => updateNewInventoryScheme({ padding: Number(event.target.value) })} />
                      </td>
                      <td>
                        <label className="switch-field" aria-label={t("settings.inventoryNumbers.newActiveAria")}>
                          <input type="checkbox" checked={newInventoryScheme.active} onChange={(event) => updateNewInventoryScheme({ active: event.target.checked })} />
                          <span />
                        </label>
                      </td>
                      <td><code>{inventorySchemePreview(newInventoryScheme.prefix, newInventoryScheme.nextNumber, newInventoryScheme.padding)}</code></td>
                      <td>
                        <button type="button" className="secondary-button compact-action" onClick={createInventoryScheme} disabled={!newInventoryScheme.category.trim() || !newInventoryScheme.prefix.trim()}>
                          <Plus size={15} />
                          {t("settings.inventoryNumbers.create")}
                        </button>
                      </td>
                    </tr>
                    {inventorySchemesLoading && inventorySchemes.length === 0 ? (
                      <tr>
                        <td colSpan={7} className="loading-cell">{t("settings.inventoryNumbers.loading")}</td>
                      </tr>
                    ) : inventorySchemes.length === 0 ? (
                      <tr>
                        <td colSpan={7} className="loading-cell">{t("settings.inventoryNumbers.empty")}</td>
                      </tr>
                    ) : (
                      inventorySchemes.map((scheme) => (
                        <tr key={scheme.id}>
                          <td><strong>{inventoryCategoryLabel(scheme.category)}</strong></td>
                          <td>
                            <input value={scheme.prefix} onChange={(event) => updateInventoryScheme(scheme.category, { prefix: event.target.value })} />
                          </td>
                          <td>
                            <input type="number" min={1} value={scheme.nextNumber} onChange={(event) => updateInventoryScheme(scheme.category, { nextNumber: Number(event.target.value) })} />
                          </td>
                          <td>
                            <input type="number" min={1} max={12} value={scheme.padding} onChange={(event) => updateInventoryScheme(scheme.category, { padding: Number(event.target.value) })} />
                          </td>
                          <td>
                            <label className="switch-field" aria-label={t("settings.inventoryNumbers.activeAria", { category: inventoryCategoryLabel(scheme.category) })}>
                              <input type="checkbox" checked={scheme.active} onChange={(event) => updateInventoryScheme(scheme.category, { active: event.target.checked })} />
                              <span />
                            </label>
                          </td>
                          <td><code>{scheme.preview}</code></td>
                          <td>
                            <button type="button" className="secondary-button compact-action" onClick={() => saveInventoryScheme(scheme)}>
                              {t("vehicles.save")}
                            </button>
                          </td>
                        </tr>
                      ))
                    )}
                  </tbody>
                </table>
              </div>
              {inventorySchemesMessage && <p className="form-message">{inventorySchemesMessage}</p>}
            </section>
          </div>

          <aside className="settings-card-stack">
            <section className="panel settings-card settings-tool-card">
              <div className="settings-card-title">
                <Database size={17} />
                <h2>{t("settings.articleSearch.title")}</h2>
              </div>
              <label className="settings-toggle-row">
                <span>
                  <strong>{t("settings.articleSearch.active")}</strong>
                  <small>{t("settings.articleSearch.help")}</small>
                </span>
                <span className="switch-field">
                  <input type="checkbox" checked={articleSearchEnabled} onChange={(event) => updateArticleSearchEnabled(event.target.checked)} />
                  <span />
                </span>
              </label>
              <div className="article-source-grid" aria-label={t("settings.articleSearch.sources")}>
                {articleSearchSourceOptions.map((option) => (
                  <label key={option.id} className="article-source-option">
                    <input
                      type="checkbox"
                      checked={articleSearchSources.includes(option.id)}
                      onChange={(event) => updateArticleSearchSource(option.id, event.target.checked)}
                      disabled={!articleSearchEnabled}
                    />
                    <span>
                      <strong>{t(option.labelKey)}</strong>
                      <small>{t(option.helpKey)}</small>
                    </span>
                  </label>
                ))}
              </div>
            </section>

            <section className="panel settings-card settings-tool-card">
              <div className="settings-card-title">
                <RefreshCw size={17} />
                <h2>{t("settings.updates.title")}</h2>
              </div>
              <label className="settings-toggle-row">
                <span>
                  <strong>{t("settings.updates.beta")}</strong>
                  <small>{t("settings.updates.betaHelp")}</small>
                </span>
                <span className="switch-field">
                  <input type="checkbox" checked={betaUpdates} onChange={(event) => updateBetaUpdates(event.target.checked)} />
                  <span />
                </span>
              </label>
              <div className="settings-action-row">
                <p>
                  {t("settings.updates.currentVersion")}: <strong>{updateVersionLabel(versionInfo?.version)}</strong>
                  {versionInfo?.latestVersion && <> · {t("settings.updates.latestVersion")}: <strong>{updateVersionLabel(versionInfo.latestVersion)}</strong></>}
                  <br />
                  <small>{updateCheckedAtLabel}</small>
                </p>
                {versionInfo?.status && (
                  <span className={`settings-pill ${versionInfo.updateAvailable ? "active" : ["unavailable", "no_release"].includes(versionInfo.status) ? "muted" : ""}`}>
                    {versionStatusLabel(versionInfo)}
                  </span>
                )}
                <button type="button" className="secondary-button" onClick={() => loadVersionInfo(true)} disabled={versionLoading}>
                  <RefreshCw size={17} />
                  {versionLoading ? t("settings.updates.checking") : t("settings.updates.checkNow")}
                </button>
              </div>
              {versionInfo?.releaseUrl && (
                <a className="settings-link-row" href={versionInfo.releaseUrl} target="_blank" rel="noreferrer">
                  <ExternalLink size={15} />
                  {t("settings.updates.openRelease")}
                </a>
              )}
              {versionInfo?.updateAvailable && (
                <div className="update-decision-panel">
                  <div>
                    <strong>{t("settings.updates.decisionTitle", { version: versionInfo.latestVersion || "" })}</strong>
                  </div>
                  {versionInfo.releaseNotes && <pre>{versionInfo.releaseNotes}</pre>}
                </div>
              )}
              {versionMessage && <p className="form-message">{localizedStatusMessage(versionMessage)}</p>}
            </section>

            <section className="panel settings-card settings-tool-card">
              <div className="settings-section-head">
                <div className="settings-card-title">
                  <HardDrive size={17} />
                  <h2>{t("settings.storage.title")}</h2>
                </div>
                <div className="settings-card-actions">
                  <button type="button" className="icon-button subtle-icon-action" onClick={optimizeStorage} aria-label={t("settings.storage.optimize")} title={t("settings.storage.optimize")} disabled={storageLoading || storageOptimizing}>
                    <Wrench size={15} />
                  </button>
                  <button type="button" className="icon-button" onClick={loadStorageUsage} aria-label={t("settings.storage.refresh")} title={t("settings.data.refresh")} disabled={storageLoading || storageOptimizing}>
                    <RefreshCw size={16} />
                  </button>
                </div>
              </div>
              <p>{t("settings.storage.subtitle")}</p>
              <div className="storage-total">
                <strong>{formatBytes(storageUsage?.totalBytes || 0)}</strong>
                <span>{storageUsage?.updatedAt ? t("settings.storage.updated", { date: formatDateTime(storageUsage.updatedAt) }) : t("settings.storage.notUpdated")}</span>
              </div>
              <div className="storage-list">
                {(storageUsage?.categories || []).map((category) => {
                  const percent = storageUsage?.totalBytes ? Math.round((category.bytes / storageUsage.totalBytes) * 100) : 0;
                  return (
                    <div className="storage-row" key={category.key}>
                      <div>
                        <strong>{storageCategoryLabel(category.key, category.label)}</strong>
                        <span>{fileCountLabel(category.files, category.bytes)}</span>
                      </div>
                      <div className="storage-bar" aria-label={storageCategoryLabel(category.key, category.label) + ": " + percent + "%"}>
                        <span style={{ width: Math.max(2, percent) + "%" }} />
                      </div>
                    </div>
                  );
                })}
                {storageLoading && <p className="empty-state compact">{t("settings.storage.loading")}</p>}
                {!storageLoading && !storageUsage && <p className="empty-state compact">{t("settings.storage.empty")}</p>}
              </div>
              {storageMessage && <p className="form-message">{storageMessage}</p>}
            </section>
          </aside>
        </section>
      )}

      {activeSettingsTab === "data" && (
        <section className="panel settings-card data-card">
          <h2>{t("settings.data.title")}</h2>
          <p>{t("settings.data.subtitle")}</p>

          <nav className="settings-secondary-tabs" aria-label={t("settings.data.nav")}>
            {masterDataTypes.map((item) => (
              <button
                key={item.type}
                type="button"
                className={item.type === activeType ? "active" : ""}
                onClick={() => setActiveType(item.type)}
              >
                {masterLabel(item.type)}
              </button>
            ))}
          </nav>

          <section className="master-data-panel">
            <div className="master-data-head">
              <div>
                <h3>{t("settings.data.manage", { label: activeDataLabel })}</h3>
                <p>{activeDataDescription}</p>
              </div>
              <button type="button" className="icon-button" onClick={reloadActiveType} aria-label={t("settings.data.refresh")} title={t("settings.data.refresh")} disabled={loading}>
                <RefreshCw size={16} />
              </button>
            </div>
            {isCV8ManufacturerData && (
              <a className="source-note" href="https://www.nmra.org/sites/default/files/standards/sandrp/DCC/S/appendix_a_s-9_2_2.pdf" target="_blank" rel="noreferrer">
                <ExternalLink size={15} />
                <span>{t("settings.data.cv8Source")}</span>
              </a>
            )}

            <>
              <label className="settings-search">
                {t("settings.data.search")}
                <input value={search} onChange={(event) => setSearch(event.target.value)} placeholder={t("settings.data.searchPlaceholder")} />
              </label>

                <form className={isCV8ManufacturerData ? "master-data-create cv8-create" : activeType === "manufacturer" ? "master-data-create manufacturer-create" : isSymbolData ? "master-data-create symbol-create" : "master-data-create"} onSubmit={submit}>
                  <strong>{editing ? t("settings.data.editEntry") : t("settings.data.newEntry")}</strong>
                  <input value={form.label} onChange={(event) => update({ label: event.target.value })} placeholder={t("settings.data.entryPlaceholder", { label: activeDataLabel })} required />
                  {isManufacturerData && (
                    <>
                      <input
                        className="master-data-input-compact"
                        value={form.nominalScalesText}
                        onChange={(event) => update({ nominalScalesText: event.target.value })}
                        placeholder={t("settings.data.nominalPlaceholder")}
                      />
                      <input
                        value={form.website}
                        onChange={(event) => update({ website: event.target.value })}
                        placeholder={t("settings.data.manufacturerWebsitePlaceholder")}
                      />
                      <input
                        value={form.searchDomainsText}
                        onChange={(event) => update({ searchDomainsText: event.target.value })}
                        placeholder={t("settings.data.searchDomainsPlaceholder")}
                      />
                      <input
                        value={form.aliasesText}
                        onChange={(event) => update({ aliasesText: event.target.value })}
                        placeholder={t("settings.data.aliasesPlaceholder")}
                      />
                      <input
                        value={form.sourceUrl}
                        onChange={(event) => update({ sourceUrl: event.target.value })}
                        placeholder={t("settings.data.sourceUrlPlaceholder")}
                      />
                    </>
                  )}
                  {isCV8ManufacturerData ? (
                    <>
                      <input
                        value={form.cvDecimal}
                        onChange={(event) => {
                          const value = event.target.value;
                          const decimal = normalizeCV8Decimal(value);
                          update({
                            cvDecimal: value,
                            cvBinary: decimal ? normalizeCV8Binary("", decimal) : form.cvBinary,
                            cvHex: decimal ? normalizeCV8Hex("", decimal) : form.cvHex
                          });
                        }}
                        placeholder={t("settings.data.cv8DecimalPlaceholder")}
                        inputMode="numeric"
                        required
                      />
                      <input
                        value={form.cvBinary}
                        onChange={(event) => update({ cvBinary: normalizeCV8Binary(event.target.value, form.cvDecimal) })}
                        placeholder={t("settings.data.cv8BinaryPlaceholder")}
                        inputMode="numeric"
                      />
                      <input
                        value={form.cvHex}
                        onChange={(event) => update({ cvHex: normalizeCV8Hex(event.target.value, form.cvDecimal) })}
                        placeholder={t("settings.data.cv8HexPlaceholder")}
                      />
                      <input
                        value={form.cvCountry}
                        onChange={(event) => update({ cvCountry: event.target.value.toUpperCase() })}
                        placeholder={t("settings.data.cv8CountryPlaceholder")}
                        maxLength={8}
                      />
                    </>
                  ) : isSymbolData ? (
                    <>
                      <textarea
                        value={form.description}
                        onChange={(event) => update({ description: event.target.value })}
                        placeholder={t("settings.data.descriptionPlaceholder")}
                        rows={2}
                      />
                      <label className="symbol-upload-field">
                        <Upload size={15} aria-hidden="true" />
                        <span>{t("settings.data.uploadImage")}</span>
                        <input
                          type="file"
                          accept="image/svg+xml,image/png,image/jpeg,image/webp,.svg"
                          onChange={(event) => {
                            selectSymbolImage(event.target.files?.[0] || null);
                            event.currentTarget.value = "";
                          }}
                        />
                      </label>
                    </>
                  ) : isManufacturerData ? null : (
                    <input value={form.sourceUrl} onChange={(event) => update({ sourceUrl: event.target.value })} placeholder={t("settings.data.sourceUrlPlaceholder")} />
                  )}
                  <button className="primary-button" disabled={saving}>
                    {saving ? t("vehicles.saving") : editing ? t("vehicles.save") : "+ " + t("settings.data.add")}
                  </button>
                  {editing && (
                    <button type="button" className="icon-button" onClick={startCreate} aria-label={t("vehicles.cancel")} title={t("vehicles.cancel")}>
                      <X size={16} />
                    </button>
                  )}
                </form>

                {isSymbolData && form.imageData && (
                  <div className="symbol-preview-card">
                    <img src={form.imageData} alt="" />
                    <span>{t("settings.data.savedSymbol")}</span>
                    <button type="button" className="icon-button danger" onClick={() => update({ imageData: "" })} aria-label={t("settings.data.removeImage")} title={t("settings.data.removeImage")}>
                      <Trash2 size={16} />
                    </button>
                  </div>
                )}

                <div className="table-wrap master-data-table">
                  <table>
                    <thead>
                      <tr>
                        {masterDataColumns.map((column) => {
                          const isActiveSort = activeSortColumn?.key === column.key;
                          const className = [
                            `master-data-col-${column.key}`,
                            column.align ? `is-${column.align}` : "",
                            column.wrap ? "is-wrapped" : ""
                          ].filter(Boolean).join(" ");
                          return (
                            <th
                              key={column.key}
                              className={className}
                              aria-sort={isActiveSort ? (masterDataSort.direction === "asc" ? "ascending" : "descending") : "none"}
                            >
                              {column.sortable === false || !column.sortValue ? (
                                column.label
                              ) : (
                                <button
                                  type="button"
                                  className={`sort-button ${isActiveSort ? "active" : ""}`}
                                  onClick={() => toggleMasterDataSort(column.key)}
                                >
                                  <span>{column.label}</span>
                                  {isActiveSort && masterDataSort.direction === "asc" ? <ChevronUp size={13} /> : <ChevronDown size={13} />}
                                </button>
                              )}
                            </th>
                          );
                        })}
                      </tr>
                    </thead>
                    <tbody>
                      {loading ? (
                        <tr>
                          <td colSpan={masterDataColumns.length} className="loading-cell">Lade aus lokaler Stammdatenbank...</td>
                        </tr>
                      ) : filteredItems.length === 0 ? (
                        <tr>
                          <td colSpan={masterDataColumns.length} className="loading-cell">{t("settings.data.noEntries")}</td>
                        </tr>
                      ) : (
                        sortedItems.map((entry) => (
                          <tr key={entry.id}>
                            {masterDataColumns.map((column) => {
                              const className = [
                                `master-data-col-${column.key}`,
                                column.align ? `is-${column.align}` : "",
                                column.wrap ? "is-wrapped" : ""
                              ].filter(Boolean).join(" ");
                              return <td key={column.key} className={className}>{column.render(entry)}</td>;
                            })}
                          </tr>
                        ))
                      )}
                    </tbody>
                  </table>
                </div>

                {message && <p className="form-message">{message}</p>}
            </>
          </section>
        </section>
      )}

      {activeSettingsTab === "importExport" && (
        <section className="panel settings-card import-export-card">
          <div className="settings-card-title">
            <Database size={18} />
            <div>
              <h2>{t("settings.import.title")}</h2>
              <p>{t("settings.import.subtitle")}</p>
            </div>
          </div>

          <section className="backup-box master-data-transfer-box">
            <div className="backup-box-head">
              <div>
                <h3>{t("settings.masterTransfer.title")}</h3>
                <p>{t("settings.masterTransfer.subtitle")}</p>
              </div>
              <div className="box-icon-actions">
                <a className="icon-button" href={api.masterDataExportUrl()} aria-label={t("settings.masterTransfer.download")} title={t("settings.masterTransfer.download")}>
                  <Download size={16} />
                </a>
                <button type="button" className="icon-button" onClick={importMasterData} disabled={masterDataSaving || !masterDataFile} aria-label={t("settings.masterTransfer.upload")} title={t("settings.masterTransfer.upload")}>
                  <Upload size={16} />
                </button>
              </div>
            </div>
            <div className="transfer-actions">
              <label className="file-picker-field">
                <span>{t("settings.masterTransfer.file")}</span>
                <span className="file-picker-shell">
                  <span className="file-picker-button">{t("settings.file.pick")}</span>
                  <span className="file-picker-name">{masterDataFile?.name || t("settings.file.none")}</span>
                </span>
                <input
                  key={masterDataFileInputKey}
                  type="file"
                  accept="application/json,.json"
                  onChange={(event) => {
                    setMasterDataFile(event.target.files?.[0] || null);
                    setMasterDataMessage("");
                  }}
                />
              </label>
              {masterDataSaving && <span className="inline-status">{t("settings.masterTransfer.saving")}</span>}
            </div>
            {masterDataMessage && <p className="form-message">{masterDataMessage}</p>}
          </section>

          <div className="backup-grid">
            <section className="backup-box">
              <div className="backup-box-head">
                <div>
                  <h3>{t("settings.backup.export.title")}</h3>
                  <p>{t("settings.backup.export.subtitle")}</p>
                </div>
              </div>
              <div className="backup-summary-strip">
                <span>
                  <strong>{formatBytes(storageUsage?.totalBytes || 0)}</strong>
                  {t("settings.backup.storage")}
                </span>
                <span>
                  <strong>{storageFileCount.toLocaleString("de-DE")}</strong>
                  {t("settings.backup.files")}
                </span>
                <button type="button" className="icon-button" onClick={loadStorageUsage} disabled={storageLoading} aria-label={t("settings.backup.refreshStorage")} title={t("settings.backup.refreshStorage")}>
                  <RefreshCw size={15} />
                </button>
              </div>
              <a className="primary-button backup-export-button" href={api.backupExportUrl()} download>
                <Download size={17} />
                {t("settings.backup.create")}
              </a>
            </section>

            <section className="backup-box warning">
              <div>
                <h3>{t("settings.backup.restore.title")}</h3>
                <p>{t("settings.backup.restore.subtitle")}</p>
              </div>
              <label className="file-picker-field">
                <span>{t("settings.backup.file")}</span>
                <span className="file-picker-shell">
                  <span className="file-picker-button">{t("settings.file.pick")}</span>
                  <span className="file-picker-name">{backupFile?.name || t("settings.file.none")}</span>
                </span>
                <input
                  key={backupFileInputKey}
                  type="file"
                  accept="application/json,.json"
                  onChange={(event) => selectBackupFile(event.target.files?.[0] || null)}
                />
              </label>
              {backupValidating && <p className="backup-validation-status">{t("settings.backup.validating")}</p>}
              {backupValidation && (
                <div className={backupValidation.compatible ? "backup-validation ok" : "backup-validation danger"}>
                  <strong>{backupValidation.compatible ? t("settings.backup.compatible") : t("settings.backup.incompatible")}</strong>
                  <dl>
                    <div>
                      <dt>{t("settings.backup.version")}</dt>
                      <dd>{backupValidation.version || "-"}</dd>
                    </div>
                    <div>
                      <dt>{t("settings.backup.tables")}</dt>
                      <dd>{backupValidation.tableCount}</dd>
                    </div>
                    <div>
                      <dt>{t("settings.backup.rows")}</dt>
                      <dd>{backupValidation.rowCount}</dd>
                    </div>
                    <div>
                      <dt>{t("settings.backup.files")}</dt>
                      <dd>
                        {backupValidation.fileCount} / {formatBytes(backupValidation.fileBytes)}
                      </dd>
                    </div>
                  </dl>
                  {backupValidation.errors.length > 0 && (
                    <div className="backup-validation-list danger">
                      <strong>{t("settings.backup.errors")}</strong>
                      <ul>
                        {backupValidation.errors.map((error) => (
                          <li key={error}>{error}</li>
                        ))}
                      </ul>
                    </div>
                  )}
                  {backupValidation.warnings.length > 0 && (
                    <div className="backup-validation-list warning">
                      <strong>{t("settings.backup.warnings")}</strong>
                      <ul>
                        {backupValidation.warnings.map((warning) => (
                          <li key={warning}>{warning}</li>
                        ))}
                      </ul>
                    </div>
                  )}
                </div>
              )}
              {backupValidation?.compatible && (
                <label className="backup-confirm-field">
                  {t("settings.backup.confirmLabel")}
                  <span>{t("settings.backup.confirmHelp", { phrase: backupRestorePhrase })}</span>
                  <input
                    value={backupRestoreConfirm}
                    onChange={(event) => setBackupRestoreConfirm(event.target.value)}
                    placeholder={backupRestorePhrase}
                    autoComplete="off"
                  />
                </label>
              )}
              <button type="button" className="secondary-button danger" onClick={restoreBackup} disabled={backupSaving || backupValidating || !backupValidation?.compatible || !backupRestoreConfirmed}>
                {backupSaving ? (
                  t("settings.backup.restoring")
                ) : (
                  <>
                    <Upload size={17} />
                    {t("settings.backup.restoreButton")}
                  </>
                )}
              </button>
            </section>
          </div>

          <p className="source-note backup-note">
            <ShieldAlert size={16} aria-hidden="true" />
            <span>{t("settings.backup.note")}</span>
          </p>
          {backupMessage && <p className="form-message">{backupMessage}</p>}
        </section>
      )}

      {activeSettingsTab === "digital" && (
        <SettingsDigitalTab canManageUsers={canManageUsers} formatDateTime={formatDateTime} />
      )}

      {activeSettingsTab === "appearance" && (
        <section className="panel settings-card settings-tool-card">
          <div className="settings-card-title">
            <Palette size={18} />
            <div>
              <h2>{t("settings.appearance.title")}</h2>
              <p>{t("settings.appearance.subtitle")}</p>
            </div>
          </div>

          <div className="appearance-mode-row" role="radiogroup" aria-label={t("settings.appearance.mode")}>
            <label className={design === "system" ? "appearance-option active" : "appearance-option"}>
              <input type="radio" name="theme" value="system" checked={design === "system"} onChange={() => updateDesign("system")} />
              <span>
                <strong>{t("settings.appearance.system")}</strong>
                <small>{t("settings.appearance.systemHelp")}</small>
              </span>
            </label>
            <label className={design === "light" ? "appearance-option active" : "appearance-option"}>
              <input type="radio" name="theme" value="light" checked={design === "light"} onChange={() => updateDesign("light")} />
              <span>
                <strong>{t("settings.appearance.light")}</strong>
                <small>{t("settings.appearance.lightHelp")}</small>
              </span>
            </label>
            <label className={design === "dark" ? "appearance-option active" : "appearance-option"}>
              <input type="radio" name="theme" value="dark" checked={design === "dark"} onChange={() => updateDesign("dark")} />
              <span>
                <strong>{t("settings.appearance.dark")}</strong>
                <small>{t("settings.appearance.darkHelp")}</small>
              </span>
            </label>
          </div>

          <div className="appearance-config-grid">
            <section className="appearance-config-card">
              <h3>{t("settings.appearance.darkMode")} <span className="settings-pill active">{t("settings.appearance.active")}</span></h3>
              <div className="settings-field-grid compact">
                <label>
                  {t("settings.appearance.background")}
                  <AppSelect value={darkBackground} onChange={(event) => setLocalSetting(localSettingKeys.darkBackground, event.target.value, setDarkBackground)}>
                    <option value="neutral">Neutral</option>
                    <option value="warm">Warm</option>
                    <option value="cool">Kühl</option>
                    <option value="oled">OLED Schwarz</option>
                  </AppSelect>
                </label>
                <label>
                  {t("settings.appearance.accent")}
                  <AppSelect value={darkAccent} onChange={(event) => setLocalSetting(localSettingKeys.darkAccent, event.target.value, setDarkAccent)}>
                    <option value="green">Grün</option>
                    <option value="blue">Blau</option>
                    <option value="gold">Gold</option>
                  </AppSelect>
                </label>
                <label>
                  {t("settings.appearance.style")}
                  <AppSelect value={darkStyle} onChange={(event) => setLocalSetting(localSettingKeys.darkStyle, event.target.value, setDarkStyle)}>
                    <option value="classic">Klassisch</option>
                    <option value="compact">Kompakt</option>
                    <option value="contrast">Kontrast</option>
                  </AppSelect>
                </label>
              </div>
            </section>
            <section className="appearance-config-card">
              <h3>{t("settings.appearance.lightMode")}</h3>
              <div className="settings-field-grid compact">
                <label>
                  {t("settings.appearance.background")}
                  <AppSelect value={lightBackground} onChange={(event) => setLocalSetting(localSettingKeys.lightBackground, event.target.value, setLightBackground)}>
                    <option value="neutral">Neutral</option>
                    <option value="warm">Warm</option>
                    <option value="cool">Kühl</option>
                  </AppSelect>
                </label>
                <label>
                  {t("settings.appearance.accent")}
                  <AppSelect value={lightAccent} onChange={(event) => setLocalSetting(localSettingKeys.lightAccent, event.target.value, setLightAccent)}>
                    <option value="green">Grün</option>
                    <option value="blue">Blau</option>
                    <option value="gold">Gold</option>
                  </AppSelect>
                </label>
                <label>
                  {t("settings.appearance.style")}
                  <AppSelect value={lightStyle} onChange={(event) => setLocalSetting(localSettingKeys.lightStyle, event.target.value, setLightStyle)}>
                    <option value="classic">Klassisch</option>
                    <option value="compact">Kompakt</option>
                    <option value="contrast">Kontrast</option>
                  </AppSelect>
                </label>
              </div>
            </section>
          </div>
        </section>
      )}

      {activeSettingsTab === "auth" && (
        <SettingsAuthTab
          t={t}
          currentSession={currentSession}
          twoFactorStatus={twoFactorStatus}
          twoFactorCode={twoFactorCode}
          setTwoFactorCode={setTwoFactorCode}
          twoFactorPassword={twoFactorPassword}
          setTwoFactorPassword={setTwoFactorPassword}
          twoFactorLoading={twoFactorLoading}
          twoFactorSaving={twoFactorSaving}
          twoFactorMessage={twoFactorMessage}
          setupTwoFactor={setupTwoFactor}
          enableTwoFactor={enableTwoFactor}
          disableTwoFactor={disableTwoFactor}
          authMessage={authMessage}
          loadCurrentSession={loadCurrentSession}
          changePassword={changePassword}
          passwordForm={passwordForm}
          setPasswordForm={setPasswordForm}
          passwordSaving={passwordSaving}
          passwordMessage={passwordMessage}
          smtpSettings={smtpSettings}
          smtpForm={smtpForm}
          setSmtpForm={setSmtpForm}
          smtpLoading={smtpLoading}
          smtpSaving={smtpSaving}
          smtpTesting={smtpTesting}
          smtpMessage={smtpMessage}
          loadSMTPSettings={loadSMTPSettings}
          saveSMTPSettings={saveSMTPSettings}
          testSMTPSettings={testSMTPSettings}
          canManageUsers={canManageUsers}
          startUserCreate={startUserCreate}
          editingUser={editingUser}
          saveUser={saveUser}
          userForm={userForm}
          setUserForm={setUserForm}
          availableRoles={availableRoles}
          toggleUserRole={toggleUserRole}
          userSaving={userSaving}
          usersLoading={usersLoading}
          users={users}
          formatDateTime={formatDateTime}
          startUserEdit={startUserEdit}
          deleteUser={deleteUser}
          loadSessions={loadSessions}
          sessionsLoading={sessionsLoading}
          sessions={sessions}
          revokeSession={revokeSession}
          sessionsMessage={sessionsMessage}
          loadAuditLog={loadAuditLog}
          auditLogLoading={auditLogLoading}
          auditLog={auditLog}
          auditLabel={auditLabel}
          auditActor={auditActor}
          auditTarget={auditTarget}
          auditLogMessage={auditLogMessage}
          roleDescription={roleDescription}
        />
      )}

      {confirmDialog && (
        <div className="confirm-layer" role="dialog" aria-modal="true" aria-label={confirmDialog.title}>
          <section className="confirm-card">
            <div className="panel-head form-head">
              <h2>{confirmDialog.title}</h2>
              <button type="button" className="icon-button" onClick={() => setConfirmDialog(null)} aria-label={t("vehicles.close")}>
                <X size={17} />
              </button>
            </div>
            {confirmDialog.body && <p>{confirmDialog.body}</p>}
            <div className="confirm-actions">
              <button type="button" className="secondary-button" onClick={() => setConfirmDialog(null)}>
                {t("vehicles.cancel")}
              </button>
              <button type="button" className={confirmDialog.danger ? "danger-button" : "primary-button"} onClick={confirmSettingsAction}>
                {confirmDialog.confirmLabel}
              </button>
            </div>
          </section>
        </div>
      )}

    </>
  );
}
