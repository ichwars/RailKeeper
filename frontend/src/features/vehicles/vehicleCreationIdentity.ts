export function validVehicleCreationIdentity(value: unknown): value is string {
  return typeof value === "string" && /^[0-9a-f]{32}$/.test(value);
}

export function newVehicleCreationIdentity(): string {
  const bytes = new Uint8Array(16);
  crypto.getRandomValues(bytes);
  return Array.from(bytes, value => value.toString(16).padStart(2, "0")).join("");
}
