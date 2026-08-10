import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import {
  api,
  type TrackGeometryLibrary,
  type TrackLibraryImportPreview,
  type TrackLibraryPackage
} from "../../shared/api";
import { TrackLibraryPanel } from "./TrackLibraryPanel";

const draftLibrary: TrackGeometryLibrary = {
  id: "library-1", manufacturer: "Kühn", trackSystem: "TT", gauge: "TT", scale: "1:120",
  version: "2026.1", sourceUrl: "https://example.com/catalogue.pdf", status: "draft",
  verificationNote: "", definitionCount: 1, createdAt: "2026-08-10T08:00:00Z"
};

const libraryPackage: TrackLibraryPackage = {
  format: "railkeeper.track-library", schemaVersion: 1,
  library: {
    manufacturer: "Kühn", trackSystem: "TT", gauge: "TT", scale: "1:120", version: "2026.1",
    sourceUrl: "https://example.com/catalogue.pdf", status: "verified"
  },
  definitions: [{
    articleNumber: "72620", name: "Gerades Gleis", kind: "straight", lengthMm: 128,
    sourceUrl: "https://example.com/72620", status: "verified",
    geometry: { schemaVersion: 1,
      ports: [{ id: "a", xMm: 0, yMm: 0, directionDegrees: 180 },
        { id: "b", xMm: 128, yMm: 0, directionDegrees: 0 }],
      routes: [{ id: "main", points: [{ xMm: 0, yMm: 0 }, { xMm: 128, yMm: 0 }] }]
    }
  }]
};

afterEach(() => vi.restoreAllMocks());

describe("TrackLibraryPanel", () => {
  it("lists library provenance and submits documented admin review", async () => {
    vi.spyOn(api, "trackLibraries").mockResolvedValue([draftLibrary]);
    const update = vi.spyOn(api, "updateTrackLibraryStatus").mockResolvedValue({
      ...draftLibrary, status: "verified", verificationNote: "Katalog geprüft"
    });
    render(<TrackLibraryPanel canManage />);

    expect(await screen.findByText("Kühn · TT")).toBeInTheDocument();
    expect(screen.getByText("TT · 1:120 · Version 2026.1")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "Prüfen und freigeben" }));
    const reviewDialog = screen.getByRole("dialog", { name: "Gleisbibliothek freigeben" });
    await userEvent.type(
      within(reviewDialog).getByRole("textbox", { name: /Prüfnachweis/ }),
      "Katalog geprüft"
    );
    await userEvent.click(within(reviewDialog).getByRole("button", { name: "Prüfen und freigeben" }));

    await waitFor(() => expect(update).toHaveBeenCalledWith("library-1", {
      confirmed: true, status: "verified", verificationNote: "Katalog geprüft"
    }));
  });

  it("previews an app-picked JSON file and imports only after confirmation", async () => {
    vi.spyOn(api, "trackLibraries").mockResolvedValue([]);
    const preview: TrackLibraryImportPreview = {
      package: libraryPackage, definitionCount: 1, warnings: ["verification_status_reset"],
      conflict: false, canImport: true
    };
    vi.spyOn(api, "previewTrackLibraryImport").mockResolvedValue(preview);
    const importLibrary = vi.spyOn(api, "importTrackLibrary").mockResolvedValue(draftLibrary);
    const { container } = render(<TrackLibraryPanel canManage />);
    await screen.findByText("Noch keine Gleisbibliothek installiert.");
    await userEvent.click(screen.getByRole("button", { name: "Bibliothek importieren" }));
    const file = new File([JSON.stringify(libraryPackage)], "kuehn.json", { type: "application/json" });
    Object.defineProperty(file, "text", { value: vi.fn().mockResolvedValue(JSON.stringify(libraryPackage)) });
    const input = container.ownerDocument.querySelector<HTMLInputElement>('input[type="file"]');
    if (!input) throw new Error("file input missing");
    fireEvent.change(input, { target: { files: [file] } });

    expect(await screen.findByText("Externe Prüfstatus werden zurückgesetzt. Die Bibliothek muss in RailKeeper erneut freigegeben werden."))
      .toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "Entwurf importieren" }));
    await waitFor(() => expect(importLibrary).toHaveBeenCalledWith({ confirmed: true, package: libraryPackage }));
  });

  it("retires a verified library through the app confirmation dialog", async () => {
    const verified = { ...draftLibrary, status: "verified" as const, verificationNote: "Katalog geprüft" };
    vi.spyOn(api, "trackLibraries").mockResolvedValue([verified]);
    const update = vi.spyOn(api, "updateTrackLibraryStatus").mockResolvedValue({
      ...verified, status: "retired"
    });
    render(<TrackLibraryPanel canManage />);
    await screen.findByText("Kühn · TT");
    await userEvent.click(screen.getByRole("button", { name: "Stilllegen" }));
    const retireDialog = screen.getByRole("dialog", { name: "Gleisbibliothek stilllegen?" });
    await userEvent.click(within(retireDialog).getByRole("button", { name: "Stilllegen" }));
    await waitFor(() => expect(update).toHaveBeenCalledWith("library-1", {
      confirmed: true, status: "retired", verificationNote: ""
    }));
  });
});
