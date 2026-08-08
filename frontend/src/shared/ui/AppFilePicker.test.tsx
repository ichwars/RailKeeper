import { fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createRef } from "react";
import { describe, expect, it, vi } from "vitest";

import { AppFilePicker } from "./AppFilePicker";

describe("AppFilePicker", () => {
  it("uses a hidden native file input and exposes the chosen filename", () => {
    const ref = createRef<HTMLInputElement>();
    const onFileChange = vi.fn();
    const file = new File(["manual"], "Anleitung.pdf", { type: "application/pdf" });

    render(
      <AppFilePicker
        ref={ref}
        label="Dokument"
        helpText="PDF oder Bild"
        file={file}
        onFileChange={onFileChange}
      />
    );

    const input = screen.getByLabelText("Dokument");
    expect(ref.current).toBe(input);
    expect(input).toHaveAttribute("type", "file");
    expect(input).toHaveClass("visually-hidden");
    expect(screen.getByText("Anleitung.pdf")).toBeVisible();
  });

  it("provides keyboard-operable choose and clear actions", async () => {
    const user = userEvent.setup();
    const onFileChange = vi.fn();
    const file = new File(["invoice"], "Rechnung.pdf", { type: "application/pdf" });

    render(<AppFilePicker label="Beleg" file={file} onFileChange={onFileChange} error="Datei ist zu groß" />);
    const input = screen.getByLabelText("Beleg");
    const clickSpy = vi.spyOn(input, "click");

    await user.tab();
    expect(screen.getByRole("button", { name: "Datei auswählen" })).toHaveFocus();
    await user.keyboard("{Enter}");
    expect(clickSpy).toHaveBeenCalledOnce();

    await user.click(screen.getByRole("button", { name: "Datei entfernen" }));
    expect(onFileChange).toHaveBeenCalledWith(null);
    expect(input).toHaveAttribute("aria-invalid", "true");

    fireEvent.change(input, { target: { files: [file] } });
    expect(onFileChange).toHaveBeenLastCalledWith(file);
  });
});
