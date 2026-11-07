import type { Vehicle } from "../../shared/api";
import { draftValue, reconcileDraft, reconcileImageDraft, retainRefreshConflicts, sameDraftValue,
  type DraftRefreshConflict } from "./vehicleDraftRefresh";
import { attachmentsToEditState, vehicleImagesToPending, type AttachmentEditState,
  type PendingArticleImage } from "./vehicleTransforms";

export type MediaDraftConflict = {
  kind: "image" | "attachment";
  id: string;
  label: string;
  deleted: boolean;
  fields: DraftRefreshConflict[];
};
export type MediaMetadataDraft = { edits: AttachmentEditState; conflicts: MediaDraftConflict[] };

const imageFields = new Set(["title", "source", "maintenanceId"]);
const attachmentFields = (value: AttachmentEditState[string]) => ({
  description: value.description, category: value.category, maintenanceId: value.maintenanceId
});

export function refreshMediaDraft(
  currentImages: PendingArticleImage[], state: MediaMetadataDraft, previous: Vehicle, incoming: Vehicle
) {
  const beforeImages = vehicleImagesToPending(previous), savedImages = vehicleImagesToPending(incoming);
  const images = reconcileImageDraft(currentImages, previous, incoming);
  const conflicts: MediaDraftConflict[] = [];
  for (const before of beforeImages) {
    const local = currentImages.find(image => image.id === before.id);
    if (!local) continue;
    const saved = savedImages.find(image => image.id === before.id);
    const retained = state.conflicts.find(conflict => conflict.kind === "image" && conflict.id === before.id);
    if (!saved) {
      if (!sameDraftValue(local, before) || retained) conflicts.push({
        kind: "image", id: before.id, label: local.title || local.fileName || "", deleted: true,
        fields: [...imageFields].filter(field => !sameDraftValue(draftValue(local, field), draftValue(before, field)))
          .map(field => ({ field, local: draftValue(local, field), saved: undefined }))
      });
      continue;
    }
    const fields = retainRefreshConflicts(retained?.fields || [], local, saved,
      reconcileDraft(local, before, saved).conflicts).filter(field => imageFields.has(field.field));
    if (fields.length) conflicts.push({ kind: "image", id: before.id, label: local.title || local.fileName || "", deleted: false, fields });
  }
  // Retain edits to records already absent from the preceding refresh as well.
  for (const conflict of state.conflicts.filter(value => value.kind === "image" && value.deleted)) {
    if (!conflicts.some(value => value.id === conflict.id) && currentImages.some(image => image.id === conflict.id) &&
      !savedImages.some(image => image.id === conflict.id)) conflicts.push(conflict);
  }
  const primary = (items: PendingArticleImage[]) => items.find(image => image.isPrimary)?.id;
  const common = new Set(beforeImages.filter(image => currentImages.some(value => value.id === image.id) &&
    savedImages.some(value => value.id === image.id)).map(image => image.id));
  const order = (items: PendingArticleImage[]) => items.filter(image => common.has(image.id)).map(image => image.id);
  const groups = [
    { id: "@primary", field: "isPrimary", before: primary(beforeImages), local: primary(currentImages), saved: primary(savedImages) },
    { id: "@order", field: "order", before: order(beforeImages), local: order(currentImages), saved: order(savedImages) }
  ];
  for (const group of groups) {
    const retained = state.conflicts.some(value => value.kind === "image" && value.id === group.id);
    if (!sameDraftValue(group.local, group.saved) && (retained ||
      (!sameDraftValue(group.local, group.before) && !sameDraftValue(group.saved, group.before)))) {
      const display = (value: unknown, items: PendingArticleImage[]) => (Array.isArray(value) ? value : [value])
        .map(id => items.find(image => image.id === id)?.title || `#${items.findIndex(image => image.id === id) + 1}`).join(", ");
      conflicts.push({ kind: "image", id: group.id, label: "", deleted: false,
        fields: [{ field: group.field, local: display(group.local, currentImages), saved: display(group.saved, savedImages) }] });
    }
  }
  const beforeEdits = attachmentsToEditState(previous.attachments), savedEdits = attachmentsToEditState(incoming.attachments);
  const edits = { ...state.edits };
  for (const [id, local] of Object.entries(state.edits)) {
    const before = beforeEdits[id], saved = savedEdits[id];
    const retained = state.conflicts.find(conflict => conflict.kind === "attachment" && conflict.id === id);
    const label = incoming.attachments?.find(item => item.id === id)?.originalName ||
      previous.attachments?.find(item => item.id === id)?.originalName || retained?.label || "";
    if (!saved) {
      if ((before && !sameDraftValue(attachmentFields(local), attachmentFields(before))) || retained) {
        conflicts.push({ kind: "attachment", id, label, deleted: true,
          fields: Object.entries(attachmentFields(local)).map(([field, value]) => ({ field, local: value, saved: undefined })) });
      } else delete edits[id];
      continue;
    }
    const merged = reconcileDraft(attachmentFields(local), attachmentFields(before || local), attachmentFields(saved));
    const fields = retainRefreshConflicts(retained?.fields || [], local, saved, merged.conflicts);
    edits[id] = { ...merged.value, expectedRevision: saved.expectedRevision };
    if (fields.length) conflicts.push({ kind: "attachment", id, label, deleted: false, fields });
  }
  for (const [id, saved] of Object.entries(savedEdits)) if (!edits[id]) edits[id] = saved;
  return { images, state: { edits, conflicts } };
}

