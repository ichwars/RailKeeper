import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
  api,
  type AccessoryArticle,
  type AccessoryArticleListResult,
  type AccessoryAsset,
  type AccessoryInstallation,
  type AccessoryPurchase,
  type AccessoryReservation,
  type AccessoryUsageEvent,
  type Layout,
  type MasterDataEntry,
  type StorageLocation
} from "../../shared/api";
import { setLanguage } from "../../shared/i18n";
import { articleViewSettingKey } from "./articleViewMode";
import { AccessoriesView } from "./AccessoriesView";

const overview: AccessoryArticleListResult = {
  items: [{
    id: "article-1",
    inventoryNumber: "RK-ART-000001",
    manufacturer: "Tillig",
    articleNumber: "83101",
    name: "Gerades Modellgleis",
    articleType: "track",
    subtype: "straight",
    gauges: ["TT"],
    inventoryStrategy: "quantity",
    archived: false,
    owned: 18,
    available: 12,
    reserved: 4,
    installed: 2,
    locationNames: ["Werkstatt / Schrank A"],
    hasUsageHistory: true,
    careHintCount: 1,
    updatedAt: "2026-08-08T10:00:00Z",
    attributes: []
  }],
  metrics: {
    articleCount: 24,
    articleTypeCount: 5,
    available: 81,
    locationCount: 7,
    reserved: 6,
    installed: 14,
    careHintCount: 3
  },
  filters: {
    manufacturers: ["Tillig", "Viessmann"],
    articleTypes: ["track", "signal"],
    gauges: ["TT", "H0"],
    storageLocations: [{ id: "location-1", name: "Werkstatt / Schrank A" }]
  }
};
const straightSubtype: MasterDataEntry = {
  id: "article-subtype-track-straight", type: "accessory_subtype", key: "track:straight", label: "Straight",
  active: true, sortOrder: 10, metadata: {}, createdAt: "2026-08-08T08:00:00Z",
  updatedAt: "2026-08-08T08:00:00Z"
};
const manufacturerEntry: MasterDataEntry = {
  id: "manufacturer:tillig", type: "manufacturer", key: "tillig", label: "Tillig",
  active: true, sortOrder: 10, metadata: {}, createdAt: "2026-08-08T08:00:00Z",
  updatedAt: "2026-08-08T08:00:00Z"
};
const gaugeEntry: MasterDataEntry = {
  id: "gauge:tt", type: "gauge", key: "tt", label: "TT",
  active: true, sortOrder: 10, metadata: {}, createdAt: "2026-08-08T08:00:00Z",
  updatedAt: "2026-08-08T08:00:00Z"
};
const stockUnitEntry: MasterDataEntry = {
  id: "stock-unit-piece", type: "stock_unit", key: "piece", label: "Piece",
  active: true, sortOrder: 10, metadata: {}, createdAt: "2026-08-08T08:00:00Z",
  updatedAt: "2026-08-08T08:00:00Z"
};
const renamedInactiveTrackType: MasterDataEntry = {
  id: "article-type-track", type: "article_type", key: "track", label: "Gleismaterial", active: false,
  sortOrder: 10, metadata: {}, createdAt: "2026-08-08T08:00:00Z", updatedAt: "2026-08-08T08:00:00Z"
};
const standardArticleTypes: MasterDataEntry[] = [
  ["track", "Track"], ["signal", "Signal"], ["decoder", "Decoder"],
  ["electrical_control", "Electrical control"], ["building_equipment", "Building equipment"],
  ["landscape_consumable", "Landscape consumable"], ["lighting", "Lighting"], ["other", "Other"]
].map(([key, label], index) => ({
  id: `article-type-${key}`, type: "article_type", key: key!, label: label!, active: true,
  sortOrder: (index + 1) * 10, metadata: {}, createdAt: "2026-08-08T08:00:00Z",
  updatedAt: "2026-08-08T08:00:00Z"
}));

function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((next) => { resolve = next; });
  return { promise, resolve };
}

