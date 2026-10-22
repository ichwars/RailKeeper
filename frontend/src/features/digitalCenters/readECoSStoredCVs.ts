import { api } from "../../shared/api";
import type { DigitalCenterWorkItem } from "./digitalCenterModel";

export async function readECoSStoredCVs(
  item: DigitalCenterWorkItem,
  onProgress: (item: DigitalCenterWorkItem) => void,
  options: { restart?: boolean; isCancelled?: () => boolean } = {}
): Promise<DigitalCenterWorkItem | null> {
  let snapshot = item;
  let start = options.restart ? 1 : item.center.cvRead?.nextCV || 1;
  while (start <= 1024 && !options.isCancelled?.()) {
    const next = await api.readDigitalCenterCVBatch(item.sessionId, item.id, start);
    const cursor = next.center.cvRead?.nextCV;
    if (!cursor || cursor <= start || cursor > 1025) throw new Error("invalid CV cursor");
    onProgress(next);
    snapshot = next;
    start = cursor;
  }
  return options.isCancelled?.() ? null : snapshot;
}
