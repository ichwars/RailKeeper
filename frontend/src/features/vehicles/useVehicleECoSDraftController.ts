import { useCallback, useEffect, useState } from "react";

import { api, type CreateVehicleRequest, type Vehicle } from "../../shared/api";
import type { FunctionEditState, PendingArticleImage } from "./vehicleTransforms";
import { vehicleToForm } from "./vehicleTransforms";
import {
  compactValue,
  ecosImportSessionStorageKey,
  ecosRequiredFields,
  ecosVehicleDraftStorageKey,
  emptyVehicle,
  type ECoSRequiredField,
  type ECoSVehicleDraftPayload
} from "./vehicleViewModel";

type Translator = (key: string, values?: Record<string, string | number>) => string;

type PreparedECoSDraft = {
  form: CreateVehicleRequest;
  images: PendingArticleImage[];
  functionEdits: FunctionEditState;
  mergeImages: (current: PendingArticleImage[]) => PendingArticleImage[];
};

type UseVehicleECoSDraftControllerOptions = {
  onOpenCreate: (prepared: PreparedECoSDraft) => void;
  onOpenUpdate: (detail: Vehicle, prepared: PreparedECoSDraft) => void;
  onFinishOpen: (draft: ECoSVehicleDraftPayload) => void;
  onMessage: (message: string) => void;
  t: Translator;
};

export function useVehicleECoSDraftController({
  onOpenCreate,
  onOpenUpdate,
  onFinishOpen,
  onMessage,
  t
}: UseVehicleECoSDraftControllerOptions) {
  const [draft, setDraft] = useState<ECoSVehicleDraftPayload | null>(null);
  const [unclearFields, setUnclearFields] = useState<Set<ECoSRequiredField>>(() => new Set());

  const clear = () => {
    setDraft(null);
    setUnclearFields(new Set());
  };

  const syncUnclearFields = (nextForm: CreateVehicleRequest) => {
    if (!draft) return;
    setUnclearFields((current) => {
      const next = new Set(current);
      ecosRequiredFields.forEach((field) => {
        if (compactValue(nextForm[field])) next.delete(field);
        else if (draft.unclearFields.includes(field)) next.add(field);
      });
      return next;
    });
  };

  const fieldClass = (field: ECoSRequiredField) => (
    draft && unclearFields.has(field) ? "ecos-unclear-field" : ""
  );

  const openDraft = useCallback((nextDraft: ECoSVehicleDraftPayload) => {
    const images: PendingArticleImage[] = [];
    const mergeImages = (current: PendingArticleImage[]) => current;
    const functionEdits: FunctionEditState = Object.fromEntries((nextDraft.functionValues || []).map((item) => [
      item.functionKey,
      {
        name: item.name || "",
        symbolKey: item.symbolKey || "",
        functionType: item.functionType || "standard",
        mode: item.mode || "dauer",
        directionDependent: Boolean(item.directionDependent),
        notes: item.notes || "",
        persisted: false
      }
    ]));
    const applyDraftValues = (base: CreateVehicleRequest) => {
      const next = { ...base };
      const keys = nextDraft.importedKeys?.length
        ? nextDraft.importedKeys
        : Object.keys(nextDraft.vehicle) as (keyof CreateVehicleRequest)[];
      keys.forEach((key) => {
        const value = nextDraft.vehicle[key];
        if (typeof value === "boolean" || (typeof value === "string" && value.trim() !== "")) {
          (next as Record<string, unknown>)[key] = value;
        }
      });
      return next;
    };
    const finish = () => {
      setDraft(nextDraft);
      setUnclearFields(new Set(nextDraft.unclearFields));
      onFinishOpen(nextDraft);
    };

    if (nextDraft.mode === "update" && nextDraft.targetVehicleId) {
      api.vehicle(nextDraft.targetVehicleId)
        .then((detail) => {
          onOpenUpdate(detail, {
            form: applyDraftValues(vehicleToForm(detail)),
            images,
            functionEdits,
            mergeImages
          });
          finish();
        })
        .catch((error: Error) => onMessage(error.message));
      return;
    }

    onOpenCreate({
      form: applyDraftValues(emptyVehicle),
      images,
      functionEdits,
      mergeImages
    });
    finish();
  }, [onFinishOpen, onMessage, onOpenCreate, onOpenUpdate]);

  useEffect(() => {
    const rawDraft = window.sessionStorage.getItem(ecosVehicleDraftStorageKey);
    if (!rawDraft) return;

    try {
      const storedDraft = JSON.parse(rawDraft) as ECoSVehicleDraftPayload;
      if (storedDraft?.source === "ecos" || storedDraft?.source === "cs3") openDraft(storedDraft);
    } catch {
      onMessage(t("vehicles.ecosDraft.invalid"));
    } finally {
      window.sessionStorage.removeItem(ecosVehicleDraftStorageKey);
      const source = new URLSearchParams(window.location.search).get("source");
      if (source === "ecos" || source === "cs3") {
        window.history.replaceState(null, "", "/vehicles");
      }
    }
  }, [onMessage, openDraft, t]);

  const markImportSessionSaved = (savedDraft: ECoSVehicleDraftPayload, vehicleId: string) => {
    if (!savedDraft.returnToEcos) return;
    try {
      const rawSession = window.sessionStorage.getItem(ecosImportSessionStorageKey);
      if (!rawSession) return;
      const session = JSON.parse(rawSession) as {
        id?: string;
        statuses?: Record<string, { status: string; vehicleId?: string; updatedAt?: string }>;
        updatedAt?: string;
      };
      if (session.id !== savedDraft.returnToEcos.sessionId) return;
      const key = String(savedDraft.returnToEcos.objectId);
      const now = new Date().toISOString();
      window.sessionStorage.setItem(ecosImportSessionStorageKey, JSON.stringify({
        ...session,
        updatedAt: now,
        statuses: {
          ...(session.statuses || {}),
          [key]: { ...(session.statuses || {})[key], status: "saved", vehicleId, updatedAt: now }
        }
      }));
    } catch {
      window.sessionStorage.removeItem(ecosImportSessionStorageKey);
    }
  };

  const returnToImportSession = (savedDraft: ECoSVehicleDraftPayload) => {
    const destination = savedDraft.returnToDigitalCenters
      ? `/digital-centers?sessionId=${encodeURIComponent(savedDraft.returnToDigitalCenters.sessionId)}` +
        `&objectId=${encodeURIComponent(savedDraft.returnToDigitalCenters.objectId)}`
      : "/import-export?source=ecos";
    window.history.pushState(null, "", destination);
    window.dispatchEvent(new PopStateEvent("popstate"));
  };

  return {
    state: { draft, unclearFields },
    setters: { setDraft, setUnclearFields },
    commands: { clear, syncUnclearFields, fieldClass, markImportSessionSaved, returnToImportSession }
  };
}
