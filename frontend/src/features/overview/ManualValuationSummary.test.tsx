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
    await userEvent.click(screen.getByRole("button", { name: "Manuell bewertete Sammlung" }));
    expect(await screen.findByText("0.123456")).toBeInTheDocument();
    expect(screen.getByText("999999999999.999999")).toBeInTheDocument();
    expect(screen.getByText("2 von 4 Fahrzeugen bewertet, 2 ohne Bewertung zum Stichtag.")).toBeInTheDocument();
    expect(screen.getByText(/keine Umrechnung/)).toBeInTheDocument();
  });
  it("does not replace the selected date with a late earlier response", async () => {
    let finish!: (value: VehicleValuationSummary) => void;
    vi.mocked(api.valuationSummary).mockReturnValueOnce(new Promise(resolve => { finish = resolve; }));
    render(<ManualValuationSummary />);
    await userEvent.click(screen.getByRole("button", { name: "Manuell bewertete Sammlung" }));
    fireEvent.change(screen.getByLabelText("Stichtag"), { target: { value: "2020-01-01" } });
    fireEvent.blur(screen.getByLabelText("Stichtag"));
    await screen.findByText("0.123456");
    await act(async () => finish({ ...summary, totals: [{ currency: "USD", amount: "1.00", vehicles: 2 }] }));
    expect(screen.queryByText("USD")).not.toBeInTheDocument();
    expect(screen.getByText("CHF")).toBeInTheDocument();
    expect(api.valuationSummary).toHaveBeenLastCalledWith("2020-01-01");
  });
  it("does not fabricate values for an empty date or an unvalued collection", async () => {
    vi.mocked(api.valuationSummary).mockResolvedValue({ ...summary, valuedVehicles: 0, missingVehicles: 4, totals: [] });
    render(<ManualValuationSummary />);
    await userEvent.click(screen.getByRole("button", { name: "Manuell bewertete Sammlung" }));
    expect(await screen.findByText(/keine gültigen Bewertungen/)).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("Stichtag"), { target: { value: "" } });
    fireEvent.blur(screen.getByLabelText("Stichtag"));
    expect(screen.getByText("Wähle einen Stichtag.")).toBeInTheDocument();
    expect(api.valuationSummary).toHaveBeenCalledOnce();
  });
  it("shows a failed calculation and retries only after an explicit request", async () => {
    vi.mocked(api.valuationSummary).mockRejectedValueOnce(new Error("offline"));
    const user = userEvent.setup();
    render(<ManualValuationSummary />);
    await user.click(screen.getByRole("button", { name: "Manuell bewertete Sammlung" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("nicht geladen");
    expect(api.valuationSummary).toHaveBeenCalledOnce();
    await user.click(screen.getByRole("button", { name: "Neu laden" }));
    await waitFor(() => expect(screen.queryByRole("alert")).not.toBeInTheDocument());
    expect(await screen.findByText("0.123456")).toBeInTheDocument();
  });
  it("starts collapsed and retains the cutoff through collapse using the shared calendar", async () => {
    const user = userEvent.setup();
    render(<ManualValuationSummary />);
    const toggle = screen.getByRole("button", { name: "Manuell bewertete Sammlung" });
    expect(toggle).toHaveAttribute("aria-expanded", "false");
    expect(screen.queryByRole("textbox", { name: "Stichtag" })).not.toBeInTheDocument();
    expect(api.valuationSummary).not.toHaveBeenCalled();
    await user.click(toggle);
    const date = screen.getByRole("textbox", { name: "Stichtag" });
    expect(date.closest(".app-date")).not.toBeNull();
    await user.clear(date);
    await user.type(date, "15.02.2026{Enter}");
    await waitFor(() => expect(api.valuationSummary).toHaveBeenLastCalledWith("2026-02-15"));
    await user.click(toggle);
    expect(toggle).toHaveAttribute("aria-expanded", "false");
    await user.click(toggle);
    expect(screen.getByRole("textbox", { name: "Stichtag" })).toHaveValue("15.02.2026");
    expect(screen.getByRole("button", { name: "Kalender öffnen" })).toBeInTheDocument();
  });
  it("closes the calendar portal when the section is collapsed from the keyboard", async () => {
    const user = userEvent.setup();
    render(<ManualValuationSummary />);
    const toggle = screen.getByRole("button", { name: "Manuell bewertete Sammlung" });
    await user.click(toggle);
    await user.click(screen.getByRole("button", { name: "Kalender öffnen" }));
    expect(screen.getByRole("dialog", { name: "Datum auswählen" })).toBeInTheDocument();
    toggle.focus();
    await user.keyboard("{Enter}");
    expect(toggle).toHaveAttribute("aria-expanded", "false");
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });
});
