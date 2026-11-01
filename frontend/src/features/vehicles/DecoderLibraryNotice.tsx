import { BookOpen } from "lucide-react";
import { useI18n } from "../../shared/i18n";
import { useDecoderLibrary } from "../../shared/DecoderLibraryProvider";
import { resolveDecoderLibrary } from "../../shared/decoderLibrary";
import type { DecoderLibraryMatch, DecoderReading } from "../../shared/decoderLibraryModel";

export function DecoderLibraryNotice({ values, profile, match: suppliedMatch, vehicleDecoderType }: {
  values: DecoderReading[]; profile: string; match?: DecoderLibraryMatch; vehicleDecoderType?: string;
}) {
  const { t, language } = useI18n();
  const library = useDecoderLibrary();
  const match = suppliedMatch || resolveDecoderLibrary(values, profile, library.profiles, library.inventory ?? null);
  const definition = library.profiles.find(p => p.id === match.profileId);
  const name = match.matchedBy === "cv" ? match.identification?.name
    : match.status === "family" ? definition?.family : language === "en" ? definition?.nameEn : definition?.name;
  return <aside className="decoder-library-notice" data-status={match.status} aria-label={t("decoderLibrary.title")}>
    <BookOpen size={17} aria-hidden="true" />
    <div><strong>{t(`decoderLibrary.status.${match.status}`, { profile: name || "" })}</strong>
      <small>{t("decoderLibrary.coverage", { known: match.knownValues, total: match.totalValues })}
        {match.revision && ` · ${t("decoderLibrary.revision")}: ${match.revision}`}</small>
      {match.identification && <small>{t(match.manufacturerId == null ? "vehicles.decoder.profileCVIdentity" : "decoderLibrary.identityEvidence", {
        manufacturer: match.manufacturerId ?? "?", cv: match.identification.cvNumber, value: match.identification.value
      })}</small>}
      {match.ecosProfile && <small>{t(`vehicles.decoder.ecosProfile.${match.ecosProfile.status}`, {
        profile: match.ecosProfile.name
      })}</small>}
      {vehicleDecoderType && <small>{t("vehicles.decoder.vehicleEvidence", {
        manufacturer: match.manufacturerId ?? "?"
      })}</small>}
      {values.some(value => value.decoderProfile === "ECoS") && <small>{t("vehicles.decoder.ecosCache")}</small>}
      {match.status === "family" && match.matchedBy === "cv" && <small>{t("decoderLibrary.identityFamilyLimits")}</small>}
      {definition?.kind === "model" && definition.firmwareMin && <small>{t(match.firmwareCompatible === false ? "decoderLibrary.firmwareOutside"
        : !match.firmware ? "decoderLibrary.firmwareUnknown" : "decoderLibrary.firmwareScope", {
        firmware: match.firmware || "", minimum: definition.firmwareMin || ""
      })}</small>}
      {library.failed && <small>{t("decoderLibrary.loadFailed")}</small>}
    </div>
    <a href="/settings?tab=data&group=general&type=decoder_profile">{t("decoderLibrary.open")}</a>
  </aside>;
}
