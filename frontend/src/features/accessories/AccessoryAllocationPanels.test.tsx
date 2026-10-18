import { vehicleFixture } from "../../test/fixtures/vehicles";
import { useState } from "react";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import {
  api,
  type AccessoryAsset,
  type AccessoryArticle,
  type AccessoryInstallation,
  type AccessoryReservation,

  type StorageLocation
} from "../../shared/api";
import { AccessoryInstallationsPanel } from "./AccessoryInstallationsPanel";
import { AccessoryReservationsPanel } from "./AccessoryReservationsPanel";

const article: AccessoryArticle = {
  id: "article-1", inventoryNumber: "RK-ART-000001", manufacturer: "Tillig", name: "Signal",
  category: "signal", trackingMode: "quantity",
  manufacturerStatus: "available", articleType: "signal", subtype: "main_signal", gauges: ["TT"], packageQuantity: 1,
  stockUnit: "piece", minimumStock: 0, inventoryStrategy: "quantity", alternativeNumbers: [], keywords: [],
  archived: false, attributes: [], createdAt: "2026-08-08T08:00:00Z", updatedAt: "2026-08-08T09:00:00Z"
};
const location: StorageLocation = {
  id: "location-1", name: "Schublade", archived: false,
  createdAt: "2026-08-08T08:00:00Z", updatedAt: "2026-08-08T09:00:00Z"
};
const vehicle = vehicleFixture({id: "vehicle-1", name: "Testlok"});
const reservation: AccessoryReservation = {
  id: "reservation-1", productId: article.id, locationId: location.id, quantity: 1, vehicleId: vehicle.id,
  placement: "Bahnhof West", digitalAddress: "17", decoderOutput: "A2", connection: "J3",
  wiringNotes: "blau/gelb", status: "active", createdBy: "planner",
  createdAt: "2026-08-08T08:00:00Z", updatedAt: "2026-08-08T08:00:00Z"
};
const installation: AccessoryInstallation = {
  id: "installation-1", productId: article.id, sourceLocationId: location.id, quantity: 1,
  condition: "ready", vehicleId: vehicle.id, installedBy: "editor", installedAt: "2026-08-08T08:00:00Z"
};
const asset: AccessoryAsset = {
  id: "asset-1", productId: article.id, inventoryNumber: "RK-83101-001", condition: "ready",
  lifecycle: "stored", storageLocationId: location.id, createdAt: "2026-08-08T08:00:00Z",
  updatedAt: "2026-08-08T08:00:00Z"
};
const hybridArticle: AccessoryArticle = {
  ...article, articleType: "track", subtype: "straight", inventoryStrategy: "quantity_later_individual"
};

async function fillTechnicalFields(user: ReturnType<typeof userEvent.setup>) {
  await user.type(screen.getByRole("textbox", { name: "Einbauort" }), "Bahnhof West");
  await user.type(screen.getByRole("textbox", { name: "Digitaladresse" }), "17");
  await user.type(screen.getByRole("textbox", { name: "Decoderausgang" }), "A2");
  await user.type(screen.getByRole("textbox", { name: "Anschluss" }), "J3");
  await user.type(screen.getByRole("textbox", { name: "Verdrahtungshinweise" }), "blau/gelb");
}

