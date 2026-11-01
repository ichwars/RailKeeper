import { api } from "../../shared/api";
import type { DigitalCenterWorkItem } from "./digitalCenterModel";

export async function readECoSStoredCVs(
  item: DigitalCenterWorkItem,
  onProgress: (item: DigitalCenterWorkItem) => void,
  options: { restart?: boolean; refresh?: boolean; useSaved?: boolean; isCancelled?: () => boolean } = {}
): Promise<DigitalCenterWorkItem | null> {
  if (options.isCancelled?.()) return null;
  if (options.useSaved && hasCompleteECoSCVSnapshot(item)) return item;
  if (options.useSaved && item.center.cvSources) {
    return { ...item, center: { ...item.center, cvs: [] } };
  }
  // One complete database export includes all stored pages. A TCP scan over
  // CV1–1024 silently misses registers such as ECoS address 4099:257.
  const snapshot = await api.readDigitalCenterCVBatch(item.sessionId, item.id, 1,
    { restart: options.restart === true, source: "database" });
  if (snapshot.center.cvRead?.source !== "database" || snapshot.center.cvRead.nextCV !== 1025) {
    throw new Error("incomplete CV database snapshot");
  }
  if (options.isCancelled?.()) return null;
  onProgress(snapshot);
  return snapshot;
}

export function hasCompleteECoSCVSnapshot(item: DigitalCenterWorkItem): boolean {
  const state = item.center.cvRead;
  return Boolean((state?.source === "database" || state?.source === "backup") &&
    state.nextCV === 1025 && state.unavailable === 0 && state.readAt && state.sourceHost &&
    item.center.cvs);
}
