import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it } from "vitest";

import type { OverviewValuation } from "../../shared/api";
import { setLanguage } from "../../shared/i18n";
import { formatOverviewMoney, OverviewValuationCard } from "./OverviewValuationCard";

const valuation: OverviewValuation = {
  vehicleListValue: "1299.90",
  vehiclePurchaseValue: "999.50",
  accessoryListValue: "250.00",
  accessoryPurchaseCost: "175.25",
  excludedForeignCurrencyPurchases: 0
};

describe("OverviewValuationCard", () => {
  beforeEach(() => setLanguage("de"));

  it("formats exact server decimals without floating point conversion", () => {
    expect(formatOverviewMoney("1299.90", "de")).toBe("1.299,90 €");
    expect(formatOverviewMoney("1299.90", "en")).toBe("€1,299.90");
  });

  it("renders all four labelled values in a semantic grid", () => {
    render(<OverviewValuationCard valuation={valuation} loading={false} error="" />);

    expect(screen.getByRole("heading", { name: "Erfasste Bestandswerte" })).toBeInTheDocument();
    for (const label of [
      "Fahrzeuge · Listenwert", "Fahrzeuge · Kaufpreis",
      "Zubehör · Listenwert", "Zubehör · Kaufkosten"
    ]) {
      expect(screen.getByText(label)).toBeInTheDocument();
    }
    for (const value of ["1.299,90 €", "999,50 €", "250,00 €", "175,25 €"]) {
      expect(screen.getByText(value)).toBeInTheDocument();
    }
  });

  it("shows local loading and failure states without invented zero values", () => {
    const view = render(<OverviewValuationCard valuation={null} loading error="" />);
    expect(screen.getByRole("status")).toBeInTheDocument();
    expect(screen.queryByText(/0,00 €/)).not.toBeInTheDocument();

    view.rerender(<OverviewValuationCard valuation={null} loading={false}
      error="Bestandswerte konnten nicht geladen werden." />);
    expect(screen.getByRole("alert")).toHaveTextContent("Bestandswerte konnten nicht geladen werden.");
  });

  it("only shows the foreign-currency hint for excluded purchases", () => {
    const view = render(<OverviewValuationCard valuation={valuation} loading={false} error="" />);
    expect(screen.queryByText(/Fremdwährung/)).not.toBeInTheDocument();

    view.rerender(<OverviewValuationCard valuation={{
      ...valuation, excludedForeignCurrencyPurchases: 2
    }} loading={false} error="" />);
    expect(screen.getByText("Nicht eingerechnete Einkäufe in Fremdwährung: 2.")).toBeInTheDocument();
  });
});
