"""Reviewed feature boundaries and a frozen baseline of all registered routes.

This is source coverage, not proof that every route was exercised in a browser.
Behavioral evidence comes from the referenced tests and the isolated baseline run.
"""

import argparse
import hashlib
import json
import re
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
ROUTES_SHA256 = "cae085fbfa4ae00906babdd05e3ca85ccbbbd677f1f9d89ccf4f6be6e2abae53"

# Longest matching prefix wins. New top-level areas must be classified explicitly.
FEATURES = {
    "setup-auth": {
        "prefixes": ["/setup", "/auth", "/users", "/roles", "/sessions"],
        "ui": "Setup, Anmeldung, Konten/Rollen, TOTP, Sitzungen",
        "storage": "users, roles, user_roles, sessions, rate_limit_attempts; installation-local",
        "tests": ["backend/internal/api/smoke_test.go", "backend/internal/application/auth_roles_test.go"],
    },
    "system": {
        "prefixes": ["/health", "/version", "/system", "/profile"],
        "ui": "Einstellungen, Speicherpfad, Update, SMTP, Drucker, Audit, Profil",
        "storage": "app_settings, user_settings, audit_logs; installation-local",
        "tests": ["backend/internal/api/storage_info_handlers_test.go", "backend/internal/api/version_test.go"],
    },
    "bug-reporting": {
        "prefixes": ["/bug-reports"], "ui": "Fehlerbericht mit optionaler Diagnose/Bild/Kontakt",
        "storage": "bug_report_submissions; installation-local",
        "tests": ["backend/internal/api/bug_report_handlers_test.go"],
    },
    "vehicles-sets": {
        "prefixes": ["/vehicles", "/vehicle-sets", "/inventory-number-schemes"],
        "ui": "Bestand, Einzelstück, Set, Inventarnummer, Bericht und QR",
        "storage": "vehicles, vehicle_sets, vehicle_set_members, inventory_number_history, vehicle_external_mappings",
        "tests": ["backend/internal/application/vehicle_sets_test.go", "frontend/src/features/vehicles/vehicleReports.test.ts"],
    },
    "media": {
        "prefixes": ["/vehicles/{id}/images", "/vehicles/{id}/attachments"],
        "ui": "Uploads, Vorschau, Download, Hauptbild, Verknüpfungen",
        "storage": "vehicle_images, vehicle_attachments, file_blobs and referenced files",
        "tests": ["backend/internal/application/vehicles_test.go", "backend/internal/application/backup_test.go"],
    },
    "workshop": {
        "prefixes": ["/vehicles/{id}/maintenance", "/vehicles/{id}/spare-parts"],
        "ui": "Wartung, Zustand, Ersatzteile und Dokumentauswertung",
        "storage": "vehicle_maintenance, vehicle_spare_parts",
        "tests": ["backend/internal/application/vehicles_test.go", "backend/internal/application/backup_test.go"],
    },
    "functions-cv": {
        "prefixes": ["/vehicles/{id}/functions", "/vehicles/{id}/cv-values", "/vehicles/{id}/cv-files", "/cv-files", "/decoder-library"],
        "ui": "Steuerung, Funktionsentwurf, CVs/Indexseiten, Decoderbibliothek, Kennlinie",
        "storage": "vehicle_functions, vehicle_cv_values, vehicle_cv_value_history, vehicle_cv_files, master_data_entries",
        "tests": ["backend/internal/application/vehicle_function_changes_test.go", "backend/internal/application/vehicle_cv_provenance_test.go"],
    },
    "accessories": {
        "prefixes": ["/accessory-products", "/accessory-assets", "/accessory-reservations", "/accessory-installations", "/storage-locations"],
        "ui": "Artikel, Lager, Mengen/Einzelstücke, Kaufbelege, Reservierung/Einbau und Historie",
        "storage": "accessory_* and storage_locations",
        "tests": ["backend/internal/api/accessory_handlers_test.go", "backend/internal/api/accessory_allocation_handlers_test.go"],
    },
    "exhibition": {
        "prefixes": ["/exhibition-lists"], "ui": "Fahrtag, Teilnehmer, Konflikte, Sperre, Druck",
        "storage": "exhibition_lists, exhibition_entries, exhibition_conflict_exceptions",
        "tests": ["backend/internal/api/exhibition_workspace_test.go"],
    },
    "transfer-search": {
        "prefixes": ["/data-transfer", "/article-search"],
        "ui": "Import/Export, Profile, Vorschau, Konflikte, Jobs; Artikelsuche",
        "storage": "transfer profiles in app backup; job history/artifacts installation-local",
        "tests": ["backend/internal/application/data_transfer_export_test.go", "backend/internal/application/article_search_test.go"],
    },
    "master-data": {
        "prefixes": ["/master-data", "/master-data-all", "/master-data-relations"],
        "ui": "Stammdaten, Aktivierung, Beziehungen, eigene Symbole, Import/Export",
        "storage": "master_data_entries, master_data_relations",
        "tests": ["backend/internal/api/master_data_api_test.go"],
    },
    "overview": {
        "prefixes": ["/overview"], "ui": "Übersicht, Wartung, Bestand und Bewertung",
        "storage": "derived from inventory; dashboard preferences installation-local",
        "tests": ["backend/internal/application/overview_valuation_test.go", "frontend/src/features/overview/OverviewView.test.tsx"],
    },
    "backup-restore": {
        "prefixes": ["/backup"], "ui": "Sicherung, Vorprüfung, Bestätigung und Wiederherstellung",
        "storage": "application tables, transfer profiles, blobs and referenced files; local auth excluded",
        "tests": ["backend/internal/application/backup_test.go", "backend/internal/api/smoke_test.go"],
    },
    "command-stations": {
        "prefixes": ["/ecos", "/digital-centers", "/vehicles/{id}/ecos-functions", "/vehicles/{id}/ecos-image"],
        "ui": "Zentralen, Lesen/Vergleichen, Freigaben, ECoS-Bild und Dateisicherung",
        "storage": "vehicle mappings/CVs in app backup; workspace sessions/grants installation-local",
        "tests": ["backend/internal/application/digital_centers_test.go", "backend/internal/api/ecos_image_handlers_test.go"],
    },
    "legacy-archive": {
        "prefixes": [], "ui": "Keine aktive Gleisplanung; Altbestand und Restore-Kompatibilität",
        "storage": "retired_allocation_records and dormant planning tables",
        "tests": ["backend/internal/infrastructure/legacy_layout_retirement_test.go"],
    },
}


