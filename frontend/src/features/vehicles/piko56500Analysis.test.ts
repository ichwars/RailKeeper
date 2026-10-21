import { describe, expect, it } from "vitest";
import { translate } from "../../shared/i18n";
import { pikoDETranslations } from "../../shared/i18n/piko.de";
import { pikoENTranslations } from "../../shared/i18n/piko.en";
import { decoderAnalysis, decoderMapping } from "./decoderAnalysis";
import { cvExplanationRows } from "./cvExplanationRows";
import { buildSpeedCurveProfiles } from "./speedCurve";
import { functionCVSuggestion } from "./functionCVSuggestions";
import type { FunctionEdit } from "./vehicleTransforms";

const label = "PIKO 56500";
const rows = (pairs: number[][], profile = label) => [[8, 162], ...pairs]
  .map(([cvNumber, value]) => ({ cvNumber, value, decoderProfile: profile }));
const analyze = (pairs: number[][]) => decoderAnalysis(rows(pairs), label);
const t = (key: string, values?: Record<string, string | number>) => translate("de", key, values);
const bank = (start: number, bytes: number[], indexCV31 = 18, indexCV32 = 0) => bytes.map((value, i) => ({
  cvNumber: start + i, value, decoderProfile: label, indexCV31, indexCV32
}));
const outputAnalysis = (bytes: number[], base = [[180, 6]]) => decoderAnalysis([...rows(base), ...bank(257, bytes)], label);

