import { fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { BugReportScreenshotField } from "./BugReportScreenshotField";

describe("BugReportScreenshotField", () => {
  beforeEach(() => {
    vi.stubGlobal("URL", {
      ...URL,
      createObjectURL: vi.fn(() => "blob:preview"),
      revokeObjectURL: vi.fn()
    });
  });

  it("accepts a bounded PNG and revokes its preview when removed", async () => {
    const user = userEvent.setup();
    const onFileChange = vi.fn();
    const file = new File(["png"], "screen.png", { type: "image/png" });
    const { rerender } = render(<BugReportScreenshotField
      file={null} onFileChange={onFileChange} consent={false} onConsentChange={vi.fn()}
      maxBytes={5 * 1024 * 1024} acceptedTypes={["image/png", "image/jpeg", "image/webp"]}
    />);
    const input = screen.getByLabelText(/Screenshot/i, { selector: "input" });
    expect(input).toHaveAttribute("capture", "environment");
    await user.upload(input, file);
    expect(onFileChange).toHaveBeenCalledWith(file);

    rerender(<BugReportScreenshotField
      file={file} onFileChange={onFileChange} consent={false} onConsentChange={vi.fn()}
      maxBytes={5 * 1024 * 1024} acceptedTypes={["image/png", "image/jpeg", "image/webp"]}
    />);
    expect(screen.getByRole("img", { name: /Vorschau|preview/i })).toHaveAttribute("src", "blob:preview");
    expect(screen.getByText(/3 B/)).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: /entfernen|remove/i }));
    expect(onFileChange).toHaveBeenLastCalledWith(null);
  });

  it("accepts drop and focused paste but rejects unsupported or oversized files", () => {
    const onFileChange = vi.fn();
    render(<BugReportScreenshotField
      file={null} onFileChange={onFileChange} consent={false} onConsentChange={vi.fn()}
      maxBytes={4} acceptedTypes={["image/png"]}
    />);
    const field = screen.getByRole("group", { name: /Screenshot/i });
    const png = new File(["png"], "drop.png", { type: "image/png" });
    fireEvent.drop(field, { dataTransfer: { files: [png] } });
    expect(onFileChange).toHaveBeenCalledWith(png);

    const jpeg = new File(["jpeg"], "paste.jpg", { type: "image/jpeg" });
    fireEvent.paste(field, { clipboardData: { files: [jpeg] } });
    expect(screen.getByRole("alert")).toHaveTextContent(/Dateityp|file type/i);

    const large = new File(["12345"], "large.png", { type: "image/png" });
    fireEvent.drop(field, { dataTransfer: { files: [large] } });
    expect(screen.getByRole("alert")).toHaveTextContent(/groß|large/i);
  });

  it("requires explicit public-upload consent when a file is selected", async () => {
    const user = userEvent.setup();
    const onConsentChange = vi.fn();
    render(<BugReportScreenshotField
      file={new File(["png"], "screen.png", { type: "image/png" })}
      onFileChange={vi.fn()} consent={false} onConsentChange={onConsentChange}
      maxBytes={5 * 1024 * 1024} acceptedTypes={["image/png"]}
      error="Zustimmung erforderlich"
    />);
    expect(screen.getByRole("alert")).toHaveTextContent("Zustimmung erforderlich");
    await user.click(screen.getByRole("checkbox"));
    expect(onConsentChange).toHaveBeenCalledWith(true);
  });
});
