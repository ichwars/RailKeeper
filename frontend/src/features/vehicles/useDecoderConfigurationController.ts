import { useCallback, useEffect, useRef, useState } from "react";
import { api, ApiError, type DecoderConfigurationContext, type Vehicle } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { decoderConfigurationDraft, decoderDraftFromCurrent, decoderDraftStored, decoderDraftValid, saveDecoderConfigurationDraft,
  type DecoderCommand, type DecoderConfigurationDraft } from "./decoderConfigurationDraft";
import { reconcileDraft, retainRefreshConflicts, type DraftRefreshConflict } from "./vehicleDraftRefresh";

type Options = {
  ownerKey: string;
  selected: Vehicle | null;
  active: boolean;
  readonly: boolean;
  busy: boolean;
  setBusy: (value: boolean) => void;
  captureScope: () => () => boolean;
};
type ScopedState = { key: string; context: DecoderConfigurationContext | null; draft: DecoderConfigurationDraft | null;
  baseline: DecoderConfigurationDraft | null; conflicts: DraftRefreshConflict[];
  loading: boolean; saving: boolean; needsReview: boolean; error: string; status: string };
const emptyState = (key: string): ScopedState => ({ key, context: null, draft: null,
  baseline: null, conflicts: [], loading: false, saving: false, needsReview: false, error: "", status: "" });

export function useDecoderConfigurationController(options: Options) {
  const { t } = useI18n();
  const translate = useRef(t); translate.current = t;
  const vehicleId = options.selected?.id || "", key = `${options.ownerKey}:${vehicleId}`;
  const latest = useRef(options); latest.current = options;
  const scope = useRef(key); scope.current = key;
  const request = useRef(0), writing = useRef(false);
  const [stored, setStored] = useState<ScopedState>(() => emptyState(key));
  const state = stored.key === key ? stored : emptyState(key);
  const stateRef = useRef(state); stateRef.current = state;
  const comparedCVs = useRef<{ key: string; values: Vehicle["cvValues"]; files: Vehicle["cvFiles"] } | null>(null);
  useEffect(() => {
    setStored(emptyState(key)); writing.current = false;
    return () => { request.current++; };
  }, [key]);

  const load = useCallback(async (compare = false) => {
    if (!vehicleId || writing.current || latest.current.busy) return;
    const sequence = ++request.current, validScope = latest.current.captureScope();
    const valid = () => sequence === request.current && scope.current === key && validScope();
    setStored(current => ({ ...(current.key === key ? current : emptyState(key)), loading: true, error: "" }));
    try {
      const context = await api.decoderConfigurationContext(vehicleId);
      if (!valid()) return;
      if (context.vehicleId !== vehicleId) throw new Error("Unexpected decoder context");
      setStored(current => {
        if (current.key !== key) return current;
        const acknowledged = current.draft && decoderDraftStored(current.draft, context);
        let draft = current.draft, baseline = current.baseline, conflicts = current.conflicts;
        if (acknowledged) { draft = null; baseline = null; conflicts = []; }
        else if (draft && baseline && compare) {
          const incoming = decoderDraftFromCurrent(draft, context);
          const refreshed = reconcileDraft(draft, baseline, incoming);
          conflicts = retainRefreshConflicts(conflicts, refreshed.value, incoming, refreshed.conflicts);
          draft = refreshed.value; baseline = incoming;
        }
        return { ...current, context, draft, baseline, conflicts,
          needsReview: acknowledged || compare ? false : current.needsReview,
          status: acknowledged ? translate.current("decoderManage.alreadySaved") : current.status };
      });
    } catch {
      if (valid()) setStored(current => ({ ...current, error: translate.current("decoderManage.loadFailed") }));
    } finally {
      if (valid()) setStored(current => ({ ...current, loading: false }));
    }
  }, [key, vehicleId]);
  useEffect(() => {
    if (!options.active || options.busy || !vehicleId) return;
    const previous = comparedCVs.current;
    if (!stateRef.current.context || previous?.key !== key || previous.values !== options.selected?.cvValues || previous.files !== options.selected?.cvFiles) {
      comparedCVs.current = { key, values: options.selected?.cvValues, files: options.selected?.cvFiles };
      void load();
    }
  }, [options.active, options.busy, options.selected?.cvValues, options.selected?.cvFiles, key, vehicleId, load]);

  const start = (kind: DecoderCommand, targetId = "") => {
    if (!state.context || state.draft || options.readonly || options.busy || state.loading || writing.current) return;
    const draft = decoderConfigurationDraft(kind, state.context, targetId);
    setStored(current => ({ ...current, draft, baseline: draft, conflicts: [], needsReview: false, error: "", status: "" }));
  };
  const update = (patch: Partial<Pick<DecoderConfigurationDraft, "label" | "manufacturer" | "model" | "serialNumber" |
    "notes" | "profile" | "decoderId" | "installationId" | "installedAt" | "removedAt" | "note">>) => {
    if (writing.current || options.busy || options.readonly) return;
    setStored(current => {
      const draft = current.draft ? { ...current.draft, ...patch } : null;
      return { ...current, draft, conflicts: draft && current.baseline
        ? retainRefreshConflicts(current.conflicts, draft, current.baseline, []) : [], error: "" };
    });
  };
  const discard = () => {
    if (writing.current || options.busy) return;
    setStored(current => ({ ...current, draft: null, baseline: null, conflicts: [], needsReview: false, error: "" }));
  };
  const resolve = (field: string, useSaved: boolean) => {
    if (writing.current || options.busy || options.readonly) return;
    setStored(current => {
      const conflict = current.conflicts.find(value => value.field === field);
      if (!conflict || !current.draft) return current;
      return { ...current, draft: useSaved ? { ...current.draft, [field]: conflict.saved } : current.draft,
        conflicts: current.conflicts.filter(value => value.field !== field) };
    });
  };
  const confirm = async () => {
    const { context, draft, loading, needsReview, conflicts } = stateRef.current;
    if (!context || !draft || loading || needsReview || conflicts.length || options.readonly || options.busy || writing.current ||
      context.revision !== draft.expectedRevision || !decoderDraftValid(draft, context)) return;
    const sequence = ++request.current, validScope = options.captureScope();
    const valid = () => sequence === request.current && scope.current === key && validScope();
    writing.current = true; options.setBusy(true);
    setStored(current => ({ ...current, saving: true, error: "", status: "" }));
    try {
      const context = await saveDecoderConfigurationDraft(vehicleId, draft);
      if (!valid()) return;
      if (context.vehicleId !== vehicleId) throw new Error("Unexpected decoder context");
      setStored(current => ({ ...current, context, draft: null, baseline: null, conflicts: [], needsReview: false, status: t("decoderManage.saved") }));
    } catch (reason) {
      if (!valid()) return;
      const invalid = reason instanceof ApiError && reason.status === 400;
      setStored(current => ({ ...current, needsReview: !invalid, error: t(invalid ? "decoderManage.invalid"
        : reason instanceof ApiError && reason.status === 409 ? "decoderManage.conflict" : "decoderManage.unconfirmed") }));
    } finally {
      if (valid()) {
        writing.current = false; latest.current.setBusy(false);
        setStored(current => ({ ...current, saving: false }));
      }
    }
  };
  return { ...state, readonly: options.readonly, busy: options.busy,
    stale: Boolean(state.draft && state.context && state.draft.expectedRevision !== state.context.revision),
    load, start, update, discard, resolve, confirm };
}

export type DecoderConfigurationController = ReturnType<typeof useDecoderConfigurationController>;
