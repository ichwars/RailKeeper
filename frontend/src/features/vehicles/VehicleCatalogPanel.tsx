import { BookOpen } from "lucide-react";
import { formatDateTime, useI18n } from "../../shared/i18n";
import { ImportReviewContext, ImportReviewStatus } from "../../shared/ImportReview";
import { AppSelect } from "../../shared/ui/AppSelect";
import { catalogFieldNames, catalogRequiredFields, missingCatalogFields } from "./vehicleCatalogModel";
import { useVehicleCatalogController, type VehicleCatalogOptions } from "./useVehicleCatalogController";

export function VehicleCatalogPanel({ targetName, ...options }: VehicleCatalogOptions & { targetName: string }) {
  const { t, language } = useI18n();
  const model = useVehicleCatalogController(options);
  const { context, draft, loading, saving, error, status, stale, alreadyStored, available } = model;
  const disabled = Boolean(options.busy || loading || saving);
  const source = context?.revisions.find(revision => revision.id === draft?.revisionID);
  const unchanged = context?.authority === "catalog" && !draft?.creating &&
    draft?.revisionID === context.pinned?.id && !draft?.selected.length;
  const missing = draft?.creating ? missingCatalogFields(draft.fields) : [];
  return <details className="vehicle-catalog-panel" aria-busy={loading || saving}
    onKeyDown={event => {
      if (event.key === "Enter" && event.target instanceof HTMLInputElement) {
        event.preventDefault(); event.stopPropagation();
      }
    }}
    onToggle={event => { if (event.currentTarget.open && !context && !loading) void model.load(false); }}>
    <summary><BookOpen size={16} aria-hidden="true" />{t("catalog.title")}</summary>
    <div className="catalog-panel-content vehicle-form">
      <p className="catalog-explanation">{t("catalog.scope")}</p>
      {loading && <p role="status">{t("catalog.loading")}</p>}
      {error && <p className="error-text" role="alert">{error}</p>}
      {status && <p className="catalog-saved" role="status">{status}</p>}
      {context && draft && <>
        <p>{context.authority === "legacy" ? t("catalog.pending") : context.authority === "unlinked"
          ? t("catalog.unlinked") : t("catalog.assigned", { number: context.pinned?.revisionNumber ?? 0 })}</p>
        <div className="catalog-toolbar">
          {context.revisions.length > 0 && <label>{t("catalog.revisionChoice")}
            <AppSelect value={draft.revisionID} disabled={disabled || draft.creating}
              onChange={event => model.chooseRevision(event.target.value)}>
              {!available && <option value={draft.revisionID} disabled>{t("catalog.previousSelection")}</option>}
              {context.revisions.map(revision => <option key={revision.id} value={revision.id}>
                {t("catalog.revisionLabel", { number: revision.revisionNumber })}
              </option>)}
            </AppSelect>
          </label>}
          {!options.readonly && <button type="button" className="secondary-button" disabled={disabled}
            onClick={draft.creating ? model.discardRevision : model.startRevision}>
            {t(draft.creating ? "catalog.discard" : "catalog.new")}
          </button>}
        </div>
        {source && !draft.creating && <p className="catalog-revision-note">
          {t("catalog.recorded", { date: formatDateTime(source.createdAt, language) })}
          {source.note && <span>{source.note}</span>}
        </p>}
        {Object.keys(context.overrides).length > 0 && <details className="catalog-overrides">
          <summary>{t(context.authority === "legacy" ? "catalog.pendingValues" : "catalog.individualValues")}</summary>
          <dl>{catalogFieldNames.filter(field => Object.hasOwn(context.overrides, field)).map(field =>
            <div key={field}><dt>{t(`vehicle.field.${field}`)}</dt>
              <dd>{context.overrides[field] || t("catalog.empty")}</dd></div>)}</dl>
        </details>}
        <ImportReviewContext source={draft.creating ? t("catalog.newSource") : source
          ? t("catalog.revisionLabel", { number: source.revisionNumber }) : t("catalog.savedSource")}
          target={targetName} effect={t(options.readonly ? "catalog.readEffect" : "catalog.effect")}
          selected={options.readonly ? undefined : draft.selected.length} total={catalogFieldNames.length} />
        {stale && <p className="catalog-review-message" role="alert">{t("catalog.stale")}</p>}
        {alreadyStored && <p className="catalog-review-message" role="status">{t("catalog.alreadyStored")}</p>}
        <p className="catalog-scroll-hint">{t("catalog.scrollHint")}</p>
        <div className="table-wrap catalog-review-table" tabIndex={0} aria-label={t("catalog.comparison")}>
          <table>
            <thead><tr>{!options.readonly && <th className="catalog-selection">{t("catalog.selection")}</th>}
              <th className="catalog-field">{t("vehicles.saveState.field")}</th>
              <th>{t("catalog.savedValue")}</th><th>{t("catalog.sourceValue")}</th>
              <th className="catalog-status">{t("importReview.state.same")} / {t("importReview.state.changed")}</th>
            </tr></thead>
            <tbody>{catalogFieldNames.map(field => {
              const different = draft.fields[field] !== context.currentFields[field];
              const individual = Object.hasOwn(context.overrides, field);
              return <tr key={field}>
                {!options.readonly && <td className="catalog-selection"><label className="checkbox-field">
                  <input type="checkbox" aria-label={`${t("catalog.select")}: ${t(`vehicle.field.${field}`)}`}
                    checked={draft.selected.includes(field)} disabled={disabled || !different}
                    onChange={() => model.toggleField(field)} />
                </label></td>}
                <th scope="row">{t(`vehicle.field.${field}`)}
                  {draft.creating && catalogRequiredFields.some(key => key === field) && <span aria-hidden="true"> *</span>}
                  {individual && <small>{t(context.authority === "legacy" ? "catalog.unresolved" : "catalog.individual")}</small>}
                </th>
                <td>{context.currentFields[field] || t("catalog.empty")}</td>
                <td>{draft.creating ? field === "description"
                  ? <textarea rows={2} value={draft.fields[field]} disabled={disabled}
                    aria-label={`${t("catalog.sourceValue")}: ${t(`vehicle.field.${field}`)}`}
                    onChange={event => model.updateField(field, event.target.value)} />
                  : <input value={draft.fields[field]} disabled={disabled}
                    aria-label={`${t("catalog.sourceValue")}: ${t(`vehicle.field.${field}`)}`}
                    aria-required={catalogRequiredFields.some(key => key === field)}
                    aria-invalid={missing.some(key => key === field) || undefined}
                    onChange={event => model.updateField(field, event.target.value)} />
                  : draft.fields[field] || t("catalog.empty")}</td>
                <td><ImportReviewStatus state={different ? "changed" : "same"} /></td>
              </tr>;
            })}</tbody>
          </table>
        </div>
        {draft.creating && <p>{t("catalog.requiredHint")}</p>}
        {draft.creating && <div className="catalog-new-options">
          <label className="checkbox-field catalog-checkbox"><input type="checkbox" checked={draft.independent} disabled={disabled}
            onChange={event => model.updateDraft({ independent: event.target.checked })} />{t("catalog.independent")}</label>
          <label>{t("catalog.note")}<textarea rows={2} value={draft.note} maxLength={2000} disabled={disabled}
            onChange={event => model.updateDraft({ note: event.target.value })} /></label>
        </div>}
        <div className="catalog-actions">
          {model.refreshPending && <button type="button" className="secondary-button" disabled={disabled}
            onClick={() => void model.refreshView()}>{t("catalog.refreshView")}</button>}
          <button type="button" className="secondary-button" disabled={disabled} onClick={() => void model.load()}>
            {t("catalog.reload")}
          </button>
          {!options.readonly && <button type="button" className="primary-button"
            disabled={disabled || stale || alreadyStored || unchanged || (!draft.creating && !available)}
            onClick={() => void model.confirm()}>
            {saving ? t("catalog.saving") : t(draft.creating ? "catalog.createConfirm" : "catalog.adoptConfirm")}
          </button>}
          {alreadyStored && <button type="button" className="secondary-button" disabled={disabled} onClick={model.newConfirmation}>
            {t("catalog.newConfirmation")}
          </button>}
        </div>
      </>}
      {!context && !loading && <button type="button" className="secondary-button" disabled={Boolean(options.busy)} onClick={() => void model.load(false)}>
        {t("catalog.reload")}
      </button>}
    </div>
  </details>;
}
