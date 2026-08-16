import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api, type MasterDataEntry, type StorageLocation } from "../../shared/api";
import { ArticleManagementSettings, type ArticleDataSection } from "./ArticleManagementSettings";
import { StorageLocationsSettings } from "./StorageLocationsSettings";

function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (reason?: unknown) => void;
  const promise = new Promise<T>((resolvePromise, rejectPromise) => {
    resolve = resolvePromise;
    reject = rejectPromise;
  });
  return { promise, resolve, reject };
}

const entry = (type: string, key: string, label: string): MasterDataEntry => ({
  id: `${type}-${key}`,
  type,
  key,
  label,
  active: true,
  sortOrder: 10,
  metadata: {},
  origin: "bundled",
  capabilities: { canDeactivate: true, canReactivate: false, canDelete: false },
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

function ArticleManagementHarness({
  roles,
  initialSection = "stock_unit"
}: {
  roles: string[];
  initialSection?: ArticleDataSection;
}) {
  const [activeSection, setActiveSection] = useState<ArticleDataSection>(initialSection);
  return (
    <ArticleManagementSettings
      roles={roles}
      activeSection={activeSection}
      onSectionChange={setActiveSection}
      onConfirmAction={({ onConfirm }) => onConfirm()}
    />
  );
}

describe("ArticleManagementSettings", () => {
  beforeEach(() => {
    const entries: Record<string, MasterDataEntry[]> = {
      manufacturer: [entry("manufacturer", "tillig", "Tillig")],
      stock_unit: [entry("stock_unit", "piece", "Piece")],
      article_type: [entry("article_type", "track", "Track")],
      accessory_subtype: [entry("accessory_subtype", "track:straight", "Straight")],
      accessory_custom_field: [entry("accessory_custom_field", "material", "Material")]
    };
    vi.spyOn(api, "managedMasterData").mockImplementation(async (type) => entries[type] || []);
    vi.spyOn(api, "storageLocations").mockResolvedValue([]);
  });

  it("loads every article master-data category and provides a useful read-only view", async () => {
    const user = userEvent.setup();
    render(<ArticleManagementHarness roles={["Viewer"]} />);

    expect(await screen.findByText("Stück")).toBeInTheDocument();
    expect(screen.queryByRole("tab", { name: "Hersteller" })).not.toBeInTheDocument();
    expect(screen.getByRole("tab", { name: "Bestandseinheiten" })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: "Artikelarten und Unterarten" })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: "Kontrollierte Zusatzfelder" })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: "Lagerorte" })).toBeInTheDocument();
    expect(screen.getByText(/Änderungen sind nur für Admins und Editoren möglich/)).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /anlegen/i })).not.toBeInTheDocument();

    await user.click(screen.getByRole("tab", { name: "Artikelarten und Unterarten" }));
    expect(await screen.findByText("Gleis")).toBeInTheDocument();
    expect(screen.getByText("Gerade")).toBeInTheDocument();
    await user.click(screen.getByRole("tab", { name: "Kontrollierte Zusatzfelder" }));
    expect(await screen.findByText("Material")).toBeInTheDocument();
    await user.click(screen.getByRole("tab", { name: "Lagerorte" }));
    expect(await screen.findByRole("heading", { name: "Lagerorthierarchie" })).toBeInTheDocument();
    expect(api.storageLocations).toHaveBeenCalledOnce();

    await waitFor(() => expect(api.managedMasterData).toHaveBeenCalledTimes(4));
    expect(vi.mocked(api.managedMasterData).mock.calls.map(([type]) => type)).toEqual([
      "stock_unit",
      "article_type",
      "accessory_subtype",
      "accessory_custom_field"
    ]);
  });

  it.each(["Planner", "Viewer"])("keeps the article settings read-only for %s", async (role) => {
    render(<ArticleManagementHarness roles={[role]} />);

    expect(await screen.findByText("Stück")).toBeInTheDocument();
    expect(screen.getByText(/Änderungen sind nur für Admins und Editoren möglich/)).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /bearbeiten|archivieren|anlegen/i })).not.toBeInTheDocument();
  });

  it("renders semantic section tabs and supports keyboard navigation", async () => {
    const user = userEvent.setup();
    render(<ArticleManagementHarness roles={["Viewer"]} />);

    expect(screen.getByRole("tablist", { name: "Bereiche der Artikelverwaltung" })).toBeInTheDocument();
    const units = screen.getByRole("tab", { name: "Bestandseinheiten" });
    const locationsTab = screen.getByRole("tab", { name: "Lagerorte" });
    expect(units).toHaveAttribute("aria-selected", "true");
    expect(screen.getByRole("tabpanel", { name: "Bestandseinheiten" })).toBeInTheDocument();

    units.focus();
    await user.keyboard("{ArrowRight}");
    const types = screen.getByRole("tab", { name: "Artikelarten und Unterarten" });
    expect(types).toHaveFocus();
    expect(types).toHaveAttribute("aria-selected", "true");
    expect(screen.getByRole("tabpanel", { name: "Artikelarten und Unterarten" })).toBeInTheDocument();

    await user.keyboard("{End}");
    expect(locationsTab).toHaveFocus();
    expect(locationsTab).toHaveAttribute("aria-selected", "true");
  });

  it("settles the active location request when an earlier master-data request finishes late", async () => {
    const user = userEvent.setup();
    const stockUnitRequest = deferred<MasterDataEntry[]>();
    vi.mocked(api.managedMasterData).mockImplementation((type) => type === "stock_unit"
      ? stockUnitRequest.promise
      : Promise.resolve([]));
    vi.mocked(api.storageLocations).mockResolvedValue(locations);
    render(<ArticleManagementHarness roles={["Viewer"]} />);

    await user.click(screen.getByRole("tab", { name: "Lagerorte" }));
    expect(await screen.findByRole("heading", { name: "Lagerorthierarchie" })).toBeInTheDocument();

    stockUnitRequest.resolve([entry("stock_unit", "piece", "Piece")]);
    await waitFor(() => expect(screen.queryByText("Artikeldaten werden geladen..."))
      .not.toBeInTheDocument());
  });

  it("shows one settled location error and retries only after an explicit action", async () => {
    const user = userEvent.setup();
    const retryRequest = deferred<StorageLocation[]>();
    vi.mocked(api.storageLocations)
      .mockRejectedValueOnce(new Error("Lagerorte konnten nicht geladen werden."))
      .mockImplementationOnce(() => retryRequest.promise);
    render(<ArticleManagementHarness roles={["Viewer"]} />);

    await user.click(await screen.findByRole("tab", { name: "Lagerorte" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("Lagerorte konnten nicht geladen werden.");
    expect(api.storageLocations).toHaveBeenCalledTimes(1);

    await user.click(screen.getByRole("button", { name: "Erneut versuchen" }));
    expect(api.storageLocations).toHaveBeenCalledTimes(2);
    retryRequest.resolve(locations);
    expect(await screen.findByRole("heading", { name: "Lagerorthierarchie" })).toBeInTheDocument();
  });

  it("updates a standard article type without allowing its key to change", async () => {
    const user = userEvent.setup();
    vi.spyOn(api, "updateMasterData").mockResolvedValue(entry("article_type", "track", "Gleismaterial"));
    render(<ArticleManagementHarness roles={["Editor"]} />);

    await user.click(await screen.findByRole("tab", { name: "Artikelarten und Unterarten" }));
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

  it("shows a localized standard label without persisting it as an override", async () => {
    const user = userEvent.setup();
    vi.spyOn(api, "updateMasterData").mockResolvedValue(entry("article_type", "track", "Track"));
    render(<ArticleManagementHarness roles={["Editor"]} />);

    await user.click(await screen.findByRole("tab", { name: "Artikelarten und Unterarten" }));
    await user.click(screen.getByRole("button", { name: "Gleis bearbeiten" }));
    expect(screen.getByRole("textbox", { name: "Bezeichnung" })).toHaveValue("Gleis");
    await user.click(screen.getByRole("button", { name: "Änderungen speichern" }));

    await waitFor(() => expect(api.updateMasterData).toHaveBeenCalledWith(
      "article_type",
      "track",
      expect.objectContaining({ label: "Track" })
    ));
  });

  it("preserves an active-state toggle when saving an editor that was already open", async () => {
    const user = userEvent.setup();
    vi.spyOn(api, "setMasterDataActive").mockImplementation(async (type, key, active) => ({
      ...entry(type, key, "Track"),
      active,
      capabilities: { canDeactivate: active, canReactivate: !active, canDelete: false }
    }));
    vi.spyOn(api, "updateMasterData").mockImplementation(async (type, key, input) => ({
      ...entry(type, key, input.label),
      active: false,
      capabilities: { canDeactivate: false, canReactivate: true, canDelete: false }
    }));
    render(<ArticleManagementHarness roles={["Editor"]} />);

    await user.click(await screen.findByRole("tab", { name: "Artikelarten und Unterarten" }));
    await screen.findByText("Gleis");
    await user.click(screen.getByRole("button", { name: "Gleis bearbeiten" }));
    await user.click(screen.getByRole("button", { name: "Gleis deaktivieren" }));
    await waitFor(() => expect(api.setMasterDataActive).toHaveBeenCalledWith("article_type", "track", false));

    const label = screen.getByRole("textbox", { name: "Bezeichnung" });
    await user.clear(label);
    await user.type(label, "Gleismaterial");
    await user.click(screen.getByRole("button", { name: "Änderungen speichern" }));

    await waitFor(() => expect(api.updateMasterData).toHaveBeenLastCalledWith(
      "article_type",
      "track",
      expect.objectContaining({ label: "Gleismaterial", active: false })
    ));
  });

  it("shows origin and only deletes an unused custom entry after confirmation", async () => {
    const user = userEvent.setup();
    const custom = {
      ...entry("stock_unit", "box", "Box"),
      origin: "custom" as const,
      capabilities: { canDeactivate: true, canReactivate: false, canDelete: true }
    };
    vi.mocked(api.managedMasterData).mockResolvedValue([custom]);
    vi.spyOn(api, "deleteMasterData").mockResolvedValue(undefined);
    const onConfirmAction = vi.fn(({ onConfirm }: { onConfirm: () => void }) => onConfirm());
    render(
      <ArticleManagementSettings roles={["Editor"]} activeSection="stock_unit"
        onSectionChange={vi.fn()} onConfirmAction={onConfirmAction} />
    );

    expect(await screen.findByText("Eigener Eintrag")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Box endgültig löschen" }));

    expect(onConfirmAction).toHaveBeenCalledWith(expect.objectContaining({
      title: "Stammdateneintrag endgültig löschen",
      confirmLabel: "Endgültig löschen",
      danger: true
    }));
    await waitFor(() => expect(api.deleteMasterData).toHaveBeenCalledWith("stock_unit", "box"));
  });

  it("creates a typed controlled custom field", async () => {
    const user = userEvent.setup();
    const createdField = {
      ...entry("accessory_custom_field", "color", "Farbe"),
      origin: "custom" as const,
      capabilities: { canDeactivate: true, canReactivate: false, canDelete: true },
      metadata: { kind: "single_select", options: ["Rot", "Grün"] }
    };
    vi.spyOn(api, "createMasterData").mockResolvedValue({
      ...createdField,
      capabilities: undefined
    });
    vi.mocked(api.managedMasterData).mockImplementation(async (type) => {
      if (type !== "accessory_custom_field") return [];
      return vi.mocked(api.createMasterData).mock.calls.length > 0
        ? [entry("accessory_custom_field", "material", "Material"), createdField]
        : [entry("accessory_custom_field", "material", "Material")];
    });
    render(<ArticleManagementHarness roles={["Admin"]} />);

    await user.click(await screen.findByRole("tab", { name: "Kontrollierte Zusatzfelder" }));
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
    expect(await screen.findByRole("button", { name: "Farbe endgültig löschen" })).toBeInTheDocument();
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

  it("closes an open location editor after archiving that location", async () => {
    const user = userEvent.setup();
    vi.spyOn(api, "updateStorageLocation").mockImplementation(async (id, input) => ({
      ...locationBase,
      id,
      name: input.name,
      parentId: input.parentId,
      description: input.description,
      archived: input.archived || false
    }));
    render(<StorageLocationsSettings locations={locations} canEdit onChanged={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: "Werkstatt bearbeiten" }));
    expect(screen.getByRole("heading", { name: "Lagerort bearbeiten" })).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Werkstatt archivieren" }));

    await waitFor(() => expect(screen.queryByRole("heading", { name: "Lagerort bearbeiten" }))
      .not.toBeInTheDocument());
    expect(screen.getByRole("heading", { name: "Lagerort anlegen" })).toBeInTheDocument();
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
