import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import { api, ExhibitionList, ExhibitionWorkspace, MasterDataEntry } from "../../shared/api";
import { ExhibitionView } from "./ExhibitionView";
import { printFunctionChips } from "./exhibitionPrint";
import * as exhibitionPrint from "./exhibitionPrint";

const list: ExhibitionList = {
  id: "messe-koeln",
  designation: "Modellbahntage Köln",
  date: "2026-08-22",
  endDate: "2026-08-24",
  status: "open",
  revision: 4,
  locked: false,
  entryCount: 2,
  createdAt: "2026-08-01T10:00:00Z",
  updatedAt: "2026-08-01T10:00:00Z"
};

const workspace: ExhibitionWorkspace = {
  list,
  summary: { entryCount: 2, ownerCount: 2, conflictCount: 1, readyCount: 1 },
  readiness: { total: 2, addressesChecked: 2, functionsDocumented: 1, imagesPresent: 1, problems: 1 },
  dayScopes: ["day1", "day2", "day3"],
  conflicts: [{
    id: "conflict-1",
    kind: "address",
    entryIds: ["entry-1", "entry-2"],
    interfaceName: "ECoS",
    address: "103",
    dayScopes: ["day2"],
    excepted: false
  }],
  entries: [
    {
      id: "entry-1", listId: list.id, owner: "Michael Weber", locomotiveName: "BR 103 113-7",
      dayScope: "day1,day2,day3", dtDecoder: true, decoderNumber: "103", interfaceName: "ECoS",
      analog: false, availability: "available", revision: 1, functionKeys: "F0", imageUrl: "/train.png",
      sortOrder: 10, createdAt: list.createdAt, updatedAt: list.updatedAt, status: "ready", conflictIds: []
    },
    {
      id: "entry-2", listId: list.id, owner: "Thomas Neumann", locomotiveName: "V 200 033",
      dayScope: "day2,day3", dtDecoder: true, decoderNumber: "103", interfaceName: "ECoS",
      analog: false, availability: "available", revision: 1, sortOrder: 20,
      createdAt: list.createdAt, updatedAt: list.updatedAt, status: "addressConflict", conflictIds: ["conflict-1"]
    }
  ]
};

const lightSymbol: MasterDataEntry = {
  id: "symbols:light",
  type: "symbols",
  key: "light",
  label: "Licht",
  active: true,
  sortOrder: 10,
  metadata: { imageData: "data:image/png;base64,cHJpbnQ=", activeImageData: "data:image/png;base64,YWN0aXZl" },
  createdAt: list.createdAt,
  updatedAt: list.updatedAt,
};

