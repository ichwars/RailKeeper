import { useState } from "react";
import { useI18n } from "../../shared/i18n";
import type { CVImportPreview } from "./cvImport";
import { CVPageControls, cvIndexLabel, cvPageSize } from "./CVPageControls";

export function CVImportRowsTable({ preview, saving, readonly, toggle }: {
  preview: CVImportPreview; saving: boolean; readonly: boolean; toggle: (id: string, selected: boolean) => void;
}) {
  const { t } = useI18n();
  const [requestedPage, setPage] = useState(0);
  const page = Math.min(requestedPage, Math.max(0, Math.ceil(preview.rows.length / cvPageSize) - 1));
  return <>
    <CVPageControls page={page} total={preview.rows.length} onChange={setPage} />
    <div className="table-wrap compact-table cv-import-table">
      <table>
        <thead><tr>
          <th>{t("vehicles.articleSearch.apply")}</th><th>CV</th><th>{t("vehicles.cv.index")}</th>
          <th>{t("vehicles.articleSearch.current")}</th><th>Import</th>
          <th>{t("vehicles.cv.profiles")}</th><th>{t("vehicles.articleSearch.status")}</th>
        </tr></thead>
        <tbody>{preview.rows.slice(page * cvPageSize, (page + 1) * cvPageSize).map((row) => <tr key={row.id} className={`cv-import-${row.status}`}>
          <td><input type="checkbox" checked={row.selected}
            onChange={(event) => toggle(row.id, event.target.checked)} disabled={row.status === "invalid" || saving || readonly}
            aria-label={t("vehicles.cv.applyCv", { cv: `${row.input.cvNumber}${row.input.indexCV31 == null ? "" : ` (${cvIndexLabel(row.input, "")})`}` })} /></td>
          <td>{row.input.cvNumber || "-"}</td><td>{cvIndexLabel(row.input, t("vehicles.cv.base"))}</td>
          <td>{row.existing?.value ?? "-"}</td><td>{Number.isFinite(row.input.value) ? row.input.value : "-"}</td>
          <td>{row.input.decoderProfile || "-"}</td><td>{row.message}</td>
        </tr>)}</tbody>
      </table>
    </div>
  </>;
}
