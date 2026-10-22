import { Cpu } from "lucide-react";
import { useI18n } from "../../shared/i18n";
import { DecoderProfileSelect, useDecoderProfile } from "./DecoderProfileSelect";
import { DecoderLibraryNotice } from "./DecoderLibraryNotice";

export function DecoderProfileStrip({ state }: { state: ReturnType<typeof useDecoderProfile> }) {
  const { t } = useI18n();
  const a = state.analysis;
  return <><div className="decoder-profile-strip">
    <Cpu size={20} aria-hidden="true" />
    <div className="decoder-profile-identity">
      {state.profiles.length > 1 ? <DecoderProfileSelect {...state} />
        : <strong>{a.get(8) === 151 && "ESU · "}{state.profile || t("vehicles.cv.noProfile")}</strong>}
      <small>{t("vehicles.decoder.recordCount", { count: a.rows.length, pages: a.indexPages })}</small>
    </div>
    <span>{a.address == null ? t("vehicles.decoder.unknown")
      : t("vehicles.decoder.profileAddress", { address: a.address })}</span>
  </div><DecoderLibraryNotice values={a.rows} profile={state.profile} match={a.libraryMatch}
    vehicleDecoderType={a.vehicleDecoderType} /></>;
}
