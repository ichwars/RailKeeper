import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { api } from "../shared/api";
import { setLanguage } from "../shared/i18n";
import { App } from "./App";
import { markSetupWelcome, readSetupWelcome } from "./setupWelcome";

vi.mock("../features/setup/SetupView",()=>({SetupView:({onComplete}:{onComplete:(username:string)=>void})=><button onClick={()=>onComplete("first-owner")}>Complete successful setup</button>}));
vi.mock("../features/auth/LoginView",()=>({LoginView:({onLogin}:{onLogin:(session:object)=>void})=><button onClick={()=>onLogin({username:"first-owner",roles:["Admin"],csrfToken:"synthetic-csrf",twoFactorEnabled:false})}>Sign in</button>}));
vi.mock("../features/overview/OverviewView",()=>({OverviewView:()=> <h1>Collection overview</h1>}));
vi.mock("../features/exhibition/ExhibitionView",()=>({ExhibitionView:()=> <h1>Exhibition planning</h1>}));
vi.mock("../shared/DecoderLibraryProvider",()=>({DecoderLibraryProvider:({children}:{children:React.ReactNode})=>children}));
vi.mock("./Shell",()=>({Shell:({children}:{children:React.ReactNode})=>children}));

describe("setup-only personal view trigger",()=>{
 beforeEach(()=>{
  window.history.replaceState(null,"","/");setLanguage("de");
  vi.spyOn(api,"setupStatus").mockResolvedValue({setupRequired:false,cookieSecure:false});
  vi.spyOn(api,"session").mockResolvedValue({username:"first-owner",roles:["Admin"],csrfToken:"csrf",twoFactorEnabled:false});
  vi.spyOn(api,"profileSettings").mockResolvedValue({settings:{}});
  vi.spyOn(api,"updateProfileSettings").mockResolvedValue({settings:{}});
 });
 it("never treats missing preferences on an existing account as a new setup",async()=>{
  render(<App/>);await screen.findByRole("heading",{name:"Collection overview"});
  expect(screen.queryByRole("heading",{name:"Dein Einstieg in RailKeeper"})).not.toBeInTheDocument();
  expect(api.updateProfileSettings).not.toHaveBeenCalled();
 });
 it("offers the choice only after successful setup and the same user's login",async()=>{
  vi.mocked(api.setupStatus).mockResolvedValue({setupRequired:true,cookieSecure:false});
  render(<App/>);
  await userEvent.click(await screen.findByRole("button",{name:"Complete successful setup"}));
  expect(readSetupWelcome()).toBe("first-owner");
  expect(screen.queryByRole("heading",{name:"Dein Einstieg in RailKeeper"})).not.toBeInTheDocument();
  await userEvent.click(screen.getByRole("button",{name:"Sign in"}));
  await screen.findByRole("heading",{name:"Dein Einstieg in RailKeeper"});
  await userEvent.click(screen.getByRole("button",{name:"Vorerst bisherige Ansicht behalten"}));
  await screen.findByRole("heading",{name:"Collection overview"});
  expect(readSetupWelcome()).toBe("");expect(api.updateProfileSettings).not.toHaveBeenCalled();
 });
 it("retains the explicit setup marker across a reload until the user chooses",async()=>{
  markSetupWelcome("first-owner");render(<App/>);
  await screen.findByRole("heading",{name:"Dein Einstieg in RailKeeper"});
 });
 it.each([{username:"other-owner",roles:["Admin"],title:"Collection overview"},{username:"first-owner",roles:["Messe"],title:"Exhibition planning"}])("keeps a setup marker away from another account or a Messe-only session ($username/$roles)",async session=>{
  markSetupWelcome("first-owner");
  vi.mocked(api.session).mockResolvedValue({...session,csrfToken:"csrf",twoFactorEnabled:false});
  render(<App/>);await screen.findByRole("heading",{name:session.title});
  expect(screen.queryByRole("heading",{name:"Dein Einstieg in RailKeeper"})).not.toBeInTheDocument();
 });
});
