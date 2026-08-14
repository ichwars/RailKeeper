import { render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { api } from "../../shared/api";
import { LayoutTwinHistory } from "./LayoutTwinHistory";

describe("LayoutTwinHistory", () => {
  afterEach(() => vi.restoreAllMocks());

  it("loads product history lazily and keeps only events for the selected position", async () => {
    const history = vi.spyOn(api, "accessoryUsageHistory").mockResolvedValue({
      productId: "product-1",
      events: [{
        id: "condition-1", type: "condition_changed", productId: "product-1",
        installationId: "installation-1", quantity: 1, technicalPositionId: "position-1",
        previousCondition: "ready", condition: "maintenance_due", occurredAt: "2026-08-09T12:00:00Z"
      }, {
        id: "installation-other", type: "installation", productId: "product-1",
        installationId: "installation-2", quantity: 2, technicalPositionId: "position-2",
        condition: "ready", occurredAt: "2026-08-09T11:00:00Z"
      }]
    });

    render(<LayoutTwinHistory positionID="position-1" productID="product-1" />);

    expect(screen.getByRole("status")).toHaveTextContent("Verlauf wird geladen");
    expect(await screen.findByText("Zustandsänderung")).toBeInTheDocument();
    expect(screen.getByText(/Einsatzbereit → Wartung fällig/)).toBeInTheDocument();
    expect(screen.queryByText("Einbau")).not.toBeInTheDocument();
    expect(history).toHaveBeenCalledWith("product-1");
  });

  it("does not request history when the position has no linked article", async () => {
    const history = vi.spyOn(api, "accessoryUsageHistory");
    render(<LayoutTwinHistory positionID="position-1" />);

    expect(screen.getByText(/verknüpften Artikel/)).toBeInTheDocument();
    await waitFor(() => expect(history).not.toHaveBeenCalled());
  });

  it("keeps loading failures inside the inspector", async () => {
    vi.spyOn(api, "accessoryUsageHistory").mockRejectedValue(new Error("offline"));
    render(<LayoutTwinHistory positionID="position-1" productID="product-1" />);

    expect(await screen.findByRole("alert")).toHaveTextContent("Verlauf konnte nicht geladen werden");
  });
});
