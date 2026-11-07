import { useEffect, useMemo, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { Archive, Check, RefreshCw, X } from "lucide-react";
import { api, type DecoderCVSnapshotSummary } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { useModalDialogLayer } from "../../shared/ui/useModalDialogLayer";
import { DecoderSnapshotDetails } from "./DecoderSnapshotDetails";
import { decoderProfileLabel } from "./decoderProfileLabel";
import { formatDateTime } from "./vehicleFormat";

export function DecoderCVArchive({ vehicleId }: { vehicleId?: string }) {
  const { t } = useI18n();
  const [records, setRecords] = useState<DecoderCVSnapshotSummary[]>([]);
  const [loading, setLoading] = useState(true), [failed, setFailed] = useState(false);
  const [revision, reload] = useState(0), [search, setSearch] = useState("");
  const [requestedPage, setPage] = useState(0), [requestedId, select] = useState<string | null>(null);
  useEffect(() => {
    let active = true;
    setLoading(true); setFailed(false); setRecords([]); select(null); setPage(0);
    void api.decoderCVSnapshots(vehicleId).then(values => { if (active) { setRecords(values); setLoading(false); } })
      .catch(() => { if (active) { setFailed(true); setLoading(false); } });
    return () => { active = false; };
  }, [vehicleId, revision]);
  const filtered = useMemo(() => records.filter(record =>
    `${record.vehicleName} ${record.vehicleInventoryNumber} ${record.profile} ${record.note}`.toLowerCase().includes(search.toLowerCase())), [records, search]);
  const page = Math.min(requestedPage, Math.max(0, Math.ceil(filtered.length / 50) - 1));
  const visible = filtered.slice(page * 50, (page + 1) * 50);
  const selected = visible.find(record => record.id === requestedId) || visible[0];
  return <section className="decoder-archive" aria-label={t("decoderConfig.archive")} onKeyDown={event => {
    if (event.key === "Enter" && event.target instanceof HTMLInputElement && !event.nativeEvent.isComposing) event.preventDefault();
  }}>
    <p className="decoder-archive-intro">{t("decoderConfig.archiveIntro")}</p>
    <p className="decoder-archive-scope">{t("decoderConfig.archiveScope")}</p>
    <div className="decoder-archive-filterbar"><label className="app-text-input">{t("decoderConfig.search")}
      <input type="search" value={search} onChange={event => { setSearch(event.target.value); setPage(0); select(null); }} /></label>
      <button type="button" className="secondary-button" disabled={loading} onClick={() => reload(value => value + 1)}>
        <RefreshCw size={16} aria-hidden="true" />{t("decoderConfig.reload")}</button></div>
    {loading && <p role="status">{t("decoderConfig.loading")}</p>}
    {failed && <p className="form-message" role="alert">{t("decoderConfig.loadFailed")}</p>}
    {!loading && !failed && !records.length && <p className="empty-state">{t("decoderConfig.empty")}</p>}
    {!loading && !failed && records.length > 0 && <>
      <div className="decoder-archive-pages"><span role="status">{t("decoderConfig.pageSnapshots", {
        from: filtered.length ? page * 50 + 1 : 0, to: Math.min((page + 1) * 50, filtered.length), total: filtered.length
      })}</span><div>
        <button type="button" className="secondary-button" disabled={page === 0} onClick={() => setPage(page - 1)}>{t("vehicles.cv.previous")}</button>
        <button type="button" className="secondary-button" disabled={(page + 1) * 50 >= filtered.length} onClick={() => setPage(page + 1)}>{t("vehicles.cv.next")}</button>
      </div></div>
      {!filtered.length && <p className="empty-state">{t("decoderConfig.noMatch")}</p>}
      <div className="decoder-archive-layout"><div className="decoder-archive-list">
        {visible.map(record => <button type="button" key={record.id} className="decoder-archive-record"
          aria-pressed={selected?.id === record.id} onClick={() => select(record.id)}>
          <strong>{decoderProfileLabel(record.profile, t("decoderConfig.noProfile"))}
            {selected?.id === record.id && <Check size={16} aria-hidden="true" />}</strong>
          <span>{record.vehicleName} · {record.vehicleInventoryNumber}</span>
          <small>{formatDateTime(record.recordedAt)} · {t("decoderConfig.values", { count: record.valueCount })}</small>
          <small>{t(`decoderConfig.kind.${record.kind}`)}</small>
        </button>)}
      </div>{selected && <DecoderSnapshotDetails key={`${selected.id}:${revision}`} snapshotId={selected.id} />}</div>
    </>}
  </section>;
}

export function DecoderArchiveDialog({ onClose }: { onClose: () => void }) {
  const { t } = useI18n();
  const closeRef = useRef<HTMLButtonElement>(null);
  const { anchorRef, layerRef, onKeyDown } = useModalDialogLayer(onClose, closeRef);
  useEffect(() => {
    const escape = (event: KeyboardEvent) => {
      if (event.key === "Escape" && !event.defaultPrevented && document.activeElement === document.body) {
        event.preventDefault(); onClose();
      }
    };
    window.addEventListener("keydown", escape);
    return () => window.removeEventListener("keydown", escape);
  }, [onClose]);
  return <><span ref={anchorRef} aria-hidden="true" />{createPortal(
    <div className="modal-layer decoder-archive-layer" role="dialog" aria-modal="true"
      aria-label={t("decoderConfig.archive")} ref={layerRef} onKeyDown={onKeyDown}>
      <section className="vehicle-modal decoder-archive-dialog"><header className="modal-head">
        <h2><Archive size={20} aria-hidden="true" />{t("decoderConfig.archive")}</h2>
        <button type="button" className="icon-button" ref={closeRef} onClick={onClose} aria-label={t("vehicles.close")}>
          <X size={18} aria-hidden="true" /></button></header>
        <div className="modal-body"><DecoderCVArchive /></div>
      </section>
    </div>, document.body)}</>;
}
