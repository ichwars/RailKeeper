import { fileURLToPath, URL } from "node:url";

import { defineConfig } from "vitepress";

const repositoryUrl = "https://github.com/ichwars/RailKeeper";

export default defineConfig({
  title: "RailKeeper",
  description: "Documentation for the local-first model railway inventory and operations tool.",
  base: "/RailKeeper/",
  head: [
    [
      "link",
      {
        rel: "icon",
        type: "image/png",
        href: "/RailKeeper/brand/railkeeper-mark.png",
      },
    ],
  ],
  srcDir: "site",
  cleanUrls: true,
  lastUpdated: true,
  vite: {
    publicDir: fileURLToPath(new URL("../../frontend/public", import.meta.url)),
  },
  locales: {
    root: { label: "English", lang: "en" },
    de: {
      label: "Deutsch",
      lang: "de",
      link: "/de/",
      description: "Dokumentation für RailKeeper.",
      themeConfig: {
        nav: [
          { text: "Benutzerhandbuch", link: "/de/guide/" },
          { text: "Administration", link: "/de/administration/" },
          { text: "Hilfe und Begriffe", link: "/de/reference/navigation-and-glossary" },
          { text: "Technik", items: [{ text: "Entwicklung", link: "/de/development/" }, { text: "Referenz", link: "/de/reference/" }] },
        ],
        sidebar: {
          "/de/guide/": [
            {
              text: "Start",
              items: [
                { text: "Was möchtest du erledigen?", link: "/de/guide/" },
                { text: "Dein erstes Fahrzeug", link: "/de/guide/quick-start" },
                { text: "Erste Schritte", link: "/de/guide/getting-started/" },
                { text: "Dashboard und Datenqualität", link: "/de/guide/overview/" },
              ],
            },
            {
              text: "Fahrzeuge",
              items: [
                { text: "Fahrzeugbestand", link: "/de/guide/vehicles/" },
                { text: "Fahrzeugsets", link: "/de/guide/vehicles/sets" },
                { text: "Fahrzeugbilder und Beilagen", link: "/de/guide/vehicles/media" },
                { text: "Fahrzeugwartung und Zustand", link: "/de/guide/vehicles/maintenance" },
                { text: "Artikelsuche, Web-Dokumente und Ersatzteile", link: "/de/guide/vehicles/search-and-spares" },
              ],
            },
            {
              text: "Decoder und Zentralen",
              items: [
                { text: "Decoder, Funktionen und CV-Daten", link: "/de/guide/vehicles/decoder-cv" },
                { text: "Decoder nachschlagen", link: "/de/guide/vehicles/decoder-library" },
                { text: "Decoder sicher erkennen", link: "/de/guide/vehicles/decoder-identification" },
                { text: "F0, F1 und F3 übernehmen", link: "/de/guide/vehicles/functions-walkthrough" },
                { text: "CV-Referenz", link: "/de/guide/vehicles/decoder-cv-reference" },
                { text: "Modell- und Firmwarehinweise", link: "/de/guide/vehicles/decoder-model-notes" },
                { text: "CV-Import ausprobieren", link: "/de/guide/vehicles/cv-import-example" },
                { text: "Lokbilder für Zentralen", link: "/de/guide/vehicles/locomotive-images" },
                { text: "CS3-Lokdaten read-only lesen", link: "/de/guide/import-export/cs3-import" },
                { text: "ECoS-Lokabgleich", link: "/de/guide/import-export/ecos-sync" },
              ],
            },
            {
              text: "Zubehör",
              items: [
                { text: "Zubehörübersicht", link: "/de/guide/accessories/" },
                { text: "Artikelstammdaten und Fachangaben", link: "/de/guide/accessories/article-records" },
                { text: "Bestand, Käufe und Dokumente", link: "/de/guide/accessories/stock-purchases-documents" },
                { text: "Reservierungen, Einbauten und Verwendung", link: "/de/guide/accessories/allocations-history" },
              ],
            },
            {
              text: "Messe",
              items: [
                { text: "Messearbeitsbereich", link: "/de/guide/exhibition/" },
                { text: "Listen und Sperren", link: "/de/guide/exhibition/lists-and-locking" },
                { text: "Einträge und Drucken", link: "/de/guide/exhibition/entries-and-printing" },
              ],
            },
            {
              text: "Import und Export",
              items: [
                { text: "Überblick und Berechtigungen", link: "/de/guide/import-export/" },
                { text: "Fahrzeugdateien importieren", link: "/de/guide/import-export/file-import" },
                { text: "Importbeispiele", link: "/de/guide/import-export/examples" },
                { text: "Bestand exportieren", link: "/de/guide/import-export/exports" },
              ],
            },
            {
              text: "Einstellungen",
              items: [
                { text: "Überblick und Berechtigungen", link: "/de/guide/settings/" },
                { text: "Persönliche Einstellungen", link: "/de/guide/settings/personal-preferences" },
                { text: "Darstellung", link: "/de/guide/settings/appearance" },
              ],
            },
          ],
          "/de/administration/": [
            {
              text: "Administration",
              items: [
                { text: "Überblick", link: "/de/administration/" },
                { text: "Installieren", link: "/de/administration/deployment-configuration" },
                { text: "Konten und Rechte", link: "/de/administration/users-sessions-security" },
                { text: "Sichern und wiederherstellen", link: "/de/administration/backup-restore" },
                { text: "Zentrale einrichten", link: "/de/administration/digital-centers" },
                { text: "Updates und Betrieb", link: "/de/administration/system-operations" },
                { text: "Stammdaten-Administration", link: "/de/administration/master-data" },
                { text: "Allgemeine Stammdaten", link: "/de/administration/master-data-general" },
                { text: "Artikelstammdaten und Lagerorte", link: "/de/administration/master-data-articles" },
                { text: "Inventarnummernschemata", link: "/de/administration/master-data-inventory-numbers" },
                { text: "Stammdatentransfer", link: "/de/administration/master-data-transfer" },
              ],
            },
          ],
          "/de/development/": [
            {
              text: "Entwicklung",
              items: [{ text: "Überblick", link: "/de/development/" }, { text: "Architektur und Beiträge", link: "/de/development/architecture" }],
            },
          ],
          "/de/reference/": [
            {
              text: "Referenz",
              items: [
                { text: "Überblick", link: "/de/reference/" },
                { text: "Dokumentationsabdeckung", link: "/de/reference/coverage" },
                { text: "Begriffe und Navigation", link: "/de/reference/navigation-and-glossary" },
                { text: "Versionen und Hilfe", link: "/de/reference/releases-and-support" },
              ],
            },
          ],
        },
        editLink: {
          pattern: `${repositoryUrl}/edit/main/docs/site/:path`,
          text: "Diese Seite auf GitHub bearbeiten",
        },
        darkModeSwitchLabel: "Darstellung",
        lightModeSwitchTitle: "Zum hellen Design wechseln",
        darkModeSwitchTitle: "Zum dunklen Design wechseln",
        sidebarMenuLabel: "Menü",
        returnToTopLabel: "Nach oben",
        langMenuLabel: "Sprache wechseln",
        skipToContentLabel: "Zum Inhalt springen",
        docFooter: {
          prev: "Vorherige Seite",
          next: "Nächste Seite",
        },
        outline: { label: "Auf dieser Seite" },
        lastUpdated: {
          text: "Letzte Änderung",
          formatOptions: { dateStyle: "medium", timeStyle: "short", forceLocale: true },
        },
      },
    },
  },
  themeConfig: {
    logo: "/brand/railkeeper-mark.png",
    siteTitle: "RailKeeper Docs",
    socialLinks: [{ icon: "github", link: repositoryUrl }],
    editLink: {
      pattern: `${repositoryUrl}/edit/main/docs/site/:path`,
      text: "Edit this page on GitHub",
    },
    nav: [
      { text: "User Guide", link: "/guide/" },
      { text: "Administration", link: "/administration/" },
      { text: "Help and terms", link: "/reference/navigation-and-glossary" },
      { text: "Technical details", items: [{ text: "Development", link: "/development/" }, { text: "Reference", link: "/reference/" }] },
    ],
    sidebar: {
      "/guide/": [
        {
          text: "Start here",
          items: [
            { text: "Choose a task", link: "/guide/" },
            { text: "Your first vehicle", link: "/guide/quick-start" },
            { text: "Getting started", link: "/guide/getting-started/" },
            { text: "Dashboard and data quality", link: "/guide/overview/" },
          ],
        },
        {
          text: "Vehicles",
          items: [
            { text: "Vehicle inventory", link: "/guide/vehicles/" },
            { text: "Vehicle sets", link: "/guide/vehicles/sets" },
            { text: "Vehicle images and attachments", link: "/guide/vehicles/media" },
            { text: "Vehicle maintenance and condition", link: "/guide/vehicles/maintenance" },
            { text: "Article search, web documents, and spare parts", link: "/guide/vehicles/search-and-spares" },
          ],
        },
        {
          text: "Decoders and command stations",
          items: [
            { text: "Decoder, functions, and CV data", link: "/guide/vehicles/decoder-cv" },
            { text: "Look up a decoder", link: "/guide/vehicles/decoder-library" },
                { text: "Identify a decoder", link: "/guide/vehicles/decoder-identification" },
                { text: "Apply F0, F1 and F3", link: "/guide/vehicles/functions-walkthrough" },
                { text: "CV reference", link: "/guide/vehicles/decoder-cv-reference" },
                { text: "Model and firmware notes", link: "/guide/vehicles/decoder-model-notes" },
            { text: "Try CV import", link: "/guide/vehicles/cv-import-example" },
            { text: "Command-station images", link: "/guide/vehicles/locomotive-images" },
            { text: "Read CS3 locomotives without writing", link: "/guide/import-export/cs3-import" },
            { text: "ECoS locomotive sync", link: "/guide/import-export/ecos-sync" },
          ],
        },
        {
          text: "Accessories",
          items: [
            { text: "Accessories overview", link: "/guide/accessories/" },
            { text: "Article records and technical data", link: "/guide/accessories/article-records" },
            { text: "Stock, purchases, and documents", link: "/guide/accessories/stock-purchases-documents" },
            { text: "Reservations, installations, and usage", link: "/guide/accessories/allocations-history" },
          ],
        },
        {
          text: "Exhibition",
          items: [
            { text: "Exhibition workspace", link: "/guide/exhibition/" },
            { text: "Lists and locking", link: "/guide/exhibition/lists-and-locking" },
            { text: "Entries and printing", link: "/guide/exhibition/entries-and-printing" },
          ],
        },
        {
          text: "Import and export",
          items: [
            { text: "Overview and permissions", link: "/guide/import-export/" },
            { text: "Import vehicle files", link: "/guide/import-export/file-import" },
            { text: "Import examples", link: "/guide/import-export/examples" },
            { text: "Export inventory", link: "/guide/import-export/exports" },
          ],
        },
        {
          text: "Settings",
          items: [
            { text: "Overview and permissions", link: "/guide/settings/" },
            { text: "Personal preferences", link: "/guide/settings/personal-preferences" },
            { text: "Appearance", link: "/guide/settings/appearance" },
          ],
        },
      ],
      "/administration/": [
        {
          text: "Administration",
          items: [
            { text: "Overview", link: "/administration/" },
            { text: "Install", link: "/administration/deployment-configuration" },
                { text: "Accounts and permissions", link: "/administration/users-sessions-security" },
                { text: "Backup and restore", link: "/administration/backup-restore" },
                { text: "Connect a command station", link: "/administration/digital-centers" },
                { text: "Updates and operation", link: "/administration/system-operations" },
                { text: "Master-data administration", link: "/administration/master-data" },
            { text: "General master data", link: "/administration/master-data-general" },
            { text: "Article master data and locations", link: "/administration/master-data-articles" },
            { text: "Inventory-number schemes", link: "/administration/master-data-inventory-numbers" },
            { text: "Master-data transfer", link: "/administration/master-data-transfer" },
          ],
        },
      ],
      "/development/": [
        {
          text: "Development",
          items: [{ text: "Overview", link: "/development/" }, { text: "Architecture and contributions", link: "/development/architecture" }],
        },
      ],
      "/reference/": [
        {
          text: "Reference",
          items: [
            { text: "Overview", link: "/reference/" },
            { text: "Documentation Coverage", link: "/reference/coverage" },
            { text: "Terms and navigation", link: "/reference/navigation-and-glossary" },
            { text: "Versions and support", link: "/reference/releases-and-support" },
          ],
        },
      ],
    },
    outline: { label: "On this page" },
    lastUpdated: {
      text: "Last updated",
      formatOptions: { dateStyle: "medium", timeStyle: "short", forceLocale: true },
    },
    search: {
      provider: "local",
      options: {
        locales: {
          de: {
            translations: {
              button: {
                buttonText: "Suchen",
                buttonAriaLabel: "Dokumentation durchsuchen",
              },
              modal: {
                displayDetails: "Detaillierte Liste anzeigen",
                resetButtonTitle: "Suche zurücksetzen",
                backButtonTitle: "Suche schließen",
                noResultsText: "Keine Ergebnisse gefunden für",
                footer: {
                  selectText: "Auswählen",
                  selectKeyAriaLabel: "Eingabetaste",
                  navigateText: "Navigieren",
                  navigateUpKeyAriaLabel: "Pfeil nach oben",
                  navigateDownKeyAriaLabel: "Pfeil nach unten",
                  closeText: "Schließen",
                  closeKeyAriaLabel: "Escape-Taste",
                },
              },
            },
          },
        },
      },
    },
  },
});
