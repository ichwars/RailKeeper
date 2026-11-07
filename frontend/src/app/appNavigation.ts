import type { MouseEvent } from "react";

export function followAppLink(event: MouseEvent<HTMLAnchorElement>, navigate?: () => void) {
  const link = event.currentTarget;
  if (!navigate || event.defaultPrevented || event.button !== 0 || event.ctrlKey || event.metaKey ||
    event.shiftKey || event.altKey || link.hasAttribute("download") ||
    (link.target && link.target !== "_self") || new URL(link.href).origin !== window.location.origin) return;
  event.preventDefault();
  navigate();
}
