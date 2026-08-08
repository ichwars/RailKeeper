import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { api, type AccessoryArticle } from "../../shared/api";
import { ArticlePurchaseDocumentsTab } from "./ArticlePurchaseDocumentsTab";

const article: AccessoryArticle = {
  id: "article-1", manufacturer: "Tillig", name: "Gleis", category: "straight", trackingMode: "quantity",
  manufacturerStatus: "available", articleType: "track", subtype: "straight", gauges: ["TT"], packageQuantity: 1,
  stockUnit: "piece", minimumStock: 0, inventoryStrategy: "quantity", alternativeNumbers: [], keywords: [],
  archived: false, attributes: [], createdAt: "2026-08-08T08:00:00Z", updatedAt: "2026-08-08T09:00:00Z"
};

describe("ArticlePurchaseDocumentsTab", () => {
  it("books one purchase with one purchase command and no client-side stock adjustment", async () => {
    const user = userEvent.setup();
    const createPurchase = vi.spyOn(api, "createAccessoryPurchase").mockResolvedValue({} as never);
    const adjustStock = vi.spyOn(api, "adjustAccessoryStock");
    const onChanged = vi.fn().mockResolvedValue(undefined);
    render(<ArticlePurchaseDocumentsTab article={article} disabled={false} onChanged={onChanged} resources={{
      locations: [], stock: null, movements: [], assets: [], purchases: [], documents: [], reservations: [],
      installations: [], usageHistory: null, vehicles: [], layouts: [], units: []
    }} />);

    await user.click(screen.getByRole("button", { name: "Kauf buchen" }));

    expect(createPurchase).toHaveBeenCalledOnce();
    expect(adjustStock).not.toHaveBeenCalled();
    expect(onChanged).toHaveBeenCalledOnce();
  });
});
