import type { BugReportArea, BugReportClientDiagnostics } from "../../shared/apiBugReports";
import type { ClientErrorBuffer } from "./clientErrorBuffer";

const diagnosticRouteGroups = new Set([
  "accessories", "article-search", "auth", "backups", "bug-reports", "data-transfer", "database",
  "digital-centers", "ecos", "exhibition", "exhibitions", "import-export", "master-data", "overview",
  "password-reset", "profile", "settings", "setup", "users",
  "vehicle-sets", "vehicle-types", "vehicles", "version"
]);

export function abstractClientErrorRoute(value: string): string {
  const pathname = value.split(/[?#]/, 1)[0];
  const firstSegment = pathname.split("/").find(Boolean)?.toLowerCase() || "";
  return diagnosticRouteGroups.has(firstSegment) ? `/${firstSegment}` : "";
}

export function abstractBugReportPathname(value: string): string {
  return abstractClientErrorRoute(value) || "/";
}

function browserDetails(userAgent: string): [string, string] {
  for (const [family, pattern] of [
    ["Edge", /Edg\/(\d+)/],
    ["Firefox", /Firefox\/(\d+)/],
    ["Chrome", /Chrome\/(\d+)/],
    ["Safari", /Version\/(\d+).*Safari/]
  ] as const) {
    const match = userAgent.match(pattern);
    if (match) return [family, match[1]];
  }
  return ["Unknown", ""];
}

function operatingSystemDetails(userAgent: string): [string, string] {
  if (/Windows NT/i.test(userAgent)) return ["Windows", ""];
  const android = userAgent.match(/Android (\d+)/i);
  if (android) return ["Android", android[1]];
  const ios = userAgent.match(/(?:iPhone OS|CPU OS) (\d+)/i);
  if (ios) return ["iOS", ios[1]];
  const mac = userAgent.match(/Mac OS X (\d+)/i);
  if (mac) return ["macOS", mac[1]];
  if (/Linux/i.test(userAgent)) return ["Linux", ""];
  return ["Unknown", ""];
}

type UserAgentBrand = { brand: string; version: string };
type UserAgentData = {
  brands?: UserAgentBrand[];
  platform?: string;
  getHighEntropyValues?: (hints: string[]) => Promise<{
    fullVersionList?: UserAgentBrand[];
    platformVersion?: string;
  }>;
};

function navigatorUserAgentData(): UserAgentData | undefined {
  return (navigator as Navigator & { userAgentData?: UserAgentData }).userAgentData;
}

function browserDetailsFromBrands(brands: UserAgentBrand[] | undefined): [string, string] | null {
  if (!brands) return null;
  for (const [brandName, family] of [
    ["Microsoft Edge", "Edge"], ["Google Chrome", "Chrome"], ["Chromium", "Chrome"]
  ] as const) {
    const match = brands.find((brand) => brand.brand === brandName);
    if (match) return [family, match.version.split(".", 1)[0].replace(/\D/g, "").slice(0, 4)];
  }
  return null;
}

function operatingSystemDetailsFromHints(platform = "", platformVersion = ""): [string, string] | null {
  const families: Record<string, string> = {
    Windows: "Windows", macOS: "macOS", Android: "Android", iOS: "iOS", Linux: "Linux"
  };
  const family = families[platform];
  if (!family) return null;
  return [family, platformVersion.split(".", 1)[0].replace(/\D/g, "").slice(0, 4)];
}

export function collectBugReportDiagnostics(input: {
  area: BugReportArea;
  roles: string[];
  language: string;
  buffer: ClientErrorBuffer;
}): BugReportClientDiagnostics {
  const hints = navigatorUserAgentData();
  const [browserFamily, browserMajor] = browserDetailsFromBrands(hints?.brands) ||
    browserDetails(navigator.userAgent || "");
  const [osFamily, osMajor] = operatingSystemDetailsFromHints(hints?.platform) ||
    operatingSystemDetails(navigator.userAgent || "");
  const theme = document.documentElement.dataset.theme === "dark" ||
    (!document.documentElement.dataset.theme && window.matchMedia?.("(prefers-color-scheme: dark)").matches)
    ? "dark" : "light";
  return {
    browserFamily,
    browserMajor,
    osFamily,
    osMajor,
    viewportWidth: Math.max(0, Math.round(window.innerWidth)),
    viewportHeight: Math.max(0, Math.round(window.innerHeight)),
    language: input.language.slice(0, 100),
    timeZone: Intl.DateTimeFormat().resolvedOptions().timeZone || "",
    theme,
    pathname: abstractBugReportPathname(window.location.pathname),
    area: input.area,
    roles: input.roles.filter((role) => ["Admin", "Editor", "Viewer", "Messe"].includes(role)).slice(0, 8),
    errors: input.buffer.snapshot().slice(-30)
  };
}

export async function collectBugReportDiagnosticsWithClientHints(input: {
  area: BugReportArea;
  roles: string[];
  language: string;
  buffer: ClientErrorBuffer;
}): Promise<BugReportClientDiagnostics> {
  const diagnostics = collectBugReportDiagnostics(input);
  const hints = navigatorUserAgentData();
  if (!hints?.getHighEntropyValues) return diagnostics;
  try {
    const values = await hints.getHighEntropyValues(["fullVersionList", "platformVersion"]);
    const browser = browserDetailsFromBrands(values.fullVersionList || hints.brands);
    const operatingSystem = operatingSystemDetailsFromHints(hints.platform, values.platformVersion);
    return {
      ...diagnostics,
      ...(browser ? { browserFamily: browser[0], browserMajor: browser[1] } : {}),
      ...(operatingSystem ? { osFamily: operatingSystem[0], osMajor: operatingSystem[1] } : {})
    };
  } catch {
    return diagnostics;
  }
}
