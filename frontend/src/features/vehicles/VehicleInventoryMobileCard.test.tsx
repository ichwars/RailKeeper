import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { vehicleFixture } from "../../test/fixtures/vehicles";
import { VehicleInventoryMobileCard } from "./VehicleInventoryMobileCard";

describe("VehicleInventoryMobileCard", () => {
  it("uses the type presentation column for the default mobile thumbnail", () => {
    const { container } = render(<VehicleInventoryMobileCard
      vehicle={vehicleFixture({ images: [] })}
      columns={["type", "inventoryNumber", "name"]}
      expanded
      onToggleExpanded={vi.fn()}
      onOpenDetail={vi.fn()}
      onOpenEdit={vi.fn()}
      renderQuickMenu={() => null}
    />);

    expect(container.querySelector(".vehicle-mobile-image .image-placeholder")).toBeInTheDocument();
    expect(screen.queryByText("Typ")).not.toBeInTheDocument();
  });
});
