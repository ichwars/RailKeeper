import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, expect, it, vi } from "vitest";
import { api } from "../../shared/api";
import { bundledDecoderInventory, bundledDecoderProfiles } from "../../shared/decoderLibraryModel";
import { DecoderLibrarySettings } from "./DecoderLibrarySettings";
import { DecoderLibraryProvider } from "../../shared/DecoderLibraryProvider";
import { DecoderLibraryNotice } from "../vehicles/DecoderLibraryNotice";

const catalog = () => structuredClone({ schemaVersion: 1, profiles: bundledDecoderProfiles, inventory: bundledDecoderInventory });
beforeEach(() => {
  window.localStorage.clear();
  vi.spyOn(api, "decoderLibrary").mockResolvedValue(catalog());
  vi.spyOn(api, "decoderReferences").mockResolvedValue([]);
});

it("navigates manufacturers, families, models and reviewed indexed definitions", async () => {
  const user = userEvent.setup();
  render(<DecoderLibrarySettings roles={["Viewer"]} />);
  const navigation = await screen.findByRole("navigation", { name: "Hersteller und Decoderfamilien" });
  expect(screen.getByText("8 Hersteller · 40 Familien · 599 Modelle")).toBeInTheDocument();
  await user.click(within(navigation).getByRole("button", { name: /^ESU/ }));
  await user.click(within(navigation).getByRole("button", { name: /^LokPilot 5\d+$/ }));
  await user.click(screen.getByRole("button", { name: "LokPilot 5 micro DCC" }));
  await user.click(screen.getByRole("tab", { name: "CV-Definitionen" }));
  await user.type(screen.getByRole("searchbox"), "CV 291");
  const table = screen.getByRole("table");
  expect(within(table).getByText("AUX3: Ausgangseffekt")).toBeInTheDocument();
  expect(within(table).getByText("CV31=16 / CV32=0")).toBeInTheDocument();
  await user.click(screen.getByRole("tab", { name: "Quellen" }));
  expect(screen.getByRole("link", { name: bundledDecoderProfiles[1].sources[0].title })).toHaveAttribute("href", bundledDecoderProfiles[1].sources[0].url);
  expect(screen.queryByRole("button", { name: /deaktivieren/ })).not.toBeInTheDocument();
});

it("searches article numbers and distinguishes directory metadata from reviewed CV rules", async () => {
  const user = userEvent.setup();
  render(<DecoderLibrarySettings roles={[]} />);
  const search = await screen.findByRole("searchbox");
  await user.type(search, "10318-01");
  expect(screen.getByText("1 Modell")).toBeInTheDocument();
  await user.click(screen.getByRole("button", { name: "SILVER+ Next18" }));
  expect(screen.getByText("Stammdaten geprüft")).toBeInTheDocument();
  expect(screen.getByText(/noch keine eigenen Hersteller-CV-Regeln/)).toBeInTheDocument();
  const identification = screen.getByRole("tab", { name: "Erkennung" });
  await user.click(identification);
  expect(screen.getByText(/CV7 und CV128/)).toBeInTheDocument();
  await user.keyboard("{ArrowLeft}");
  expect(screen.getByRole("tab", { name: "CV-Definitionen" })).toHaveFocus();
  await user.type(screen.getByRole("searchbox"), "CV 51");
  expect(screen.getByText("Keine passenden Definitionen gefunden.")).toBeInTheDocument();
});

it("filters reviewed models and recovers from a search without results", async () => {
  const user = userEvent.setup();
  render(<DecoderLibrarySettings roles={[]} />);
  await screen.findByRole("table");
  await user.click(screen.getByRole("button", { name: "Prüfumfang" }));
  await user.click(screen.getByRole("option", { name: "CV-Regeln geprüft" }));
  expect(screen.getByText("19 Modelle")).toBeInTheDocument();
  expect(screen.getByRole("button", { name: "10882 · PluX16-Decoder (NEM 658)" })).toBeInTheDocument();
  await user.type(screen.getByRole("searchbox"), "nicht-vorhanden");
  expect(screen.getByText("Keine Modelle für diese Auswahl gefunden.")).toBeInTheDocument();
  await user.click(screen.getByRole("button", { name: "Auswahl zurücksetzen" }));
  expect(screen.getByText("599 Modelle")).toBeInTheDocument();
});

