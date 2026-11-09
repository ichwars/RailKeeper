import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createRef } from "react";
import { describe, expect, it, vi } from "vitest";
import { AppFileButton } from "./AppFileButton";

describe("AppFileButton", () => {
  it("opens from the keyboard and forwards a complete multiple-file selection", async () => {
    const ref = createRef<HTMLInputElement>();
    const names: string[][] = [];
    render(<AppFileButton ref={ref} multiple accept="text/plain" aria-label="Import files"
      onChange={event => names.push(Array.from(event.target.files || [], file => file.name))}>Browse</AppFileButton>);
    const click = vi.spyOn(ref.current!, "click");
    const button = screen.getByRole("button", { name: "Import files" });
    expect(screen.getAllByRole("button")).toEqual([button]);
    button.focus();
    await userEvent.keyboard("{Enter}");
    expect(click).toHaveBeenCalledTimes(1);
    await userEvent.upload(ref.current!, [new File(["a"], "a.txt", { type: "text/plain" }),
      new File(["b"], "b.txt", { type: "text/plain" })]);
    expect(names).toEqual([["a.txt", "b.txt"]]);
  });

  it("keeps both the native picker and visible trigger disabled", () => {
    const ref = createRef<HTMLInputElement>();
    render(<AppFileButton ref={ref} disabled>Browse</AppFileButton>);
    expect(screen.getByRole("button")).toBeDisabled();
    expect(ref.current).toBeDisabled();
  });
});
