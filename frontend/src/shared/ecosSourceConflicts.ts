import type { DecoderInventory } from "./decoderLibraryModel";

export function hasECoSDecoderIdentityConflict(
  conflicts: Array<{ number: number; indexCV31?: number; indexCV32?: number }> | undefined,
  inventory: DecoderInventory | null | undefined
): boolean {
  return Boolean(conflicts?.some(value => value.indexCV31 == null && value.indexCV32 == null &&
    (value.number === 7 || value.number === 8 ||
      inventory?.models.some(model => model.identification?.cv === value.number))));
}