it("refreshes interpretation after deactivation without editing vehicle values", async () => {
  const user = userEvent.setup();
  const rows = [{ cvNumber: 8, value: 151, decoderProfile: "LokPilot 5 micro DCC" }];
  vi.spyOn(api, "setMasterDataActive").mockImplementation(async () => {
    const next = catalog(); next.profiles = next.profiles.filter(p => p.id !== "esu-lokpilot-5-micro-dcc");
    vi.mocked(api.decoderLibrary).mockResolvedValue(next);
    return { id: "decoder_profile:esu-lokpilot-5-micro-dcc", type: "decoder_profile",
      key: "esu-lokpilot-5-micro-dcc", label: "ESU", active: false, sortOrder: 1,
      metadata: {}, createdAt: "now", updatedAt: "now" };
  });
  render(<DecoderLibraryProvider><DecoderLibrarySettings roles={["Editor"]} />
    <DecoderLibraryNotice values={rows} profile="LokPilot 5 micro DCC" /></DecoderLibraryProvider>);
  await screen.findByText("Profil erkannt: ESU LokPilot 5 micro DCC");
  await user.click(await screen.findByRole("button", { name: "LokPilot 5 micro DCC" }));
  await user.click(screen.getByRole("button", { name: "Für Auswertung deaktivieren" }));
  await screen.findByText("Auswertung mit DCC-Grundlagen");
  await waitFor(() => expect(screen.queryByText("Profil erkannt: ESU LokPilot 5 micro DCC")).not.toBeInTheDocument());
  expect(api.setMasterDataActive).toHaveBeenCalledWith("decoder_profile", "esu-lokpilot-5-micro-dcc", false);
  expect(rows[0].value).toBe(151);
});

it("shows loading, empty and load-error states", async () => {
  vi.mocked(api.decoderLibrary).mockResolvedValueOnce({ schemaVersion: 1, profiles: [] });
  const { unmount } = render(<DecoderLibrarySettings roles={[]} />);
  expect(screen.getByText("Decoderbibliothek wird geladen …")).toBeInTheDocument();
  await screen.findByText("Keine Decoderprofile vorhanden.");
  unmount();
  vi.mocked(api.decoderLibrary).mockRejectedValueOnce(new Error("offline"));
  render(<DecoderLibrarySettings roles={[]} />);
  expect(await screen.findByRole("alert")).toHaveTextContent("konnte nicht geladen werden");
});

it("does not request viewer-only data for exhibition-only sessions", () => {
  render(<DecoderLibraryProvider enabled={false}><p>Ausstellung</p></DecoderLibraryProvider>);
  expect(api.decoderLibrary).not.toHaveBeenCalled();
});

it("shows the documented F060 connection and leaves its unconfirmed order number open", async () => {
  const user = userEvent.setup();
  render(<DecoderLibrarySettings roles={["Viewer"]} />);
  await user.type(await screen.findByRole("searchbox"), "F060");
  await user.click(screen.getByRole("button", { name: "F060" }));
  expect(screen.getByText("Litzen / wires")).toBeInTheDocument();
  expect(screen.getByText(/Die geprüfte Anleitung nennt keine Artikelnummer/)).toBeInTheDocument();
  expect(screen.getByText("Stammdaten geprüft")).toBeInTheDocument();
  await user.click(screen.getByRole("tab", { name: "Quellen" }));
  expect(screen.getByRole("link", { name: "Kühn F060 Anleitung" })).toHaveAttribute(
    "href", "https://www.kuehn-digital.de/files/F060_de_c.pdf",
  );
});

