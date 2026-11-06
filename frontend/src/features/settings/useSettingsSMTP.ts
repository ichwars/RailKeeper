import { useState, type FormEvent } from "react";
import { api, type SMTPSettings, type SMTPSettingsInput } from "../../shared/api";
import { useI18n } from "../../shared/i18n";

export function useSettingsSMTP() {
  const { t } = useI18n();
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

  return {
    smtpSettings,
    setSmtpSettings,
    smtpForm,
    setSmtpForm,
    smtpLoading,
    setSmtpLoading,
    smtpSaving,
    setSmtpSaving,
    smtpTesting,
    setSmtpTesting,
    smtpMessage,
    setSmtpMessage,
    loadSMTPSettings,
    saveSMTPSettings,
    testSMTPSettings
  };
}
