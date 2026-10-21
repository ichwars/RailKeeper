import { createContext, useContext, useEffect, useState, type ReactNode } from "react";
import { api } from "./api";
import { bundledDecoderInventory, bundledDecoderProfiles, type DecoderInventory, type DecoderLibraryProfile } from "./decoderLibraryModel";

export const decoderLibraryChanged = "railkeeper-decoder-library-changed";
type LibraryState = { profiles: DecoderLibraryProfile[]; inventory?: DecoderInventory; failed: boolean };
const Context = createContext<LibraryState>({ profiles: bundledDecoderProfiles, inventory: bundledDecoderInventory, failed: false });
export const useDecoderLibrary = () => useContext(Context);

export function DecoderLibraryProvider({ children, enabled = true }: { children: ReactNode; enabled?: boolean }) {
  const [state, setState] = useState<LibraryState>({ profiles: [], failed: false });
  useEffect(() => {
    if (!enabled) return;
    let active = true, sequence = 0;
    const load = async () => {
      const request = ++sequence;
      try {
        const catalog = await api.decoderLibrary();
        if (active && request === sequence) setState({ profiles: catalog.profiles, inventory: catalog.inventory, failed: false });
      } catch {
        if (active && request === sequence) setState({ profiles: [], failed: true });
      }
    };
    void load();
    window.addEventListener(decoderLibraryChanged, load);
    return () => { active = false; window.removeEventListener(decoderLibraryChanged, load); };
  }, [enabled]);
  return <Context.Provider value={state}>{children}</Context.Provider>;
}
