import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { DigitalCentersView } from "./DigitalCentersView";

describe("DigitalCentersView", () => {
  it("renders the approved dedicated page heading topology", () => {
    const { container } = render(<DigitalCentersView roles={["Admin"]} />);

    expect(screen.getByText("DIGITALBETRIEB")).toHaveClass("eyebrow");
    expect(screen.getByRole("heading", { level: 1, name: "Digitalzentralen" }))
      .toBeInTheDocument();
    expect(screen.getByText("Zentralen, Live-Daten und Synchronisation in einer Arbeitsansicht."))
      .toBeInTheDocument();
    expect(container.querySelector(".digital-centers-workspace"))
      .toHaveAttribute("data-can-administer", "true");
  });
});
