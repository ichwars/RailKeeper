import { useCallback, useEffect, useRef, useState } from "react";

import { api, type MasterDataEntry } from "../../shared/api";

export type ArticleCoreMasterData = {
  manufacturers: MasterDataEntry[];
  gauges: MasterDataEntry[];
  stockUnits: MasterDataEntry[];
  loading: boolean;
  error: boolean;
  retry: () => Promise<void>;
};

export function useArticleCoreMasterData(enabled: boolean): ArticleCoreMasterData {
  const requestRef = useRef(0);
  const [manufacturers, setManufacturers] = useState<MasterDataEntry[]>([]);
  const [gauges, setGauges] = useState<MasterDataEntry[]>([]);
  const [stockUnits, setStockUnits] = useState<MasterDataEntry[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(false);

  const load = useCallback(async () => {
    const request = ++requestRef.current;
    setLoading(true);
    setError(false);
    try {
      const [nextManufacturers, nextGauges, nextStockUnits] = await Promise.all([
        api.masterData("manufacturer"),
        api.masterData("gauge"),
        api.masterData("stock_unit")
      ]);
      if (requestRef.current !== request) return;
      setManufacturers(nextManufacturers);
      setGauges(nextGauges);
      setStockUnits(nextStockUnits);
    } catch {
      if (requestRef.current === request) setError(true);
    } finally {
      if (requestRef.current === request) setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!enabled) {
      requestRef.current += 1;
      return;
    }
    void load();
    return () => { requestRef.current += 1; };
  }, [enabled, load]);

  return { manufacturers, gauges, stockUnits, loading, error, retry: load };
}
