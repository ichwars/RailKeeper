import type { VehicleCVValue, VehicleCVValueInput } from "../../shared/api";

export type SpeedCurveCVSource = Pick<VehicleCVValue | VehicleCVValueInput, "cvNumber" | "value" | "decoderProfile" | "protocol">;

export type SpeedCurvePoint = {
  step: number;
  cvNumber: number;
  value: number;
  label: string;
};

export type SpeedCurveProfile = {
  id: string;
  label: string;
  decoderProfile: string;
  protocol: string;
  speedTableActive: boolean | null;
  cv29Value: number | null;
  forwardTrim: number | null;
  reverseTrim: number | null;
  threePoint: SpeedCurvePoint[];
  speedTable: SpeedCurvePoint[];
  missingThreePointCVs: number[];
  missingSpeedTableCVs: number[];
  primaryCurve: "speedTable" | "threePoint" | "none";
  sourceCount: number;
};

const speedTableCVs = Array.from({ length: 28 }, (_, index) => 67 + index);
const threePointCVs = [2, 6, 5];

function groupKey(value: SpeedCurveCVSource) {
  const decoderProfile = (value.decoderProfile || "").trim();
  const protocol = (value.protocol || "").trim();
  return `${decoderProfile.toLocaleLowerCase("de-DE")}::${protocol.toLocaleLowerCase("de-DE")}`;
}

function groupLabel(decoderProfile: string, protocol: string, fallbackLabel: string) {
  if (decoderProfile && protocol) return `${decoderProfile} - ${protocol}`;
  if (decoderProfile) return decoderProfile;
  if (protocol) return protocol;
  return fallbackLabel;
}

function numericCV(value: SpeedCurveCVSource) {
  return {
    ...value,
    cvNumber: Number(value.cvNumber),
    value: Number(value.value),
    decoderProfile: (value.decoderProfile || "").trim(),
    protocol: (value.protocol || "").trim()
  };
}

export function buildSpeedCurveProfiles(values: SpeedCurveCVSource[], fallbackLabel: string) {
  const groups = new Map<string, ReturnType<typeof numericCV>[]>();
  values
    .map(numericCV)
    .filter((value) => Number.isInteger(value.cvNumber) && Number.isInteger(value.value))
    .filter((value) => speedRelevantCVs.has(value.cvNumber))
    .forEach((value) => {
      const key = groupKey(value);
      groups.set(key, [...(groups.get(key) || []), value]);
    });

  return Array.from(groups.entries()).map(([id, groupValues]) => {
    const byCV = new Map<number, number>();
    groupValues.forEach((value) => byCV.set(value.cvNumber, value.value));
    const first = groupValues[0];
    const threePoint = [
      { cvNumber: 2, step: 1, label: "CV 2" },
      { cvNumber: 6, step: 14, label: "CV 6" },
      { cvNumber: 5, step: 28, label: "CV 5" }
    ]
      .filter((point) => byCV.has(point.cvNumber))
      .map((point) => ({ ...point, value: byCV.get(point.cvNumber) || 0 }));
    const speedTable = speedTableCVs
      .filter((cvNumber) => byCV.has(cvNumber))
      .map((cvNumber) => ({
        cvNumber,
        step: cvNumber - 66,
        value: byCV.get(cvNumber) || 0,
        label: `CV ${cvNumber}`
      }));
    const cv29Value = byCV.has(29) ? byCV.get(29) || 0 : null;
    const speedTableActive = cv29Value === null ? null : (cv29Value & 16) === 16;
    const primaryCurve = choosePrimaryCurve(speedTableActive, speedTable.length, threePoint.length);
    return {
      id,
      label: groupLabel(first.decoderProfile, first.protocol, fallbackLabel),
      decoderProfile: first.decoderProfile,
      protocol: first.protocol,
      speedTableActive,
      cv29Value,
      forwardTrim: byCV.has(66) ? byCV.get(66) || 0 : null,
      reverseTrim: byCV.has(95) ? byCV.get(95) || 0 : null,
      threePoint,
      speedTable,
      missingThreePointCVs: threePointCVs.filter((cvNumber) => !byCV.has(cvNumber)),
      missingSpeedTableCVs: speedTableCVs.filter((cvNumber) => !byCV.has(cvNumber)),
      primaryCurve,
      sourceCount: groupValues.length
    } satisfies SpeedCurveProfile;
  }).sort((a, b) => a.label.localeCompare(b.label, "de-DE"));
}

function choosePrimaryCurve(speedTableActive: boolean | null, speedTableCount: number, threePointCount: number) {
  if (speedTableActive === true && speedTableCount > 0) return "speedTable";
  if (speedTableActive === false && threePointCount > 0) return "threePoint";
  if (speedTableCount === 28) return "speedTable";
  if (threePointCount >= 2) return "threePoint";
  if (speedTableCount > 0) return "speedTable";
  if (threePointCount > 0) return "threePoint";
  return "none";
}

export const speedRelevantCVs = new Set([
  2,
  5,
  6,
  29,
  66,
  95,
  ...speedTableCVs
]);
