import { useEffect } from "react";
import { act, fireEvent, render, screen, within } from "@testing-library/react";
import { expect, it, vi } from "vitest";
import type { MasterDataEntry, VehicleCVValueInput } from "../../shared/api";
import { translate } from "../../shared/i18n";
import { cvValueFixture, functionFixture, vehicleFixture } from "../../test/fixtures/vehicles";
import { decoderAnalysis } from "./decoderAnalysis";
import { functionCVSuggestion } from "./functionCVSuggestions";
import { useVehicleFunctionsController } from "./useVehicleFunctionsController";
import { VehicleFunctionsTab } from "./VehicleFunctionsTab";
import { emptyVehicle } from "./vehicleViewModel";

const profile = "LokPilot 5 micro DCC";
const cv = (cvNumber: number, value: number): VehicleCVValueInput => ({ cvNumber, value, decoderProfile: profile });
function mapping(row: number, a: number, b: number, output: number, logical = 0) {
  const start = 257 + row * 16;
  return [...[a,b,0,0,0,0,0,0,0].map((value, i) => ({ ...cv(start+i,value), indexCV31:16,indexCV32:3 })),
    ...[output,0,0,logical,0,0,0].map((value,i) => ({ ...cv(start+i,value),indexCV31:16,indexCV32:8 }))];
}
const readings = [cv(8,151), { ...cv(7,255), description: `LokProgrammer Decoder: ${profile} (5.14.183)` },
  ...mapping(0,20,0,1), ...mapping(1,24,0,2), ...mapping(2,64,0,4), ...mapping(3,0,1,8), ...mapping(4,0,4,0,2),
  ...[0,3,32,2,32,2,96,9].map((value,i) => ({...cv(257+i,value),indexCV31:1,indexCV32:1,
    sourceLine:175+i,sourceExportedAt:"2026-10-17T17:41:59"}))];
const symbols: MasterDataEntry[] = [
  { id:"light",type:"symbols",key:"light",label:"Licht",active:true,sortOrder:1,metadata:{ecosCode:3,category:"Licht"},createdAt:"",updatedAt:"" },
  { id:"generic",type:"symbols",key:"generic",label:"Allgemeine Funktion",active:true,sortOrder:2,metadata:{ecosCode:2},createdAt:"",updatedAt:"" },
  { id:"shunt",type:"symbols",key:"shunt",label:"Rangiergang",active:true,sortOrder:3,metadata:{ecosCode:9},createdAt:"",updatedAt:"" }
];
const source = "RailComPlus · Betriebsart: Dauer; Invertiert: nein · CV31=1 / CV32=1 · CV257=0; CV258=3 · Originalzeile 175 / 176";
const vehicle = vehicleFixture({ cvValues:readings.map((row,i)=>cvValueFixture({...row,id:`cv-${i}`})), functions:[
  functionFixture({symbolKey:"light",notes:"Meine Notiz\n"+source,revision:"a".repeat(64),directionDependent:null,inverted:false}),
  functionFixture({functionKey:"F1",name:"",symbolKey:"",revision:"b".repeat(64),directionDependent:null,inverted:false})
] });
let controller: ReturnType<typeof useVehicleFunctionsController>;
function Harness({ readonly = false, data = vehicle }: { readonly?: boolean; data?: typeof vehicle }) {
  controller = useVehicleFunctionsController({selected:data,form:emptyVehicle,setSaving:vi.fn(),onMessage:vi.fn(),refreshSelectedVehicle:vi.fn()});
  useEffect(()=>controller.commands.loadDetail(data.functions,data.id),[data]);
  return <VehicleFunctionsTab selected={data} readonly={readonly} saving={false}
    functionImportInputRef={controller.refs.importInputRef} configuredFunctionKeys={controller.state.configuredKeys}
    showConfiguredFunctionsOnly={controller.state.showConfiguredOnly} symbols={symbols}
    onImportFunctions={controller.commands.importValues} onExportFunctions={controller.commands.exportValues}
    onShowConfiguredFunctionsOnlyChange={controller.setters.setShowConfiguredOnly}
    functionEdit={controller.commands.functionEdit} updateFunctionEdit={controller.commands.updateFunctionEdit}
    inferFunctionTypeFromSymbol={(_key,_symbols,fallback)=>fallback||"standard"} deleteFunction={controller.commands.remove} />;
}

it("shows CV-mapped keys by default in one list, with a collapsed mapping last", () => {
  const {container} = render(<Harness />);
  expect(screen.getByRole("checkbox",{name:"Nur belegte"})).toBeChecked();
  expect(container.querySelectorAll(".function-row")).toHaveLength(4);
  expect(screen.getByLabelText("F3 Funktionsname")).toBeVisible();
  expect(screen.queryByText("RailComPlus-Anmeldedaten")).not.toBeInTheDocument();
  expect(screen.queryByRole("button",{name:"F0 speichern"})).not.toBeInTheDocument();
  expect(container.querySelector(".function-mapping-details")?.hasAttribute("open")).toBe(false);
  expect(container.querySelector(".function-editor")?.lastElementChild).toHaveClass("function-mapping-details");
  fireEvent.click(screen.getByRole("checkbox",{name:"Nur belegte"}));
  expect(container.querySelectorAll(".function-row")).toHaveLength(32);
  expect(screen.getAllByText("Nicht belegt").length).toBeGreaterThan(0);
});

