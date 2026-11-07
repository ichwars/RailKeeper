import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import type { FormEvent } from "react";
import { api, ApiError, type VehicleCatalogContext } from "../../shared/api";
import { catalogFieldsFrom } from "./vehicleCatalogModel";
import { VehicleCatalogPanel } from "./VehicleCatalogPanel";

const fields = catalogFieldsFrom({ manufacturer: "Piko", name: "Base", gauge: "H0", category: "Lokomotive", gattung: "Diesellok" });
const first = { id: "r1", articleId: "article", revisionNumber: 1, fields, note: "Original", createdAt: "2026-11-06T20:00:00Z" };
const second = { ...first, id: "r2", revisionNumber: 2, fields: { ...fields, manufacturer: "Corrected", name: "Catalog name" } };
const context: VehicleCatalogContext = { targetType: "vehicle", targetId: "owned", currentFields: { ...fields, name: "Owned label" },
  authority: "catalog", overrides: { name: "Owned label" }, pinned: first, revisions: [second, first], revision: "1".repeat(64) };
const props = { targetType: "vehicle" as const, targetId: "owned", targetName: "Owned label" };
const open = async () => {
  await userEvent.click(screen.getByText("Artikelrevisionen"));
  await screen.findByText(/Festgelegter Artikelstand/);
};

