import { useEffect, useRef, useState } from "react";
import { api, ApiError, type Vehicle, type VehicleAttachment, type VehicleCVFile, type VehicleImage } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import type { VehicleUploadKind } from "./vehicleUploadBatch";

type Request = { kind: "image"; record: VehicleImage } | { kind: "attachment"; record: VehicleAttachment }
  | { kind: "decoder"; record: VehicleCVFile };
export type VehicleMediaDeletion = {
  kind: VehicleUploadKind; vehicleId: string; vehicleLabel: string; id: string; label: string;
  revision: string; references: number; busy: boolean; stale: boolean; error: string;
};
type Options = {
  selected: Vehicle | null; readonly: boolean; saving: boolean;
  setSaving: (value: boolean) => void; onMessage: (value: string) => void; onStatus: (value: string) => void;
  onImageDeleted: (id: string) => void;
  refreshSaved: (id: string, kind: VehicleUploadKind) => Promise<void>;
  reload: (id: string) => Promise<void>;
};

export function useVehicleMediaDeletion(options: Options) {
  const { t } = useI18n();
  const optionsRef = useRef(options); optionsRef.current = options;
  const [candidate, setCandidate] = useState<VehicleMediaDeletion | null>(null);
  const current = useRef(candidate), generation = useRef(0), busy = useRef(false);
  useEffect(() => () => { generation.current++; }, []);
  const set = (value: VehicleMediaDeletion | null) => { current.current = value; setCandidate(value); };
  const reset = () => {
    generation.current++;
    if (busy.current) optionsRef.current.setSaving(false);
    busy.current = false; set(null);
  };
  const prepare = ({ kind, record }: Request) => {
    const opts = optionsRef.current, vehicle = opts.selected;
    if (!vehicle || opts.readonly || opts.saving || busy.current || record.vehicleId !== vehicle.id) return;
    if (kind === "image" && "maintenanceId" in record && record.maintenanceId) {
      opts.onMessage(t("vehicles.uploads.maintenanceLinkedCannotDelete")); return;
    }
    if (!record.revision) { opts.onMessage(t("vehicles.mediaConflict.missingRevision")); return; }
    set({ kind, vehicleId: vehicle.id, vehicleLabel: `${vehicle.inventoryNumber} · ${vehicle.name}`,
      id: record.id, label: "originalName" in record ? record.originalName : "title" in record ? record.title || record.id : record.id,
      revision: record.revision, references: "sourceReferenceCount" in record ? record.sourceReferenceCount || 0 : 0,
      busy: false, stale: false, error: "" });
  };
  const close = () => { if (!busy.current) set(null); };
  const confirm = async () => {
    const value = current.current, opts = optionsRef.current;
    if (!value || value.stale || busy.current || opts.readonly || opts.saving || opts.selected?.id !== value.vehicleId) return;
    const version = ++generation.current;
    const active = () => generation.current === version && optionsRef.current.selected?.id === value.vehicleId;
    busy.current = true; opts.setSaving(true); opts.onMessage(""); set({ ...value, busy: true, error: "" });
    let deleted = false;
    try {
      if (value.kind === "image") await api.deleteVehicleImage(value.vehicleId, value.id, value.revision);
      else if (value.kind === "attachment") await api.deleteVehicleAttachment(value.vehicleId, value.id, value.revision);
      else await api.deleteVehicleCVFile(value.vehicleId, value.id, value.revision);
      deleted = true;
      if (!active()) return;
      if (value.kind === "image") optionsRef.current.onImageDeleted(value.id);
      await optionsRef.current.refreshSaved(value.vehicleId, value.kind);
      if (active()) { set(null); optionsRef.current.onStatus(t("vehicles.mediaDelete.completed", { name: value.label })); }
    } catch (error) {
      if (!active()) return;
      const stale = !deleted && error instanceof ApiError && error.status === 409;
      const missing = !deleted && error instanceof ApiError && error.status === 404;
      if (stale || missing) {
        try { await optionsRef.current.reload(value.vehicleId); } catch { /* Keep the original guarded action blocked. */ }
      }
      if (active()) set({ ...value, busy: false, stale: stale || missing || deleted,
        error: deleted ? t("vehicles.mediaDelete.refreshFailed") : stale || missing ? t("vehicles.mediaDelete.reviewAgain")
          : error instanceof Error ? error.message : String(error) });
    } finally {
      if (generation.current === version) {
        busy.current = false; optionsRef.current.setSaving(false);
        if (current.current?.busy) set({ ...current.current, busy: false });
      }
    }
  };
  return { candidate, commands: { prepare, confirm, close, reset } };
}
