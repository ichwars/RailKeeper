import { fireEvent, render, screen, within } from "@testing-library/react";
import type { ComponentProps } from "react";
import { describe, expect, it, vi } from "vitest";

import { vehicleFixture } from "../../test/fixtures/vehicles";
import { VehicleInventoryMobileList } from "./VehicleInventoryMobileList";

function renderList(overrides: Partial<ComponentProps<typeof VehicleInventoryMobileList>> = {}) {
  const first = vehicleFixture({
    id: "vehicle-1",
    inventoryNumber: "RK-LOK-000001",
    name: "BR 106",
    manufacturer: "ESU",
    articleNumber: "12345",
    gauge: "H0",
    epoch: "III",
    series: "BR 106",
    purchaseDate: "2026-08-15"
  });
  const second = vehicleFixture({
    id: "vehicle-2",
    inventoryNumber: "RK-LOK-000002",
    name: "BR 218",
    manufacturer: "Piko",
    articleNumber: "57903",
    gauge: "H0",
    epoch: "IV",
    series: "BR 218",
    images: []
  });
  const props: ComponentProps<typeof VehicleInventoryMobileList> = {
    vehicles: [first, second],
    columns: [
      "image",
      "inventoryNumber",
      "name",
      "manufacturer",
      "articleNumber",
      "gauge",
      "epoch",
      "series",
      "purchaseDate"
    ],
    onOpenDetail: vi.fn(),
    onOpenEdit: vi.fn(),
    renderQuickMenu: (vehicle) => (
      <button type="button">Menü {vehicle.inventoryNumber}</button>
    ),
    ...overrides
  };
  return { ...render(<VehicleInventoryMobileList {...props} />), first, second, props };
}

describe("VehicleInventoryMobileList", () => {
	it("keeps set members hidden until the set is expanded", () => {
		const set = {
			id: "set-1", inventoryNumber: "RK-SET-000001", name: "Rheingold", manufacturer: "Roco",
			articleNumber: "45923", gauge: "H0", memberCount: 2, position: 1
		};
		renderList({
			vehicles: [
				vehicleFixture({ id: "member-1", inventoryNumber: "RK-WAG-000001", vehicleSet: set }),
				vehicleFixture({ id: "member-2", inventoryNumber: "RK-WAG-000002", vehicleSet: { ...set, position: 2 } })
			],
			onOpenSet: vi.fn()
		});

		expect(screen.getByText("RK-SET-000001")).toBeVisible();
		expect(screen.queryByText("RK-WAG-000001")).toBeNull();
		fireEvent.click(screen.getByRole("button", { name: /Set aufklappen/ }));
		expect(screen.getByText("RK-WAG-000001")).toBeVisible();
	});

  it("shows compact summaries and independently expands ordered remaining fields", () => {
    renderList();

    const firstToggle = screen.getByRole("button", {
      name: "Details für RK-LOK-000001 anzeigen"
    });
    const secondToggle = screen.getByRole("button", {
      name: "Details für RK-LOK-000002 anzeigen"
    });
    expect(firstToggle).toHaveAttribute("aria-expanded", "false");
    expect(secondToggle).toHaveAttribute("aria-expanded", "false");
    expect(screen.queryByRole("term")).not.toBeInTheDocument();

    fireEvent.click(firstToggle);

    expect(firstToggle).toHaveAttribute("aria-expanded", "true");
    expect(secondToggle).toHaveAttribute("aria-expanded", "false");
    expect(screen.getAllByRole("term").map((term) => term.textContent)).toEqual([
      "Baureihe",
      "Datum"
    ]);
    expect(screen.getAllByRole("definition").map((value) => value.textContent)).toEqual([
      "BR 106",
      "15.08.2026"
    ]);
  });

  it("hides an unselected image and keeps View, Edit, and Quick menu outside the toggle", () => {
    const onOpenDetail = vi.fn();
    const onOpenEdit = vi.fn();
    renderList({
      vehicles: [vehicleFixture({ images: [] })],
      columns: ["inventoryNumber", "series"],
      onOpenDetail,
      onOpenEdit
    });

    expect(screen.queryByText("Keine Vorschau")).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Anzeigen" }));
    fireEvent.click(screen.getByRole("button", { name: "Bearbeiten" }));

    expect(onOpenDetail).toHaveBeenCalledTimes(1);
    expect(onOpenEdit).toHaveBeenCalledTimes(1);
    expect(screen.getByRole("button", {
      name: "Details für RK-LOK-000001 anzeigen"
    })).toHaveAttribute("aria-expanded", "false");
    expect(screen.getByRole("button", { name: "Menü RK-LOK-000001" })).toBeVisible();
  });

  it("forgets expanded IDs when a filtered vehicle leaves the result list", () => {
    const { first, props, rerender } = renderList({ vehicles: [vehicleFixture()] });
    fireEvent.click(screen.getByRole("button", {
      name: "Details für RK-LOK-000001 anzeigen"
    }));

    rerender(<VehicleInventoryMobileList {...props} vehicles={[]} />);
    rerender(<VehicleInventoryMobileList {...props} vehicles={[first]} />);

    expect(screen.getByRole("button", {
      name: "Details für RK-LOK-000001 anzeigen"
    })).toHaveAttribute("aria-expanded", "false");
  });

  it("uses the neutral image treatment when the visible image has no file", () => {
    renderList({
      vehicles: [vehicleFixture({ images: [] })],
      columns: ["image", "inventoryNumber"]
    });

    const card = screen.getByRole("article");
    expect(within(card).getByText("Keine Vorschau")).toBeVisible();
  });
});
