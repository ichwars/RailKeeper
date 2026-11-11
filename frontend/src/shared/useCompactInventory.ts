import { useEffect, useState } from "react";

const compactInventoryQuery = "(max-width: 900px)";

export function useCompactInventory() {
  const [compact, setCompact] = useState(() =>
    typeof window.matchMedia === "function" && window.matchMedia(compactInventoryQuery).matches);

  useEffect(() => {
    if (typeof window.matchMedia !== "function") return;
    const mediaQuery = window.matchMedia(compactInventoryQuery);
    const update = () => setCompact(mediaQuery.matches);
    update();
    mediaQuery.addEventListener("change", update);
    return () => mediaQuery.removeEventListener("change", update);
  }, []);

  return compact;
}
