import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { act, render, screen, waitFor } from "@testing-library/react";
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

  it("ignores a superseded valuation failure after the language changes", async () => {
    let rejectFirstRequest: ((reason?: unknown) => void) | undefined;
    vi.mocked(api.overviewValuation)
      .mockImplementationOnce(() => new Promise((_resolve, reject) => {
        rejectFirstRequest = reject;
      }))
      .mockResolvedValueOnce({
        vehicleListValue: "100.00",
        vehiclePurchaseValue: "80.00",
        accessoryListValue: "50.00",
        accessoryPurchaseCost: "40.00",
        excludedForeignCurrencyPurchases: 0
      });

    render(<OverviewView />);
    await waitFor(() => expect(api.overviewValuation).toHaveBeenCalledOnce());
    act(() => setLanguage("en"));

    expect(await screen.findByRole("heading", { name: "Recorded inventory values" }))
      .toBeInTheDocument();
    await act(async () => rejectFirstRequest?.(new Error("Alter Fehler")));

    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    expect(screen.getByText("€100.00")).toBeInTheDocument();
  });

  it("defines responsive two-column and one-column valuation layouts without overflow", () => {
    const responsive = readFileSync(resolve(process.cwd(), "src/styles/overrides-responsive.css"), "utf8");
    const overview = readFileSync(resolve(process.cwd(), "src/styles/overview.css"), "utf8");

    expect(responsive).toContain(".overview-valuation-grid");
    expect(responsive).toContain("grid-template-columns: repeat(2, minmax(0, 1fr))");
    expect(responsive).toContain("grid-template-columns: minmax(0, 1fr)");
    expect(overview).toContain("grid-template-columns: repeat(5, minmax(0, 1fr))");
    expect(overview).toContain(".overview-valuation-card");
    expect(overview).toContain("grid-column: span 2");
    expect(overview).toContain("min-width: 0");
    expect(overview).toContain("overflow-wrap: anywhere");
  });
});
