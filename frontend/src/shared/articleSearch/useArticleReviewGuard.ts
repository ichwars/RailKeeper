import { useEffect, useLayoutEffect, useRef } from "react";
import type { ArticleSearchResponse, ArticleSearchResult } from "../api";

export function reviewedResultIndex(response: ArticleSearchResponse | null, result: ArticleSearchResult) {
  if (!response) return -1;
  const exact = response.results.indexOf(result);
  if (exact >= 0) return exact;
  const serialized = JSON.stringify(result);
  return response.results.findIndex(entry => JSON.stringify(entry) === serialized);
}

export function useArticleReviewGuard(scope: string, onScopeChange: () => void) {
  const reset = useRef(onScopeChange); reset.current = onScopeChange;
  const previousScope = useRef(scope);
  const current = useRef({ scope, version: 0, live: true, consumed: false,
    response: null as ArticleSearchResponse | null, expected: new Map<string, string>() });
  if (current.current.scope !== scope) {
    current.current.scope = scope; current.current.version++; current.current.response = null;
    current.current.expected.clear(); current.current.consumed = false;
  }
  useLayoutEffect(() => {
    if (previousScope.current !== scope) { previousScope.current = scope; reset.current(); }
  }, [scope]);
  useEffect(() => {
    const review = current.current;
    review.live = true;
    return () => { review.live = false; review.version++; };
  }, []);
  const available = () => current.current.live && current.current.scope === scope;
  const active = (version: number | null) => version !== null && available() && current.current.version === version;
  const reviewing = (response: ArticleSearchResponse | null) => current.current.live && current.current.scope === scope
    && Boolean(response) && current.current.response === response && !current.current.consumed;
  return {
    begin: () => { if (!available()) return null;
      current.current.response = null; current.current.expected.clear(); current.current.consumed = false;
      return ++current.current.version; },
    available,
    active,
    complete: (version: number | null, response: ArticleSearchResponse | null) => {
      if (!active(version)) return false;
      current.current.response = response; return true;
    },
    reviewing,
    remember: (key: string, value: string) => { current.current.expected.set(key, value); },
    changed: (values: Array<[string, string]>) => values.filter(([key, value]) => current.current.expected.get(key) !== value).map(([key]) => key),
    consume: (response: ArticleSearchResponse | null) => {
      if (!reviewing(response)) return false;
      current.current.consumed = true; return true;
    },
    invalidate: () => { if (available()) { current.current.version++; current.current.response = null; } }
  };
}
