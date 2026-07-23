import { describe, expect, it } from "vitest";

import type { ArticleSearchImage, ArticleSearchResult, CreateVehicleRequest } from "../../shared/api";
import {
  articleFieldStatus,
  articleSelectionKey,
  articleValueForForm,
  booleanFromArticleValue,
  currentArticleValue,
  imageSelectionKey,
  sourceDisplayName,
  sourceShortLink
} from "./articleSearch";

const result: ArticleSearchResult = {
  source: "manufacturer",
  title: "BR 106",
  url: "https://www.esu.eu/produkte/br-106",
  snippet: "Diesellokomotive",
  score: 100,
  fields: {}
};
const image: ArticleSearchImage = {
  url: "https://www.esu.eu/images/br-106.jpg",
  title: "BR 106",
  source: result.url
};

describe("article search helpers", () => {
  it("normalizes common boolean values", () => {
    expect(booleanFromArticleValue(" Ja ")).toBe(true);
    expect(booleanFromArticleValue("vorhanden")).toBe(true);
    expect(booleanFromArticleValue("nein")).toBe(false);
    expect(articleValueForForm("digital", "digital")).toBe(true);
    expect(articleValueForForm("name", "BR 106")).toBe("BR 106");
  });

  it("builds stable field and image selection keys", () => {
    expect(articleSelectionKey(result, "name", 2)).toBe(`${result.url}-2::name`);
    expect(imageSelectionKey(result, image, 2)).toBe(`${result.url}-2::image::${image.url}`);
  });

  it("formats current values and field conflicts", () => {
    const form = { manufacturer: "ESU", name: "BR 106", gauge: "H0", digital: true } satisfies CreateVehicleRequest;
    expect(currentArticleValue(form, "digital")).toBe("Ja");
    expect(currentArticleValue(form, "manufacturer")).toBe("ESU");
    expect(articleFieldStatus("", "ESU")).toBe("empty");
    expect(articleFieldStatus("esu", "ESU")).toBe("same");
    expect(articleFieldStatus("Roco", "ESU")).toBe("conflict");
  });

  it("uses safe source labels and compact links", () => {
    expect(sourceDisplayName("https://www.esu.eu/produkte/1")).toBe("Esu");
    expect(sourceDisplayName("not a url")).toBe("Quelle");
    expect(sourceShortLink("https://www.esu.eu/produkte/1")).toBe("esu.eu/produkte/1");
  });
});
