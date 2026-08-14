import { type FormEvent, useState } from "react";

import {
  api,
  type ArticleSearchImage,
  type ArticleSearchInput,
  type ArticleSearchResponse,
  type ArticleSearchResult,
  type MasterDataEntry
} from "../../shared/api";
import {
  articleSelectionKey,
  imageSelectionKey
} from "../../shared/articleSearch/articleSearchModel";
import {
  articleSearchEnabled,
  articleSearchSources
} from "../../shared/articleSearch/articleSearchPreferences";
import {
  accessorySearchInput,
  applyAccessorySearchResult,
  currentAccessorySearchValue,
  hasAccessorySearchCriteria,
  isSelectableAccessorySearchValue
} from "./accessoryArticleSearch";
import type { ArticleEditorForm } from "./articleEditorModel";

export type PendingAccessoryArticleImage = ArticleSearchImage & {
  id: string;
  isPrimary: boolean;
};

type UseAccessoryArticleSearchControllerOptions = {
  form: ArticleEditorForm;
  readOnly: boolean;
  pendingImageCount: number;
  manufacturers: MasterDataEntry[];
  gauges: MasterDataEntry[];
  updateForm: (patch: Partial<ArticleEditorForm>) => void;
  addImages: (images: PendingAccessoryArticleImage[]) => void;
};

function isSafeRemoteImage(image: ArticleSearchImage) {
  try {
    const url = new URL(image.url);
    return url.protocol === "http:" || url.protocol === "https:";
  } catch {
    return false;
  }
}

export function useAccessoryArticleSearchController({
  form,
  readOnly,
  pendingImageCount,
  manufacturers,
  gauges,
  updateForm,
  addImages
}: UseAccessoryArticleSearchControllerOptions) {
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const [response, setResponse] = useState<ArticleSearchResponse | null>(null);
  const [error, setError] = useState("");
  const [barcodeOpen, setBarcodeOpen] = useState(false);
  const [barcodeValue, setBarcodeValue] = useState("");
  const [selectedFields, setSelectedFields] = useState<Record<string, boolean>>({});
  const [selectedImages, setSelectedImages] = useState<Record<string, boolean>>({});

  const canSelectField = (key: string, value: string) =>
    isSelectableAccessorySearchValue(key, value, manufacturers, gauges, form.articleType);

  const run = (searchForm = form, searchInput?: ArticleSearchInput) => {
    if (readOnly) return;
    if (!articleSearchEnabled()) {
      setError("Die Artikeldaten-Websuche ist in den Einstellungen deaktiviert.");
      setOpen(true);
      setResponse(null);
      return;
    }
    const input = searchInput ?? accessorySearchInput(searchForm);
    if (!hasAccessorySearchCriteria(input)) {
      setError("Bitte Artikel-Nr. oder Bezeichnung sowie Hersteller und Spurweite erfassen.");
      return;
    }

    setOpen(true);
    setLoading(true);
    setError("");
    setResponse(null);
    setSelectedFields({});
    setSelectedImages({});
    void api.articleSearch(input).then((nextResponse) => {
      const initialSelection: Record<string, boolean> = {};
      nextResponse.results.forEach((result, index) => {
        Object.entries(result.fields).forEach(([key, field]) => {
          if (!canSelectField(key, field.value)) return;
          initialSelection[articleSelectionKey(result, key, index)] =
            currentAccessorySearchValue(searchForm, key) === "";
        });
      });
      setResponse(nextResponse);
      setSelectedFields(initialSelection);
    }).catch((reason: Error) => setError(reason.message))
      .finally(() => setLoading(false));
  };

  const openBarcode = () => {
    if (readOnly) return;
    setBarcodeValue(form.ean);
    setBarcodeOpen(true);
  };

  const submitBarcode = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (readOnly) return;
    const code = barcodeValue.trim();
    if (!code) return;
    const nextForm = { ...form, ean: code };
    updateForm({ ean: code });
    setBarcodeOpen(false);
    run(nextForm, {
      searchSources: articleSearchSources(),
      fields: { ean: code }
    });
  };

  const toggleField = (result: ArticleSearchResult, index: number, key: string, checked: boolean) => {
    const value = result.fields[key]?.value || "";
    if (!canSelectField(key, value)) return;
    setSelectedFields((current) => ({ ...current, [articleSelectionKey(result, key, index)]: checked }));
  };

  const toggleImage = (
    result: ArticleSearchResult,
    index: number,
    image: ArticleSearchImage,
    checked: boolean
  ) => {
    if (!isSafeRemoteImage(image)) return;
    setSelectedImages((current) => ({ ...current, [imageSelectionKey(result, image, index)]: checked }));
  };

  const applyResult = (result: ArticleSearchResult) => {
    if (readOnly) return;
    const foundIndex = response?.results.findIndex((entry) => entry.url === result.url) ?? 0;
    const resultIndex = foundIndex >= 0 ? foundIndex : 0;
    const patch = applyAccessorySearchResult({
      form,
      result,
      resultIndex,
      selectedFields,
      manufacturers,
      gauges
    });
    const images = (result.images || []).filter((image) =>
      isSafeRemoteImage(image) && selectedImages[imageSelectionKey(result, image, resultIndex)])
      .map((image, index) => ({
        ...image,
        id: `${result.url}-${image.url}`,
        isPrimary: pendingImageCount === 0 && index === 0
      }));
    if (images.length > 0) addImages(images);
    updateForm(patch);
    setOpen(false);
  };

  return {
    state: {
      open,
      loading,
      canRun: hasAccessorySearchCriteria(accessorySearchInput(form)),
      response,
      error,
      barcodeOpen,
      barcodeValue,
      selectedFields,
      selectedImages
    },
    setters: { setOpen, setBarcodeOpen, setBarcodeValue },
    commands: { run, openBarcode, submitBarcode, toggleField, toggleImage, applyResult, canSelectField }
  };
}

export type AccessoryArticleSearchController = ReturnType<typeof useAccessoryArticleSearchController>;
