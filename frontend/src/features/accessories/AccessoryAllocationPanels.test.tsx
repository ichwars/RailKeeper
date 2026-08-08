import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { api, type AccessoryArticle, type Layout, type StorageLocation } from "../../shared/api";
import { AccessoryInstallationsPanel } from "./AccessoryInstallationsPanel";
import { AccessoryReservationsPanel } from "./AccessoryReservationsPanel";

const article: AccessoryArticle = {
  id: "article-1", manufacturer: "Tillig", name: "Signal", category: "signal", trackingMode: "quantity",
  manufacturerStatus: "available", articleType: "signal", subtype: "main_signal", gauges: ["TT"], packageQuantity: 1,
  stockUnit: "piece", minimumStock: 0, inventoryStrategy: "quantity", alternativeNumbers: [], keywords: [],
  archived: false, attributes: [], createdAt: "2026-08-08T08:00:00Z", updatedAt: "2026-08-08T09:00:00Z"
};
const location: StorageLocation = {
  id: "location-1", name: "Schublade", archived: false,
  createdAt: "2026-08-08T08:00:00Z", updatedAt: "2026-08-08T09:00:00Z"
};
const layout: Layout = {
  id: "layout-1", name: "Clubanlage", kind: "club", gauge: "TT", scale: "1:120", version: 1, archived: false,
  createdAt: "2026-08-08T08:00:00Z", updatedAt: "2026-08-08T09:00:00Z"
};

async function fillTechnicalFields(user: ReturnType<typeof userEvent.setup>) {
  await user.type(screen.getByRole("textbox", { name: "Einbauort" }), "Bahnhof West");
  await user.type(screen.getByRole("textbox", { name: "Digitaladresse" }), "17");
  await user.type(screen.getByRole("textbox", { name: "Decoderausgang" }), "A2");
  await user.type(screen.getByRole("textbox", { name: "Anschluss" }), "J3");
  await user.type(screen.getByRole("textbox", { name: "Verdrahtungshinweise" }), "blau/gelb");
}

describe("accessory allocation forms", () => {
  it("submits approved technical placement fields with a reservation and uses AppNumberInput", async () => {
    const user = userEvent.setup();
    const create = vi.spyOn(api, "createAccessoryReservation").mockResolvedValue({} as never);
    const view = render(<AccessoryReservationsPanel article={article} reservations={[]} assets={[]}
      locations={[location]} vehicles={[]} layouts={[layout]} units={[]} canReserve onChanged={vi.fn()}
      onDirtyChange={vi.fn()} />);
    await fillTechnicalFields(user);
    expect(Array.from(view.container.querySelectorAll("input[type='number']"))
      .every((input) => input.closest(".app-number-input"))).toBe(true);
    await user.click(screen.getByRole("button", { name: "Reservierung anlegen" }));
    await user.click(screen.getByRole("button", { name: "Bestätigen" }));

    expect(create).toHaveBeenCalledWith(expect.objectContaining({ layoutId: "layout-1", placement: "Bahnhof West",
      digitalAddress: "17", decoderOutput: "A2", connection: "J3", wiringNotes: "blau/gelb" }));
  });

  it("submits approved technical placement fields with an installation and uses AppNumberInput", async () => {
    const user = userEvent.setup();
    const create = vi.spyOn(api, "createAccessoryInstallation").mockResolvedValue({} as never);
    const view = render(<AccessoryInstallationsPanel article={article} reservations={[]} installations={[]} assets={[]}
      locations={[location]} vehicles={[]} layouts={[layout]} units={[]} canInstall onChanged={vi.fn()}
      onDirtyChange={vi.fn()} />);
    await fillTechnicalFields(user);
    expect(Array.from(view.container.querySelectorAll("input[type='number']"))
      .every((input) => input.closest(".app-number-input"))).toBe(true);
    await user.click(screen.getByRole("button", { name: "Einbau erfassen" }));
    await user.click(screen.getByRole("button", { name: "Bestätigen" }));

    expect(create).toHaveBeenCalledWith(expect.objectContaining({ layoutId: "layout-1", placement: "Bahnhof West",
      digitalAddress: "17", decoderOutput: "A2", connection: "J3", wiringNotes: "blau/gelb" }));
  });
});
