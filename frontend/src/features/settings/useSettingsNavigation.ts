import { useEffect, useMemo, useState } from "react";
import {
  isArticleDataType,
  isGeneralDataType,
  readSettingsLocation,
  settingsLocationSearch,
  type ArticleDataType,
  type DataGroup,
  type GeneralDataType,
  type SettingsLocation
} from "./settingsDataModel";
import type { SettingsTab } from "./settingsModel";

export function useSettingsNavigation() {
  const initialSettingsLocation = useMemo(() => readSettingsLocation(window.location.search), []);
  const [activeSettingsTab, setActiveSettingsTab] = useState<SettingsTab>(initialSettingsLocation.tab);
  const [dataGroup, setDataGroup] = useState<DataGroup>(initialSettingsLocation.group);
  const [activeArticleType, setActiveArticleType] = useState<ArticleDataType>(
    isArticleDataType(initialSettingsLocation.type) ? initialSettingsLocation.type : "stock_unit"
  );
  const [activeType, setActiveType] = useState<GeneralDataType>(
    isGeneralDataType(initialSettingsLocation.type) ? initialSettingsLocation.type : "manufacturer"
  );
  useEffect(() => {
    const normalizedSearch = settingsLocationSearch(initialSettingsLocation);
    if (normalizedSearch !== window.location.search) {
      window.history.replaceState(null, "", `${window.location.pathname}${normalizedSearch}`);
    }
  }, [initialSettingsLocation]);
  const writeSettingsLocation = (location: SettingsLocation) => {
    const search = settingsLocationSearch(location);
    window.history.replaceState(null, "", `${window.location.pathname}${search}`);
  };
  const selectSettingsTab = (tab: SettingsTab) => {
    writeSettingsLocation({
      tab,
      group: dataGroup,
      type: dataGroup === "general" ? activeType : activeArticleType
    });
    setActiveSettingsTab(tab);
  };
  const selectDataGroup = (group: DataGroup) => {
    setDataGroup(group);
    writeSettingsLocation({
      tab: "data",
      group,
      type: group === "general" ? activeType : activeArticleType
    });
  };
  const selectGeneralType = (type: GeneralDataType) => {
    setActiveType(type);
    writeSettingsLocation({ tab: "data", group: "general", type });
  };
  const selectArticleType = (type: ArticleDataType) => {
    setActiveArticleType(type);
    writeSettingsLocation({ tab: "data", group: "article", type });
  };
  return {
    activeSettingsTab, dataGroup, activeArticleType, activeType,
    selectSettingsTab, selectDataGroup, selectGeneralType, selectArticleType
  };
}

export type SettingsNavigationController = ReturnType<typeof useSettingsNavigation>;
