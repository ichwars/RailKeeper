import { useCallback, useEffect, useState } from "react";
import { api, type Vehicle } from "../../shared/api";
import { emptyOptions, type MasterDataOptions } from "./vehicleViewModel";

export function useVehicleCollectionController() {
  const [vehicles, setVehicles] = useState<Vehicle[]>([]);
  const [options, setOptions] = useState<MasterDataOptions>(emptyOptions);
  const [query, setQuery] = useState("");
  const [message, setMessage] = useState("");
  const [loading, setLoading] = useState(false);
  const load = useCallback(() => {
    setLoading(true);
    setMessage("");
    api
      .vehicles(query)
      .then(setVehicles)
      .catch((error: Error) => setMessage(error.message))
      .finally(() => setLoading(false));
  }, [query]);

  useEffect(() => {
    load();
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
  }, []);

  return { vehicles, setVehicles, options, query, setQuery, message, setMessage, loading, load };
}
