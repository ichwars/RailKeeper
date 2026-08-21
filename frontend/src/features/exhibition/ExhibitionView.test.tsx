import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import { api } from "../../shared/api";
import { ExhibitionView } from "./ExhibitionView";

describe("ExhibitionView", () => {
  afterEach(() => vi.restoreAllMocks());

  it("keeps the Messe workspace isolated from general inventory master data", async () => {
    vi.spyOn(api, "exhibitionLists").mockResolvedValue([]);
    const masterData = vi.spyOn(api, "masterData").mockResolvedValue([]);
    const masterDataAll = vi.spyOn(api, "masterDataAll").mockRejectedValue(new Error("Insufficient role."));

    render(<ExhibitionView roles={["Messe"]} />);

    expect(await screen.findByText("Ausstellung")).toBeInTheDocument();
    await waitFor(() => expect(api.exhibitionLists).toHaveBeenCalledOnce());
    expect(masterData).toHaveBeenCalledOnce();
    expect(masterData).toHaveBeenCalledWith("symbols");
    expect(masterDataAll).not.toHaveBeenCalled();
    expect(screen.queryByText("Insufficient role.")).not.toBeInTheDocument();
  });

  it("offers PluX12 from the shared vehicle adapter options", async () => {
    const user = userEvent.setup();
    vi.spyOn(api, "exhibitionLists").mockResolvedValue([{
      id: "list-1",
      designation: "Clubabend",
      date: "2026-08-15",
      endDate: "2026-08-15",
      status: "open",
      revision: 1,
      locked: false,
      entryCount: 0,
      createdAt: "2026-08-15T18:00:00Z",
      updatedAt: "2026-08-15T18:00:00Z"
    }]);
    vi.spyOn(api, "exhibitionWorkspace").mockResolvedValue({
      list: {
        id: "list-1", designation: "Clubabend", date: "2026-08-15", endDate: "2026-08-15",
        status: "open", revision: 1, locked: false, entryCount: 0,
        createdAt: "2026-08-15T18:00:00Z", updatedAt: "2026-08-15T18:00:00Z"
      },
      summary: { entryCount: 0, ownerCount: 0, conflictCount: 0, readyCount: 0 },
      readiness: { total: 0, addressesChecked: 0, functionsDocumented: 0, imagesPresent: 0, problems: 0 },
      entries: [], conflicts: [], dayScopes: ["day1"]
    });
    vi.spyOn(api, "exhibitionEntries").mockResolvedValue([]);
    vi.spyOn(api, "masterData").mockResolvedValue([]);
    vi.spyOn(api, "masterDataAll").mockResolvedValue({});

    render(<ExhibitionView roles={["Admin"]} />);
    await user.click(await screen.findByRole("button", { name: "Eintrag hinzufügen" }));

    const adapter = screen.getByLabelText("Adapter / Schnittstelle");
    await user.click(adapter);
    expect(screen.getByRole("option", { name: "PluX12" })).toHaveTextContent("PluX12");
  });
});
