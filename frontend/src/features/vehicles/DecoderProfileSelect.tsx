import { useMemo, useState } from "react";
import type { DecoderCV } from "./decoderAnalysis";
import { decoderAnalysis, decoderProfiles } from "./decoderAnalysis";
import { useI18n } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";
import { useDecoderLibrary } from "../../shared/DecoderLibraryProvider";

export function useDecoderProfile(values: DecoderCV[]) {
  const library = useDecoderLibrary();
  const [requested, setProfile] = useState<string | null>(null);
  const profiles = decoderProfiles(values);
  const profile = requested != null && profiles.includes(requested) ? requested : profiles[0] || "";
  const analysis = useMemo(() => decoderAnalysis(values, profile, library.profiles, library.inventory ?? null),
    [values, profile, library.profiles, library.inventory]);
  return { profiles, profile, setProfile, analysis, failed: library.failed, loading: !library.failed && library.profiles.length === 0 };
}

export function DecoderProfileSelect({ profiles, profile, setProfile }: {
  profiles: string[]; profile: string; setProfile: (value: string) => void;
}) {
  const { t } = useI18n();
  return <label className="decoder-profile-select">{t("vehicle.field.decoderProfile")}
    <AppSelect value={profile} onChange={(event) => setProfile(event.target.value)}>
      {!profiles.length && <option value="">{t("vehicles.cv.noProfile")}</option>}
      {profiles.map((value) => <option key={value} value={value}>{value || t("vehicles.cv.noProfile")}</option>)}
    </AppSelect>
  </label>;
}