describe("PIKO56500, exact manual and explicit source conflicts", () => {
  it("requires the exact label and manufacturer and never enables ESU flags", () => {
    expect(analyze([]).piko).toBeDefined(); expect(analyze([]).esu5).toBe(false);
    expect(analyze([]).railComPlus).toBeUndefined();
    for (const profile of ["56500", "SmartDecoder XP 5.1 PluX22"]) expect(decoderAnalysis(rows([], profile), profile).piko).toBeDefined();
    for (const profile of ["PIKO 56501", "PIKO SmartDecoder XP", "SmartDecoder XP 5.1 Next18", "PIKO 56122", "Zentrale"])
      expect(decoderAnalysis(rows([], profile), profile).piko).toBeUndefined();
    expect(analyze([[8, 151]]).piko).toBeUndefined();
    expect(decoderAnalysis(rows([]).map(r => ({ ...r, protocol: "MM" })), label).piko).toBeUndefined();
    expect(decoderAnalysis([{ cvNumber: 29, value: 14, decoderProfile: label }], label).piko).toBeUndefined();
  });
  it("retains unknown and invalid register values instead of borrowing foundation meanings", () => {
    const a = analyze([[101, 20], [250, 10], [12, 2], [27, 4], [28, 8], [29, 64], [62, 2], [96, 2], [115, 8], [157, 100]]);
    for (const cv of [101, 250, 12, 27, 28, 29, 62, 96, 115, 157]) expect(a.get(cv)).toBeUndefined();
    expect(a.address).toBeUndefined(); expect(a.abc).toBeUndefined();
    expect(analyze([[29, 32], [17, 192], [18, 127]]).address).toBeUndefined();
    expect(analyze([[29, 32], [17, 192], [18, 128]]).address).toBe(128);
    expect(analyze([[29, 32], [17, 231], [18, 16]]).address).toBeUndefined();
  });
  it.each([[14,117,true,true], [10,117,false,false], [14,4,false,false], [14,1,true,false]])(
    "combines protocol permission with CV29: %i/%i", (config, protocols, dc, ac) => {
      const a = analyze([[29, config], [12, protocols]]);
      expect(a.dcAnalog).toBe(dc); expect(a.piko?.acAnalog).toBe(ac);
    });
  it("reports separate RailCom and RailComPlus permissions without decoding ESU names", () => {
    expect(analyze([[29,14],[28,0]]).railcomEnabled).toBe(true);
    expect(analyze([[29,14],[28,151]]).railcomEnabled).toBe(true);
    expect(analyze([[29,6],[28,151]]).piko?.railcomPlus).toBe(false);
    expect(analyze([[29,14],[28,128]]).piko?.railcomPlus).toBe(true);
    expect(analyze([[29,14]]).railcomEnabled).toBe(true);
    expect(analyze([[28,151]]).piko?.railcomPlus).toBeUndefined();
  });
  it.each([[0,0,0,false], [0,0,30,false], [1,0,30,false], [1,1,30,true], [2,16,30,true], [2,1,30,false], [4,0,30,true], [8,0,30,true], [16,0,30,undefined]])(
    "uses triggers and permission, not distance:145=%i/27=%i/141=%i", (triggers, brake, distance, result) => {
      const a = analyze([[145,Number(triggers)],[27,Number(brake)],[141,Number(distance)],[142,0]]);
      expect(a.constantBrake).toBe(result);
    });
  it("preserves partial braking information and the separate test-key semantics", () => {
    expect(analyze([[141,30]]).constantBrake).toBeUndefined();
    expect(analyze([[145,1],[142,0]]).constantBrake).toBeUndefined();
    expect(analyze([[142,3]]).constantBrake).toBe(true);
    expect(analyze([[145,0],[142,0]]).constantBrake).toBe(false);
    expect(analyze([[142,69]]).piko?.testKey).toBeUndefined();
    expect(analyze([[48,0],[51,30],[52,50]]).piko?.abcThresholdV).toBeCloseTo(1.6);
    expect(cvExplanationRows(analyze([[51,30]]),"braking",t).find(r=>r.key==='piko-shuttleWait')?.reading).toBe('3000 ms');
  });
  it("only interprets shifted simple mapping in mode1, and marks unavailable outputs", () => {
    const pairs=[[33,1],[34,2],[35,4],[38,1],[42,64],[43,1],[46,128]];
    expect(decoderMapping(analyze(pairs),t)).toEqual([]);
    expect(decoderMapping(analyze([[96,6],...pairs]),t)).toEqual([]);
    const m=decoderMapping(analyze([[96,1],...pairs]),t);
    expect(m.map(r=>[r.functionKey,r.direction,r.actions,r.warning])).toEqual([
      ['F0','forward',['Licht vorne'],false],['F0','reverse',['Licht hinten'],false],
      ['F1','both',['A1'],false],['F4','both',['A2'],false],
      ['F8','both',['A8 (beim56500 nicht vorhanden)'],true],['F9','both',['A5'],false],
      ['F12','both',['A12 (beim56500 nicht vorhanden)'],true]
    ]);
  });
  it("covers every output and both parameter sets only on the explicit bank18/0", () => {
    const a=decoderAnalysis([...rows([]),...bank(257,Array.from({length:90},(_,i)=>i%5===1?64:0))],label);
    expect(a.piko?.outputSets).toHaveLength(18);
    expect(a.piko?.outputSets.at(-1)).toMatchObject({output:8,set:'B',start:342,effect:0,pwm:64});
    expect(a.libraryMatch.knownValues).toBe(91);
    expect(decoderAnalysis([...rows([]),...bank(257,[1,64,31,0,0],1,1)],label).piko?.outputSets).toEqual([]);
    expect(decoderAnalysis([...rows([[31,18],[32,0]]),...bank(257,[1,64,31,0,0]).map(({indexCV31:_,indexCV32:__,...row})=>row)],label).piko?.outputSets).toEqual([]);
    const explanation=cvExplanationRows(a,'outputs',t).find(r=>r.key==='piko-output-8-B');
    expect(explanation?.mappingSource).toEqual({cvNumber:342,indexCV31:18,indexCV32:0});
  });
  it.each([[0,false],[1,false],[4,false],[5,true],[31,true],[64,false],[255,true]])(
    "auto-off requires direction AND speed flags: %i",(flags,expected)=>{
      expect(outputAnalysis([1,64,Number(flags),0,0]).piko?.outputSets[0].autoOff).toBe(expected);
    });
  it("treats smoke flags and missing/unknown effects as unresolved and validates PWM/timers",()=>{
    const normal=outputAnalysis([1,64,197,0,0]).piko?.outputSets[0];
    expect(normal).toMatchObject({autoOff:true,autoOffSeconds:3,fadeIn:true,fadeOut:true});
    for(const effect of [40,41,99,256]){
      const o=outputAnalysis([effect,65,255,20,30]).piko?.outputSets[0];
      expect(o?.autoOff).toBeUndefined();expect(o?.fadeIn).toBeUndefined();expect(o?.pwm).toBeUndefined();
    }
    const partial=decoderAnalysis([...rows([]),...bank(258,[20,31])],label);
    expect(partial.piko?.outputSets[0]).toMatchObject({effect:undefined,autoOff:undefined,sourceCV:258});
    const a=decoderAnalysis([...rows([[182,64]]),...bank(277,[1,64,31,0,0])],label);
    expect(a.piko?.outputSets[0].autoOffSeconds).toBeUndefined();
  });
  it("keeps contradictory firmware thresholds unknown even with a software byte",()=>{
    for(const version of [1,16,19,255]){
      const a=analyze([[7,version],[102,10],[103,20],[104,30]]);
      expect(a.firmware).toBeUndefined();
      expect(cvExplanationRows(a,'communication',t).find(r=>r.key==='piko-analogThresholds')?.reading).toBe('Unbekannt');
      for(const cv of [102,103,104])expect(a.definition({cvNumber:cv,value:0})?.definition.factoryDefault).toBeUndefined();
    }
  });
  it("plots only the selected valid speed characteristic and preserves raw trims",()=>{
    const curve=(extra:number[][])=>buildSpeedCurveProfiles(rows([[2,0],[6,100],[5,200],[66,128],[95,128],...extra]),'Decoder')[0];
    expect(curve([[29,14]])).toMatchObject({primaryCurve:'threePoint',forwardTrim:128,reverseTrim:128,pikoCurve:true});
    expect(curve([[29,30]])).toMatchObject({primaryCurve:'none'});
    expect(curve([[29,30],[67,10],[94,255]])).toMatchObject({primaryCurve:'speedTable'});
    expect(curve([]).primaryCurve).toBe('none');
    expect(curve([[29,64],[67,10]]).primaryCurve).toBe('none');
    expect(curve([[29,14],[2,-1],[6,256]]).threePoint).toHaveLength(1);
  });
  it("never mutates raw provenance, manual fields or ESU-only takeover",()=>{
    const values=[...rows([[96,1],[35,4]]),...bank(257,[1,64,31,0,0])].map(r=>({...r,sourceLine:9,sourceExportedAt:'2026-10-20T15:00:00'}));
    const before=structuredClone(values),a=decoderAnalysis(values,label);
    for(const current of [{},{name:'Mein Licht',mode:'moment',inverted:true,notes:'Manuell'}] satisfies FunctionEdit[]){
      const snapshot=structuredClone(current);
      for(const key of [0,1,3])expect(functionCVSuggestion(a,`F${key}`,current,[],t)).toEqual({ notesFull: false });
      expect(current).toEqual(snapshot);
    }
    cvExplanationRows(a,'outputs',t);expect(values).toEqual(before);
  });
  it("has complete German and English labels with matching placeholders",()=>{
    expect(Object.keys(pikoDETranslations).sort()).toEqual(Object.keys(pikoENTranslations).sort());
    for(const key of Object.keys(pikoDETranslations)){
      expect(pikoDETranslations[key]).not.toBe('');expect(pikoENTranslations[key]).not.toBe('');
      expect(pikoDETranslations[key].match(/\{\w+\}/g)).toEqual(pikoENTranslations[key].match(/\{\w+\}/g));
    }
  });
});
