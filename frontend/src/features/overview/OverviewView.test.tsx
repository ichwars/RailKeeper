import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { act, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api, type AccessoryArticleListResult } from "../../shared/api";
import { setLanguage } from "../../shared/i18n";
import { vehicleFixture } from "../../test/fixtures/vehicles";
import { OverviewView } from "./OverviewView";
import { overviewMetricLimitForWidth, overviewMetricProfileKey } from "./overviewModel";

const accessories: AccessoryArticleListResult = {
  items: [],
  metrics: {
    articleCount: 7,
    articleTypeCount: 2,
    available: 4,
    locationCount: 1,
    reserved: 2,
    installed: 1,
    careHintCount: 0
  },
  filters: { manufacturers: [], articleTypes: [], gauges: [], storageLocations: [] }
};

describe("OverviewView", () => {
  beforeEach(() => {
    window.localStorage.clear();
    setLanguage("de");
    vi.spyOn(HTMLCanvasElement.prototype, "getContext").mockImplementation(() => null);
    vi.spyOn(api, "vehicles").mockResolvedValue([vehicleFixture()]);
    vi.spyOn(api, "accessoryArticles").mockResolvedValue(accessories);
    vi.spyOn(api, "overviewValuation").mockResolvedValue({
      vehicleListValue: "100.00",
      vehiclePurchaseValue: "80.00",
      accessoryListValue: "50.00",
      accessoryPurchaseCost: "40.00",
      excludedForeignCurrencyPurchases: 0
    });
    vi.spyOn(api, "profileSettings").mockResolvedValue({ settings: {} });
    vi.spyOn(api, "updateProfileSettings").mockResolvedValue({ settings: {} });
  });

  it("renders the reference topology with real overview values", async () => {
    render(<OverviewView username="anna" />);

    expect(await screen.findByText("1 digital · 0 analog")).toBeInTheDocument();
    expect(screen.getByText("2 reserviert")).toBeInTheDocument();
    for (const heading of [
      "Jetzt wichtig",
      "Digitalzentralen",
      "Zuletzt bearbeitet",
      "Nächste Wartungen",
      "Bestandsentwicklung",
      "Wertaufteilung",
      "Bestandsstruktur",
      "Hersteller",
      "Datenqualität"
    ]) {
      expect(screen.getByRole("heading", { name: heading })).toBeInTheDocument();
    }
    expect(screen.getByText("150,00 €")).toBeInTheDocument();
    expect(screen.getByText("Fahrzeug anlegen")).toBeInTheDocument();
  });

  it("opens the dedicated Digital Centers workspace from the overview card", async () => {
    const disabledProvider = { enabled: false, host: "", port: "" };
    vi.spyOn(api, "digitalSettings").mockResolvedValue({
      provider: "ecos",
      ecos: disabledProvider,
      z21: disabledProvider,
      intellibox3: disabledProvider,
      cs3: disabledProvider
    });

    render(<OverviewView roles={["Admin"]} />);

    expect(await screen.findByRole("link", { name: "Öffnen" }))
      .toHaveAttribute("href", "/digital-centers");
  });

  it.each([
    ["z21", "Roco Z21", "/assets/overview-z21-10820.png"],
    ["intellibox3", "Uhlenbrock Intellibox 3", "/assets/overview-intellibox-3-65300.jpg"],
    ["cs3", "Märklin CS3", "/assets/overview-cs3-60226.jpg"]
  ] as const)("renders the configured %s product image", async (provider, label, expectedSource) => {
    const disabledProvider = { enabled: false, host: "", port: "" };
    vi.spyOn(api, "digitalSettings").mockResolvedValue({
      provider,
      ecos: disabledProvider,
      z21: disabledProvider,
      intellibox3: disabledProvider,
      cs3: disabledProvider
    });

    const { container } = render(<OverviewView roles={["Admin"]} />);

    expect(await screen.findByText(label)).toBeInTheDocument();
    expect(container.querySelector<HTMLImageElement>(".overview-device-art img")?.getAttribute("src"))
      .toBe(expectedSource);
  });

  it("keeps inventory modules available when valuation loading fails", async () => {
    vi.mocked(api.overviewValuation).mockRejectedValueOnce(new Error("Wertdienst offline"));
    render(<OverviewView />);

    expect(await screen.findByRole("alert")).toHaveTextContent("Bestandswerte konnten nicht geladen werden.");
    expect(screen.getByText("1 digital · 0 analog")).toBeInTheDocument();
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

    expect(await screen.findByRole("heading", { name: "Value distribution" })).toBeInTheDocument();
    await act(async () => rejectFirstRequest?.(new Error("Alter Fehler")));

    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    expect(screen.getByText("€100.00")).toBeInTheDocument();
  });

  it("does not refetch language-independent inventory after the language changes", async () => {
    render(<OverviewView />);
    await waitFor(() => expect(api.vehicles).toHaveBeenCalledOnce());

    act(() => setLanguage("en"));

    await screen.findByRole("heading", { name: "Value distribution" });
    expect(api.vehicles).toHaveBeenCalledOnce();
    expect(api.accessoryArticles).toHaveBeenCalledOnce();
    expect(api.overviewValuation).toHaveBeenCalledTimes(2);
  });

  it("previews metric changes immediately and persists them only after finishing", async () => {
    const user = userEvent.setup();
    render(<OverviewView username="anna" />);
    await screen.findByText("1 digital · 0 analog");

    await user.click(screen.getByRole("button", { name: "Kacheln anpassen" }));
    const metricRegion = screen.getByRole("region", { name: "Konfigurierbare Kennzahlen" });
    const checkboxes = screen.getAllByRole("checkbox");
    expect(checkboxes).toHaveLength(10);
    expect(checkboxes.filter((checkbox) => (checkbox as HTMLInputElement).checked)).toHaveLength(4);
    expect(checkboxes.filter((checkbox) => (checkbox as HTMLInputElement).disabled)).toHaveLength(6);

    await user.click(screen.getByRole("checkbox", { name: "Wartung" }));
    expect(within(metricRegion).queryByText("Wartung")).not.toBeInTheDocument();
    await user.click(screen.getByRole("checkbox", { name: "Digitalisiert" }));
    expect(within(metricRegion).getByText("Digitalisiert")).toBeInTheDocument();
    expect(api.updateProfileSettings).not.toHaveBeenCalled();
    await user.click(screen.getByRole("button", { name: "Fertig" }));

    expect(api.updateProfileSettings).toHaveBeenCalledWith(expect.objectContaining({
      [overviewMetricProfileKey]: expect.stringContaining("digitalized")
    }));
    expect(window.localStorage.getItem(`${overviewMetricProfileKey}:anna`)).toContain("digitalized");
  });

  it("retries a transient profile synchronization failure before showing an error", async () => {
    vi.mocked(api.updateProfileSettings)
      .mockRejectedValueOnce(new Error("connection interrupted"))
      .mockResolvedValueOnce({ settings: {} });
    const user = userEvent.setup();
    render(<OverviewView username="anna" />);
    await screen.findByText("1 digital · 0 analog");

    await user.click(screen.getByRole("button", { name: "Kacheln anpassen" }));
    await user.click(screen.getByRole("button", { name: "Fertig" }));

    await waitFor(() => expect(api.updateProfileSettings).toHaveBeenCalledTimes(2));
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("toggles the metric dialog and discards an unfinished preview", async () => {
    const user = userEvent.setup();
    render(<OverviewView username="anna" />);
    await screen.findByText("1 digital · 0 analog");

    const trigger = screen.getByRole("button", { name: "Kacheln anpassen" });
    await user.click(trigger);
    const metricRegion = screen.getByRole("region", { name: "Konfigurierbare Kennzahlen" });
    await user.click(screen.getByRole("checkbox", { name: "Wartung" }));
    expect(within(metricRegion).queryByText("Wartung")).not.toBeInTheDocument();

    await user.click(trigger);
    expect(screen.queryByRole("dialog", { name: "Kacheln anpassen" })).not.toBeInTheDocument();
    expect(within(metricRegion).getByText("Wartung")).toBeInTheDocument();
    expect(api.updateProfileSettings).not.toHaveBeenCalled();
  });

  it("resets the metric preview immediately without persisting before finishing", async () => {
    window.localStorage.setItem(`${overviewMetricProfileKey}:anna`, JSON.stringify({
      active: ["vehicles", "digitalized"],
      order: ["digitalized", "vehicles"]
    }));
    const user = userEvent.setup();
    render(<OverviewView username="anna" />);
    await screen.findByText("100% des Fahrzeugbestands");

    await user.click(screen.getByRole("button", { name: "Kacheln anpassen" }));
    await user.click(screen.getByRole("button", { name: "Zurücksetzen" }));

    const metricRegion = screen.getByRole("region", { name: "Konfigurierbare Kennzahlen" });
    expect(within(metricRegion).getByText("Wartung")).toBeInTheDocument();
    expect(within(metricRegion).queryByText("Digitalisiert")).not.toBeInTheDocument();
    expect(api.updateProfileSettings).not.toHaveBeenCalled();
  });

  it("adapts the metric capacity to the available desktop width", () => {
    expect(overviewMetricLimitForWidth(900)).toBe(3);
    expect(overviewMetricLimitForWidth(1320)).toBe(5);
    expect(overviewMetricLimitForWidth(1800)).toBe(6);
  });

  it("uses the full workspace width and keeps the GitHub report action available", () => {
    const css = readFileSync(resolve(process.cwd(), "src/styles/overview-dashboard.css"), "utf8");

    expect(css).toMatch(/\.overview-page\s*\{[^}]*max-width:\s*none/s);
    expect(css).not.toMatch(/\.layout:has\(\.overview-page\) \.feedback-button\s*\{[^}]*display:\s*none/s);
    expect(css).toMatch(/\.layout:has\(\.overview-page\) \.feedback-button\s*\{[^}]*bottom:\s*104px/s);
    expect(css).toContain("grid-template-columns: minmax(0, 3fr) minmax(360px, 2fr)");
    expect(css).toContain("grid-template-columns: minmax(0, 1.35fr) minmax(360px, 0.85fr)");
    expect(css).toContain("conic-gradient(var(--accent-strong)");
    expect(css).toContain("position: fixed");
    expect(css).toContain("left: 216px");
    expect(css).toContain(".sidebar-collapsed .overview-action-footer");
    expect(css).toContain("@container overview (max-width: 1120px)");
    expect(css).toContain("@container overview (max-width: 860px)");
    expect(css).toContain("@media (max-width: 760px)");
  });

  it("defines readable hierarchy and separate semantic status colors", () => {
    const css = readFileSync(resolve(process.cwd(), "src/styles/overview-dashboard.css"), "utf8");
    const baseCss = readFileSync(resolve(process.cwd(), "src/styles/base.css"), "utf8");

    expect(baseCss).toMatch(/--font-size-micro:\s*11px/);
    expect(baseCss).toMatch(/--font-size-caption:\s*12px/);
    expect(baseCss).toMatch(/--font-size-3xs:\s*var\(--font-size-micro\)/);
    expect(baseCss).toMatch(/--font-size-2xs:\s*var\(--font-size-caption\)/);
    expect(baseCss).toMatch(/--font-size-xs:\s*var\(--font-size-caption\)/);
    expect(baseCss).toMatch(/--font-size-sm:\s*13px/);
    expect(baseCss).toMatch(/--font-size-base:\s*15px/);
    expect(css).toContain("--overview-interactive:");
    expect(css).toContain("--overview-info:");
    expect(css).toContain("--overview-warning:");
    expect(css).toContain("--overview-danger:");
    expect(css).toMatch(/\.overview-card-head h2,[\s\S]*font-size:\s*var\(--font-size-md\)/);
    expect(css).toMatch(/\.overview-priority-card\s*\{[^}]*border-color:/s);
    expect(css).toMatch(/\.overview-maintenance-card:has\(\.overdue\)\s*\{[^}]*border-color:/s);
    expect(css).toMatch(/\.overview-analysis-primary \.overview-card,[\s\S]*background:/s);
  });
});
