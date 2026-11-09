import type {
  ExhibitionEntry, ExhibitionEntryInput, ExhibitionList, ExhibitionListInput, ExhibitionStatusInput,
  ExhibitionWorkspace
} from "./apiExhibitionModels";
import type { APIRequest } from "./apiTransport";

export function createExhibitionAPI(request: APIRequest) {
  return {
    exhibitionLists: () => request<ExhibitionList[]>("/exhibition-lists"),
    exhibitionList: (id: string) => request<ExhibitionList>(`/exhibition-lists/${encodeURIComponent(id)}`),
    exhibitionWorkspace: (id: string) =>
      request<ExhibitionWorkspace>(`/exhibition-lists/${encodeURIComponent(id)}/workspace`),
    checkExhibitionConflicts: (id: string) =>
      request<ExhibitionWorkspace>(`/exhibition-lists/${encodeURIComponent(id)}/conflicts/check`, { method: "POST" }),
    createExhibitionList: (input: ExhibitionListInput) =>
      request<ExhibitionList>("/exhibition-lists", {
        method: "POST",
        body: JSON.stringify(input)
      }),
    updateExhibitionList: (id: string, input: ExhibitionListInput) =>
      request<ExhibitionList>(`/exhibition-lists/${encodeURIComponent(id)}`, {
        method: "PUT",
        body: JSON.stringify(input)
      }),
    deleteExhibitionList: (id: string) =>
      request<void>(`/exhibition-lists/${encodeURIComponent(id)}`, {
        method: "DELETE"
      }),
    setExhibitionListLocked: (id: string, locked: boolean) =>
      request<ExhibitionList>(`/exhibition-lists/${encodeURIComponent(id)}/lock`, {
        method: "PUT",
        body: JSON.stringify({ locked })
      }),
    setExhibitionStatus: (id: string, input: ExhibitionStatusInput) =>
      request<ExhibitionList>(`/exhibition-lists/${encodeURIComponent(id)}/status`, {
        method: "PUT",
        body: JSON.stringify(input)
      }),
    setExhibitionConflictException: (listId: string, conflictId: string, reason: string, expectedRevision: number) =>
      request<ExhibitionWorkspace>(
        `/exhibition-lists/${encodeURIComponent(listId)}/conflicts/${encodeURIComponent(conflictId)}/exception`,
        { method: "PUT", body: JSON.stringify({ reason, expectedRevision }) }
      ),
    exhibitionEntries: (listId: string) => request<ExhibitionEntry[]>(`/exhibition-lists/${encodeURIComponent(listId)}/entries`),
    createExhibitionEntry: (listId: string, input: ExhibitionEntryInput) =>
      request<ExhibitionEntry>(`/exhibition-lists/${encodeURIComponent(listId)}/entries`, {
        method: "POST",
        body: JSON.stringify(input)
      }),
    updateExhibitionEntry: (listId: string, entryId: string, input: ExhibitionEntryInput) =>
      request<ExhibitionEntry>(`/exhibition-lists/${encodeURIComponent(listId)}/entries/${encodeURIComponent(entryId)}`, {
        method: "PUT",
        body: JSON.stringify(input)
      }),
    deleteExhibitionEntry: (listId: string, entryId: string) =>
      request<void>(`/exhibition-lists/${encodeURIComponent(listId)}/entries/${encodeURIComponent(entryId)}`, {
        method: "DELETE"
      })
  };
}
