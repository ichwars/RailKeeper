import { useEffect, useRef, useState } from "react";

import { api } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import type { DigitalCenterWorkItem } from "./digitalCenterModel";
import { readECoSStoredCVs } from "./readECoSStoredCVs";

export function ECoSStoredCVPanel({ item, onChange, onBusy, externalReading = false,
  externalError = "", onCancelRead }: {
  item: DigitalCenterWorkItem;
  onChange: (item: DigitalCenterWorkItem) => void;
  onBusy: (busy: boolean) => void;
  externalReading?: boolean;
  externalError?: string;
  onCancelRead?: () => void;
}) {
  const { t } = useI18n();
  const [reading, setReading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [result, setResult] = useState<{ imported: number; preserved: number } | null>(null);
  const cancelled = useRef(false);
  const alive = useRef(true);
  const busy = useRef(false);
  useEffect(() => {
    alive.current = true;
    return () => { alive.current = false; cancelled.current = true; };
  }, []);
  const values = item.center.cvs || [];
  const state = item.center.cvRead;
  const checked = state ? state.nextCV - 1 : 0;
  const complete = checked === 1024;
  const readFailed = Boolean(error || externalError);

  const readAll = async () => {
    if (busy.current) return;
    busy.current = true;
    cancelled.current = false;
    setReading(true);
    onBusy(true);
    setError("");
    setResult(null);
    try {
      await readECoSStoredCVs(item, (next) => { if (alive.current) onChange(next); }, {
        restart: complete && !readFailed, refresh: complete && readFailed,
        isCancelled: () => cancelled.current || !alive.current
      });
    } catch {
      if (alive.current) setError(t("digitalCenters.cvs.readError"));
    } finally {
      busy.current = false;
      if (alive.current) { setReading(false); onBusy(false); }
    }
  };

  const importValues = async () => {
    if (busy.current || !item.vehicleId) return;
    busy.current = true;
    cancelled.current = false;
    setSaving(true);
    onBusy(true);
    setError("");
    try {
      setReading(true);
      const snapshot = await readECoSStoredCVs(item, (next) => { if (alive.current) onChange(next); },
        { refresh: true, isCancelled: () => cancelled.current || !alive.current });
      if (!snapshot) return;
      if (!snapshot.center.cvs?.length) { setResult(null); return; }
      setReading(false);
      const imported = await api.importDigitalCenterCVs(item.sessionId, item.id, item.vehicleId);
      if (alive.current) setResult(imported);
    } catch {
      if (alive.current) setError(t("digitalCenters.cvs.importError"));
    } finally {
      busy.current = false;
      if (alive.current) { setSaving(false); setReading(false); onBusy(false); }
    }
  };

  return <section className="ecos-stored-cvs" aria-label={t("digitalCenters.cvs.title")}>
    <h3>{t("digitalCenters.cvs.title")}</h3>
    <p>{t("digitalCenters.cvs.explanation")}</p>
    <p>{t("digitalCenters.cvs.indexHint")}</p>
    <div className="ecos-stored-cvs-actions">
      <button type="button" className="digital-center-button" disabled={reading || saving || externalReading}
        onClick={() => void readAll()}>
        {reading || externalReading ? t("digitalCenters.cvs.reading") : (checked > 0 && !complete) || readFailed
          ? t("digitalCenters.cvs.resume") : t("digitalCenters.cvs.readAll")}
      </button>
      {(reading || externalReading) && <button type="button" className="digital-center-button"
        onClick={() => { cancelled.current = true; onCancelRead?.(); }}>{t("digitalCenters.cvs.stop")}</button>}
      {item.vehicleId && values.length > 0 && <button type="button" className="digital-center-button"
        disabled={reading || saving || externalReading} onClick={() => void importValues()}>
        {saving ? t("digitalCenters.cvs.saving") : t("digitalCenters.cvs.import")}
      </button>}
    </div>
    <div role="status" aria-live="polite">
      <p>{t("digitalCenters.cvs.count", { count: values.length, checked })}
        {state && ` · ${complete ? t("digitalCenters.cvs.complete") : t("digitalCenters.cvs.partial")}`}</p>
      {(reading || externalReading || checked > 0) && <progress max={1024} value={checked}
        aria-label={t("digitalCenters.cvs.progress")} />}
      {complete && !values.length && <p>{t("digitalCenters.cvs.unavailable")}</p>}
      {result && <p>{t("digitalCenters.cvs.imported", result)}</p>}
    </div>
    {(error || externalError) && <p role="alert">{error || externalError}</p>}
    {values.length > 0 && <details>
      <summary>{t("digitalCenters.cvs.showValues")}</summary>
      {state && <p>{state.sourceHost} · #{item.centerObjectId} · {state.readAt}</p>}
      <div className="ecos-stored-cvs-table"><table>
        <thead><tr><th scope="col">CV</th><th scope="col">{t("digitalCenters.cvs.value")}</th></tr></thead>
        <tbody>{values.map((value) => <tr key={value.number}>
          <th scope="row">{value.number}</th><td>{value.value}</td>
        </tr>)}</tbody>
      </table></div>
    </details>}
    {values.length > 0 && <p>{t("digitalCenters.cvs.preserveHint")}</p>}
  </section>;
}
