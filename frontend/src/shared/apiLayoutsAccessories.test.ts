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

    const currentArticle = completeTypedArticle();
    fetchMock
      .mockResolvedValueOnce(jsonResponse({ items: [], metrics: {}, filters: {} }))
      .mockResolvedValueOnce(jsonResponse(currentArticle))
      .mockResolvedValueOnce(jsonResponse(currentArticle))
      .mockResolvedValueOnce(jsonResponse(currentArticle))
      .mockResolvedValueOnce(jsonResponse(currentArticle));
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
      ["POST", "/api/v1/accessory-products", {
        ...productInput,
        articleType: "other",
        subtype: productInput.category,
        packageQuantity: 1,
        stockUnit: "piece",
        inventoryStrategy: productInput.trackingMode
      }],
      ["GET", "/api/v1/accessory-products/product%2F1"],
      ["GET", "/api/v1/accessory-products/product%2F1"],
      ["PUT", "/api/v1/accessory-products/product%2F1", {
        ...productInput,
        ean: currentArticle.ean,
        manufacturerStatus: currentArticle.manufacturerStatus,
        articleType: currentArticle.articleType,
        subtype: currentArticle.subtype,
        gauges: currentArticle.gauges,
        scale: currentArticle.scale,
        packageQuantity: currentArticle.packageQuantity,
        stockUnit: currentArticle.stockUnit,
        minimumStock: currentArticle.minimumStock,
        inventoryStrategy: currentArticle.inventoryStrategy,
        manufacturerUrl: currentArticle.manufacturerUrl,
        productUrl: currentArticle.productUrl,
        alternativeNumbers: currentArticle.alternativeNumbers,
        keywords: currentArticle.keywords,
        compatibilityNotes: currentArticle.compatibilityNotes,
        internalNotes: currentArticle.internalNotes,
        archived: currentArticle.archived,
        attributes: currentArticle.attributes
      }],
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

  it("calls article purchase, transfer, individualization, archive, restore, and history routes", async () => {
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
    await api.accessoryUsageHistory("product/1");

    expectRequests([
      ["GET", "/api/v1/accessory-products/product%2F1/purchases"],
      ["POST", "/api/v1/accessory-products/product%2F1/purchases", purchase],
      ["GET", "/api/v1/accessory-products/product%2F1/stock-movements"],
      ["POST", "/api/v1/accessory-products/product%2F1/stock-transfers", transfer],
      ["POST", "/api/v1/accessory-products/product%2F1/individualizations", individualization],
      ["POST", "/api/v1/accessory-products/product%2F1/archive"],
      ["POST", "/api/v1/accessory-products/product%2F1/restore"],
      ["GET", "/api/v1/accessory-products/product%2F1/usage-history"]
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

  it("adapts the article list result to the legacy array consumed by AccessoriesView", async () => {
    fetchMock.mockResolvedValueOnce(jsonResponse({
      items: [{
        id: "product-1",
        manufacturer: "Tillig",
        articleNumber: "83101",
        name: "Gerades Gleis",
        articleType: "track",
        subtype: "track:straight",
        gauges: ["TT"],
        inventoryStrategy: "quantity_later_individual",
        archived: false,
        owned: 3,
        available: 2,
        reserved: 1,
        installed: 0,
        locationNames: ["Schrank A"],
        hasUsageHistory: true,
        careHintCount: 0,
        updatedAt: "2026-08-08T12:00:00Z",
        attributes: []
      }],
      metrics: {
        articleCount: 1,
        articleTypeCount: 1,
        available: 2,
        locationCount: 1,
        reserved: 1,
        installed: 0,
        careHintCount: 0
      },
      filters: {
        manufacturers: ["Tillig"],
        articleTypes: ["track"],
        gauges: ["TT"],
        storageLocations: [{ id: "location-1", name: "Schrank A" }]
      }
    }));

    const products = await api.accessoryProducts("Tillig");

    expect(Array.isArray(products)).toBe(true);
    expect(products).toEqual([{
      id: "product-1",
      manufacturer: "Tillig",
      articleNumber: "83101",
      name: "Gerades Gleis",
      category: "track:straight",
      trackingMode: "quantity",
      createdAt: "2026-08-08T12:00:00Z",
      updatedAt: "2026-08-08T12:00:00Z"
    }]);
  });

  it("adapts reachable legacy detail and create responses without weakening new article DTOs", async () => {
    const article = {
      id: "product-1",
      manufacturer: "Tillig",
      articleNumber: "83101",
      name: "Gerades Gleis",
      category: "Gleismaterial",
      trackingMode: "quantity",
      description: "Modellgleis",
      manufacturerStatus: "available",
      articleType: "track",
      subtype: "track:straight",
      gauges: ["TT"],
      packageQuantity: 1,
      stockUnit: "piece",
      minimumStock: 0,
      inventoryStrategy: "quantity",
      alternativeNumbers: [],
      keywords: [],
      archived: false,
      attributes: [],
      createdAt: "2026-08-08T11:00:00Z",
      updatedAt: "2026-08-08T12:00:00Z"
    };
    fetchMock
      .mockResolvedValueOnce(jsonResponse(article))
      .mockResolvedValueOnce(jsonResponse(article));
    const input = {
      manufacturer: "Tillig",
      articleNumber: "83101",
      name: "Gerades Gleis",
      category: "Gleismaterial",
      trackingMode: "quantity" as const,
      description: "Modellgleis"
    };

    const detail = await api.accessoryProduct("product-1");
    const created = await api.createAccessoryProduct(input);

    const expected = {
      id: "product-1",
      manufacturer: "Tillig",
      articleNumber: "83101",
      name: "Gerades Gleis",
      category: "Gleismaterial",
      trackingMode: "quantity",
      description: "Modellgleis",
      createdAt: "2026-08-08T11:00:00Z",
      updatedAt: "2026-08-08T12:00:00Z"
    };
    expect(detail).toEqual(expected);
    expect(created).toEqual(expected);
    const expectedWrite = {
      ...input,
      articleType: "other",
      subtype: "Gleismaterial",
      packageQuantity: 1,
      stockUnit: "piece",
      inventoryStrategy: "quantity"
    };
    expect(fetchMock.mock.calls[1]?.[1].body).toBe(JSON.stringify(expectedWrite));
  });

  it("preserves the complete typed article when an unchanged legacy form saves", async () => {
    const current = completeTypedArticle();
    fetchMock
      .mockResolvedValueOnce(jsonResponse(current))
      .mockResolvedValueOnce(jsonResponse(current));

    const result = await api.updateAccessoryProduct("product/1", {
      manufacturer: current.manufacturer,
      articleNumber: current.articleNumber,
      name: current.name,
      category: current.category,
      trackingMode: current.trackingMode,
      description: ""
    });

    expect(fetchMock).toHaveBeenCalledTimes(2);
    expect(fetchMock.mock.calls[0]?.[0]).toBe("/api/v1/accessory-products/product%2F1");
    expect(fetchMock.mock.calls[0]?.[1].method ?? "GET").toBe("GET");
    expect(fetchMock.mock.calls[1]?.[0]).toBe("/api/v1/accessory-products/product%2F1");
    expect(fetchMock.mock.calls[1]?.[1].method).toBe("PUT");
    expect(JSON.parse(String(fetchMock.mock.calls[1]?.[1].body))).toEqual({
      manufacturer: current.manufacturer,
      articleNumber: current.articleNumber,
      name: current.name,
      category: current.category,
      trackingMode: current.trackingMode,
      description: current.description,
      ean: current.ean,
      manufacturerStatus: current.manufacturerStatus,
      articleType: current.articleType,
      subtype: current.subtype,
      gauges: current.gauges,
      scale: current.scale,
      packageQuantity: current.packageQuantity,
      stockUnit: current.stockUnit,
      minimumStock: current.minimumStock,
      inventoryStrategy: current.inventoryStrategy,
      manufacturerUrl: current.manufacturerUrl,
      productUrl: current.productUrl,
      alternativeNumbers: current.alternativeNumbers,
      keywords: current.keywords,
      compatibilityNotes: current.compatibilityNotes,
      internalNotes: current.internalNotes,
      archived: current.archived,
      attributes: current.attributes
    });
    expect(result).toMatchObject({
      id: current.id,
      category: current.category,
      trackingMode: current.trackingMode,
      description: current.description
    });
  });

  it("overlays deliberate legacy common-field edits while retaining typed classification", async () => {
    const current = completeTypedArticle();
    const updated = {
      ...current,
      manufacturer: "Tillig Modellbahnen",
      articleNumber: "83101-A",
      name: "Gerades Modellgleis",
      description: "Neue Beschreibung"
    };
    fetchMock
      .mockResolvedValueOnce(jsonResponse(current))
      .mockResolvedValueOnce(jsonResponse(updated));

    await api.updateAccessoryProduct("product/1", {
      manufacturer: updated.manufacturer,
      articleNumber: updated.articleNumber,
      name: updated.name,
      category: current.category,
      trackingMode: current.trackingMode,
      description: updated.description
    });

    const body = JSON.parse(String(fetchMock.mock.calls[1]?.[1].body));
    expect(body).toMatchObject({
      manufacturer: updated.manufacturer,
      articleNumber: updated.articleNumber,
      name: updated.name,
      description: updated.description,
      articleType: "track",
      subtype: "track:straight",
      inventoryStrategy: "quantity_later_individual",
      minimumStock: 4,
      attributes: current.attributes
    });
  });

  function completeTypedArticle() {
    return {
      id: "product-1",
      manufacturer: "Tillig",
      articleNumber: "83101",
      name: "Gerades Gleis",
      category: "Gleismaterial",
      trackingMode: "quantity" as const,
      description: "Bestehende Beschreibung",
      ean: "4012500831012",
      manufacturerStatus: "available" as const,
      articleType: "track" as const,
      subtype: "track:straight",
      gauges: ["TT"],
      scale: "1:120",
      packageQuantity: 6,
      stockUnit: "piece",
      minimumStock: 4,
      inventoryStrategy: "quantity_later_individual" as const,
      manufacturerUrl: "https://example.test/manufacturer",
      productUrl: "https://example.test/product",
      alternativeNumbers: ["83101-1"],
      keywords: ["Modellgleis", "gerade"],
      compatibilityNotes: "Tillig Modellgleis",
      internalNotes: "Vereinsbestand",
      archived: false,
      attributes: [
        { key: "length", kind: "number" as const, numberValue: 166, unit: "mm" },
        { key: "weathered", kind: "boolean" as const, booleanValue: false }
      ],
      primaryImageUrl: "/api/v1/accessory-products/product-1/documents/image-1/download",
      createdAt: "2026-08-08T11:00:00Z",
      updatedAt: "2026-08-08T12:00:00Z"
    };
  }

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

  function jsonResponse(value: unknown) {
    return {
      ok: true,
      status: 200,
      statusText: "OK",
      json: async () => value
    };
  }
});
