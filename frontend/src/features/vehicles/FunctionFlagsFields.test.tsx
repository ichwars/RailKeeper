import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { FunctionFlagsFields } from "./FunctionFlagsFields";
import { functionMappingsFromImport } from "./cvImport";

describe("independent function flags", () => {
  it("keeps missing flags unknown and edits each flag without changing the other", () => {
    const update = vi.fn();
    render(<FunctionFlagsFields functionKey="F1" edit={{ inverted: true, directionDependent: null, legacyDirectionDependent: false }}
      disabled={false} update={update} />);
    expect(screen.getByLabelText("F1 Richtungsabhängig")).toHaveTextContent("Unbekannt");
    expect(screen.getByLabelText("F1 Invertiert")).toHaveTextContent("ja");
    expect(screen.getByText(/Altangabe: nein/)).toBeVisible();
    fireEvent.click(screen.getByLabelText("F1 Richtungsabhängig"));
    fireEvent.click(screen.getByRole("option", { name: "nein" }));
    expect(update).toHaveBeenLastCalledWith({ directionDependent: false, flagsVersion: 2 });
    fireEvent.click(screen.getByLabelText("F1 Invertiert"));
    fireEvent.click(screen.getByRole("option", { name: "Unbekannt" }));
    expect(update).toHaveBeenLastCalledWith({ inverted: null, flagsVersion: 2 });
  });
  it("locks both controls for read-only users", () => {
    render(<FunctionFlagsFields functionKey="F2" edit={{}} disabled update={vi.fn()} />);
    expect(screen.getByLabelText("F2 Richtungsabhängig")).toBeDisabled();
    expect(screen.getByLabelText("F2 Invertiert")).toBeDisabled();
  });
  it("shows both independent fields in English", () => {
    window.localStorage.setItem("railkeeper.settings.language", "en");
    render(<FunctionFlagsFields functionKey="F1" edit={{}} disabled={false} update={vi.fn()} />);
    expect(screen.getByLabelText("F1 Direction-dependent")).toHaveTextContent("Unknown");
    expect(screen.getByLabelText("F1 Inverted")).toHaveTextContent("Unknown");
  });
  it("does not reinterpret old function JSON flags", () => {
    for (const value of [true, false]) {
      const [row] = functionMappingsFromImport(JSON.stringify({ functions: [{ functionKey: "F1", directionDependent: value }] }));
      expect(row).toMatchObject({ directionDependent: null, inverted: null, legacyDirectionDependent: value, flagsVersion: 1 });
    }
  });
  it("round trips explicit false, true, unknown and legacy references in current JSON", () => {
    for (const directionDependent of [true, false, null]) for (const inverted of [true, false, null]) {
      const [row] = functionMappingsFromImport(JSON.stringify({ flagsVersion: 2,
        functions: [{ functionKey: "F1", directionDependent, inverted, legacyDirectionDependent: false }] }));
      expect(row).toMatchObject({ directionDependent, inverted, legacyDirectionDependent: false, flagsVersion: 2 });
    }
    expect(() => functionMappingsFromImport('{"flagsVersion":2,"functions":[{"functionKey":"F1","inverted":"false"}]}'))
      .toThrow(/Funktionsangaben/);
  });
});
