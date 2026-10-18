import { fireEvent, render, screen, within } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { cvValueFixture } from "../../test/fixtures/vehicles";
import { VehicleCVValuesTable } from "./VehicleCVValuesTable";
import { CVImportRowsTable } from "./CVImportRowsTable";
import { buildCVImportPreview } from "./cvImport";

describe("indexed CV display", () => {
  beforeEach(() => localStorage.clear());

  it("shows explicit index pages and edits the selected row only", () => {
    const edit = vi.fn();
    const values = [cvValueFixture({ cvNumber: 257, indexCV31: 16, indexCV32: 0 }),
      cvValueFixture({ id: "other", cvNumber: 257, value: 99, indexCV31: 16, indexCV32: 3 })];
    render(<VehicleCVValuesTable values={values} readonly={false} saving={false} edit={edit} remove={vi.fn()} />);
    const row = screen.getByText("CV31=16 / CV32=3").closest("tr")!;
    fireEvent.click(within(row).getByRole("button", { name: "CV bearbeiten" }));
    expect(edit).toHaveBeenCalledWith(values[1]);
    fireEvent.change(screen.getByRole("searchbox"), { target: { value: "99" } });
    expect(screen.getByRole("status")).toHaveTextContent("1–1 von 1 Werten");
    expect(screen.queryByText("CV31=16 / CV32=0")).not.toBeInTheDocument();
  });

  it("searches readable meanings as well as raw numbers", () => {
    render(<VehicleCVValuesTable values={[cvValueFixture({ cvNumber: 19, value: 0, description: "" }),
      cvValueFixture({ id: "indexed", cvNumber: 19, value: 8, indexCV31: 0, indexCV32: 0 })]}
      readonly={false} saving={false} edit={vi.fn()} remove={vi.fn()} />);
    fireEvent.change(screen.getByRole("searchbox"), { target: { value: "Verbundadresse" } });
    expect(screen.getByRole("status")).toHaveTextContent("1–1 von 1 Werten");
    expect(screen.getByText("DCC-Verbundadresse")).toBeVisible();
    expect(screen.queryByText("CV31=0 / CV32=0")).not.toBeInTheDocument();
  });

  it("keeps all 1367 import rows selected while paging the preview", () => {
    const values = Array.from({ length: 1367 }, (_, i) => ({ cvNumber: 257 + i % 256,
      value: i % 256, indexCV31: 16, indexCV32: Math.floor(i / 256) }));
    const preview = buildCVImportPreview("full.txt", values, []);
    const toggle = vi.fn();
    render(<CVImportRowsTable preview={preview} saving={false} readonly={false} toggle={toggle} />);
    expect(screen.getAllByRole("checkbox")).toHaveLength(50);
    expect(preview.rows.filter((row) => row.selected)).toHaveLength(1367);
    fireEvent.click(screen.getByRole("button", { name: "Weiter" }));
    expect(screen.getByRole("status")).toHaveTextContent("51–100 von 1367 Werten");
    fireEvent.click(screen.getAllByRole("checkbox")[0]);
    expect(toggle).toHaveBeenCalledWith(preview.rows[50].id, false);
  });
});
