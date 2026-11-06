import { useI18n } from "../../shared/i18n";
import type { Language } from "../../shared/i18n";
import { Eye, EyeOff, ChevronUp, ChevronDown } from "lucide-react";
import { availableStartView } from "../../app/navigationAvailability";
import { AppSelect } from "../../shared/ui/AppSelect";
import { localSettingKeys } from "./settingsModel";
import type { SettingsPreferencesController } from "./useSettingsPreferences";

export function SettingsProfileCard({ preferences }: { preferences: SettingsPreferencesController }) {
  const { language, setLanguage, t } = useI18n();
  const {
    defaultView,
    setDefaultView,
    sidebarOrder,
    sidebarHidden,
    dateFormat,
    setDateFormat,
    timeFormat,
    setTimeFormat,
    defaultPrinter,
    setDefaultPrinter,
    systemPrinters,
    setLocalSetting,
    moveSidebarItem,
    resetSidebarOrder,
    toggleSidebarVisibility
  } = preferences;

  return (
    <section className="panel settings-card settings-tool-card">
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
          <AppSelect
            value={defaultView}
            onChange={(event) => setLocalSetting(
              localSettingKeys.defaultView,
              event.target.value,
              (value) => setDefaultView(availableStartView(value))
            )}
          >
            <option value="overview">{t("nav.overview")}</option>
            <option value="vehicles">{t("nav.vehicles")}</option>
            <option value="accessories">{t("nav.accessories")}</option>
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
  );
}
