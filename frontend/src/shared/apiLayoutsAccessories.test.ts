import { beforeEach, describe, expect, it, vi } from "vitest";

import { api } from "./api";

describe("layout and accessory API client", () => {
  const fetchMock = vi.fn().mockResolvedValue({
    ok: true,
    status: 200,
    statusText: "OK",
    json: async () => ({})
  });

  beforeEach(() => {
    fetchMock.mockClear();
    vi.stubGlobal("fetch", fetchMock);
    document.cookie = "rk_csrf=client-test-token; path=/";
  });

  it("encodes layout paths, methods, bodies, and CSRF headers", async () => {
    const layoutInput = {
      name: "Clubanlage",
      kind: "club" as const,
      gauge: "TT",
      scale: "1:120",
      description: "Vereinsanlage",
      archived: false
    };
    const unitInput = {
      name: "Modul 1",
      kind: "module" as const,
      ownerLabel: "Daniel",
      widthMm: 1000,
      heightMm: 500,
      archived: false
    };
    const configurationInput = {
      name: "Ausstellung",
      description: "Aufbau A",
      expectedVersion: 2,
      archived: false,
      units: [{
        unitId: "unit/1",
        planRevisionId: "revision/1",
        positionXMm: 10,
        positionYMm: 20,
        rotationDegrees: 90,
        sortOrder: 0
      }]
    };

    await api.layouts();
    await api.createLayout(layoutInput);
    await api.layout("layout/1");
    await api.updateLayout("layout/1", { ...layoutInput, expectedVersion: 3 });
    await api.layoutUnits("layout/1");
    await api.createLayoutUnit("layout/1", unitInput);
    await api.updateLayoutUnit("unit/1", { ...unitInput, expectedVersion: 2 });
    await api.layoutConfigurations("layout/1");
    await api.createLayoutConfiguration("layout/1", configurationInput);
    await api.updateLayoutConfiguration("configuration/1", configurationInput);
    await api.planVariants("unit/1");
    await api.createPlanVariant("unit/1", { name: "Sommer", description: "Variante" });
    await api.createPlanRevision("variant/1", { baseRevisionId: "revision/1" });
    await api.submitPlanRevision("revision/1", 4);
    await api.publishPlanRevision("revision/1", 5);

    expectRequests([
      ["GET", "/api/v1/layouts"],
      ["POST", "/api/v1/layouts", layoutInput],
      ["GET", "/api/v1/layouts/layout%2F1"],
      ["PUT", "/api/v1/layouts/layout%2F1", { ...layoutInput, expectedVersion: 3 }],
      ["GET", "/api/v1/layouts/layout%2F1/units"],
      ["POST", "/api/v1/layouts/layout%2F1/units", unitInput],
      ["PUT", "/api/v1/layout-units/unit%2F1", { ...unitInput, expectedVersion: 2 }],
      ["GET", "/api/v1/layouts/layout%2F1/configurations"],
      ["POST", "/api/v1/layouts/layout%2F1/configurations", configurationInput],
      ["PUT", "/api/v1/layout-configurations/configuration%2F1", configurationInput],
      ["GET", "/api/v1/layout-units/unit%2F1/plan-variants"],
      ["POST", "/api/v1/layout-units/unit%2F1/plan-variants", { name: "Sommer", description: "Variante" }],
      ["POST", "/api/v1/plan-variants/variant%2F1/revisions", { baseRevisionId: "revision/1" }],
      ["POST", "/api/v1/plan-revisions/revision%2F1/submit", { expectedVersion: 4 }],
      ["POST", "/api/v1/plan-revisions/revision%2F1/publish", { expectedVersion: 5 }]
    ]);
  });

  it("encodes accessory paths, filters, bodies, and CSRF headers", async () => {
    const productInput = {
      manufacturer: "Tillig",
      articleNumber: "83101",
      name: "Gerades Gleis",
      category: "Gleismaterial",
      trackingMode: "quantity" as const,
      description: "Modellgleis"
    };
    const locationInput = { name: "Schrank A", parentId: "room/1", description: "Fach 2", archived: false };
    const assetInput = {
      inventoryNumber: "RK-Z-0001",
      serialNumber: "123",
      condition: "ready" as const,
      lifecycle: "stored" as const,
      storageLocationId: "location/1",
      purchaseDate: "2026-08-07",
      purchasePrice: "19.90",
      warrantyUntil: "2028-08-07",
      notes: "Test"
    };
    const reservationInput = {
      productId: "product/1",
      locationId: "location/1",
      quantity: 2,
      layoutId: "layout/1" as const,
      note: "Bahnhof"
    };
    const installationInput = {
      reservationId: "reservation/1",
      productId: "product/1",
      sourceLocationId: "location/1",
      quantity: 2,
      layoutId: "layout/1" as const,
      condition: "ready" as const,
      notes: "Montiert"
    };

    await api.accessoryProducts("Tillig & TT");
    await api.createAccessoryProduct(productInput);
    await api.accessoryProduct("product/1");
    await api.updateAccessoryProduct("product/1", productInput);
    await api.storageLocations();
    await api.createStorageLocation(locationInput);
    await api.updateStorageLocation("location/1", locationInput);
    await api.accessoryStock("product/1");
    await api.adjustAccessoryStock("product/1", { locationId: "location/1", delta: 5 });
    await api.accessoryAssets("product/1");
    await api.createAccessoryAsset("product/1", assetInput);
    await api.updateAccessoryAsset("asset/1", assetInput);
    await api.accessoryAllocationSummary("product/1");
    await api.accessoryReservations("product/1");
    await api.createAccessoryReservation(reservationInput);
    await api.cancelAccessoryReservation("reservation/1");
    await api.accessoryInstallations("product/1");
    await api.createAccessoryInstallation(installationInput);
    await api.removeAccessoryInstallation("installation/1", {
      disposition: "stored",
      storageLocationId: "location/1",
      notes: "Ausgebaut"
    });
    await api.updateAccessoryInstallationCondition("installation/1", { condition: "maintenance_due" });

    expectRequests([
      ["GET", "/api/v1/accessory-products?query=Tillig%20%26%20TT"],
      ["POST", "/api/v1/accessory-products", productInput],
      ["GET", "/api/v1/accessory-products/product%2F1"],
      ["PUT", "/api/v1/accessory-products/product%2F1", productInput],
      ["GET", "/api/v1/storage-locations"],
      ["POST", "/api/v1/storage-locations", locationInput],
      ["PUT", "/api/v1/storage-locations/location%2F1", locationInput],
      ["GET", "/api/v1/accessory-products/product%2F1/stock"],
      ["POST", "/api/v1/accessory-products/product%2F1/stock-adjustments", { locationId: "location/1", delta: 5 }],
      ["GET", "/api/v1/accessory-products/product%2F1/assets"],
      ["POST", "/api/v1/accessory-products/product%2F1/assets", assetInput],
      ["PUT", "/api/v1/accessory-assets/asset%2F1", assetInput],
      ["GET", "/api/v1/accessory-products/product%2F1/allocation-summary"],
      ["GET", "/api/v1/accessory-reservations?productId=product%2F1"],
      ["POST", "/api/v1/accessory-reservations", reservationInput],
      ["POST", "/api/v1/accessory-reservations/reservation%2F1/cancel"],
      ["GET", "/api/v1/accessory-installations?productId=product%2F1"],
      ["POST", "/api/v1/accessory-installations", installationInput],
      ["POST", "/api/v1/accessory-installations/installation%2F1/remove", {
        disposition: "stored", storageLocationId: "location/1", notes: "Ausgebaut"
      }],
      ["PUT", "/api/v1/accessory-installations/installation%2F1/condition", {
        condition: "maintenance_due"
      }]
    ]);
  });

  function expectRequests(expected: Array<[string, string, object?]>) {
    expect(fetchMock).toHaveBeenCalledTimes(expected.length);
    expected.forEach(([method, path, body], index) => {
      const [url, init] = fetchMock.mock.calls[index];
      expect(url).toBe(path);
      expect(init.method ?? "GET").toBe(method);
      expect(init.body).toBe(body === undefined ? undefined : JSON.stringify(body));
      if (method === "GET") {
        expect(init.headers["X-CSRF-Token"]).toBeUndefined();
      } else {
        expect(init.headers["X-CSRF-Token"]).toBe("client-test-token");
      }
    });
  }
});
