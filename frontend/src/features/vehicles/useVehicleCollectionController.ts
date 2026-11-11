import { useCallback, useEffect, useRef, useState } from "react";
import { api, type Vehicle } from "../../shared/api";
import { searchDebounceMs } from "../../shared/useDebouncedSearch";
import { emptyOptions, type MasterDataOptions } from "./vehicleViewModel";

export function useVehicleCollectionController() {
  const [vehicles, setVehicles] = useState<Vehicle[]>([]);
  const [options, setOptions] = useState<MasterDataOptions>(emptyOptions);
  const [query, setQueryState] = useState("");
  const queryRef = useRef("");
  const timer = useRef<number | undefined>(undefined);
  const requestSequence = useRef(0);
  const mounted = useRef(false);
  const [notice, setNotice] = useState({ message: "", tone: "error" as "error" | "status" });
  const setMessage = useCallback((message: string) => setNotice({ message, tone: "error" }), []);
  const setStatusMessage = useCallback((message: string) => setNotice({ message, tone: "status" }), []);
  const [loading, setLoading] = useState(false);
  const load = useCallback(() => {
    if (!mounted.current) return;
    window.clearTimeout(timer.current);
    const request = ++requestSequence.current;
    setLoading(true);
    setMessage("");
    api
      .vehicles(queryRef.current)
      .then((next) => {
        if (request === requestSequence.current) setVehicles(next);
      })
      .catch((error: Error) => {
        if (request === requestSequence.current) setMessage(error.message);
      })
      .finally(() => {
        if (request === requestSequence.current) setLoading(false);
      });
  }, [setMessage]);

  const setQuery = useCallback((value: string) => {
    if (value === queryRef.current) return;
    queryRef.current = value;
    setQueryState(value);
    requestSequence.current += 1;
    window.clearTimeout(timer.current);
    timer.current = window.setTimeout(load, searchDebounceMs);
  }, [load]);

  useEffect(() => {
    mounted.current = true;
    load();
    return () => {
      mounted.current = false;
      requestSequence.current += 1;
      window.clearTimeout(timer.current);
    };
  }, [load]);

  useEffect(() => {
    const reloadVisible = () => {
      if (!document.hidden) {
        load();
      }
    };
    window.addEventListener("focus", reloadVisible);
    window.addEventListener("online", reloadVisible);
    document.addEventListener("visibilitychange", reloadVisible);
    return () => {
      window.removeEventListener("focus", reloadVisible);
      window.removeEventListener("online", reloadVisible);
      document.removeEventListener("visibilitychange", reloadVisible);
    };
  }, [load]);

  useEffect(() => {
    Promise.all([
      api.masterDataAll(),
      api.masterDataRelations("vehicle_category", "vehicle_gattung")
    ])
      .then(([entriesByType, categoryRelations]) => {
        setOptions({
          manufacturers: entriesByType.manufacturer || [],
          gauges: entriesByType.gauge || [],
          epochs: entriesByType.epoch || [],
          railwayCompanies: entriesByType.railway_company || [],
          categories: entriesByType.vehicle_category || [],
          gattungen: entriesByType.vehicle_gattung || [],
          symbols: entriesByType.symbols || [],
          categoryRelations
        });
      })
      .catch((error: Error) => setMessage(error.message));
  }, [setMessage]);

  return { vehicles, setVehicles, options, query, setQuery, message: notice.message, messageTone: notice.tone,
    setMessage, setStatusMessage, loading, load };
}