describe("accessory allocation forms", () => {
  it("marks reservation and installation forms for the responsive two-column grid", () => {
    const reservationView = render(<AccessoryReservationsPanel article={article} reservations={[]} assets={[]}
      locations={[location]} vehicles={[vehicle]} canReserve onChanged={vi.fn()}
      onDirtyChange={vi.fn()} />);
    expect(screen.getByRole("heading", { name: "Zubehör reservieren" }).closest("form"))
      .toHaveClass("accessory-allocation-form");
    reservationView.unmount();

    render(<AccessoryInstallationsPanel article={article} reservations={[]} installations={[]} assets={[]}
      locations={[location]} vehicles={[vehicle]} canInstall onChanged={vi.fn()}
      onDirtyChange={vi.fn()} />);
    expect(screen.getByRole("heading", { name: "Zubehör einbauen" }).closest("form"))
      .toHaveClass("accessory-allocation-form");
  });

  it("submits approved technical placement fields with a reservation and uses AppNumberInput", async () => {
    const user = userEvent.setup();
    const create = vi.spyOn(api, "createAccessoryReservation").mockResolvedValue({} as never);
    const view = render(<AccessoryReservationsPanel article={article} reservations={[]} assets={[]}
      locations={[location]} vehicles={[vehicle]} canReserve onChanged={vi.fn()}
      onDirtyChange={vi.fn()} />);
    await fillTechnicalFields(user);
    expect(Array.from(view.container.querySelectorAll("input[type='number']"))
      .every((input) => input.closest(".app-number-input"))).toBe(true);
    await user.click(screen.getByRole("button", { name: "Reservierung anlegen" }));
    await user.click(screen.getByRole("button", { name: "Bestätigen" }));

    expect(create).toHaveBeenCalledWith(expect.objectContaining({ vehicleId: "vehicle-1", placement: "Bahnhof West",
      digitalAddress: "17", decoderOutput: "A2", connection: "J3", wiringNotes: "blau/gelb" }));
    expect(screen.queryByRole("button", { name: "Bestandsquelle" })).not.toBeInTheDocument();
  });

  it("resets the vehicle reservation dirty signal after a successful confirmation", async () => {
    const user = userEvent.setup();
    vi.spyOn(api, "createAccessoryReservation").mockResolvedValue({} as never);
    const onDirtyChange = vi.fn();
    render(<AccessoryReservationsPanel article={article} reservations={[]} assets={[]}
      locations={[location]} vehicles={[vehicle]} canReserve onChanged={vi.fn()}
      onDirtyChange={onDirtyChange} />);

    await user.type(screen.getByRole("textbox", { name: "Einbauort" }), "Bahnhof West");
    await waitFor(() => expect(onDirtyChange).toHaveBeenLastCalledWith(true));
    await user.click(screen.getByRole("button", { name: "Reservierung anlegen" }));
    await user.click(screen.getByRole("button", { name: "Bestätigen" }));

    await waitFor(() => expect(onDirtyChange).toHaveBeenLastCalledWith(false));
  });

  it("focuses the stable reservations heading after a successful cancellation refresh", async () => {
    const user = userEvent.setup();
    vi.spyOn(api, "cancelAccessoryReservation").mockResolvedValue({} as never);
    function Harness() {
      const [reservations, setReservations] = useState([reservation]);
      return <AccessoryReservationsPanel article={article} reservations={reservations} assets={[]}
        locations={[location]} vehicles={[vehicle]} canReserve
        onChanged={async () => setReservations([])} onDirtyChange={vi.fn()} />;
    }
    render(<Harness />);
    const heading = screen.getByRole("heading", { name: "Reservierungen" });

    await user.click(screen.getByRole("button", { name: "Stornieren" }));
    await user.click(screen.getByRole("button", { name: "Bestätigen" }));

    await waitFor(() => expect(screen.queryByRole("button", { name: "Stornieren" })).not.toBeInTheDocument());
    expect(heading).toHaveFocus();
  });

  it("lets hybrid stock reserve free quantity even when a stored asset exists", async () => {
    const user = userEvent.setup();
    const create = vi.spyOn(api, "createAccessoryReservation").mockResolvedValue({} as never);
    render(<AccessoryReservationsPanel article={hybridArticle} reservations={[]} assets={[asset]}
      locations={[location]} vehicles={[vehicle]} canReserve onChanged={vi.fn()}
      onDirtyChange={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: "Bestandsquelle" }));
    await user.click(screen.getByRole("option", { name: "Menge" }));
    const quantity = screen.getByRole("spinbutton", { name: "Menge" });
    await user.clear(quantity);
    await user.type(quantity, "2");
    await user.click(screen.getByRole("button", { name: "Reservierung anlegen" }));
    await user.click(screen.getByRole("button", { name: "Bestätigen" }));

    expect(create).toHaveBeenCalledWith(expect.objectContaining({
      productId: article.id, quantity: 2, locationId: location.id
    }));
    expect(create.mock.calls[0]?.[0]).not.toHaveProperty("assetId");
  });

  it("lets hybrid stock reserve a stored asset independently from free quantity", async () => {
    const user = userEvent.setup();
    const create = vi.spyOn(api, "createAccessoryReservation").mockResolvedValue({} as never);
    render(<AccessoryReservationsPanel article={hybridArticle} reservations={[]} assets={[asset]}
      locations={[location]} vehicles={[vehicle]} canReserve onChanged={vi.fn()}
      onDirtyChange={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: "Bestandsquelle" }));
    await user.click(screen.getByRole("option", { name: "Einzelstück" }));
    expect(screen.getByRole("button", { name: "Einzelstück" })).toHaveTextContent("RK-83101-001");
    await user.click(screen.getByRole("button", { name: "Reservierung anlegen" }));
    await user.click(screen.getByRole("button", { name: "Bestätigen" }));

    expect(create).toHaveBeenCalledWith(expect.objectContaining({
      productId: article.id, assetId: asset.id, quantity: 1, locationId: location.id
    }));
  });

  it("submits approved technical placement fields with an installation and uses AppNumberInput", async () => {
    const user = userEvent.setup();
    const create = vi.spyOn(api, "createAccessoryInstallation").mockResolvedValue({} as never);
    const view = render(<AccessoryInstallationsPanel article={article} reservations={[]} installations={[]} assets={[]}
      locations={[location]} vehicles={[vehicle]} canInstall onChanged={vi.fn()}
      onDirtyChange={vi.fn()} />);
    await fillTechnicalFields(user);
    expect(Array.from(view.container.querySelectorAll("input[type='number']"))
      .every((input) => input.closest(".app-number-input"))).toBe(true);
    await user.click(screen.getByRole("button", { name: "Einbau erfassen" }));
    await user.click(screen.getByRole("button", { name: "Bestätigen" }));

    expect(create).toHaveBeenCalledWith(expect.objectContaining({ vehicleId: "vehicle-1", placement: "Bahnhof West",
      digitalAddress: "17", decoderOutput: "A2", connection: "J3", wiringNotes: "blau/gelb" }));
    expect(screen.queryByRole("button", { name: "Bestandsquelle" })).not.toBeInTheDocument();
  });

  it("focuses the stable installations heading after a successful removal refresh", async () => {
    const user = userEvent.setup();
    vi.spyOn(api, "removeAccessoryInstallation").mockResolvedValue({} as never);
    function Harness() {
      const [installations, setInstallations] = useState([installation]);
      return <AccessoryInstallationsPanel article={article} reservations={[]} installations={installations}
        assets={[]} locations={[location]} vehicles={[vehicle]} canInstall
        onChanged={async () => setInstallations([])} onDirtyChange={vi.fn()} />;
    }
    render(<Harness />);
    const heading = screen.getByRole("heading", { name: "Einbau und Historie" });

    await user.click(screen.getByRole("button", { name: "Ausbauen" }));
    await user.click(screen.getAllByRole("button", { name: "Ausbauen" })[1]!);
    await user.click(screen.getByRole("button", { name: "Bestätigen" }));

    await waitFor(() => expect(screen.queryByRole("button", { name: "Ausbauen" })).not.toBeInTheDocument());
    expect(heading).toHaveFocus();
  });

  it("lets hybrid stock install quantity and asset through separate direct allocation modes", async () => {
    const user = userEvent.setup();
    const create = vi.spyOn(api, "createAccessoryInstallation").mockResolvedValue({} as never);
    const view = render(<AccessoryInstallationsPanel article={hybridArticle} reservations={[]} installations={[]}
      assets={[asset]} locations={[location]} vehicles={[vehicle]} canInstall
      onChanged={vi.fn()} onDirtyChange={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: "Bestandsquelle" }));
    await user.click(screen.getByRole("option", { name: "Menge" }));
    await user.click(screen.getByRole("button", { name: "Einbau erfassen" }));
    await user.click(screen.getByRole("button", { name: "Bestätigen" }));
    expect(create).toHaveBeenLastCalledWith(expect.objectContaining({ quantity: 1 }));
    expect(create.mock.calls[0]?.[0]).not.toHaveProperty("assetId");

    view.unmount();
    render(<AccessoryInstallationsPanel article={hybridArticle} reservations={[]} installations={[]}
      assets={[asset]} locations={[location]} vehicles={[vehicle]} canInstall
      onChanged={vi.fn()} onDirtyChange={vi.fn()} />);
    await user.click(screen.getByRole("button", { name: "Bestandsquelle" }));
    await user.click(screen.getByRole("option", { name: "Einzelstück" }));
    await user.click(screen.getByRole("button", { name: "Einbau erfassen" }));
    await user.click(screen.getByRole("button", { name: "Bestätigen" }));
    expect(create).toHaveBeenLastCalledWith(expect.objectContaining({ assetId: asset.id, quantity: 1 }));
  });

  it("keeps an asset-bound reservation locked to its asset when installing", async () => {
    const user = userEvent.setup();
    const create = vi.spyOn(api, "createAccessoryInstallation").mockResolvedValue({} as never);
    const assetReservation = { ...reservation, assetId: asset.id };
    render(<AccessoryInstallationsPanel article={hybridArticle} reservations={[assetReservation]}
      installations={[]} assets={[asset]} locations={[location]} vehicles={[vehicle]}
      canInstall onChanged={vi.fn()} onDirtyChange={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: "Reservierung" }));
    await user.click(screen.getByRole("option", { name: "RK-LOK-000001 · Testlok" }));
    expect(screen.queryByRole("button", { name: "Bestandsquelle" })).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Einzelstück" })).toBeDisabled();
    await user.click(screen.getByRole("button", { name: "Einbau erfassen" }));
    await user.click(screen.getByRole("button", { name: "Bestätigen" }));

    expect(create).toHaveBeenCalledWith(expect.objectContaining({
      reservationId: reservation.id, assetId: asset.id, quantity: 1
    }));
  });

  it("prefills reservation technical data, keeps it editable, and sends explicit overrides", async () => {
    const user = userEvent.setup();
    const create = vi.spyOn(api, "createAccessoryInstallation").mockResolvedValue({} as never);
    render(<AccessoryInstallationsPanel article={article} reservations={[reservation]} installations={[]} assets={[]}
      locations={[location]} vehicles={[vehicle]} canInstall onChanged={vi.fn()}
      onDirtyChange={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: "Reservierung" }));
    await user.click(screen.getByRole("option", { name: "RK-LOK-000001 · Testlok" }));
    expect(screen.getByRole("textbox", { name: "Einbauort" })).toHaveValue("Bahnhof West");
    expect(screen.getByRole("textbox", { name: "Digitaladresse" })).toHaveValue("17");
    expect(screen.getByRole("textbox", { name: "Decoderausgang" })).toHaveValue("A2");
    const connection = screen.getByRole("textbox", { name: "Anschluss" });
    expect(connection).toHaveValue("J3");
    expect(screen.getByRole("textbox", { name: "Verdrahtungshinweise" })).toHaveValue("blau/gelb");
    await user.clear(connection);
    await user.type(connection, "J4");
    await user.click(screen.getByRole("button", { name: "Einbau erfassen" }));
    await user.click(screen.getByRole("button", { name: "Bestätigen" }));

    expect(create).toHaveBeenCalledWith(expect.objectContaining({ reservationId: reservation.id,
      placement: "Bahnhof West", digitalAddress: "17", decoderOutput: "A2", connection: "J4",
      wiringNotes: "blau/gelb" }));
  });
});
