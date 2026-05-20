export const attachmentCategories = ["Anleitung", "Rechnung", "Decoder-Datei", "Dokumentation", "Ersatzteilliste", "Zertifikat", "Sonstiges"];
export const attachmentAccept = ".pdf,.jpg,.jpeg,.png,.webp,.txt,.csv,.json,.xml,.zip";
export const cvFileAccept = ".json,.csv,.txt,.xml,.z21,.esu,.esux,.lokprogrammer,.zip";
export const imageAccept = ".jpg,.jpeg,.png,.webp";

const blockedAttachmentExtensions = new Set(["exe", "bat", "cmd", "com", "scr", "msi", "dll", "ps1", "vbs", "js", "jar", "sh"]);
const allowedAttachmentExtensions = new Set(["pdf", "jpg", "jpeg", "png", "webp", "txt", "csv", "json", "xml", "zip"]);
const allowedCVFileExtensions = new Set(["json", "csv", "txt", "xml", "z21", "esu", "esux", "lokprogrammer", "zip"]);
const allowedImageExtensions = new Set(["jpg", "jpeg", "png", "webp"]);

function fileExtension(fileName: string) {
  return fileName.split(".").pop()?.toLocaleLowerCase("de-DE") || "";
}

export function isBlockedAttachmentFile(file: File) {
  const extension = fileExtension(file.name);
  return blockedAttachmentExtensions.has(extension) || !allowedAttachmentExtensions.has(extension);
}

export function isBlockedCVFile(file: File) {
  const extension = fileExtension(file.name);
  return blockedAttachmentExtensions.has(extension) || !allowedCVFileExtensions.has(extension);
}

export function isAllowedImageFile(file: File) {
  return allowedImageExtensions.has(fileExtension(file.name));
}

export function attachmentCategoryForFile(file: File) {
  const lower = file.name.toLocaleLowerCase("de-DE");
  if (lower.includes("rechnung") || lower.includes("invoice")) return "Rechnung";
  if (lower.includes("decoder") || lower.endsWith(".json") || lower.endsWith(".xml")) return "Decoder-Datei";
  if (lower.includes("ersatzteil")) return "Ersatzteilliste";
  if (lower.includes("zertifikat") || lower.includes("certificate")) return "Zertifikat";
  if (lower.includes("anleitung") || lower.includes("manual") || lower.includes("bedienung")) return "Anleitung";
  if (lower.endsWith(".pdf")) return "Dokumentation";
  return "Sonstiges";
}