describe("Exhibition reference workspace", () => {
  afterEach(() => vi.restoreAllMocks());

  it("renders the reference topology from the persistent workspace", async () => {
    vi.spyOn(api, "exhibitionLists").mockResolvedValue([list]);
    vi.spyOn(api, "exhibitionWorkspace").mockResolvedValue(workspace);
    vi.spyOn(api, "masterData").mockResolvedValue([]);
    vi.spyOn(api, "masterDataAll").mockResolvedValue({});

    render(<ExhibitionView roles={["Admin"]} />);

    expect(await screen.findByRole("heading", { name: "Ausstellung", level: 1 })).toBeInTheDocument();
    expect(screen.getAllByRole("button", { name: /Neue Veranstaltung/i })[0]).toBeInTheDocument();
    expect(await screen.findByRole("region", { name: "Veranstaltungsübersicht" })).toBeInTheDocument();
    expect(screen.getByRole("region", { name: "Veranstaltungen" })).toBeInTheDocument();
    const operations = screen.getByRole("region", { name: "Teilnehmer und Fahrzeuge" });
    expect(within(operations).getByText("BR 103 113-7")).toBeInTheDocument();
    expect(within(operations).getByRole("button", { name: /Tag 3/i })).toBeInTheDocument();
    expect(screen.getByRole("region", { name: "Bereitschaft" })).toBeInTheDocument();
    await waitFor(() => expect(api.exhibitionWorkspace).toHaveBeenCalledWith(list.id));
  });

  it("filters persistent entries by day and readiness", async () => {
    const user = userEvent.setup();
    vi.spyOn(api, "exhibitionLists").mockResolvedValue([list]);
    vi.spyOn(api, "exhibitionWorkspace").mockResolvedValue(workspace);
    vi.spyOn(api, "masterData").mockResolvedValue([]);
    vi.spyOn(api, "masterDataAll").mockResolvedValue({});

    render(<ExhibitionView roles={["Admin"]} />);
    await screen.findByText("BR 103 113-7");
    await user.click(screen.getByRole("button", { name: /^Bereit 1$/i }));
    expect(screen.getByText("BR 103 113-7")).toBeInTheDocument();
    expect(screen.queryByText("V 200 033")).not.toBeInTheDocument();
  });

  it("uses the print palette for function symbols in generated print markup", () => {
    const html = printFunctionChips(
      JSON.stringify([{ key: "F0", name: "Fahrlicht", type: "licht", symbolKey: "light" }]),
      [lightSymbol],
    );

    expect(html).toContain('src="data:image/png;base64,cHJpbnQ="');
    expect(html).not.toContain('src="data:image/png;base64,YWN0aXZl"');
  });

  it("prints all saved entries in the selected event despite screen filters", async () => {
    const user = userEvent.setup();
    vi.spyOn(api, "exhibitionLists").mockResolvedValue([list]);
    vi.spyOn(api, "exhibitionWorkspace").mockResolvedValue(workspace);
    vi.spyOn(api, "masterData").mockResolvedValue([lightSymbol]);
    const print = vi.spyOn(exhibitionPrint, "printList").mockResolvedValue();
    const pagePrint = vi.spyOn(window, "print").mockImplementation(() => {});
    render(<ExhibitionView roles={["Messe"]} />);
    await screen.findByText("BR 103 113-7");
    await user.click(screen.getByRole("button", { name: /^Tag 1/ }));
    await user.type(screen.getByRole("textbox", { name: "Einträge durchsuchen" }), "BR 103");
    expect(screen.queryByText("V 200 033")).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Drucken" }));
    expect(print).toHaveBeenCalledWith(list, workspace.entries, [lightSymbol], "de", expect.any(Function));
    expect(pagePrint).not.toHaveBeenCalled();
  });

  it("blocks printing the previous workspace while another event loads and reports failures", async () => {
    const user = userEvent.setup();
    const nextList = { ...list, id: "next-event", designation: "Nächste Ausstellung" };
    let finishLoad: (value: ExhibitionWorkspace) => void = () => {};
    vi.spyOn(api, "exhibitionLists").mockResolvedValue([list, nextList]);
    vi.spyOn(api, "exhibitionWorkspace").mockResolvedValueOnce(workspace)
      .mockReturnValueOnce(new Promise((resolve) => { finishLoad = resolve; }));
    vi.spyOn(api, "masterData").mockResolvedValue([]);
    const print = vi.spyOn(exhibitionPrint, "printList").mockRejectedValue(new Error("Druck nicht verfügbar"));
    render(<ExhibitionView roles={["Messe"]} />);
    await screen.findByText("BR 103 113-7");
    await user.click(screen.getByRole("button", { name: /Nächste Ausstellung/ }));
    expect(screen.getByRole("button", { name: "Drucken" })).toBeDisabled();
    finishLoad({ ...workspace, list: nextList, entries: [] });
    await waitFor(() => expect(screen.getByRole("button", { name: "Drucken" })).toBeEnabled());
    await user.click(screen.getByRole("button", { name: "Drucken" }));
    expect(print).toHaveBeenCalledWith(nextList, [], [], "de", expect.any(Function));
    expect(await screen.findByRole("alert")).toHaveTextContent("Druck nicht verfügbar");
  });

  it("persists lifecycle transitions from the event menu", async () => {
	const user = userEvent.setup();
	const lockedList: ExhibitionList = { ...list, status: "locked", locked: true, revision: 5 };
	const lockedWorkspace: ExhibitionWorkspace = { ...workspace, list: lockedList };
	vi.spyOn(api, "exhibitionLists").mockResolvedValue([lockedList]);
	vi.spyOn(api, "exhibitionWorkspace").mockResolvedValue(lockedWorkspace);
	vi.spyOn(api, "masterData").mockResolvedValue([]);
	vi.spyOn(api, "masterDataAll").mockResolvedValue({});
	const setStatus = vi.spyOn(api, "setExhibitionStatus").mockResolvedValue({
	  ...lockedList, status: "running", locked: true, revision: 6
	});

	render(<ExhibitionView roles={["Admin"]} />);
	await screen.findByText("BR 103 113-7");
	await user.click(screen.getByRole("button", { name: "Weitere Aktionen" }));
	await user.click(screen.getByRole("button", { name: "Veranstaltung starten" }));

	await waitFor(() => expect(setStatus).toHaveBeenCalledWith(lockedList.id, {
	  status: "running", expectedRevision: 5, reason: "", confirmConflicts: false
	}));
  });

  it("opens create and the single row action directly as modal dialogs", async () => {
    const user = userEvent.setup();
    vi.spyOn(api, "exhibitionLists").mockResolvedValue([list]);
    vi.spyOn(api, "exhibitionWorkspace").mockResolvedValue(workspace);
    vi.spyOn(api, "masterData").mockResolvedValue([]);
    vi.spyOn(api, "masterDataAll").mockResolvedValue({});

    render(<ExhibitionView roles={["Admin"]} />);
    await screen.findByText("BR 103 113-7");

    await user.click(screen.getAllByRole("button", { name: "Neue Veranstaltung" })[0]);
    const createDialog = screen.getByRole("dialog", { name: "Neue Veranstaltung" });
    expect(createDialog).toBeInTheDocument();
    expect(createDialog.parentElement).toHaveClass("modal-layer");
    expect(createDialog.querySelector('input[type="date"]')).not.toBeInTheDocument();
    await user.keyboard("{Escape}");

    const editButton = screen.getByRole("button", { name: "BR 103 113-7 bearbeiten" });
    expect(editButton).toHaveAttribute("title", "BR 103 113-7 bearbeiten");
    await user.click(editButton);
    const editDialog = screen.getByRole("dialog", { name: "Eintrag bearbeiten" });
    expect(editDialog).toBeInTheDocument();
    expect(editDialog.parentElement).toHaveClass("modal-layer");
    expect(within(editDialog).getByRole("tab", { name: "Allgemein" })).toHaveAttribute("aria-selected", "true");
    expect(within(editDialog).getByText("Hersteller")).toBeInTheDocument();
    expect(within(editDialog).getByText("Baureihe")).toBeInTheDocument();
    expect(within(editDialog).getByText("Gattung")).toBeInTheDocument();
    expect(within(editDialog).getByText("Epoche")).toBeInTheDocument();
    expect(within(editDialog).getByText("Bahnverwaltung")).toBeInTheDocument();
    expect(within(editDialog).getByText("Decoder-Typ")).toBeInTheDocument();
    expect(within(editDialog).getByText("Adresse DCC")).toBeInTheDocument();
    expect(within(editDialog).getByText("Adresse SX")).toBeInTheDocument();
    expect(within(editDialog).getByText("Analog")).toBeInTheDocument();
    expect(editDialog.querySelector("select")).not.toBeInTheDocument();

    await user.click(within(editDialog).getByRole("tab", { name: /Fahrzeugbild/ }));
    expect(within(editDialog).getByRole("button", { name: /Bild auswählen/ })).toBeInTheDocument();
    expect(within(editDialog).queryByText("Bild-URL")).not.toBeInTheDocument();

    await user.click(within(editDialog).getByRole("tab", { name: /Funktionstasten/ }));
    expect(within(editDialog).getByDisplayValue("Fahrlicht")).toBeInTheDocument();
    expect(within(editDialog).getByRole("button", { name: "Funktion hinzufügen" })).toBeInTheDocument();
    expect(within(editDialog).queryByPlaceholderText("F0, F1, F2")).not.toBeInTheDocument();
  });
});
