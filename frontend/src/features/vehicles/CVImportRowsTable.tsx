import { AppChoiceInput } from "../../shared/ui/AppChoiceInput";
import { RailComPlusPreview } from "./RailComPlusPreview";
import { useState } from "react";
import { useI18n } from "../../shared/i18n";
import { ImportReviewContext, ImportReviewStatus } from "../../shared/ImportReview";
import type { CVImportPreview } from "./cvImport";
import { CVPageControls, cvIndexLabel, cvPageSize } from "./CVPageControls";
import { DecoderLibraryNotice } from "./DecoderLibraryNotice";
import { useDecoderLibrary } from "../../shared/DecoderLibraryProvider";
import { resolveDecoderLibrary, decoderDefinition } from "../../shared/decoderLibrary";
import { decoderProfileLabel } from "./decoderProfileLabel";

export function CVImportRowsTable({ preview, saving, readonly, toggle }: {
  preview: CVImportPreview; saving: boolean; readonly: boolean; toggle: (id: string, selected: boolean) => void;
}) {
  const { t, language } = useI18n();
  const library = useDecoderLibrary();
  const [requestedPage, setPage] = useState(0);
  const page = Math.min(requestedPage, Math.max(0, Math.ceil(preview.rows.length / cvPageSize) - 1));
  const pageRows = preview.rows.slice(page * cvPageSize, (page + 1) * cvPageSize);
  const profiles = [...new Set(pageRows.map(row => row.input.decoderProfile || ""))];
  const groups = new Map(profiles.map(profile => [profile, preview.rows
    .filter(row => (row.input.decoderProfile || "") === profile).map(row => row.input)]));
  const matches = new Map(profiles.map(profile => [profile,
    resolveDecoderLibrary(groups.get(profile) || [], profile, library.profiles, library.inventory ?? null)]));
  const meaning = (input: CVImportPreview["rows"][number]["input"]) => {
    const match = matches.get(input.decoderProfile || "");
    return match ? decoderDefinition(library.profiles, match, input)?.definition.name[language] : undefined;
  };
  return <>
    <ImportReviewContext source={preview.fileName} target={preview.workingTarget
      ? `${t("importReview.target.cv")} · ${decoderProfileLabel(preview.workingTarget.profile, t("vehicles.cv.noProfile"))}` : t("importReview.target.cv")}
      effect={t("importReview.effect.cv")} selected={preview.rows.filter(row => row.selected).length} total={preview.rows.length} />
    {preview.previousWorkingGroupLabels?.length ? <p className="cv-working-assignments">{t("cvScope.previousAssignment")}: {preview.previousWorkingGroupLabels.join("; ")}</p> : null}
    {preview.workingGroupLabels?.length ? <p className="cv-working-assignments">{t("cvScope.confirmedAssignment")}: {preview.workingGroupLabels.join("; ")}</p> : null}
    {profiles.slice(0, 3).map(profile => <DecoderLibraryNotice key={profile} profile={profile} values={groups.get(profile) || []} />)}
    <RailComPlusPreview values={preview.rows.map(row => row.input)} />
    <CVPageControls page={page} total={preview.rows.length} onChange={setPage} />
    <div className="table-wrap compact-table cv-import-table">
      <table>
        <thead><tr>
          <th>{t("vehicles.articleSearch.apply")}</th><th>CV</th><th>{t("vehicles.cv.index")}</th>
          <th>{t("vehicles.articleSearch.current")}</th><th>Import</th>
          <th>{t("vehicles.cv.profiles")}</th><th>{t("vehicles.articleSearch.status")}</th>
        </tr></thead>
        <tbody>{pageRows.map((row) => <tr key={row.id} className={`cv-import-${row.status}`}>
          <td><AppChoiceInput type="checkbox" checked={row.selected}
            onChange={(event) => toggle(row.id, event.target.checked)} disabled={row.status === "invalid" || saving || readonly}
            aria-label={t("vehicles.cv.applyCv", { cv: `${row.input.cvNumber}${row.input.indexCV31 == null ? "" : ` (${cvIndexLabel(row.input, "")})`}` })} /></td>
          <td>{row.input.cvNumber || "-"}<small>{meaning(row.input)}</small></td><td>{cvIndexLabel(row.input, t("vehicles.cv.base"))}</td>
          <td>{row.existing?.value ?? "-"}</td><td>{Number.isFinite(row.input.value) ? row.input.value : "-"}</td>
          <td>{decoderProfileLabel(row.input.decoderProfile || "", "-")}
            {row.sourceProfile !== undefined && row.sourceProfile !== (row.input.decoderProfile || "") &&
              <small>{t("cvScope.sourceProfile", { profile: decoderProfileLabel(row.sourceProfile, t("vehicles.cv.noProfile")) })}</small>}</td><td>
            <ImportReviewStatus state={row.status === "invalid" ? "excluded" : row.status} />
            {row.status === "invalid" && <small>{row.reason ? t(`importReview.cv.${row.reason}`) : row.message}</small>}
            {row.status === "changed" && <small>{row.changedFields ? t("importReview.cv.changes", {
              fields: row.changedFields.map(field => t(`importReview.cv.field.${field}`)).join(", ")
            }) : row.message}</small>}
          </td>
        </tr>)}</tbody>
      </table>
    </div>
  </>;
}
