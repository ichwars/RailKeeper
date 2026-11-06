import { useSettingsPreferences } from "./useSettingsPreferences";
import { useSettingsGeneral } from "./useSettingsGeneral";
import { SettingsProfileCard } from "./SettingsProfileCard";
import { SettingsInventoryCard } from "./SettingsInventoryCard";
import { SettingsArticleSearchCard } from "./SettingsArticleSearchCard";
import { SettingsUpdatesCard } from "./SettingsUpdatesCard";
import { SettingsAppearanceTab } from "./SettingsAppearanceTab";
import { FormEvent, useEffect, useMemo, useState } from "react";
import type { ReactNode } from "react";
import { ChevronDown, ChevronUp, ExternalLink, RefreshCw, Trash2, Upload, X } from "lucide-react";

import {
  api,
  AuditLogEntry,
  MasterDataEntry,
  MasterDataInput,
  Role,
  Session,
  SessionRecord,
  SMTPSettings,
  SMTPSettingsInput,
  TwoFactorStatus,
  UserAccount
} from "../../shared/api";
import { useI18n } from "../../shared/i18n";

import { SettingsAuthTab } from "./SettingsAuthTab";
import { SettingsDigitalTab } from "./SettingsDigitalTab";
import { ArticleManagementSettings } from "./ArticleManagementSettings";
import { DecoderLibrarySettings } from "./DecoderLibrarySettings";
import {
  MasterDataBulkEntryCheckbox,
  MasterDataBulkSelectAllCheckbox,
  MasterDataBulkToolbar,
  useMasterDataBulkSelection
} from "./MasterDataBulkSelection";
import { MasterDataLifecycleActions } from "./MasterDataLifecycleActions";
import {
  generalDataTypes,
  isArticleDataType,
  isGeneralDataType,
  readSettingsLocation,
  settingsLocationSearch,
  type ArticleDataType,
  type DataGroup,
  type GeneralDataType,
  type SettingsLocation
} from "./settingsDataModel";
import { SettingsTabList } from "./SettingsTabList";
import { SettingsStorageCard } from "./SettingsStorageCard";
import { SettingsTransferTab } from "./SettingsTransferTab";
import { useSettingsStorage } from "./useSettingsStorage";
import { useSettingsTransfer } from "./useSettingsTransfer";