describe("AccessoriesView", () => {
  afterEach(() => vi.unstubAllGlobals());

  beforeEach(() => {
    window.localStorage.removeItem(articleViewSettingKey);
    setLanguage("de");
    vi.spyOn(api, "accessoryArticles").mockResolvedValue(overview);
    vi.spyOn(api, "storageLocations").mockResolvedValue([]);
    vi.spyOn(api, "masterData").mockImplementation(async (type) =>
      type === "accessory_subtype" ? [straightSubtype]
        : type === "article_type" ? standardArticleTypes
          : type === "manufacturer" ? [manufacturerEntry]
            : type === "gauge" ? [gaugeEntry]
              : type === "stock_unit" ? [stockUnitEntry] : []);
    vi.spyOn(api, "archiveAccessoryProduct").mockResolvedValue({} as never);
    vi.spyOn(api, "restoreAccessoryProduct").mockResolvedValue({} as never);
    vi.spyOn(api, "deleteAccessoryProduct").mockResolvedValue(undefined);
  });

  it("keeps permanent deletion hidden from editors", async () => {
    const user = userEvent.setup();
    render(<AccessoriesView roles={["Editor"]} />);
    await screen.findByText("Gerades Modellgleis");

    await user.click(screen.getByRole("button", { name: /Weitere Aktionen/ }));
    expect(screen.queryByRole("menuitem", { name: "Artikel löschen" })).not.toBeInTheDocument();
  });

  it("confirms admin deletion with article identity and reloads", async () => {
    const user = userEvent.setup();
    render(<AccessoriesView roles={["Admin"]} />);
    await screen.findByText("Gerades Modellgleis");

    await user.click(screen.getByRole("button", { name: /Weitere Aktionen/ }));
    await user.click(screen.getByRole("menuitem", { name: "Artikel löschen" }));
    const dialog = screen.getByRole("dialog", { name: "Artikel endgültig löschen" });
    expect(within(dialog).getByText(/RK-ART-000001/)).toBeInTheDocument();
    expect(within(dialog).getByText(/Gerades Modellgleis/)).toBeInTheDocument();
    await user.click(within(dialog).getByRole("button", { name: "Endgültig löschen" }));

    expect(api.deleteAccessoryProduct).toHaveBeenCalledWith("article-1");
    await waitFor(() => expect(api.accessoryArticles).toHaveBeenCalledTimes(2));
    expect(screen.queryByRole("dialog", { name: "Artikel endgültig löschen" }))
      .not.toBeInTheDocument();
  });

  it("keeps the dialog and article visible when deletion is blocked", async () => {
    vi.mocked(api.deleteAccessoryProduct).mockRejectedValueOnce(
      new Error("Accessory product has stock or usage history and cannot be deleted.")
    );
    const user = userEvent.setup();
    render(<AccessoriesView roles={["Admin"]} />);
    await screen.findByText("Gerades Modellgleis");

    await user.click(screen.getByRole("button", { name: /Weitere Aktionen/ }));
    await user.click(screen.getByRole("menuitem", { name: "Artikel löschen" }));
    await user.click(screen.getByRole("button", { name: "Endgültig löschen" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("stock or usage history");
    expect(screen.getByRole("dialog", { name: "Artikel endgültig löschen" })).toBeInTheDocument();
    expect(screen.getByText("Gerades Modellgleis")).toBeInTheDocument();
  });

  it("localizes the admin delete action and confirmation in English", async () => {
    setLanguage("en");
    const user = userEvent.setup();
    render(<AccessoriesView roles={["Admin"]} />);
    await screen.findByText("Gerades Modellgleis");

    await user.click(screen.getByRole("button", { name: /More actions/ }));
    await user.click(screen.getByRole("menuitem", { name: "Delete article" }));

    const dialog = screen.getByRole("dialog", { name: "Permanently delete article" });
    expect(within(dialog).getByRole("button", { name: "Delete permanently" })).toBeInTheDocument();
    expect(within(dialog).getByText(/RK-ART-000001: Gerades Modellgleis/)).toBeInTheDocument();
  });

  it("renders the table article overview with four global metrics by default", async () => {
    render(<AccessoriesView roles={["Editor"]} />);

    expect(await screen.findByRole("heading", { name: "Zubehör" })).toBeInTheDocument();
    expect(screen.getByText("Modellbahnartikel suchen, erfassen und pflegen")).toBeInTheDocument();
    expect(screen.queryByText("WERKSTATT UND SAMMLUNG")).not.toBeInTheDocument();
    const metrics = screen.getByLabelText("Kennzahlen der Artikelverwaltung");
    expect(metrics).toHaveClass("inventory-status-row");
    expect(screen.getAllByTestId("article-metric")).toHaveLength(4);
    for (const metric of screen.getAllByTestId("article-metric")) {
      expect(metric).toHaveClass("inventory-status-card");
    }
    expect(screen.getByText("24 Artikel")).toBeInTheDocument();
    expect(screen.getByText("5 Arten")).toBeInTheDocument();
    expect(screen.getByText("81 frei")).toBeInTheDocument();
    expect(screen.getByText("7 Lagerorte")).toBeInTheDocument();
    expect(screen.getByText("20 gebunden")).toBeInTheDocument();
    expect(screen.getByText("6 reserviert · 14 eingebaut")).toBeInTheDocument();
    expect(screen.getByText("6 reserviert · 14 eingebaut").closest("article")).toHaveClass("wide");
    expect(screen.getByText("3")).toBeInTheDocument();
    const careMetric = screen.getByText("3").closest("article");
    expect(careMetric).not.toBeNull();
    expect(within(careMetric!).queryByRole("button")).not.toBeInTheDocument();
    expect(screen.queryByRole("tablist")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Tabellenansicht" })).toHaveClass("active");
    expect(screen.getByRole("button", { name: "Kachelansicht" })).toBeInTheDocument();
    expect(screen.getByText("1 Ergebnis")).toBeInTheDocument();
    const list = screen.getByRole("region", { name: "Artikel" });
    expect(list).toHaveClass("inventory-panel");
    expect(within(list).getByRole("searchbox", { name: "Artikel suchen" })).toBeInTheDocument();
    expect(within(list).getByRole("table")).toHaveClass("inventory-table");
  });

  it("switches desktop views without reloading data and restores the persisted choice", async () => {
    const user = userEvent.setup();
    const view = render(<AccessoriesView roles={["Editor"]} />);
    await screen.findByText("Gerades Modellgleis");

    expect(screen.getByRole("button", { name: "Tabellenansicht" })).toHaveClass("active");
    expect(screen.getByRole("table")).toBeInTheDocument();
    const requestCount = vi.mocked(api.accessoryArticles).mock.calls.length;

    await user.click(screen.getByRole("button", { name: "Kachelansicht" }));
    expect(screen.getByRole("button", { name: "Kachelansicht" })).toHaveClass("active");
    expect(screen.getByRole("list", { name: "Artikel-Kachelansicht" })).toBeInTheDocument();
    expect(screen.queryByRole("list", { name: "Kompakte Artikelliste" })).not.toBeInTheDocument();
    expect(api.accessoryArticles).toHaveBeenCalledTimes(requestCount);
    expect(window.localStorage.getItem(articleViewSettingKey)).toBe("cards");

    view.unmount();
    render(<AccessoriesView roles={["Editor"]} />);
    expect(await screen.findByRole("list", { name: "Artikel-Kachelansicht" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Kachelansicht" })).toHaveClass("active");
  });

  it("localizes the view controls", async () => {
    setLanguage("en");
    vi.mocked(api.accessoryArticles).mockResolvedValueOnce({
      ...overview,
      items: overview.items.map((item) => ({ ...item, name: "Straight model track" }))
    });
    render(<AccessoriesView roles={["Viewer"]} />);

    await screen.findByText("Straight model track");
    expect(screen.getByRole("button", { name: "Table view" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Card view" })).toBeInTheDocument();
  });

  it("uses the compact article list automatically at the mobile breakpoint", async () => {
    const matchMedia = vi.fn().mockReturnValue({
      matches: true,
      media: "(max-width: 900px)",
      onchange: null,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      addListener: vi.fn(),
      removeListener: vi.fn(),
      dispatchEvent: vi.fn()
    });
    vi.stubGlobal("matchMedia", matchMedia);

    render(<AccessoriesView roles={["Editor"]} />);

    expect(await screen.findByRole("list", { name: "Kompakte Artikelliste" })).toBeInTheDocument();
    expect(screen.queryByRole("table")).not.toBeInTheDocument();
    expect(matchMedia).toHaveBeenCalledWith("(max-width: 900px)");
  });

  it("uses English singular metric nouns for one article and one type", async () => {
    setLanguage("en");
    vi.mocked(api.accessoryArticles).mockResolvedValueOnce({ ...overview, metrics: {
      ...overview.metrics, articleCount: 1, articleTypeCount: 1
    } });
    render(<AccessoriesView roles={["Viewer"]} />);
    expect(await screen.findByText("1 article")).toBeInTheDocument();
    expect(screen.getByText("1 type")).toBeInTheDocument();
    setLanguage("de");
  });

  it("uses renamed inactive result types in the table and filter", async () => {
    const user = userEvent.setup();
    vi.mocked(api.masterData).mockImplementation(async (type) =>
      type === "accessory_subtype" ? [straightSubtype]
        : type === "article_type" ? [renamedInactiveTrackType] : []);
    render(<AccessoriesView roles={["Viewer"]} />);

    expect(await screen.findByText("Gleismaterial")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Artikelart" }));
    expect(screen.getByRole("option", { name: "Gleismaterial" })).toBeInTheDocument();
  });

  it("searches instantly, maps every visible filter, and resets them", async () => {
    const user = userEvent.setup();
    render(<AccessoriesView roles={["Viewer"]} />);
    await screen.findByText("Gerades Modellgleis");

    await user.type(screen.getByRole("searchbox", { name: "Artikel suchen" }), "Tillig 83101");
    await waitFor(() => expect(api.accessoryArticles).toHaveBeenLastCalledWith(expect.objectContaining({
      query: "Tillig 83101"
    })));

    const choose = async (label: string, option: string) => {
      await user.click(screen.getByRole("button", { name: label }));
      await user.click(screen.getByRole("option", { name: option }));
    };
    await choose("Artikelart", "Gleis");
    await choose("Hersteller", "Tillig");
    await choose("Spurweite", "TT");
    await choose("Status", "Verfügbar");
    await choose("Lagerort", "Werkstatt / Schrank A");

    await waitFor(() => expect(api.accessoryArticles).toHaveBeenLastCalledWith({
      query: "Tillig 83101",
      articleTypes: ["track"],
      manufacturer: "Tillig",
      gauges: ["TT"],
      statuses: ["available"],
      locationId: "location-1",
      sort: "inventoryNumber",
      direction: "asc"
    }));

    await user.click(screen.getByRole("button", { name: "Filter zurücksetzen" }));
    await waitFor(() => expect(api.accessoryArticles).toHaveBeenLastCalledWith({
      sort: "inventoryNumber",
      direction: "asc"
    }));
  });

  it("keeps a metric-applied status filter visible and removable", async () => {
    const user = userEvent.setup();
    render(<AccessoriesView roles={["Viewer"]} />);
    await screen.findByText("Gerades Modellgleis");

    await user.click(screen.getByRole("button", { name: "Gebundene Menge filtern" }));
    await waitFor(() => expect(api.accessoryArticles).toHaveBeenLastCalledWith(expect.objectContaining({
      statuses: ["reserved", "installed"]
    })));
    expect(screen.getByRole("button", { name: "Status" })).toHaveTextContent("Gebunden");

    await user.click(screen.getByRole("button", { name: "Filter zurücksetzen" }));
    expect(screen.getByRole("button", { name: "Status" })).toHaveTextContent("Alle Status");
  });

  it("keeps care hints informational and maps maintenance due only from the status filter", async () => {
    const user = userEvent.setup();
    render(<AccessoriesView roles={["Viewer"]} />);
    await screen.findByText("Gerades Modellgleis");
    const initialCalls = vi.mocked(api.accessoryArticles).mock.calls.length;

    await user.click(screen.getByText("3"));
    expect(api.accessoryArticles).toHaveBeenCalledTimes(initialCalls);
    expect(screen.queryByRole("button", { name: "Pflegehinweise filtern" })).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Status" }));
    await user.click(screen.getByRole("option", { name: "Wartung fällig" }));
    await waitFor(() => expect(api.accessoryArticles).toHaveBeenLastCalledWith(expect.objectContaining({
      statuses: ["maintenance_due"]
    })));
    expect(screen.getByRole("button", { name: "Status" })).toHaveTextContent("Wartung fällig");
  });

  it("shows loading, error, no-article, and no-result states", async () => {
    const pending = deferred<AccessoryArticleListResult>();
    vi.mocked(api.accessoryArticles).mockReturnValueOnce(pending.promise);
    const { unmount } = render(<AccessoriesView roles={["Editor"]} />);
    expect(screen.getByText("Artikel werden geladen …")).toBeInTheDocument();
    unmount();

    vi.mocked(api.accessoryArticles).mockRejectedValueOnce(new Error("Artikel nicht erreichbar"));
    const errorView = render(<AccessoriesView roles={["Editor"]} onCreateArticle={vi.fn()} />);
    expect(await screen.findByRole("alert")).toHaveTextContent("Artikel nicht erreichbar");
    expect(screen.queryByText("Noch keine Artikel vorhanden.")).not.toBeInTheDocument();
    expect(screen.queryByText("Keine Artikel entsprechen den aktiven Filtern.")).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Ersten Artikel anlegen" })).not.toBeInTheDocument();
    errorView.unmount();

    vi.mocked(api.accessoryArticles).mockResolvedValueOnce({ ...overview, items: [], metrics: {
      ...overview.metrics, articleCount: 0
    } });
    const emptyView = render(<AccessoriesView roles={["Editor"]} onCreateArticle={vi.fn()} />);
    expect(await screen.findByText("Noch keine Artikel vorhanden.")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Ersten Artikel anlegen" })).toBeInTheDocument();
    emptyView.unmount();

    vi.mocked(api.accessoryArticles).mockResolvedValue({ ...overview, items: [] });
    const user = userEvent.setup();
    render(<AccessoriesView roles={["Viewer"]} />);
    await screen.findByText("Keine Artikel entsprechen den aktiven Filtern.");
    await user.type(screen.getByRole("searchbox", { name: "Artikel suchen" }), "unbekannt");
    expect(screen.getByText("Keine Artikel entsprechen den aktiven Filtern.")).toBeInTheDocument();
  });

  it("exposes mutation seams only to writers and explains read-only access", async () => {
    const onOpenArticle = vi.fn();
    const onCreateArticle = vi.fn();
    const user = userEvent.setup();
    const editorView = render(
      <AccessoriesView roles={["Editor"]} onOpenArticle={onOpenArticle} onCreateArticle={onCreateArticle} />
    );
    await screen.findByText("Gerades Modellgleis");
    await user.click(screen.getByRole("button", { name: "Neuer Artikel" }));
    await user.click(screen.getByRole("button", { name: "Artikel bearbeiten: Gerades Modellgleis" }));
    expect(onCreateArticle).toHaveBeenCalledOnce();
    expect(onOpenArticle).toHaveBeenCalledWith("article-1", "edit");
    editorView.unmount();

    const plannerView = render(<AccessoriesView roles={["Planner"]} />);
    await screen.findByText("Gerades Modellgleis");
    expect(screen.queryByRole("button", { name: "Neuer Artikel" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /bearbeiten/i })).not.toBeInTheDocument();
    expect(screen.getByText("Planungszugriff: Sie können Artikel ansehen sowie Reservierungen anlegen und stornieren.")).toBeInTheDocument();
    plannerView.unmount();

    render(<AccessoriesView roles={["Viewer"]} />);
    await screen.findByText("Gerades Modellgleis");
    expect(screen.queryByRole("button", { name: "Neuer Artikel" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /bearbeiten/i })).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Artikel ansehen: Gerades Modellgleis" })).toBeInTheDocument();
    expect(screen.getByText("Schreibgeschützter Zugriff: Sie können Artikel ansehen, aber nicht ändern."))
      .toBeInTheDocument();
  });

  it("wires create, view, edit, and article-name actions to the shared Task 11 controller", async () => {
    const user = userEvent.setup();
    render(<AccessoriesView roles={["Editor"]} />);
    await screen.findByText("Gerades Modellgleis");

    expect(screen.getByRole("button", { name: "Neuer Artikel" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Artikel ansehen/ })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Artikel bearbeiten/ })).toBeInTheDocument();
    expect(screen.getByText("Gerades Modellgleis").closest("button")).not.toBeNull();
    expect(screen.getByRole("button", { name: /Weitere Aktionen/ })).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Neuer Artikel" }));
    expect(screen.getByRole("dialog", { name: "Artikel anlegen" })).toBeInTheDocument();
  });

  it("renders no article workspace and performs no request for Messe", () => {
    render(<AccessoriesView roles={["Messe"]} />);
    expect(screen.getByText("Kein Zugriff auf Zubehör.")).toBeInTheDocument();
    expect(api.accessoryArticles).not.toHaveBeenCalled();
  });

  it("runs Tillig 83101 through purchase, individualization, reservation, installation, removal, and history", async () => {
    const timestamp = "2026-08-08T10:00:00Z";
    const location: StorageLocation = {
      id: "location-1", name: "Werkstatt", archived: false, createdAt: timestamp, updatedAt: timestamp
    };
    const layout: Layout = {
      id: "layout-1", name: "Testanlage", kind: "private", gauge: "TT", scale: "1:120", version: 1,
      archived: false, createdAt: timestamp, updatedAt: timestamp
    };
    let article: AccessoryArticle | null = null;
    let stockQuantity = 0;
    let assets: AccessoryAsset[] = [];
    let purchases: AccessoryPurchase[] = [];
    let reservations: AccessoryReservation[] = [];
    let installations: AccessoryInstallation[] = [];
    let events: AccessoryUsageEvent[] = [];
    const currentOverview = (): AccessoryArticleListResult => article ? {
      items: [{
        id: article.id, inventoryNumber: article.inventoryNumber, manufacturer: article.manufacturer,
        articleNumber: article.articleNumber || "",
        name: article.name, articleType: article.articleType, subtype: article.subtype, gauges: article.gauges,
        inventoryStrategy: article.inventoryStrategy, archived: false,
        owned: stockQuantity + assets.length, available: stockQuantity,
        reserved: reservations.filter((item) => item.status === "active").length,
        installed: installations.filter((item) => !item.removedAt).length,
        locationNames: [location.name], hasUsageHistory: events.length > 0, careHintCount: 0,
        updatedAt: timestamp, attributes: article.attributes
      }],
      metrics: {
        articleCount: 1, articleTypeCount: 1, available: stockQuantity, locationCount: 1,
        reserved: reservations.filter((item) => item.status === "active").length,
        installed: installations.filter((item) => !item.removedAt).length, careHintCount: 0
      },
      filters: { manufacturers: [article.manufacturer], articleTypes: [article.articleType],
        gauges: article.gauges, storageLocations: [{ id: location.id, name: location.name }] }
    } : {
      items: [], metrics: { articleCount: 0, articleTypeCount: 0, available: 0, locationCount: 0,
        reserved: 0, installed: 0, careHintCount: 0 },
      filters: { manufacturers: [], articleTypes: [], gauges: [], storageLocations: [] }
    };

    vi.mocked(api.accessoryArticles).mockImplementation(async () => currentOverview());
    vi.mocked(api.storageLocations).mockResolvedValue([location]);
    vi.spyOn(api, "masterData").mockImplementation(async (type) =>
      type === "accessory_subtype" ? [straightSubtype]
        : type === "article_type" ? standardArticleTypes
          : type === "manufacturer" ? [manufacturerEntry]
            : type === "gauge" ? [gaugeEntry]
              : type === "stock_unit" ? [stockUnitEntry] : []);
    vi.spyOn(api, "checkAccessoryArticleDuplicates").mockResolvedValue({ candidates: [] });
    vi.spyOn(api, "createAccessoryArticle").mockImplementation(async (input) => {
      article = {
        id: "article-83101", inventoryNumber: "RK-ART-000001", manufacturer: input.manufacturer,
        articleNumber: input.articleNumber,
        name: input.name, category: input.subtype, trackingMode: "quantity", manufacturerStatus: "available",
        articleType: input.articleType, subtype: input.subtype, gauges: input.gauges || [],
        packageQuantity: input.packageQuantity, stockUnit: input.stockUnit, minimumStock: input.minimumStock || 0,
        inventoryStrategy: input.inventoryStrategy, alternativeNumbers: [], keywords: [], archived: false,
        attributes: input.attributes || [], createdAt: timestamp, updatedAt: timestamp
      };
      return article;
    });
    vi.spyOn(api, "accessoryArticle").mockImplementation(async () => {
      if (!article) throw new Error("article missing");
      return article;
    });
    vi.spyOn(api, "vehicles").mockResolvedValue([]);
    vi.spyOn(api, "layouts").mockResolvedValue([layout]);
    vi.spyOn(api, "layoutUnits").mockResolvedValue([]);
    vi.spyOn(api, "accessoryStock").mockImplementation(async () => ({
      productId: "article-83101", trackingMode: "quantity", totalQuantity: stockQuantity,
      locations: [{ locationId: location.id, locationName: location.name, quantity: stockQuantity,
        updatedAt: timestamp }]
    }));
    vi.spyOn(api, "accessoryStockMovements").mockResolvedValue([]);
    vi.spyOn(api, "accessoryAssets").mockImplementation(async () => assets);
    vi.spyOn(api, "accessoryPurchases").mockImplementation(async () => purchases);
    vi.spyOn(api, "accessoryDocuments").mockResolvedValue([]);
    vi.spyOn(api, "accessoryReservations").mockImplementation(async () => reservations);
    vi.spyOn(api, "accessoryInstallations").mockImplementation(async () => installations);
    vi.spyOn(api, "accessoryUsageHistory").mockImplementation(async () => ({
      productId: "article-83101", events
    }));
    vi.spyOn(api, "createAccessoryPurchase").mockImplementation(async (productId, input) => {
      const purchase: AccessoryPurchase = {
        id: "purchase-1", productId, storageLocationId: input.storageLocationId, quantity: input.quantity,
        purchasedAt: input.purchasedAt, supplier: input.supplier, currency: input.currency,
        bookToStock: Boolean(input.bookToStock), createdAt: timestamp, updatedAt: timestamp
      };
      purchases = [purchase];
      if (input.bookToStock) stockQuantity += input.quantity;
      return purchase;
    });
    const individualize = vi.spyOn(api, "individualizeAccessoryProduct").mockImplementation(async (productId, input) => {
      stockQuantity -= 1;
      const asset: AccessoryAsset = {
        id: "asset-1", productId, inventoryNumber: input.asset.inventoryNumber, condition: "ready",
        lifecycle: "stored", storageLocationId: input.locationId, createdAt: timestamp, updatedAt: timestamp
      };
      assets = [asset];
      return asset;
    });
    const reserve = vi.spyOn(api, "createAccessoryReservation").mockImplementation(async (input) => {
      const reservation: AccessoryReservation = {
        ...input, id: "reservation-1", status: "active", createdBy: "editor", createdAt: timestamp,
        updatedAt: timestamp
      };
      reservations = [reservation];
      events = [{ id: "event-reservation", productId: input.productId, reservationId: reservation.id,
        assetId: input.assetId, locationId: input.locationId, layoutId: layout.id, quantity: 1,
        type: "reservation", occurredAt: timestamp }];
      return reservation;
    });
    vi.spyOn(api, "createAccessoryInstallation").mockImplementation(async (input) => {
      const installation: AccessoryInstallation = {
        ...input, id: "installation-1", condition: "ready", installedBy: "editor", installedAt: timestamp
      };
      installations = [installation];
      reservations = reservations.map((item) => ({ ...item, status: "fulfilled" }));
      assets = assets.map((item) => ({ ...item, lifecycle: "installed" }));
      events = [...events, { id: "event-installation", productId: input.productId,
        installationId: installation.id, assetId: input.assetId, layoutId: layout.id, quantity: 1,
        type: "installation", condition: "ready", occurredAt: timestamp }];
      return installation;
    });
    vi.spyOn(api, "removeAccessoryInstallation").mockImplementation(async (id, input) => {
      const current = installations.find((item) => item.id === id);
      if (!current) throw new Error("installation missing");
      const removed: AccessoryInstallation = {
        ...current, removedAt: timestamp, removedBy: "editor", removalDisposition: input.disposition
      };
      installations = [removed];
      assets = assets.map((item) => ({ ...item, lifecycle: "stored", storageLocationId: location.id }));
      events = [...events, { id: "event-removal", productId: current.productId,
        installationId: current.id, assetId: current.assetId, layoutId: layout.id, quantity: 1,
        type: "removal", removalDisposition: input.disposition, occurredAt: timestamp }];
      return removed;
    });

    const user = userEvent.setup();
    render(<AccessoriesView roles={["Editor"]} />);
    await user.click(await screen.findByRole("button", { name: "Neuer Artikel" }));
    const createDialog = screen.getByRole("dialog", { name: "Artikel anlegen" });
    await user.click(within(createDialog).getByRole("button", { name: "Hersteller" }));
    await user.click(screen.getByRole("option", { name: "Tillig" }));
    await user.type(within(createDialog).getByRole("textbox", { name: "Artikelnummer" }), "83101");
    await user.type(within(createDialog).getByRole("textbox", { name: "Bezeichnung" }), "TT Modellgleis");
    await user.click(within(createDialog).getByRole("button", { name: "Artikelart" }));
    await user.click(screen.getByRole("option", { name: "Gleis" }));
    await user.click(within(createDialog).getByRole("button", { name: "Unterart" }));
    await user.click(screen.getByRole("option", { name: "Gerade" }));
    await user.click(within(createDialog).getByRole("button", { name: /Spurweite/ }));
    await user.click(screen.getByRole("option", { name: "TT" }));
    await user.click(screen.getByRole("tab", { name: "Bestand" }));
    await user.click(screen.getByRole("button", { name: "Bestandsstrategie" }));
    await user.click(screen.getByRole("option", { name: "Menge mit späterer Individualisierung" }));
    await user.click(screen.getByRole("tab", { name: "Fachangaben: Gleis" }));
    await user.type(screen.getByRole("textbox", { name: "Gleissystem" }), "Tillig TT Modellgleis");
    await user.type(screen.getByRole("spinbutton", { name: "Länge (mm)" }), "166");
    await user.click(screen.getByRole("button", { name: "Artikel anlegen" }));
    await waitFor(() => expect(screen.queryByRole("dialog", { name: "Artikel anlegen" })).not.toBeInTheDocument());

    await user.click(await screen.findByRole("button", { name: "Artikel bearbeiten: TT Modellgleis" }));
    await screen.findByRole("dialog", { name: "Artikel bearbeiten" });
    await user.click(screen.getByRole("tab", { name: "Kauf & Dokumente" }));
    await user.clear(screen.getByRole("spinbutton", { name: "Menge" }));
    await user.type(screen.getByRole("spinbutton", { name: "Menge" }), "2");
    await user.click(screen.getByRole("checkbox", { name: "Kauf bestandswirksam buchen" }));
    await user.click(screen.getByRole("button", { name: "Kauf buchen" }));
    await waitFor(() => expect(stockQuantity).toBe(2));

    await user.click(screen.getByRole("tab", { name: "Bestand" }));
    await user.type(screen.getByRole("textbox", { name: "Inventarnummer" }), "RK-83101-001");
    await user.click(screen.getByRole("button", { name: "Einzelstück speichern" }));
    await user.click(within(screen.getByRole("dialog", { name: "Einzelstück bestätigen" }))
      .getByRole("button", { name: "Bestätigen" }));
    await waitFor(() => expect(individualize).toHaveBeenCalledWith("article-83101", expect.objectContaining({
      locationId: location.id, asset: expect.objectContaining({ inventoryNumber: "RK-83101-001" })
    })));

    await user.click(screen.getAllByRole("button", { name: "Bestandsquelle" })[0]!);
    await user.click(screen.getByRole("option", { name: "Einzelstück" }));
    expect(screen.getAllByRole("button", { name: "Einzelstück" })[0]).toHaveTextContent("RK-83101-001");
    await user.click(screen.getByRole("button", { name: "Reservierung anlegen" }));
    await user.click(within(screen.getByRole("dialog", { name: "Reservierung bestätigen" }))
      .getByRole("button", { name: "Bestätigen" }));
    await waitFor(() => expect(reserve).toHaveBeenCalledWith(expect.objectContaining({
      productId: "article-83101", assetId: "asset-1", layoutId: layout.id, quantity: 1
    })));

    expect(screen.getByRole("tab", { name: "Verwendung & Historie" })).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Reservierung" }));
    await user.click(screen.getByRole("option", { name: "Testanlage" }));
    await user.click(screen.getByRole("button", { name: "Einbau erfassen" }));
    await user.click(within(screen.getByRole("dialog", { name: "Einbau bestätigen" }))
      .getByRole("button", { name: "Bestätigen" }));
    await screen.findByRole("button", { name: "Ausbauen" });
    await user.click(screen.getByRole("button", { name: "Ausbauen" }));
    await user.click(screen.getAllByRole("button", { name: "Ausbauen" }).at(-1)!);
    await user.click(within(screen.getByRole("dialog", { name: "Ausbau bestätigen" }))
      .getByRole("button", { name: "Bestätigen" }));
    await user.click(screen.getByRole("tab", { name: "Verwendung & Historie" }));
    expect(await screen.findByText("Ausbau")).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: "Verwendung & Historie" })).toBeInTheDocument();
  }, 15_000);
});
