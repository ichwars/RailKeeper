import { expect, it } from "vitest";
import cases from "../../../backend/internal/decoderlibrary/definition_cases_test.json";
import { bundledDecoderProfiles, type DecoderReading } from "./decoderLibraryModel";
import { decoderDefinition, resolveDecoderLibrary } from "./decoderLibrary";

for (const test of cases) it(test.name, () => {
  const decoderProfile = "LokPilot 5 micro DCC";
  const rows: DecoderReading[] = [{ cvNumber: 8, value: 151, decoderProfile }];
  if (test.firmware) rows.push({ cvNumber: 7, value: 255, decoderProfile,
    description: `LokProgrammer Decoder: LokPilot 5 micro DCC (${test.firmware})` });
  const match = resolveDecoderLibrary(rows, decoderProfile);
  const result = decoderDefinition(bundledDecoderProfiles, match, test.row);
  if (test.meaning) expect(result?.definition.name.de).toContain(test.meaning);
  else expect(result).toBeUndefined();
});
