import { useId, useState } from "react";
import { ArrowLeft, ExternalLink, FileCheck2 } from "lucide-react";
import type { DecoderInventory, DecoderLibraryProfile, DecoderModel, DecoderSource } from "../../shared/decoderLibraryModel";
import { formatDate, useI18n } from "../../shared/i18n";
import { modelContext } from "./decoderLibraryDirectory";
import { DecoderLibraryDefinitions } from "./DecoderLibraryDefinitions";
import { DecoderManufacturerReferences } from "./DecoderManufacturerReferences";

export function DecoderLibraryModelDetail({ model, inventory, profiles, roles, saving, onToggle, onBack }: {
  model?: DecoderModel; inventory?: DecoderInventory; profiles: DecoderLibraryProfile[]; roles: string[];
  saving: boolean; onToggle: (profile: DecoderLibraryProfile) => void; onBack: () => void;
}) {
  const { t, language } = useI18n();
  const [tab, setTab] = useState("overview");
  const id = useId();
  const { family, maker } = model && inventory ? modelContext(inventory, model) : {};
  const profile = profiles.find(p => p.id === (model ? model.profileId || model.documentationProfileId : "dcc-basics"));
  const technicalMaker = inventory?.manufacturers.find(item => item.cv8 === profile?.manufacturerId);
  const sources = [...(profile?.sources || []), ...(model?.sources || []), ...(model?.variants?.map(v => v.source) || []),
    ...(family?.sources || []), ...(maker ? [maker.source] : [])]
    .filter((source, index, all) => all.findIndex(item => item.url === source.url) === index);
  const title = model?.name || t("decoderLibrary.foundation");
  const tabs = ["overview", "definitions", "identification", "sources"];
  const onTabKey = (event: React.KeyboardEvent<HTMLButtonElement>) => {
    const current = tabs.indexOf(tab);
    const next = event.key === "ArrowRight" ? (current + 1) % tabs.length : event.key === "ArrowLeft"
      ? (current + tabs.length - 1) % tabs.length : event.key === "Home" ? 0 : event.key === "End" ? tabs.length - 1 : -1;
    if (next < 0) return;
    event.preventDefault(); setTab(tabs[next]); document.getElementById(`${id}-${tabs[next]}`)?.focus();
  };
  return <section className="decoder-model-detail" aria-label={title}>
    <button type="button" className="decoder-back" onClick={onBack}><ArrowLeft size={15} aria-hidden="true" />{t("decoderLibrary.back")}</button>
    <header className="decoder-library-profile-head">
      <div><p className="decoder-breadcrumb">{maker ? `${maker.name} / ${family?.name}` : t("decoderLibrary.standards")}</p>
        <h4>{title}</h4><p>{maker && `${t(model?.documentationProfileId ? "decoderLibrary.registeredBrandCV8" : "decoderLibrary.registeredCV8", { value: maker.cv8 })} · `}{t("decoderLibrary.dccScope")}</p>
      {model?.documentationProfileId && profile?.manufacturerId != null && <p>{t("decoderLibrary.documentedTechnicalCV8", {
        manufacturer: technicalMaker?.name || profile.name, value: profile.manufacturerId
      })}</p>}</div>
      <span className="decoder-coverage" data-reviewed={!!profile}><FileCheck2 size={15} aria-hidden="true" />
        {t(profile ? "decoderLibrary.cvReviewed" : "decoderLibrary.dataReviewed")}</span>
    </header>
    <div className="decoder-detail-tabs" role="tablist" aria-label={t("decoderLibrary.modelDetails")}>
      {tabs.map(item => <button key={item} type="button" role="tab" id={`${id}-${item}`} aria-selected={tab === item}
        aria-controls={`${id}-panel`} tabIndex={tab === item ? 0 : -1} onKeyDown={onTabKey} onClick={() => setTab(item)}>
        {t(`decoderLibrary.tab.${item}`)}</button>)}
    </div>
    <div role="tabpanel" id={`${id}-panel`} aria-labelledby={`${id}-${tab}`} tabIndex={0}>
      {tab === "overview" && <>
        <dl className="decoder-model-facts">
          <div><dt>{t("decoderLibrary.manufacturer")}</dt><dd>{maker?.name || "NMRA / DCC"}</dd></div>
          <div><dt>{t("decoderLibrary.family")}</dt><dd>{family?.name || t("decoderLibrary.foundation")}</dd></div>
          <div><dt>{t("decoderLibrary.kind")}</dt><dd>{model ? t(`decoderLibrary.kind.${model.kind}`) : t("decoderLibrary.standard")}</dd></div>
          {!model?.variants?.length && <>
            <div><dt>{t("decoderLibrary.interfaces")}</dt><dd>{model?.interfaces.join(", ") || t("decoderLibrary.notRecorded")}</dd></div>
            <div><dt>{t("decoderLibrary.articleNumbers")}</dt><dd>{model?.articleNumbers.join(", ") || t("decoderLibrary.notRecorded")}</dd></div>
          </>}
          <div><dt>{t("decoderLibrary.reviewedAt")}</dt><dd>{inventory ? formatDate(inventory.reviewedAt, language) : profile?.revision}</dd></div>
        </dl>
        {model?.documentStatus && <p className="decoder-library-note">{t(`decoderLibrary.documentStatus.${model.documentStatus}`)}</p>}
        {!!model?.variants?.length && <section className="decoder-model-variants" aria-label={t("decoderLibrary.variants")}>
          <h5>{t("decoderLibrary.variants")}</h5>
          <div className="table-wrap decoder-variant-table"><table><thead><tr>
            <th>{t("decoderLibrary.articleNumber")}</th><th>{t("decoderLibrary.interfaces")}</th><th>{t("decoderLibrary.tab.sources")}</th>
          </tr></thead><tbody>{model.variants.map(variant => <tr key={variant.articleNumber}>
            <td>{variant.articleNumber}</td><td>{variant.interface}</td><td>
              <a href={variant.source.url} target="_blank" rel="noreferrer">{variant.source.title}<ExternalLink size={13} aria-hidden="true" /></a>
              <small>{variant.source.version}</small></td>
          </tr>)}</tbody></table></div>
        </section>}
        {model?.brandOnly && <p className="decoder-library-note">{t("decoderLibrary.brandOnly")}</p>}
        {maker?.coverage && <p className="decoder-library-note">{maker.coverage[language]}</p>}
        {family && <p className="decoder-family-notes">{family.notes[language]}</p>}
        {model?.notes && <p>{model.notes[language]}</p>}
        {profile ? <section className="decoder-profile-scope"><h5>{t("decoderLibrary.interpretation")}</h5>
          <p>{language === "en" ? profile.nameEn : profile.name} · {t(profile.active ? "decoderLibrary.active" : "decoderLibrary.inactive")} · {profile.revision}</p>
          {profile.firmwareMin && <p>{t("decoderLibrary.minimumFirmware", { minimum: profile.firmwareMin })}</p>}
          <p>{t("decoderLibrary.partialCoverage")}</p>
          {roles.some(role => role === "Admin" || role === "Editor") && <button type="button" className="secondary-button" disabled={saving}
            onClick={() => onToggle(profile)}>{t(profile.active ? "decoderLibrary.deactivate" : "decoderLibrary.activate")}</button>}
        </section> : <p className="decoder-library-note">{t("decoderLibrary.basicOnly")}</p>}
        <p className="decoder-library-note">{t("decoderLibrary.noValueChanges")}</p>
      </>}
      {tab === "definitions" && <>
        {model && <DecoderManufacturerReferences key={`references-${model.id}`} modelId={model.id} />}
        <h5>{t("decoderLibrary.interpretation")}</h5>
        <DecoderLibraryDefinitions key={`definitions-${model?.id || "standard"}`} profile={profile} profiles={profiles} />
      </>}
      {tab === "identification" && <>
        <h5>{t("decoderLibrary.identificationTitle")}</h5>
        {model?.identification && <section className="decoder-profile-scope">
          <h5>{t("decoderLibrary.automaticIdentification")}</h5>
          <p>CV8 = {maker?.cv8} · CV{model.identification.cv} = {model.identification.values.join(" / ")}</p>
          <p>{t(model.identification.scope === "model" ? "decoderLibrary.identityModelScope" : "decoderLibrary.identityFamilyScope",
            { name: model.identification.name })}</p>
        </section>}
        <p className="decoder-family-notes">{family?.identification[language] || t("decoderLibrary.standardIdentification")}</p>
        <p className="decoder-library-note">{t("decoderLibrary.identificationLimits")}</p>
        {profile && <><h5>{t("decoderLibrary.importNames")}</h5>
          <ul className="decoder-aliases">{profile.aliases.map(alias => <li key={alias}>{alias}</li>)}</ul>
          {profile.familyAliases.length > 0 && <p>{t("decoderLibrary.familyNames")} {profile.familyAliases.join(", ")}</p>}</>}
      </>}
      {tab === "sources" && <>
        <p className="decoder-library-note">{t("decoderLibrary.sourceMethod")}</p>
        {sources.map(source => <SourceReference key={source.url} source={source} />)}
        <p>{t("decoderLibrary.limits")}</p>
      </>}
    </div>
  </section>;
}

function SourceReference({ source }: { source: DecoderSource }) {
  return <article className="decoder-source-reference">
    <a href={source.url} target="_blank" rel="noreferrer">{source.title}<ExternalLink size={14} aria-hidden="true" /></a>
    <p>{source.version}</p><p>{source.scope}</p>
  </article>;
}
