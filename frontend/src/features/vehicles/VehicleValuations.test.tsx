import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { api } from "../../shared/api";
import { setLanguage } from "../../shared/i18n";
import type { VehicleValuation } from "../../shared/apiValuations";
import { VehicleValuations } from "./VehicleValuations";

const record: VehicleValuation = {  
id: "1".repeat(32), vehicleId: "a", amount: "1.123456", currency: "CHF", valuedOn: "2026-01-01",
  source: "Original source", notes: "Historical note", replacesId: "", withdrawn: false, recordedAt: "2026-11-07T12:00:00.000000000Z", replacedById: ""
};

describe("manual vehicle valuation presentation", () => {
  beforeEach(() => {
    setLanguage("de");
    vi.spyOn(api, "vehicleValuations").mockResolvedValue({ entries: [record], total: 1 });
    vi.spyOn(api, "recordVehicleValuation").mockRejectedValue(new Error("response lost"));
  });
  it("shows exact value, provenance and actual recording time without write controls for readers", async () => {
    render(<VehicleValuations vehicleID="a" />);
    expect(await screen.findByText("1.123456 CHF")).toBeInTheDocument();
    expect(screen.getByText("Original source")).toBeInTheDocument();
    expect(screen.getByText(/Tatsächlich erfasst:/)).toBeInTheDocument();
    expect(screen.queryByRole("textbox")).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Bewertung berichtigen" })).not.toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Verlauf als JSON" })).toHaveAttribute("href", "/api/v1/vehicles/a/valuations/export");
  });
  it("requires explicit correction and a fresh withdrawal reason", async () => {
    const user = userEvent.setup();
    render(<VehicleValuations vehicleID="a" canEdit />);
    await user.click(await screen.findByRole("button", { name: "Bewertung berichtigen" }));
    const correction = screen.getByRole("group", { name: "Bewertung berichtigen" });
    expect(within(correction).getByRole("textbox", { name: /Herkunft/ })).toHaveValue("Original source");
    await user.click(screen.getByRole("button", { name: "Bewertung zurückziehen" }));
    const withdrawal = screen.getByRole("group", { name: "Bewertung zurückziehen" });
    expect(within(withdrawal).getByRole("textbox", { name: /Begründung/ })).toHaveValue("");
    expect(within(withdrawal).getByRole("textbox", { name: /Betrag/ })).toHaveAttribute("readonly");
    expect(api.recordVehicleValuation).not.toHaveBeenCalled();
  });
  it("locks an uncertain input and leaves the read-only verification action available", async () => {
    const user = userEvent.setup();
    render(<VehicleValuations vehicleID="a" canEdit />);
    await screen.findByText("1.123456 CHF");
    await user.type(screen.getByRole("textbox", { name: /Betrag/ }), "12,345");
    await user.type(screen.getByRole("textbox", { name: /Herkunft/ }), "Manual");
    await user.click(screen.getByRole("button", { name: "Bewertung jetzt speichern" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("Speicherbestätigung");
    expect(screen.getByRole("textbox", { name: /Betrag/ })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Gespeicherten Stand prüfen" })).toBeEnabled();
    expect(api.recordVehicleValuation).toHaveBeenCalledOnce();
  });
  it("keeps displayed history when a reload fails", async () => {
    const user = userEvent.setup();
    render(<VehicleValuations vehicleID="a" />);
    await screen.findByText("1.123456 CHF");
    vi.mocked(api.vehicleValuations).mockRejectedValueOnce(new Error("offline"));
    await user.click(screen.getByRole("button", { name: "Neu laden" }));
    await waitFor(() => expect(screen.getByRole("alert")).toHaveTextContent("nicht geladen"));
    expect(screen.getByText("1.123456 CHF")).toBeInTheDocument();
  });
});
