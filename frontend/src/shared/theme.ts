export type ThemePreference = "system" | "light" | "dark";

export const themePreferenceKey = "railkeeper.theme";

export function readThemePreference(): ThemePreference {
  const stored = window.localStorage.getItem(themePreferenceKey);
  return stored === "light" || stored === "dark" || stored === "system" ? stored : "system";
}

export function applyThemePreference(preference: ThemePreference) {
  window.localStorage.setItem(themePreferenceKey, preference);
  document.documentElement.dataset.theme = preference;
}
