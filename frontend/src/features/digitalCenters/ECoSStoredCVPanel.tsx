import { useEffect, useMemo, useRef, useState } from "react";

import { api } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import type { DigitalCenterWorkItem } from "./digitalCenterModel";
import { hasCompleteECoSCVSnapshot, readECoSStoredCVs } from "./readECoSStoredCVs";
import { useDecoderLibrary } from "../../shared/DecoderLibraryProvider";
import { resolveDecoderLibrary } from "../../shared/decoderLibrary";
import { DecoderLibraryNotice } from "../vehicles/DecoderLibraryNotice";
import { hasECoSDecoderIdentityConflict } from "../../shared/ecosSourceConflicts";

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
  const library = useDecoderLibrary();
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
  const database = state?.source === "database";
  const backup = state?.source === "backup";
  const comparedIncomplete = Boolean(item.center.cvSources && !hasCompleteECoSCVSnapshot(item));
  const indexed = values.filter(value => value.indexCV31 != null).length;
  const readFailed = Boolean(error || externalError);
  const readings = useMemo(() => values.map(value => ({ cvNumber: value.number, value: value.value,
    decoderProfile: "ECoS", protocol: item.protocol, indexCV31: value.indexCV31, indexCV32: value.indexCV32
  })), [values, item.protocol]);
  const match = useMemo(() => {
    const identityConflict = hasECoSDecoderIdentityConflict(item.center.cvSources?.conflicts, library.inventory);
    return resolveDecoderLibrary(readings, "ECoS", identityConflict ? [] : library.profiles,
      identityConflict ? null : library.inventory ?? null, item.railkeeper.decoderType || "ECoS", item.center.profile);
  }, [readings, library.profiles, library.inventory, item.railkeeper.decoderType, item.center.profile, item.center.cvSources]);

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
    if (busy.current || !item.vehicleId || comparedIncomplete) return;
    busy.current = true;
    cancelled.current = false;
    setSaving(true);
    onBusy(true);
    setError("");
    try {
      setReading(!hasCompleteECoSCVSnapshot(item));
      const snapshot = await readECoSStoredCVs(item,
        (next) => { if (alive.current) onChange(next); },
        { useSaved: true, isCancelled: () => cancelled.current || !alive.current });
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
    <p>{t(backup ? "digitalCenters.backup.cvExplanation" : "digitalCenters.cvs.explanation")}</p>
    {(item.center.sourceProfile || item.center.profile) && <p>{t("digitalCenters.cvs.profile", {
      profile: item.center.sourceProfile || item.center.profile || ""
    })}</p>}
    {item.center.sourceProfile && !item.center.profile && <p>{t("digitalCenters.cvs.profileUnresolved")}</p>}
    {(item.center.profile || item.center.sourceProfile) && <DecoderLibraryNotice values={readings} profile="ECoS" match={match} />}
    <p>{t("digitalCenters.cvs.indexHint")}</p>
    <div className="ecos-stored-cvs-actions">
      {!backup && !hasCompleteECoSCVSnapshot(item) && !comparedIncomplete && <button type="button" className="digital-center-button" disabled={reading || saving || externalReading}
        onClick={() => void readAll()}>
        {reading || externalReading ? t("digitalCenters.cvs.reading") : (checked > 0 && !complete) || readFailed
          ? t("digitalCenters.cvs.resume") : t("digitalCenters.cvs.readAll")}
      </button>}
      {(reading || externalReading) && <button type="button" className="digital-center-button"
        onClick={() => { cancelled.current = true; onCancelRead?.(); }}>{t("digitalCenters.cvs.stop")}</button>}
      {item.vehicleId && values.length > 0 && <button type="button" className="digital-center-button"
        disabled={reading || saving || externalReading || comparedIncomplete || (backup && !complete)} onClick={() => void importValues()}>
        {saving ? t("digitalCenters.cvs.saving") : t("digitalCenters.cvs.import")}
      </button>}
    </div>
    {database && complete && <p>{t("digitalCenters.cvs.readWithComparison")}</p>}
    {comparedIncomplete && <p role="status">{t("digitalCenters.cvs.incompleteComparison")}</p>}
    {(item.center.cvSources?.backupStatus === "failed" || item.center.cvSources?.warning ||
      (item.center.cvSources?.backupStatus === "checked" && item.center.cvSources.storedCacheStatus === "unavailable")) && <p role="status">
      {t("digitalCenters.cvs.additionalReadFailed")}</p>}
    {Boolean(item.center.cvSources?.conflicts.length) && <details>
      <summary>{t("digitalCenters.cvs.sourceConflicts", { count: item.center.cvSources!.conflicts.length })}</summary>
      <p>{t("digitalCenters.cvs.sourceConflictsHint")}</p>
      <div className="ecos-stored-cvs-table ecos-source-conflicts" tabIndex={0} role="region"
        aria-label={t("digitalCenters.cvs.sourceConflicts", { count: item.center.cvSources!.conflicts.length })}>
        <table><thead><tr><th>CV</th><th>CV31 / CV32</th>
          <th>{t("digitalCenters.cvs.databaseSource")}</th><th>{t("digitalCenters.cvs.backupSource")}</th></tr></thead>
          <tbody>{item.center.cvSources!.conflicts.map(value => <tr
            key={`${value.number}:${value.indexCV31}:${value.indexCV32}`}>
            <th scope="row">{value.number}</th>
            <td>{value.indexCV31 == null ? "–" : `${value.indexCV31} / ${value.indexCV32}`}</td>
            <td>{value.databaseValue}</td><td>{value.backupValue}
              {value.backupSource === "stored_cache" && <small>{t("digitalCenters.cvs.rawStatusUnknown")}</small>}
              {value.rawCacheValue != null && <small>{t("digitalCenters.cvs.additionalRawValue", { value: value.rawCacheValue })}</small>}
            </td>
          </tr>)}</tbody>
        </table>
      </div>
    </details>}
    {Boolean(item.center.cvSources?.unverifiedValues?.length) && <details>
      <summary>{t("digitalCenters.cvs.unverifiedCount", { count: item.center.cvSources!.unverifiedValues!.length })}</summary>
      <p>{t("digitalCenters.cvs.unverifiedHint")}</p>
      <div className="ecos-stored-cvs-table" tabIndex={0} role="region"
        aria-label={t("digitalCenters.cvs.rawSource")}>
        <table><thead><tr><th>CV</th><th>{t("digitalCenters.cvs.value")}</th></tr></thead>
          <tbody>{item.center.cvSources!.unverifiedValues!.map(value => <tr key={value.number}>
            <th scope="row">{value.number}</th><td>{value.value}</td>
          </tr>)}</tbody>
        </table>
      </div>
    </details>}
    <div role="status" aria-live="polite">
      <p>{comparedIncomplete ? t("digitalCenters.cvs.partialCount", { count: values.length })
        : database || backup ? t("digitalCenters.cvs.databaseCount", { count: values.length, indexed })
        : t("digitalCenters.cvs.count", { count: values.length, checked })}
        {state && ` · ${backup ? t("digitalCenters.backup.source")
          : complete ? t("digitalCenters.cvs.complete") : t("digitalCenters.cvs.partial")}`}</p>
      {!backup && (reading || externalReading || checked > 0) && <progress max={1024} value={reading || externalReading ? undefined : checked}
        aria-label={t("digitalCenters.cvs.progress")} />}
      {complete && !values.length && !item.center.cvSources?.conflicts.length &&
        <p>{t("digitalCenters.cvs.unavailable")}</p>}
      {result && <p>{t("digitalCenters.cvs.imported", result)}</p>}
    </div>
    {(error || externalError) && <p role="alert">{error || externalError}</p>}
    {values.length > 0 && <details>
      <summary>{t("digitalCenters.cvs.showValues")}</summary>
      {state && <p>{state.sourceHost} · #{item.centerObjectId} · {backup
        ? t("digitalCenters.backup.checkedAt", { time: state.readAt }) : state.readAt}</p>}
      <div className="ecos-stored-cvs-table" tabIndex={0} role="region"
        aria-label={t("digitalCenters.cvs.showValues")}><table>
        <thead><tr><th scope="col">CV</th><th scope="col">CV31 / CV32</th><th scope="col">{t("digitalCenters.cvs.value")}</th></tr></thead>
        <tbody>{values.map((value) => <tr key={`${value.number}:${value.indexCV31}:${value.indexCV32}`}>
          <th scope="row">{value.number}</th>
          <td>{value.indexCV31 == null ? (value.number > 256 ? "?" : "–") : `${value.indexCV31} / ${value.indexCV32}`}</td>
          <td>{value.value}</td>
        </tr>)}</tbody>
      </table></div>
    </details>}
    {values.length > 0 && <p>{t(backup ? "digitalCenters.backup.preserveHint" : "digitalCenters.cvs.preserveHint")}</p>}
  </section>;
}
