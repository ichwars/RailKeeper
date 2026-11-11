import { useState } from "react";

export function useExpandedInventoryRows() {
  return useState(() => new Set<string>());
}

export type ExpandedInventoryRows = ReturnType<typeof useExpandedInventoryRows>;
