export type ThemePreference = "system" | "light" | "dark";

export const themePreferenceKey = "railkeeper.theme";
const themeOptionKeys = {
  darkBackground: "railkeeper.settings.darkBackground",
  darkAccent: "railkeeper.settings.darkAccent",
  darkStyle: "railkeeper.settings.darkStyle",
  lightBackground: "railkeeper.settings.lightBackground",
  lightAccent: "railkeeper.settings.lightAccent",
  lightStyle: "railkeeper.settings.lightStyle"
};

export function readThemePreference(): ThemePreference {
  const stored = window.localStorage.getItem(themePreferenceKey);
  return stored === "light" || stored === "dark" || stored === "system" ? stored : "system";
}

export function applyThemePreference(preference: ThemePreference) {
  window.localStorage.setItem(themePreferenceKey, preference);
  document.documentElement.dataset.theme = preference;
  applyStoredThemeOptions();
}

export function applyStoredThemeOptions() {
  const root = document.documentElement;
  root.dataset.darkBackground = window.localStorage.getItem(themeOptionKeys.darkBackground) || "neutral";
  root.dataset.darkAccent = window.localStorage.getItem(themeOptionKeys.darkAccent) || "green";
  root.dataset.darkStyle = window.localStorage.getItem(themeOptionKeys.darkStyle) || "classic";
  root.dataset.lightBackground = window.localStorage.getItem(themeOptionKeys.lightBackground) || "neutral";
  root.dataset.lightAccent = window.localStorage.getItem(themeOptionKeys.lightAccent) || "green";
  root.dataset.lightStyle = window.localStorage.getItem(themeOptionKeys.lightStyle) || "classic";
}
