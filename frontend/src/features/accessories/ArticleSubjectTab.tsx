import { useEffect, useMemo, useState } from "react";

import {
  api,
  type AccessoryAttributeValue,
  type MasterDataEntry
} from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppDateInput } from "../../shared/ui/AppDateInput";
import { AppMultiSelect } from "../../shared/ui/AppMultiSelect";
import { AppNumberInput } from "../../shared/ui/AppNumberInput";
import { AppSelect } from "../../shared/ui/AppSelect";
import { AppTextInput } from "../../shared/ui/AppTextInput";
import type { ArticleEditorForm } from "./articleEditorModel";
import {
  customFieldDefinitions,
  fieldDefinitionsForType,
  type ArticleSubjectFieldDefinition,
  type CustomArticleSubjectFieldDefinition
} from "./articleTypeFields";
import { articleTypeLabel } from "./articleTypes";

type SubjectDefinition = ArticleSubjectFieldDefinition | CustomArticleSubjectFieldDefinition;

function isCustomDefinition(definition: SubjectDefinition): definition is CustomArticleSubjectFieldDefinition {
  return "label" in definition;
}

function replaceAttribute(
  attributes: readonly AccessoryAttributeValue[],
  next: AccessoryAttributeValue | null,
  key: string
): AccessoryAttributeValue[] {
  const remaining = attributes.filter((attribute) => attribute.key !== key);
  return next ? [...remaining, next] : remaining;
}

function attributeFor(
  attributes: readonly AccessoryAttributeValue[],
  definition: SubjectDefinition
): AccessoryAttributeValue | undefined {
  return attributes.find((attribute) => attribute.key === definition.key && attribute.kind === definition.kind);
}

