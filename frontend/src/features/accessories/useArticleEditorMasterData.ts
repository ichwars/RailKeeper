import type { Dispatch, RefObject, SetStateAction } from "react";
import { useCallback } from "react";
import {
  api,
  type AccessoryArticleType,
  type MasterDataEntry
} from "../../shared/api";
import {
  type ArticleEditorForm
} from "./articleEditorModel";
import {
  customFieldDefinitions,
  type CustomArticleSubjectFieldDefinition
} from "./articleTypeFields";
import { articleTypeOrder } from "./articleTypes";



type useArticleEditorMasterDataProps = {
  setCustomFieldsLoading: Dispatch<SetStateAction<boolean>>;
  setCustomFieldsError: Dispatch<SetStateAction<string>>;
  isCurrent: (generation: number) => boolean;
  setCustomFields: Dispatch<SetStateAction<CustomArticleSubjectFieldDefinition[]>>;
  t: (key: string, values?: Record<string, string | number>) => string;
  setSubtypeEntriesLoading: Dispatch<SetStateAction<boolean>>;
  setSubtypeEntriesError: Dispatch<SetStateAction<string>>;
  setSubtypeEntries: Dispatch<SetStateAction<MasterDataEntry[]>>;
  setArticleTypeEntriesLoading: Dispatch<SetStateAction<boolean>>;
  setArticleTypeEntriesError: Dispatch<SetStateAction<string>>;
  setArticleTypeEntries: Dispatch<SetStateAction<MasterDataEntry[]>>;
  createGenerationRef: RefObject<number | null>;
  setForm: Dispatch<SetStateAction<ArticleEditorForm>>;
  setInitialForm: Dispatch<SetStateAction<ArticleEditorForm>>;
};

export function useArticleEditorMasterData({
  setCustomFieldsLoading, setCustomFieldsError, isCurrent, setCustomFields, t, setSubtypeEntriesLoading, setSubtypeEntriesError, setSubtypeEntries, setArticleTypeEntriesLoading, setArticleTypeEntriesError, setArticleTypeEntries, createGenerationRef, setForm, setInitialForm
}: useArticleEditorMasterDataProps) {
  const loadCustomFields = useCallback(async (generation: number) => {
    setCustomFieldsLoading(true);
    setCustomFieldsError("");
    try {
      const entries = await api.masterData("accessory_custom_field", true);
      if (isCurrent(generation)) setCustomFields(customFieldDefinitions(entries));
    } catch {
      if (isCurrent(generation)) setCustomFieldsError(t("accessories.subject.customLoadError"));
    } finally {
      if (isCurrent(generation)) setCustomFieldsLoading(false);
    }
  }, [isCurrent, setCustomFields, setCustomFieldsError, setCustomFieldsLoading, t]);

  const loadSubtypeEntries = useCallback(async (generation: number) => {
    setSubtypeEntriesLoading(true);
    setSubtypeEntriesError("");
    try {
      const entries = await api.masterData("accessory_subtype");
      if (isCurrent(generation)) setSubtypeEntries(entries);
    } catch {
      if (isCurrent(generation)) setSubtypeEntriesError(t("accessories.editor.subtypes.loadError"));
    } finally {
      if (isCurrent(generation)) setSubtypeEntriesLoading(false);
    }
  }, [isCurrent, setSubtypeEntries, setSubtypeEntriesError, setSubtypeEntriesLoading, t]);

  const loadArticleTypeEntries = useCallback(async (generation: number) => {
    setArticleTypeEntriesLoading(true);
    setArticleTypeEntriesError("");
    try {
      const entries = await api.masterData("article_type");
      if (!isCurrent(generation)) return;
      setArticleTypeEntries(entries);
      if (createGenerationRef.current === generation) {
        const activeEntry = entries.find((entry) => entry.type === "article_type" && entry.active &&
          articleTypeOrder.some((articleType) => articleType === entry.key));
        if (!activeEntry) {
          setArticleTypeEntriesError(t("accessories.editor.articleTypes.noneActive"));
          return;
        }
        const articleType = activeEntry.key as AccessoryArticleType;
        setForm((current) => ({ ...current, articleType }));
        setInitialForm((current) => ({ ...current, articleType }));
      }
    } catch {
      if (isCurrent(generation)) setArticleTypeEntriesError(t("accessories.editor.articleTypes.loadError"));
    } finally {
      if (isCurrent(generation)) setArticleTypeEntriesLoading(false);
    }
  }, [createGenerationRef, isCurrent, setArticleTypeEntries, setArticleTypeEntriesError, setArticleTypeEntriesLoading, setForm, setInitialForm, t]);

  return {
    loadCustomFields,
    loadArticleTypeEntries,
    loadSubtypeEntries,
  };
}
