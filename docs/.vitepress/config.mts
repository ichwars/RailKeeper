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
              ],
            },
          ],
          "/de/administration/": [
            {
              text: "Administration",
              items: [{ text: "Überblick", link: "/de/administration/" }],
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
          ],
        },
      ],
      "/administration/": [
        {
          text: "Administration",
          items: [{ text: "Overview", link: "/administration/" }],
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
