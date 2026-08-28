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
          { text: "Entwicklung", link: "/de/development/" },
          { text: "Referenz", link: "/de/reference/" },
        ],
        sidebar: {
          "/de/guide/": [
            {
              text: "Benutzerhandbuch",
              items: [
                { text: "Überblick", link: "/de/guide/" },
                { text: "Erste Schritte", link: "/de/guide/getting-started/" },
                { text: "Dashboard und Datenqualität", link: "/de/guide/overview/" },
                { text: "Fahrzeugbestand", link: "/de/guide/vehicles/" },
                { text: "Fahrzeugbilder und Beilagen", link: "/de/guide/vehicles/media" },
                { text: "Fahrzeugwartung und Zustand", link: "/de/guide/vehicles/maintenance" },
                { text: "Decoder, Funktionen und CV-Daten", link: "/de/guide/vehicles/decoder-cv" },
                {
                  text: "Artikelsuche, Web-Dokumente und Ersatzteile",
                  link: "/de/guide/vehicles/search-and-spares",
                },
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
                { text: "Bestand exportieren", link: "/de/guide/import-export/exports" },
                { text: "ECoS-Lokabgleich", link: "/de/guide/import-export/ecos-sync" },
                { text: "CS3-Lokdaten read-only lesen", link: "/de/guide/import-export/cs3-import" },
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
              items: [{ text: "Überblick", link: "/de/development/" }],
            },
          ],
          "/de/reference/": [
            {
              text: "Referenz",
              items: [
                { text: "Überblick", link: "/de/reference/" },
                { text: "Dokumentationsabdeckung", link: "/de/reference/coverage" },
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
        outlineTitle: "Auf dieser Seite",
        lastUpdated: { text: "Zuletzt geprüft" },
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
      { text: "Development", link: "/development/" },
      { text: "Reference", link: "/reference/" },
    ],
    sidebar: {
      "/guide/": [
        {
          text: "User Guide",
          items: [
            { text: "Overview", link: "/guide/" },
            { text: "Getting started", link: "/guide/getting-started/" },
            { text: "Dashboard and data quality", link: "/guide/overview/" },
            { text: "Vehicle inventory", link: "/guide/vehicles/" },
            { text: "Vehicle images and attachments", link: "/guide/vehicles/media" },
            { text: "Vehicle maintenance and condition", link: "/guide/vehicles/maintenance" },
            { text: "Decoder, functions, and CV data", link: "/guide/vehicles/decoder-cv" },
            {
              text: "Article search, web documents, and spare parts",
              link: "/guide/vehicles/search-and-spares",
            },
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
            { text: "Export inventory", link: "/guide/import-export/exports" },
            { text: "ECoS locomotive sync", link: "/guide/import-export/ecos-sync" },
            { text: "Read CS3 locomotives without writing", link: "/guide/import-export/cs3-import" },
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
          items: [{ text: "Overview", link: "/development/" }],
        },
      ],
      "/reference/": [
        {
          text: "Reference",
          items: [
            { text: "Overview", link: "/reference/" },
            { text: "Documentation Coverage", link: "/reference/coverage" },
          ],
        },
      ],
    },
    outlineTitle: "On this page",
    lastUpdated: { text: "Last reviewed" },
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
