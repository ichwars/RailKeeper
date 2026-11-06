import { useEffect, useState, type FormEvent } from "react";
import {
  api, type AuditLogEntry, type Role, type Session, type SessionRecord, type UserAccount
} from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import {
  emptyPasswordForm, emptyUserForm, type PasswordFormState, type UserFormState,
  type SettingsTab, type SettingsConfirmDialog
} from "./settingsModel";
import { useSettingsSMTP } from "./useSettingsSMTP";
import { useSettingsTwoFactor } from "./useSettingsTwoFactor";

type SettingsAuthOptions = {
  username: string;
  activeSettingsTab: SettingsTab;
  requestConfirmation: (dialog: SettingsConfirmDialog) => void;
};

export function useSettingsAuth({ username, activeSettingsTab, requestConfirmation }: SettingsAuthOptions) {
  const { t } = useI18n();
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
  const canManageUsers = Boolean(currentSession?.roles.includes("Admin"));
  const canEditMasterData = Boolean(currentSession?.roles.some((role) => role === "Admin" || role === "Editor"));

  const loadCurrentSession = () => {
    setAuthMessage("");
    api
      .session()
      .then(setCurrentSession)
      .catch((error: Error) => setAuthMessage(error.message));
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

  const revokeSession = (session: SessionRecord) => {
    requestConfirmation({
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
    requestConfirmation({
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

  const smtp = useSettingsSMTP();
  const { loadSMTPSettings } = smtp;
  const twoFactor = useSettingsTwoFactor({
    canManageUsers, loadCurrentSession, loadUsersAndRoles, loadSessions, loadAuditLog
  });
  const { loadTwoFactorStatus } = twoFactor;

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

  return {
    currentSession,
    setCurrentSession,
    authMessage,
    setAuthMessage,
    availableRoles,
    setAvailableRoles,
    users,
    setUsers,
    usersLoading,
    setUsersLoading,
    auditLog,
    setAuditLog,
    auditLogLoading,
    setAuditLogLoading,
    auditLogMessage,
    setAuditLogMessage,
    sessions,
    setSessions,
    sessionsLoading,
    setSessionsLoading,
    sessionsMessage,
    setSessionsMessage,
    editingUser,
    setEditingUser,
    userForm,
    setUserForm,
    userSaving,
    setUserSaving,
    passwordForm,
    setPasswordForm,
    passwordSaving,
    setPasswordSaving,
    passwordMessage,
    setPasswordMessage,
    canManageUsers,
    canEditMasterData,
    loadCurrentSession,
    loadUsersAndRoles,
    loadAuditLog,
    loadSessions,
    revokeSession,
    changePassword,
    startUserCreate,
    startUserEdit,
    toggleUserRole,
    saveUser,
    deleteUser,
    ...smtp,
    ...twoFactor
  };
}

export type SettingsAuthController = ReturnType<typeof useSettingsAuth>;
