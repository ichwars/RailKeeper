"""Synthetic collection created through the real API, never from production data."""

import hashlib
import struct
import zlib


def sample_png():
    def chunk(kind, payload):
        return struct.pack(">I", len(payload)) + kind + payload + struct.pack(">I", zlib.crc32(kind + payload))
    return (b"\x89PNG\r\n\x1a\n" + chunk(b"IHDR", struct.pack(">IIBBBBB", 2, 2, 8, 2, 0, 0, 0))
            + chunk(b"IDAT", zlib.compress(b"\0" + b"\x28\x80\x50" * 2 + b"\0" + b"\x28\x80\x50" * 2))
            + chunk(b"IEND", b""))


def seed(client):
    post = lambda path, data: client.request("POST", "/api/v1/" + path, data, 201)
    base = {"manufacturer": "Roco", "articleNumber": "SYNTHETIC-218", "gauge": "H0",
            "category": "Lokomotive", "gattung": "Diesellokomotive", "digital": True,
            "decoderType": "Unbekannter Decoder", "condition": "ready", "qrCodeEnabled": True}
    vehicles = [post("vehicles", {**base, "inventoryNumber": f"REF-00{i}",
                                 "name": f"Referenzlok {i}", "purchasePrice": price,
                                 "listPrice": "199.00", "digitalDecoderNumber": str(20 + i)})
                for i, price in [(1, "100.00"), (2, "125.00")]]
    vehicle_set = post("vehicle-sets", {
        "set": {"name": "Referenzset", "manufacturer": "Roco", "articleNumber": "SYNTHETIC-SET",
                "gauge": "H0", "category": "Wagen", "gattung": "Reisezugwagen", "purchasePrice": "175.00"},
        "members": [{"name": "Referenzwagen A", "additionalInfo": "Individuelle Änderung A"},
                    {"name": "Referenzwagen B", "additionalInfo": "Abweichender Zustand B"}],
    })
    first = "vehicles/" + vehicles[0]["id"]
    maintenance = post(first + "/maintenance", {"kind": "Reinigung", "status": "geplant",
                                               "dueDate": "2026-12-01", "notes": "Referenzwartung"})
    png = sample_png()
    image = client.upload("/api/v1/" + first + "/images", "referenz.png", png, "image/png",
                          {"title": "Referenzbild", "isPrimary": "true", "maintenanceId": maintenance["id"]})
    attachment = client.upload("/api/v1/" + first + "/attachments", "anleitung.txt",
                               b"Synthetic manual: no real manufacturer data.\n", "text/plain")
    cv_file = client.upload("/api/v1/" + first + "/cv-files", "referenz.csv", b"CV;Value\n1;3\n8;151\n",
                           "text/csv", {"decoderProfile": "Legacy-Profil"})
    for key, direction, inverted in [("F0", None, False), ("F1", True, None), ("F3", False, True)]:
        client.request("PUT", "/api/v1/" + first + "/functions/" + key,
                       {"name": "Referenz " + key, "functionType": "standard", "mode": "dauer",
                        "directionDependent": direction, "inverted": inverted, "flagsVersion": 2,
                        "notes": "Herkunft: synthetischer Referenzbestand"})
    cv_base = {"decoderProfile": "Legacy-Profil", "description": "Referenzwert", "sourceFileId": cv_file["id"],
               "sourceLine": 2, "sourceExportedAt": "2026-10-20T08:30:00"}
    cv = post(first + "/cv-values", {**cv_base, "cvNumber": 1, "value": 3})
    client.request("PUT", "/api/v1/" + first + "/cv-values/" + cv["id"],
                   {**cv_base, "cvNumber": 1, "value": 5})
    for indices, value in [({}, 10), ({"indexCV31": 0, "indexCV32": 0}, 11),
                           ({"indexCV31": 1, "indexCV32": 0}, 12)]:
        post(first + "/cv-values", {**cv_base, **indices, "cvNumber": 257, "value": value})
    post(first + "/cv-values", {"decoderProfile": "Zweites Profil", "cvNumber": 1, "value": 7})
    client.request("POST", "/api/v1/" + first + "/cv-values/import",
                   {"values": [{**cv_base, "cvNumber": 8, "value": 151, "sourceLine": 3}]})
    client.request("POST", "/api/v1/" + first + "/external-mappings",
                   {"provider": "ecos", "externalId": "ref-1001",
                    "externalName": "Referenzlok", "externalAddress": "21"})
    post(first + "/spare-parts", {"articleNumber": "REF-SPARE", "description": "Referenzersatzteil",
                                  "price": "3.50", "url": "https://example.test/spare"})
    root = post("storage-locations", {"name": "Referenzwerkstatt"})
    location = post("storage-locations", {"name": "Schublade A", "parentId": root["id"]})
    product = post("accessory-products", {"manufacturer": "Referenz", "name": "Referenzbeleuchtung",
                                          "category": "Elektronik", "trackingMode": "quantity"})
    product_path = "accessory-products/" + product["id"]
    client.request("POST", "/api/v1/" + product_path + "/stock-adjustments",
                   {"locationId": location["id"], "delta": 10})
    post("accessory-reservations", {"productId": product["id"], "locationId": location["id"],
                                    "quantity": 2, "vehicleId": vehicles[0]["id"]})
    installation = post("accessory-installations", {"productId": product["id"],
                         "sourceLocationId": location["id"], "quantity": 1,
                         "vehicleId": vehicles[1]["id"], "condition": "ready"})
    individual = post("accessory-products", {"manufacturer": "Referenz", "name": "Referenzdecoder",
                                              "category": "Decoder", "trackingMode": "individual"})
    asset = post("accessory-products/" + individual["id"] + "/assets",
                 {"inventoryNumber": "REF-Z-001", "condition": "maintenance_due", "lifecycle": "maintenance",
                  "storageLocationId": location["id"]})
    client.upload("/api/v1/" + product_path + "/documents", "beleg.txt", b"Synthetic receipt\n", "text/plain",
                  {"category": "invoice"})
    event = post("exhibition-lists", {"designation": "Referenzfahrtag", "date": "2026-12-12",
                                      "endDate": "2026-12-13", "location": "Referenzraum"})
    entry = post("exhibition-lists/" + event["id"] + "/entries",
                 {"vehicleId": vehicles[0]["id"], "owner": "Referenzverein", "locomotiveName": "Referenzlok 1",
                  "decoderNumber": "21", "interfaceName": "Referenz-DCC", "availability": "available"})
    profile = post("data-transfer/profiles", {"name": "Referenzexport", "direction": "export", "format": "railkeeper-json",
                                               "areas": ["vehicles"], "options": {}})
    job = post("data-transfer/jobs/export", {"profileId": profile["id"]})
    exported = client.request("POST", "/api/v1/data-transfer/jobs/" + job["id"] + "/execute")
    if exported["job"]["state"] != "completed":
        raise AssertionError("Reference transfer export did not complete")
    return {"vehicleId": vehicles[0]["id"], "secondVehicleId": vehicles[1]["id"],
            "setId": vehicle_set["id"], "maintenanceId": maintenance["id"], "imageId": image["id"],
            "attachmentId": attachment["id"], "cvFileId": cv_file["id"], "productId": product["id"],
            "assetId": asset["id"], "installationId": installation["id"], "eventId": event["id"],
            "entryId": entry["id"], "imageSHA256": hashlib.sha256(png).hexdigest(),
            "artifactId": exported["artifact"]["id"], "artifactSHA256": exported["artifact"]["sha256"]}


