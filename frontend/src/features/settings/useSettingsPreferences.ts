import { useEffect, useState } from "react";
import type { AppView } from "../../app/App";
import { availableStartView } from "../../app/navigationAvailability";
import { api, type SystemPrinters } from "../../shared/api";
import {
  applyStoredThemeOptions, applyThemePreference, readThemePreference, themePreferenceKey,
  type ThemePreference
} from "../../shared/theme";
import {
  articleSearchSettingKey, articleSearchSourcesSettingKey, defaultSidebarOrder, localSettingKeys,
  normalizeSidebarPrefs, readArticleSearchSources, readLocalBool, readLocalSetting,
  readSidebarPrefs, sidebarOrderChangedEvent, sidebarPrefsKey
} from "./settingsModel";

export function useSettingsPreferences(username: string) {
  const [articleSearchEnabled, setArticleSearchEnabled] = useState(
    () => window.localStorage.getItem(articleSearchSettingKey) !== "false"
  );
  const [articleSearchSources, setArticleSearchSources] = useState<string[]>(readArticleSearchSources);
  const [design, setDesign] = useState<ThemePreference>(readThemePreference);
  const [defaultView, setDefaultView] = useState(() =>
    availableStartView(readLocalSetting(localSettingKeys.defaultView, "overview"))
  );
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
  const [systemPrinters, setSystemPrinters] = useState<SystemPrinters | null>(null);

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
        store(localSettingKeys.defaultView, (value) => setDefaultView(availableStartView(value)));
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
            const normalized = normalizeSidebarPrefs(prefs);
            window.localStorage.setItem(sidebarPrefsKey(username), JSON.stringify(normalized));
            setSidebarOrder(normalized.order);
            setSidebarHidden(normalized.hidden);
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

  const updateBetaUpdates = (enabled: boolean) => {
    setLocalBool(localSettingKeys.betaUpdates, enabled, setBetaUpdates);
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

  return {
    articleSearchEnabled,
    setArticleSearchEnabled,
    articleSearchSources,
    setArticleSearchSources,
    design,
    setDesign,
    defaultView,
    setDefaultView,
    sidebarOrder,
    setSidebarOrder,
    sidebarHidden,
    setSidebarHidden,
    dateFormat,
    setDateFormat,
    timeFormat,
    setTimeFormat,
    defaultPrinter,
    setDefaultPrinter,
    betaUpdates,
    setBetaUpdates,
    darkBackground,
    setDarkBackground,
    darkAccent,
    setDarkAccent,
    darkStyle,
    setDarkStyle,
    lightBackground,
    setLightBackground,
    lightAccent,
    setLightAccent,
    lightStyle,
    setLightStyle,
    systemPrinters,
    setSystemPrinters,
    saveProfileSetting,
    setLocalSetting,
    setLocalBool,
    saveSidebarPrefs,
    saveSidebarOrder,
    moveSidebarItem,
    resetSidebarOrder,
    toggleSidebarVisibility,
    updateArticleSearchEnabled,
    updateArticleSearchSource,
    updateDesign,
    updateBetaUpdates,
    loadSystemPrinters
  };
}

export type SettingsPreferencesController = ReturnType<typeof useSettingsPreferences>;