import {
  applyVisibleMetadata,
  auditActor,
  auditTarget,
  cv8BinaryText,
  cv8CountryText,
  cv8DecimalText,
  cv8HexText,
  cv8NameText,
  emptyForm,
  emptyPasswordForm,
  emptyUserForm,
  entryToForm,
  externalLink,
  formatDateTime,
  loadableMasterDataTypes,
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
  settingsTabs
} from "./settingsModel";
import type { FormState, PasswordFormState, SettingsConfirmDialog, SettingsTab, UserFormState } from "./settingsModel";

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
  const initialSettingsLocation = useMemo(() => readSettingsLocation(window.location.search), []);
  const [activeSettingsTab, setActiveSettingsTab] = useState<SettingsTab>(initialSettingsLocation.tab);
  const [dataGroup, setDataGroup] = useState<DataGroup>(initialSettingsLocation.group);
  const [activeArticleType, setActiveArticleType] = useState<ArticleDataType>(
    isArticleDataType(initialSettingsLocation.type) ? initialSettingsLocation.type : "stock_unit"
  );
  const [activeType, setActiveType] = useState<GeneralDataType>(
    isGeneralDataType(initialSettingsLocation.type) ? initialSettingsLocation.type : "manufacturer"
  );
  useEffect(() => {
    const normalizedSearch = settingsLocationSearch(initialSettingsLocation);
    if (normalizedSearch !== window.location.search) {
      window.history.replaceState(null, "", `${window.location.pathname}${normalizedSearch}`);
    }
  }, [initialSettingsLocation]);
  const writeSettingsLocation = (location: SettingsLocation) => {
    const search = settingsLocationSearch(location);
    window.history.replaceState(null, "", `${window.location.pathname}${search}`);
  };
  const selectSettingsTab = (tab: SettingsTab) => {
    writeSettingsLocation({
      tab,
      group: dataGroup,
      type: dataGroup === "general" ? activeType : activeArticleType
    });
    setActiveSettingsTab(tab);
  };
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
  const storage = useSettingsStorage();
  const { storageUsage, storageLoading, loadStorageUsage } = storage;
  const preferences = useSettingsPreferences(username);
  const general = useSettingsGeneral(activeSettingsTab, preferences, loadStorageUsage);
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
  // Keep controller state mounted when navigating between settings tabs.
  const transfer = useSettingsTransfer({
    requestConfirmation: setConfirmDialog,
    invalidateMasterData: () => {
      setLoadedTypes({});
      setItemsByType({});
    },
    clearMasterDataSearch: () => setSearch(""),
    loadStorageUsage
  });
  const canManageUsers = Boolean(currentSession?.roles.includes("Admin"));
  const canEditMasterData = Boolean(currentSession?.roles.some((role) => role === "Admin" || role === "Editor"));
  const activeDataType = useMemo(
    () => masterDataTypes.find((item) => item.type === activeType) || masterDataTypes[0],
    [activeType]
  );
  const masterLabel = (type: string) => t(`settings.master.${type}`);
  const masterDescription = (type: string) => t(`settings.master.${type}Desc`);
  const dataGroupOptions = [
    { id: "general" as const, label: t("settings.data.group.general") },
    { id: "article" as const, label: t("settings.data.group.article") }
  ];
  const generalTypeOptions = generalDataTypes.map((type) => ({ id: type, label: masterLabel(type) }));
  const selectDataGroup = (group: DataGroup) => {
    setDataGroup(group);
    writeSettingsLocation({
      tab: "data",
      group,
      type: group === "general" ? activeType : activeArticleType
    });
  };
  const selectGeneralType = (type: GeneralDataType) => {
    setActiveType(type);
    writeSettingsLocation({ tab: "data", group: "general", type });
  };
  const selectArticleType = (type: ArticleDataType) => {
    setActiveArticleType(type);
    writeSettingsLocation({ tab: "data", group: "article", type });
  };
  const auditLabel = (action: string) => {
    const label = t(`settings.auditAction.${action}`);
    return label === `settings.auditAction.${action}` ? action : label;
  };
  const roleDescription = (role: string) => {
    const description = t(`settings.role.${role}`);
    return description === `settings.role.${role}` ? t("settings.roles.custom") : description;
  };
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
    setConfirmDialog({
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
    setConfirmDialog({
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
    setConfirmDialog({
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

  const masterDataColumns: MasterDataTableColumn[] = [];
  if (canEditMasterData) {
    masterDataColumns.push({
      key: "selection",
      label: "",
      align: "center",
      sortable: false,
      render: (entry) => <MasterDataBulkEntryCheckbox
        selection={masterDataSelection}
        entry={entry}
        label={isCV8ManufacturerData ? cv8NameText(entry) : entry.label}
        disabled={masterDataBatchBusy}
      />
    });
  }
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
  masterDataColumns.push(
    {
      key: "status",
      label: t("settings.master.status"),
      align: "center",
      sortValue: (entry) => entry.active ? 1 : 0,
      render: (entry) => t(entry.active ? "settings.articleManagement.active" :
        "settings.articleManagement.inactive")
    },
    {
      key: "origin",
      label: t("settings.master.origin"),
      align: "center",
      sortValue: (entry) => entry.origin || "custom",
      render: (entry) => t(`settings.master.origin.${entry.origin || "custom"}`)
    }
  );
  masterDataColumns.push({
    key: "actions",
    label: t("settings.data.actions"),
    align: "right",
    sortable: false,
    render: (entry) => (
      <MasterDataLifecycleActions
        entry={entry}
        displayLabel={isCV8ManufacturerData ? cv8NameText(entry) : entry.label}
        onEdit={() => startEdit(entry)}
        onDeactivate={() => deactivateEntry(entry)}
        onReactivate={() => setEntryActive(entry, true)}
        onDelete={() => deleteEntry(entry)}
        disabled={masterDataBatchBusy}
      />
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
  const activeSettingsDefinition =
    settingsTabs.find((tab) => tab.id === activeSettingsTab) || settingsTabs[0];

  return (
    <>
      <section className="settings-head">
        <h1>{t(activeSettingsDefinition.labelKey)}</h1>
        <p>{t(activeSettingsDefinition.descriptionKey)}</p>
      </section>

      <nav className="settings-primary-tabs" aria-label={t("settings.title")}>
        {settingsTabs.map((tab) => (
          <button
            key={tab.id}
            type="button"
            className={activeSettingsTab === tab.id ? "active" : ""}
            onClick={() => selectSettingsTab(tab.id)}
          >
            {t(tab.labelKey)}
          </button>
        ))}
      </nav>

      {activeSettingsTab === "general" && (
        <section className="settings-dashboard-grid">
          <div className="settings-card-stack">
            <SettingsProfileCard preferences={preferences} />
            <SettingsInventoryCard general={general} />
          </div>
          <aside className="settings-card-stack">
            <SettingsArticleSearchCard preferences={preferences} />
            <SettingsUpdatesCard preferences={preferences} general={general} />
            <SettingsStorageCard storage={storage} canManageUsers={canManageUsers} />
          </aside>
        </section>
      )}

      {activeSettingsTab === "data" && (
        <section className="panel settings-card data-card">
          <SettingsTabList
            ariaLabel={t("settings.data.groups")}
            options={dataGroupOptions}
            value={dataGroup}
            onChange={selectDataGroup}
            className="settings-data-groups"
          />

          {dataGroup === "general" ? <>
            <SettingsTabList
              ariaLabel={t("settings.data.nav")}
              options={generalTypeOptions}
              value={activeType}
              onChange={selectGeneralType}
              className="settings-data-types"
            />

            {activeType === "decoder_profile" ? <DecoderLibrarySettings roles={currentSession?.roles || []} /> : <section className="master-data-panel">
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

                {canEditMasterData && <MasterDataBulkToolbar
                  count={masterDataSelection.selectedCount}
                  busy={masterDataBatchBusy}
                  onDeactivate={deactivateSelectedEntries}
                />}

                <div className={`table-wrap master-data-table master-data-table-${activeType}`}>
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
                              {column.key === "selection" ? (
                                <MasterDataBulkSelectAllCheckbox
                                  selection={masterDataSelection}
                                  disabled={masterDataBatchBusy}
                                />
                              ) : column.sortable === false || !column.sortValue ? (
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
                          <tr key={entry.id} className={entry.active ? "" : "muted-row"}>
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
            </section>}
          </> : (
            <ArticleManagementSettings
              roles={currentSession?.roles || []}
              activeSection={activeArticleType}
              onSectionChange={selectArticleType}
              onConfirmAction={setConfirmDialog}
            />
          )}
        </section>
      )}

      {activeSettingsTab === "importExport" && (
        <SettingsTransferTab transfer={transfer} storage={storage} />
      )}

      {activeSettingsTab === "digital" && (
        <SettingsDigitalTab canManageUsers={canManageUsers} formatDateTime={formatDateTime} username={username} />
      )}

      {activeSettingsTab === "appearance" && <SettingsAppearanceTab preferences={preferences} />}

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