describe("article revision comparison", () => {
  afterEach(() => { vi.restoreAllMocks(); });

  it("loads only on opening and protects individual overrides when selecting another revision", async () => {
    const read = vi.spyOn(api, "catalogContext").mockResolvedValue(context);
    const adopt = vi.spyOn(api, "adoptCatalogRevision").mockResolvedValue({ ...context, pinned: second });
    const saved = vi.fn().mockResolvedValue(undefined);
    render(<VehicleCatalogPanel {...props} onSaved={saved} />);
    expect(read).not.toHaveBeenCalled();
    await open();
    await userEvent.click(screen.getByRole("button", { name: "Artikelrevision vergleichen" }));
    await userEvent.click(screen.getByRole("option", { name: "Revision 2" }));
    expect(screen.getByRole("checkbox", { name: "Übernehmen: Hersteller" })).toBeChecked();
    expect(screen.getByRole("checkbox", { name: "Übernehmen: Bezeichnung" })).not.toBeChecked();
    await userEvent.click(screen.getByRole("button", { name: "Auswahl jetzt übernehmen" }));
    await waitFor(() => expect(saved).toHaveBeenCalledOnce());
    expect(adopt).toHaveBeenCalledWith("vehicle", "owned", {
      expectedRevision: context.revision, revisionId: "r2", selectedFields: ["manufacturer"]
    });
  });

  it("retains fields, note and identity across conflict refresh and explicit retry", async () => {
    vi.spyOn(api, "catalogContext").mockResolvedValueOnce(context).mockResolvedValue({ ...context, revision: "2".repeat(64) });
    const create = vi.spyOn(api, "createCatalogRevision").mockRejectedValue(new ApiError("Conflict", "catalog_conflict", 409));
    render(<VehicleCatalogPanel {...props} />);
    await open();
    await userEvent.click(screen.getByRole("button", { name: "Neue Revision entwerfen" }));
    await userEvent.type(screen.getByRole("textbox", { name: "Artikelwert: Beschreibung" }), "Unsent article draft");
    await userEvent.type(screen.getByRole("textbox", { name: "Notiz zur neuen Revision" }), "Retain this note");
    await userEvent.click(screen.getByRole("button", { name: "Revision erstellen und Auswahl speichern" }));
    await screen.findByRole("alert");
    const identity = create.mock.calls[0][2].id;
    await userEvent.click(screen.getByRole("button", { name: "Vergleich aktualisieren" }));
    await waitFor(() => expect(screen.queryByRole("alert")).not.toBeInTheDocument());
    expect(screen.getByRole("textbox", { name: "Artikelwert: Beschreibung" })).toHaveValue("Unsent article draft");
    expect(screen.getByRole("textbox", { name: "Notiz zur neuen Revision" })).toHaveValue("Retain this note");
    await userEvent.click(screen.getByRole("button", { name: "Revision erstellen und Auswahl speichern" }));
    await waitFor(() => expect(create).toHaveBeenCalledTimes(2));
    expect(create.mock.calls[1][2]).toMatchObject({ id: identity, expectedRevision: "2".repeat(64),
      fields: { description: "Unsent article draft" }, note: "Retain this note" });
  });

  it("recognizes an already stored identity and requires a deliberate new confirmation while retaining the draft", async () => {
    const read = vi.spyOn(api, "catalogContext").mockResolvedValue(context);
    const create = vi.spyOn(api, "createCatalogRevision").mockRejectedValue(new ApiError("Unknown result", "catalog_write_failed", 500));
    render(<VehicleCatalogPanel {...props} />);
    await open();
    await userEvent.click(screen.getByRole("button", { name: "Neue Revision entwerfen" }));
    await userEvent.type(screen.getByRole("textbox", { name: "Artikelwert: Beschreibung" }), "Keep after stored identity");
    await userEvent.click(screen.getByRole("button", { name: "Revision erstellen und Auswahl speichern" }));
    await screen.findByRole("alert");
    const submitted = create.mock.calls[0][2];
    read.mockResolvedValue({ ...context, revision: "3".repeat(64), revisions: [
      { ...second, id: submitted.id, fields: submitted.fields }, ...context.revisions
    ] });
    await userEvent.click(screen.getByRole("button", { name: "Vergleich aktualisieren" }));
    const next = await screen.findByRole("button", { name: "Neue Bestätigung für diesen Entwurf" });
    expect(screen.getByRole("button", { name: "Revision erstellen und Auswahl speichern" })).toBeDisabled();
    await userEvent.click(next);
    expect(screen.getByRole("textbox", { name: "Artikelwert: Beschreibung" })).toHaveValue("Keep after stored identity");
    await userEvent.click(screen.getByRole("button", { name: "Revision erstellen und Auswahl speichern" }));
    await waitFor(() => expect(create).toHaveBeenCalledTimes(2));
    expect(create.mock.calls[1][2].id).not.toBe(submitted.id);
  });

  it("ignores a late confirmation after its editor scope ends and blocks double clicks", async () => {
    vi.spyOn(api, "catalogContext").mockResolvedValue(context);
    let finish: (value: VehicleCatalogContext) => void = () => undefined;
    const create = vi.spyOn(api, "createCatalogRevision").mockReturnValue(new Promise(resolve => { finish = resolve; }));
    let active = true;
    const saved = vi.fn();
    render(<VehicleCatalogPanel {...props} captureScope={() => () => active} onSaved={saved} />);
    await open();
    await userEvent.click(screen.getByRole("button", { name: "Neue Revision entwerfen" }));
    const button = screen.getByRole("button", { name: "Revision erstellen und Auswahl speichern" });
    fireEvent.click(button); fireEvent.click(button);
    expect(create).toHaveBeenCalledOnce();
    active = false;
    finish(context);
    await Promise.resolve();
    expect(saved).not.toHaveBeenCalled();
  });

  it("shows saved pending values in read-only mode and provides no write action", async () => {
    vi.spyOn(api, "catalogContext").mockResolvedValue({ ...context, authority: "legacy", overrides: { name: "Pending proposal" } });
    const create = vi.spyOn(api, "createCatalogRevision");
    render(<VehicleCatalogPanel {...props} readonly />);
    await userEvent.click(screen.getByText("Artikelrevisionen"));
    await screen.findByText(/Diese Zuordnung enthält einen ungeklärten Vorschlag/);
    await userEvent.click(screen.getByText("Gespeicherte ungeklärte Vorschlagswerte"));
    expect(screen.getByText("Pending proposal")).toBeVisible();
    expect(screen.queryByRole("button", { name: "Neue Revision entwerfen" })).not.toBeInTheDocument();
    expect(screen.queryByRole("checkbox")).not.toBeInTheDocument();
    expect(create).not.toHaveBeenCalled();
  });

  it("keeps an incomplete article draft independent from the enclosing vehicle form submission", async () => {
    vi.spyOn(api, "catalogContext").mockResolvedValue(context);
    const submit = vi.fn((event: FormEvent<HTMLFormElement>) => event.preventDefault());
    render(<form onSubmit={submit}><VehicleCatalogPanel {...props} /><button type="submit">Hauptformular speichern</button></form>);
    await open();
    await userEvent.click(screen.getByRole("button", { name: "Neue Revision entwerfen" }));
    const input = screen.getByRole("textbox", { name: "Artikelwert: Hersteller" });
    await userEvent.clear(input);
    await userEvent.type(input, "{Enter}");
    expect(submit).not.toHaveBeenCalled();
    await userEvent.click(screen.getByRole("button", { name: "Hauptformular speichern" }));
    expect(submit).toHaveBeenCalledOnce();
  });

  it("starts no write when the captured editor scope is already closed", async () => {
    vi.spyOn(api, "catalogContext").mockResolvedValue(context);
    const create = vi.spyOn(api, "createCatalogRevision");
    let active = true;
    render(<VehicleCatalogPanel {...props} captureScope={() => () => active} />);
    await open();
    await userEvent.click(screen.getByRole("button", { name: "Neue Revision entwerfen" }));
    active = false;
    await userEvent.click(screen.getByRole("button", { name: "Revision erstellen und Auswahl speichern" }));
    expect(create).not.toHaveBeenCalled();
  });

  it("retries a failed saved-view refresh without repeating the confirmed write", async () => {
    vi.spyOn(api, "catalogContext").mockResolvedValue(context);
    const create = vi.spyOn(api, "createCatalogRevision").mockResolvedValue(context);
    const saved = vi.fn().mockRejectedValueOnce(new Error("View unavailable")).mockResolvedValue(undefined);
    render(<VehicleCatalogPanel {...props} onSaved={saved} />);
    await open();
    await userEvent.click(screen.getByRole("button", { name: "Neue Revision entwerfen" }));
    await userEvent.click(screen.getByRole("button", { name: "Revision erstellen und Auswahl speichern" }));
    await screen.findByText("Der Artikelstand ist gespeichert. Die Fahrzeugansicht konnte noch nicht aktualisiert werden.");
    await userEvent.click(screen.getByRole("button", { name: "Gespeicherte Ansicht erneut laden" }));
    await waitFor(() => expect(saved).toHaveBeenCalledTimes(2));
    expect(create).toHaveBeenCalledOnce();
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });
});
