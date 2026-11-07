import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import type { VersionInfo } from "../../shared/api";
import { languageSettingKey } from "../../shared/i18n";
import { SettingsBuildIdentity } from "./SettingsBuildIdentity";

type Build = NonNullable<VersionInfo["build"]>;
const labels = {
  de: { release: "Release", development: "Entwicklungsstand", modified: "Lokale Änderungen", unknown: "Build-Herkunft unbekannt" },
  en: { release: "Release", development: "Development build", modified: "Local changes", unknown: "Unknown build origin" }
};

describe("Settings candidate identity", () => {
  for (const language of ["de", "en"] as const) {
    for (const status of ["release", "development", "modified", "unknown"] as const) {
      it(`distinguishes ${status} in ${language} while retaining the full commit for inspection`, () => {
        localStorage.setItem(languageSettingKey, language);
        const build: Build = { commit: "a".repeat(40), shortCommit: "a".repeat(12), status, goVersion: "go-test", platform: "linux/test" };
        render(<SettingsBuildIdentity build={build} />);
        expect(screen.getByText(labels[language][status], { exact: false })).toBeInTheDocument();
        expect(screen.getByText(build.shortCommit!)).toBeInTheDocument();
        expect(screen.getByTitle(build.commit!)).toBeInTheDocument();
      });
    }
  }

  it("labels an unknown source explicitly", () => {
    render(<SettingsBuildIdentity build={{ status: "unknown", goVersion: "go-test", platform: "test" }} />);
    expect(screen.getByText("Commit unbekannt")).toBeInTheDocument();
    expect(screen.getByText("Build-Herkunft unbekannt", { exact: false })).toBeInTheDocument();
  });

  it("accepts version responses from an older server without build fields", () => {
    const { container } = render(<SettingsBuildIdentity />);
    expect(container).toBeEmptyDOMElement();
  });
});
