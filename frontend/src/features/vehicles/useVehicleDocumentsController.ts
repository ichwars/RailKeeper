import { useState } from "react";

import { api, type ArticleSearchDocument, type Vehicle } from "../../shared/api";
import { categoryForWebDocument, uniqueWebDocuments, webDocumentKey } from "./vehicleDocuments";
import { vehicleToForm } from "./vehicleTransforms";
import {
  articleSearchEnabled,
  articleSearchSources,
  hasArticleSearchCriteria,
  sanitizeArticleSearchResponse,
  vehicleFieldsForSearch
} from "./vehicleViewModel";

type Translate = (key: string, values?: Record<string, string | number>) => string;

type UseVehicleDocumentsControllerOptions = {
  selected: Vehicle | null;
  setSaving: (saving: boolean) => void;
  onMessage: (message: string) => void;
  refreshSelectedVehicle: (vehicleId?: string) => Promise<void>;
  t: Translate;
};

export function useVehicleDocumentsController({
  selected,
  setSaving,
  onMessage,
  refreshSelectedVehicle,
  t
}: UseVehicleDocumentsControllerOptions) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [ran, setRan] = useState(false);
  const [documents, setDocuments] = useState<ArticleSearchDocument[]>([]);
  const [selectedDocuments, setSelectedDocuments] = useState<Record<string, boolean>>({});

  const reset = () => {
    setLoading(false);
    setError("");
    setRan(false);
    setDocuments([]);
    setSelectedDocuments({});
  };

  const search = async () => {
    if (!selected) return;
    if (!articleSearchEnabled()) {
      setError("Die Artikeldaten-Websuche ist in den Einstellungen deaktiviert.");
      setRan(true);
      return;
    }
    const searchForm = vehicleToForm(selected);
    if (!hasArticleSearchCriteria(searchForm)) {
      setError(t("vehicles.articleSearch.missingInput"));
      setRan(true);
      return;
    }
    setLoading(true);
    setError("");
    setRan(true);
    setDocuments([]);
    setSelectedDocuments({});
    try {
      const response = sanitizeArticleSearchResponse(await api.articleSearch({
        manufacturer: searchForm.manufacturer,
        articleNumber: searchForm.articleNumber,
        name: searchForm.name,
        gauge: searchForm.gauge,
        searchSources: articleSearchSources(),
        fields: vehicleFieldsForSearch(searchForm)
      }));
      setDocuments(uniqueWebDocuments(response.results.flatMap((result) => result.documents || [])));
    } catch (searchError) {
      setError(searchError instanceof Error ? searchError.message : String(searchError));
    } finally {
      setLoading(false);
    }
  };

  const importDocuments = (requestedDocuments: ArticleSearchDocument[]) => {
    if (!selected) return;
    const importable = requestedDocuments.filter((document) => document.url);
    if (importable.length === 0) return;
    setSaving(true);
    onMessage("");
    (async () => {
      for (const document of importable) {
        await api.importVehicleAttachmentFromUrl(selected.id, {
          url: document.url,
          title: document.title || "Dokument",
          description: `Quelle: ${document.source || document.url}\n${document.url}`,
          category: categoryForWebDocument(document),
          maintenanceId: ""
        });
      }
    })()
      .then(() => refreshSelectedVehicle(selected.id))
      .then(() => {
        setSelectedDocuments({});
        const messageKey = importable.length === 1
          ? "vehicles.uploads.webDocumentImported"
          : "vehicles.uploads.webDocumentsImported";
        onMessage(t(messageKey, { count: importable.length }));
      })
      .catch((importError: Error) => onMessage(importError.message))
      .finally(() => setSaving(false));
  };

  const toggle = (document: ArticleSearchDocument, index: number, checked: boolean) => {
    const key = webDocumentKey(document, index);
    setSelectedDocuments((current) => {
      const next = { ...current };
      if (checked) next[key] = true;
      else delete next[key];
      return next;
    });
  };

  const toggleAll = (checked: boolean) => {
    if (!checked) {
      setSelectedDocuments({});
      return;
    }
    const existingDescriptions = new Set(
      (selected?.attachments || []).map((attachment) => attachment.description || "")
    );
    setSelectedDocuments(Object.fromEntries(documents.flatMap((document, index) => {
      const alreadyImported = document.url && Array.from(existingDescriptions)
        .some((description) => description.includes(document.url));
      if (!document.url || alreadyImported) return [];
      return [[webDocumentKey(document, index), true]];
    })));
  };

  const importSelected = () => {
    importDocuments(documents.filter(
      (document, index) => selectedDocuments[webDocumentKey(document, index)]
    ));
  };

  return {
    state: { loading, error, ran, documents, selectedDocuments },
    commands: {
      reset,
      search,
      importOne: (document: ArticleSearchDocument) => importDocuments([document]),
      importSelected,
      toggle,
      toggleAll
    }
  };
}
