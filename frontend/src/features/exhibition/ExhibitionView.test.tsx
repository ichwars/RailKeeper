import { render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { api } from "../../shared/api";
import { ExhibitionView } from "./ExhibitionView";

describe("ExhibitionView", () => {
  afterEach(() => vi.restoreAllMocks());

  it("keeps the Messe workspace isolated from general inventory master data", async () => {
    vi.spyOn(api, "exhibitionLists").mockResolvedValue([]);
    vi.spyOn(api, "masterData").mockResolvedValue([]);
    const masterDataAll = vi.spyOn(api, "masterDataAll").mockRejectedValue(new Error("Insufficient role."));

    render(<ExhibitionView roles={["Messe"]} />);

    expect(await screen.findByText("Messeliste")).toBeInTheDocument();
    await waitFor(() => expect(api.exhibitionLists).toHaveBeenCalledOnce());
    expect(api.masterData).toHaveBeenCalledWith("symbols", true);
    expect(masterDataAll).not.toHaveBeenCalled();
    expect(screen.queryByText("Insufficient role.")).not.toBeInTheDocument();
  });
});
