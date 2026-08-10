# Checkpoint

## Implementierung

- `5332f48` definiert Shape-Domäne und Validierung.
- `d9bafaf` persistiert freie Planobjekte über Migration `0054_free_plan_objects.sql`.
- `6e46326` integriert Revisionsklon, Änderungsvorschau und Backup-Version 13.
- `b10a59f` ergänzt Anwendung, HTTP-Routen und OpenAPI.
- `ee891ca` ergänzt API-Adapter, app-eigenen Dialog und SVG-Darstellung.
- `4ffaaa1` integriert Erstellen, Auswählen, Verschieben, Drehen, Bearbeiten und Löschen in den
  Planer.

## Grenzen

Freie Planobjekte erzeugen keine Gleisanschlüsse, Prüfhinweise, Höhen, Stücklistenpositionen oder
Reservierungen. Polygone, Bildimporte, Gruppierung, Layerreihenfolge und Inventarverknüpfungen
bleiben ausdrücklich außerhalb dieses Pakets.
