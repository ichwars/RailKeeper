export function decoderProfileLabel(profile: string, fallback: string) {
  if (!profile) return fallback;
  return profile !== profile.trim() || /["\\\r\n\t]/.test(profile) ? JSON.stringify(profile) : profile;
}
