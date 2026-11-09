import { useState } from "react";
import { DigitalCenterSettings } from "../../shared/api";
import { type DigitalProvider } from "./DigitalCenterWorkflowView";
import { localSettingKeys, readLocalBool, readLocalSetting } from "./settingsModel";

export const digitalProviders: DigitalProvider[] = ["ecos", "z21", "intellibox3", "cs3"];

export function useDigitalProviderSettings() {
  const [provider, setProvider] = useState<DigitalProvider>(() => {
    const stored = readLocalSetting(localSettingKeys.digitalProvider, "ecos");
    return digitalProviders.includes(stored as DigitalProvider) ? stored as DigitalProvider : "ecos";
  });

  const [activeDialogProvider, setActiveDialogProvider] = useState<DigitalProvider | null>(null);

  const [ecosEnabled, setEcosEnabled] = useState(() => {
    const stored = window.localStorage.getItem(localSettingKeys.digitalEcosEnabled);
    if (stored !== null) return readLocalBool(localSettingKeys.digitalEcosEnabled, false);
    return Boolean(readLocalSetting(localSettingKeys.digitalEcosHost, "").trim());
  });

  const [ecosHost, setEcosHost] = useState(() => readLocalSetting(localSettingKeys.digitalEcosHost, ""));

  const [ecosPort, setEcosPort] = useState(() => readLocalSetting(localSettingKeys.digitalEcosPort, "15471"));

  const [z21Enabled, setZ21Enabled] = useState(() => readLocalBool(localSettingKeys.digitalZ21Enabled, false));

  const [z21Host, setZ21Host] = useState(() => readLocalSetting(localSettingKeys.digitalZ21Host, ""));

  const [z21Port, setZ21Port] = useState(() => readLocalSetting(localSettingKeys.digitalZ21Port, "21105"));

  const [intellibox3Enabled, setIntellibox3Enabled] = useState(() => readLocalBool(localSettingKeys.digitalIntellibox3Enabled, false));

  const [intellibox3Host, setIntellibox3Host] = useState(() => readLocalSetting(localSettingKeys.digitalIntellibox3Host, ""));

  const [intellibox3Port, setIntellibox3Port] = useState(() => readLocalSetting(localSettingKeys.digitalIntellibox3Port, "21105"));

  const [cs3Enabled, setCS3Enabled] = useState(() => readLocalBool(localSettingKeys.digitalCS3Enabled, false));

  const [cs3Host, setCS3Host] = useState(() => readLocalSetting(localSettingKeys.digitalCS3Host, ""));

  const [cs3Port, setCS3Port] = useState(() => readLocalSetting(localSettingKeys.digitalCS3Port, "80"));

  const ecosInput = () => ({
    host: ecosHost.trim(),
    port: Number(ecosPort) || 15471
  });

  const providerEnabled = (providerId: DigitalProvider) => {
    if (providerId === "ecos") return ecosEnabled;
    if (providerId === "z21") return z21Enabled;
    if (providerId === "intellibox3") return intellibox3Enabled;
    return cs3Enabled;
  };

  const providerHost = (providerId: DigitalProvider) => {
    if (providerId === "ecos") return ecosHost.trim();
    if (providerId === "z21") return z21Host.trim();
    if (providerId === "intellibox3") return intellibox3Host.trim();
    return cs3Host.trim();
  };

  const providerPort = (providerId: DigitalProvider) => {
    if (providerId === "ecos") return ecosPort.trim() || "15471";
    if (providerId === "z21") return z21Port.trim() || "21105";
    if (providerId === "intellibox3") return intellibox3Port.trim() || "21105";
    return cs3Port.trim() || "80";
  };

  const providerPortDraft = (providerId: DigitalProvider) => ({
    ecos: ecosPort, z21: z21Port, intellibox3: intellibox3Port, cs3: cs3Port
  })[providerId];

  const providerInput = (providerId: DigitalProvider) => ({
    host: providerHost(providerId),
    port: Number(providerPort(providerId))
  });

  const providerConfigured = (providerId: DigitalProvider) => Boolean(providerHost(providerId));

  const providerActive = (providerId: DigitalProvider) => providerEnabled(providerId) && providerConfigured(providerId);

  const currentDigitalSettings = (nextProvider: DigitalProvider = provider): DigitalCenterSettings => ({
    provider: nextProvider,
    ecos: { enabled: ecosEnabled, host: ecosHost.trim(), port: ecosPort.trim() || "15471" },
    z21: { enabled: z21Enabled, host: z21Host.trim(), port: z21Port.trim() || "21105" },
    intellibox3: { enabled: intellibox3Enabled, host: intellibox3Host.trim(), port: intellibox3Port.trim() || "21105" },
    cs3: { enabled: cs3Enabled, host: cs3Host.trim(), port: cs3Port.trim() || "80" }
  });

  const applyDigitalSettings = (settings: DigitalCenterSettings) => {
    const nextProvider = digitalProviders.includes(settings.provider) ? settings.provider : "ecos";
    setProvider(nextProvider);
    setEcosEnabled(Boolean(settings.ecos?.enabled));
    setEcosHost(settings.ecos?.host || "");
    setEcosPort(settings.ecos?.port || "15471");
    setZ21Enabled(Boolean(settings.z21?.enabled));
    setZ21Host(settings.z21?.host || "");
    setZ21Port(settings.z21?.port || "21105");
    setIntellibox3Enabled(Boolean(settings.intellibox3?.enabled));
    setIntellibox3Host(settings.intellibox3?.host || "");
    setIntellibox3Port(settings.intellibox3?.port || "21105");
    setCS3Enabled(Boolean(settings.cs3?.enabled));
    setCS3Host(settings.cs3?.host || "");
    setCS3Port(settings.cs3?.port || "80");
  };

  const hasDigitalConfig = (settings: DigitalCenterSettings) =>
    Boolean(
      settings.ecos?.host ||
      settings.z21?.host ||
      settings.intellibox3?.host ||
      settings.cs3?.host ||
      settings.ecos?.enabled ||
      settings.z21?.enabled ||
      settings.intellibox3?.enabled ||
      settings.cs3?.enabled
    );

  const setActiveProviderHost = (host: string) => {
    if (activeDialogProvider === "ecos") setEcosHost(host);
    if (activeDialogProvider === "z21") setZ21Host(host);
    if (activeDialogProvider === "intellibox3") setIntellibox3Host(host);
    if (activeDialogProvider === "cs3") setCS3Host(host);
  };

  const setActiveProviderPort = (port: string) => {
    if (activeDialogProvider === "ecos") setEcosPort(port);
    if (activeDialogProvider === "z21") setZ21Port(port);
    if (activeDialogProvider === "intellibox3") setIntellibox3Port(port);
    if (activeDialogProvider === "cs3") setCS3Port(port);
  };

  const storeSettingsLocally = (settings: DigitalCenterSettings) => {
    const nextProvider = settings.provider;
    window.localStorage.setItem(localSettingKeys.digitalProvider, nextProvider);
    window.localStorage.setItem(localSettingKeys.digitalEcosEnabled, String(settings.ecos.enabled));
    window.localStorage.setItem(localSettingKeys.digitalEcosHost, settings.ecos.host);
    window.localStorage.setItem(localSettingKeys.digitalEcosPort, settings.ecos.port);
    window.localStorage.setItem(localSettingKeys.digitalZ21Enabled, String(settings.z21.enabled));
    window.localStorage.setItem(localSettingKeys.digitalZ21Host, settings.z21.host);
    window.localStorage.setItem(localSettingKeys.digitalZ21Port, settings.z21.port);
    window.localStorage.setItem(localSettingKeys.digitalIntellibox3Enabled, String(settings.intellibox3.enabled));
    window.localStorage.setItem(localSettingKeys.digitalIntellibox3Host, settings.intellibox3.host);
    window.localStorage.setItem(localSettingKeys.digitalIntellibox3Port, settings.intellibox3.port);
    window.localStorage.setItem(localSettingKeys.digitalCS3Enabled, String(settings.cs3.enabled));
    window.localStorage.setItem(localSettingKeys.digitalCS3Host, settings.cs3.host);
    window.localStorage.setItem(localSettingKeys.digitalCS3Port, settings.cs3.port);
    window.dispatchEvent(new Event("railkeeper-digital-settings-changed"));
  };

  return {
    provider, setProvider, activeDialogProvider, setActiveDialogProvider, ecosEnabled,
    setEcosEnabled, ecosHost, setEcosHost, ecosPort, setEcosPort, z21Enabled, setZ21Enabled,
    z21Host, setZ21Host, z21Port, setZ21Port, intellibox3Enabled, setIntellibox3Enabled,
    intellibox3Host, setIntellibox3Host, intellibox3Port, setIntellibox3Port, cs3Enabled,
    setCS3Enabled, cs3Host, setCS3Host, cs3Port, setCS3Port, ecosInput, providerEnabled,
    providerHost, providerPort, providerPortDraft, providerInput, providerConfigured,
    providerActive, currentDigitalSettings, applyDigitalSettings, hasDigitalConfig,
    setActiveProviderHost, setActiveProviderPort, storeSettingsLocally,
  };
}
