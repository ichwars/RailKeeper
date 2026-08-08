import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api, type AccessoryArticleListResult } from "../../shared/api";
import { AccessoriesView } from "./AccessoriesView";

const overview: AccessoryArticleListResult = {
  items: [{
    id: "article-1",
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

function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((next) => { resolve = next; });
  return { promise, resolve };
}

describe("AccessoriesView", () => {
  beforeEach(() => {
    vi.spyOn(api, "accessoryArticles").mockResolvedValue(overview);
    vi.spyOn(api, "archiveAccessoryProduct").mockResolvedValue({} as never);
    vi.spyOn(api, "restoreAccessoryProduct").mockResolvedValue({} as never);
  });

  it("renders one table-only article overview with four global metrics", async () => {
    render(<AccessoriesView roles={["Editor"]} />);

    expect(await screen.findByRole("heading", { name: "Artikelübersicht" })).toBeInTheDocument();
    expect(screen.getByText("Modellbahnartikel suchen, erfassen und pflegen")).toBeInTheDocument();
    expect(screen.getAllByTestId("article-metric")).toHaveLength(4);
    expect(screen.getByText("24 Artikel · 5 Arten")).toBeInTheDocument();
    expect(screen.getByText("81 frei · 7 Lagerorte")).toBeInTheDocument();
    expect(screen.getByText("6 reserviert · 14 eingebaut")).toBeInTheDocument();
    expect(screen.getByText("3 unvollständig")).toBeInTheDocument();
    expect(screen.queryByRole("tablist")).not.toBeInTheDocument();
    expect(screen.queryByLabelText(/Kartenansicht/i)).not.toBeInTheDocument();
    expect(screen.getByText("1 Ergebnis")).toBeInTheDocument();
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
      sort: "article",
      direction: "asc"
    }));

    await user.click(screen.getByRole("button", { name: "Filter zurücksetzen" }));
    await waitFor(() => expect(api.accessoryArticles).toHaveBeenLastCalledWith({
      sort: "article",
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

  it("shows loading, error, no-article, and no-result states", async () => {
    const pending = deferred<AccessoryArticleListResult>();
    vi.mocked(api.accessoryArticles).mockReturnValueOnce(pending.promise);
    const { unmount } = render(<AccessoriesView roles={["Editor"]} />);
    expect(screen.getByText("Artikel werden geladen …")).toBeInTheDocument();
    unmount();

    vi.mocked(api.accessoryArticles).mockRejectedValueOnce(new Error("Artikel nicht erreichbar"));
    const errorView = render(<AccessoriesView roles={["Viewer"]} />);
    expect(await screen.findByRole("alert")).toHaveTextContent("Artikel nicht erreichbar");
    errorView.unmount();

    vi.mocked(api.accessoryArticles).mockResolvedValueOnce({ ...overview, items: [], metrics: {
      ...overview.metrics, articleCount: 0
    } });
    const emptyView = render(<AccessoriesView roles={["Editor"]} />);
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

    render(<AccessoriesView roles={["Planner"]} />);
    await screen.findByText("Gerades Modellgleis");
    expect(screen.queryByRole("button", { name: "Neuer Artikel" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /bearbeiten/i })).not.toBeInTheDocument();
    expect(screen.getByText("Schreibgeschützter Zugriff: Sie können Artikel ansehen, aber nicht ändern.")).toBeInTheDocument();
  });

  it("renders no article workspace and performs no request for Messe", () => {
    render(<AccessoriesView roles={["Messe"]} />);
    expect(screen.getByText("Kein Zugriff auf die Artikelübersicht.")).toBeInTheDocument();
    expect(api.accessoryArticles).not.toHaveBeenCalled();
  });
});
