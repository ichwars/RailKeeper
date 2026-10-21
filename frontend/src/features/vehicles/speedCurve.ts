import type { VehicleCVValue, VehicleCVValueInput } from "../../shared/api";
import { decoderDefinition, resolveDecoderLibrary } from "../../shared/decoderLibrary";
import { bundledDecoderInventory, bundledDecoderProfiles } from "../../shared/decoderLibraryModel";
import { zimoSpeedParameters } from "./zimoMNAnalysis";

export type SpeedCurveCVSource = Pick<VehicleCVValue | VehicleCVValueInput,
  "cvNumber" | "value" | "decoderProfile" | "protocol" | "indexCV31" | "indexCV32" | "description">;

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
  motorCharacteristic?: { cvNumber: number; value: number | null };
  zimoCurve?: ReturnType<typeof zimoSpeedParameters>;
  lenzCurve?: boolean;
  kuehnCurve?: boolean;
  uhlenbrockCurve?: boolean;
  rocoCurve?: boolean;
  pikoCurve?: boolean;
};

const speedTableCVs = Array.from({ length: 28 }, (_, index) => 67 + index);
const threePointCVs = [2, 6, 5];

function groupKey(value: SpeedCurveCVSource) {
  const decoderProfile = (value.decoderProfile || "").trim();
  const protocol = (value.protocol || "").trim();
  return `${decoderProfile}::${protocol.toLocaleLowerCase("de-DE")}`;
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

export function buildSpeedCurveProfiles(values: SpeedCurveCVSource[], fallbackLabel: string,
  catalog = bundledDecoderProfiles, inventory = bundledDecoderInventory) {
  const groups = new Map<string, ReturnType<typeof numericCV>[]>();
  values
    .map(numericCV)
    .filter((value) => Number.isInteger(value.cvNumber) && Number.isInteger(value.value))
    .filter((value) => value.indexCV31 == null && value.indexCV32 == null)
    .forEach((value) => {
      const key = groupKey(value);
      groups.set(key, [...(groups.get(key) || []), value]);
    });

  return Array.from(groups.entries()).map(([id, groupValues]) => {
    const byCV = new Map<number, number>();
    groupValues.forEach((value) => byCV.set(value.cvNumber, value.value));
    const first = groupValues[0];
    const match = resolveDecoderLibrary(groupValues, first.decoderProfile, catalog, inventory);
    const model = catalog.find(p => p.active && p.id === match.profileId);
    const characteristicCV = ["matched", "family"].includes(match.status) && match.firmwareCompatible !== false
      ? model?.rules?.motorCurveCV : undefined;
    const zimoCurve = ["matched", "family"].includes(match.status) && match.firmwareCompatible !== false && model?.rules?.zimoMN
      ? zimoSpeedParameters(cv => byCV.get(cv)) : undefined;
    const lenzCurve = ["matched", "family"].includes(match.status) && match.firmwareCompatible !== false && model?.rules?.lenzStandardV2;
    const kuehnCurve = ["matched", "family"].includes(match.status) && match.firmwareCompatible !== false && model?.rules?.kuehnN45;
    const pikoCurve = ["matched", "family"].includes(match.status) && match.firmwareCompatible !== false && model?.rules?.piko56500;
    const rocoCurve = ["matched", "family"].includes(match.status) && match.firmwareCompatible !== false && model?.rules?.roco10882;
    const uhlenbrockCurve = ["matched", "family"].includes(match.status) && match.firmwareCompatible !== false && model?.rules?.uhlenbrock76420;
    const smoothed = Boolean(zimoCurve || lenzCurve || kuehnCurve || uhlenbrockCurve);
    const validByte = (value: number | undefined) => value != null && value >= 0 && value <= 255;
    const validParameter = (cv: number) => {
      const value = byCV.get(cv);
      return value != null && value >= 1 && value <= (rocoCurve ? cv === 6 ? 252 : 255 : 63);
    };
    const trim = (cv: number) => {
      const value = byCV.get(cv);
      return value != null && value >= 0 && value <= 63 ? value : null;
    };
    const threePoint = [
      { cvNumber: 2, step: 1, label: "CV 2" },
      { cvNumber: 6, step: 14, label: "CV 6" },
      { cvNumber: 5, step: 28, label: "CV 5" }
    ]
      .filter((point) => byCV.has(point.cvNumber) && (!(kuehnCurve || pikoCurve) || validByte(byCV.get(point.cvNumber))) &&
        (!(uhlenbrockCurve || rocoCurve) || validParameter(point.cvNumber)))
      .map((point) => ({ ...point, value: byCV.get(point.cvNumber) || 0 }));
    const speedTable = speedTableCVs
      .filter((cvNumber) => byCV.has(cvNumber) && (!(kuehnCurve || uhlenbrockCurve || rocoCurve || pikoCurve) || validByte(byCV.get(cvNumber))))
      .map((cvNumber) => ({
        cvNumber,
        step: cvNumber - 66,
        value: byCV.get(cvNumber) || 0,
        label: `CV ${cvNumber}`
      }));
    const cv29Value = byCV.has(29) ? byCV.get(29) || 0 : null;
    const curveConfig = (smoothed || rocoCurve || pikoCurve) && cv29Value != null && (cv29Value < 0 || cv29Value > (uhlenbrockCurve || rocoCurve ? 255 : 63)) ? null : cv29Value;
    const speedTableActive = characteristicCV || curveConfig === null ? null : (curveConfig & 16) === 16;
    const primaryCurve = characteristicCV ? "none" : rocoCurve || pikoCurve
      ? speedTableActive === true && speedTable.length > 0 ? "speedTable"
        : speedTableActive === false && threePoint.length > 0 ? "threePoint" : "none"
      : smoothed
      ? speedTableActive === true && speedTable.length > 0 ? "speedTable" : "none"
      : choosePrimaryCurve(speedTableActive, speedTable.length, threePoint.length);
    const relevantCVs = rocoCurve ? rocoSpeedCVs : uhlenbrockCurve ? uhlenbrockSpeedCVs : speedRelevantCVs;
    const characteristic = characteristicCV == null ? undefined : byCV.get(characteristicCV);
    const characteristicDefinition = characteristicCV == null ? undefined
      : decoderDefinition(catalog, match, { cvNumber: characteristicCV, value: characteristic ?? 0 })?.definition;
    const validCharacteristic = characteristic != null && characteristicDefinition != null &&
      characteristic >= characteristicDefinition.minimum && characteristic <= characteristicDefinition.maximum &&
      (!characteristicDefinition.allowedValues || characteristicDefinition.allowedValues.includes(characteristic));
    return {
      id,
      label: groupLabel(first.decoderProfile, first.protocol, fallbackLabel),
      decoderProfile: first.decoderProfile,
      protocol: first.protocol,
      speedTableActive,
      cv29Value,
      forwardTrim: pikoCurve ? validByte(byCV.get(66)) ? byCV.get(66)! : null : rocoCurve ? null : uhlenbrockCurve ? trim(47) : !kuehnCurve && byCV.has(66) ? byCV.get(66) || 0 : null,
      reverseTrim: pikoCurve ? validByte(byCV.get(95)) ? byCV.get(95)! : null : rocoCurve ? null : uhlenbrockCurve ? trim(48) : byCV.has(95) ? byCV.get(95) || 0 : null,
      threePoint,
      speedTable,
      missingThreePointCVs: threePointCVs.filter((cvNumber) => (!byCV.has(cvNumber) || ((uhlenbrockCurve || rocoCurve) && !validParameter(cvNumber)) || (pikoCurve && !validByte(byCV.get(cvNumber))))),
      missingSpeedTableCVs: speedTableCVs.filter((cvNumber) => !byCV.has(cvNumber) || ((rocoCurve || pikoCurve) && !validByte(byCV.get(cvNumber)))),
      primaryCurve,
      sourceCount: groupValues.filter(value => relevantCVs.has(value.cvNumber) || value.cvNumber === characteristicCV).length,
      motorCharacteristic: characteristicCV ? { cvNumber: characteristicCV, value: validCharacteristic ? characteristic! : null } : undefined,
      zimoCurve, lenzCurve, kuehnCurve, uhlenbrockCurve, rocoCurve, pikoCurve
    } satisfies SpeedCurveProfile;
  }).filter(profile => profile.sourceCount > 0).sort((a, b) => a.label.localeCompare(b.label, "de-DE"));
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

const rocoSpeedCVs = new Set([2, 5, 6, 29, ...speedTableCVs]);
const uhlenbrockSpeedCVs = new Set([2, 5, 6, 29, 47, 48, ...speedTableCVs]);

export const speedRelevantCVs = new Set([
  2,
  5,
  6,
  29,
  66,
  95,
  ...speedTableCVs
]);
