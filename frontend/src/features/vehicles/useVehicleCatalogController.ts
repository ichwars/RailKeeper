import { useEffect, useRef, useState } from "react";
import { api, ApiError, type VehicleCatalogContext, type VehicleCatalogFields } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { newVehicleCreationIdentity as newCatalogIdentity } from "./vehicleCreationIdentity";
import {
  catalogDraftFrom, catalogSelection, missingCatalogFields, sameCatalogFields, type CatalogDraft, type CatalogField
} from "./vehicleCatalogModel";

export type VehicleCatalogOptions = {
  targetType: "vehicle" | "set";
  targetId: string;
  savedFields?: VehicleCatalogFields;
  readonly?: boolean;
  busy?: boolean;
  captureScope?: () => () => boolean;
  onSaved?: (confirmed: boolean) => Promise<void>;
  onPendingChange?: (pending: boolean) => void;
};

export function useVehicleCatalogController({ targetType, targetId, savedFields, readonly = false,
  busy = false, captureScope, onSaved, onPendingChange }: VehicleCatalogOptions) {
  const { t } = useI18n();
  const [context, setContext] = useState<VehicleCatalogContext | null>(null);
  const [draft, setDraft] = useState<CatalogDraft | null>(null);
  const [comparedFields, setComparedFields] = useState<VehicleCatalogFields | null>(null);
  const [loading, setLoading] = useState(false), [saving, setSaving] = useState(false);
  const [refreshPending, setRefreshPending] = useState(false);
  const [error, setError] = useState(""), [status, setStatus] = useState("");
  const mounted = useRef(true), request = useRef(0), writing = useRef(false);
  const confirmedWrite = useRef(false);
  useEffect(() => {
    mounted.current = true;
    setContext(null); setDraft(null); setComparedFields(null); setLoading(false); setSaving(false);
    setRefreshPending(false);
    confirmedWrite.current = false;
    setError(""); setStatus(""); writing.current = false;
    return () => { mounted.current = false; request.current++; };
  }, [targetType, targetId]);

  const capture = () => {
    const version = ++request.current, isCurrent = captureScope?.() ?? (() => true);
    return () => mounted.current && version === request.current && isCurrent();
  };
  const load = async (keepDraft = true) => {
    if (writing.current || busy) return;
    const isCurrent = capture();
    if (!isCurrent()) return;
    setLoading(true); setError(""); setStatus("");
    try {
      const result = await api.catalogContext(targetType, targetId);
      if (!isCurrent()) return;
      setContext(result); setComparedFields(savedFields ? { ...savedFields } : null);
      setDraft(current => keepDraft && current ? current : catalogDraftFrom(result));
    } catch {
      if (isCurrent()) setError(t("catalog.loadFailed"));
    } finally {
      if (isCurrent()) setLoading(false);
    }
  };
  const stale = Boolean(context && savedFields && comparedFields &&
    !sameCatalogFields(savedFields, comparedFields) && !sameCatalogFields(savedFields, context.currentFields));
  const alreadyStored = Boolean(draft?.creating && context?.revisions.some(revision => revision.id === draft.identity));
  const available = Boolean(draft && context?.revisions.some(revision => revision.id === draft.revisionID));

  const chooseRevision = (id: string) => {
    const revision = context?.revisions.find(value => value.id === id);
    if (revision && context && !draft?.creating) { setDraft(catalogDraftFrom(context, revision)); setError(""); setStatus(""); }
  };
  const startRevision = () => {
    if (!context || !draft || readonly || busy || saving) return;
    const fields = context.authority === "legacy" && draft.revisionID === context.pinned?.id
      ? { ...draft.fields, ...context.overrides } : { ...draft.fields };
    setDraft({ ...draft, fields, selected: catalogSelection(context, fields), creating: true,
      identity: newCatalogIdentity(), independent: !context.pinned, note: "" });
    setError(""); setStatus("");
  };
  const discardRevision = () => {
    if (context) { setDraft(catalogDraftFrom(context)); setError(""); setStatus(""); }
  };
  const updateField = (field: CatalogField, value: string) => {
    if (!context) return;
    setDraft(current => {
      if (!current?.creating) return current;
      const selected = current.selected.filter(key => key !== field);
      if (value !== context.currentFields[field] &&
        (!Object.hasOwn(context.overrides, field) || current.selected.includes(field))) selected.push(field);
      return { ...current, fields: { ...current.fields, [field]: value }, selected };
    });
  };
  const toggleField = (field: CatalogField) => setDraft(current => current ? {
    ...current, selected: current.selected.includes(field) ? current.selected.filter(key => key !== field) : [...current.selected, field]
  } : current);

  const confirm = async () => {
    if (!context || !draft || readonly || busy || loading || writing.current || stale || alreadyStored ||
      (!draft.creating && !available)) return;
    if (draft.creating && missingCatalogFields(draft.fields).length) { setError(t("catalog.required")); return; }
    const isCurrent = capture();
    if (!isCurrent()) return;
    writing.current = true; confirmedWrite.current = false;
    setSaving(true); setRefreshPending(true); setError(""); setStatus(""); onPendingChange?.(true);
    let acknowledged = false;
    try {
      const result = draft.creating
        ? await api.createCatalogRevision(targetType, targetId, { id: draft.identity, expectedRevision: context.revision,
          fields: draft.fields, note: draft.note, independent: draft.independent, selectedFields: draft.selected })
        : await api.adoptCatalogRevision(targetType, targetId, { expectedRevision: context.revision,
          revisionId: draft.revisionID, selectedFields: draft.selected });
      acknowledged = true;
      confirmedWrite.current = true;
      if (!isCurrent()) return;
      setContext(result); setDraft(catalogDraftFrom(result)); setComparedFields(savedFields ? { ...savedFields } : null);
      setStatus(t("catalog.saved"));
      setRefreshPending(true);
      await onSaved?.(true);
      if (isCurrent()) { setRefreshPending(false); onPendingChange?.(false); }
    } catch (reason) {
      if (!isCurrent()) return;
      setError(acknowledged ? t("catalog.refreshFailed") : reason instanceof ApiError && reason.status === 409
        ? t("catalog.conflict") : reason instanceof ApiError && reason.status === 400
          ? t("catalog.invalid") : t("catalog.unconfirmed"));
    } finally {
      if (isCurrent()) {
        writing.current = false; setSaving(false);
      }
    }
  };

  const refreshView = async () => {
    if (!refreshPending || writing.current || busy || loading) return;
    const isCurrent = capture();
    if (!isCurrent()) return;
    setLoading(true); setError("");
    try {
      await onSaved?.(confirmedWrite.current);
      if (isCurrent()) { setRefreshPending(false); onPendingChange?.(false); }
    } catch {
      if (isCurrent()) setError(t("catalog.refreshFailed"));
    } finally {
      if (isCurrent()) setLoading(false);
    }
  };

  return { context, draft, loading, saving, error, status, stale, alreadyStored, available, refreshPending, refreshView,
    load, chooseRevision, startRevision, discardRevision, updateField, toggleField, confirm,
    updateDraft: (patch: Partial<Pick<CatalogDraft, "note" | "independent">>) => setDraft(current => current ? { ...current, ...patch } : current),
    newConfirmation: () => { setDraft(current => current ? { ...current, identity: newCatalogIdentity() } : current); setError(""); } };
}
