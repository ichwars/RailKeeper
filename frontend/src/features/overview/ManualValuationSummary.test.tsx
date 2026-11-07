import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { api, type VehicleValuationSummary } from "../../shared/api";
import { setLanguage } from "../../shared/i18n";
import { ManualValuationSummary } from "./ManualValuationSummary";

const summary: VehicleValuationSummary = {  
asOf: "2026-01-01", vehicleCount: 4, valuedVehicles: 2, missingVehicles: 2,
  totals: [{ currency: "CHF", amount: "0.123456", vehicles: 1 }, { currency: "EUR", amount: "999999999999.999999", vehicles: 1 }]
};

describe("manual collection summary", () => {
  beforeEach(() => { setLanguage("de"); vi.spyOn(api, "valuationSummary").mockResolvedValue(summary); });
  it("retains exact currency totals and makes missing coverage explicit", async () => {
    render(<ManualValuationSummary />);
    expect(await screen.findByText("0.123456")).toBeInTheDocument();
    expect(screen.getByText("999999999999.999999")).toBeInTheDocument();
    expect(screen.getByText("2 von 4 Fahrzeugen bewertet, 2 ohne Bewertung zum Stichtag.")).toBeInTheDocument();
    expect(screen.getByText(/keine Umrechnung/)).toBeInTheDocument();
  });
  it("does not replace the selected date with a late earlier response", async () => {
    let finish!: (value: VehicleValuationSummary) => void;
    vi.mocked(api.valuationSummary).mockReturnValueOnce(new Promise(resolve => { finish = resolve; }));
    render(<ManualValuationSummary />);
    fireEvent.change(screen.getByLabelText("Stichtag"), { target: { value: "2020-01-01" } });
    await screen.findByText("0.123456");
    await act(async () => finish({ ...summary, totals: [{ currency: "USD", amount: "1.00", vehicles: 2 }] }));
    expect(screen.queryByText("USD")).not.toBeInTheDocument();
    expect(screen.getByText("CHF")).toBeInTheDocument();
    expect(api.valuationSummary).toHaveBeenLastCalledWith("2020-01-01");
  });
  it("does not fabricate values for an empty date or an unvalued collection", async () => {
    vi.mocked(api.valuationSummary).mockResolvedValue({ ...summary, valuedVehicles: 0, missingVehicles: 4, totals: [] });
    render(<ManualValuationSummary />);
    expect(await screen.findByText(/keine gültigen Bewertungen/)).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("Stichtag"), { target: { value: "" } });
    expect(screen.getByText("Wähle einen Stichtag.")).toBeInTheDocument();
    expect(api.valuationSummary).toHaveBeenCalledOnce();
  });
  it("shows a failed calculation and retries only after an explicit request", async () => {
    vi.mocked(api.valuationSummary).mockRejectedValueOnce(new Error("offline"));
    const user = userEvent.setup();
    render(<ManualValuationSummary />);
    expect(await screen.findByRole("alert")).toHaveTextContent("nicht geladen");
    expect(api.valuationSummary).toHaveBeenCalledOnce();
    await user.click(screen.getByRole("button", { name: "Neu laden" }));
    await waitFor(() => expect(screen.queryByRole("alert")).not.toBeInTheDocument());
    expect(await screen.findByText("0.123456")).toBeInTheDocument();
  });
});
