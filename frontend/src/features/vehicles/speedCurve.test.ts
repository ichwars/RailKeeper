import { describe, expect, it } from "vitest";
import { buildSpeedCurveProfiles } from "./speedCurve";

describe("speed curve decoder identity", () => {
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
