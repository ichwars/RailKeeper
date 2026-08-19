import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import type { ArticleSearchResult } from "../../shared/api";
import { VehicleCreateArticleResults } from "./VehicleCreateArticleResults";

function result(title: string, fieldCount: number, score: number): ArticleSearchResult {
  return {
    source: "test",
    title,
    url: `https://example.test/${encodeURIComponent(title)}`,
    snippet: "",
    score,
    fields: Object.fromEntries(Array.from({ length: fieldCount }, (_, index) => [
      `field-${index}`,
      { label: `Field ${index}`, value: String(index), confidence: 1 }
    ])),
    conflicts: []
  };
}

describe("VehicleCreateArticleResults", () => {
  it("sorts by field count, then score, while preserving stable ties and the response", () => {
    const results = [
      result("Few high score", 1, 900),
      result("Many low score", 3, 100),
      result("Many high score first", 3, 200),
      result("Many high score second", 3, 200)
    ];
    const originalTitles = results.map((entry) => entry.title);

    render(<VehicleCreateArticleResults response={{ query: "test", results }}
      onSelect={vi.fn()} onRevise={vi.fn()} />);

    expect(screen.getAllByRole("heading", { level: 4 }).map((heading) => heading.textContent)).toEqual([
      "Many high score first",
      "Many high score second",
      "Many low score",
      "Few high score"
    ]);
    expect(results.map((entry) => entry.title)).toEqual(originalTitles);
  });

  it("selects the original response index after visual sorting", async () => {
    const user = userEvent.setup();
    const onSelect = vi.fn();
    const results = [result("Few", 1, 900), result("Most", 4, 100)];

    render(<VehicleCreateArticleResults response={{ query: "test", results }}
      onSelect={onSelect} onRevise={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: /Most auswählen/ }));
    expect(onSelect).toHaveBeenCalledWith(1);
  });
});
