import { useCallback, useEffect, useRef, useState } from "react";
import { api, ApiError } from "../../shared/api";
import type { VehicleValuation, VehicleValuationHistory } from "../../shared/apiValuations";
import { newValuationDraft, sameValuationRequest, valuationChangedEvent, valuationHistoryPageSize, type ValuationDraft } from "./valuationModel";

export function useVehicleValuations(vehicleID: string, active: boolean) {
  const [history, setHistory] = useState<VehicleValuationHistory>({ entries: [], total: 0 });
  const [draft, setDraft] = useState<ValuationDraft>(() => newValuationDraft());
  const [loading, setLoading] = useState(false);
  const [loadFailed, setLoadFailed] = useState(false);
  const [saving, setSaving] = useState(false);
  const [uncertain, setUncertain] = useState<ValuationDraft | null>(null);
  const [error, setError] = useState("");
  const [status, setStatus] = useState("");
  const readGeneration = useRef(0), lifetime = useRef(0), writeBusy = useRef(false);

  const load = useCallback(async (offset = 0) => {
    const generation = ++readGeneration.current;
    setLoading(true); setLoadFailed(false);
    try {
      const result = await api.vehicleValuations(vehicleID, offset);
      if (generation !== readGeneration.current) return;
      setHistory(previous => offset === 0 ? result : {
        total: result.total,
        entries: [...previous.entries, ...result.entries.filter(entry => !previous.entries.some(old => old.id === entry.id))]
      });
    } catch {
      if (generation === readGeneration.current) setLoadFailed(true);
    } finally {
      if (generation === readGeneration.current) setLoading(false);
    }
  }, [vehicleID]);

  useEffect(() => {
    if (active) void load();
    return () => { readGeneration.current += 1; };
  }, [active, load]);
  useEffect(() => () => { lifetime.current += 1; }, []);

  function confirmed(record: VehicleValuation, knownHistory?: VehicleValuationHistory) {
    setHistory(previous => knownHistory ?? ({
      total: previous.total + (previous.entries.some(old => old.id === record.id) ? 0 : 1),
      entries: [record, ...previous.entries.filter(old => old.id !== record.id)
        .map(old => old.id === record.replacesId ? { ...old, replacedById: record.id } : old)]
    }));
    setDraft(newValuationDraft()); setUncertain(null); setError(""); setStatus("valuation.saved");
    window.dispatchEvent(new Event(valuationChangedEvent));
  }

  async function save() {
    if (writeBusy.current || uncertain) return;
    const owner = lifetime.current;
    const input = { ...draft };
    writeBusy.current = true; setSaving(true); setError(""); setStatus("");
    try {
      const record = await api.recordVehicleValuation(vehicleID, input);
      if (owner !== lifetime.current) return;
      if (!sameValuationRequest(record, input) || record.vehicleId !== vehicleID) {
        setUncertain(input); setError("valuation.uncertain"); return;
      }
      confirmed(record);
      await load();
    } catch (failure) {
      if (owner !== lifetime.current) return;
      if (failure instanceof ApiError && failure.status === 400) setError("valuation.invalid");
      else if (failure instanceof ApiError && failure.status === 409) setError("valuation.conflict");
      else if (failure instanceof ApiError && failure.status === 404) setError("valuation.notFound");
      else if (failure instanceof ApiError && (failure.status === 401 || failure.status === 403)) setError("valuation.forbidden");
      else { setUncertain(input); setError("valuation.uncertain"); }
    } finally {
      if (owner === lifetime.current) { writeBusy.current = false; setSaving(false); }
    }
  }

  async function review() {
    if (writeBusy.current || !uncertain) return;
    const owner = lifetime.current;
    writeBusy.current = true; setSaving(true);
    try {
      const exported = await api.vehicleValuationsExport(vehicleID);
      if (owner !== lifetime.current) return;
      if (exported.vehicleId !== vehicleID) throw new Error("valuation owner differs");
      const record = exported.entries.find(entry => entry.id === uncertain.id);
      if (record && record.vehicleId !== vehicleID) throw new Error("valuation owner differs");
      if (record) {
        if (!sameValuationRequest(record, uncertain)) {
          setUncertain(null); setError("valuation.conflict");
          await load(); return;
        }
        confirmed(record, { entries: exported.entries.slice(0, valuationHistoryPageSize), total: exported.entries.length });
      } else { setUncertain(null); setError(""); setStatus("valuation.reviewMissing"); }
      await load();
    } catch {
      if (owner === lifetime.current) setError("valuation.reviewFailed");
    } finally {
      if (owner === lifetime.current) { writeBusy.current = false; setSaving(false); }
    }
  }

  const startCorrection = (record: VehicleValuation, withdrawn: boolean) => {
    if (writeBusy.current || uncertain || record.replacedById || record.vehicleId !== vehicleID) return;
    setDraft(newValuationDraft(record, withdrawn)); setError(""); setStatus("");
  };
  const reset = () => {
    if (writeBusy.current || uncertain) return;
    setDraft(newValuationDraft()); setError(""); setStatus("");
  };

  const rebaseCorrection = (record: VehicleValuation) => {
    if (writeBusy.current || uncertain || error !== "valuation.conflict" || !draft.replacesId || record.replacedById || record.vehicleId !== vehicleID) return;
    const target = newValuationDraft(record, draft.withdrawn);
    setDraft(previous => ({ ...previous, id: target.id, replacesId: record.id,
      ...(previous.withdrawn ? { amount: record.amount, currency: record.currency, valuedOn: record.valuedOn } : {}) }));
    setError(""); setStatus("valuation.rebased");
  };

  return {
    history, draft, setDraft, loading, loadFailed, saving, uncertain, error, status,
    load, save, review, startCorrection, reset, rebaseCorrection
  };
}

export type VehicleValuationController = ReturnType<typeof useVehicleValuations>;
