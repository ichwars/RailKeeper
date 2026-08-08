import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api, type MasterDataEntry, type StorageLocation } from "../../shared/api";
import { ArticleManagementSettings } from "./ArticleManagementSettings";
import { StorageLocationsSettings } from "./StorageLocationsSettings";

const entry = (type: string, key: string, label: string): MasterDataEntry => ({
  id: `${type}-${key}`,
  type,
  key,
  label,
  active: true,
  sortOrder: 10,
  metadata: {},
  createdAt: "2026-08-08T10:00:00Z",
  updatedAt: "2026-08-08T10:00:00Z"
});

const locationBase = {
  archived: false,
  createdAt: "2026-08-08T10:00:00Z",
  updatedAt: "2026-08-08T10:00:00Z"
};

const locations: StorageLocation[] = [
  { ...locationBase, id: "workshop", name: "Werkstatt" },
  { ...locationBase, id: "cabinet", parentId: "workshop", name: "Schrank" },
  { ...locationBase, id: "archive", name: "Altlager", archived: true }
];

describe("ArticleManagementSettings", () => {
  beforeEach(() => {
    const entries: Record<string, MasterDataEntry[]> = {
      manufacturer: [entry("manufacturer", "tillig", "Tillig")],
      stock_unit: [entry("stock_unit", "piece", "Stück")],
      article_type: [entry("article_type", "track", "Gleis")],
      accessory_subtype: [entry("accessory_subtype", "track:straight", "Gerade")],
      accessory_custom_field: [entry("accessory_custom_field", "material", "Material")]
    };
    vi.spyOn(api, "masterData").mockImplementation(async (type) => entries[type] || []);
    vi.spyOn(api, "storageLocations").mockResolvedValue([]);
  });

  it("loads every article master-data category and provides a useful read-only view", async () => {
    const user = userEvent.setup();
    render(<ArticleManagementSettings roles={["Viewer"]} />);

    expect(await screen.findByText("Tillig")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Hersteller" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Bestandseinheiten" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Artikelarten und Unterarten" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Kontrollierte Zusatzfelder" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Lagerorte" })).toBeInTheDocument();
    expect(screen.getByText(/Änderungen sind nur für Admins und Editoren möglich/)).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /anlegen/i })).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Bestandseinheiten" }));
    expect(await screen.findByText("Stück")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Artikelarten und Unterarten" }));
    expect(await screen.findByText("Gleis")).toBeInTheDocument();
    expect(screen.getByText("Gerade")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Kontrollierte Zusatzfelder" }));
    expect(await screen.findByText("Material")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Lagerorte" }));
    expect(await screen.findByRole("heading", { name: "Lagerorthierarchie" })).toBeInTheDocument();
    expect(api.storageLocations).toHaveBeenCalledOnce();

    await waitFor(() => expect(api.masterData).toHaveBeenCalledTimes(5));
    expect(vi.mocked(api.masterData).mock.calls.map(([type]) => type)).toEqual([
      "manufacturer",
      "stock_unit",
      "article_type",
      "accessory_subtype",
      "accessory_custom_field"
    ]);
  });

  it("updates a standard article type without allowing its key to change", async () => {
    const user = userEvent.setup();
    vi.spyOn(api, "updateMasterData").mockResolvedValue(entry("article_type", "track", "Gleismaterial"));
    render(<ArticleManagementSettings roles={["Editor"]} />);

    await user.click(await screen.findByRole("button", { name: "Artikelarten und Unterarten" }));
    expect(within(screen.getByRole("region", { name: "Artikelarten" }))
      .queryByRole("button", { name: "Eintrag anlegen" })).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Gleis bearbeiten" }));

    expect(screen.getAllByText("track")).toHaveLength(2);
    expect(screen.queryByRole("textbox", { name: "Schlüssel" })).not.toBeInTheDocument();
    const label = screen.getByRole("textbox", { name: "Bezeichnung" });
    await user.clear(label);
    await user.type(label, "Gleismaterial");
    await user.click(screen.getByRole("button", { name: "Änderungen speichern" }));

    await waitFor(() => expect(api.updateMasterData).toHaveBeenCalledWith(
      "article_type",
      "track",
      expect.not.objectContaining({ key: expect.anything() })
    ));
  });

  it("creates a typed controlled custom field", async () => {
    const user = userEvent.setup();
    vi.spyOn(api, "createMasterData").mockResolvedValue({
      ...entry("accessory_custom_field", "color", "Farbe"),
      metadata: { kind: "single_select", options: ["Rot", "Grün"] }
    });
    render(<ArticleManagementSettings roles={["Admin"]} />);

    await user.click(await screen.findByRole("button", { name: "Kontrollierte Zusatzfelder" }));
    await screen.findByText("Material");
    await user.click(screen.getByRole("button", { name: "Eintrag anlegen" }));
    await user.type(screen.getByRole("textbox", { name: "Schlüssel" }), "color");
    await user.type(screen.getByRole("textbox", { name: "Bezeichnung" }), "Farbe");
    await user.click(screen.getByRole("button", { name: "Feldtyp" }));
    await user.click(screen.getByRole("option", { name: "Einfachauswahl" }));
    await user.type(screen.getByRole("textbox", { name: "Auswahlwerte" }), "Rot, Grün");
    await user.click(screen.getByRole("button", { name: "Eintrag speichern" }));

    await waitFor(() => expect(api.createMasterData).toHaveBeenCalledWith("accessory_custom_field", {
      key: "color",
      label: "Farbe",
      active: true,
      metadata: { kind: "single_select", options: ["Rot", "Grün"] }
    }));
  });
});

describe("StorageLocationsSettings", () => {
  it("creates a child location in the visible hierarchy", async () => {
    const user = userEvent.setup();
    vi.spyOn(api, "createStorageLocation").mockResolvedValue({
      ...locationBase,
      id: "drawer",
      parentId: "cabinet",
      name: "Schublade"
    });
    const onChanged = vi.fn().mockResolvedValue(undefined);
    render(<StorageLocationsSettings locations={locations} canEdit onChanged={onChanged} />);

    expect(screen.getByText("Werkstatt / Schrank")).toBeInTheDocument();
    await user.type(screen.getByRole("textbox", { name: "Bezeichnung" }), "Schublade");
    await user.click(screen.getByRole("button", { name: "Übergeordneter Lagerort" }));
    await user.click(screen.getByRole("option", { name: "Werkstatt / Schrank" }));
    await user.click(screen.getByRole("button", { name: "Lagerort speichern" }));

    await waitFor(() => expect(api.createStorageLocation).toHaveBeenCalledWith({
      name: "Schublade",
      parentId: "cabinet",
      description: undefined
    }));
    expect(onChanged).toHaveBeenCalledOnce();
  });

  it("edits, archives and reactivates locations without offering cyclic parents", async () => {
    const user = userEvent.setup();
    vi.spyOn(api, "updateStorageLocation").mockImplementation(async (id, input) => ({
      ...locationBase,
      id,
      name: input.name,
      parentId: input.parentId,
      description: input.description,
      archived: input.archived || false
    }));
    const onChanged = vi.fn().mockResolvedValue(undefined);
    render(<StorageLocationsSettings locations={locations} canEdit onChanged={onChanged} />);

    await user.click(screen.getByRole("button", { name: "Werkstatt bearbeiten" }));
    await user.click(screen.getByRole("button", { name: "Übergeordneter Lagerort" }));
    expect(screen.queryByRole("option", { name: "Werkstatt / Schrank" })).not.toBeInTheDocument();
    await user.keyboard("{Escape}");

    await user.click(screen.getByRole("button", { name: "Werkstatt archivieren" }));
    await waitFor(() => expect(api.updateStorageLocation).toHaveBeenCalledWith("workshop", {
      name: "Werkstatt",
      parentId: undefined,
      description: undefined,
      archived: true
    }));

    await user.click(screen.getByRole("button", { name: "Altlager reaktivieren" }));
    await waitFor(() => expect(api.updateStorageLocation).toHaveBeenCalledWith("archive", {
      name: "Altlager",
      parentId: undefined,
      description: undefined,
      archived: false
    }));
  });

  it.each([
    "Der Lagerort würde einen Zyklus in der Hierarchie erzeugen.",
    "Ein archivierter Lagerort kann nicht als übergeordneter Lagerort verwendet werden."
  ])("keeps hierarchy errors visible: %s", async (errorMessage) => {
    const user = userEvent.setup();
    vi.spyOn(api, "updateStorageLocation").mockRejectedValue(new Error(errorMessage));
    render(<StorageLocationsSettings locations={locations} canEdit onChanged={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: "Schrank archivieren" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(errorMessage);
  });
});
