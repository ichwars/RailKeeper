import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from "react";
import { api } from "../../shared/api";
import { useDecoderLibrary } from "../../shared/DecoderLibraryProvider";
import type { DecoderCVReference, DecoderReading } from "../../shared/decoderLibraryModel";
import { useI18n } from "../../shared/i18n";
import { decoderAnalysis } from "./decoderAnalysis";
import { manufacturerCVHints, manufacturerReferenceModel, type ManufacturerCVHint } from "./manufacturerCVHints";

type HintState = {
  lookup: (value: DecoderReading) => ManufacturerCVHint | undefined;
  status: "idle" | "loading" | "ready" | "missing" | "failed"; retry: () => void;
};
const Context = createContext<HintState>({ lookup: () => undefined, status: "idle", retry: () => {} });
export const useManufacturerCVHints = () => useContext(Context);

export function ManufacturerCVProvider({ values, profile, manualType, children }: {
  values: DecoderReading[]; profile: string; manualType: string; children: ReactNode;
}) {
  const library = useDecoderLibrary();
  const model = useMemo(() => manufacturerReferenceModel(
    decoderAnalysis(values, profile, library.profiles, library.inventory ?? null), library.inventory, profile, manualType
  ), [values, profile, manualType, library]);
  const [attempt, setAttempt] = useState(0);
  const [result, setResult] = useState<{
    modelId: string; library: typeof library; attempt: number; references?: DecoderCVReference[]; failed?: boolean;
  }>();
  useEffect(() => {
    if (!model) return;
    let current = true;
    void api.decoderReferences(model.id).then(references => {
      if (current) setResult({ modelId: model.id, library, attempt, references });
    }).catch(() => {
      if (current) setResult({ modelId: model.id, library, attempt, failed: true });
    });
    return () => { current = false; };
  }, [model, library, attempt]);
  const state = useMemo<HintState>(() => {
    const retry = () => setAttempt(value => value + 1);
    const empty = { lookup: () => undefined, retry };
    if (!model) return { ...empty, status: "idle" };
    if (!result || result.modelId !== model.id || result.library !== library || result.attempt !== attempt)
      return { ...empty, status: "loading" };
    if (result.failed) return { ...empty, status: "failed" };
    if (!result.references?.some(reference => reference.active && reference.modelIds.includes(model.id) && reference.rows.length))
      return { ...empty, status: "missing" };
    const lookup = manufacturerCVHints(model, result.references ?? []);
    return { status: "ready", retry, lookup: value =>
      (value.decoderProfile || "") === profile ? lookup(value) : undefined };
  }, [model, result, library, attempt, profile]);
  return <Context.Provider value={state}>{children}</Context.Provider>;
}

export function ManufacturerCVStatus() {
  const { status, retry } = useManufacturerCVHints();
  const { t } = useI18n();
  if (status === "idle") return null;
  return <div className="decoder-library-note" role="status">
    <span>{t(status === "failed" ? "decoderLibrary.referencesFailed" : status === "loading"
      ? "decoderLibrary.hintsLoading" : status === "missing" ? "decoderLibrary.referencesMissing" : "decoderLibrary.hintsScope")}</span>
    {status === "failed" && <button type="button" className="secondary-button" onClick={retry}>
      {t("decoderLibrary.hintsRetry")}</button>}
  </div>;
}