def observations(client, fixture):
    vehicle = "/api/v1/vehicles/" + fixture["vehicleId"]
    endpoints = [vehicle, "/api/v1/vehicles", "/api/v1/vehicle-sets/" + fixture["setId"],
                 vehicle + "/cv-values", vehicle + "/functions", vehicle + "/maintenance",
                 "/api/v1/overview/valuation", "/api/v1/storage-locations",
                 "/api/v1/accessory-products/" + fixture["productId"] + "/stock",
                 "/api/v1/accessory-products/" + fixture["productId"] + "/usage-history",
                 "/api/v1/exhibition-lists/" + fixture["eventId"] + "/workspace"]
    return {path: record_comparison_projection(client.request("GET", path)) for path in endpoints}


def record_comparison_projection(value):
    """Ignore only new computed vehicle/media/CV revisions in old-to-new API comparisons.

    All stored fields and other entity revisions remain compared. Raw DB/file manifests have no
    projection; transaction/HTTP tests separately verify these content-derived revision guards.
    """
    if isinstance(value, list):
        return [record_comparison_projection(item) for item in value]
    if isinstance(value, dict):
        cv = {"id", "vehicleId", "cvNumber", "value"}.issubset(value)
        vehicle = {"id", "inventoryNumber", "manufacturer", "name", "gauge", "qrCodeEnabled"}.issubset(value)
        image = {"id", "vehicleId", "url", "isPrimary", "sortOrder"}.issubset(value)
        attachment = {"id", "vehicleId", "fileName", "originalName", "sizeBytes"}.issubset(value)
        computed = {"revision"} if cv or image or attachment else set()
        if vehicle:
            computed.update({"revision", "imagesRevision"})
        return {key: record_comparison_projection(item) for key, item in value.items() if key not in computed}
    return value
