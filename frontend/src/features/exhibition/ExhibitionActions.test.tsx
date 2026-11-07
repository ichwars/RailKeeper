import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { ApiError, api, type ExhibitionList, type ExhibitionWorkspace } from "../../shared/api";
import { vehicleFixture } from "../../test/fixtures/vehicles";
import { ExhibitionView } from "./ExhibitionView";

const list: ExhibitionList = {
  id: "event", designation: "Clubausstellung", date: "2026-11-07", endDate: "2026-11-08",
  status: "open", revision: 4, locked: false, entryCount: 1,
  createdAt: "2026-10-01T10:00:00Z", updatedAt: "2026-10-01T10:00:00Z"
};
function reference(): ExhibitionWorkspace {
  return {
    list: { ...list }, dayScopes: ["day1", "day2"],
    summary: { entryCount: 1, ownerCount: 1, conflictCount: 1, readyCount: 0 },
    readiness: { total: 1, addressesChecked: 0, functionsDocumented: 1, imagesPresent: 0, problems: 1 },
    entries: [{
      id: "entry", listId: list.id, owner: "Alex", locomotiveName: "BR 103", dayScope: "day2", dtDecoder: true,
      decoderNumber: "103", interfaceName: "ECoS", analog: false, availability: "available", revision: 2,
      sortOrder: 10, createdAt: list.createdAt, updatedAt: list.updatedAt, status: "addressConflict", conflictIds: ["conflict"]
    }],
    conflicts: [{ id: "conflict", kind: "address", entryIds: ["entry"], interfaceName: "ECoS", address: "103", dayScopes: ["day2"], excepted: false }]
  };
}

async function open() {
  render(<ExhibitionView roles={["Admin"]} />);
  await screen.findByText("BR 103");
  return userEvent.setup();
}

