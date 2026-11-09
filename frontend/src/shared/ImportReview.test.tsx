import { fireEvent, render, screen, within } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { ArticleSearchDialog } from "./articleSearch/ArticleSearchDialog";
import { articleSelectionKey } from "./articleSearch/articleSearchModel";
import { languageSettingKey } from "./i18n";
import { ImportReviewStatus } from "./ImportReview";
import type { ArticleSearchResult } from "./api";

describe("shared import review", () => {
  beforeEach(() => localStorage.clear());
  it("labels image choices and retains image-selection events", () => {
    const image = { url: "https://example.test/image.png", title: "Seitenansicht", source: "manufacturer" };
    const result: ArticleSearchResult = { source: "manufacturer", title: "Reference", score: 100,
      url: "https://example.test/model", snippet: "", fields: {}, images: [image] };
    const toggle = vi.fn();
    render(<ArticleSearchDialog loading={false} response={{ query: "Reference", results: [result] }} error=""
      fieldGroups={[]} currentValue={() => ""} selectedFields={{}} selectedImages={{}}
      onApply={vi.fn()} onClose={vi.fn()} onToggleField={vi.fn()} onToggleImage={toggle} />);
    fireEvent.click(screen.getByRole("checkbox", { name: "Bild 1 auswählen: Seitenansicht" }));
    expect(toggle).toHaveBeenCalledExactlyOnceWith(result, 0, image, true);
    expect(document.querySelector(".article-image-option > .app-choice-control + img"))
      .toHaveAttribute("src", image.url);
  });

  it("honors an explicit view language independently of the global preference", () => {
    render(<ImportReviewStatus state="changed" language="en" />);
    expect(screen.getByText("Different")).toBeInTheDocument();
  });
  it("shows actual article source, target, selected scope and protected/unclear fields without writing", () => {
    localStorage.setItem(languageSettingKey, "en");
    const result: ArticleSearchResult = { source: "manufacturer", title: "Reference", score: 100,
      url: "https://example.test/catalog/variant", snippet: "", conflicts: ["weightG"], fields: {
        name: { label: "Name", value: "New name", confidence: 100 },
        manufacturer: { label: "Manufacturer", value: "Piko", confidence: 100 },
        gauge: { label: "Gauge", value: "TT", confidence: 100 },
        description: { label: "Description", value: "Remote", confidence: 100 },
        weightG: { label: "Weight", value: "100", confidence: 30 }
      } };
    const onApply = vi.fn(), onToggleField = vi.fn();
    const current: Record<string, string> = { name: "Original", gauge: "TT", description: "Protected" };
    render(<ArticleSearchDialog loading={false} response={{ query: "Reference", results: [result] }} error=""
      reviewTarget="Accessory draft RK-A-1" fieldGroups={[{ key: "core", label: "Core", fields: Object.entries(result.fields)
        .map(([key, value]) => ({ key, label: value.label })) }]}
      currentValue={key => current[key] || ""}
      selectedFields={Object.fromEntries(["name", "manufacturer", "description"].map(key => [articleSelectionKey(result, key), true]))}
      selectedImages={{}} canSelectField={key => key !== "description"} onApply={onApply} onClose={vi.fn()}
      onToggleField={onToggleField} onToggleImage={vi.fn()} />);
    const scope = screen.getByRole("region", { name: "Source, target and effect" });
    expect(within(scope).getByText(result.url)).toBeInTheDocument();
    expect(within(scope).getByText("Accessory draft RK-A-1")).toBeInTheDocument();
    expect(within(scope).getByText(/2 of 5 selected/)).toBeInTheDocument();
    for (const [label, status] of [["Name", "Different"], ["Manufacturer", "New"], ["Gauge", "Identical"],
      ["Description", "Excluded"], ["Weight", "Unclear"]]) {
      const row = screen.getByRole("row", { name: new RegExp(label) });
      expect(within(row).getByText(status)).toBeInTheDocument();
    }
    const protectedRow = screen.getByRole("row", { name: /Description/ });
    expect(within(protectedRow).getByRole("checkbox")).toBeDisabled();
    fireEvent.click(within(screen.getByRole("row", { name: /Manufacturer/ })).getByRole("checkbox"));
    expect(onToggleField).toHaveBeenCalledExactlyOnceWith(result, 0, "manufacturer", false);
    expect(onApply).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: "Apply selected fields" }));
    expect(onApply).toHaveBeenCalledExactlyOnceWith(result);
  });
});
