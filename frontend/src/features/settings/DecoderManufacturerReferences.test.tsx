import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, expect, it, vi } from "vitest";
import { api } from "../../shared/api";
import type { DecoderCVReference } from "../../shared/decoderLibraryModel";
import { DecoderManufacturerReferences } from "./DecoderManufacturerReferences";

const reference: DecoderCVReference = {
  id: "manufacturer-table", revision: "2026-10-20.1", modelIds: ["model"], active: true, status: "partial",
  title: { de: "Herstellertabelle", en: "Manufacturer table" }, note: { de: "Firmware beachten.", en: "Check firmware." },
  sources: [{ title: "Originalanleitung", version: "Edition 2", url: "https://example.test/manual.pdf", scope: "Modell" }],
  rows: [
    { cv: 257, label: "Ausgang A", page: 11, sourceIndex: 0, indexScope: "indexed", indexCV31: 18, indexCV32: 0,
      applicability: "manual", valueRange: "0–255", factoryValue: "0" },
    { cv: 267, label: "Ungeklärte Indexseite", page: 12, sourceIndex: 0, indexScope: "unspecified", applicability: "manual" },
    { cv: 67, cvEnd: 94, label: "Geschwindigkeitstabelle", page: 8, sourceIndex: 0, indexScope: "base", applicability: "model" },
  ],
};
beforeEach(() => { localStorage.clear(); vi.spyOn(api, "decoderReferences").mockResolvedValue([structuredClone(reference)]); });

it("filters CV ranges while keeping exact index pages, source and unknown applicability visible", async () => {
  const user = userEvent.setup();
  render(<DecoderManufacturerReferences modelId="model" />);
  await screen.findByRole("heading", { name: "Herstellertabelle" });
  expect(api.decoderReferences).toHaveBeenCalledWith("model");
  expect(screen.getByText("CV31=18 / CV32=0")).toBeInTheDocument();
  expect(screen.getByText("Indexseite nicht eindeutig erfasst")).toBeInTheDocument();
  expect(screen.getByRole("link", { name: "Quelle, PDF-Seite 11" })).toHaveAttribute("href", "https://example.test/manual.pdf#page=11");
  await user.type(screen.getByRole("searchbox"), "CV80");
  const table = screen.getByRole("table");
  expect(within(table).getByText("Geschwindigkeitstabelle")).toBeInTheDocument();
  expect(within(table).queryByText("Ausgang A")).not.toBeInTheDocument();
  expect(screen.queryByRole("button", { name: /übernehmen|speichern/i })).not.toBeInTheDocument();
});

it("shows gaps, empty results and a retry after a reference loading failure", async () => {
  const user = userEvent.setup();
  vi.mocked(api.decoderReferences).mockRejectedValueOnce(new Error("offline"));
  const { unmount } = render(<DecoderManufacturerReferences modelId="model" />);
  expect(await screen.findByRole("alert")).toHaveTextContent("konnten nicht geladen werden");
  await user.click(screen.getByRole("button", { name: "Aktualisieren" }));
  await screen.findByRole("table");
  await user.type(screen.getByRole("searchbox"), "unauffindbar");
  expect(screen.queryByRole("table")).not.toBeInTheDocument();
  expect(screen.getByText("Keine passenden Definitionen gefunden.")).toBeInTheDocument();
  unmount();
  vi.mocked(api.decoderReferences).mockResolvedValue([{ ...reference, status: "unavailable", rows: [] }]);
  render(<DecoderManufacturerReferences modelId="model" />);
  await screen.findByText(/Offene Dokumentationslücke/);
  expect(screen.queryByRole("searchbox")).not.toBeInTheDocument();
});

it("does not render a late response for a previously selected model", async () => {
  let resolve: (value: DecoderCVReference[]) => void = () => undefined;
  vi.mocked(api.decoderReferences).mockImplementationOnce(() => new Promise(r => { resolve = r; }));
  const { rerender } = render(<DecoderManufacturerReferences modelId="old" />);
  vi.mocked(api.decoderReferences).mockResolvedValue([]);
  rerender(<DecoderManufacturerReferences modelId="new" />);
  await screen.findByText("Für dieses Modell ist noch keine belastbare CV-Tabelle hinterlegt.");
  resolve([reference]);
  expect(screen.queryByRole("heading", { name: "Herstellertabelle" })).not.toBeInTheDocument();
});
