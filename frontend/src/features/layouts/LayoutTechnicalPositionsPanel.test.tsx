import { act, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import {
  api,
  type AccessoryArticleListResult,
  type LayoutTechnicalPosition,
  type LayoutUnit
} from "../../shared/api";
import { LayoutTechnicalPositionsPanel } from "./LayoutTechnicalPositionsPanel";

const units: LayoutUnit[] = [{
  id: "unit-1", layoutId: "layout-1", name: "Bahnhof", kind: "module", widthMm: 1000, heightMm: 400,
  version: 1, archived: false, createdAt: "2026-08-09T10:00:00Z", updatedAt: "2026-08-09T10:00:00Z"
}, {
  id: "unit-2", layoutId: "layout-1", name: "Schattenbahnhof", kind: "module", widthMm: 1200,
  heightMm: 500, version: 1, archived: false, createdAt: "2026-08-09T10:00:00Z",
  updatedAt: "2026-08-09T10:00:00Z"
}];

const positions: LayoutTechnicalPosition[] = units.map((unit, index) => ({
  id: `position-${index + 1}`, layoutUnitId: unit.id, label: `${unit.name} Signal`, kind: "signal",
  positionXMm: 100, positionYMm: 80, rotationDegrees: 0, version: 1, archived: false,
  createdAt: "2026-08-09T10:00:00Z", updatedAt: "2026-08-09T10:00:00Z"
}));

const emptyArticles: AccessoryArticleListResult = {
  items: [],
  metrics: {
    articleCount: 0, articleTypeCount: 0, available: 0, locationCount: 0,
    reserved: 0, installed: 0, careHintCount: 0
  },
  filters: { manufacturers: [], articleTypes: [], gauges: [], storageLocations: [] }
};

function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((next) => { resolve = next; });
  return { promise, resolve };
}

describe("LayoutTechnicalPositionsPanel", () => {
  afterEach(() => vi.restoreAllMocks());

  it("ignores a stale position response after switching units", async () => {
    const user = userEvent.setup();
    const first = deferred<LayoutTechnicalPosition[]>();
    const second = deferred<LayoutTechnicalPosition[]>();
    vi.spyOn(api, "accessoryArticles").mockResolvedValue(emptyArticles);
    const load = vi.spyOn(api, "layoutTechnicalPositions").mockImplementation((unitID) =>
      unitID === units[0].id ? first.promise : second.promise);
    render(<LayoutTechnicalPositionsPanel units={units} canPlan />);
    await waitFor(() => expect(load).toHaveBeenCalledWith(units[0].id));

    await user.click(screen.getByRole("button", { name: "Anlageneinheit" }));
    await user.click(screen.getByRole("option", { name: units[1].name }));
    await waitFor(() => expect(load).toHaveBeenCalledWith(units[1].id));
    await act(async () => second.resolve([positions[1]]));
    expect(await screen.findByText("Schattenbahnhof Signal")).toBeInTheDocument();

    await act(async () => first.resolve([positions[0]]));
    await waitFor(() => expect(screen.queryByText("Bahnhof Signal")).not.toBeInTheDocument());
    expect(screen.getByText("Schattenbahnhof Signal")).toBeInTheDocument();
  });
});
