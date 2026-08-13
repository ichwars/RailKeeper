export const articleSearchSettingKey = "railkeeper.articleSearchEnabled";
export const articleSearchSourcesSettingKey = "railkeeper.articleSearchSources";
export const articleSearchSourceIds = ["web", "manufacturer", "catalogs", "dealers", "wiki"];
export const defaultArticleSearchSources = ["manufacturer", "catalogs", "dealers", "web"];

const legacyArticleSearchSources = ["web", "manufacturer", "dealers", "wiki"];
const previousArticleSearchSources = ["manufacturer", "dealers", "web"];

export function articleSearchEnabled() {
  return window.localStorage.getItem(articleSearchSettingKey) !== "false";
}

function isLegacyArticleSearchDefault(sources: string[]) {
  return (
    sources.length === legacyArticleSearchSources.length &&
    legacyArticleSearchSources.every((source) => sources.includes(source))
  ) || (
    sources.length === previousArticleSearchSources.length &&
    previousArticleSearchSources.every((source) => sources.includes(source))
  );
}

export function articleSearchSources() {
  try {
    const stored = JSON.parse(window.localStorage.getItem(articleSearchSourcesSettingKey) || "[]") as string[];
    const allowed = new Set(articleSearchSourceIds);
    const sources = stored.filter((source) => allowed.has(source));
    if (isLegacyArticleSearchDefault(sources)) {
      window.localStorage.setItem(articleSearchSourcesSettingKey, JSON.stringify(defaultArticleSearchSources));
      return defaultArticleSearchSources;
    }
    return sources.length > 0 ? sources : defaultArticleSearchSources;
  } catch {
    return defaultArticleSearchSources;
  }
}

export const readArticleSearchSources = articleSearchSources;
