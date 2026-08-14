# Interaktiver Anlagenzwilling, Paket D: Verlauf und Abnahme

## Ziel

Paket D vervollständigt den Inspector um den Einbau- und Zustandsverlauf der ausgewählten
technischen Position. Bestehende Stage-2-Backup-, Migrations- und Rollenprüfungen werden gegen die
Designabnahme bestätigt. Es entstehen keine neuen parallelen Historientabellen.

## Umsetzung

1. Der Inspector lädt beim Öffnen die vorhandene Produkthistorie verzögert nach und filtert sie über
   `technicalPositionId` auf die gewählte Position.
2. Reservierung, Einbau, Zustandsänderung und Ausbau erscheinen als kompakte, zeitlich sortierte
   Verlaufsliste mit Menge, Zustand beziehungsweise Verbleib und Zeitpunkt.
3. Laden, Leerzustand und Fehler verwenden app-eigene Zustände und bleiben auf den Inspector
   begrenzt.
4. Positionen ohne verknüpften Artikel zeigen einen erklärenden Leerzustand und lösen keine unnötige
   API-Anfrage aus.
5. Deutsch und Englisch sowie schmale Ansichten werden im bestehenden Designsystem ergänzt.

## Verifikation

- Frontendtests für Lazy Loading, Positionsfilter, Verlaufstypen, Leer- und Fehlerzustand
- Bestätigung der vorhandenen Backup-Version-4-, Restore-, Migrations-, Rollen- und CSRF-Tests
- vollständige Backend- und Frontendtests, Build und lokale Browserprüfung
- ausschließlich lokale Commits, kein Push, keine PR und kein Merge