it("shows the eight manufacturers, exact order connections and limits for older and branded entries", async () => {
  const user = userEvent.setup();
  render(<DecoderLibrarySettings roles={["Viewer"]} />);
  const navigation = await screen.findByRole("navigation", { name: "Hersteller und Decoderfamilien" });
  for (const maker of ["Kühn", "Uhlenbrock", "Roco", "PIKO"]) {
    expect(within(navigation).getByRole("button", { name: new RegExp(`^${maker}`) })).toBeInTheDocument();
  }
  expect(screen.getByText(/kein vollständiges Herstellerarchiv/)).toBeInTheDocument();
  await user.type(screen.getByRole("searchbox"), "59857");
  await user.click(screen.getByRole("button", { name: "6 Bestellvarianten ansehen" }));
  const variants = screen.getByRole("region", { name: "Dokumentierte Bestellvarianten" });
  const straight = within(variants).getByRole("row", { name: /59827/ });
  const angled = within(variants).getByRole("row", { name: /59857/ });
  expect(straight).toHaveTextContent("NEM 651 (direkt)");
  expect(angled).toHaveTextContent("NEM 651 (direkt, gewinkelt)");
  expect(within(angled).getByRole("link")).toHaveAttribute("href", "https://www.esu.eu/produkte/lokpilot/lokpilot-5-micro-dcc/");
  await user.click(screen.getByRole("button", { name: "Zur Modellübersicht" }));
  await user.clear(screen.getByRole("searchbox"));
  await user.type(screen.getByRole("searchbox"), "82350");
  await user.click(screen.getByRole("button", { name: "N45" }));
  expect(screen.getByText(/^Historischer Dokumentationsstand \/ ältere Generation/)).toBeInTheDocument();
  expect(screen.getByText("Registrierte CV8-ID: 157 · Auswertungsbereich: DCC")).toBeInTheDocument();
  await user.click(screen.getByRole("button", { name: "Zur Modellübersicht" }));
  await user.clear(screen.getByRole("searchbox"));
  await user.type(screen.getByRole("searchbox"), "10891");
  expect(screen.getByRole("button", { name: "1 Bestellvariante ansehen" })).toBeInTheDocument();
  await user.click(screen.getByRole("button", { name: /10891 ·/ }));
  expect(screen.getByText(/Die tatsächliche CV8 bleibt maßgeblich/)).toBeInTheDocument();
  expect(screen.getByText(/noch keine eigenen Hersteller-CV-Regeln/)).toBeInTheDocument();
});


it("opens Roco's documentation profile without inventing an identity link or unsupported foundation CVs", async () => {
  const user = userEvent.setup();
  render(<DecoderLibrarySettings roles={["Viewer"]} />);
  await user.type(await screen.findByRole("searchbox"), "10882");
  await user.click(screen.getByRole("button", { name: "10882 · PluX16-Decoder (NEM 658)" }));
  expect(screen.getByText("CV-Regeln geprüft")).toBeInTheDocument();
  expect(screen.getByText("Technischer Decoderhersteller laut Anleitung: ZIMO · CV8=145")).toBeVisible();
  await user.click(screen.getByRole("tab", { name: "CV-Definitionen" }));
  await user.type(screen.getByRole("searchbox"), "CV 124");
  expect(within(screen.getByRole("table")).getByText("Bremsstreckenerkennung")).toBeVisible();
  await user.clear(screen.getByRole("searchbox")); await user.type(screen.getByRole("searchbox"), "CV 19");
  expect(screen.getByText("Keine passenden Definitionen gefunden.")).toBeVisible();
  await user.click(screen.getByRole("tab", { name: "Erkennung" }));
  expect(screen.queryByText("Automatische Decodererkennung")).not.toBeInTheDocument();
  expect(screen.queryByRole("button", { name: /deaktivieren/ })).not.toBeInTheDocument();
  const model = bundledDecoderInventory.models.find(m => m.documentationProfileId === "roco-10882")!;
  expect(model.brandOnly).toBe(true); expect(model.profileId).toBeUndefined(); expect(model.identification).toBeUndefined();
});
