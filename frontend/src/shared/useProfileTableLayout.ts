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
  const loadingRef = useRef(true);
  const pendingChangesRef = useRef<Array<{
    change: (current: Layout) => Layout;
    persist: boolean;
  }>>([]);
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
    loadingRef.current = true;
    pendingChangesRef.current = [];
    setLoading(true);

    api.profileSettings()
      .then(({ settings }) => {
        if (cancelled) return;
        const stored = settings[settingKey];
        const legacy = stored === undefined ? legacyValueRef.current?.() : undefined;
        const pendingChanges = pendingChangesRef.current;
        const next = pendingChanges.reduce(
          (current, pending) => pending.change(current),
          parseRef.current(stored ?? legacy)
        );
        const shouldPersist = pendingChanges.some((pending) => pending.persist) ||
          (stored === undefined && legacy !== undefined);
        pendingChangesRef.current = [];
        loadingRef.current = false;
        layoutRef.current = next;
        setLayout(next);
        if (shouldPersist) queueSave(next);
        setLoading(false);
      })
      .catch(() => {
        if (cancelled) return;
        const shouldPersist = pendingChangesRef.current.some((pending) => pending.persist);
        pendingChangesRef.current = [];
        loadingRef.current = false;
        onLoadErrorRef.current();
        if (shouldPersist) queueSave(layoutRef.current);
        setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [queueSave, settingKey]);

  const update = useCallback((change: (current: Layout) => Layout, persist: boolean) => {
    if (loadingRef.current) pendingChangesRef.current.push({ change, persist });
    const next = change(layoutRef.current);
    layoutRef.current = next;
    setLayout(next);
    if (persist && !loadingRef.current) queueSave(next);
  }, [queueSave]);

  const preview = useCallback((change: (current: Layout) => Layout) => {
    update(change, false);
  }, [update]);

  const commit = useCallback((change: (current: Layout) => Layout) => {
    update(change, true);
  }, [update]);

  return { commit, layout, loading, preview };
}
