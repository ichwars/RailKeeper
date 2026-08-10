import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import type { PlanFreeObject } from "../../shared/api";
import { FreePlanObjectLayer } from "./FreePlanObjectLayer";

const objects: PlanFreeObject[] = [
  {
    id: "rectangle", lineageId: "rectangle", revisionId: "revision-1", name: "Gebäude",
    category: "structure", positionXMm: 10, positionYMm: 20, rotationDegrees: 15,
    shape: { schemaVersion: 1, kind: "rectangle", widthMm: 100, heightMm: 50 },
    version: 1, createdAt: "now", updatedAt: "now"
  },
  {
    id: "ellipse", lineageId: "ellipse", revisionId: "revision-1", name: "Teich",
    category: "scenery", positionXMm: 200, positionYMm: 20, rotationDegrees: 0,
    shape: { schemaVersion: 1, kind: "ellipse", widthMm: 80, heightMm: 40 },
    version: 1, createdAt: "now", updatedAt: "now"
  },
  {
    id: "line", lineageId: "line", revisionId: "revision-1", name: "Trennlinie",
    category: "annotation", positionXMm: 0, positionYMm: 0, rotationDegrees: 0,
    shape: { schemaVersion: 1, kind: "line", endXMm: 100, endYMm: 30 },
    version: 1, createdAt: "now", updatedAt: "now"
  },
  {
    id: "label", lineageId: "label", revisionId: "revision-1", name: "Gleistext",
    category: "annotation", positionXMm: 50, positionYMm: 70, rotationDegrees: 0,
    shape: { schemaVersion: 1, kind: "label", text: "Gleis 1", fontSizeMm: 8 },
    version: 1, createdAt: "now", updatedAt: "now"
  }
];

describe("FreePlanObjectLayer", () => {
  it("renders semantic SVG shapes, category classes and a non-color selection marker", () => {
    render(<svg><FreePlanObjectLayer objects={objects} selectedID="rectangle" onSelect={vi.fn()} /></svg>);

    const rectangle = screen.getByRole("button", { name: "Gebäude" });
    expect(rectangle).toHaveAttribute("transform", "translate(10 20) rotate(15)");
    expect(rectangle).toHaveClass("free-plan-object-category-structure", "is-selected");
    expect(rectangle).toHaveAttribute("aria-pressed", "true");
    expect(rectangle.querySelector("rect.free-plan-object-selection")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Teich" }).querySelector("ellipse")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Trennlinie" }).querySelector("line")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Gleistext" }).querySelector("text")).toHaveTextContent("Gleis 1");
  });

  it("selects with pointer and keyboard", () => {
    const onSelect = vi.fn();
    render(<svg><FreePlanObjectLayer objects={objects} selectedID={null} onSelect={onSelect} /></svg>);

    fireEvent.click(screen.getByRole("button", { name: "Teich" }));
    fireEvent.keyDown(screen.getByRole("button", { name: "Gleistext" }), { key: "Enter" });
    expect(onSelect).toHaveBeenNthCalledWith(1, objects[1]);
    expect(onSelect).toHaveBeenNthCalledWith(2, objects[3]);
  });
});
