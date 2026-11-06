import { useI18n } from "../../shared/i18n";
import { ArticleManagementSettings } from "./ArticleManagementSettings";
import { DecoderLibrarySettings } from "./DecoderLibrarySettings";
import { SettingsMasterDataPanel } from "./SettingsMasterDataPanel";
import { SettingsTabList } from "./SettingsTabList";
import { generalDataTypes } from "./settingsDataModel";
import type { SettingsConfirmDialog } from "./settingsModel";
import type { SettingsNavigationController } from "./useSettingsNavigation";
import type { SettingsMasterDataController } from "./useSettingsMasterData";

type SettingsDataTabProps = {
  navigation: SettingsNavigationController;
  masterData: SettingsMasterDataController;
  roles: string[];
  requestConfirmation: (dialog: SettingsConfirmDialog) => void;
};

export function SettingsDataTab({ navigation, masterData, roles, requestConfirmation }: SettingsDataTabProps) {
  const { t } = useI18n();
  const { dataGroup, activeType, activeArticleType, selectDataGroup, selectGeneralType, selectArticleType } = navigation;
  const masterLabel = (type: string) => t(`settings.master.${type}`);
  const dataGroupOptions = [
    { id: "general" as const, label: t("settings.data.group.general") },
    { id: "article" as const, label: t("settings.data.group.article") }
  ];
  const generalTypeOptions = generalDataTypes.map((type) => ({ id: type, label: masterLabel(type) }));

  return (
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

        {activeType === "decoder_profile" ? <DecoderLibrarySettings roles={roles} /> : <SettingsMasterDataPanel controller={masterData} />}
      </> : (
        <ArticleManagementSettings
          roles={roles}
          activeSection={activeArticleType}
          onSectionChange={selectArticleType}
          onConfirmAction={requestConfirmation}
        />
      )}
    </section>
  );
}
