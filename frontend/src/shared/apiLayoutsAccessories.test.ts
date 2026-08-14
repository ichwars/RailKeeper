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
        rotationDegrees: 90
      }]
    };

    await api.layouts();
    await api.createLayout(layoutInput);
    await api.layout("layout/1");
    await api.layoutTwin("layout/1", { configurationId: "configuration/1" });
    await api.layoutTwin("layout/1", { unitId: "unit/1" });
    await api.updateLayout("layout/1", { ...layoutInput, expectedVersion: 3 });
    await api.layoutUnits("layout/1");
    await api.createLayoutUnit("layout/1", unitInput);
    await api.updateLayoutUnit("unit/1", { ...unitInput, expectedVersion: 2 });
    const outlineInput = { expectedVersion: 2, points: [
      { xMm: 0, yMm: 0 }, { xMm: 100, yMm: 0 }, { xMm: 100, yMm: 50 }
    ] };
    await api.updateLayoutUnitOutline("unit/1", outlineInput);
    const positionInput = {
      label: "Signal A",
      kind: "signal" as const,
      positionXMm: 125,
      positionYMm: 40,
      rotationDegrees: 90,
      productId: "product/1",
      description: "Einfahrt"
    };
    await api.layoutTechnicalPositions("unit/1");
    await api.createLayoutTechnicalPosition("unit/1", positionInput);
    await api.updateLayoutTechnicalPosition("position/1", { ...positionInput, expectedVersion: 2 });
    await api.layoutConfigurations("layout/1");
    await api.createLayoutConfiguration("layout/1", configurationInput);
    await api.updateLayoutConfiguration("configuration/1", configurationInput);
    await api.layoutConfigurationPortAnalysis("configuration/1");
    await api.previewLayoutConfigurationUnitSnap("configuration/1", {
      unitId: "unit/1", positionXMm: 10, positionYMm: 20, rotationDegrees: 90
    });
    await api.planVariants("unit/1");
    await api.createPlanVariant("unit/1", { name: "Sommer", description: "Variante" });
    await api.createPlanRevision("variant/1", { baseRevisionId: "revision/1" });
    await api.submitPlanRevision("revision/1", 4);
    await api.publishPlanRevision("revision/1", 5);
    const trackObjectInput = {
      geometryId: "geometry/1", positionXMm: 100, positionYMm: 50, rotationDegrees: 15,
      elevationStartMm: 0, elevationEndMm: 4.15
    };
    await api.trackGeometries("TT");
    await api.trackPlan("revision/1");
    await api.trackPlanAnalysis("revision/1");
    await api.trackPlanChangePreview("revision/1");
    await api.reserveTrackPlanMaterials("revision/1", {
      confirmed: true,
      items: [{
        trackObjectId: "object/1", productId: "product/1", locationId: "location/1",
        expectedObjectVersion: 2
      }]
    });
    await api.createPlanTrackObject("revision/1", trackObjectInput);
    await api.updatePlanTrackObject("object/1", {
      positionXMm: 110, positionYMm: 55, rotationDegrees: 30,
      elevationStartMm: 1, elevationEndMm: 5.15, expectedVersion: 2
    });
    await api.deletePlanTrackObject("object/1", 3);

    expectRequests([
      ["GET", "/api/v1/layouts"],
      ["POST", "/api/v1/layouts", layoutInput],
      ["GET", "/api/v1/layouts/layout%2F1"],
      ["GET", "/api/v1/layouts/layout%2F1/twin?configurationId=configuration%2F1"],
      ["GET", "/api/v1/layouts/layout%2F1/twin?unitId=unit%2F1"],
      ["PUT", "/api/v1/layouts/layout%2F1", { ...layoutInput, expectedVersion: 3 }],
      ["GET", "/api/v1/layouts/layout%2F1/units"],
      ["POST", "/api/v1/layouts/layout%2F1/units", unitInput],
      ["PUT", "/api/v1/layout-units/unit%2F1", { ...unitInput, expectedVersion: 2 }],
      ["PUT", "/api/v1/layout-units/unit%2F1/outline", outlineInput],
      ["GET", "/api/v1/layout-units/unit%2F1/technical-positions"],
      ["POST", "/api/v1/layout-units/unit%2F1/technical-positions", positionInput],
      ["PUT", "/api/v1/layout-technical-positions/position%2F1", { ...positionInput, expectedVersion: 2 }],
      ["GET", "/api/v1/layouts/layout%2F1/configurations"],
      ["POST", "/api/v1/layouts/layout%2F1/configurations", configurationInput],
      ["PUT", "/api/v1/layout-configurations/configuration%2F1", configurationInput],
      ["GET", "/api/v1/layout-configurations/configuration%2F1/port-analysis"],
      ["POST", "/api/v1/layout-configurations/configuration%2F1/unit-snap-preview", {
        unitId: "unit/1", positionXMm: 10, positionYMm: 20, rotationDegrees: 90
      }],
      ["GET", "/api/v1/layout-units/unit%2F1/plan-variants"],
      ["POST", "/api/v1/layout-units/unit%2F1/plan-variants", { name: "Sommer", description: "Variante" }],
      ["POST", "/api/v1/plan-variants/variant%2F1/revisions", { baseRevisionId: "revision/1" }],
      ["POST", "/api/v1/plan-revisions/revision%2F1/submit", { expectedVersion: 4 }],
      ["POST", "/api/v1/plan-revisions/revision%2F1/publish", { expectedVersion: 5 }],
      ["GET", "/api/v1/track-geometries?gauge=TT"],
      ["GET", "/api/v1/plan-revisions/revision%2F1/track-plan"],
      ["GET", "/api/v1/plan-revisions/revision%2F1/track-analysis"],
      ["GET", "/api/v1/plan-revisions/revision%2F1/track-change-preview"],
      ["POST", "/api/v1/plan-revisions/revision%2F1/track-reservations", {
        confirmed: true,
        items: [{
          trackObjectId: "object/1", productId: "product/1", locationId: "location/1",
          expectedObjectVersion: 2
        }]
      }],
      ["POST", "/api/v1/plan-revisions/revision%2F1/track-objects", trackObjectInput],
      ["PUT", "/api/v1/plan-track-objects/object%2F1",
		{ positionXMm: 110, positionYMm: 55, rotationDegrees: 30,
		  elevationStartMm: 1, elevationEndMm: 5.15, expectedVersion: 2 }],
      ["DELETE", "/api/v1/plan-track-objects/object%2F1?expectedVersion=3"]
    ]);
  });


  it("encodes article filters as repeated query parameters with stable sorting", async () => {
    await api.accessoryArticles({
      query: "Tillig & TT",
      articleTypes: ["track", "lighting"],
      manufacturer: "Märklin / Trix",
      gauges: ["TT", "H0 + H0e"],
      statuses: ["available", "maintenance_due"],
      locationId: "room/1",
      sort: "updatedAt",
      direction: "desc"
    });

    expect(fetchMock.mock.calls[0]?.[0]).toBe(
      "/api/v1/accessory-products?query=Tillig+%26+TT&articleType=track&articleType=lighting" +
      "&manufacturer=M%C3%A4rklin+%2F+Trix&gauge=TT&gauge=H0+%2B+H0e" +
      "&status=available&status=maintenance_due&locationId=room%2F1&sort=updatedAt&direction=desc"
    );
  });

  it("calls article purchase, transfer, individualization, archive, restore, delete, and history routes", async () => {
    const purchase = {
      purchasedAt: "2026-08-08",
      supplier: "Fachhändler",
      quantity: 2,
      unitPrice: "12.90",
      currency: "EUR",
      invoiceNumber: "R-1",
      warrantyUntil: "2028-08-08",
      storageLocationId: "location/1",
      bookToStock: true,
      notes: "Vereinsbedarf"
    };
    const transfer = {
      fromLocationId: "location/1",
      toLocationId: "location/2",
      quantity: 1,
      note: "Umbuchung"
    };
    const individualization = {
      locationId: "location/2",
      asset: {
        inventoryNumber: "RK-1",
        condition: "ready" as const,
        lifecycle: "stored" as const,
        storageLocationId: "location/2"
      }
    };

    await api.accessoryPurchases("product/1");
    await api.createAccessoryPurchase("product/1", purchase);
    await api.accessoryStockMovements("product/1");
    await api.transferAccessoryStock("product/1", transfer);
    await api.individualizeAccessoryProduct("product/1", individualization);
    await api.archiveAccessoryProduct("product/1");
    await api.restoreAccessoryProduct("product/1");
    await api.deleteAccessoryProduct("product/1");
    await api.accessoryUsageHistory("product/1");

    expectRequests([
      ["GET", "/api/v1/accessory-products/product%2F1/purchases"],
      ["POST", "/api/v1/accessory-products/product%2F1/purchases", purchase],
      ["GET", "/api/v1/accessory-products/product%2F1/stock-movements"],
      ["POST", "/api/v1/accessory-products/product%2F1/stock-transfers", transfer],
      ["POST", "/api/v1/accessory-products/product%2F1/individualizations", individualization],
      ["POST", "/api/v1/accessory-products/product%2F1/archive"],
      ["POST", "/api/v1/accessory-products/product%2F1/restore"],
      ["DELETE", "/api/v1/accessory-products/product%2F1"],
      ["GET", "/api/v1/accessory-products/product%2F1/usage-history"]
    ]);
  });

  it("keeps typed stock, asset, reservation, and installation command routes covered", async () => {
    const asset = { inventoryNumber: "RK-1", serialNumber: "S-1", condition: "ready" as const,
      lifecycle: "stored" as const, storageLocationId: "location/1", purchaseDate: "2026-08-08",
      purchasePrice: "12.50", warrantyUntil: "2028-08-08", notes: "Geprüft" };
    const technical = { placement: "Bahnhof West", digitalAddress: "17", decoderOutput: "A2",
      connection: "J3", wiringNotes: "blau/gelb" };
    const reservation = { productId: "product/1", layoutId: "layout/1", locationId: "location/1",
      quantity: 2, note: "Aufbau", ...technical };
    const installation = { productId: "product/1", layoutId: "layout/1", sourceLocationId: "location/1",
      quantity: 2, condition: "ready" as const, notes: "Montiert", ...technical };

    await api.accessoryStock("product/1");
    await api.adjustAccessoryStock("product/1", { locationId: "location/1", delta: 2 });
    await api.accessoryAssets("product/1");
    await api.createAccessoryAsset("product/1", asset);
    await api.updateAccessoryAsset("asset/1", asset);
    await api.accessoryReservations("product/1");
    await api.createAccessoryReservation(reservation);
    await api.cancelAccessoryReservation("reservation/1");
    await api.accessoryInstallations("product/1");
    await api.createAccessoryInstallation(installation);
    await api.updateAccessoryInstallationCondition("installation/1", { condition: "defective" });
    await api.removeAccessoryInstallation("installation/1", { disposition: "retired", notes: "Verbraucht" });

    expectRequests([
      ["GET", "/api/v1/accessory-products/product%2F1/stock"],
      ["POST", "/api/v1/accessory-products/product%2F1/stock-adjustments", { locationId: "location/1", delta: 2 }],
      ["GET", "/api/v1/accessory-products/product%2F1/assets"],
      ["POST", "/api/v1/accessory-products/product%2F1/assets", asset],
      ["PUT", "/api/v1/accessory-assets/asset%2F1", asset],
      ["GET", "/api/v1/accessory-reservations?productId=product%2F1"],
      ["POST", "/api/v1/accessory-reservations", reservation],
      ["POST", "/api/v1/accessory-reservations/reservation%2F1/cancel"],
      ["GET", "/api/v1/accessory-installations?productId=product%2F1"],
      ["POST", "/api/v1/accessory-installations", installation],
      ["PUT", "/api/v1/accessory-installations/installation%2F1/condition", { condition: "defective" }],
      ["POST", "/api/v1/accessory-installations/installation%2F1/remove",
        { disposition: "retired", notes: "Verbraucht" }]
    ]);
  });

  it("keeps track library preview, import, export, and review routes typed", async () => {
    const libraryPackage = {
      format: "railkeeper.track-library" as const,
      schemaVersion: 1 as const,
      library: {
        manufacturer: "Kühn", trackSystem: "TT", gauge: "TT", scale: "1:120",
        version: "2026.1", sourceUrl: "https://example.com/catalogue.pdf", status: "verified" as const
      },
      definitions: []
    };
    await api.trackLibraries();
    await api.exportTrackLibrary("library/1");
    await api.previewTrackLibraryImport(libraryPackage);
    await api.importTrackLibrary({ confirmed: true, package: libraryPackage });
    await api.updateTrackLibraryStatus("library/1", {
      confirmed: true, status: "verified", verificationNote: "Katalog geprüft"
    });

    expectRequests([
      ["GET", "/api/v1/track-libraries"],
      ["GET", "/api/v1/track-libraries/library%2F1/export"],
      ["POST", "/api/v1/track-libraries/import/preview", libraryPackage],
      ["POST", "/api/v1/track-libraries/import", { confirmed: true, package: libraryPackage }],
      ["PUT", "/api/v1/track-libraries/library%2F1/status", {
        confirmed: true, status: "verified", verificationNote: "Katalog geprüft"
      }]
    ]);
  });

  it("uploads article documents as multipart form data without a content-type override", async () => {
    const file = new File(["invoice"], "Rechnung 1.pdf", { type: "application/pdf" });

    await api.uploadAccessoryDocument("product/1", {
      file,
      category: "invoice",
      description: "Originalrechnung",
      isPrimary: false
    });

    const [url, init] = fetchMock.mock.calls[0] ?? [];
    expect(url).toBe("/api/v1/accessory-products/product%2F1/documents");
    expect(init.method).toBe("POST");
    expect(init.body).toBeInstanceOf(FormData);
    if (!(init.body instanceof FormData)) throw new Error("expected multipart form data");
    const form = init.body;
    expect(form.get("file")).toBe(file);
    expect(form.get("category")).toBe("invoice");
    expect(form.get("description")).toBe("Originalrechnung");
    expect(form.get("isPrimary")).toBe("false");
    expect(init.headers["Content-Type"]).toBeUndefined();
    expect(init.headers["X-CSRF-Token"]).toBe("client-test-token");
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
