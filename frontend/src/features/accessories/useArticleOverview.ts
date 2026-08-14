import { useCallback, useEffect, useMemo, useRef, useState } from "react";

import {
  api,
  type AccessoryArticleListQuery,
  type AccessoryArticleListResult,
  type AccessoryArticleSort,
  type AccessoryArticleStatus,
  type AccessoryArticleType,
  type AccessorySortDirection
} from "../../shared/api";

export type ArticleOverviewStatusFilter = "" | AccessoryArticleStatus | "allocated";

export type ArticleOverviewFilters = {
  query: string;
  articleType: "" | AccessoryArticleType;
  manufacturer: string;
  gauge: string;
  status: ArticleOverviewStatusFilter;
  locationId: string;
};

const emptyFilters: ArticleOverviewFilters = {
  query: "",
  articleType: "",
  manufacturer: "",
  gauge: "",
  status: "",
  locationId: ""
};

const emptyResult: AccessoryArticleListResult = {
  items: [],
  metrics: {
    articleCount: 0,
    articleTypeCount: 0,
    available: 0,
    locationCount: 0,
    reserved: 0,
    installed: 0,
    careHintCount: 0
  },
  filters: {
    manufacturers: [],
    articleTypes: [],
    gauges: [],
    storageLocations: []
  }
};

function statusesForFilter(status: ArticleOverviewStatusFilter): AccessoryArticleStatus[] | undefined {
  if (!status) return undefined;
  if (status === "allocated") return ["reserved", "installed"];
  return [status];
}

function buildQuery(
  filters: ArticleOverviewFilters,
  sort: AccessoryArticleSort,
  direction: AccessorySortDirection
): AccessoryArticleListQuery {
  return {
    ...(filters.query ? { query: filters.query } : {}),
    ...(filters.articleType ? { articleTypes: [filters.articleType] } : {}),
    ...(filters.manufacturer ? { manufacturer: filters.manufacturer } : {}),
    ...(filters.gauge ? { gauges: [filters.gauge] } : {}),
    ...(filters.status ? { statuses: statusesForFilter(filters.status) } : {}),
    ...(filters.locationId ? { locationId: filters.locationId } : {}),
    sort,
    direction
  };
}

export function useArticleOverview({ enabled = true }: { enabled?: boolean } = {}) {
  const [filters, setFilters] = useState<ArticleOverviewFilters>(emptyFilters);
  const [sort, setSortState] = useState<AccessoryArticleSort>("inventoryNumber");
  const [direction, setDirection] = useState<AccessorySortDirection>("asc");
  const [data, setData] = useState<AccessoryArticleListResult>(emptyResult);
  const [loading, setLoading] = useState(enabled);
  const [error, setError] = useState("");
  const [reloadToken, setReloadToken] = useState(0);
  const requestSequence = useRef(0);

  const query = useMemo(() => buildQuery(filters, sort, direction), [direction, filters, sort]);

  useEffect(() => {
    if (!enabled) {
      setLoading(false);
      return;
    }
    const request = ++requestSequence.current;
    setLoading(true);
    setError("");
    api.accessoryArticles(query)
      .then((next) => {
        if (request === requestSequence.current) setData(next);
      })
      .catch((reason: unknown) => {
        if (request !== requestSequence.current) return;
        setError(reason instanceof Error ? reason.message : String(reason));
      })
      .finally(() => {
        if (request === requestSequence.current) setLoading(false);
      });
    return () => {
      if (request === requestSequence.current) requestSequence.current += 1;
    };
  }, [enabled, query, reloadToken]);

  const setFilter = useCallback(<Key extends keyof ArticleOverviewFilters>(
    key: Key,
    value: ArticleOverviewFilters[Key]
  ) => {
    setFilters((current) => ({ ...current, [key]: value }));
  }, []);

  const resetFilters = useCallback(() => setFilters(emptyFilters), []);

  const setSort = useCallback((nextSort: AccessoryArticleSort) => {
    if (nextSort === sort) {
      setDirection((current) => current === "asc" ? "desc" : "asc");
      return;
    }
    setSortState(nextSort);
    setDirection("asc");
  }, [sort]);

  const reload = useCallback(() => setReloadToken((current) => current + 1), []);

  const archiveArticle = useCallback(async (id: string) => {
    setError("");
    try {
      await api.archiveAccessoryProduct(id);
      reload();
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : String(reason));
    }
  }, [reload]);

  const restoreArticle = useCallback(async (id: string) => {
    setError("");
    try {
      await api.restoreAccessoryProduct(id);
      reload();
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : String(reason));
    }
  }, [reload]);

  const deleteArticle = useCallback(async (id: string) => {
    setError("");
    await api.deleteAccessoryProduct(id);
    reload();
  }, [reload]);

  return {
    data,
    filters,
    sort,
    direction,
    loading,
    error,
    hasActiveFilters: Object.values(filters).some(Boolean),
    setFilter,
    resetFilters,
    setSort,
    reload,
    archiveArticle,
    restoreArticle,
    deleteArticle
  };
}
