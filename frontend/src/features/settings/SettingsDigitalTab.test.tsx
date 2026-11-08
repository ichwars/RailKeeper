import { act, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api, type DigitalCenterSettings, type ECoSLiveStatus } from "../../shared/api";
import { SettingsDigitalTab } from "./SettingsDigitalTab";

const settings = (ecosEnabled = false): DigitalCenterSettings => ({
  provider: "ecos",
  ecos: { enabled: ecosEnabled, host: "192.168.2.151", port: "15471" },
  z21: { enabled: false, host: "", port: "21105" },
  intellibox3: { enabled: false, host: "", port: "21105" },
  cs3: { enabled: false, host: "", port: "80" }
});

const liveStatus: ECoSLiveStatus = {
  provider: "ecos",
  connected: false,
  state: "stopped",
  blocksReceived: 0,
  repliesReceived: 0,
  eventsReceived: 0,
  pulseSamples: [],
  recentEvents: [],
  diagnosis: { connectionState: "stopped", passive: true },
  message: "ECoS-Live-Verbindung inaktiv."
};

describe("SettingsDigitalTab commissioning workflow", () => {
  beforeEach(() => {
    vi.spyOn(api, "digitalSettings").mockResolvedValue(settings());
    vi.spyOn(api, "updateDigitalSettings").mockImplementation(async (input) => input);
    vi.spyOn(api, "getECoSLiveStatus").mockResolvedValue(liveStatus);
    vi.spyOn(api, "testECoSConnection").mockResolvedValue({
      connected: true,
      host: "192.168.2.151",
      port: 15471,
      applicationVersion: "4.3.1",
      compatibilityStatus: "unverified",
      missingFields: [],
      message: "ECoS-Verbindung erfolgreich getestet."
    });
  });

  it("requires a successful test before activating the adapter", async () => {
    const user = userEvent.setup();
    render(
      <SettingsDigitalTab
        canManageUsers
        formatDateTime={(value) => value}
        username="admin"
      />
    );

    expect(await screen.findByText("192.168.2.151:15471")).toBeInTheDocument();
    const activate = screen.getByRole("button", { name: "Adapter aktivieren" });
    expect(activate).toBeDisabled();

    await user.click(screen.getByRole("button", { name: "Verbindung testen" }));
    expect(await screen.findByText("Verbindung erfolgreich geprüft")).toBeInTheDocument();
    expect(screen.getByText("Vollständig, Firmware noch nicht geräteverifiziert")).toBeInTheDocument();
    expect(activate).toBeEnabled();

    await user.click(activate);
    await waitFor(() => {
      expect(api.updateDigitalSettings).toHaveBeenLastCalledWith(
        expect.objectContaining({ ecos: expect.objectContaining({ enabled: true }) })
      );
    });
    expect(await screen.findByText("Adapter und Importbereich wurden aktiviert.")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Live-Monitor starten" })).toBeEnabled();
  });

  it("makes the diagnostic and ECoS message tabs interactive", async () => {
    const user = userEvent.setup();
    render(
      <SettingsDigitalTab
        canManageUsers
        formatDateTime={(value) => value}
        username="admin"
      />
    );

    await screen.findByText("192.168.2.151:15471");
    await user.click(screen.getByRole("tab", { name: "Letzte Diagnose" }));
    expect(screen.getByRole("tabpanel")).toHaveTextContent("Dieser Adapter bietet derzeit nur den Verbindungstest.");

    await user.click(screen.getByRole("tab", { name: "ECoS-Meldungen" }));
    expect(screen.getByRole("tabpanel")).toHaveTextContent("ECoS-Live-Verbindung inaktiv.");
    expect(screen.getByRole("button", { name: "Status aktualisieren" })).toBeEnabled();
  });

  it("removes a configured connection only after confirmation", async () => {
    const user = userEvent.setup();
    vi.mocked(api.digitalSettings).mockResolvedValue(settings(true));
    const confirm = vi.spyOn(window, "confirm").mockReturnValue(true);

    render(
      <SettingsDigitalTab
        canManageUsers
        formatDateTime={(value) => value}
        username="admin"
      />
    );

    await screen.findByText("192.168.2.151:15471");
    await user.click(screen.getByRole("button", { name: "Verbindung entfernen" }));

    expect(confirm).toHaveBeenCalledWith(expect.stringContaining("ESU ECoS"));
    await waitFor(() => {
      expect(api.updateDigitalSettings).toHaveBeenLastCalledWith(
        expect.objectContaining({ ecos: { enabled: false, host: "", port: "15471" } })
      );
    });
    expect(await screen.findByText("Adapter-Verbindung wurde entfernt.")).toBeInTheDocument();
  });

  it("runs CS3 read-only diagnostics and exposes the import safety boundary", async () => {
    const user = userEvent.setup();
    vi.mocked(api.digitalSettings).mockResolvedValue({
      ...settings(),
      provider: "cs3",
      cs3: { enabled: true, host: "192.168.2.46", port: "80" }
    });
    vi.spyOn(api, "probeCS3Connection").mockResolvedValue({
      provider: "cs3",
      connected: true,
      host: "192.168.2.46",
      port: 80,
      message: "CS3-Diagnose abgeschlossen.",
      fields: { readOnly: "true", locomotiveCount: "1" },
      commands: [{
        name: "CS3_LOCOMOTIVE_API",
        description: "Read-only Loklisten-API",
        request: "GET /app/api/locos",
        commandHex: "",
        ok: true,
        fields: { readOnly: "true", locomotiveCount: "1" }
      }]
    });

    render(
      <SettingsDigitalTab
        canManageUsers
        formatDateTime={(value) => value}
        username="admin"
      />
    );

    expect(await screen.findByText("192.168.2.46:80")).toBeInTheDocument();
    expect(screen.getByText("Liest Name, Adresse und Protokoll nur lesend in den Vergleichsbereich ein.")).toBeInTheDocument();
    await user.click(screen.getByRole("tab", { name: "Letzte Diagnose" }));
    await user.click(screen.getByRole("button", { name: "Diagnosedaten lesen" }));

    await waitFor(() => {
      expect(api.probeCS3Connection).toHaveBeenCalledWith({ host: "192.168.2.46", port: 80 });
    });
    expect(await screen.findByText("GET /app/api/locos")).toBeInTheDocument();
    expect(screen.getAllByText("true").length).toBeGreaterThan(0);
  });

  it("shows the bounded address-only read scope for Z21", async () => {
    vi.mocked(api.digitalSettings).mockResolvedValue({
      ...settings(),
      provider: "z21",
      z21: { enabled: true, host: "192.168.2.45", port: "21105" }
    });

    render(
      <SettingsDigitalTab
        canManageUsers
        formatDateTime={(value) => value}
        username="admin"
      />
    );

    expect(await screen.findByText("192.168.2.45:21105")).toBeInTheDocument();
    expect(screen.getByText("Fragt nur bekannte RailKeeper-Decoderadressen ab. Loknamen, Fahrzustände und Funktionen werden nicht übernommen.")).toBeInTheDocument();
    expect(screen.getByText("Lokdaten lesen").closest("article")?.querySelector(".capability-state.available")).toBeInTheDocument();
    expect(screen.getByText("Schreibzugriffe geschützt").closest("article")).toHaveTextContent("Dieser Adapter führt keine Schreibzugriffe aus.");
  });

  it("separates available Z21 UDP from disabled LocoNet TCP for Intellibox 3", async () => {
    vi.mocked(api.digitalSettings).mockResolvedValue({
      ...settings(),
      provider: "intellibox3",
      intellibox3: { enabled: true, host: "192.168.2.47", port: "21105" }
    });

    render(
      <SettingsDigitalTab
        canManageUsers
        formatDateTime={(value) => value}
        username="admin"
      />
    );

    expect(await screen.findByText("192.168.2.47:21105")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Transportwege" })).toBeInTheDocument();
    expect(screen.getByText("Z21 über UDP")).toBeInTheDocument();
    expect(screen.getByText("Verfügbar")).toBeInTheDocument();
    expect(screen.getByText("LocoNet-over-TCP")).toBeInTheDocument();
    expect(screen.getByText("Geplant")).toBeInTheDocument();
    expect(screen.getByText(/es werden keine TCP-Anfragen gesendet/)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Live-Monitor starten" })).toBeDisabled();
  });
});

describe("digital adapter failure and isolation contracts", () => {
  const cases = [
    { provider: "ecos", label: "ESU ECoS", hostLabel: "ECoS-IP oder Hostname", portLabel: "Port", port: "15471" },
    { provider: "z21", label: "Roco/Fleischmann Z21", hostLabel: "Z21-IP oder Hostname", portLabel: "UDP-Port", port: "21105" },
    { provider: "intellibox3", label: "Uhlenbrock Intellibox 3", hostLabel: "Intellibox-3-IP oder Hostname", portLabel: "Z21-UDP-Port", port: "21105" },
    { provider: "cs3", label: "Märklin CS3", hostLabel: "CS3-IP oder Hostname", portLabel: "HTTP-Port", port: "80" }
  ] as const;
  beforeEach(() => {
    const configured = settings();
    configured.z21.host = "z21.test"; configured.intellibox3.host = "ibox.test"; configured.cs3.host = "cs3.test";
    vi.spyOn(api,"digitalSettings").mockResolvedValue(configured);
    vi.spyOn(api,"updateDigitalSettings").mockImplementation(async input=>input);
    vi.spyOn(api,"getECoSLiveStatus").mockResolvedValue(liveStatus);
  });

  it.each(cases)("keeps each provider's host/port separate when editing $provider",async item=>{
    const user=userEvent.setup();
    render(<SettingsDigitalTab canManageUsers username="synthetic-admin" formatDateTime={value=>value}/>);
    await screen.findByText("192.168.2.151:15471");
    const selector=screen.getByRole("region",{name:"Adapter auswählen"});
    await user.click(within(selector).getByRole("button",{name:new RegExp(item.label.replace(/[.*+?^${}()|[\]\\]/g,"\\$&"))}));
    await user.click(screen.getAllByRole("button",{name:"Konfiguration ändern"})[0]);
    const dialog=screen.getByRole("dialog");
    await user.clear(within(dialog).getByLabelText(item.hostLabel));
    await user.type(within(dialog).getByLabelText(item.hostLabel),`${item.provider}.updated.test`);
    await user.clear(within(dialog).getByLabelText(item.portLabel));
    await user.type(within(dialog).getByLabelText(item.portLabel),"3456");
    await user.click(within(dialog).getByRole("button",{name:"Speichern"}));
    await waitFor(()=>expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
    expect(api.updateDigitalSettings).toHaveBeenLastCalledWith(expect.objectContaining({provider:item.provider,[item.provider]:{host:`${item.provider}.updated.test`,port:"3456",enabled:false}}));
    const stored=JSON.parse(window.localStorage.getItem("railkeeper.digital.configurationMeta")||"{}");
    expect(stored[item.provider].configuredBy).toBe("synthetic-admin");
  });

  it.each(cases)("keeps activation disabled after a failed $provider test and allows a fresh test",async item=>{
    const methods={ecos:"testECoSConnection",z21:"testZ21Connection",intellibox3:"testIntellibox3Connection",cs3:"testCS3Connection"} as const;
    const test=vi.spyOn(api,methods[item.provider]).mockRejectedValueOnce(new Error("Synthetic connection failure"));
    if(item.provider==="ecos") vi.mocked(api.testECoSConnection).mockResolvedValue({connected:true,host:"station.test",port:Number(item.port),message:"Confirmed connection",compatibilityStatus:"unverified",missingFields:[]});
    else test.mockResolvedValue({provider:item.provider,connected:true,host:"station.test",port:Number(item.port),message:"Confirmed connection"});
    const user=userEvent.setup();
    render(<SettingsDigitalTab canManageUsers username="admin" formatDateTime={value=>value}/>);
    await screen.findByText("192.168.2.151:15471");
    await user.click(within(screen.getByRole("region",{name:"Adapter auswählen"})).getByRole("button",{name:new RegExp(item.label.replace(/[.*+?^${}()|[\]\\]/g,"\\$&"))}));
    await user.click(screen.getByRole("button",{name:"Verbindung testen"}));
    await screen.findAllByText("Synthetic connection failure");
    expect(screen.getByRole("button",{name:"Adapter aktivieren"})).toBeDisabled();
    await user.click(screen.getByRole("button",{name:"Verbindung testen"}));
    await waitFor(()=>expect(screen.getByRole("button",{name:"Adapter aktivieren"})).toBeEnabled());
    expect(test).toHaveBeenCalledTimes(2);
  });

  it("retains a configured connection after cancellation or a failed removal",async()=>{
    vi.mocked(api.digitalSettings).mockResolvedValue(settings(true));
    const confirm=vi.spyOn(window,"confirm").mockReturnValueOnce(false).mockReturnValue(true);
    const user=userEvent.setup();
    render(<SettingsDigitalTab canManageUsers username="admin" formatDateTime={value=>value}/>);
    await screen.findByText("192.168.2.151:15471");
    const remove=screen.getByRole("button",{name:"Verbindung entfernen"});
    await user.click(remove);
    expect(api.updateDigitalSettings).not.toHaveBeenCalled();
    vi.mocked(api.updateDigitalSettings).mockRejectedValueOnce(new Error("Removal failed"));
    await user.click(remove);
    await screen.findByText("Removal failed");
    expect(confirm).toHaveBeenCalledTimes(2);
    expect(screen.getByText("192.168.2.151:15471")).toBeInTheDocument();
    expect(remove).toBeEnabled();
  });

  it("starts, refreshes and stops the read-only monitor with explicit recoverable errors",async()=>{
    vi.mocked(api.digitalSettings).mockResolvedValue(settings(true));
    vi.mocked(api.getECoSLiveStatus).mockResolvedValue({...liveStatus,connected:true,state:"running"});
    vi.spyOn(api,"stopECoSLive").mockRejectedValueOnce(new Error("Stop failed")).mockResolvedValue({...liveStatus,message:"Stopped safely"});
    vi.spyOn(api,"startECoSLive").mockRejectedValueOnce(new Error("Start failed")).mockResolvedValue({...liveStatus,connected:true,state:"running",message:"Monitor connected"});
    vi.spyOn(api,"countECoSLocomotives").mockResolvedValue({host:"station.test",port:15471,count:2,message:"Two synthetic locomotives"});
    vi.spyOn(api,"testECoSConnection").mockResolvedValue({connected:true,host:"station.test",port:15471,message:"Connection rechecked",compatibilityStatus:"unverified",missingFields:[]});
    const user=userEvent.setup();
    render(<SettingsDigitalTab canManageUsers username="admin" formatDateTime={value=>value}/>);
    await screen.findByText("192.168.2.151:15471");
    await user.click(await screen.findByRole("button",{name:"Monitor stoppen"}));
    await screen.findByText("Stop failed");
    await user.click(screen.getByRole("button",{name:"Monitor stoppen"}));
    await screen.findByText("Stopped safely");
    expect(screen.getByRole("button",{name:"Live-Monitor starten"})).toBeDisabled();
    await user.click(screen.getByRole("button",{name:"Verbindung testen"}));
    await waitFor(()=>expect(screen.getByRole("button",{name:"Live-Monitor starten"})).toBeEnabled());
    await user.click(screen.getByRole("button",{name:"Live-Monitor starten"}));
    await screen.findByText("Start failed");
    await user.click(screen.getByRole("button",{name:"Live-Monitor starten"}));
    await waitFor(()=>expect(api.countECoSLocomotives).toHaveBeenCalledTimes(2));
    await user.click(screen.getByRole("tab",{name:"ECoS-Meldungen"}));
    vi.mocked(api.getECoSLiveStatus).mockRejectedValueOnce(new Error("Refresh failed"));
    await user.click(screen.getByRole("button",{name:"Status aktualisieren"}));
    await screen.findByText("Refresh failed");
    expect(screen.getByRole("button",{name:"Status aktualisieren"})).toBeEnabled();
  });

  it("does not load administrator settings or start a monitor for a read-only viewer",async()=>{
    render(<SettingsDigitalTab canManageUsers={false} username="viewer" formatDateTime={value=>value}/>);
    expect(api.digitalSettings).not.toHaveBeenCalled();
    expect(api.getECoSLiveStatus).not.toHaveBeenCalled();
    expect(screen.getByRole("button",{name:"Verbindung testen"})).toBeDisabled();
    expect(screen.getByRole("button",{name:"Adapter aktivieren"})).toBeDisabled();
    expect(screen.getByRole("button",{name:"Live-Monitor starten"})).toBeDisabled();
  });
  it("keeps adapter inputs after a rejected save and acknowledges only the successful retry", async () => {
    const user = userEvent.setup();
    render(<SettingsDigitalTab canManageUsers username="admin" formatDateTime={value => value} />);
    await screen.findByText("192.168.2.151:15471");
    await user.click(screen.getAllByRole("button", { name: "Konfiguration ändern" })[0]);
    const dialog = screen.getByRole("dialog");
    await user.clear(within(dialog).getByLabelText("Port"));
    await user.type(within(dialog).getByLabelText("Port"), "3456");
    vi.mocked(api.updateDigitalSettings).mockRejectedValueOnce(new Error("Save unavailable"));
    await user.click(within(dialog).getByRole("button", { name: "Speichern" }));
    await within(dialog).findByText("Save unavailable");
    expect(within(dialog).getByLabelText("Port")).toHaveValue("3456");
    expect(window.localStorage.getItem("railkeeper.digital.configurationMeta")).toBeNull();
    expect(screen.queryByText("Adapter-Konfiguration gespeichert.")).not.toBeInTheDocument();
    await user.click(within(dialog).getByRole("button", { name: "Speichern" }));
    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
    expect(api.updateDigitalSettings).toHaveBeenCalledTimes(2);
    expect(api.updateDigitalSettings).toHaveBeenLastCalledWith(expect.objectContaining({ ecos: expect.objectContaining({ port: "3456" }) }));
  });

  it("prevents editing or closing a configuration while its save acknowledgement is pending", async () => {
    let finish: () => void = () => { throw new Error("save has not started"); };
    vi.mocked(api.updateDigitalSettings).mockImplementationOnce(input => new Promise(resolve => { finish = () => resolve(input); }));
    const user = userEvent.setup();
    render(<SettingsDigitalTab canManageUsers username="admin" formatDateTime={value => value} />);
    await screen.findByText("192.168.2.151:15471");
    await user.click(screen.getAllByRole("button", { name: "Konfiguration ändern" })[0]);
    const dialog = screen.getByRole("dialog");
    await user.click(within(dialog).getByRole("button", { name: "Speichern" }));
    expect(within(dialog).getByLabelText("Port")).toBeDisabled();
    expect(within(dialog).getByRole("button", { name: "Schließen" })).toBeDisabled();
    expect(within(dialog).getByRole("button", { name: "Wird gespeichert..." })).toBeDisabled();
    await act(async () => finish());
    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
  });

});
