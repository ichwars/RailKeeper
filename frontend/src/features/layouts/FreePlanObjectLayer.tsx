import { KeyboardEvent } from "react";

import type { PlanFreeObject } from "../../shared/api";

export function FreePlanObjectLayer({ objects, selectedID, onSelect }: {
  objects: PlanFreeObject[];
  selectedID: string | null;
  onSelect: (object: PlanFreeObject) => void;
}) {
  const selectWithKeyboard = (event: KeyboardEvent<SVGGElement>, object: PlanFreeObject) => {
    if (event.key !== "Enter" && event.key !== " ") return;
    event.preventDefault();
    onSelect(object);
  };
  return <g className="free-plan-objects">
    {objects.map((object) => {
      const selected = object.id === selectedID;
      return <g key={object.id} role="button" tabIndex={0} aria-label={object.name}
        aria-pressed={selected} transform={`translate(${object.positionXMm} ${object.positionYMm}) rotate(${object.rotationDegrees})`}
        className={`free-plan-object free-plan-object-category-${object.category}${selected ? " is-selected" : ""}`}
        onClick={() => onSelect(object)} onKeyDown={(event) => selectWithKeyboard(event, object)}>
        {renderShape(object)}
        {selected ? renderSelection(object) : null}
      </g>;
    })}
  </g>;
}

function renderShape(object: PlanFreeObject) {
  const shape = object.shape;
  switch (shape.kind) {
    case "rectangle":
      return <rect className="free-plan-object-shape" width={shape.widthMm} height={shape.heightMm} />;
    case "ellipse":
      return <ellipse className="free-plan-object-shape" cx={(shape.widthMm ?? 0) / 2}
        cy={(shape.heightMm ?? 0) / 2} rx={(shape.widthMm ?? 0) / 2} ry={(shape.heightMm ?? 0) / 2} />;
    case "line":
      return <line className="free-plan-object-shape" x1={0} y1={0} x2={shape.endXMm} y2={shape.endYMm} />;
    case "label":
      return <text className="free-plan-object-shape" fontSize={shape.fontSizeMm}>{shape.text}</text>;
  }
}

function renderSelection(object: PlanFreeObject) {
  const shape = object.shape;
  switch (shape.kind) {
    case "rectangle":
      return <rect className="free-plan-object-selection" x={-3} y={-3}
        width={(shape.widthMm ?? 0) + 6} height={(shape.heightMm ?? 0) + 6} />;
    case "ellipse":
      return <ellipse className="free-plan-object-selection" cx={(shape.widthMm ?? 0) / 2}
        cy={(shape.heightMm ?? 0) / 2} rx={(shape.widthMm ?? 0) / 2 + 3}
        ry={(shape.heightMm ?? 0) / 2 + 3} />;
    case "line":
      return <line className="free-plan-object-selection" x1={0} y1={0}
        x2={shape.endXMm} y2={shape.endYMm} />;
    case "label": {
      const fontSize = shape.fontSizeMm ?? 8;
      return <rect className="free-plan-object-selection" x={-3} y={-fontSize - 3}
        width={Math.max(fontSize, (shape.text?.length ?? 1) * fontSize * 0.65) + 6}
        height={fontSize * 1.4 + 6} />;
    }
  }
}