export function ArticleSubjectTab({
  form,
  disabled,
  error,
  active = true,
  customFieldEntries,
  customFields,
  articleTypeEntries = [],
  loading = false,
  loadError: externalLoadError = "",
  subjectFieldErrors = {},
  onChange
}: {
  form: ArticleEditorForm;
  disabled: boolean;
  error?: string;
  active?: boolean;
  customFieldEntries?: readonly MasterDataEntry[];
  customFields?: readonly CustomArticleSubjectFieldDefinition[];
  articleTypeEntries?: readonly MasterDataEntry[];
  loading?: boolean;
  loadError?: string;
  subjectFieldErrors?: Readonly<Record<string, string>>;
  onChange: (patch: Partial<ArticleEditorForm>) => void;
}) {
  const { t } = useI18n();
  const [loadedCustomFields, setLoadedCustomFields] = useState<readonly MasterDataEntry[]>(customFieldEntries || []);
  const [loadingCustomFields, setLoadingCustomFields] = useState(false);
  const [loadError, setLoadError] = useState("");

  useEffect(() => {
    if (customFieldEntries) setLoadedCustomFields(customFieldEntries);
  }, [customFieldEntries]);

  useEffect(() => {
    if (!active || form.articleType !== "other" || customFieldEntries || customFields) return;
    let current = true;
    setLoadingCustomFields(true);
    setLoadError("");
    void api.masterData("accessory_custom_field", true).then((entries) => {
      if (current) setLoadedCustomFields(entries);
    }).catch(() => {
      if (current) setLoadError(t("accessories.subject.customLoadError"));
    }).finally(() => {
      if (current) setLoadingCustomFields(false);
    });
    return () => { current = false; };
  }, [active, customFieldEntries, customFields, form.articleType, t]);

  const customDefinitions = useMemo(() => customFields || customFieldDefinitions(loadedCustomFields),
    [customFields, loadedCustomFields]);
  const definitions = fieldDefinitionsForType(form.articleType, customDefinitions);
  const labelFor = (definition: SubjectDefinition) => {
    const label = isCustomDefinition(definition) ? definition.label : t(definition.labelKey);
    return definition.unit ? `${label} (${definition.unit})` : label;
  };
  const optionLabel = (definition: SubjectDefinition, option: string) =>
    isCustomDefinition(definition) ? option : t(`accessories.subject.option.${option}`);

  const setNumber = (definition: SubjectDefinition, draft: string) => {
    const normalized = draft.replace(",", ".");
    const numberValue = Number(normalized);
    const nextAttribute = draft.trim() !== "" && Number.isFinite(numberValue)
      ? { key: definition.key, kind: "number" as const, numberValue,
          ...(definition.unit ? { unit: definition.unit } : {}) }
      : null;
    onChange({
      attributeNumberDrafts: { ...form.attributeNumberDrafts, [definition.key]: draft },
      attributes: replaceAttribute(form.attributes, nextAttribute, definition.key)
    });
  };

  const renderField = (definition: SubjectDefinition) => {
    const value = attributeFor(form.attributes, definition);
    const label = labelFor(definition);
    const helpText = !isCustomDefinition(definition) && definition.helpKey ? t(definition.helpKey) : undefined;
    const fieldError = subjectFieldErrors[definition.key];
    switch (definition.kind) {
    case "text":
      return <AppTextInput key={definition.key} label={label} helpText={helpText} error={fieldError} disabled={disabled}
        value={value?.kind === "text" ? value.textValue : ""}
        onChange={(event) => onChange({ attributes: replaceAttribute(form.attributes,
          event.target.value ? { key: definition.key, kind: "text", textValue: event.target.value } : null,
          definition.key) })} />;
    case "number":
      return <AppNumberInput key={definition.key} label={label} helpText={helpText} error={fieldError} disabled={disabled}
        min={definition.min} max={definition.max} step={definition.step || "any"}
        value={form.attributeNumberDrafts[definition.key] ??
          (value?.kind === "number" ? String(value.numberValue) : "")}
        onValueChange={(draft) => setNumber(definition, draft)} />;
    case "boolean":
      return <label key={definition.key} className={`app-field article-checkbox ${fieldError ? "has-error" : ""}`}>
        <span className="app-field-label">{label}</span>
        <span className="article-checkbox-control"><input type="checkbox" disabled={disabled}
          aria-label={label}
          checked={value?.kind === "boolean" ? value.booleanValue : false}
          onChange={(event) => onChange({ attributes: replaceAttribute(form.attributes,
            { key: definition.key, kind: "boolean", booleanValue: event.target.checked }, definition.key) })} />
        {t("accessories.subject.booleanHint")}</span>
        {fieldError ? <span className="app-field-error" role="alert">{fieldError}</span> : null}
      </label>;
    case "date": {
      const id = `article-subject-${definition.key}`;
      return <label key={definition.key} className={`app-field ${fieldError ? "has-error" : ""}`} htmlFor={id}>
        <span className="app-field-label">{label}</span>
        <AppDateInput id={id} aria-label={label} aria-invalid={fieldError ? true : undefined} disabled={disabled}
          value={value?.kind === "date" ? value.dateValue : ""}
          onChange={(event) => onChange({ attributes: replaceAttribute(form.attributes,
            event.target.value ? { key: definition.key, kind: "date", dateValue: event.target.value } : null,
            definition.key) })} />
        {fieldError ? <span className="app-field-error" role="alert">{fieldError}</span> : null}
      </label>;
    }
    case "single_select": {
      const id = `article-subject-${definition.key}`;
      const errorId = fieldError ? `${id}-error` : undefined;
      return <label key={definition.key} className={`app-field ${fieldError ? "has-error" : ""}`}>
        <span className="app-field-label">{label}</span>
        <AppSelect id={id} value={value?.kind === "single_select" ? value.optionValues[0] : ""}
          disabled={disabled} aria-label={label} aria-invalid={fieldError ? true : undefined}
          aria-describedby={errorId}
          onChange={(event) => onChange({ attributes: replaceAttribute(form.attributes,
            event.target.value ? { key: definition.key, kind: "single_select",
              optionValues: [event.target.value] } : null, definition.key) })}>
          <option value="">{t("accessories.subject.selectPlaceholder")}</option>
          {(definition.options || []).map((option) =>
            <option key={option} value={option}>{optionLabel(definition, option)}</option>)}
        </AppSelect>
        {fieldError ? <span id={errorId} className="app-field-error" role="alert">{fieldError}</span> : null}
      </label>;
    }
    case "multi_select":
      return <AppMultiSelect key={definition.key} label={label} helpText={helpText} error={fieldError} disabled={disabled}
        value={value?.kind === "multi_select" ? value.optionValues : []}
        options={(definition.options || []).map((option) => ({
          value: option,
          label: optionLabel(definition, option)
        }))}
        placeholder={t("accessories.subject.multiPlaceholder")}
        onValueChange={(options) => onChange({ attributes: replaceAttribute(form.attributes,
          options.length > 0 ? { key: definition.key, kind: "multi_select", optionValues: options } : null,
          definition.key) })} />;
    }
  };

  return <section className="article-editor-tab article-subject-tab" data-testid="article-subject-tab"
    aria-label={t("accessories.editor.tabs.subject", {
      type: articleTypeLabel(form.articleType, articleTypeEntries, t)
    })}>
    {error ? <p className="form-message" role="alert">{error}</p> : null}
    {loading || loadingCustomFields ? <p className="loading-cell">{t("accessories.subject.customLoading")}</p> : null}
    {(!customFields && externalLoadError) || loadError
      ? <p className="form-message" role="alert">{externalLoadError || loadError}</p> : null}
    {!loading && !loadingCustomFields && !externalLoadError && !loadError && definitions.length === 0
      ? <p className="article-editor-hint">{t("accessories.subject.empty")}</p>
      : <div className="article-editor-grid article-subject-grid">{definitions.map(renderField)}</div>}
  </section>;
}
