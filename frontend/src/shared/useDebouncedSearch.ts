import { useEffect, useState } from "react";

export const searchDebounceMs = 300;

export function useDebouncedSearch(query: string) {
  const [value, setValue] = useState(query);
  useEffect(() => {
    if (query === value) return;
    const timer = window.setTimeout(() => setValue(query), searchDebounceMs);
    return () => window.clearTimeout(timer);
  }, [query, value]);
  return value;
}
