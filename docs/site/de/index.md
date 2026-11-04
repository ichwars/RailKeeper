---
layout: home
title: Modellbahnsammlung lokal verwalten
description: Fahrzeuge, Bilder, Unterlagen, Wartung und Decoderdaten an einem Ort. RailKeeper für Windows und Docker, mit deutschem und englischem Handbuch.
audience: reference
status: development
reviewedVersion: main
lastReviewed: 2026-11-04
hero:
  name: RailKeeper
  text: Deine Modellbahnsammlung an einem Ort
  tagline: Fahrzeuge, Bilder, Unterlagen, Wartung und Decoderdaten zusammen verwalten. Lokal auf deinem Rechner oder Heimserver. Kostenlos und quelloffen.
  image:
    src: /brand/railkeeper-mark.png
    alt: RailKeeper
  actions:
    - theme: brand
      text: Für Windows herunterladen
      link: https://github.com/ichwars/RailKeeper/releases/latest
    - theme: alt
      text: Mit Docker starten
      link: /de/try#docker
    - theme: alt
      text: Mit Beispieldaten ausprobieren
      link: /de/try#beispieldaten
features:
  - title: Den Bestand wiederfinden
    details: Loks, Wagen und Sets erfassen, nach Hersteller oder Inventarnummer suchen und eigene Tabellen über eine Importvorschau übernehmen.
    link: /de/guide/vehicles/
  - title: Unterlagen am Fahrzeug
    details: Bilder, Anleitungen und Beilagen direkt beim passenden Modell ablegen. Decoderdaten und Funktionen schrittweise ergänzen.
    link: /de/guide/vehicles/media
  - title: Wartung im Blick
    details: Reinigung, Reparaturen und Umbauten dokumentieren. Fällige Arbeiten und den Zustand deiner Fahrzeuge nachschlagen.
    link: /de/guide/vehicles/maintenance
---

## So arbeitest du mit RailKeeper

Du kannst mit einer einzelnen Lok anfangen oder eine vorhandene Liste importieren.
Eine Digitalzentrale brauchst du für die Bestandsverwaltung nicht.
Die Bilder zeigen künstliche Beispieldaten aus einer getrennten Demo-Installation.

### 1. Deine Sammlung durchsuchen

Suche nach einem Modell und öffne seine Daten. Fahrzeuge einer gemeinsamen Verpackung
lassen sich als Set zusammenfassen.

<GuideFigure name="vehicles" alt="Fahrzeugbestand mit einer Demo-Lok, Wagen und einem aufklappbaren Set." caption="Bestand, Filter und Fahrzeugsets auf einen Blick. Anklicken für die volle Größe." />

### 2. Bilder und Unterlagen zuordnen

Fotos und Dokumente gehören zum Fahrzeug. So findest du die passende Anleitung später
bei derselben Lok wieder.

<GuideFigure name="media" alt="Fahrzeugeditor mit den Bereichen Bilder und Beilagen sowie Schaltflächen zum Hochladen." caption="Bilder und Beilagen direkt am Fahrzeug verwalten." />

### 3. Die nächste Wartung vormerken

Halte fest, was du erledigt hast und welche Arbeit noch ansteht.

<GuideFigure name="maintenance" alt="Wartungsbereich eines Fahrzeugs mit einem geplanten Reinigungseintrag." caption="Wartungen planen und den Verlauf am Fahrzeug nachschlagen." />

## Dein erster Versuch

Der [Tester-Einstieg](try) führt vom Start über zwei importierte Beispielfahrzeuge bis
zum eigenen dritten Fahrzeug mit Bild. Du erhältst die Dateien und eine kurze Checkliste
auf derselben Seite.

Windows startet aus einem ZIP. Für den Heimserver gibt es Docker-Images für AMD64 und ARM64.
Das [Handbuch](guide/) hilft bei den nächsten Schritten, die
[Installationsanleitung](administration/deployment-configuration) beim dauerhaften Betrieb.
Die Abbildungen zeigen den dokumentierten Entwicklungsstand; den Umfang der veröffentlichten
Version findest du in den [Release-Notizen](https://github.com/ichwars/RailKeeper/releases/latest).

## Deine Erfahrung hilft weiter

Teile in den [Discussions](https://github.com/ichwars/RailKeeper/discussions/categories/general),
was funktioniert und wo du hängen geblieben bist. Die [Feedbackfragen](try#feedback)
helfen beim ersten Bericht.

RailKeeper hilft dir? Mit einem [Stern auf GitHub](https://github.com/ichwars/RailKeeper)
unterstützt du das Projekt.
