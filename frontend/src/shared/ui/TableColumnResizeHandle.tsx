import { useRef, type KeyboardEvent, type PointerEvent } from "react";

type TableColumnResizeHandleProps = {
  label: string;
  width: number;
  minWidth: number;
  maxWidth: number;
  defaultWidth: number;
  onPreview: (width: number) => void;
  onCommit: (width: number) => void;
};

type DragState = {
  pointerID: number;
  startX: number;
  startWidth: number;
  currentWidth: number;
};

function boundedWidth(width: number, minWidth: number, maxWidth: number) {
  return Math.min(maxWidth, Math.max(minWidth, Math.round(width)));
}

export function TableColumnResizeHandle({
  label,
  width,
  minWidth,
  maxWidth,
  defaultWidth,
  onPreview,
  onCommit
}: TableColumnResizeHandleProps) {
  const dragRef = useRef<DragState | null>(null);

  const startDrag = (event: PointerEvent<HTMLSpanElement>) => {
    if (event.button !== 0) return;
    event.preventDefault();
    event.stopPropagation();
    event.currentTarget.setPointerCapture?.(event.pointerId);
    dragRef.current = {
      pointerID: event.pointerId,
      startX: event.clientX,
      startWidth: width,
      currentWidth: width
    };
  };

  const moveDrag = (event: PointerEvent<HTMLSpanElement>) => {
    const drag = dragRef.current;
    if (!drag || drag.pointerID !== event.pointerId) return;
    event.preventDefault();
    event.stopPropagation();
    const next = boundedWidth(
      drag.startWidth + event.clientX - drag.startX,
      minWidth,
      maxWidth
    );
    drag.currentWidth = next;
    onPreview(next);
  };

  const finishDrag = (event: PointerEvent<HTMLSpanElement>) => {
    const drag = dragRef.current;
    if (!drag || drag.pointerID !== event.pointerId) return;
    event.preventDefault();
    event.stopPropagation();
    if (event.currentTarget.hasPointerCapture?.(event.pointerId)) {
      event.currentTarget.releasePointerCapture?.(event.pointerId);
    }
    dragRef.current = null;
    onCommit(drag.currentWidth);
  };

  const resizeWithKeyboard = (event: KeyboardEvent<HTMLSpanElement>) => {
    let next: number | undefined;
    const step = event.shiftKey ? 32 : 8;
    if (event.key === "ArrowLeft") next = width - step;
    if (event.key === "ArrowRight") next = width + step;
    if (event.key === "Home") next = minWidth;
    if (event.key === "End") next = maxWidth;
    if (next === undefined) return;
    event.preventDefault();
    event.stopPropagation();
    onCommit(boundedWidth(next, minWidth, maxWidth));
  };

  return (
    <span
      className="table-column-resize-handle"
      role="separator"
      aria-label={label}
      aria-orientation="vertical"
      aria-valuemin={minWidth}
      aria-valuemax={maxWidth}
      aria-valuenow={width}
      tabIndex={0}
      title={label}
      onClick={(event) => {
        event.preventDefault();
        event.stopPropagation();
      }}
      onDoubleClick={(event) => {
        event.preventDefault();
        event.stopPropagation();
        onCommit(defaultWidth);
      }}
      onKeyDown={resizeWithKeyboard}
      onPointerDown={startDrag}
      onPointerMove={moveDrag}
      onPointerUp={finishDrag}
      onPointerCancel={finishDrag}
    />
  );
}
