import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import type { AccessoryArticleListItem, LayoutUnit } from "../../shared/api";
import { LayoutTechnicalPositionDialog } from "./LayoutTechnicalPositionDialog";

const unit: LayoutUnit = {
  id: "unit-1", layoutId: "layout-1", name: "Bahnhofsmodul", kind: "module", widthMm: 1200,
  heightMm: 500, version: 1, archived: false, createdAt: "2026-08-09T10:00:00Z",
  updatedAt: "2026-08-09T10:00:00Z"
};

const product: AccessoryArticleListItem = {
  id: "product-1", inventoryNumber: "RK-ART-000001", manufacturer: "Tillig", articleNumber: "83101",
  name: "Gleisstück", articleType: "track", subtype: "straight", gauges: ["TT"],
  inventoryStrategy: "quantity", archived: false, owned: 4, available: 4, reserved: 0, installed: 0,
  locationNames: [], hasUsageHistory: false, careHintCount: 0, updatedAt: "2026-08-09T10:00:00Z",
  attributes: []
};

describe("LayoutTechnicalPositionDialog", () => {
  it("uses app-owned controls, focuses the label, and submits numeric coordinates", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    render(<LayoutTechnicalPositionDialog unit={unit} products={[product]} saving={false} message=""
      conflict={false} onSubmit={onSubmit} onClose={() => undefined} />);

    const dialog = screen.getByRole("dialog", { name: "Technische Position anlegen" });
    const label = within(dialog).getByRole("textbox", { name: "Bezeichnung" });
    expect(label).toHaveFocus();
    expect(dialog.querySelector("select")).toBeNull();
    expect(within(dialog).getByRole("spinbutton", { name: "X-Position (mm)" }).closest(".app-number-input"))
      .not.toBeNull();

    await user.type(label, "Signal B");
    await user.click(within(dialog).getByRole("button", { name: "Positionsart" }));
    await user.click(screen.getByRole("option", { name: "Signal" }));
    await user.click(within(dialog).getByRole("button", { name: "Zugeordneter Artikel" }));
    await user.click(screen.getByRole("option", { name: /RK-ART-000001/ }));
    await user.click(within(dialog).getByRole("button", { name: "Position speichern" }));

    expect(onSubmit).toHaveBeenCalledWith(expect.objectContaining({
      label: "Signal B", kind: "signal", positionXMm: 0, positionYMm: 0,
      rotationDegrees: 0, productId: product.id
    }), undefined);
  });

  it("asks before discarding a changed position", async () => {
    const user = userEvent.setup();
    const onClose = vi.fn();
    render(<LayoutTechnicalPositionDialog unit={unit} products={[]} saving={false} message=""
      conflict={false} onSubmit={() => undefined} onClose={onClose} />);

    await user.type(screen.getByRole("textbox", { name: "Bezeichnung" }), "Entwurf");
    await user.click(screen.getByRole("button", { name: "Abbrechen" }));
    expect(onClose).not.toHaveBeenCalled();
    const confirm = screen.getByRole("dialog", { name: "Änderungen verwerfen?" });
    await user.click(within(confirm).getByRole("button", { name: "Verwerfen" }));
    expect(onClose).toHaveBeenCalledTimes(1);
  });
});
