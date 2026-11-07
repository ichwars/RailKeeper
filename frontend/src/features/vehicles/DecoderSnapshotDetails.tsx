import { Fragment, useEffect, useState } from "react";
import { Download } from "lucide-react";
import { api, type DecoderCVSnapshot, type DecoderRecordedCVValue } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { CVPageControls, cvIndexLabel, cvPageSize } from "./CVPageControls";
import { formatDateTime, formatFileSize } from "./vehicleFormat";
import { decoderProfileLabel } from "./decoderProfileLabel";

function downloadSnapshot(snapshot: DecoderCVSnapshot) {
  const url = URL.createObjectURL(new Blob([JSON.stringify(snapshot, null, 2)], { type: "application/json" }));
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = `railkeeper-cv-snapshot-${snapshot.recordedAt.slice(0, 10)}.json`;
  anchor.click();
  URL.revokeObjectURL(url);
}

export function DecoderSnapshotDetails({ snapshotId }: { snapshotId: string }) {
  const { t } = useI18n();
  const [snapshot, setSnapshot] = useState<DecoderCVSnapshot | null>(null);
  const [failed, setFailed] = useState(false);
  useEffect(() => {
    let active = true;
    setSnapshot(null); setFailed(false);
    void api.decoderCVSnapshot(snapshotId).then(value => {
      if (active) { if (value.id === snapshotId) setSnapshot(value); else setFailed(true); }
    }).catch(() => { if (active) setFailed(true); });
    return () => { active = false; };
  }, [snapshotId]);
  if (failed) return <p className="form-message" role="alert">{t("decoderConfig.detailFailed")}</p>;
  if (!snapshot || snapshot.id !== snapshotId) return <p role="status">{t("decoderConfig.loading")}</p>;
  const hardware = snapshot.hardware;
  return <section className="decoder-archive-detail" aria-label={t("decoderConfig.snapshot")}>
    <header className="decoder-archive-detail-head"><div>
      <h3>{decoderProfileLabel(snapshot.profile, t("decoderConfig.noProfile"))}</h3>
      <p>{t("decoderConfig.recorded", { date: formatDateTime(snapshot.recordedAt) })}</p>
      <span className="decoder-archive-partial">{t("decoderConfig.partial")}</span>
    </div><button type="button" className="secondary-button" onClick={() => downloadSnapshot(snapshot)}>
      <Download size={16} aria-hidden="true" />{t("decoderConfig.export")}</button></header>
    <p className="decoder-archive-scope">{t("decoderConfig.exportHint")}</p>
    <dl className="decoder-archive-metadata">
      <div><dt>{t("decoderConfig.historicalVehicle")}</dt><dd>{snapshot.vehicleName}
        {snapshot.vehicleInventoryNumber && <small>{snapshot.vehicleInventoryNumber}</small>}</dd></div>
      <div><dt>{t("decoderConfig.hardware")}</dt><dd>{hardware ? <>{hardware.decoderLabel}
        <small>{t(`decoderConfig.state.${hardware.installationState}`)}</small></> : t("decoderConfig.unassigned")}</dd></div>
      {hardware && <>
        <div><dt>{t("decoderConfig.manufacturer")}</dt><dd>{hardware.manufacturer || t("decoderConfig.unknown")}</dd></div>
        <div><dt>{t("decoderConfig.model")}</dt><dd>{hardware.model || t("decoderConfig.unknown")}</dd></div>
        <div><dt>{t("decoderConfig.serial")}</dt><dd>{hardware.serialNumber || t("decoderConfig.unknown")}</dd></div>
        <div><dt>{t("decoderConfig.installedAt")}</dt><dd>{hardware.installedAt || t("decoderConfig.unknown")}</dd></div>
        <div><dt>{t("decoderConfig.removedAt")}</dt><dd>{hardware.removedAt || t("decoderConfig.unknown")}</dd></div>
      </>}
    </dl>
    {snapshot.note && <div className="decoder-archive-note"><h4>{t("decoderConfig.note")}</h4><p>{snapshot.note}</p></div>}
    <section className="decoder-archive-sources"><h4>{t("decoderConfig.sources")}</h4>
      {!snapshot.sources.length && <p>{t("decoderConfig.noSources")}</p>}
      {snapshot.sources.map(source => <div key={source.sourceFileId} className="decoder-archive-source">
        <div><strong>{source.metadata.originalName || t("decoderConfig.unavailable")}</strong>
          <small>{t(source.contentState === "stored" ? "decoderConfig.available" : "decoderConfig.unavailable")}
            {source.metadata.sizeBytes > 0 && ` · ${formatFileSize(source.metadata.sizeBytes)}`}</small></div>
        {source.contentState === "stored" && <a className="secondary-button"
          href={api.decoderSnapshotSourceUrl(snapshot.id, source.sourceFileId)}>
          <Download size={15} aria-hidden="true" />{t("decoderConfig.downloadSource")}</a>}
      </div>)}
    </section>
    <RecordedCVValues snapshot={snapshot} />
  </section>;
}

