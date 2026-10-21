import { useState } from "react";
import type { DecoderLibraryProfile } from "../../shared/decoderLibraryModel";
import { useI18n } from "../../shared/i18n";
import { decoderCVMatches } from "../../shared/decoderLibrary";

export function DecoderLibraryDefinitions({ profile, profiles }: {
  profile?: DecoderLibraryProfile; profiles: DecoderLibraryProfile[];
}) {
  const { t, language } = useI18n();
  const [search, setSearch] = useState("");
  const foundation = profiles.find(p => p.id === "dcc-basics");
  // Model definitions override the same base address; both origins remain visible in the source column.
  const own = profile?.id === "dcc-basics" ? [] : (profile?.definitions || []);
  const modelOnly = profile?.rules?.modelDefinitionsOnly || profile?.rules?.roco10882 || profile?.rules?.uhlenbrock76420 || profile?.rules?.piko56500;
  const definitions = [
    ...((modelOnly ? [] : foundation?.definitions)?.filter(d => !own.some(item => item.cv === d.cv && item.indexCV31 === d.indexCV31 && item.indexCV32 === d.indexCV32)) || [])
      .map(definition => ({ definition, origin: foundation })),
    ...own.map(definition => ({ definition, origin: profile }))
  ].sort((a, b) => a.definition.cv - b.definition.cv);
  const query = search.trim().toLocaleLowerCase(), cv = query.match(/^(?:cv\s*)?(\d+)$/);
  const visible = definitions.filter(({ definition: d }) => cv
    ? decoderCVMatches(d, Number(cv[1]))
    : `${d.name[language]} ${d.help[language]} ${d.cv} ${d.indexCV31 ?? ""} ${d.indexCV32 ?? ""}`.toLocaleLowerCase().includes(query));
  return <>
    <p className="decoder-library-note">{t("decoderLibrary.referenceOnly")}</p>
    {!own.length && profile?.id !== "dcc-basics" && <p className="decoder-library-note">{t("decoderLibrary.basicOnly")}</p>}
    <label className="settings-search">{t("decoderLibrary.search")}<input type="search" value={search} onChange={event => setSearch(event.target.value)} /></label>
    <div className="table-wrap decoder-library-table"><table><thead><tr>
      <th>CV</th><th>{t("vehicles.cv.index")}</th><th>{t("vehicles.cv.description")}</th><th>{t("decoderLibrary.valueRange")}</th>
    </tr></thead><tbody>{visible.map(({ definition: d, origin }, i) => <tr key={`${d.cv}-${d.indexCV31}-${d.indexCV32}-${i}`}>
      <td><strong>{d.cv}{d.cvEnd && `–${d.cvEnd}`}</strong>
        {d.cvStride && <small>{t("decoderLibrary.cvStride", { stride: d.cvStride })}</small>}</td>
      <td>{d.indexCV31 == null ? t("vehicles.cv.base") : `CV31=${d.indexCV31} / CV32=${d.indexCV32}${d.indexCV32End != null ? `–${d.indexCV32End}` : ""}`}</td>
      <td><strong>{d.name[language]}</strong><p>{d.help[language]}</p>
        <small className="decoder-definition-origin">{language === "en" ? origin?.nameEn : origin?.name} · {origin?.revision}
          {origin && !origin.active && ` · ${t("decoderLibrary.inactive")}`}</small>
        {d.firmwareMin && <small>{t("decoderLibrary.minimumFirmware", { minimum: d.firmwareMin })}</small>}
        {!!d.bits?.length && <details><summary>{t("vehicles.decoder.individualSettings")}</summary>
          <dl>{d.bits.map(bit => <div key={bit.bit}><dt>Bit {bit.bit} · {bit.name[language]}</dt>
            <dd>0: {bit.off[language]} · 1: {bit.on[language]}</dd></div>)}</dl>
        </details>}</td><td className="decoder-definition-values">{d.reserved ? t("decoderLibrary.reservedRange") : `${d.minimum}–${d.maximum}`}
        {d.unit && <small>{d.unit[language]}</small>}
        {d.factoryDefault != null && <small>{t("decoderLibrary.documentedDefault")}: {d.factoryDefault}</small>}
        {d.reference && origin && <small><a href={origin.sources[d.reference.sourceIndex].url}
          target="_blank" rel="noreferrer">{d.reference.location}</a></small>}
      </td>
    </tr>)}</tbody></table></div>
    {!visible.length && <p className="empty-state">{t("decoderLibrary.noDefinitions")}</p>}
  </>;
}