describe("Exhibition confirmed actions", () => {
  beforeEach(() => {
    vi.spyOn(api, "exhibitionLists").mockResolvedValue([{ ...list }]);
    vi.spyOn(api, "exhibitionWorkspace").mockResolvedValue(reference());
    vi.spyOn(api, "masterData").mockResolvedValue([]);
    vi.spyOn(api, "masterDataAll").mockResolvedValue({});
    vi.spyOn(api, "vehicles").mockResolvedValue([vehicleFixture()]);
  });

  it("preserves event edits after a conflict and submits the original revision on explicit retry", async () => {
    const update = vi.spyOn(api, "updateExhibitionList").mockRejectedValueOnce(new ApiError("Veranstaltung wurde geändert", "conflict", 409)).mockResolvedValue(list);
    const user = await open();
    await user.click(screen.getByRole("button", { name: "Weitere Aktionen" }));
    await user.click(screen.getByRole("button", { name: "Veranstaltung bearbeiten" }));
    const dialog = within(screen.getByRole("dialog", { name: "Veranstaltung bearbeiten" }));
    await user.clear(dialog.getByLabelText("Bezeichnung"));
    await user.type(dialog.getByLabelText("Bezeichnung"), "Geänderter Clubabend");
    await user.type(dialog.getByLabelText("Ort"), "Vereinsheim");
    await user.type(dialog.getByLabelText("Beschreibung"), "Gastfahrzeuge");
    await user.type(dialog.getByLabelText("Organisation (intern)"), "Aufbau ab 9 Uhr");
    await user.click(dialog.getByRole("button", { name: "Speichern" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("Veranstaltung wurde geändert");
    expect(dialog.getByLabelText("Bezeichnung")).toHaveValue("Geänderter Clubabend");
    expect(update).toHaveBeenCalledOnce();
    await user.click(dialog.getByRole("button", { name: "Speichern" }));
    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
    expect(update).toHaveBeenLastCalledWith("event", expect.objectContaining({
      expectedRevision: 4, designation: "Geänderter Clubabend", location: "Vereinsheim",
      description: "Gastfahrzeuge", organizationNotes: "Aufbau ab 9 Uhr"
    }));
  });

  it("creates an event only on save and keeps cancellation free of writes", async () => {
    const create = vi.spyOn(api, "createExhibitionList").mockResolvedValue({ ...list, id: "new-event" });
    const user = await open();
    await user.click(screen.getByRole("button", { name: "Neue Veranstaltung" }));
    await user.type(screen.getByLabelText("Bezeichnung"), "Entwurf");
    await user.click(screen.getByRole("button", { name: "Abbrechen" }));
    expect(create).not.toHaveBeenCalled();
    await user.click(screen.getByRole("button", { name: "Neue Veranstaltung" }));
    expect(screen.getByLabelText("Bezeichnung")).toHaveValue("");
    await user.type(screen.getByLabelText("Bezeichnung"), "Neue Ausstellung");
    await user.click(screen.getByRole("button", { name: "Speichern" }));
    await waitFor(() => expect(create).toHaveBeenCalledOnce());
    expect(create).toHaveBeenCalledWith(expect.objectContaining({ designation: "Neue Ausstellung", status: "open" }));
  });

  it("retains rejected entry edits and sends the selected event and entry revision", async () => {
    const update = vi.spyOn(api, "updateExhibitionEntry").mockRejectedValueOnce(new ApiError("Eintrag wurde geändert", "conflict", 409)).mockResolvedValue(reference().entries[0]);
    const user = await open();
    await user.click(screen.getByRole("button", { name: "BR 103 bearbeiten" }));
    const dialog = within(screen.getByRole("dialog", { name: "Eintrag bearbeiten" }));
    await user.clear(dialog.getByLabelText("Besitzer"));
    await user.type(dialog.getByLabelText("Besitzer"), "Alex und Chris");
    await user.click(dialog.getByRole("button", { name: "Speichern" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("Eintrag wurde geändert");
    expect(dialog.getByLabelText("Besitzer")).toHaveValue("Alex und Chris");
    expect(update).toHaveBeenCalledWith("event", "entry", expect.objectContaining({ owner: "Alex und Chris", expectedRevision: 2 }));
    await user.click(dialog.getByRole("button", { name: "Speichern" }));
    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
  });

  it("requires a reason and a separate confirmation before locking with conflicts", async () => {
    const status = vi.spyOn(api, "setExhibitionStatus").mockRejectedValueOnce(new ApiError("Konflikte bestehen", "exhibition_conflicts", 409)).mockResolvedValue({ ...list, status: "locked", locked: true, revision: 5 });
    const user = await open();
    await user.click(screen.getByRole("button", { name: "Liste sperren" }));
    const dialog = within(await screen.findByRole("dialog", { name: "Liste trotz Konflikten sperren?" }));
    expect(status).toHaveBeenCalledWith("event", { status: "locked", expectedRevision: 4, reason: "", confirmConflicts: false });
    expect(dialog.getByRole("button", { name: "Trotzdem sperren" })).toBeDisabled();
    await user.type(dialog.getByLabelText("Begründung"), "Getrennte Fahrzeiten vereinbart");
    await user.click(dialog.getByRole("button", { name: "Trotzdem sperren" }));
    await waitFor(() => expect(status).toHaveBeenLastCalledWith("event", { status: "locked", expectedRevision: 4, reason: "Getrennte Fahrzeiten vereinbart", confirmConflicts: true }));
  });

  it("records a conflict exception without claiming the conflict is solved", async () => {
    const workspace = reference();
    const updated = { ...workspace, conflicts: [{ ...workspace.conflicts[0], excepted: true, exceptionReason: "Getrennte Gleise" }] };
    const exception = vi.spyOn(api, "setExhibitionConflictException").mockResolvedValue(updated);
    const user = await open();
    await user.click(screen.getByRole("button", { name: "Konflikte prüfen" }));
    const dialog = within(screen.getByRole("dialog", { name: "Konflikte und Prüfungen" }));
    await user.click(dialog.getByRole("button", { name: "Ausnahme speichern" }));
    expect(exception).not.toHaveBeenCalled();
    await user.type(dialog.getByPlaceholderText("Begründung der Ausnahme"), "  Getrennte Gleise  ");
    await user.click(dialog.getByRole("button", { name: "Ausnahme speichern" }));
    expect(await dialog.findByText(/Dokumentierte Ausnahme.*Getrennte Gleise/)).toBeInTheDocument();
    expect(exception).toHaveBeenCalledWith("event", "conflict", "Getrennte Gleise", 4);
    expect(dialog.getByText(/gelten nicht als vollständig gelöst/)).toBeInTheDocument();
    await user.click(dialog.getAllByRole("button", { name: "Schließen" })[0]);
    expect(screen.getByText("Adresskonflikt")).toBeInTheDocument();
  });

  it("keeps day/search filters reversible and hides archived events until requested", async () => {
    const archived = { ...list, id: "archive", designation: "Vorjahr", status: "archived" as const };
    vi.mocked(api.exhibitionLists).mockResolvedValue([list, archived]);
    const user = await open();
    expect(screen.queryByText("Vorjahr")).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: /^Tag 1/ }));
    expect(screen.queryByText("BR 103")).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Alle Tage" }));
    await user.type(screen.getByRole("textbox", { name: "Einträge durchsuchen" }), "not-present");
    expect(screen.getByText("Keine Einträge für diese Auswahl.")).toBeInTheDocument();
    await user.clear(screen.getByRole("textbox", { name: "Einträge durchsuchen" }));
    expect(screen.getByText("BR 103")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Alle Veranstaltungen anzeigen" }));
    expect(screen.getByText("Vorjahr")).toBeInTheDocument();
  });

  it("deletes an entry only after the explicit confirmation", async () => {
    const remove = vi.spyOn(api, "deleteExhibitionEntry").mockResolvedValue(undefined);
    const confirm = vi.spyOn(window, "confirm").mockReturnValueOnce(false).mockReturnValueOnce(true);
    const user = await open();
    const row = within(screen.getByRole("row", { name: /BR 103/ }));
    await user.click(row.getByRole("button", { name: "Löschen" }));
    expect(remove).not.toHaveBeenCalled();
    await user.click(row.getByRole("button", { name: "Löschen" }));
    await waitFor(() => expect(remove).toHaveBeenCalledWith("event", "entry"));
    expect(confirm).toHaveBeenCalledTimes(2);
  });

  it("displays a load error and allows dismissing it without any write", async () => {
    vi.mocked(api.exhibitionLists).mockRejectedValue(new Error("Listen nicht erreichbar"));
    render(<ExhibitionView roles={["Messe"]} />);
    expect(await screen.findByRole("alert")).toHaveTextContent("Listen nicht erreichbar");
    await userEvent.setup().click(screen.getByRole("button", { name: "Meldung schließen" }));
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Neue Veranstaltung" })).not.toBeInTheDocument();
  });
});
