import { useEffect, useRef, useState } from "react";

import {
  api,
  type ArticleSearchResponse,
  type ArticleSearchSparePart,
  type Vehicle,
  type VehicleAttachment,
  type VehicleSparePart,
  type VehicleSparePartInput
} from "../../shared/api";
import { sourceDisplayName } from "./articleSearch";
import {
  sparePartResultKey,
  strictCleanSparePartDescription,
  type SparePartSortDirection,
  type SparePartSortKey
} from "./VehicleSparePartsTab";
import type { AttachmentEditState } from "./vehicleTransforms";
import {
  sparePartCatalogKey,
  sparePartImportKey,
  sparePartLookupCandidates,
  sparePartLookupMode,
  sparePartSearchCandidateMatches,
  sparePartSearchSourcesForLookup,
  sparePartStatusCandidate,
  visibleSparePartUrl
} from "./vehicleSpareParts";
import {
  articleSearchEnabled,
  articleSearchSources,
  emptySparePartForm,
  sanitizeArticleSearchResponse
} from "./vehicleViewModel";

type Translate = (key: string, values?: Record<string, string | number>) => string;

type UseVehicleSparePartsControllerOptions = {
  selected: Vehicle | null;
  active: boolean;
  attachmentEdits: AttachmentEditState;
  setSaving: (saving: boolean) => void;
  onMessage: (message: string) => void;
  onOpenSpareParts: () => void;
  refreshSelectedVehicle: (vehicleId?: string) => Promise<void>;
  t: Translate;
};

