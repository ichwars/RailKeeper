import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { TableColumnResizeHandle } from "./TableColumnResizeHandle";

describe("TableColumnResizeHandle", () => {
  it("exposes bounded separator semantics", () => {
    render(<TableColumnResizeHandle
      label="Breite von Name ändern"
      width={160}
      minWidth={100}
      maxWidth={300}
      defaultWidth={180}
      onPreview={vi.fn()}
      onCommit={vi.fn()}
    />);

    const handle = screen.getByRole("separator", { name: "Breite von Name ändern" });
    expect(handle).toHaveAttribute("aria-orientation", "vertical");
    expect(handle).toHaveAttribute("aria-valuemin", "100");
    expect(handle).toHaveAttribute("aria-valuemax", "300");
    expect(handle).toHaveAttribute("aria-valuenow", "160");
  });

  it("previews pointer movement and commits once when dragging ends", () => {
    const onPreview = vi.fn();
    const onCommit = vi.fn();
    render(<TableColumnResizeHandle label="Name" width={160} minWidth={100} maxWidth={300}
      defaultWidth={180} onPreview={onPreview} onCommit={onCommit} />);

    const handle = screen.getByRole("separator", { name: "Name" });
    fireEvent.pointerDown(handle, { clientX: 100, pointerId: 7, button: 0 });
    fireEvent.pointerMove(handle, { clientX: 154, pointerId: 7 });
    fireEvent.pointerUp(handle, { clientX: 154, pointerId: 7 });

    expect(onPreview).toHaveBeenLastCalledWith(214);
    expect(onCommit).toHaveBeenCalledOnce();
    expect(onCommit).toHaveBeenCalledWith(214);
  });

  it("does not persist clicks and restores previews when pointer dragging is cancelled", () => {
    const onPreview = vi.fn();
    const onCommit = vi.fn();
    render(<TableColumnResizeHandle label="Name" width={160} minWidth={100} maxWidth={300}
      defaultWidth={180} onPreview={onPreview} onCommit={onCommit} />);

    const handle = screen.getByRole("separator", { name: "Name" });
    fireEvent.pointerDown(handle, { clientX: 100, pointerId: 7, button: 0 });
    fireEvent.pointerUp(handle, { clientX: 100, pointerId: 7 });
    expect(onCommit).not.toHaveBeenCalled();

    fireEvent.pointerDown(handle, { clientX: 100, pointerId: 8, button: 0 });
    fireEvent.pointerMove(handle, { clientX: 154, pointerId: 8 });
    fireEvent.pointerCancel(handle, { clientX: 154, pointerId: 8 });

    expect(onPreview.mock.calls.map(([width]) => width)).toEqual([214, 160]);
    expect(onCommit).not.toHaveBeenCalled();
  });

  it("persists only the reset during a complete double-click sequence", () => {
    const onCommit = vi.fn();
    render(<TableColumnResizeHandle label="Name" width={160} minWidth={100} maxWidth={300}
      defaultWidth={180} onPreview={vi.fn()} onCommit={onCommit} />);

    const handle = screen.getByRole("separator", { name: "Name" });
    for (const pointerId of [7, 8]) {
      fireEvent.pointerDown(handle, { clientX: 100, pointerId, button: 0 });
      fireEvent.pointerUp(handle, { clientX: 100, pointerId });
    }
    fireEvent.doubleClick(handle);

    expect(onCommit).toHaveBeenCalledOnce();
    expect(onCommit).toHaveBeenCalledWith(180);
  });

  it("supports bounded keyboard steps, endpoints, and double-click reset", () => {
    const onCommit = vi.fn();
    render(<TableColumnResizeHandle label="Name" width={160} minWidth={100} maxWidth={300}
      defaultWidth={180} onPreview={vi.fn()} onCommit={onCommit} />);

    const handle = screen.getByRole("separator", { name: "Name" });
    fireEvent.keyDown(handle, { key: "ArrowRight" });
    fireEvent.keyDown(handle, { key: "ArrowLeft", shiftKey: true });
    fireEvent.keyDown(handle, { key: "Home" });
    fireEvent.keyDown(handle, { key: "End" });
    fireEvent.doubleClick(handle);

    expect(onCommit.mock.calls.map(([width]) => width)).toEqual([168, 128, 100, 300, 180]);
  });
});
