# Checkpoint

## Umgesetzte Pakete

- `cdcbffe`: Entwurf und Implementierungsplan
- `c36982f`: Austauschdomäne und strikte Schema-1-Validierung
- `c151bb3`: Migration 0055, Geometriesnapshots und Backupformat 14
- `89ad295`: Vorschau, Draftimport, Admin-Freigabe, Stilllegung und Audit
- `90ad341`: rollen- und CSRF-geschützte API mit OpenAPI-Vertrag
- `93f76fe`: Planerpanel und app-eigene Import-, Prüf- und Stilllegungsdialoge
- `c68efbe`: stabiler einmaliger Bibliotheks-Loader
- `009d2ec`: Hersteller aus importierten Bibliotheken in Planer und Materialfluss erhalten

## Sicherheitsgrenzen

- Vorschau schreibt nicht.
- Import setzt externe Prüfstatus immer auf `draft` zurück.
- Freigabe und Stilllegung verlangen Admin, CSRF und explizite Bestätigung.
- Importgrenze: 4 MiB, ein Dokument, höchstens 500 Definitionen.
- Bibliotheksversionen werden nicht überschrieben.
- Bestehende Planobjekte lesen unveränderliche Geometriesnapshots.