export function useVehicleSparePartsController({
  selected,
  active,
  attachmentEdits,
  setSaving,
  onMessage,
  onOpenSpareParts,
  refreshSelectedVehicle,
  t
}: UseVehicleSparePartsControllerOptions) {
  const [form, setForm] = useState<VehicleSparePartInput>(emptySparePartForm);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [searchLoading, setSearchLoading] = useState(false);
  const [searchError, setSearchError] = useState("");
  const [searchRan, setSearchRan] = useState(false);
  const [foundParts, setFoundParts] = useState<ArticleSearchSparePart[]>([]);
  const [selectedFoundParts, setSelectedFoundParts] = useState<Record<string, boolean>>({});
  const [sort, setSort] = useState<{ key: SparePartSortKey; direction: SparePartSortDirection }>({
    key: "articleNumber",
    direction: "asc"
  });
  const [lookupLoadingId, setLookupLoadingId] = useState("");
  const [lookupErrors, setLookupErrors] = useState<Record<string, string>>({});
  const [lookupResults, setLookupResults] = useState<Record<string, ArticleSearchSparePart[]>>({});
  const [statusLoading, setStatusLoading] = useState<Record<string, boolean>>({});
  const [statuses, setStatuses] = useState<Record<string, ArticleSearchSparePart>>({});
  const [importAllLoading, setImportAllLoading] = useState(false);
  const statusCheckKeyRef = useRef("");

  const updateForm = (patch: Partial<VehicleSparePartInput>) => {
    setForm((current) => ({ ...current, ...patch }));
  };

  const resetForm = () => {
    setForm(emptySparePartForm);
    setEditingId(null);
  };

  const resetSearch = () => {
    setSearchLoading(false);
    setSearchError("");
    setSearchRan(false);
    setFoundParts([]);
    setSelectedFoundParts({});
  };

  const resetDetail = () => {
    resetForm();
    resetSearch();
    setStatusLoading({});
    setStatuses({});
    setImportAllLoading(false);
    statusCheckKeyRef.current = "";
  };

  const edit = (part: VehicleSparePart) => {
    setForm({
      articleNumber: part.articleNumber || "",
      description: part.description || "",
      price: part.price || "",
      url: part.url?.startsWith("/api/v1/vehicles/") ? "" : part.url || ""
    });
    setEditingId(part.id);
    window.setTimeout(() => {
      const editor = document.getElementById("vehicle-spare-parts-editor") as HTMLDetailsElement | null;
      editor?.scrollIntoView({ behavior: "smooth", block: "start" });
      if (editor) editor.open = true;
      document.getElementById("vehicle-spare-part-article-number")?.focus({ preventScroll: true });
    }, 0);
  };

  const save = () => {
    if (!selected) return;
    const payload: VehicleSparePartInput = {
      articleNumber: form.articleNumber?.trim() || "",
      description: form.description?.trim() || "",
      price: form.price?.trim() || "",
      url: form.url?.trim() || ""
    };
    if (!payload.articleNumber && !payload.description && !payload.url) {
      onMessage(t("vehicles.spareParts.incomplete"));
      return;
    }
    setSaving(true);
    onMessage("");
    const action = editingId
      ? api.updateVehicleSparePart(selected.id, editingId, payload)
      : api.createVehicleSparePart(selected.id, payload);
    action
      .then(() => refreshSelectedVehicle(selected.id))
      .then(resetForm)
      .catch((error: Error) => onMessage(error.message))
      .finally(() => setSaving(false));
  };

  const remove = (part: VehicleSparePart) => {
    if (!selected) return;
    setSaving(true);
    onMessage("");
    api.deleteVehicleSparePart(selected.id, part.id)
      .then(() => refreshSelectedVehicle(selected.id))
      .catch((error: Error) => onMessage(error.message))
      .finally(() => setSaving(false));
  };

  const toggleFound = (key: string, checked: boolean) => {
    setSelectedFoundParts((current) => ({ ...current, [key]: checked }));
  };

  const toggleAllFound = (checked: boolean) => {
    setSelectedFoundParts(Object.fromEntries(
      foundParts.map((part, index) => [sparePartResultKey(part, index), checked])
    ));
  };

  const toggleSort = (key: SparePartSortKey) => {
    setSort((current) => ({
      key,
      direction: current.key === key && current.direction === "asc" ? "desc" : "asc"
    }));
  };

  const selectedInputs = () => foundParts
    .filter((part, index) => selectedFoundParts[sparePartResultKey(part, index)])
    .map((part) => ({
      articleNumber: part.articleNumber || "",
      description: strictCleanSparePartDescription(part.description) || part.description || "",
      price: part.price || "",
      url: part.url || ""
    }));

  const refreshStatuses = async (vehicle: Vehicle) => {
    const linkedParts = (vehicle.spareParts || []).filter((part) => visibleSparePartUrl(part));
    if (linkedParts.length === 0 || !articleSearchEnabled()) return;
    setStatusLoading((current) => ({
      ...current,
      ...Object.fromEntries(linkedParts.map((part) => [part.id, true]))
    }));
    try {
      const lookupMode = sparePartLookupMode(vehicle);
      if (lookupMode) {
        const response = await api.articleSearch({
          manufacturer: vehicle.manufacturer,
          articleNumber: vehicle.articleNumber || linkedParts[0]?.articleNumber || "",
          name: "",
          gauge: "",
          searchSources: sparePartSearchSourcesForLookup(),
          fields: {
            manufacturer: vehicle.manufacturer || "",
            articleNumber: vehicle.articleNumber || linkedParts[0]?.articleNumber || "",
            vehicleArticleNumber: vehicle.articleNumber || "",
            sparePartLookup: lookupMode
          }
        });
        const nextStatuses: Record<string, ArticleSearchSparePart> = {};
        linkedParts.forEach((part) => {
          const candidate = sparePartStatusCandidate(part, response);
          if (candidate?.availability || candidate?.price || candidate?.url) nextStatuses[part.id] = candidate;
        });
        setStatuses((current) => ({ ...current, ...nextStatuses }));
        return;
      }

      const batch = linkedParts.filter((part) => part.articleNumber?.trim()).slice(0, 4);
      const results = await Promise.allSettled(batch.map((part) => api.articleSearch({
        manufacturer: vehicle.manufacturer,
        articleNumber: part.articleNumber || "",
        name: "",
        gauge: "",
        searchSources: sparePartSearchSourcesForLookup(),
        fields: {
          manufacturer: vehicle.manufacturer || "",
          articleNumber: part.articleNumber || "",
          vehicleArticleNumber: vehicle.articleNumber || "",
          sparePartLookup: lookupMode
        }
      }).then((response) => ({ part, response }))));
      const nextStatuses: Record<string, ArticleSearchSparePart> = {};
      results.forEach((result) => {
        if (result.status !== "fulfilled") return;
        const candidate = sparePartStatusCandidate(result.value.part, result.value.response);
        if (candidate?.availability || candidate?.price || candidate?.url) {
          nextStatuses[result.value.part.id] = candidate;
        }
      });
      setStatuses((current) => ({ ...current, ...nextStatuses }));
    } finally {
      setStatusLoading((current) => {
        const next = { ...current };
        linkedParts.forEach((part) => delete next[part.id]);
        return next;
      });
    }
  };

  useEffect(() => {
    if (!active || !selected) return;
    const linkedParts = (selected.spareParts || []).filter((part) => visibleSparePartUrl(part));
    if (linkedParts.length === 0) return;
    const statusKey = `${selected.id}|${selected.articleNumber}|${selected.manufacturer}|${linkedParts
      .map((part) => `${part.id}:${part.articleNumber}:${part.url}`).join("|")}`;
    if (statusCheckKeyRef.current === statusKey) return;
    statusCheckKeyRef.current = statusKey;
    refreshStatuses(selected).catch(() => setStatusLoading({}));
  }, [active, selected]);

  const searchSingle = async (part: VehicleSparePart) => {
    if (!selected || !part.articleNumber?.trim()) return;
    if (!articleSearchEnabled()) {
      setLookupErrors((current) => ({ ...current, [part.id]: t("vehicles.spareParts.searchDisabledSettings") }));
      setLookupResults((current) => ({ ...current, [part.id]: [] }));
      return;
    }
    setLookupLoadingId(part.id);
    setLookupErrors((current) => ({ ...current, [part.id]: "" }));
    setLookupResults((current) => ({ ...current, [part.id]: [] }));
    try {
      const response = await api.articleSearch({
        manufacturer: selected.manufacturer,
        articleNumber: part.articleNumber || "",
        name: "",
        gauge: "",
        searchSources: sparePartSearchSourcesForLookup(),
        fields: {
          manufacturer: selected.manufacturer || "",
          articleNumber: part.articleNumber || "",
          vehicleArticleNumber: selected.articleNumber || "",
          sparePartLookup: sparePartLookupMode(selected)
        }
      });
      setLookupResults((current) => ({ ...current, [part.id]: sparePartLookupCandidates(part, response) }));
    } catch (error) {
      setLookupErrors((current) => ({
        ...current,
        [part.id]: error instanceof Error ? error.message : String(error)
      }));
    } finally {
      setLookupLoadingId("");
    }
  };

  const applyLookup = (part: VehicleSparePart, result: ArticleSearchSparePart) => {
    if (!selected) return;
    setSaving(true);
    onMessage("");
    const pricedResult = result.price
      ? result
      : (lookupResults[part.id] || []).find((candidate) => candidate.price) || result;
    api.updateVehicleSparePart(selected.id, part.id, {
      articleNumber: part.articleNumber || "",
      description: part.description || "",
      price: result.price || pricedResult.price || part.price || "",
      url: pricedResult.url || result.url || part.url || ""
    })
      .then(() => {
        setLookupResults((current) => {
          const next = { ...current };
          delete next[part.id];
          return next;
        });
        setLookupErrors((current) => {
          const next = { ...current };
          delete next[part.id];
          return next;
        });
        return refreshSelectedVehicle(selected.id);
      })
      .catch((error: Error) => onMessage(error.message))
      .finally(() => setSaving(false));
  };

  const importAll = async () => {
    if (!selected) return;
    const lookupMode = sparePartLookupMode(selected);
    if (!lookupMode || !selected.articleNumber?.trim()) return;
    if (!articleSearchEnabled()) {
      onMessage(t("vehicles.spareParts.searchDisabledSettings"));
      return;
    }
    setImportAllLoading(true);
    onMessage("");
    try {
      const response = await api.articleSearch({
        manufacturer: selected.manufacturer,
        articleNumber: selected.articleNumber || "",
        name: "",
        gauge: "",
        searchSources: sparePartSearchSourcesForLookup(),
        fields: {
          manufacturer: selected.manufacturer || "",
          articleNumber: selected.articleNumber || "",
          vehicleArticleNumber: selected.articleNumber || "",
          sparePartLookup: lookupMode
        }
      });
      const existingPartsByKey = new Map((selected.spareParts || []).map((part) => [sparePartCatalogKey(part), part]));
      const importable = new Map<string, VehicleSparePartInput>();
      const updates = new Map<string, { id: string; input: VehicleSparePartInput }>();
      sanitizeArticleSearchResponse(response).results.forEach((searchResult) => {
        (searchResult.spareParts || []).forEach((part) => {
          const input: VehicleSparePartInput = {
            articleNumber: part.articleNumber?.trim() || "",
            description: strictCleanSparePartDescription(part.description || "") || part.description || "",
            price: part.price?.trim() || "",
            url: part.url?.trim() || searchResult.url || ""
          };
          const key = sparePartCatalogKey(input);
          if (key === "url:") return;
          const existingPart = existingPartsByKey.get(key);
          if (existingPart) {
            const existingPrice = existingPart.price && existingPart.price !== "-" ? existingPart.price : "";
            const existingUrl = existingPart.url?.startsWith("/api/v1/vehicles/") ? "" : existingPart.url || "";
            const nextPrice = existingPrice || input.price || "";
            const nextUrl = existingUrl || input.url || "";
            if ((nextPrice && nextPrice !== existingPrice) || (nextUrl && nextUrl !== existingUrl)) {
              updates.set(existingPart.id, {
                id: existingPart.id,
                input: {
                  articleNumber: existingPart.articleNumber || input.articleNumber || "",
                  description: existingPart.description || input.description || "",
                  price: nextPrice,
                  url: nextUrl
                }
              });
            }
            return;
          }
          const current = importable.get(key);
          importable.set(key, current ? {
            articleNumber: current.articleNumber || input.articleNumber || "",
            description: current.description || input.description || "",
            price: current.price || input.price || "",
            url: current.url || input.url || ""
          } : input);
        });
      });
      if (importable.size === 0 && updates.size === 0) {
        onMessage(t("vehicles.spareParts.importAllNone"));
        return;
      }
      for (const update of updates.values()) {
        await api.updateVehicleSparePart(selected.id, update.id, update.input);
      }
      for (const part of importable.values()) {
        await api.createVehicleSparePart(selected.id, part);
      }
      await refreshSelectedVehicle(selected.id);
      onMessage(t("vehicles.spareParts.importAllDone", { count: importable.size + updates.size }));
    } catch (error) {
      onMessage(error instanceof Error ? error.message : String(error));
    } finally {
      setImportAllLoading(false);
    }
  };

  const extractAttachment = async (attachment: VehicleAttachment) => {
    if (!selected) return;
    const attachmentEdit = attachmentEdits[attachment.id] || { description: "", category: "", maintenanceId: "" };
    setSaving(true);
    onMessage("");
    setSearchError("");
    try {
      await api.updateVehicleAttachment(selected.id, attachment.id, attachmentEdit);
      const localParts = await api.vehicleSparePartSuggestions(selected.id, attachment.id);
      const candidates = localParts.filter((part) => Boolean(
        part.articleNumber || strictCleanSparePartDescription(part.description) || part.url
      ));
      const existingKeys = new Set((selected.spareParts || []).map((part) => sparePartImportKey(part)));
      const importableParts: VehicleSparePartInput[] = [];
      for (const part of candidates) {
        const input: VehicleSparePartInput = {
          articleNumber: part.articleNumber?.trim() || "",
          description: strictCleanSparePartDescription(part.description) || part.description || "",
          price: part.price?.trim() || "",
          url: part.url?.trim().startsWith("/api/v1/vehicles/") ? "" : part.url?.trim() || ""
        };
        if (!input.articleNumber && !input.description && !input.url) continue;
        const key = sparePartImportKey(input);
        if (existingKeys.has(key)) continue;
        existingKeys.add(key);
        importableParts.push(input);
      }
      if (importableParts.length > 0) {
        for (const part of importableParts) {
          await api.createVehicleSparePart(selected.id, part);
        }
        setFoundParts([]);
        setSearchRan(false);
        setSelectedFoundParts({});
        onMessage(t("vehicles.uploads.extractSparePartsDone", { count: importableParts.length }));
      } else {
        setFoundParts(candidates);
        setSearchRan(true);
        setSelectedFoundParts({});
        onMessage(t("vehicles.uploads.extractSparePartsEmpty"));
      }
      await refreshSelectedVehicle(selected.id);
      onOpenSpareParts();
    } catch (error) {
      onMessage(error instanceof Error ? error.message : String(error));
    } finally {
      setSaving(false);
    }
  };

  const runSearch = async () => {
    if (!selected) return;
    const storedParts = (selected.spareParts || []).filter((part) => part.articleNumber?.trim());
    if (storedParts.length === 0) {
      setSearchError(t("vehicles.spareParts.searchDisabledEmpty"));
      setSearchRan(true);
      return;
    }
    if (!articleSearchEnabled()) {
      setSearchError(t("vehicles.spareParts.searchDisabledSettings"));
      setSearchRan(true);
      return;
    }
    setSearchLoading(true);
    setSearchError("");
    setSearchRan(true);
    setFoundParts([]);
    setSelectedFoundParts({});
    const configuredSources = articleSearchSources().filter((source) => source === "manufacturer" || source === "catalogs");
    const searchSources = configuredSources.length > 0 ? configuredSources : ["manufacturer", "catalogs"];
    try {
      const responses: { part: VehicleSparePart; response: ArticleSearchResponse }[] = [];
      let failedSearches = 0;
      for (let index = 0; index < storedParts.length; index += 3) {
        const batchResults = await Promise.allSettled(storedParts.slice(index, index + 3).map((part) =>
          api.articleSearch({
            manufacturer: selected.manufacturer,
            articleNumber: part.articleNumber || "",
            name: "",
            gauge: "",
            searchSources,
            fields: { manufacturer: selected.manufacturer || "", articleNumber: part.articleNumber || "" }
          }).then((response) => ({ part, response }))
        ));
        batchResults.forEach((result) => {
          if (result.status === "fulfilled") responses.push(result.value);
          else failedSearches += 1;
        });
      }
      const parts = new Map<string, ArticleSearchSparePart>();
      responses.forEach(({ part, response }) => {
        sanitizeArticleSearchResponse(response).results.forEach((searchResult) => {
          const resultSignal = `${searchResult.title || ""} ${searchResult.snippet || ""} ${searchResult.url || ""}`
            .toLocaleLowerCase("de-DE");
          (searchResult.spareParts || []).forEach((candidate) => {
            const candidateSignal = `${candidate.articleNumber || ""} ${candidate.description || ""} ${candidate.url || ""}`
              .toLocaleLowerCase("de-DE");
            if (!sparePartSearchCandidateMatches(part, candidateSignal)) return;
            const key = `${candidate.articleNumber || ""}|${candidate.description || ""}|${candidate.url || ""}`
              .toLocaleLowerCase("de-DE");
            if (key !== "||" && !parts.has(key)) parts.set(key, candidate);
          });
          if (sparePartSearchCandidateMatches(part, resultSignal) && searchResult.url) {
            const fallback: ArticleSearchSparePart = {
              articleNumber: part.articleNumber || "",
              description: part.description || searchResult.title || "",
              price: part.price || "",
              url: searchResult.url,
              source: searchResult.source || sourceDisplayName(searchResult.url)
            };
            const key = `${fallback.articleNumber || ""}|${fallback.description || ""}|${fallback.url || ""}`
              .toLocaleLowerCase("de-DE");
            if (key !== "||" && !parts.has(key)) parts.set(key, fallback);
          }
        });
      });
      setFoundParts(Array.from(parts.values()));
      setSelectedFoundParts({});
      if (failedSearches > 0 && parts.size === 0) setSearchError(t("vehicles.spareParts.searchFailed"));
      else if (failedSearches > 0) {
        setSearchError(t("vehicles.spareParts.searchPartialFailed", { count: failedSearches }));
      }
    } catch (error) {
      setSearchError(error instanceof Error ? error.message : String(error));
    } finally {
      setSearchLoading(false);
    }
  };

  const canImportAll = Boolean(selected && sparePartLookupMode(selected) && selected.articleNumber?.trim());
  const importAllTitle = !selected
    ? t("vehicles.spareParts.importAllNoVehicle")
    : !sparePartLookupMode(selected)
      ? t("vehicles.spareParts.importAllUnsupported")
      : !selected.articleNumber?.trim()
        ? t("vehicles.spareParts.importAllMissingArticle")
        : t("vehicles.spareParts.importAllTitle");

  return {
    state: {
      form,
      editingId,
      searchLoading,
      searchError,
      searchRan,
      foundParts,
      selectedFoundParts,
      sort,
      lookupLoadingId,
      lookupErrors,
      lookupResults,
      statusLoading,
      statuses,
      importAllLoading,
      canImportAll,
      importAllTitle
    },
    commands: {
      updateForm,
      resetForm,
      resetSearch,
      resetDetail,
      edit,
      save,
      remove,
      toggleFound,
      toggleAllFound,
      toggleSort,
      selectedInputs,
      clearSelectedFound: () => setSelectedFoundParts({}),
      searchSingle,
      applyLookup,
      importAll,
      extractAttachment,
      runSearch
    }
  };
}
