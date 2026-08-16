import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api } from "../../shared/api";
import { setLanguage } from "../../shared/i18n";
import { vehicleFixture } from "../../test/fixtures/vehicles";
import { OverviewView } from "./OverviewView";

describe("OverviewView valuation", () => {
  beforeEach(() => {
    setLanguage("de");
    vi.spyOn(api, "vehicles").mockResolvedValue([vehicleFixture()]);
    vi.spyOn(api, "overviewValuation").mockResolvedValue({
      vehicleListValue: "100.00",
      vehiclePurchaseValue: "80.00",
      accessoryListValue: "50.00",
      accessoryPurchaseCost: "40.00",
      excludedForeignCurrencyPurchases: 0
    });
  });

  it("replaces the old list-value KPI with the detailed valuation card", async () => {
    render(<OverviewView />);

    expect(await screen.findByRole("heading", { name: "Erfasste Bestandswerte" })).toBeInTheDocument();
    expect(screen.queryByText("Erfasster Listenwert")).not.toBeInTheDocument();
  });

  it("keeps vehicle KPIs available when only valuation loading fails", async () => {
    vi.mocked(api.overviewValuation).mockRejectedValueOnce(new Error("Wertdienst offline"));
    render(<OverviewView />);

    expect(await screen.findByRole("alert")).toHaveTextContent("Bestandswerte konnten nicht geladen werden.");
    expect(screen.getByText("Gesamtbestand")).toBeInTheDocument();
    expect(screen.queryByText("Wertdienst offline")).not.toBeInTheDocument();
  });

  it("defines responsive two-column and one-column valuation layouts without overflow", () => {
    const responsive = readFileSync(resolve(process.cwd(), "src/styles/overrides-responsive.css"), "utf8");
    const overview = readFileSync(resolve(process.cwd(), "src/styles/overview.css"), "utf8");

    expect(responsive).toContain(".overview-valuation-grid");
    expect(responsive).toContain("grid-template-columns: repeat(2, minmax(0, 1fr))");
    expect(responsive).toContain("grid-template-columns: minmax(0, 1fr)");
    expect(overview).toContain("min-width: 0");
    expect(overview).toContain("overflow-wrap: anywhere");
  });
});
