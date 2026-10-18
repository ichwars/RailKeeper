import { Fragment, useState } from "react";
import { CircleHelp, Pencil, Trash2 } from "lucide-react";
import type { VehicleCVValue } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";
import { decoderAnalysis } from "./decoderAnalysis";
import { cvValueInfo } from "./cvValueInfo";
import { CVPageControls, cvIndexLabel, cvPageSize } from "./CVPageControls";
import { formatDateTime } from "./vehicleFormat";

export function VehicleCVValuesTable({ values, readonly, saving, edit, remove, onExplain }: {
  values: VehicleCVValue[]; readonly: boolean; saving: boolean;
  edit: (value: VehicleCVValue) => void; remove: (value: VehicleCVValue) => void;
  onExplain?: (value: VehicleCVValue) => void;
}) {
  const { language, t } = useI18n();
  const [profile, setProfile] = useState("*");
  const [index, setIndex] = useState("*");
  const [search, setSearch] = useState("");
  const [hideZero, setHideZero] = useState(false);
  const [requestedPage, setPage] = useState(0);
  const profiles = Array.from(new Set(values.map((value) => value.decoderProfile || "")));
  const profileValues = values.filter((value) => profile === "*" || (value.decoderProfile || "") === profile);
  const indexes = Array.from(new Map(profileValues.map((value) => [indexKey(value), value])).entries());
  const analyses = new Map(profiles.map(profile => [profile, decoderAnalysis(values, profile)]));
  const info = (value: VehicleCVValue) => cvValueInfo(value,
    analyses.get(value.decoderProfile || "") || decoderAnalysis([], ""), t);
  const meaning = (value: VehicleCVValue) => value.description || info(value).name;
  const cvSearch = search.trim().match(/^CV\s*(\d+)$/i);
  const filtered = profileValues.filter((value) => (index === "*" || indexKey(value) === index) &&
    (!hideZero || value.value !== 0) && (cvSearch ? value.cvNumber === Number(cvSearch[1]) :
      `${value.cvNumber} ${value.value} ${meaning(value)} ${info(value).name} ${value.category || ""}`.toLowerCase().includes(search.toLowerCase())));
  const page = Math.min(requestedPage, Math.max(0, Math.ceil(filtered.length / cvPageSize) - 1));
  return <>
    <div className="cv-value-filters">
      <label>{t("vehicle.field.decoderProfile")}<AppSelect value={profile} onChange={(event) => { setProfile(event.target.value); setIndex("*"); setPage(0); }}>
        <option value="*">{t("vehicles.cv.allProfiles")}</option>
        {profiles.map((profile) => <option key={profile} value={profile}>{profile || t("vehicles.cv.noProfile")}</option>)}
      </AppSelect></label>
      <label>{t("vehicles.cv.index")}<AppSelect value={index} onChange={(event) => { setIndex(event.target.value); setPage(0); }}>
        <option value="*">{t("vehicles.cv.allPages")}</option>
        {indexes.map(([key, value]) => <option key={key} value={key}>{cvIndexLabel(value, t("vehicles.cv.base"))}</option>)}
      </AppSelect></label>
      <label>{t("common.search")}<input type="search" placeholder="CV29 / Text" value={search} onChange={(event) => { setSearch(event.target.value); setPage(0); }} /></label>
    </div>
    <label className="cv-zero-filter"><input type="checkbox" checked={hideZero} onChange={(event) => { setHideZero(event.target.checked); setPage(0); }} />{t("vehicles.decoder.hideZero")}</label>
    <CVPageControls page={page} total={filtered.length} onChange={setPage} />
    <div className="table-wrap compact-table cv-values-table">
      <table><thead><tr>
        <th>CV</th><th>{t("vehicles.cv.description")}</th><th>{t("vehicle.field.value")}</th>
        <th>{t("vehicles.cv.index")}</th><th>{t("vehicle.field.decoderProfile")}</th><th>{t("vehicles.actions")}</th>
      </tr></thead><tbody>
        {filtered.slice(page * cvPageSize, (page + 1) * cvPageSize).map((value) => <Fragment key={value.id}>
          <tr><td>CV {value.cvNumber}</td>
            <td>{meaning(value)}{value.category && <small> · {value.category}</small>}
              {value.protocol && <small> · {value.protocol}</small>}</td>
            <td>{value.value}</td><td>{cvIndexLabel(value, t("vehicles.cv.base"))}</td><td>{value.decoderProfile || "-"}</td>
            <td><div className="table-actions">
              {onExplain && <button type="button" className="icon-button" aria-haspopup="dialog" onClick={() => onExplain(value)}
                aria-label={t("vehicles.decoder.explainValue", { number: value.cvNumber })}><CircleHelp size={15} /></button>}
              <button type="button" className="icon-button" onClick={() => edit(value)} disabled={readonly || saving} aria-label={t("vehicles.cv.edit")} title={t("vehicles.cv.edit")}><Pencil size={15} /></button>
              <button type="button" className="icon-button danger" onClick={() => remove(value)} disabled={readonly || saving} aria-label={t("vehicles.cv.delete")} title={t("vehicles.cv.delete")}><Trash2 size={15} /></button>
            </div></td>
          </tr>
          {!!value.history?.length && <tr className="cv-history-row"><td colSpan={6}><details>
            <summary>{t("vehicles.cv.history", { count: value.history.length, suffix: value.history.length === 1 ? "" : language === "de" ? "en" : "s" })}</summary>
            <div className="cv-history-list">{value.history.slice(0, 5).map((entry) => <span key={entry.id}>{formatDateTime(entry.changedAt)}: {entry.oldValue} → {entry.newValue}</span>)}</div>
          </details></td></tr>}
        </Fragment>)}
      </tbody></table>
      {!filtered.length && <p className="empty-state compact">{t("vehicles.cv.noMatch")}</p>}
    </div>
  </>;
}

function indexKey(value: VehicleCVValue) {
  return value.indexCV31 == null ? "base" : `${value.indexCV31}/${value.indexCV32}`;
}
