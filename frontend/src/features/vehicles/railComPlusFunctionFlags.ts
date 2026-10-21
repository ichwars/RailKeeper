import type { VehicleFunctionInput } from "../../shared/api";
import type { RailComPlusData, RailComPlusFunction } from "../../shared/railComPlus";
import type { Translator } from "./decoderAnalysis";
import type { FunctionEdit } from "./vehicleTransforms";

export function railComPlusFlagAvailability(current: FunctionEdit) {
  return {
    // A new row's default is not a manual choice. Every saved or explicitly chosen mode is protected.
    mode: !current.persisted && !current.modeAssigned && (!current.mode || current.mode === "dauer"),
    inverted: current.inverted == null
  };
}

export function railComPlusFlagPatch(data: RailComPlusData, row: RailComPlusFunction,
  current: FunctionEdit, t: Translator): Partial<VehicleFunctionInput> | undefined {
  const verification = data.flagsVerification;
  if (!verification) return undefined;
  const available = railComPlusFlagAvailability(current);
  if (!available.mode && !available.inverted) return undefined;
  const fields = [
    ...(available.mode ? [`${t("vehicles.functions.mode")}: ${t(row.momentary
      ? "vehicles.functionMode.moment" : "vehicles.functionMode.dauer")}`] : []),
    ...(available.inverted ? [`${t("railcom.inverted")}: ${t(row.inverted ? "common.yes" : "common.no")}`] : [])
  ];
  const lines = [row.sourceLine, row.secondSourceLine].filter((line): line is number => line != null);
  const source = [
    "RailComPlus", fields.join("; "), "CV31=1 / CV32=1",
    `CV${row.firstCV}=${row.firstValue}; CV${row.firstCV + 1}=${row.secondValue}`,
    ...(lines.length ? [t("railcom.line", { line: lines.join(" / ") })] : []),
    ...(row.sourceExportedAt ? [`${t("vehicles.decoder.exportedAt")}: ${row.sourceExportedAt}`] : []),
    t("railcom.flags.reference", verification)
  ].join(" · ");
  const notes = [current.notes, source].filter(Boolean).join("\n");
  // The API limit is UTF-8 bytes. Never truncate manual notes or the source evidence.
  if (new TextEncoder().encode(notes).length > 1000) return undefined;
  return {
    ...(available.mode ? { mode: row.momentary ? "moment" : "dauer" } : {}),
    ...(available.inverted ? { inverted: row.inverted } : {}),
    flagsVersion: 2,
    notes
  };
}
