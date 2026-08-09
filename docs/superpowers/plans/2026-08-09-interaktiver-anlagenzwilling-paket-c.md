# Interaktiver Anlagenzwilling, Paket C: Bearbeitungsmodus und Konturen

## Ziel

Paket C ergänzt die reine Leseansicht um einen ausdrücklich aktivierten Bearbeitungsmodus. Planner
und Admin können technische Positionen und Konturpunkte direkt in der SVG-Arbeitsfläche verschieben.
Ohne Bearbeitungsmodus bleiben Maus- und Touchgesten rein lesend.

## Umsetzung

1. Ein versionsgeprüfter Endpunkt ersetzt die Kontur einer Anlageneinheit atomar und schreibt einen
   Audit-Eintrag. Mindestens drei endliche Punkte sind erforderlich.
2. Die Twin-Leseansicht liefert lokale und transformierte Konturen sowie die Versionsstände von
   Einheiten und technischen Positionen.
3. Der Frontend-Client und OpenAPI werden um Konturtypen und den Schreibendpunkt ergänzt.
4. Nur Planner und Admin sehen die Aktion `Bearbeiten`. Erst nach Aktivierung werden Marker und
   Konturpunkte ziehbar.
5. Änderungen werden nach Ende der Zeigergeste kurz verzögert gespeichert. Währenddessen ist das
   betroffene Element gesperrt.
6. Bei Versionskonflikten bleibt der lokale Entwurf sichtbar. Der Benutzer kann den Serverstand neu
   laden oder den lokalen Entwurf bewusst verwerfen.
7. Maus, Tastatur und Touch bleiben getrennt: Auswahl funktioniert immer, Verschieben ausschließlich
   im Bearbeitungsmodus.

## Verifikation

- Service-, Repository-, Rollen-, CSRF- und Konflikttests
- Client- und OpenAPI-Vertragstests
- Frontendtests für Rollen, Aktivierung, Verschieben, Speichern und Konflikte
- vollständige Backend- und Frontendtests, Build und reale lokale Browserprüfung
- ausschließlich lokale Commits, kein Push, keine PR und kein Merge

