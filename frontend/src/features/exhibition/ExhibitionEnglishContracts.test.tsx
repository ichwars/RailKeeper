import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { api, type ExhibitionList, type ExhibitionWorkspace } from "../../shared/api";
import { setLanguage } from "../../shared/i18n";
import { vehicleFixture } from "../../test/fixtures/vehicles";
import { ExhibitionView } from "./ExhibitionView";

const at = "2026-11-08T12:00:00Z";
const list: ExhibitionList = { id: "english-event", designation: "English event", date: "2026-11-10", endDate: "2026-11-11", status: "open", revision: 7, locked: false, entryCount: 1, createdAt: at, updatedAt: at };
const source = vehicleFixture({ id: "mapped-source", name: "Mapped source", manufacturer: "Stored maker", series: "218", dtDecoder: true, dtDecoderNumber: "31", digitalDecoderNumber: "32", decoderType: "Stored decoder" });
function reference(current = list): ExhibitionWorkspace {
  return { list: current, dayScopes: ["day1", "day2"], summary: { entryCount: 1, ownerCount: 1, conflictCount: 2, readyCount: 0 },
    readiness: { total: 1, addressesChecked: 0, functionsDocumented: 0, imagesPresent: 0, problems: 1 },
    entries: [{ id: "guest", listId: current.id, owner: "Guest owner", locomotiveName: "Guest locomotive", dayScope: "day2", analog: false, dtDecoder: true, availability: "unavailable", revision: 2, sortOrder: 10, createdAt: at, updatedAt: at, status: "unavailable", conflictIds: ["missing", "duplicate"] }],
    conflicts: [{ id: "missing", kind: "missing", entryIds: ["guest"], fields: ["decoderNumber"], excepted: false }, { id: "duplicate", kind: "duplicateVehicle", entryIds: ["guest"], excepted: false }]
  };
}

describe("English exhibition workflows and revision contracts", () => {
  beforeEach(() => {
    setLanguage("en");
    vi.spyOn(api,"exhibitionLists").mockResolvedValue([list]);
    vi.spyOn(api,"exhibitionWorkspace").mockResolvedValue(reference());
    vi.spyOn(api,"vehicles").mockResolvedValue([source]);
    vi.spyOn(api,"masterData").mockResolvedValue([]);
    vi.spyOn(api,"masterDataAll").mockResolvedValue({});
  });

  it("keeps day, readiness, search and archive filters independent and exposes the actual conflicts", async () => {
    const archived = { ...list, id: "archived", designation: "Old archived event", date: "2025-01-01", endDate: "2025-01-02", status: "archived" as const };
    vi.mocked(api.exhibitionLists).mockResolvedValue([list, archived]);
    const user = userEvent.setup(); render(<ExhibitionView roles={["Admin"]} />);
    await screen.findByText("Guest locomotive");
    expect(screen.queryByText("Old archived event")).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: /^Day 1/ }));
    expect(screen.queryByText("Guest locomotive")).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "All days" }));
    await user.type(screen.getByRole("textbox", { name: "Search entries" }), "Guest owner");
    expect(screen.getByText("Guest locomotive")).toBeInTheDocument();
    await user.clear(screen.getByRole("textbox", { name: "Search entries" }));
    await user.click(screen.getByRole("button", { name: /^Ready 0$/ }));
    expect(screen.queryByText("Guest locomotive")).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: /^All1$/ }));
    for (const label of [/Addresses checked/, /Function keys/, /^Images/]) {
      await user.click(screen.getByRole("button", { name: label }));
      expect(screen.getByText("Guest locomotive")).toBeInTheDocument();
    }
    await user.click(screen.getByRole("button", { name: "More actions" }));
    await user.click(screen.getByRole("button", { name: "Show archive" }));
    expect(screen.getByText("Old archived event")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Hide archive" }));
    expect(screen.queryByText("Old archived event")).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Show all events" }));
    expect(screen.getByText("Old archived event")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Check conflicts" }));
    const dialog = screen.getByRole("dialog", { name: "Conflicts and checks" });
    expect(within(dialog).getByText("Vehicle is entered more than once.")).toBeInTheDocument();
    expect(within(dialog).getByText(/decoderNumber/)).toBeInTheDocument();
    await user.click(within(dialog).getAllByRole("button", { name: "Close" })[0]);
  });

  it("copies a selected inventory reference only into the event and can return to guest input", async () => {
    const save = vi.spyOn(api,"createExhibitionEntry").mockResolvedValue(reference().entries[0]);
    const user = userEvent.setup(); render(<ExhibitionView roles={["Admin"]} />);
    await screen.findByText("Guest locomotive");
    await user.click(screen.getByRole("button", { name: "Add entry" }));
    const dialog = screen.getByRole("dialog", { name: "Add entry" });
    const vehicle = within(dialog).getByRole("button", { name: "Inventory vehicle (optional)" });
    await user.click(vehicle); await user.click(screen.getByRole("option", { name: /Mapped source/ }));
    expect(within(dialog).getByLabelText("Locomotive name")).toHaveValue(source.name);
    expect(within(dialog).getByLabelText("DCC address")).toHaveValue("31");
    await user.click(vehicle); await user.click(screen.getByRole("option", { name: "Guest vehicle / manual entry" }));
    expect(within(dialog).getByLabelText("Locomotive name")).toHaveValue(source.name);
    await user.click(within(dialog).getByRole("button", { name: "Day 2" }));
    fireEvent.submit(dialog);
    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
    expect(save).toHaveBeenCalledWith(list.id, expect.objectContaining({ vehicleId: "", locomotiveName: source.name, manufacturer: source.manufacturer, decoderNumber: "31", decoderType: source.decoderType, dayScope: "day2" }));
  });

  it.each([
    { from: "draft", action: "Open planning", to: "open" },
    { from: "running", action: "Complete event", to: "completed" },
    { from: "completed", action: "Reopen planning", to: "open" },
    { from: "archived", action: "Restore from archive", to: "open" },
    { from: "open", action: "Archive", to: "archived" }
  ] as const)("submits the current revision for $from → $to", async item => {
    const current = { ...list, status: item.from, locked: item.from === "running" };
    vi.mocked(api.exhibitionLists).mockResolvedValue([current]);
    vi.mocked(api.exhibitionWorkspace).mockResolvedValue(reference(current));
    const change = vi.spyOn(api,"setExhibitionStatus").mockResolvedValue({ ...current, status: item.to });
    const user = userEvent.setup(); render(<ExhibitionView roles={["Admin"]} />);
    await screen.findByText("Guest locomotive");
    await user.click(screen.getByRole("button", { name: "More actions" }));
    await user.click(screen.getByRole("button", { name: item.action }));
    await waitFor(() => expect(change).toHaveBeenCalledWith(list.id, { status: item.to, expectedRevision: 7, reason: "", confirmConflicts: false }));
  });
});
