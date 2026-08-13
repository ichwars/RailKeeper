import { useEffect, useRef, type KeyboardEvent, type RefObject } from "react";

const focusableSelector = [
  "button:not([disabled])",
  "input:not([disabled]):not([type='hidden'])",
  "select:not([disabled])",
  "textarea:not([disabled])",
  "a[href]",
  "[contenteditable='true']",
  "[tabindex]:not([tabindex='-1'])"
].join(",");

function focusableElements(container: HTMLElement | null) {
  return Array.from(container?.querySelectorAll<HTMLElement>(focusableSelector) || []).filter((element) => {
    if (element.tabIndex < 0 || element.closest("[hidden], [aria-hidden='true'], [inert]")) return false;
    const style = window.getComputedStyle(element);
    return style.display !== "none" && style.visibility !== "hidden";
  });
}

export function useModalDialogLayer(
  onClose: () => void,
  initialFocusRef: RefObject<HTMLElement | null>
) {
  const layerRef = useRef<HTMLDivElement | null>(null);
  const anchorRef = useRef<HTMLSpanElement | null>(null);
  const invokerRef = useRef<HTMLElement | null>(null);
  const onCloseRef = useRef(onClose);
  onCloseRef.current = onClose;

  useEffect(() => {
    invokerRef.current = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    const parentDialog = anchorRef.current?.closest<HTMLElement>("[role='dialog']") || null;
    const previousAriaHidden = parentDialog?.getAttribute("aria-hidden") ?? null;
    const previouslyInert = parentDialog?.hasAttribute("inert") ?? false;
    parentDialog?.setAttribute("aria-hidden", "true");
    parentDialog?.setAttribute("inert", "");
    initialFocusRef.current?.focus();

    return () => {
      if (parentDialog) {
        if (previousAriaHidden === null) parentDialog.removeAttribute("aria-hidden");
        else parentDialog.setAttribute("aria-hidden", previousAriaHidden);
        if (!previouslyInert) parentDialog.removeAttribute("inert");
      }
      if (invokerRef.current?.isConnected) invokerRef.current.focus();
    };
  }, [initialFocusRef]);

  const onKeyDown = (event: KeyboardEvent<HTMLDivElement>) => {
    event.stopPropagation();
    if (event.key === "Escape") {
      event.preventDefault();
      onCloseRef.current();
      return;
    }
    if (event.key !== "Tab") return;
    const focusable = focusableElements(layerRef.current);
    if (focusable.length === 0) return;
    const first = focusable[0];
    const last = focusable[focusable.length - 1];
    if ((!event.shiftKey && event.target === last) || (event.shiftKey && event.target === first)) {
      event.preventDefault();
      (event.shiftKey ? last : first).focus();
    }
  };

  return { anchorRef, layerRef, onKeyDown };
}
