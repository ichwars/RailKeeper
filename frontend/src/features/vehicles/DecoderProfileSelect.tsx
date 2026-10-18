import { useState } from "react";
import type { DecoderCV } from "./decoderAnalysis";
import { decoderAnalysis, decoderProfiles } from "./decoderAnalysis";
import { useI18n } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";

export function useDecoderProfile(values: DecoderCV[]) {
  const [requested, setProfile] = useState<string | null>(null);
  const profiles = decoderProfiles(values);
  const profile = requested != null && profiles.includes(requested) ? requested : profiles[0] || "";
  return { profiles, profile, setProfile, analysis: decoderAnalysis(values, profile) };
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