def inventory():
    source = (ROOT / "backend/internal/api/routes.go").read_text(encoding="utf-8")
    routes = sorted(re.findall(r'\{http.Method(\w+), "([^\"]+)", routeAccess(\w+)', source))
    actual = hashlib.sha256(json.dumps(routes, separators=(",", ":")).encode()).hexdigest()
    if actual != ROUTES_SHA256:
        raise AssertionError("Registered routes/access changed: review the feature baseline before updating its digest")
    result = {name: {**feature, "routes": []} for name, feature in FEATURES.items()}
    for feature in result.values():
        for relative in feature["tests"]:
            if not (ROOT / relative).is_file():
                raise AssertionError("Missing evidence source: " + relative)
    for method, path, access in routes:
        short = path.removeprefix("/api/v1")
        candidates = [(len(prefix), name) for name, feature in FEATURES.items() for prefix in feature["prefixes"]
                      if short == prefix or short.startswith(prefix + "/")]
        if not candidates:
            raise AssertionError("Unmapped route: " + path)
        _, name = max(candidates)
        result[name]["routes"].append({"method": method.upper(), "path": path, "access": access})
    return {"routeCount": len(routes), "routesSHA256": actual, "features": result,
            "evidence": "source-mapped; behavioral/platform/hardware evidence must be recorded separately"}


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", type=Path)
    args = parser.parse_args()
    matrix = inventory()
    if args.output:
        args.output.write_text(json.dumps(matrix, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(f"Feature matrix: {len(FEATURES)} areas, {matrix['routeCount']} registered routes; evidence sources exist")