it("opens the symbol picker from the marked control and preserves a manual name", () => {
  render(<Harness />);
  fireEvent.change(screen.getByLabelText("F1 Funktionsname"),{target:{value:"Mein Name"}});
  const picker=screen.getByLabelText("F1 Symbol");
  expect(picker.closest("details")).toHaveClass("missing-symbol");
  fireEvent.click(picker);
  expect(picker.closest("details")).toHaveAttribute("open");
  fireEvent.click(within(picker.closest("details")!).getByRole("button",{name:/^Licht$/}));
  expect(screen.getByLabelText("F1 Funktionsname")).toHaveValue("Mein Name");
  expect(picker.closest("details")).not.toHaveClass("missing-symbol");
  expect(controller.commands.changes()[0].value).toMatchObject({name:"Mein Name",symbolKey:"light"});
});

it("fills all safe fields together, retaining manual settings and historical source text", () => {
  render(<Harness />);
  act(()=>controller.commands.updateFunctionEdit("F1",{mode:"moment",inverted:true}));
  fireEvent.click(screen.getByRole("button",{name:"Fehlende Angaben aus CVs ergänzen"}));
  expect(controller.commands.functionEdit("F1")).toMatchObject({name:"",symbolKey:"generic",mode:"moment",inverted:true});
  expect(controller.commands.functionEdit("F3")).toMatchObject({name:"Rangiergang",symbolKey:"shunt",mode:"dauer",inverted:false});
  expect(controller.commands.functionEdit("F0").notes).toBe("Meine Notiz\n"+source);
  const f3=controller.commands.changes().find(change=>change.functionKey==="F3");
  expect(f3?.value?.notes).toContain("CV263=96; CV264=9");
  expect(f3?.value?.notes).toContain("181 / 182");
  expect(f3?.value?.notes).toContain("2026-10-17T17:41:59");
  expect(f3?.value?.notes).toContain("F0-/F1-/F3-Vergleich");
});

it("keeps provenance under Details and preserves it when editing the own note", () => {
  const {container}=render(<Harness />);
  const row=container.querySelector(".function-row")!;
  fireEvent.click(row.querySelector(".function-row-details > summary")!);
  expect(screen.getByLabelText("F0 Notiz")).toHaveValue("Meine Notiz");
  fireEvent.change(screen.getByLabelText("F0 Notiz"),{target:{value:"Neue Notiz"}});
  expect(controller.commands.functionEdit("F0").notes).toBe("Neue Notiz\n"+source);
  fireEvent.click(within(row as HTMLElement).getByText("Herkunft und CV-Werte ansehen"));
  expect(within(row as HTMLElement).getByText(source)).toBeVisible();
});

it("preserves read-only access and leaves an empty list usable", () => {
  const {rerender,container}=render(<Harness readonly />);
  expect(screen.getByRole("button",{name:"Fehlende Angaben aus CVs ergänzen"})).toBeDisabled();
  fireEvent.click(screen.getByLabelText("F1 Symbol"));
  expect(screen.getByLabelText("F1 Symbol").closest("details")).not.toHaveAttribute("open");
  expect(controller.commands.changes()).toEqual([]);
  rerender(<Harness data={vehicleFixture({cvValues:[],functions:[]})} />);
  expect(container.querySelectorAll(".function-row")).toHaveLength(0);
  expect(screen.getByText(/Noch keine Funktionstasten/)).toBeVisible();
});

it("retains firmware and notes guards in the combined action", () => {
  const t=(key:string,args?:Record<string,string|number>)=>translate("de",key,args);
  const older=readings.map(row=>row.cvNumber===7?{...row,description:`LokProgrammer Decoder: ${profile} (5.14.184)`}:row);
  const candidate=functionCVSuggestion(decoderAnalysis(older,profile),"F1",{},symbols,t);
  expect(candidate.patch).toMatchObject({symbolKey:"generic"});
  expect(candidate.patch).not.toHaveProperty("mode");
  expect(candidate.patch).not.toHaveProperty("inverted");
  expect(functionCVSuggestion(decoderAnalysis(readings,profile),"F3",{notes:"ä".repeat(500)},symbols,t)).toEqual({notesFull:true});
});

it("adds the untouched F0 default explicitly without saving when no CV mapping exists", () => {
  const {container}=render(<Harness data={vehicleFixture({cvValues:[],functions:[]})} />);
  fireEvent.click(screen.getByRole("checkbox",{name:"Nur belegte"}));
  expect(controller.commands.changes()).toEqual([]);
  fireEvent.click(screen.getByRole("button",{name:"F0 hinzufügen"}));
  expect(controller.commands.changes()).toMatchObject([{functionKey:"F0",expectedRevision:"",value:{name:"Fahrlicht",symbolKey:"light"}}]);
  fireEvent.click(screen.getByRole("checkbox",{name:"Nur belegte"}));
  expect(container.querySelectorAll(".function-row")).toHaveLength(1);
});