export function resolveMediaImage(
  images: PendingArticleImage[], conflict: MediaDraftConflict, incoming: Vehicle, choice: "draft" | "saved"
) {
  if (choice === "draft" && !conflict.deleted) return images;
  const saved = vehicleImagesToPending(incoming);
  if (conflict.id === "@primary") {
    const id = saved.find(image => image.isPrimary)?.id;
    return images.map(image => ({ ...image, isPrimary: image.id === id }));
  }
  if (conflict.id === "@order") {
    const ids = saved.map(image => image.id);
    return [...images.filter(image => ids.includes(image.id)).sort((a, b) => ids.indexOf(a.id) - ids.indexOf(b.id)),
      ...images.filter(image => !ids.includes(image.id))];
  }
  const record = saved.find(image => image.id === conflict.id);
  if (!record) return images.filter(image => image.id !== conflict.id);
  return images.map(image => image.id !== record.id ? image : {
    ...image, ...Object.fromEntries(conflict.fields.map(field => [field.field, draftValue(record, field.field)]))
  });
}

// Display and guard against the live draft, including typing after a conflict response.
export function activeMediaConflicts(
  state: MediaMetadataDraft, images: PendingArticleImage[], incoming: Vehicle | null
): MediaDraftConflict[] {
  if (!incoming) return [];
  const savedImages = vehicleImagesToPending(incoming), savedEdits = attachmentsToEditState(incoming.attachments);
  const label = (items: PendingArticleImage[], id: string | undefined) => {
    const index = items.findIndex(image => image.id === id), image = items[index];
    return image ? image.title || image.fileName || `#${index + 1}` : "";
  };
  return state.conflicts.flatMap(conflict => {
    if (conflict.kind === "image" && conflict.id === "@primary") {
      const local = images.find(image => image.isPrimary)?.id, saved = savedImages.find(image => image.isPrimary)?.id;
      return local === saved ? [] : [{ ...conflict, fields: [{ field: "isPrimary", local: label(images, local), saved: label(savedImages, saved) }] }];
    }
    if (conflict.kind === "image" && conflict.id === "@order") {
      const common = new Set(images.filter(image => savedImages.some(saved => saved.id === image.id)).map(image => image.id));
      const local = images.filter(image => common.has(image.id)).map(image => image.id);
      const saved = savedImages.filter(image => common.has(image.id)).map(image => image.id);
      return sameDraftValue(local, saved) ? [] : [{ ...conflict,
        fields: [{ field: "order", local: local.map(id => label(images, id)).join(", "), saved: saved.map(id => label(savedImages, id)).join(", ") }] }];
    }
    const local = conflict.kind === "image" ? images.find(image => image.id === conflict.id) : state.edits[conflict.id];
    const saved = conflict.kind === "image" ? savedImages.find(image => image.id === conflict.id) : savedEdits[conflict.id];
    if (!local) return [];
    const fields = conflict.fields.map(field => ({ field: field.field,
      local: draftValue(local, field.field), saved: saved ? draftValue(saved, field.field) : undefined
    })).filter(field => conflict.deleted || !sameDraftValue(field.local, field.saved));
    return conflict.deleted || fields.length ? [{ ...conflict, fields }] : [];
  });
}
