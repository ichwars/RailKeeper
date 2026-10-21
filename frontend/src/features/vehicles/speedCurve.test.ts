import { describe, expect, it } from "vitest";
import { buildSpeedCurveProfiles } from "./speedCurve";

describe("speed curve decoder identity", () => {
  it("uses DH48 as a characteristic parameter instead of plotting unrelated DCC points", () => {
    const values = [[8, 97], [261, 180], [29, 30], [48, 5], [2, 0], [5, 92], [6, 100], [67, 123]]
      .map(([cvNumber, value]) => ({ cvNumber, value, decoderProfile: "DH18A" }));
    const before = structuredClone(values);
    const [profile] = buildSpeedCurveProfiles(values, "Ohne Profil");
    expect(profile.motorCharacteristic).toEqual({ cvNumber: 48, value: 5 });
    expect(profile.primaryCurve).toBe("none");
    expect(profile.speedTableActive).toBeNull();
    expect(values).toEqual(before);
  });

  it("does not replace base curve data with an indexed copy or plot missing DH48", () => {
    const [profile] = buildSpeedCurveProfiles([
      { cvNumber: 8, value: 97, decoderProfile: "DH18A" },
      { cvNumber: 261, value: 180, decoderProfile: "DH18A" },
      { cvNumber: 29, value: 14, decoderProfile: "DH18A" },
      { cvNumber: 48, value: 7, decoderProfile: "DH18A", indexCV31: 1, indexCV32: 1 }
    ], "Ohne Profil");
    expect(profile.motorCharacteristic).toEqual({ cvNumber: 48, value: null });
    expect(profile.primaryCurve).toBe("none");
    const [generic] = buildSpeedCurveProfiles([
      { cvNumber: 2, value: 3 }, { cvNumber: 29, value: 14 },
      { cvNumber: 29, value: 30, indexCV31: 0, indexCV32: 0 }
    ], "Ohne Profil");
    expect(generic.cv29Value).toBe(14);
    expect(generic.primaryCurve).toBe("threePoint");
  });
  it("keeps profiles distinct using the same exact identity as the CV table", () => {
    const profiles = buildSpeedCurveProfiles([
      { cvNumber: 2, value: 3, decoderProfile: "Front" },
      { cvNumber: 2, value: 7, decoderProfile: "front" }
    ], "Ohne Profil");
    expect(profiles).toHaveLength(2);
    expect(profiles.find((profile) => profile.decoderProfile === "Front")?.threePoint[0].value).toBe(3);
    expect(profiles.find((profile) => profile.decoderProfile === "front")?.threePoint[0].value).toBe(7);
  });
});
