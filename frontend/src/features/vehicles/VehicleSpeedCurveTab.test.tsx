import { render, screen } from "@testing-library/react";
import { expect, it } from "vitest";
import { vehicleFixture, cvValueFixture } from "../../test/fixtures/vehicles";
import { VehicleSpeedCurveTab } from "./VehicleSpeedCurveTab";

it("keeps N45 special parameters raw and states the speed-table 28/128 ambiguity", () => {
  const cvValues = [[8, 157], [29, 14], [2, 3], [5, 0], [6, 0], [67, 3], [94, 255]]
    .map(([cvNumber, value]) => cvValueFixture({ id: String(cvNumber), cvNumber, value, decoderProfile: "Kühn N45" }));
  const before = structuredClone(cvValues);
  const { rerender } = render(<VehicleSpeedCurveTab selected={vehicleFixture({ cvValues })} ecosDraft={null} />);
  expect(screen.getByText("N45-Kennlinienparameter (Rohwerte)")).toBeVisible();
  expect(screen.getByText("N45-Kennlinie aus gespeicherten Parametern")).toBeVisible();
  expect(screen.queryByRole("img")).not.toBeInTheDocument();
  rerender(<VehicleSpeedCurveTab selected={vehicleFixture({ cvValues: cvValues.map(cv => cv.cvNumber === 29
    ? { ...cv, value: 30 } : cv) })} ecosDraft={null} />);
  expect(screen.getByRole("img")).toBeVisible();
  expect(screen.getByText("30 · Tabelle ausgewählt")).toBeVisible();
  expect(screen.queryByText("30 · 28-Punkt aktiv")).not.toBeInTheDocument();
  expect(screen.getByText(/bei 128 nicht/)).toBeVisible();
  expect(cvValues).toEqual(before);
});

it("explains the internally smoothed Lenz characteristic instead of drawing a guessed curve", () => {
  const cvValues = [[8, 99], [7, 79], [29, 14], [2, 1], [5, 254], [6, 48], [67, 20]]
    .map(([cvNumber, value]) => cvValueFixture({ id: String(cvNumber), cvNumber, value, decoderProfile: "STANDARD+ V2" }));
  const before = structuredClone(cvValues);
  render(<VehicleSpeedCurveTab selected={vehicleFixture({ cvValues })} ecosDraft={null} />);
  expect(screen.getByText("Geglättete Lenz-Kennlinie")).toBeVisible();
  expect(screen.getByText("Lenz-Stützparameter (Rohwerte)")).toBeVisible();
  expect(screen.getByText("14 · intern geglättet")).toBeVisible();
  expect(screen.queryByRole("img")).not.toBeInTheDocument();
  expect(cvValues).toEqual(before);
});

it("explains the manufacturer's characteristic and preserves raw values without a guessed graph", () => {
  const cvValues = [[8, 97], [261, 180], [29, 30], [48, 5], [2, 0], [5, 92], [6, 100], [67, 123]]
    .map(([cvNumber, value]) => cvValueFixture({ id: String(cvNumber), cvNumber, value, decoderProfile: "DH18A" }));
  const before = structuredClone(cvValues);
  render(<VehicleSpeedCurveTab selected={vehicleFixture({ cvValues })} ecosDraft={null} />);
  expect(screen.getByText("Kennlinie über CV48")).toBeVisible();
  expect(screen.getByText(/CV48 = 5/)).toBeVisible();
  expect(screen.getByText("30 · Rohwert")).toBeVisible();
  expect(screen.queryByText("3-Punkt-Kurve")).not.toBeInTheDocument();
  expect(screen.queryByRole("img")).not.toBeInTheDocument();
  expect(cvValues).toEqual(before);
});

it("explains ZIMO special values while keeping their raw parameters separate from the chart", () => {
  const cvValues = [[8, 145], [29, 14], [2, 1], [5, 1], [6, 1]]
    .map(([cvNumber, value]) => cvValueFixture({ id: String(cvNumber), cvNumber, value, decoderProfile: "MN180N18" }));
  const before = structuredClone(cvValues);
  render(<VehicleSpeedCurveTab selected={vehicleFixture({ cvValues })} ecosDraft={null} />);
  expect(screen.getByText("Geglättete ZIMO-Kennlinie")).toBeVisible();
  expect(screen.getByText(/Interner Höchstwert: 255/)).toBeVisible();
  expect(screen.getByText(/etwa ¼ des Höchstwerts/)).toBeVisible();
  expect(screen.getByText("ZIMO-Stützparameter (Rohwerte)")).toBeVisible();
  expect(screen.queryByRole("img")).not.toBeInTheDocument();
  expect(cvValues).toEqual(before);
});

it("keeps 76420 parameters on their documented scale and labels the correct trim CVs", () => {
  const cvValues = [[8, 85], [29, 6], [2, 1], [5, 48], [6, 24], [47, 32], [48, 30], [66, 200], [67, 1], [94, 255]]
    .map(([cvNumber, value]) => cvValueFixture({ id: String(cvNumber), cvNumber, value, decoderProfile: "Uhlenbrock 76420" }));
  const { rerender } = render(<VehicleSpeedCurveTab selected={vehicleFixture({ cvValues })} ecosDraft={null} />);
  expect(screen.getByText("76420-Kennlinienparameter")).toBeVisible();
  expect(screen.getByText("CV2/5/6 · Rohparameter 1–63")).toBeVisible();
  expect(screen.getByText("Korrektur CV47 / CV48")).toBeVisible(); expect(screen.getByText("32 / 30")).toBeVisible();
  expect(screen.queryByRole("img")).not.toBeInTheDocument();
  rerender(<VehicleSpeedCurveTab selected={vehicleFixture({ cvValues: cvValues.map(cv => cv.cvNumber === 29 ? { ...cv, value: 22 } : cv) })} ecosDraft={null} />);
  expect(screen.getByRole("img")).toBeVisible();
  expect(screen.getByText(/Motorola-Programmierseite/)).toBeVisible();
});

it("shows only the Roco10882-selected characteristic and keeps an invalid configuration raw", () => {
  const cvValues = [[8, 145], [29, 14], [2, 4], [5, 252], [6, 85], [67, 1], [94, 255]]
    .map(([cvNumber, value]) => cvValueFixture({ id: String(cvNumber), cvNumber, value, decoderProfile: "Roco 10882" }));
  const { rerender } = render(<VehicleSpeedCurveTab selected={vehicleFixture({ cvValues })} ecosDraft={null} />);
  expect(screen.getByRole("img")).toBeVisible();
  expect(screen.getByText("14 · 3-Punkt aktiv")).toBeVisible();
  rerender(<VehicleSpeedCurveTab selected={vehicleFixture({ cvValues: cvValues.map(cv => cv.cvNumber === 29 ? { ...cv, value: 30 } : cv) })} ecosDraft={null} />);
  expect(screen.getByRole("img")).toBeVisible(); expect(screen.getByText("30 · 28-Punkt aktiv")).toBeVisible();
  rerender(<VehicleSpeedCurveTab selected={vehicleFixture({ cvValues: cvValues.map(cv => cv.cvNumber === 29 ? { ...cv, value: 256 } : cv) })} ecosDraft={null} />);
  expect(screen.queryByRole("img")).not.toBeInTheDocument(); expect(screen.getByText("256 · Rohwert")).toBeVisible();
});
