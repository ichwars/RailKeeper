import type { ArticleSearchDocument } from "../../shared/api";

export function webDocumentKey(document: ArticleSearchDocument, index: number) {
  return `${document.url || document.title || "document"}-${index}`;
}

export function categoryForWebDocument(document: ArticleSearchDocument) {
  const signal = `${document.kind || ""} ${document.title || ""}`.toLocaleLowerCase("de-DE");
  if (signal.includes("spare") || signal.includes("ersatzteil") || signal.includes("et-blatt")) {
    return "Ersatzteilliste";
  }
  if (signal.includes("manual") || signal.includes("anleitung") || signal.includes("bedienung")) {
    return "Anleitung";
  }
  return "Dokumentation";
}

export function uniqueWebDocuments(documents: ArticleSearchDocument[]) {
  const unique = new Map<string, ArticleSearchDocument>();
  documents.forEach((document) => {
    const key = (document.url || document.title || "").toLocaleLowerCase();
    if (key && !unique.has(key)) unique.set(key, document);
  });
  return Array.from(unique.values());
}
