import { useDecoderLibrary } from "../../shared/DecoderLibraryProvider";
import { useI18n } from "../../shared/i18n";
import { decoderAnalysis, type DecoderCV } from "./decoderAnalysis";
import { RailComPlusSuggestions } from "./RailComPlusSuggestions";

export function RailComPlusPreview({ values, defaultProfile = "" }: { values: DecoderCV[]; defaultProfile?: string }) {
  const library = useDecoderLibrary(), { t } = useI18n();
  const rows = values.map(value => ({ ...value, decoderProfile: value.decoderProfile || defaultProfile }));
  const groups = [...new Set(rows.map(value => value.decoderProfile))];
  const results = groups.map(profile => ({ profile,
    data: decoderAnalysis(rows, profile, library.profiles, library.inventory ?? null).railComPlus
  })).filter(({ data }) => data && (data.name || data.nameIssue || data.functions.length ||
    data.incompleteFunctions || data.unassignedFunctions));
  if (!results.length) return null;
  return <details className="railcom-preview"><summary>{t("railcom.preview")}</summary>
    <p className="decoder-hint">{t("railcom.previewHelp")}</p>
    {results.map(result => <div key={result.profile}><p><strong>{result.profile}</strong></p>
      <RailComPlusSuggestions data={result.data} /></div>)}
  </details>;
}
