import { type FormEvent, useState } from "react";

import {
  api,
  type ArticleSearchImage,
  type ArticleSearchInput,
  type ArticleSearchResponse,
  type ArticleSearchResult,
  type CreateVehicleRequest
} from "../../shared/api";
import {
  articleSelectionKey,
  articleValueForForm,
  currentArticleValue,
  imageSelectionKey,
  isArticleFieldKey
} from "./articleSearch";
import type { PendingArticleImage } from "./vehicleTransforms";
import type { VehicleCreateArticleDraft } from "./vehicleCreateWizardState";
import {
  articleSearchEnabled,
  articleSearchSources,
  hasArticleSearchCriteria,
  isBadArticleValue,
  sanitizeArticleSearchResponse,
  vehicleFieldsForSearch
} from "./vehicleViewModel";

type Translate = (key: string, values?: Record<string, string | number>) => string;

type UseArticleSearchControllerOptions = {
  form: CreateVehicleRequest;
  pendingImageCount: number;
  replaceForm: (form: CreateVehicleRequest) => void;
  updateForm: (patch: Partial<CreateVehicleRequest>) => void;
  addImages: (images: PendingArticleImage[]) => void;
  onMessage: (message: string) => void;
  t: Translate;
};

export function useArticleSearchController({
  form,
  pendingImageCount,
  replaceForm,
  updateForm,
  addImages,
  onMessage,
  t
}: UseArticleSearchControllerOptions) {
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const [response, setResponse] = useState<ArticleSearchResponse | null>(null);
  const [error, setError] = useState("");
  const [barcodeOpen, setBarcodeOpen] = useState(false);
  const [barcodeValue, setBarcodeValue] = useState("");
  const [selectedFields, setSelectedFields] = useState<Record<string, boolean>>({});
  const [selectedImages, setSelectedImages] = useState<Record<string, boolean>>({});

  const run = (searchForm = form, searchInput?: ArticleSearchInput) => {
    if (!articleSearchEnabled()) {
      setError("Die Artikeldaten-Websuche ist in den Einstellungen deaktiviert.");
      setOpen(true);
      setResponse(null);
      return;
    }

    if (!hasArticleSearchCriteria(searchForm, searchInput)) {
      const missingInput = t("vehicles.articleSearch.missingInput");
      setOpen(false);
      setLoading(false);
      setResponse(null);
      setError(missingInput);
      onMessage(missingInput);
      return;
    }

    setOpen(true);
    setLoading(true);
    setError("");
    setResponse(null);
    setSelectedFields({});
    setSelectedImages({});

    api.articleSearch(searchInput ?? {
      manufacturer: searchForm.manufacturer,
      articleNumber: searchForm.articleNumber,
      name: searchForm.name,
      gauge: searchForm.gauge,
      searchSources: articleSearchSources(),
      fields: vehicleFieldsForSearch(searchForm)
    })
      .then((nextResponse) => {
        const sanitized = sanitizeArticleSearchResponse(nextResponse);
        const initialSelection: Record<string, boolean> = {};
        sanitized.results.forEach((result, index) => {
          Object.keys(result.fields).filter(isArticleFieldKey).forEach((key) => {
            initialSelection[articleSelectionKey(result, key, index)] = !currentArticleValue(searchForm, key);
          });
        });
        setResponse(sanitized);
        setSelectedFields(initialSelection);
      })
      .catch((requestError: Error) => setError(requestError.message))
      .finally(() => setLoading(false));
  };

  const openBarcode = () => {
    setBarcodeValue(form.ean || "");
    setBarcodeOpen(true);
  };

  const submitBarcode = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const code = barcodeValue.trim();
    if (!code) {
      onMessage("Bitte einen Barcode oder eine EAN eingeben.");
      return;
    }

    const nextForm = { ...form, ean: code };
    replaceForm(nextForm);
    setBarcodeOpen(false);
    run(nextForm, {
      searchSources: articleSearchSources(),
      fields: { ean: code }
    });
  };

  const toggleField = (result: ArticleSearchResult, index: number, key: string, checked: boolean) => {
    setSelectedFields((current) => ({ ...current, [articleSelectionKey(result, key, index)]: checked }));
  };

  const toggleImage = (
    result: ArticleSearchResult,
    index: number,
    image: ArticleSearchImage,
    checked: boolean
  ) => {
    setSelectedImages((current) => ({ ...current, [imageSelectionKey(result, image, index)]: checked }));
  };

  const applyResult = (result: ArticleSearchResult) => {
    const patch: Partial<CreateVehicleRequest> = {};
    const foundResultIndex = response?.results.findIndex((entry) => entry.url === result.url) ?? 0;
    const resultIndex = foundResultIndex >= 0 ? foundResultIndex : 0;

    Object.entries(result.fields).forEach(([key, field]) => {
      if (!isArticleFieldKey(key) || !selectedFields[articleSelectionKey(result, key, resultIndex)]) return;
      if (isBadArticleValue(key, field.value)) return;
      Object.assign(patch, { [key]: articleValueForForm(key, field.value) });
    });

    const images = (result.images || [])
      .filter((image) => selectedImages[imageSelectionKey(result, image, resultIndex)])
      .map((image, imageIndex) => ({
        ...image,
        id: `${result.url}-${image.url}`,
        isPrimary: pendingImageCount === 0 && imageIndex === 0
      }));
    if (images.length > 0) addImages(images);

    updateForm(patch);
    setOpen(false);
  };

  const restoreDraft = (
    draft: VehicleCreateArticleDraft | null,
    selectedResultIndex: number | null,
    imagesApplied: boolean
  ) => {
    setResponse(draft?.response || null);
    setSelectedFields(draft?.selectedFields || {});
    setSelectedImages(draft?.selectedImages || {});
    setOpen(false);
    setError("");
    if (!imagesApplied || !draft?.response || selectedResultIndex === null) return;
    const result = draft.response.results[selectedResultIndex];
    if (!result) return;
    const images = (result.images || [])
      .filter((image) => draft.selectedImages[imageSelectionKey(result, image, selectedResultIndex)])
      .map((image, imageIndex) => ({
        ...image,
        id: `${result.url}-${image.url}`,
        isPrimary: pendingImageCount === 0 && imageIndex === 0
      }));
    if (images.length > 0) addImages(images);
  };

  return {
    state: {
      open,
      loading,
      response,
      error,
      barcodeOpen,
      barcodeValue,
      selectedFields,
      selectedImages
    },
    setters: {
      setOpen,
      setBarcodeOpen,
      setBarcodeValue
    },
    commands: {
      run,
      openBarcode,
      submitBarcode,
      toggleField,
      toggleImage,
      applyResult,
      restoreDraft
    }
  };
}

export type ArticleSearchController = ReturnType<typeof useArticleSearchController>;
