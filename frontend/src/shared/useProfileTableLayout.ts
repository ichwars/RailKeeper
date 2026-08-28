import { useCallback, useEffect, useRef, useState } from "react";

import { api } from "./api";

type ProfileTableLayoutOptions<Layout> = {
  settingKey: string;
  defaultLayout: Layout;
  parse: (raw: string | undefined) => Layout;
  serialize: (layout: Layout) => string;
  legacyValue?: () => string | undefined;
  onLoadError: () => void;
  onSaveError: () => void;
};

export function useProfileTableLayout<Layout>({
  settingKey,
  defaultLayout,
  parse,
  serialize,
  legacyValue,
  onLoadError,
  onSaveError
}: ProfileTableLayoutOptions<Layout>) {
  const [layout, setLayout] = useState(defaultLayout);
  const [loading, setLoading] = useState(true);
  const layoutRef = useRef(layout);
  const saveQueue = useRef<Promise<void>>(Promise.resolve());
  const parseRef = useRef(parse);
  const serializeRef = useRef(serialize);
  const legacyValueRef = useRef(legacyValue);
  const onLoadErrorRef = useRef(onLoadError);
  const onSaveErrorRef = useRef(onSaveError);

  useEffect(() => {
    parseRef.current = parse;
    serializeRef.current = serialize;
    legacyValueRef.current = legacyValue;
    onLoadErrorRef.current = onLoadError;
    onSaveErrorRef.current = onSaveError;
  }, [legacyValue, onLoadError, onSaveError, parse, serialize]);

  const queueSave = useCallback((next: Layout) => {
    const value = serializeRef.current(next);
    saveQueue.current = saveQueue.current
      .catch(() => undefined)
      .then(() => api.updateProfileSettings({ [settingKey]: value }))
      .then(() => undefined)
      .catch(() => onSaveErrorRef.current());
  }, [settingKey]);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);

    api.profileSettings()
      .then(({ settings }) => {
        if (cancelled) return;
        const stored = settings[settingKey];
        const legacy = stored === undefined ? legacyValueRef.current?.() : undefined;
        const next = parseRef.current(stored ?? legacy);
        layoutRef.current = next;
        setLayout(next);
        if (stored === undefined && legacy !== undefined) queueSave(next);
      })
      .catch(() => {
        if (!cancelled) onLoadErrorRef.current();
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [queueSave, settingKey]);

  const update = useCallback((change: (current: Layout) => Layout, persist: boolean) => {
    const next = change(layoutRef.current);
    layoutRef.current = next;
    setLayout(next);
    if (persist) queueSave(next);
  }, [queueSave]);

  const preview = useCallback((change: (current: Layout) => Layout) => {
    update(change, false);
  }, [update]);

  const commit = useCallback((change: (current: Layout) => Layout) => {
    update(change, true);
  }, [update]);

  return { commit, layout, loading, preview };
}
