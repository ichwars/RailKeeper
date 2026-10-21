import { ArrowRight, Check, ChevronRight, FileText } from "lucide-react";
import { useId, type ReactNode } from "react";
import type { CreateVehicleRequest, VehicleCVValue } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { cvSource } from "./decoderAnalysis";
import { DecoderProfileSelect, useDecoderProfile } from "./DecoderProfileSelect";
import { DecoderLibraryNotice } from "./DecoderLibraryNotice";
import { useDecoderLibrary } from "../../shared/DecoderLibraryProvider";

type DigitalFieldProps = {
  label: string; value: ReactNode; source: string; known?: boolean; title?: string;
};

function DigitalField({ label, value, source, known = false, title }: DigitalFieldProps) {
  return <div className={known ? "digital-field derived" : "digital-field"}>
    <span className="digital-field-caption">{label}</span>
    <output className="digital-field-value" aria-label={label} title={typeof value === "string" ? value : undefined}>
      <strong>{value}</strong>{known && <Check size={14} aria-hidden="true" />}
    </output>
    <small className="digital-source" title={title}><FileText size={12} aria-hidden="true" />{source}</small>
  </div>;
}

export function VehicleDigitalFields({ form, values, readonly, update, onOpenCV }: {
  form: CreateVehicleRequest; values: VehicleCVValue[]; readonly: boolean;
  update: (patch: Partial<CreateVehicleRequest>) => void; onOpenCV?: () => void;
}) {
  const { t } = useI18n();
  const dtNumberId = useId();
  const state = useDecoderProfile(values), a = state.analysis;
  const { inventory } = useDecoderLibrary();
  const manufacturerName = inventory?.manufacturers.find(maker => maker.cv8 === a.get(8))?.name;
  const unknown = t("vehicles.decoder.unknown");
  const typeName = a.libraryMatch.matchedBy === "cv" && ["matched", "identified", "family"].includes(a.libraryMatch.status)
    ? a.libraryMatch.identification?.name : a.modelRules ? state.profile : undefined;
  const yesNo = (value: boolean | undefined) => value == null ? unknown : t(value ? "common.yes" : "common.no");
  const apply = () => update({
    ...(a.address != null ? { digital: true, digitalDecoderNumber: String(a.address) } : {}),
    ...(typeName ? { decoderType: typeName } : {}),
    ...(a.abc != null ? { abcBrakes: a.abc } : {})
  });
  const canApply = a.address != null || Boolean(typeName) || a.abc != null;
  const longActive = a.config != null && Boolean(a.config & 32);
  const address = a.address == null ? unknown : `${a.address} · ${t(longActive
    ? "vehicles.decoder.longDccAddress" : "vehicles.decoder.shortDccAddress")}`;
  const headerSource = a.libraryMatch.identification
    ? `CV8=${a.libraryMatch.manufacturerId} · CV${a.libraryMatch.identification.cvNumber}=${a.libraryMatch.identification.value}`
    : a.firmware ? t("vehicles.decoder.txtHeader") : t("vehicle.field.decoderProfile");
  const firmwareSource = t(a.firmware ? "vehicles.decoder.firmwareHeader" : "vehicles.decoder.firmwareSource");
  const abcBits = Array.from({ length: 8 }, (_, bit) => bit).filter(bit => (a.rules?.abc?.mask || 0) & (1 << bit));

  return <div className="vehicle-digital-fields">
    <DecoderLibraryNotice values={values} profile={state.profile} />
    <div className="digital-state-row">
      <label className="digital-equipped"><input type="checkbox" checked={Boolean(form.digital)}
        aria-label="Digital" onChange={(event) => update({ digital: event.target.checked })}
        disabled={readonly} />{t("vehicles.decoder.equipped")}</label>
      <button type="button" className="secondary-button" disabled={readonly || !canApply} onClick={apply}>
        {t("vehicles.decoder.applyFields")}</button>
    </div>
    {state.profiles.length > 1 && <div className="digital-profile-choice"><DecoderProfileSelect {...state} /></div>}
    {!values.length && <p className="digital-empty">{t("vehicles.decoder.noValues")}</p>}
    <div className="digital-core">
      <DigitalField label={t("vehicle.field.digitalDecoderNumber")} value={address} known={a.address != null}
        source={a.address == null ? cvSource(a, 1, 29) : t(longActive
          ? "vehicles.decoder.longAddressSource" : "vehicles.decoder.shortAddressSource")}
        title={cvSource(a, 1, 17, 18, 29)} />
      <div className="digital-field dt-primary">
        <label className="digital-field-caption dt-enable-control"><input type="checkbox" aria-label={t("vehicles.decoder.dtEnabled")} checked={Boolean(form.dtDecoder)}
          onChange={(event) => update({ dtDecoder: event.target.checked })} disabled={readonly} />{t("vehicle.field.dtDecoderNumber")}</label>
        <input id={dtNumberId} aria-label={t("vehicle.field.dtDecoderNumber")} value={form.dtDecoderNumber || ""} placeholder={t("vehicles.decoder.doubleHeading")}
          onChange={(event) => update({ dtDecoderNumber: event.target.value })} disabled={readonly || !form.dtDecoder} />
        <small className="digital-source">{t("vehicles.decoder.dtManual")}</small>
      </div>
      <DigitalField label={t("vehicles.decoder.consistAddress")} known={a.consistAddress != null}
        value={a.consistAddress == null ? unknown : a.consistAddress === 0 ? t("vehicles.decoder.inactive")
          : a.consistReverse == null ? a.consistAddress : `${a.consistAddress} · ${t(a.consistReverse ? "vehicles.decoder.reverse" : "vehicles.decoder.forward")}`}
        source={cvSource(a, ...a.consistCVs)} />
    </div>
    <section className="consist-permissions">
      <h4>{t("vehicles.decoder.consistFunctions")}</h4>
      <div className="consist-chips">{a.enabledFunctions.map((key) => <span key={key}>{key}</span>)}
        {!a.enabledFunctions.length && <span className="consist-empty">{a.functionsComplete ? t("vehicles.decoder.noneEnabled") : unknown}</span>}
      </div>
      <small className="digital-source"><FileText size={12} aria-hidden="true" />{cvSource(a, 21, 22)} · {t("vehicles.decoder.consistPermissionHelp")}
        {!a.functionsComplete && ` · ${t(a.roco ? "vehicles.decoder.roco.consistPermissions" : a.uhlenbrock ? "vehicles.decoder.uhlenbrock.consistPermissions" : a.kuehn ? "vehicles.decoder.kuehn.consistPermissions" : a.lenz ? "vehicles.decoder.lenz.consistPermissions"
          : a.zimo?.ambiguousFunctionBit ? "vehicles.decoder.zimo.ambiguousFunction" : "vehicles.decoder.incomplete")}`}
        {a.zimo?.autoConsist && ` · ${t("vehicles.decoder.zimo.autoConsist")}`}</small>
    </section>
    <div className="brake-summary">
      <div className="digital-brake"><div><h4>{t("vehicles.decoder.abc")}</h4>
        <span className={a.abc ? "digital-state-badge enabled" : "digital-state-badge"}>{yesNo(a.abc)}</span></div>
        <small className="digital-source"><FileText size={12} aria-hidden="true" />{a.rules?.abc
          ? `${cvSource(a, a.rules.abc.cv)} · ${t(abcBits.length === 1 ? "decoderLibrary.bit" : "decoderLibrary.bits", { bits: abcBits.join("/") })}` : t("decoderLibrary.noRule")}</small></div>
      <div className="digital-brake"><div><h4>{t("vehicles.decoder.constantBrake")}</h4>
        <span className={a.constantBrake ? "digital-state-badge enabled" : "digital-state-badge"}>{yesNo(a.constantBrake)}</span></div>
        <small className="digital-source"><FileText size={12} aria-hidden="true" />{a.constantBrakeCVs.length
          ? cvSource(a, ...a.constantBrakeCVs) : t("decoderLibrary.noRule")}
          {a.zimo?.brakeMode != null && ` · ${t(`vehicles.decoder.zimo.mode.${a.zimo.brakeMode}`)}`}</small></div>
    </div>
    <div className="digital-metadata digital-core">
      <DigitalField label={t("vehicle.field.decoderType")} value={typeName || unknown} source={headerSource} known={Boolean(typeName)} />
      <DigitalField label={t("vehicles.decoder.manufacturer")} value={manufacturerName || (a.get(8) ?? unknown)}
        source={cvSource(a, 8)} known={a.get(8) != null} />
      <DigitalField label={t("vehicles.decoder.firmware")} value={a.firmware || unknown}
        source={firmwareSource} known={Boolean(a.firmware)} />
    </div>
    {a.railComPlus?.name && <div className="railcom-name"><DigitalField label={t("railcom.name")} value={a.railComPlus.name}
      source={`CV31=1 / CV32=0 · CV261–${a.railComPlus.nameLastCV}`} known />
      <p className="decoder-hint">{t("railcom.nameHelp")}</p></div>}
    {a.railComPlus?.nameIssue && <p className="decoder-notice">{t(`railcom.name.${a.railComPlus.nameIssue}`)}</p>}
    <details className="digital-more"><summary><ChevronRight size={14} aria-hidden="true" />
      {t("vehicles.decoder.more")}<span>{t("vehicles.decoder.moreHint")}</span></summary>
      <div className="digital-more-content">
        <div className="decoder-profile-bar"><DecoderProfileSelect {...state} />
          <span>{t("vehicles.decoder.recordCount", { count: a.rows.length, pages: a.indexPages })}</span></div>
        <DigitalField label={t("vehicles.decoder.longAddress")} value={a.longAddress ?? unknown}
          source={cvSource(a, 17, 18, 29)} known={a.longAddress != null} />
        <section className="digital-vehicle-fields"><h4>{t("vehicles.decoder.vehicleFields")}</h4>
          <div className="form-row three-columns">
            <label>{t("vehicle.field.digitalDecoderNumber")}<input value={form.digitalDecoderNumber || ""}
              onChange={(event) => update({ digitalDecoderNumber: event.target.value })} disabled={readonly || !form.digital} /></label>
            <label>{t("vehicle.field.decoderType")}<input value={form.decoderType || ""} placeholder={typeName || ""}
              onChange={(event) => update({ decoderType: event.target.value })} disabled={readonly} /></label>
            <label className="digital-manual-abc"><input type="checkbox" checked={Boolean(form.abcBrakes)}
              onChange={(event) => update({ abcBrakes: event.target.checked })} disabled={readonly} />{t("vehicles.decoder.manualAbc")}</label>
          </div><p className="decoder-hint">{t("vehicles.decoder.applyFieldsHelp")}</p>
        </section>
        {!a.modelRules && <p className="decoder-hint">{t("vehicles.decoder.unsupported")}</p>}
      </div>
    </details>
    <p className="digital-help"><FileText size={16} aria-hidden="true" /><span>{t("vehicles.decoder.snapshotCompact")}{" "}
      {onOpenCV && <button type="button" onClick={onOpenCV}>{t("vehicles.decoder.openCV")}<ArrowRight size={15} aria-hidden="true" /></button>}</span></p>
  </div>;
}
