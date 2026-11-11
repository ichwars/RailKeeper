import { flushSync } from "react-dom";
import { useLayoutEffect, useMemo, useRef, useState, type KeyboardEvent, type ReactNode } from "react";

const estimatedRowHeight = 54;
const overscan = 8;

function firstOffsetAfter(offsets: number[], position: number) {
  let start = 0;
  let end = offsets.length;
  while (start < end) {
    const middle = Math.floor((start + end) / 2);
    if (offsets[middle] <= position) start = middle + 1;
    else end = middle;
  }
  return start;
}

function rowTabStops(row: HTMLTableRowElement) {
  return Array.from(row.querySelectorAll<HTMLElement>("a[href],button,input,select,textarea,[tabindex]"))
    .filter(element => element.tabIndex >= 0 && !element.matches(":disabled") &&
      element.getClientRects().length > 0 && getComputedStyle(element).visibility === "visible");
}

export function VehicleInventoryWindow({ rowKeys, columnCount, renderRow }: {
  rowKeys: readonly string[];
  columnCount: number;
  renderRow: (index: number) => ReactNode;
}) {
  const bodyRef = useRef<HTMLTableSectionElement>(null);
  const [heights, setHeights] = useState<ReadonlyMap<string, number>>(() => new Map());
  const [range, setRange] = useState({ start: 0, end: overscan });
  const [focusedKey, setFocusedKey] = useState<string>();
  const [requestedIndex, setRequestedIndex] = useState<number>();
  const indexes = useMemo(() => new Map(rowKeys.map((key, index) => [key, index])), [rowKeys]);
  const offsets = useMemo(() => {
    const positions = [0];
    let offset = 0;
    for (const key of rowKeys) {
      offset += heights.get(key) ?? estimatedRowHeight;
      positions.push(offset);
    }
    return positions;
  }, [heights, rowKeys]);

  useLayoutEffect(() => {
    const body = bodyRef.current;
    if (!body) return;
    const scroller = body.closest("main");
    let frame = 0;
    const update = () => {
      const bodyTop = body.getBoundingClientRect().top;
      const viewport = scroller?.getBoundingClientRect();
      const top = Math.max(0, viewport?.top ?? 0) - bodyTop;
      const bottom = Math.min(window.innerHeight, viewport?.bottom ?? window.innerHeight) - bodyTop;
      const start = Math.max(0, firstOffsetAfter(offsets, top) - 1 - overscan);
      const end = Math.min(rowKeys.length, firstOffsetAfter(offsets, bottom) + overscan);
      setRange(previous => previous.start === start && previous.end === end ? previous : { start, end });
    };
    const schedule = () => {
      window.cancelAnimationFrame(frame);
      frame = window.requestAnimationFrame(update);
    };
    update();
    const observer = new ResizeObserver(schedule);
    const panelHead = body.closest(".inventory-panel")?.querySelector(".inventory-list-head");
    if (scroller) observer.observe(scroller);
    if (panelHead) observer.observe(panelHead);
    window.addEventListener("scroll", schedule, true);
    window.addEventListener("resize", schedule);
    return () => {
      observer.disconnect();
      window.removeEventListener("scroll", schedule, true);
      window.removeEventListener("resize", schedule);
      window.cancelAnimationFrame(frame);
    };
  }, [offsets, rowKeys.length]);

  const focusedIndex = focusedKey === undefined ? undefined : indexes.get(focusedKey);
  useLayoutEffect(() => {
    const body = bodyRef.current;
    if (!body) return;
    const observer = new ResizeObserver(entries => {
      const measured = entries.flatMap(({ target }) => {
        const key = (target as HTMLTableRowElement).dataset.inventoryRow;
        const height = target.getBoundingClientRect().height;
        return key && height > 0 ? [[key, height] as const] : [];
      });
      setHeights(previous => measured.some(([key, height]) => previous.get(key) !== height)
        ? new Map([...previous, ...measured]) : previous);
    });
    for (const row of body.querySelectorAll("tr[data-inventory-row]")) observer.observe(row);
    return () => observer.disconnect();
  }, [rowKeys, range, focusedIndex, requestedIndex]);

  const mountedIndexes = new Set<number>();
  for (let index = range.start; index < Math.min(range.end, rowKeys.length); index++) mountedIndexes.add(index);
  // A pointer scroll must keep the active control mounted even when it is far outside the window.
  if (focusedIndex !== undefined) mountedIndexes.add(focusedIndex);
  if (requestedIndex !== undefined && requestedIndex < rowKeys.length) mountedIndexes.add(requestedIndex);

  const children: ReactNode[] = [];
  const spacer = (start: number, end: number) => (
    <tr key={`space-${start}-${end}`} aria-hidden="true">
      <td colSpan={columnCount} style={{ height: offsets[end] - offsets[start], padding: 0, border: 0 }} />
    </tr>
  );
  let next = 0;
  for (const index of Array.from(mountedIndexes).sort((left, right) => left - right)) {
    if (next < index) children.push(spacer(next, index));
    children.push(renderRow(index));
    next = index + 1;
  }
  if (next < rowKeys.length) children.push(spacer(next, rowKeys.length));

  function handleTab(event: KeyboardEvent<HTMLTableSectionElement>) {
    if (event.defaultPrevented || event.key !== "Tab" || event.altKey || event.ctrlKey || event.metaKey) return;
    const row = (event.target as HTMLElement).closest<HTMLTableRowElement>("tr[data-inventory-row]");
    if (!row || !event.currentTarget.contains(row)) return;
    const index = indexes.get(row.dataset.inventoryRow ?? "");
    if (index === undefined) return;
    const stops = rowTabStops(row);
    if (event.target !== (event.shiftKey ? stops[0] : stops.at(-1))) return;
    const target = index + (event.shiftKey ? -1 : 1);
    if (target < 0 || target >= rowKeys.length || mountedIndexes.has(target)) return;

    // Tab can outrun the scroll window. Mount the adjacent logical row before moving focus.
    flushSync(() => setRequestedIndex(target));
    const adjacent = bodyRef.current?.querySelector<HTMLTableRowElement>(`tr[aria-rowindex="${target + 2}"]`);
    const destinations = adjacent ? rowTabStops(adjacent) : [];
    const destination = event.shiftKey ? destinations.at(-1) : destinations[0];
    if (destination) {
      event.preventDefault();
      destination.focus();
    }
    setRequestedIndex(undefined);
  }

  return (
    <tbody
      ref={bodyRef}
      style={{ overflowAnchor: "none" }}
      onFocus={event => {
        const row = (event.target as HTMLElement).closest<HTMLTableRowElement>("tr[data-inventory-row]");
        if (row && event.currentTarget.contains(row)) setFocusedKey(row.dataset.inventoryRow);
      }}
      onBlur={event => {
        if (!(event.relatedTarget instanceof Node) || !event.currentTarget.contains(event.relatedTarget)) {
          setFocusedKey(undefined);
        }
      }}
      onKeyDown={handleTab}
    >
      {children}
    </tbody>
  );
}
