import { useEffect, useState } from "react";
import { BookOpen, ChevronRight, FileCheck2, RefreshCw, Search } from "lucide-react";
import { api } from "../../shared/api";
import type { DecoderLibrary, DecoderLibraryProfile } from "../../shared/decoderLibraryModel";
import { decoderLibraryChanged } from "../../shared/DecoderLibraryProvider";
import { formatDate, translate, useI18n } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";
import { directoryModels, modelContext } from "./decoderLibraryDirectory";
import { DecoderLibraryModelDetail } from "./DecoderLibraryModelDetail";

export function DecoderLibrarySettings({ roles }: { roles: string[] }) {
  const { t, language } = useI18n();
  const [catalog, setCatalog] = useState<DecoderLibrary>({ schemaVersion: 1, profiles: [] });
  const [manufacturer, setManufacturer] = useState("");
  const [family, setFamily] = useState("");
  const [modelId, setModelId] = useState("");
  const [standard, setStandard] = useState(false);
  const [search, setSearch] = useState("");
  const [kind, setKind] = useState("");
  const [coverage, setCoverage] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");
  const [refresh, setRefresh] = useState(0);
  useEffect(() => {
    let active = true;
    setLoading(true); setMessage("");
    api.decoderLibrary(true).then(value => { if (active) setCatalog(value); })
      .catch(() => { if (active) setMessage(translate(language, "decoderLibrary.loadFailed")); })
      .finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, [refresh, language]);
  const { profiles, inventory, referenceCoverage } = catalog;
  const selected = inventory?.models.find(model => model.id === modelId);
  const visible = inventory ? directoryModels(inventory, profiles, { manufacturer, family, kind, coverage, search }) : [];
  const selectScope = (maker = "", group = "") => {
    setManufacturer(maker); setFamily(group); setModelId(""); setStandard(false);
  };
  const toggle = async (profile: DecoderLibraryProfile) => {
    setSaving(true); setMessage("");
    try {
      await api.setMasterDataActive("decoder_profile", profile.id, !profile.active);
      setCatalog(value => ({ ...value, profiles: value.profiles.map(p => p.id === profile.id ? { ...p, active: !p.active } : p) }));
      window.dispatchEvent(new Event(decoderLibraryChanged));
    } catch { setMessage(t("decoderLibrary.saveFailed")); }
    finally { setSaving(false); }
  };
  return <section className="decoder-library-settings" aria-label={t("decoderLibrary.title")}>
    <header className="master-data-head"><div><h3><BookOpen size={19} aria-hidden="true" />{t("decoderLibrary.title")}</h3>
      <p>{t("decoderLibrary.directoryIntro")}</p></div>
      <button type="button" className="icon-button" disabled={loading || saving} onClick={() => {
        setRefresh(n => n + 1); window.dispatchEvent(new Event(decoderLibraryChanged));
      }} aria-label={t("settings.data.refresh")}><RefreshCw size={16} /></button>
    </header>
    {message && <p role="alert" className="form-message">{message}</p>}
    {loading ? <p role="status">{t("decoderLibrary.loading")}</p> : !profiles.length && !inventory
      ? <p className="empty-state">{t("decoderLibrary.empty")}</p> : <>
      <div className="decoder-directory-summary">
        <span>{t("decoderLibrary.counts", { manufacturers: inventory?.manufacturers.length || 0,
          families: inventory?.families.length || 0, models: inventory?.models.length || 0 })}</span>
        <span>{inventory && `${t("decoderLibrary.reviewedAt")}: ${formatDate(inventory.reviewedAt, language)} · ${inventory.revision}`}</span>
      </div>
      <p className="decoder-directory-hint">{t("decoderLibrary.inventoryScope")}</p>
      <div className="decoder-directory-layout">
        <div className="decoder-directory-mobile">
          <label><span>{t("decoderLibrary.manufacturer")}</span>
            <AppSelect value={standard ? "standard" : manufacturer} aria-label={t("decoderLibrary.manufacturer")}
              onChange={event => {
                if (event.target.value === "standard") { setStandard(true); setModelId(""); }
                else selectScope(event.target.value);
              }}>
              <option value="">{t("decoderLibrary.allModels")}</option>
              {inventory?.manufacturers.map(maker => <option key={maker.id} value={maker.id}>{maker.name}</option>)}
              <option value="standard">{t("decoderLibrary.foundation")}</option>
            </AppSelect></label>
          <label><span>{t("decoderLibrary.family")}</span><AppSelect value={family} disabled={!manufacturer || standard}
            aria-label={t("decoderLibrary.family")} onChange={event => selectScope(manufacturer, event.target.value)}>
            <option value="">{t("decoderLibrary.allFamilies")}</option>
            {inventory?.families.filter(group => group.manufacturerId === manufacturer)
              .map(group => <option key={group.id} value={group.id}>{group.name}</option>)}
          </AppSelect></label>
        </div>
        <nav className="decoder-directory-nav" aria-label={t("decoderLibrary.navigation")}>
          <button type="button" aria-current={!manufacturer && !standard ? "page" : undefined}
            onClick={() => selectScope()}><BookOpen size={16} aria-hidden="true" />{t("decoderLibrary.allModels")}<small>{inventory?.models.length || 0}</small></button>
          <h4>{t("decoderLibrary.manufacturers")}</h4>
          {inventory?.manufacturers.map(maker => <div key={maker.id}>
            <button type="button" aria-current={manufacturer === maker.id && !family && !standard ? "page" : undefined}
              aria-expanded={manufacturer === maker.id} onClick={() => selectScope(maker.id)}>
              <ChevronRight size={14} className={manufacturer === maker.id ? "expanded" : ""} aria-hidden="true" />{maker.name}
              <small>{inventory.models.filter(model => modelContext(inventory, model).maker?.id === maker.id).length}</small></button>
            {manufacturer === maker.id && <div className="decoder-directory-families">
              {inventory.families.filter(group => group.manufacturerId === maker.id).map(group => <button key={group.id} type="button"
                aria-current={family === group.id && !standard ? "page" : undefined} onClick={() => selectScope(maker.id, group.id)}>
                {group.name}<small>{inventory.models.filter(model => model.familyId === group.id).length}</small></button>)}
            </div>}
          </div>)}
          <h4>{t("decoderLibrary.standards")}</h4>
          <button type="button" aria-current={standard ? "page" : undefined}
            onClick={() => { setStandard(true); setModelId(""); }}><FileCheck2 size={16} aria-hidden="true" />{t("decoderLibrary.foundation")}</button>
        </nav>
        <div className="decoder-directory-content">
          {selected || standard ? <DecoderLibraryModelDetail key={selected?.id || "standard"} model={selected} inventory={inventory}
            profiles={profiles} roles={roles} saving={saving} onToggle={profile => void toggle(profile)}
            onBack={() => { setModelId(""); setStandard(false); }} /> : <>
            <div className="decoder-directory-title"><h4>{family ? inventory?.families.find(group => group.id === family)?.name
              : manufacturer ? inventory?.manufacturers.find(maker => maker.id === manufacturer)?.name : t("decoderLibrary.allModels")}</h4>
              <span>{t(visible.length === 1 ? "decoderLibrary.resultCountOne" : "decoderLibrary.resultCount", { count: visible.length })}</span></div>
            {inventory?.manufacturers.find(maker => maker.id === manufacturer)?.coverage &&
              <p className="decoder-library-note">{inventory.manufacturers.find(maker => maker.id === manufacturer)?.coverage?.[language]}</p>}
            <div className="decoder-directory-toolbar">
              <label className="decoder-directory-search settings-search"><span>{t("decoderLibrary.modelSearch")}</span><div><Search size={16} aria-hidden="true" />
                <input type="search" value={search} onChange={event => setSearch(event.target.value)} /></div></label>
              <label><span>{t("decoderLibrary.kind")}</span><AppSelect value={kind} onChange={event => setKind(event.target.value)} aria-label={t("decoderLibrary.kind")}>
                <option value="">{t("decoderLibrary.allKinds")}</option>{["motor", "sound", "function"].map(value => <option key={value} value={value}>{t(`decoderLibrary.kind.${value}`)}</option>)}
              </AppSelect></label>
              <label><span>{t("decoderLibrary.coverageLabel")}</span><AppSelect value={coverage} onChange={event => setCoverage(event.target.value)} aria-label={t("decoderLibrary.coverageLabel")}>
                <option value="">{t("decoderLibrary.allCoverage")}</option><option value="reviewed">{t("decoderLibrary.cvReviewed")}</option>
                <option value="basic">{t("decoderLibrary.basicsOnlyShort")}</option></AppSelect></label>
            </div>
            <p className="decoder-directory-hint">{t("decoderLibrary.coverageHint")}</p>
            {visible.length ? <div className="table-wrap decoder-directory-table"><table><thead><tr>
              <th>{t("decoderLibrary.model")}</th><th>{t("decoderLibrary.interfaces")}</th><th>{t("decoderLibrary.articleNumbers")}</th><th>{t("decoderLibrary.coverageLabel")}</th>
            </tr></thead><tbody>{visible.map(model => {
              const { family: group, maker } = inventory ? modelContext(inventory, model) : {};
              const profile = profiles.find(item => item.id === (model.profileId || model.documentationProfileId));
              const hasReference = referenceCoverage?.some(item => item.active && item.rowCount > 0 && item.modelIds.includes(model.id));
              const referenceGap = referenceCoverage?.some(item => item.active && item.status === "unavailable" && item.modelIds.includes(model.id));
              return <tr key={model.id}><td><button type="button" className="decoder-model-link" onClick={() => setModelId(model.id)}>{model.name}<ChevronRight size={14} aria-hidden="true" /></button>
                <small>{maker?.name} · {group?.name} · {t(`decoderLibrary.kind.${model.kind}`)}</small>
                {model.documentStatus && <small>{t(`decoderLibrary.documentStatus.${model.documentStatus}`)}</small>}</td>
                <td>{model.interfaces.join(", ") || t("decoderLibrary.notRecorded")}</td><td>{model.articleNumbers.join(", ") || "—"}
                  {!!model.variants?.length && <button type="button" className="decoder-variant-link" onClick={() => setModelId(model.id)}>
                    {t(model.variants.length === 1 ? "decoderLibrary.variantCountOne" : "decoderLibrary.variantCount",
                      { count: model.variants.length })}</button>}</td>
                <td><span className="decoder-coverage" data-reviewed={!!profile}>{t(profile ? "decoderLibrary.cvReviewed" : "decoderLibrary.basicsOnlyShort")}</span>
                  {hasReference && <small>{t("decoderLibrary.referencesAvailable")}</small>}
                  {referenceGap && <small>{t("decoderLibrary.referenceStatus.unavailable")}</small>}
                  {profile && !profile.active && <small>{t("decoderLibrary.inactive")}</small>}</td></tr>;
            })}</tbody></table></div> : <div className="empty-state"><p>{t("decoderLibrary.noModels")}</p>
              <button type="button" className="secondary-button" onClick={() => { setSearch(""); setKind(""); setCoverage(""); selectScope(); }}>{t("decoderLibrary.resetFilters")}</button></div>}
          </>}
        </div>
      </div>
    </>}
  </section>;
}
