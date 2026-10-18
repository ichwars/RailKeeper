import { useI18n } from "../../shared/i18n";

export const cvPageSize = 50;

export function CVPageControls({ page, total, onChange }: { page: number; total: number; onChange: (page: number) => void }) {
  const { t } = useI18n();
  return <div className="cv-page-controls">
    <span role="status">{t("vehicles.cv.pageRows", { from: total ? page * cvPageSize + 1 : 0,
      to: Math.min((page + 1) * cvPageSize, total), total })}</span>
    <button type="button" className="secondary-button" disabled={page === 0} onClick={() => onChange(page - 1)}>
      {t("vehicles.cv.previous")}
    </button>
    <button type="button" className="secondary-button" disabled={(page + 1) * cvPageSize >= total} onClick={() => onChange(page + 1)}>
      {t("vehicles.cv.next")}
    </button>
  </div>;
}

export function cvIndexLabel(value: { indexCV31?: number | null; indexCV32?: number | null }, baseLabel: string) {
  return value.indexCV31 == null && value.indexCV32 == null ? baseLabel
    : `CV31=${value.indexCV31 ?? "?"} / CV32=${value.indexCV32 ?? "?"}`;
}
