import { fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createRef, FormEvent, useState } from "react";
import { describe, expect, it, vi } from "vitest";

import { AppFilePicker } from "./AppFilePicker";

describe("AppFilePicker", () => {
  it("uses a hidden native file input and exposes the chosen filename", () => {
    const ref = createRef<HTMLInputElement>();
    const onFileChange = vi.fn();
    const file = new File(["manual"], "Anleitung.pdf", { type: "application/pdf" });

    const { container } = render(
      <AppFilePicker
        ref={ref}
        label="Dokument"
        helpText="PDF oder Bild"
        file={file}
        onFileChange={onFileChange}
        triggerLabel="Datei wählen"
        clearLabel="Datei entfernen"
        emptyLabel="Keine Datei"
      />
    );

    const input = container.querySelector('input[type="file"]');
    if (!(input instanceof HTMLInputElement)) throw new Error("expected file input");
    expect(ref.current).toBe(input);
    expect(input).toHaveAttribute("type", "file");
    expect(input).toHaveClass("visually-hidden");
    expect(screen.getByText("Anleitung.pdf")).toBeVisible();
    expect(screen.getByRole("button", { name: "Dokument Datei wählen Anleitung.pdf" })).toBeVisible();
  });

  it("provides keyboard-operable choose and clear actions", async () => {
    const user = userEvent.setup();
    const onFileChange = vi.fn();
    const file = new File(["invoice"], "Rechnung.pdf", { type: "application/pdf" });

    const { container } = render(
      <AppFilePicker
        label="Beleg"
        file={file}
        onFileChange={onFileChange}
        error="Datei ist zu groß"
        triggerLabel="Datei wählen"
        clearLabel="Datei entfernen"
        emptyLabel="Keine Datei"
      />
    );
    const input = container.querySelector('input[type="file"]');
    if (!(input instanceof HTMLInputElement)) throw new Error("expected file input");
    const clickSpy = vi.spyOn(input, "click");

    await user.tab();
    expect(screen.getByRole("button", { name: "Beleg Datei wählen Rechnung.pdf" })).toHaveFocus();
    await user.keyboard("{Enter}");
    expect(clickSpy).toHaveBeenCalledOnce();

    await user.click(screen.getByRole("button", { name: "Datei entfernen" }));
    expect(onFileChange).toHaveBeenCalledWith(null);
    expect(input).toHaveAttribute("aria-invalid", "true");

    fireEvent.change(input, { target: { files: [file] } });
    expect(onFileChange).toHaveBeenLastCalledWith(file);
  });

  it("clears the native input on controlled reset and permits same-file reselection", async () => {
    const user = userEvent.setup();
    const onFileChange = vi.fn();
    const file = new File(["manual"], "Anleitung.pdf", { type: "application/pdf" });
    const labels = {
      triggerLabel: "Choose file",
      clearLabel: "Remove file",
      emptyLabel: "No file selected"
    };
    const { container, rerender } = render(
      <AppFilePicker label="Document" file={null} onFileChange={onFileChange} {...labels} />
    );
    const input = container.querySelector('input[type="file"]');
    if (!(input instanceof HTMLInputElement)) throw new Error("expected file input");

    await user.upload(input, file);
    expect(onFileChange).toHaveBeenCalledWith(file);
    rerender(<AppFilePicker label="Document" file={file} onFileChange={onFileChange} {...labels} />);
    rerender(<AppFilePicker label="Document" file={null} onFileChange={onFileChange} {...labels} />);
    expect(input.files).toHaveLength(0);

    await user.click(screen.getByRole("button", { name: "Document Choose file No file selected" }));
    await user.upload(input, file);
    expect(onFileChange).toHaveBeenCalledTimes(2);
    expect(onFileChange).toHaveBeenLastCalledWith(file);
  });

  it("keeps the controlled file and native form value when choosing is cancelled", async () => {
    const user = userEvent.setup();
    const file = new File(["manual"], "Anleitung.pdf", { type: "application/pdf" });
    const labels = {
      triggerLabel: "Choose file",
      clearLabel: "Remove file",
      emptyLabel: "No file selected"
    };
    function Harness() {
      const [selectedFile, setSelectedFile] = useState<File | null>(null);
      return (
        <form>
          <AppFilePicker
            label="Document"
            name="document"
            file={selectedFile}
            onFileChange={setSelectedFile}
            {...labels}
          />
        </form>
      );
    }
    const { container } = render(<Harness />);
    const input = container.querySelector<HTMLInputElement>('input[type="file"]');
    const form = container.querySelector("form");
    if (!input || !form) throw new Error("expected file form control");

    await user.upload(input, file);
    expect(input.files?.[0]?.name).toBe("Anleitung.pdf");
    const initialFormFile = new FormData(form).get("document");
    if (!(initialFormFile instanceof File)) throw new Error("expected initial submitted file");
    await user.click(screen.getByRole("button", { name: "Document Choose file Anleitung.pdf" }));

    expect(screen.getByText("Anleitung.pdf")).toBeVisible();
    expect(input.files?.[0]?.name).toBe("Anleitung.pdf");
    const formFile = new FormData(form).get("document");
    expect(formFile).toBeInstanceOf(File);
    if (!(formFile instanceof File)) throw new Error("expected submitted file");
    expect({ name: formFile.name, size: formFile.size, type: formFile.type }).toEqual({
      name: initialFormFile.name,
      size: initialFormFile.size,
      type: initialFormFile.type
    });
  });

  it("mirrors required validation onto the visible trigger and focuses it", () => {
    const { container } = render(
      <AppFilePicker
        label="Document"
        file={null}
        required
        triggerLabel="Choose file"
        clearLabel="Remove file"
        emptyLabel="No file selected"
      />
    );
    const input = container.querySelector('input[type="file"]');
    if (!(input instanceof HTMLInputElement)) throw new Error("expected file input");
    const trigger = screen.getByRole("button", { name: "Document Choose file No file selected" });

    expect(trigger).toHaveAttribute("aria-required", "true");
    expect(trigger).toHaveAttribute("aria-invalid", "true");
    fireEvent.invalid(input);
    expect(trigger).toHaveFocus();
  });

  it("does not constrain an empty read-only or disabled required picker", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn((event: FormEvent) => event.preventDefault());
    const { container, rerender } = render(
      <form onSubmit={onSubmit}>
        <AppFilePicker
          label="Document"
          file={null}
          required
          readOnly
          triggerLabel="Choose file"
          clearLabel="Remove file"
          emptyLabel="No file selected"
        />
        <button type="submit">Save</button>
      </form>
    );
    const input = container.querySelector<HTMLInputElement>('input[type="file"]');
    if (!input) throw new Error("expected file input");
    const trigger = screen.getByRole("button", { name: "Document Choose file No file selected" });

    expect(input).not.toBeRequired();
    expect(input.checkValidity()).toBe(true);
    await user.click(screen.getByRole("button", { name: "Save" }));
    expect(onSubmit).toHaveBeenCalledOnce();
    expect(trigger).not.toHaveFocus();

    rerender(
      <form onSubmit={onSubmit}>
        <AppFilePicker
          label="Document"
          file={null}
          required
          disabled
          triggerLabel="Choose file"
          clearLabel="Remove file"
          emptyLabel="No file selected"
        />
      </form>
    );
    expect(input).not.toBeRequired();
    expect(input.checkValidity()).toBe(true);
  });
});
