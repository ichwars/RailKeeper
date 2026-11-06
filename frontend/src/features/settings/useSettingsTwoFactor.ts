import { useState } from "react";
import { api, type TwoFactorStatus } from "../../shared/api";
import { useI18n } from "../../shared/i18n";

type TwoFactorOptions = {
  canManageUsers: boolean;
  loadCurrentSession: () => void;
  loadUsersAndRoles: () => void;
  loadSessions: () => void;
  loadAuditLog: () => void;
};

export function useSettingsTwoFactor({ canManageUsers, loadCurrentSession, loadUsersAndRoles, loadSessions, loadAuditLog }: TwoFactorOptions) {
  const { t } = useI18n();
  const [twoFactorStatus, setTwoFactorStatus] = useState<TwoFactorStatus | null>(null);
  const [twoFactorCode, setTwoFactorCode] = useState("");
  const [twoFactorPassword, setTwoFactorPassword] = useState("");
  const [twoFactorLoading, setTwoFactorLoading] = useState(false);
  const [twoFactorSaving, setTwoFactorSaving] = useState(false);
  const [twoFactorMessage, setTwoFactorMessage] = useState("");

  const loadTwoFactorStatus = () => {
    setTwoFactorLoading(true);
    setTwoFactorMessage("");
    api
      .twoFactorStatus()
      .then(setTwoFactorStatus)
      .catch((error: Error) => setTwoFactorMessage(error.message))
      .finally(() => setTwoFactorLoading(false));
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

  return {
    twoFactorStatus,
    setTwoFactorStatus,
    twoFactorCode,
    setTwoFactorCode,
    twoFactorPassword,
    setTwoFactorPassword,
    twoFactorLoading,
    setTwoFactorLoading,
    twoFactorSaving,
    setTwoFactorSaving,
    twoFactorMessage,
    setTwoFactorMessage,
    loadTwoFactorStatus,
    setupTwoFactor,
    enableTwoFactor,
    disableTwoFactor
  };
}
