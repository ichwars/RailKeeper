export type FunctionSymbolImageVariant = "active" | "inactive" | "print";

function metadataString(metadata: Record<string, unknown> | undefined, key: string) {
  const value = metadata?.[key];
  return typeof value === "string" ? value : "";
}

export function functionSymbolImageData(
  metadata?: Record<string, unknown>,
  variant: FunctionSymbolImageVariant = "active",
) {
  const keys =
    variant === "active"
      ? ["activeImageData", "imageData", "svgData"]
      : variant === "inactive"
        ? ["inactiveImageData", "imageData", "activeImageData", "svgData"]
        : ["imageData", "activeImageData", "svgData"];
  return keys.map((key) => metadataString(metadata, key)).find(Boolean) || "";
}
