import { useEffect, useState } from "react";
import { ExternalLink } from "lucide-react";
import { api } from "../../shared/api";
import type { DecoderCVReference } from "../../shared/decoderLibraryModel";
import { useI18n } from "../../shared/i18n";

export function DecoderManufacturerReferences({ modelId }: { modelId: string }) {
  const { t, language } = useI18n();
  const [references, setReferences] = useState<DecoderCVReference[]>([]);
  const [status, setStatus] = useState<"loading" | "ready" | "error">("loading");
  const [search, setSearch] = useState("");
  const [retry, setRetry] = useState(0);
  useEffect(() => {
    let active = true;
    setStatus("loading");
    api.decoderReferences(modelId).then(value => {
      if (active) { setReferences(value); setStatus("ready"); }
    }).catch(() => { if (active) setStatus("error"); });
    return () => { active = false; };
  }, [modelId, retry]);
  const query = search.trim().toLocaleLowerCase();
  const cv = query.match(/^(?:cv\s*)?(\d+)$/);
  return <section className="decoder-manufacturer-references" aria-label={t("decoderLibrary.referencesTitle")}>
    <h5>{t("decoderLibrary.referencesTitle")}</h5>
    <p className="decoder-library-note">{t("decoderLibrary.referencesScope")}</p>
    {status === "loading" && <p role="status">{t("decoderLibrary.loading")}</p>}
    {status === "error" && <div role="alert"><p>{t("decoderLibrary.referencesFailed")}</p>
      <button type="button" className="secondary-button" onClick={() => setRetry(n => n + 1)}>{t("settings.data.refresh")}</button></div>}
    {status === "ready" && <>
      {!references.length && <p>{t("decoderLibrary.referencesMissing")}</p>}
      {references.some(reference => reference.rows.length > 0) && <label className="settings-search">
        {t("decoderLibrary.referencesSearch")}<input type="search" value={search} onChange={event => setSearch(event.target.value)} />
      </label>}
      {references.map(reference => {
        const rows = reference.rows.filter(row => cv ? row.cv <= Number(cv[1]) && Number(cv[1]) <= (row.cvEnd ?? row.cv)
          : `${row.cv} ${row.label} ${row.valueRange ?? ""}`.toLocaleLowerCase().includes(query));
        return <article key={reference.id} className="decoder-source-reference">
          <h5>{reference.title[language]}</h5>
          <p>{t(`decoderLibrary.referenceStatus.${reference.status}`)} · {reference.rows.length} {t("decoderLibrary.referenceRows")}</p>
          <p className="decoder-library-note">{reference.note[language]}</p>
          {reference.sources.map(source => <p key={source.url}><a href={source.url} target="_blank" rel="noreferrer">
            {source.title}<ExternalLink size={13} aria-hidden="true" /></a><small>{source.version} · {source.scope}</small></p>)}
          {rows.length > 0 && <div className="table-wrap decoder-library-table"><table><thead><tr>
            <th>CV</th><th>{t("vehicles.cv.index")}</th><th>{t("vehicles.cv.description")}</th>
            <th>{t("decoderLibrary.valueRange")}</th><th>{t("decoderLibrary.documentedDefault")}</th>
          </tr></thead><tbody>{rows.map((row, index) => {
            const source = reference.sources[row.sourceIndex];
            return <tr key={`${row.cv}-${row.page}-${index}`}>
              <td><strong>{row.cv}{row.cvEnd && `–${row.cvEnd}`}</strong></td>
              <td>{row.indexScope === "indexed" ? `CV31=${row.indexCV31} / CV32=${row.indexCV32}`
                : t(row.indexScope === "base" ? "vehicles.cv.base" : "decoderLibrary.indexUnspecified")}</td>
              <td><strong>{row.label}</strong><small>{t(row.applicability === "manual"
                ? "decoderLibrary.manualApplicability" : "decoderLibrary.modelApplicability")}</small>
                {source && <a href={`${source.url}${row.page > 0 ? `#page=${row.page}` : ""}`} target="_blank" rel="noreferrer">
                  {row.page > 0 ? t("decoderLibrary.sourcePage", { page: row.page }) : t("decoderLibrary.tab.sources")}</a>}</td>
              <td>{row.valueRange || "—"}</td><td>{row.factoryValue || "—"}</td>
            </tr>;
          })}</tbody></table></div>}
          {reference.rows.length > 0 && !rows.length && <p>{t("decoderLibrary.noDefinitions")}</p>}
        </article>;
      })}
    </>}
  </section>;
}
