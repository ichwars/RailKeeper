import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { followAppLink } from "./appNavigation";

describe("existing-view navigation without dropping queued work", () => {
  it.each([{}, { ctrlKey: true }, { metaKey: true }, { shiftKey: true }, { altKey: true }, { button: 1 }])(
    "keeps native modifier behavior for %j", options => {
      const navigate = vi.fn();
      let intercepted = false;
      render(<a href="/vehicles" onClick={event => {
        followAppLink(event, navigate);
        intercepted = event.defaultPrevented;
        event.preventDefault();
      }}>Open</a>);
      fireEvent.click(screen.getByRole("link"), options);
      const regularClick = Object.keys(options).length === 0;
      expect(intercepted).toBe(regularClick);
      expect(navigate).toHaveBeenCalledTimes(regularClick ? 1 : 0);
    }
  );

  it.each([{ href: "https://example.test/" }, { href: "/vehicles", target: "_blank" },
    { href: "/vehicles", download: "data.json" }])("keeps external and download navigation native", attributes => {
      const navigate = vi.fn();
      render(<a {...attributes} onClick={event => {
        followAppLink(event, navigate);
        expect(event.defaultPrevented).toBe(false);
        event.preventDefault();
      }}>Open</a>);
      fireEvent.click(screen.getByRole("link"));
      expect(navigate).not.toHaveBeenCalled();
    });
});
