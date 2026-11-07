import { fireEvent, render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { languageSettingKey } from "../i18n";
import { AppDateInput } from "./AppDateInput";

function DateField({ initial = "2026-11-07", onCommit = vi.fn() }: { initial?: string; onCommit?: (value: string) => void }) {
  const [value, setValue] = useState(initial);
  return <><AppDateInput aria-label="Ereignisdatum" value={value} onChange={event => { setValue(event.target.value); onCommit(event.target.value); }} /><button type="button">Außerhalb</button></>;
}

describe("Date input preserves known and unknown event dates", () => {
  afterEach(() => vi.useRealTimers());
  it.each(["2028-02-29", "29.2.2028"])("commits the valid input %s as the same ISO date", async input => {
    const onCommit = vi.fn();
    render(<DateField onCommit={onCommit} />);
    const user = userEvent.setup();
    await user.clear(screen.getByLabelText("Ereignisdatum"));
    await user.type(screen.getByLabelText("Ereignisdatum"), `${input}{Enter}`);
    expect(onCommit).toHaveBeenLastCalledWith("2028-02-29");
    expect(screen.getByLabelText("Ereignisdatum")).toHaveValue("29.02.2028");
  });

  it.each(["29.02.2027", "31.11.2026", "2026-13-01", "unknown"])("rejects %s without inventing a replacement date", async input => {
    const onCommit = vi.fn();
    render(<DateField onCommit={onCommit} />);
    const user = userEvent.setup();
    await user.clear(screen.getByLabelText("Ereignisdatum"));
    await user.type(screen.getByLabelText("Ereignisdatum"), input);
    await user.click(screen.getByRole("button", { name: "Außerhalb" }));
    expect(onCommit).not.toHaveBeenCalled();
    expect(screen.getByLabelText("Ereignisdatum")).toHaveValue("07.11.2026");
  });

  it("restores an uncommitted draft on Escape and clears only after an explicit action", async () => {
    const onCommit = vi.fn();
    render(<DateField onCommit={onCommit} />);
    const user = userEvent.setup();
    await user.clear(screen.getByLabelText("Ereignisdatum"));
    await user.type(screen.getByLabelText("Ereignisdatum"), "12.12.2026{Escape}");
    expect(screen.getByLabelText("Ereignisdatum")).toHaveValue("07.11.2026");
    expect(onCommit).not.toHaveBeenCalled();
    await user.click(screen.getByRole("button", { name: "Datum leeren" }));
    expect(onCommit).toHaveBeenLastCalledWith("");
    expect(screen.getByLabelText("Ereignisdatum")).toHaveValue("");
  });

  it("opens an unknown date without substituting today, and closes on outside pointer", async () => {
    const onCommit = vi.fn();
    render(<DateField initial="" onCommit={onCommit} />);
    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "Kalender öffnen" }));
    expect(screen.getByRole("dialog", { name: "Datum auswählen" })).toBeInTheDocument();
    expect(onCommit).not.toHaveBeenCalled();
    await user.click(screen.getByRole("button", { name: "Außerhalb" }));
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    expect(screen.getByLabelText("Ereignisdatum")).toHaveValue("");
  });

  it("navigates months and chooses the exact displayed day using the keyboard opener", async () => {
    const onCommit = vi.fn();
    render(<DateField onCommit={onCommit} />);
    const user = userEvent.setup();
    await user.click(screen.getByLabelText("Ereignisdatum"));
    await user.keyboard("{ArrowDown}");
    const dialog = within(screen.getByRole("dialog", { name: "Datum auswählen" }));
    expect(dialog.getByText("November 2026")).toBeInTheDocument();
    await user.click(dialog.getByRole("button", { name: "Nächster Monat" }));
    expect(dialog.getByText("Dezember 2026")).toBeInTheDocument();
    await user.click(dialog.getByRole("button", { name: "Vorheriger Monat" }));
    await user.click(dialog.getByRole("button", { name: "15" }));
    expect(onCommit).toHaveBeenLastCalledWith("2026-11-15");
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("keeps the popup inside the viewport and updates its position on scrolling", async () => {
    render(<DateField />);
    const root = screen.getByLabelText("Ereignisdatum").closest(".app-date")!;
    vi.spyOn(root, "getBoundingClientRect").mockReturnValue({ x: 1100, y: 650, top: 650, bottom: 680, left: 1100, right: 1200, width: 100, height: 30, toJSON: () => ({}) });
    await userEvent.setup().click(screen.getByRole("button", { name: "Kalender öffnen" }));
    const dialog = screen.getByRole("dialog");
    expect(Number.parseFloat(dialog.style.left)).toBeLessThanOrEqual(window.innerWidth - 284);
    expect(Number.parseFloat(dialog.style.top)).toBeLessThan(650);
    fireEvent.scroll(window);
    fireEvent.resize(window);
    expect(dialog).toHaveStyle({ position: "fixed" });
  });

  it("offers English labels and an explicit clear without changing the date grammar", async () => {
    localStorage.setItem(languageSettingKey, "en");
    const onCommit = vi.fn();
    render(<DateField onCommit={onCommit} />);
    const user = userEvent.setup();
    expect(screen.getByLabelText("Ereignisdatum")).toHaveAttribute("placeholder", "DD.MM.YYYY");
    await user.click(screen.getByRole("button", { name: "Open calendar" }));
    const dialog = within(screen.getByRole("dialog", { name: "Choose date" }));
    expect(dialog.getByRole("button", { name: "Next month" })).toBeInTheDocument();
    expect(dialog.getByRole("button", { name: "Today" })).toBeInTheDocument();
    await user.click(dialog.getByRole("button", { name: "Clear" }));
    expect(onCommit).toHaveBeenLastCalledWith("");
  });

  it("records today only after its explicit calendar action", async () => {
    vi.useFakeTimers({ toFake: ["Date"] });
    vi.setSystemTime(new Date(2028, 1, 29, 12));
    const onCommit = vi.fn();
    render(<DateField initial="" onCommit={onCommit} />);
    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "Kalender öffnen" }));
    expect(onCommit).not.toHaveBeenCalled();
    await user.click(screen.getByRole("button", { name: "Heute" }));
    expect(onCommit).toHaveBeenLastCalledWith("2028-02-29");
    expect(screen.getByLabelText("Ereignisdatum")).toHaveValue("29.02.2028");
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("accepts an explicitly emptied text field on blur", async () => {
    const onCommit = vi.fn();
    render(<DateField onCommit={onCommit} />);
    const user = userEvent.setup();
    await user.clear(screen.getByLabelText("Ereignisdatum"));
    await user.tab();
    expect(onCommit).toHaveBeenLastCalledWith("");
  });

  it("disables editing and the calendar together", () => {
    const onChange = vi.fn();
    render(<AppDateInput aria-label="Datum" value="2026-11-07" disabled required onChange={onChange} />);
    expect(screen.getByLabelText("Datum")).toBeDisabled();
    expect(screen.getByLabelText("Datum")).toBeRequired();
    expect(screen.getByRole("button", { name: "Kalender öffnen" })).toBeDisabled();
    expect(screen.queryByRole("button", { name: "Datum leeren" })).not.toBeInTheDocument();
    expect(onChange).not.toHaveBeenCalled();
  });
});
