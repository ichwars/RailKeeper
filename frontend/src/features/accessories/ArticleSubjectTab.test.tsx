import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
import { describe, expect, it, vi } from "vitest";

import { api, type MasterDataEntry } from "../../shared/api";
import { ArticleSubjectTab } from "./ArticleSubjectTab";
import { emptyArticleEditorForm, type ArticleEditorForm } from "./articleEditorModel";

function customField(key: string, label: string, kind: string, metadata: Record<string, unknown> = {}): MasterDataEntry {
  return {
    id: `field-${key}`, type: "accessory_custom_field", key, label, active: true, sortOrder: 0,
    metadata: { kind, ...metadata }, createdAt: "2026-08-08T08:00:00Z", updatedAt: "2026-08-08T08:00:00Z"
  };
}

describe("ArticleSubjectTab", () => {
  it("renders only the selected type definitions through one subject component", () => {
    const { rerender } = render(<ArticleSubjectTab form={{ ...emptyArticleEditorForm(), articleType: "track" }}
      disabled={false} onChange={vi.fn()} />);

    expect(screen.getByRole("textbox", { name: "Gleissystem" })).toBeInTheDocument();
    expect(screen.getByRole("spinbutton", { name: /Länge/ })).toBeInTheDocument();
    expect(screen.queryByRole("textbox", { name: "Vorbild" })).not.toBeInTheDocument();

    rerender(<ArticleSubjectTab form={{ ...emptyArticleEditorForm(), articleType: "decoder" }}
      disabled={false} onChange={vi.fn()} />);
    expect(screen.getByRole("textbox", { name: "Firmware" })).toBeInTheDocument();
    expect(screen.queryByRole("textbox", { name: "Gleissystem" })).not.toBeInTheDocument();
  });

  it("writes typed standard values and preserves incomplete number drafts", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    function Harness() {
      const [form, setForm] = useState<ArticleEditorForm>({
        ...emptyArticleEditorForm(),
        articleType: "track"
      });
      return <ArticleSubjectTab form={form} disabled={false} onChange={(patch) => {
        onChange(patch);
        setForm((current) => ({ ...current, ...patch }));
      }} />;
    }
    render(<Harness />);

    await user.type(screen.getByRole("textbox", { name: "Gleissystem" }), "Tillig TT Modellgleis");
    expect(onChange).toHaveBeenLastCalledWith(expect.objectContaining({
      attributes: [{ key: "trackSystem", kind: "text", textValue: "Tillig TT Modellgleis" }]
    }));

    fireEvent.change(screen.getByRole("spinbutton", { name: /Länge/ }), { target: { value: "16.5" } });
    expect(onChange).toHaveBeenCalledWith(expect.objectContaining({
      attributeNumberDrafts: { lengthMm: "16.5" }
    }));
  });

  it("renders all six configured custom kinds for other and excludes inactive fields", () => {
    render(<ArticleSubjectTab form={{ ...emptyArticleEditorForm(), articleType: "other" }} disabled={false}
      onChange={vi.fn()} customFieldEntries={[
        customField("note", "Notiz", "text"),
        customField("weight", "Gewicht", "number", { unit: "g" }),
        customField("weatherproof", "Wetterfest", "boolean"),
        customField("released", "Erschienen", "date"),
        customField("color", "Farbe", "single_select", { options: ["Rot", "Grün"] }),
        customField("uses", "Einsatz", "multi_select", { options: ["Innen", "Außen"] }),
        { ...customField("hidden", "Versteckt", "text"), active: false }
      ]} />);

    expect(screen.getByRole("textbox", { name: "Notiz" })).toBeInTheDocument();
    expect(screen.getByRole("spinbutton", { name: /Gewicht.*g/ })).toBeInTheDocument();
    expect(screen.getByRole("checkbox", { name: "Wetterfest" })).toBeInTheDocument();
    expect(screen.getByRole("textbox", { name: "Erschienen" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Farbe" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Einsatz/ })).toBeInTheDocument();
    expect(screen.queryByText("Versteckt")).not.toBeInTheDocument();
  });

  it("loads active custom fields and writes each configured kind as a typed value", async () => {
    const user = userEvent.setup();
    const entries = [
      customField("note", "Notiz", "text"),
      customField("weight", "Gewicht", "number", { unit: "g" }),
      customField("weatherproof", "Wetterfest", "boolean"),
      customField("released", "Erschienen", "date"),
      customField("color", "Farbe", "single_select", { options: ["Rot", "Grün"] }),
      customField("uses", "Einsatz", "multi_select", { options: ["Innen", "Außen"] })
    ];
    vi.spyOn(api, "masterData").mockResolvedValue(entries);
    const patches: Array<Partial<ArticleEditorForm>> = [];
    function Harness() {
      const [form, setForm] = useState<ArticleEditorForm>({
        ...emptyArticleEditorForm(),
        articleType: "other"
      });
      return <ArticleSubjectTab form={form} disabled={false} onChange={(patch) => {
        patches.push(patch);
        setForm((current) => ({ ...current, ...patch }));
      }} />;
    }
    render(<Harness />);

    expect(await screen.findByRole("textbox", { name: "Notiz" })).toBeInTheDocument();
    expect(api.masterData).toHaveBeenCalledWith("accessory_custom_field", true);
    await user.type(screen.getByRole("textbox", { name: "Notiz" }), "Werkstatt");
    fireEvent.change(screen.getByRole("spinbutton", { name: /Gewicht/ }), { target: { value: "12.5" } });
    await user.click(screen.getByRole("checkbox", { name: "Wetterfest" }));
    fireEvent.change(screen.getByRole("textbox", { name: "Erschienen" }), {
      target: { value: "01.08.2026" }
    });
    fireEvent.blur(screen.getByRole("textbox", { name: "Erschienen" }));
    await user.click(screen.getByRole("button", { name: "Farbe" }));
    await user.click(screen.getByRole("option", { name: "Rot" }));
    await user.click(screen.getByRole("button", { name: /Einsatz/ }));
    await user.click(screen.getByRole("option", { name: "Innen" }));

    await waitFor(() => expect(patches.some((patch) => patch.attributes?.some((attribute) =>
      attribute.key === "note" && attribute.kind === "text" && attribute.textValue === "Werkstatt"))).toBe(true));
    expect(patches.some((patch) => patch.attributes?.some((attribute) =>
      attribute.key === "weight" && attribute.kind === "number" && attribute.numberValue === 12.5 &&
      attribute.unit === "g"))).toBe(true);
    expect(patches.some((patch) => patch.attributes?.some((attribute) =>
      attribute.key === "weatherproof" && attribute.kind === "boolean" && attribute.booleanValue))).toBe(true);
    expect(patches.some((patch) => patch.attributes?.some((attribute) =>
      attribute.key === "released" && attribute.kind === "date" && attribute.dateValue === "2026-08-01"))).toBe(true);
    expect(patches.some((patch) => patch.attributes?.some((attribute) =>
      attribute.key === "color" && attribute.kind === "single_select" && attribute.optionValues[0] === "Rot"))).toBe(true);
    expect(patches.some((patch) => patch.attributes?.some((attribute) =>
      attribute.key === "uses" && attribute.kind === "multi_select" && attribute.optionValues.includes("Innen"))))
      .toBe(true);
  });

  it("disables every subject control in read-only mode", () => {
    render(<ArticleSubjectTab form={{ ...emptyArticleEditorForm(), articleType: "track" }}
      disabled onChange={vi.fn()} />);
    expect(screen.getAllByRole("textbox").every((control) => control.hasAttribute("disabled"))).toBe(true);
    expect(screen.getAllByRole("spinbutton").every((control) => control.hasAttribute("disabled"))).toBe(true);
    expect(screen.getAllByRole("checkbox").every((control) => control.hasAttribute("disabled"))).toBe(true);
    expect(screen.getAllByRole("button").every((control) => control.hasAttribute("disabled"))).toBe(true);
  });

  it("keeps a long configured label in the responsive subject-grid contract", () => {
    const longLabel = "Sehr lange kontrollierte Zusatzfeldbezeichnung für schmale Artikeldialoge";
    const { container } = render(<ArticleSubjectTab
      form={{ ...emptyArticleEditorForm(), articleType: "other" }}
      disabled={false}
      customFieldEntries={[customField("long_label", longLabel, "text")]}
      onChange={vi.fn()}
    />);

    expect(screen.getByRole("textbox", { name: longLabel })).toBeInTheDocument();
    expect(container.querySelector(".article-subject-grid")).toHaveClass("article-editor-grid");
    expect(screen.getByText(longLabel)).toHaveTextContent(longLabel);
  });

  it("associates a single-select subject error with its app-owned trigger", () => {
    render(<ArticleSubjectTab
      form={{ ...emptyArticleEditorForm(), articleType: "track", attributes: [
        { key: "direction", kind: "single_select", optionValues: ["left"] }
      ] }}
      disabled={false}
      subjectFieldErrors={{ direction: "Auswahl ist ungültig" }}
      onChange={vi.fn()}
    />);

    const trigger = screen.getByRole("button", { name: "Richtung" });
    expect(trigger).toHaveAttribute("id", "article-subject-direction");
    expect(trigger).toHaveAttribute("aria-describedby", "article-subject-direction-error");
    expect(screen.getByText("Auswahl ist ungültig")).toHaveAttribute("id", "article-subject-direction-error");
  });
});
