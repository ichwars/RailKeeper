import { useI18n } from "../../shared/i18n";

import { AppSelect } from "../../shared/ui/AppSelect";
import { localSettingKeys } from "./settingsModel";
import type { SettingsPreferencesController } from "./useSettingsPreferences";

export function SettingsAppearanceTab({ preferences }: { preferences: SettingsPreferencesController }) {
  const { t } = useI18n();
  const {
    design,
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
    setLocalSetting,
    updateDesign
  } = preferences;

  return (
<section className="panel settings-card settings-tool-card">
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
  );
}
