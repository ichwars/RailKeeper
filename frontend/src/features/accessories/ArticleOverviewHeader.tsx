import { Plus } from "lucide-react";

import { useI18n } from "../../shared/i18n";

export function ArticleOverviewHeader({ canEdit, onCreate }: { canEdit: boolean; onCreate: () => void }) {
  const { t } = useI18n();

  return (
    <section className="inventory-head accessory-head">
      <div>
        <p className="eyebrow">{t("accessories.overview.eyebrow")}</p>
        <h1>{t("accessories.overview.title")}</h1>
        <p>{t("accessories.overview.subtitle")}</p>
      </div>
      {canEdit ? (
        <button type="button" className="primary-button article-create-button" onClick={onCreate}>
          <Plus size={16} aria-hidden="true" />
          {t("accessories.overview.create")}
        </button>
      ) : null}
    </section>
  );
}