function RecordedCVValues({ snapshot }: { snapshot: DecoderCVSnapshot }) {
  const { t } = useI18n();
  const [search, setSearch] = useState("");
  const [requestedPage, setPage] = useState(0);
  const sources = new Map(snapshot.sources.map(source => [source.sourceFileId, source]));
  const rows = snapshot.payload.values.filter(value => `${value.cvNumber} ${value.value} ${value.description || ""} ${value.category || ""}`
    .toLowerCase().includes(search.toLowerCase()));
  const page = Math.min(requestedPage, Math.max(0, Math.ceil(rows.length / cvPageSize) - 1));
  const displayValue = (value: number) => value < 0 || value > 255
    ? t("decoderConfig.unknownValue", { value }) : String(value);
  const provenance = (value: DecoderRecordedCVValue) => {
    const source = value.sourceFileId ? sources.get(value.sourceFileId) : undefined;
    return <>{value.sourceFileId ? source?.metadata.originalName || t("decoderConfig.unavailable") : t("decoderConfig.noFileSource")}
      {value.sourceLine != null && <small>{t("decoderConfig.sourceLine", { line: value.sourceLine })}</small>}
      {value.sourceExportedAt && <small>{t("decoderConfig.sourceTime", { date: value.sourceExportedAt.replace("T", " ") })}</small>}</>;
  };
  return <section className="decoder-archive-values">
    <label className="decoder-archive-search app-text-input">{t("common.search")}<input type="search" value={search}
      onChange={event => { setSearch(event.target.value); setPage(0); }} /></label>
    <CVPageControls page={page} total={rows.length} onChange={setPage} />
    <div className="table-wrap compact-table"><table><thead><tr>
      <th>CV</th><th>{t("vehicle.field.value")}</th><th>{t("vehicles.cv.index")}</th>
      <th>{t("vehicles.cv.description")}</th><th>{t("decoderConfig.source")}</th>
    </tr></thead><tbody>{rows.slice(page * cvPageSize, (page + 1) * cvPageSize).map(value => <Fragment key={value.id}>
      <tr><td>CV {value.cvNumber}</td><td>{displayValue(value.value)}</td>
        <td>{cvIndexLabel(value, t("decoderConfig.noIndex"))}</td>
        <td>{value.description}{value.category && <small>{value.category}</small>}{value.protocol && <small>{value.protocol}</small>}</td>
        <td>{provenance(value)}</td></tr>
      {!!value.history?.length && <tr><td colSpan={5}><details>
        <summary>{t("decoderConfig.history", { count: value.history.length })}</summary>
        <ol>{value.history.map(entry => <li key={entry.id}>{formatDateTime(entry.changedAt)}: {displayValue(entry.oldValue)} → {displayValue(entry.newValue)}</li>)}</ol>
      </details></td></tr>}
    </Fragment>)}</tbody></table>
      {!rows.length && <p className="empty-state compact">{t("vehicles.cv.noMatch")}</p>}
    </div>
  </section>;
}
