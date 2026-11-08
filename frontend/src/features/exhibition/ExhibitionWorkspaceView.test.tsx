import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { api, ApiError, ExhibitionList, ExhibitionWorkspace, MasterDataEntry } from "../../shared/api";
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

describe("exhibition mutation failures and retained input", () => {
  beforeEach(() => {
    vi.spyOn(api,"exhibitionLists").mockResolvedValue([list]);
    vi.spyOn(api,"exhibitionWorkspace").mockResolvedValue(workspace);
    vi.spyOn(api,"vehicles").mockResolvedValue([]);
    vi.spyOn(api,"masterData").mockResolvedValue([]);
    vi.spyOn(api,"masterDataAll").mockResolvedValue({});
  });

  it("retains edited event fields after a failed save and retries the original revision",async()=>{
    const update=vi.spyOn(api,"updateExhibitionList").mockRejectedValueOnce(new Error("Event conflict")).mockResolvedValue(list);
    const user=userEvent.setup();render(<ExhibitionView roles={["Admin"]}/>);
    await screen.findByText("BR 103 113-7");
    await user.click(screen.getByRole("button",{name:"Weitere Aktionen"}));
    await user.click(screen.getByRole("button",{name:"Veranstaltung bearbeiten"}));
    const dialog=screen.getByRole("dialog",{name:"Veranstaltung bearbeiten"});
    for(const [label,value]of [["Bezeichnung","Updated event"],["Ort","Updated location"],["Beschreibung","Public details"],["Organisation (intern)","Keep internal details"]]) {
      fireEvent.change(within(dialog).getByLabelText(label),{target:{value}});
    }
    fireEvent.change(within(dialog).getByLabelText("Beginn"),{target:{value:"2026-11-09"}});
    fireEvent.blur(within(dialog).getByLabelText("Beginn"));
    fireEvent.change(within(dialog).getByLabelText("Ende"),{target:{value:"2026-11-10"}});
    fireEvent.blur(within(dialog).getByLabelText("Ende"));
    fireEvent.submit(dialog);
    expect(await within(dialog).findByRole("alert")).toHaveTextContent("Event conflict");
    expect(within(dialog).getByLabelText("Ort")).toHaveValue("Updated location");
    fireEvent.submit(dialog);
    await waitFor(()=>expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
    expect(update).toHaveBeenLastCalledWith(list.id,expect.objectContaining({designation:"Updated event",date:"2026-11-09",endDate:"2026-11-10",location:"Updated location",description:"Public details",organizationNotes:"Keep internal details",expectedRevision:4}));
  });

  it("shows a rejected entry save beside the retained guest input and never changes inventory",async()=>{
    const save=vi.spyOn(api,"createExhibitionEntry").mockRejectedValueOnce(new Error("Entry conflict")).mockResolvedValue(workspace.entries[0]);
    const user=userEvent.setup();render(<ExhibitionView roles={["Messe"]}/>);
    await screen.findByText("BR 103 113-7");
    await user.click(screen.getByRole("button",{name:"Eintrag hinzufügen"}));
    const dialog=screen.getByRole("dialog",{name:"Eintrag hinzufügen"});
    await user.type(within(dialog).getByLabelText("Besitzer"),"Guest owner");
    await user.type(within(dialog).getByLabelText("Lok-Bezeichnung"),"Guest unit");
    fireEvent.submit(dialog);
    expect(await within(dialog).findByRole("alert")).toHaveTextContent("Entry conflict");
    expect(within(dialog).getByLabelText("Lok-Bezeichnung")).toHaveValue("Guest unit");
    expect(api.vehicles).not.toHaveBeenCalled();
    fireEvent.submit(dialog);
    await waitFor(()=>expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
    expect(save).toHaveBeenLastCalledWith(list.id,expect.objectContaining({locomotiveName:"Guest unit",owner:"Guest owner"}));
  });

  it("retains a conflict reason after rejection and submits it against the unchanged list revision",async()=>{
    const next={...workspace,conflicts:[{...workspace.conflicts[0],excepted:true,exceptionReason:"Reviewed reason"}]};
    const save=vi.spyOn(api,"setExhibitionConflictException").mockRejectedValueOnce(new Error("Exception conflict")).mockResolvedValue(next);
    const user=userEvent.setup();render(<ExhibitionView roles={["Messe"]}/>);
    await screen.findByText("BR 103 113-7");
    await user.click(screen.getByRole("button",{name:"Konflikte prüfen"}));
    const dialog=screen.getByRole("dialog",{name:"Konflikte und Prüfungen"});
    const reason=within(dialog).getByPlaceholderText("Begründung der Ausnahme");
    await user.click(within(dialog).getByRole("button",{name:"Ausnahme speichern"}));
    expect(save).not.toHaveBeenCalled();
    await user.type(reason,"  Reviewed reason  ");
    await user.click(within(dialog).getByRole("button",{name:"Ausnahme speichern"}));
    expect(await within(dialog).findByRole("alert")).toHaveTextContent("Exception conflict");
    expect(reason).toHaveValue("  Reviewed reason  ");
    await user.click(within(dialog).getByRole("button",{name:"Ausnahme speichern"}));
    await within(dialog).findByText(/Dokumentierte Ausnahme: Reviewed reason/);
    expect(save).toHaveBeenLastCalledWith(list.id,"conflict-1","Reviewed reason",4);
  });

  it("keeps events and entries after failed deletion and respects cancelled confirmations",async()=>{
    const remove=vi.spyOn(api,"deleteExhibitionList").mockRejectedValueOnce(new Error("Delete unavailable")).mockResolvedValue(undefined);
    const confirm=vi.spyOn(window,"confirm").mockReturnValueOnce(false).mockReturnValue(true);
    const user=userEvent.setup();render(<ExhibitionView roles={["Admin"]}/>);
    await screen.findByText("BR 103 113-7");
    await user.click(screen.getByRole("button",{name:"Weitere Aktionen"}));
    const menu=screen.getByRole("button",{name:"Weitere Aktionen"}).parentElement!;
    await user.click(within(menu).getByRole("button",{name:"Löschen"}));
    expect(remove).not.toHaveBeenCalled();
    await user.click(within(menu).getByRole("button",{name:"Löschen"}));
    expect(await screen.findByRole("alert")).toHaveTextContent("Delete unavailable");
    expect(screen.getByText("BR 103 113-7")).toBeInTheDocument();
    expect(confirm).toHaveBeenCalledTimes(2);
  });

  it("opens lock confirmation after a conflict response without silently confirming it",async()=>{
    const change=vi.spyOn(api,"setExhibitionStatus").mockRejectedValueOnce(new ApiError("Conflicts","exhibition_conflicts",409)).mockResolvedValue({...list,status:"locked",locked:true});
    const user=userEvent.setup();render(<ExhibitionView roles={["Admin"]}/>);
    await screen.findByText("BR 103 113-7");
    await user.click(screen.getByRole("button",{name:"Liste sperren"}));
    const dialog=await screen.findByRole("dialog",{name:"Liste trotz Konflikten sperren?"});
    expect(change).toHaveBeenCalledTimes(1);
    await user.type(within(dialog).getByLabelText("Begründung"),"Explicit reason");
    fireEvent.submit(dialog);
    await waitFor(()=>expect(change).toHaveBeenLastCalledWith(list.id,{status:"locked",expectedRevision:4,reason:"Explicit reason",confirmConflicts:true}));
  });
  it("retains an edited entry when deletion is cancelled or rejected and closes only after confirmation", async () => {
    const remove = vi.spyOn(api, "deleteExhibitionEntry").mockRejectedValueOnce(new Error("Entry delete unavailable")).mockResolvedValue(undefined);
    const confirm = vi.spyOn(window, "confirm").mockReturnValueOnce(false).mockReturnValue(true);
    const user = userEvent.setup(); render(<ExhibitionView roles={["Admin"]} />);
    await screen.findByText("BR 103 113-7");
    await user.click(screen.getByRole("button", { name: "BR 103 113-7 bearbeiten" }));
    const dialog = screen.getByRole("dialog", { name: "Eintrag bearbeiten" });
    await user.type(within(dialog).getByLabelText("Besitzer"), " retained draft");
    await user.click(within(dialog).getByRole("button", { name: "Löschen" }));
    expect(remove).not.toHaveBeenCalled();
    expect(dialog).toBeInTheDocument();
    await user.click(within(dialog).getByRole("button", { name: "Löschen" }));
    expect(await within(dialog).findByRole("alert")).toHaveTextContent("Entry delete unavailable");
    expect(within(dialog).getByLabelText("Besitzer")).toHaveValue("Michael Weber retained draft");
    await user.click(within(dialog).getByRole("button", { name: "Löschen" }));
    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
    expect(confirm).toHaveBeenCalledTimes(3);
    expect(remove).toHaveBeenLastCalledWith(list.id, "entry-1");
  });

});
