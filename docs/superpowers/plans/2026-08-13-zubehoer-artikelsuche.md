# Zubehör-Artikelsuche Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development
> (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use
> checkbox (`- [ ]`) syntax for tracking.

**Goal:** Die Artikelübersicht heißt `Zubehör` und erhält im Artikel-Dialog dieselbe Barcode- und
Artikeldatensuche wie der Fahrzeugbestand, einschließlich kontrollierter Feldübernahme und sicherem
Import ausgewählter Trefferbilder.

**Architecture:** Die bestehenden Fahrzeugdialoge werden zu gemeinsam nutzbaren
Artikel-Suchkomponenten extrahiert. Fahrzeuge und Zubehör liefern jeweils eigene Felddefinitionen
und Formularadapter. Die bestehende Suchroute bleibt unverändert; nur der sichere URL-Import eines
Zubehörbildes erhält eine neue Editor-Route.

**Tech Stack:** React 19, TypeScript 7, Vitest, Testing Library, Vite, Go 1.26, SQLite, OpenAPI 3,
gemeinsame `safefetch`- und Dateiblob-Infrastruktur.

## Global Constraints

- Der Branch `dev/accessory-article-search` basiert direkt auf `origin/main` und übernimmt keine
  lokalen Anlage-Commits.
- Die Route `/accessories`, bestehende Datenmodelle und Suchanbieter-Konfiguration bleiben stabil.
- Deutsche Bereichsbezeichnung: `Zubehör`; englische Bereichsbezeichnung: `Accessories`.
- Artikelart und Unterart werden nie automatisch aus Suchtreffern verändert.
- Unbekannte Hersteller oder Spurweiten werden nicht in Dropdownfelder geschrieben.
- Externe Treffer bleiben Vorschläge und werden nur nach sichtbarer Auswahl übernommen.
- Schreibzugriffe bleiben CSRF-geschützt und serverseitig auf Admin und Editor beschränkt.
- Externe URLs werden ausschließlich über die vorhandenen SSRF-, Redirect-, Größen-, MIME- und
  Dateityp-Schutzmechanismen geladen.
- Produktionscode entsteht ausschließlich nach einem fachlich passenden fehlgeschlagenen Test.

---

## Vorgesehene Dateistruktur

- `frontend/src/shared/articleSearch/articleSearchModel.ts`: domänenneutrale Treffer-, Auswahl- und
  Statushelfer.
- `frontend/src/shared/articleSearch/articleSearchPreferences.ts`: gemeinsame Aktivierungs- und
  Quellenpräferenzen aus dem lokalen Anwendungsspeicher.
- `frontend/src/shared/articleSearch/ArticleSearchDialog.tsx`: gemeinsamer app-eigener
  Ergebnisdialog.
- `frontend/src/shared/articleSearch/BarcodeSearchDialog.tsx`: gemeinsamer manueller,
  Tastatur-Scanner- und Kameraablauf.
- `frontend/src/features/vehicles/articleSearch.ts`: nur noch Fahrzeugfelddefinitionen und
  Fahrzeugwert-Konvertierung.
- `frontend/src/features/accessories/accessoryArticleSearch.ts`: Zubehörsuchkriterien,
  Stammdatenprüfung und Feldübernahme.
- `frontend/src/features/accessories/useAccessoryArticleSearchController.ts`: Zubehörsuchzustand
  und Dialogbefehle.
- `backend/internal/api/remote_file_import.go`: gemeinsam nutzbares, begrenztes Laden externer
  Dateien nach erfolgreicher URL-Prüfung.
- `backend/internal/api/accessory_document_handlers.go`: schmale Zubehörbild-Import-Route.

---

### Task 1: Bereich in Zubehör umbenennen

**Files:**
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`
- Test: `frontend/src/app/Shell.test.tsx`
- Test: `frontend/src/features/accessories/AccessoriesView.test.tsx`

**Interfaces:**
- Consumes: bestehende Schlüssel `nav.accessories`, `accessories.overview.title` und
  `accessories.overview.noAccess`.
- Produces: unveränderte Schlüssel mit den sichtbaren Bereichsnamen `Zubehör` und `Accessories`.

- [ ] **Step 1: Erwartungen auf die neuen Bereichsnamen ändern**

```tsx
expect(navigation).toHaveTextContent("Zubehör");
expect(await screen.findByRole("heading", { name: "Zubehör" })).toBeInTheDocument();
expect(screen.getByText("Kein Zugriff auf Zubehör.")).toBeInTheDocument();
```

Die englische Shell-Prüfung setzt `setLanguage("en")` und erwartet den Link `Accessories`.

- [ ] **Step 2: Zieltests ausführen und das erwartete Rot prüfen**

Run:

```powershell
cd frontend
npm.cmd run test:run -- src/app/Shell.test.tsx src/features/accessories/AccessoriesView.test.tsx
```

Expected: FAIL, weil Navigation und Überschrift noch `Artikelübersicht` beziehungsweise
`Article overview` enthalten.

- [ ] **Step 3: Nur die Bereichsübersetzungen ändern**

```ts
"nav.accessories": "Zubehör",
"accessories.overview.title": "Zubehör",
"accessories.overview.noAccess": "Kein Zugriff auf Zubehör.",
```

```ts
"nav.accessories": "Accessories",
"accessories.overview.title": "Accessories",
"accessories.overview.noAccess": "No access to accessories.",
```

Artikelbezogene Tabellen-, Formular- und Aktionsbegriffe bleiben unverändert.

- [ ] **Step 4: Zieltests grün ausführen**

Run: derselbe Vitest-Befehl wie in Step 2.
Expected: PASS.

- [ ] **Step 5: Commit erstellen**

```powershell
git add frontend/src/shared/i18n/de.ts frontend/src/shared/i18n/en.ts `
  frontend/src/app/Shell.test.tsx frontend/src/features/accessories/AccessoriesView.test.tsx
git commit -m "feat: rename article overview to accessories"
```

---

### Task 2: Fahrzeugsuche in gemeinsame UI-Komponenten extrahieren

**Files:**
- Create: `frontend/src/shared/articleSearch/articleSearchModel.ts`
- Create: `frontend/src/shared/articleSearch/articleSearchPreferences.ts`
- Create: `frontend/src/shared/articleSearch/ArticleSearchDialog.tsx`
- Create: `frontend/src/shared/articleSearch/BarcodeSearchDialog.tsx`
- Modify: `frontend/src/features/vehicles/articleSearch.ts`
- Modify: `frontend/src/features/vehicles/useArticleSearchController.ts`
- Modify: `frontend/src/features/vehicles/VehicleModelTab.tsx`
- Modify: `frontend/src/features/vehicles/VehiclesView.tsx`
- Delete: `frontend/src/features/vehicles/ArticleSearchDialog.tsx`
- Delete: `frontend/src/features/vehicles/BarcodeSearchDialog.tsx`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`
- Test: `frontend/src/features/vehicles/articleSearch.test.ts`
- Test: `frontend/src/features/vehicles/useArticleSearchController.test.tsx`
- Test: `frontend/src/features/vehicles/VehiclesView.test.tsx`

**Interfaces:**
- Consumes: `ArticleSearchResponse`, `ArticleSearchResult`, `ArticleSearchImage` und die bestehende
  Fahrzeugsteuerung.
- Produces:

```ts
export type ArticleSearchFieldDefinition = {
  key: string;
  label: string;
};

export type Translate = (key: string, values?: Record<string, string | number>) => string;

export function articleSearchEnabled(): boolean;
export function articleSearchSources(): string[];
export function isBadCommonArticleValue(key: string, value: string): boolean;

export type ArticleSearchFieldGroup = {
  key: string;
  label: string;
  fields: ArticleSearchFieldDefinition[];
};

export type ArticleSearchDialogProps = {
  fieldGroups: ArticleSearchFieldGroup[];
  currentValue: (key: string) => string;
  loading: boolean;
  response: ArticleSearchResponse | null;
  error: string;
  selectedFields: Record<string, boolean>;
  selectedImages: Record<string, boolean>;
  onApply: (result: ArticleSearchResult) => void;
  onClose: () => void;
  onToggleField: (result: ArticleSearchResult, index: number, key: string, checked: boolean) => void;
  onToggleImage: (result: ArticleSearchResult, index: number, image: ArticleSearchImage,
    checked: boolean) => void;
};
```

- [ ] **Step 1: Fahrzeugtests um gemeinsame Importe und unverändertes Verhalten ergänzen**

```tsx
expect(screen.getByRole("dialog", { name: "Artikeldaten prüfen" })).toBeInTheDocument();
expect(screen.getByRole("button", { name: "Barcode suchen" })).toBeEnabled();
expect(screen.getByText("Bestehender Wert")).toBeInTheDocument();
```

`articleSearch.test.ts` importiert Auswahl- und Statushelfer anschließend aus
`../../shared/articleSearch/articleSearchModel`.

- [ ] **Step 2: Fahrzeugtests ausführen und Importfehler als erwartetes Rot bestätigen**

Run:

```powershell
cd frontend
npm.cmd run test:run -- src/features/vehicles/articleSearch.test.ts `
  src/features/vehicles/useArticleSearchController.test.tsx src/features/vehicles/VehiclesView.test.tsx
```

Expected: FAIL, weil die gemeinsamen Dateien noch nicht existieren.

- [ ] **Step 3: Domänenneutrale Helfer extrahieren**

```ts
export function articleResultKey(result: ArticleSearchResult, index = 0) {
  return `${result.url || result.title}-${index}`;
}

export function articleSelectionKey(result: ArticleSearchResult, key: string, index = 0) {
  return `${articleResultKey(result, index)}::${key}`;
}

export function imageSelectionKey(result: ArticleSearchResult, image: ArticleSearchImage, index = 0) {
  return `${articleResultKey(result, index)}::image::${image.url}`;
}

export function articleFieldStatus(current: string, found: string) {
  if (!current) return "empty" as const;
  if (current.localeCompare(found, "de", { sensitivity: "accent" }) === 0) return "same" as const;
  return "conflict" as const;
}
```

Die beiden Dialoge werden unverändert app-eigen gerendert. Ihre sichtbaren Texte wechseln von
fahrzeugspezifischen Schlüsseln zu `articleSearch.*` und `barcodeSearch.*`.

`articleSearchEnabled` und `articleSearchSources` wechseln ohne Verhaltensänderung aus
`vehicleViewModel.ts` nach `articleSearchPreferences.ts`. Die allgemeinen Filter für leere Werte
und ungeeignete Beschreibungen wechseln nach `articleSearchModel.ts`; fahrzeugspezifische Plausibilitätsregeln
für Länge und Technik bleiben im Fahrzeugfeature.

- [ ] **Step 4: Fahrzeugadapter auf die gemeinsame Schnittstelle umstellen**

```ts
export const vehicleArticleSearchGroups: ArticleSearchFieldGroup[] = articleFieldGroups.map((group) => ({
  key: group.title,
  label: articleGroupLabel(group.title),
  fields: group.keys.map((key) => ({ key, label: articleFieldLabels[key] || key }))
}));

const currentValue = (key: string) => currentArticleValue(form, key as ArticleFieldKey);
```

`VehiclesView` importiert die Dialoge aus `../../shared/articleSearch/...`; der Fahrzeugcontroller
behält Suchkriterien, Feldkonvertierung und Bildentwurf unverändert.

- [ ] **Step 5: Fahrzeugtests und TypeScript-Build grün ausführen**

Run:

```powershell
npm.cmd run test:run -- src/features/vehicles/articleSearch.test.ts `
  src/features/vehicles/useArticleSearchController.test.tsx src/features/vehicles/VehiclesView.test.tsx
npm.cmd run build
```

Expected: alle Tests PASS, Build exit code 0.

- [ ] **Step 6: Commit erstellen**

```powershell
git add frontend/src/shared/articleSearch frontend/src/features/vehicles `
  frontend/src/shared/i18n/de.ts frontend/src/shared/i18n/en.ts
git commit -m "refactor: share article search dialogs"
```

---

### Task 3: Zubehör-Feldadapter testgetrieben ergänzen

**Files:**
- Create: `frontend/src/features/accessories/accessoryArticleSearch.ts`
- Create: `frontend/src/features/accessories/accessoryArticleSearch.test.ts`

**Interfaces:**
- Consumes: `ArticleEditorForm`, `ArticleSearchInput`, `ArticleSearchResult`, `MasterDataEntry` und
  die gemeinsamen Auswahlhelfer.
- Produces:

```ts
export function accessorySearchInput(form: ArticleEditorForm): ArticleSearchInput;
export function hasAccessorySearchCriteria(input: ArticleSearchInput): boolean;
export function accessorySearchFieldGroups(t: Translate): ArticleSearchFieldGroup[];
export function currentAccessorySearchValue(form: ArticleEditorForm, key: string): string;
export function applyAccessorySearchResult(options: {
  form: ArticleEditorForm;
  result: ArticleSearchResult;
  resultIndex: number;
  selectedFields: Record<string, boolean>;
  manufacturers: MasterDataEntry[];
  gauges: MasterDataEntry[];
}): Partial<ArticleEditorForm>;
export function isUsableAccessorySearchValue(key: string, value: string): boolean;
```

- [ ] **Step 1: Fehlende Adapterfunktionen mit fachlichen Tests beschreiben**

```ts
it("maps article criteria without type inference", () => {
  expect(accessorySearchInput({ ...emptyArticleEditorForm(), manufacturer: "Tillig",
    articleNumber: "83101", gauges: ["TT"], ean: "4012500831012" })).toEqual(expect.objectContaining({
    manufacturer: "Tillig",
    articleNumber: "83101",
    gauge: "TT",
    fields: expect.objectContaining({ ean: "4012500831012" })
  }));
});

it("applies selected compatible values but preserves type and unknown master data", () => {
  const patch = applyAccessorySearchResult({
    form: { ...emptyArticleEditorForm(), articleType: "track", subtype: "straight" },
    result,
    resultIndex: 0,
    selectedFields,
    manufacturers: [{ id: "m1", type: "manufacturer", key: "tillig", label: "Tillig", active: true }],
    gauges: [{ id: "g1", type: "gauge", key: "tt", label: "TT", active: true }]
  });
  expect(patch).toMatchObject({ manufacturer: "Tillig", gauges: ["TT"] });
  expect(patch).not.toHaveProperty("articleType");
  expect(patch).not.toHaveProperty("subtype");
});
```

Weitere Tests prüfen: unbekannten Hersteller ignorieren, unbekannte Spurweite ignorieren,
`articleSourceUrl` nach `productUrl` abbilden, Cookie-Text als Beschreibung verwerfen und nicht
ausgewählte Konfliktwerte erhalten.

- [ ] **Step 2: Adaptertests rot ausführen**

Run:

```powershell
cd frontend
npm.cmd run test:run -- src/features/accessories/accessoryArticleSearch.test.ts
```

Expected: FAIL, weil das Adaptermodul fehlt.

- [ ] **Step 3: Minimalen Zubehöradapter implementieren**

```ts
const fieldMap = {
  manufacturer: "manufacturer",
  articleNumber: "articleNumber",
  name: "name",
  ean: "ean",
  scale: "scale",
  description: "description",
  articleSourceUrl: "productUrl"
} as const;

export function accessorySearchInput(form: ArticleEditorForm): ArticleSearchInput {
  return {
    manufacturer: form.manufacturer.trim() || undefined,
    articleNumber: form.articleNumber.trim() || undefined,
    name: form.name.trim() || undefined,
    gauge: form.gauges[0] || undefined,
    searchSources: articleSearchSources(),
    fields: Object.fromEntries(Object.entries({ ean: form.ean, scale: form.scale,
      description: form.description, articleSourceUrl: form.productUrl })
      .map(([key, value]) => [key, value.trim()]).filter(([, value]) => value))
  };
}
```

Die Übernahme iteriert nur über die explizite Feldliste, prüft Auswahlstatus und Stammdaten und
liefert einen Patch statt den Entwurf selbst zu mutieren.

- [ ] **Step 4: Adaptertests grün ausführen**

Run: derselbe Vitest-Befehl wie in Step 2.
Expected: PASS.

- [ ] **Step 5: Commit erstellen**

```powershell
git add frontend/src/features/accessories/accessoryArticleSearch.ts `
  frontend/src/features/accessories/accessoryArticleSearch.test.ts
git commit -m "feat: map accessory article search fields"
```

---

### Task 4: Barcode- und Artikeldatensuche in den Zubehördialog integrieren

**Files:**
- Create: `frontend/src/features/accessories/useAccessoryArticleSearchController.ts`
- Create: `frontend/src/features/accessories/useAccessoryArticleSearchController.test.tsx`
- Modify: `frontend/src/features/accessories/ArticleCoreTab.tsx`
- Modify: `frontend/src/features/accessories/ArticleEditorDialog.tsx`
- Modify: `frontend/src/features/accessories/AccessoriesView.tsx`
- Modify: `frontend/src/features/accessories/useArticleEditorController.ts`
- Modify: `frontend/src/features/accessories/ArticleEditorDialog.test.tsx`
- Modify: `frontend/src/features/accessories/AccessoriesView.test.tsx`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`

**Interfaces:**
- Consumes: Task 2 Dialoge und Task 3 Zubehöradapter.
- Produces:

```ts
export type PendingAccessoryArticleImage = ArticleSearchImage & {
  id: string;
  isPrimary: boolean;
};

export type AccessoryArticleSearchController = {
  state: {
    open: boolean;
    loading: boolean;
    response: ArticleSearchResponse | null;
    error: string;
    barcodeOpen: boolean;
    barcodeValue: string;
    selectedFields: Record<string, boolean>;
    selectedImages: Record<string, boolean>;
  };
  setters: {
    setOpen: (open: boolean) => void;
    setBarcodeOpen: (open: boolean) => void;
    setBarcodeValue: (value: string) => void;
  };
  commands: {
    run: (searchForm?: ArticleEditorForm, searchInput?: ArticleSearchInput) => void;
    openBarcode: () => void;
    submitBarcode: (event: FormEvent<HTMLFormElement>) => void;
    toggleField: (result: ArticleSearchResult, index: number, key: string, checked: boolean) => void;
    toggleImage: (result: ArticleSearchResult, index: number, image: ArticleSearchImage,
      checked: boolean) => void;
    applyResult: (result: ArticleSearchResult) => void;
  };
};

export function useAccessoryArticleSearchController(options: {
  form: ArticleEditorForm;
  readonly: boolean;
  manufacturers: MasterDataEntry[];
  gauges: MasterDataEntry[];
  pendingImageCount: number;
  replaceForm: (form: ArticleEditorForm) => void;
  updateForm: (patch: Partial<ArticleEditorForm>) => void;
  addImages: (images: PendingAccessoryArticleImage[]) => void;
  onMessage: (message: string) => void;
  t: Translate;
}): AccessoryArticleSearchController;
```

- [ ] **Step 1: Controller- und Dialogtests zuerst ergänzen**

```tsx
expect(screen.getByText("Artikelsuche")).toBeInTheDocument();
expect(screen.getByRole("button", { name: "Barcode suchen" })).toBeEnabled();
expect(screen.getByRole("button", { name: "Artikeldaten suchen" })).toBeDisabled();
```

Controller-Test für Barcode:

```tsx
await act(async () => result.current.commands.submitBarcode(submitEvent));
expect(api.articleSearch).toHaveBeenCalledWith(expect.objectContaining({
  fields: { ean: "4012500831012" }
}));
expect(result.current.state.barcodeOpen).toBe(false);
```

Zusätzliche Tests prüfen Nur-Lese-Modus, deaktivierte Websuche, Ladefehler, leere Treffer,
Feldkonflikte und ausgewählte Bilder als Entwurf.

- [ ] **Step 2: Zubehörtests rot ausführen**

Run:

```powershell
cd frontend
npm.cmd run test:run -- src/features/accessories/useAccessoryArticleSearchController.test.tsx `
  src/features/accessories/ArticleEditorDialog.test.tsx src/features/accessories/AccessoriesView.test.tsx
```

Expected: FAIL, weil Controller, Suchblock und Dialogintegration fehlen.

- [ ] **Step 3: Zubehörcontroller mit demselben Zustandsmodell wie die Fahrzeugsuche implementieren**

```ts
const run = (searchForm = form, searchInput = accessorySearchInput(searchForm)) => {
  if (readonly) return;
  if (!articleSearchEnabled()) {
    setOpen(true);
    setResponse(null);
    setError(t("articleSearch.disabled"));
    return;
  }
  if (!hasAccessorySearchCriteria(searchInput)) {
    onMessage(t("articleSearch.missingInput"));
    return;
  }
  setOpen(true);
  setLoading(true);
  setError("");
  void api.articleSearch(searchInput).then(selectInitialEmptyFields)
    .catch((reason: Error) => setError(reason.message)).finally(() => setLoading(false));
};
```

Barcodebestätigung ersetzt nur `ean`, schließt den Barcode-Dialog und ruft `run` mit
`fields: { ean: code }` auf.

- [ ] **Step 4: Suchblock und Dialoge im Artikel-Tab verdrahten**

`ArticleCoreTab` erhält vor Bild und Raster:

```tsx
<div className="article-search-box">
  <div>
    <strong>{t("articleSearch.title")}</strong>
    <span>{t("articleSearch.subtitle")}</span>
  </div>
  <div className="article-search-actions">
    <button type="button" className="secondary-button" onClick={onOpenBarcodeSearch}
      disabled={disabled || articleSearchLoading}>
      <Barcode size={15} aria-hidden="true" />{t("articleSearch.barcode")}
    </button>
    <button type="button" className="secondary-button" onClick={onRunArticleSearch}
      disabled={disabled || articleSearchLoading || !canRunArticleSearch}>
      <PackageSearch size={15} aria-hidden="true" />{t("articleSearch.search")}
    </button>
  </div>
</div>
```

`ArticleEditorDialog` rendert die gemeinsamen Dialoge außerhalb des scrollenden Formularbereichs,
damit Fokusfang und Überlagerung wie beim Fahrzeugdialog funktionieren.

`useArticleEditorController` hält in diesem Task außerdem `pendingArticleImages` und
`setPendingArticleImages`. Öffnen oder Schließen einer Editorsitzung leert diesen Entwurf. Task 6
ergänzt anschließend ausschließlich seine dauerhafte Speicherung.

- [ ] **Step 5: Zubehörtests und Fahrzeugregression grün ausführen**

Run:

```powershell
npm.cmd run test:run -- src/features/accessories/useAccessoryArticleSearchController.test.tsx `
  src/features/accessories/ArticleEditorDialog.test.tsx src/features/accessories/AccessoriesView.test.tsx `
  src/features/vehicles/useArticleSearchController.test.tsx src/features/vehicles/VehiclesView.test.tsx
```

Expected: PASS.

- [ ] **Step 6: Commit erstellen**

```powershell
git add frontend/src/features/accessories frontend/src/shared/i18n/de.ts frontend/src/shared/i18n/en.ts
git commit -m "feat: add accessory article search dialogs"
```

---

### Task 5: Sicheren Zubehörbild-Import bereitstellen

**Files:**
- Create: `backend/internal/api/remote_file_import.go`
- Create: `backend/internal/api/remote_file_import_test.go`
- Modify: `backend/internal/api/vehicle_attachment_handlers.go`
- Modify: `backend/internal/api/accessory_document_handlers.go`
- Modify: `backend/internal/api/accessory_document_handlers_test.go`
- Modify: `backend/internal/api/routes.go`
- Modify: `backend/internal/api/openapi_contract_test.go`
- Modify: `openapi/railkeeper.yaml`

**Interfaces:**
- Consumes: `safefetch.IsPublicHTTPURL`, `remoteDocumentHTTPClient`,
  `normalizeAccessoryDocumentMime`, Fileblob-Speicherung und `AccessoryDocumentService`.
- Produces:

```go
type remoteFileImportInput struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	IsPrimary   bool   `json:"isPrimary"`
}

type fetchedRemoteFile struct {
	Data         []byte
	OriginalName string
	MIMEType     string
}

func fetchRemoteFile(
	ctx context.Context,
	input remoteFileImportInput,
	maxBytes int64,
	accept string,
) (fetchedRemoteFile, error)

func remoteImportedFileName(title, rawURL, mimeType string) string
```

Neue Route:

```text
POST /api/v1/accessory-products/{id}/documents/import-url
```

- [ ] **Step 1: Handler-, Sicherheits- und Vertragsprüfungen schreiben**

```go
func TestAccessoryDocumentImportURLRejectsPrivateTarget(t *testing.T) {
	fixture := newAccessoryAPIFixture(t)
	response := fixture.request(http.MethodPost,
		"/api/v1/accessory-products/"+fixture.product.ID+"/documents/import-url",
		`{"url":"http://127.0.0.1/private.png","title":"Bild"}`)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}
```

Weitere Tests prüfen fehlende URL, Größenlimit, leere Antwort, Nicht-Bild-MIME, Blob-Aufräumen bei
Metadatenfehler, `isPrimary`, Viewer-Verbot und OpenAPI-Schema einschließlich CSRF-Sicherheit.

- [ ] **Step 2: Backendtests rot ausführen**

Run:

```powershell
cd backend
$env:GOCACHE='C:\Users\droth\Documents\GitHub\RailKeeper\.cache\go-build-accessory-search'
go test ./internal/api -run "TestAccessoryDocumentImport|TestOpenAPI" -count=1
```

Expected: FAIL, weil Route, Handler, Helfer und Vertrag fehlen.

- [ ] **Step 3: Gemeinsames begrenztes Remote-Laden extrahieren**

```go
func fetchRemoteFile(
	ctx context.Context,
	input remoteFileImportInput,
	maxBytes int64,
	accept string,
) (fetchedRemoteFile, error) {
	value := strings.TrimSpace(input.URL)
	if value == "" || !isPublicImageURL(ctx, value) {
		return fetchedRemoteFile{}, errRemoteFileURL
	}
	requestCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(requestCtx, http.MethodGet, value, nil)
	if err != nil { return fetchedRemoteFile{}, errRemoteFileURL }
	req.Header.Set("User-Agent", "RailKeeper/0.1 document-fetch")
	req.Header.Set("Accept", accept)
	resp, err := remoteDocumentHTTPClient(ctx).Do(req)
	if err != nil { return fetchedRemoteFile{}, errRemoteFileFetch }
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fetchedRemoteFile{}, errRemoteFileFetch
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil || int64(len(data)) > maxBytes { return fetchedRemoteFile{}, errRemoteFileTooLarge }
	if len(data) == 0 { return fetchedRemoteFile{}, errRemoteFileEmpty }
	mimeType := http.DetectContentType(data)
	return fetchedRemoteFile{Data: data,
		OriginalName: remoteImportedFileName(input.Title, value, mimeType), MIMEType: mimeType}, nil
}
```

Der bestehende Fahrzeug-URL-Import verwendet denselben Helfer, behält aber seine erlaubten
Dokumenttypen. Der Zubehörhandler akzeptiert ausschließlich JPEG, PNG und WebP.

- [ ] **Step 4: Zubehörhandler und Route implementieren**

```go
func (a *App) importAccessoryDocumentFromURL(w http.ResponseWriter, r *http.Request) {
	var input remoteFileImportInput
	if !decodeAccessoryJSON(w, r, &input) { return }
	file, err := fetchRemoteFile(
		r.Context(), input, a.maxAttachmentBytes, "image/jpeg,image/png,image/webp",
	)
	if err != nil { a.respondRemoteFileError(w, err); return }
	if !strings.HasPrefix(file.MIMEType, "image/") {
		respondProblem(w, http.StatusUnsupportedMediaType,
			"accessory_document_type_unsupported", "Accessory image type is not supported.")
		return
	}
	blobID, err := a.storeFileBlob(r.Context(), file.Data)
	if err != nil {
		a.logger.Error("accessory image blob store failed", "error", err)
		respondProblem(w, http.StatusInternalServerError,
			"accessory_document_store_failed", "Accessory image could not be stored.")
		return
	}
	document, err := a.accessoryDocumentService.CreateDocument(r.Context(),
		application.CreateAccessoryDocumentInput{
			ProductID: r.PathValue("id"), FileBlobID: blobID,
			AccessoryDocumentUploadMetadata: application.AccessoryDocumentUploadMetadata{
				FileName: file.OriginalName, OriginalName: file.OriginalName,
				Category: application.AccessoryDocumentImage, MIMEType: file.MIMEType,
				SizeBytes: int64(len(file.Data)), IsPrimary: input.IsPrimary,
			}, Description: strings.TrimSpace(input.Description),
		}, a.maxAttachmentBytes, actorUserID(r))
	if err != nil { a.deleteFileBlobIfUnreferenced(r.Context(), blobID); a.accessoryError(w, err,
		"import accessory image"); return }
	respondJSON(w, http.StatusCreated, document)
}
```

- [ ] **Step 5: OpenAPI-Vertrag ergänzen und Backend formatieren**

Der Vertrag definiert `AccessoryDocumentImportInput` mit erforderlicher `url`, optionalen
`title`, `description`, `isPrimary`, 201-Antwort sowie 400, 403, 404, 413, 415 und 502.

Run:

```powershell
gofmt -w internal/api/remote_file_import.go internal/api/remote_file_import_test.go `
  internal/api/vehicle_attachment_handlers.go internal/api/accessory_document_handlers.go `
  internal/api/accessory_document_handlers_test.go internal/api/routes.go internal/api/openapi_contract_test.go
go test ./internal/api -run "TestAccessoryDocumentImport|TestOpenAPI" -count=1
```

Expected: PASS.

- [ ] **Step 6: Gesamte Backend-Suite ausführen**

Run: `go test ./...`
Expected: alle Pakete PASS.

- [ ] **Step 7: Commit erstellen**

```powershell
git add backend/internal/api openapi/railkeeper.yaml
git commit -m "feat: import accessory images safely"
```

---

### Task 6: Trefferbilder nach dem Artikelspeichern importieren

**Files:**
- Modify: `frontend/src/shared/apiLayoutsAccessories.ts`
- Modify: `frontend/src/shared/apiLayoutsAccessories.test.ts`
- Modify: `frontend/src/features/accessories/useArticleEditorController.ts`
- Modify: `frontend/src/features/accessories/useArticleEditorController.test.tsx`
- Modify: `frontend/src/features/accessories/useAccessoryArticleSearchController.ts`
- Modify: `frontend/src/features/accessories/ArticleEditorDialog.tsx`
- Modify: `frontend/src/features/accessories/AccessoriesView.tsx`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`

**Interfaces:**
- Consumes: Task 4 `PendingAccessoryArticleImage` und Task 5 Importroute.
- Produces:

```ts
export type AccessoryDocumentImportInput = {
  url: string;
  title?: string;
  description?: string;
  isPrimary?: boolean;
};

api.importAccessoryDocumentFromUrl(
  productId: string,
  input: AccessoryDocumentImportInput
): Promise<AccessoryDocument>;
```

- [ ] **Step 1: API- und Speicherablauf als fehlgeschlagene Tests beschreiben**

```ts
await api.importAccessoryDocumentFromUrl("product/1", {
  url: "https://example.invalid/image.png",
  title: "Produktbild",
  isPrimary: true
});
expect(fetch).toHaveBeenCalledWith(
  "/api/v1/accessory-products/product%2F1/documents/import-url",
  expect.objectContaining({ method: "POST" })
);
```

Controller-Test für Teilerfolg:

```ts
vi.mocked(api.createAccessoryArticle).mockResolvedValue(article);
vi.mocked(api.importAccessoryDocumentFromUrl)
  .mockResolvedValueOnce(primaryImage)
  .mockRejectedValueOnce(new Error("Bildimport fehlgeschlagen"));
await act(() => result.current.submit());
expect(result.current.article?.id).toBe(article.id);
expect(result.current.mode).toBe("edit");
expect(result.current.error).toContain("Bildimport fehlgeschlagen");
expect(result.current.pendingArticleImages).toHaveLength(1);
```

Weitere Tests prüfen: erstes Bild nur ohne vorhandenes Primärbild als primär, vollständiger Erfolg
schließt den Dialog, Wiederholung importiert nur fehlgeschlagene Bilder und Öffnen eines anderen
Artikels verwirft alte Bildentwürfe.

- [ ] **Step 2: Zieltests rot ausführen**

Run:

```powershell
cd frontend
npm.cmd run test:run -- src/shared/apiLayoutsAccessories.test.ts `
  src/features/accessories/useArticleEditorController.test.tsx `
  src/features/accessories/useAccessoryArticleSearchController.test.tsx
```

Expected: FAIL, weil API-Methode und Bildentwurfs-Speicherablauf fehlen.

- [ ] **Step 3: API-Client ergänzen**

```ts
importAccessoryDocumentFromUrl: (productId: string, input: AccessoryDocumentImportInput) =>
  request<AccessoryDocument>(
    `/accessory-products/${encodeURIComponent(productId)}/documents/import-url`,
    json("POST", input)
  ),
```

- [ ] **Step 4: Bildentwürfe in den Editorcontroller aufnehmen**

```ts
const importPendingImages = async (saved: AccessoryArticle) => {
  let primaryAssigned = resources.documents.some((document) =>
    document.category === "image" && document.isPrimary);
  const failed: PendingAccessoryArticleImage[] = [];
  let firstError = "";
  for (const image of pendingArticleImages) {
    try {
      await api.importAccessoryDocumentFromUrl(saved.id, {
        url: image.url,
        title: image.title,
        description: image.source,
        isPrimary: !primaryAssigned
      });
      primaryAssigned = true;
    } catch (reason) {
      failed.push(image);
      if (!firstError) firstError = errorMessage(reason);
    }
  }
  setPendingArticleImages(failed);
  return { failed, firstError };
};
```

Nach dem Stammdatenspeichern wechselt ein neu angelegter Artikel sofort in `edit`, damit ein
fehlgeschlagener Bildimport bei erneutem Speichern nicht noch einmal einen Artikel anlegt. Nur bei
vollständigem Erfolg schließt `closeNow()` den Dialog.

- [ ] **Step 5: Zieltests und betroffene Zubehörtests grün ausführen**

Run:

```powershell
npm.cmd run test:run -- src/shared/apiLayoutsAccessories.test.ts `
  src/features/accessories/useArticleEditorController.test.tsx `
  src/features/accessories/useAccessoryArticleSearchController.test.tsx `
  src/features/accessories/ArticleEditorDialog.test.tsx src/features/accessories/AccessoriesView.test.tsx
```

Expected: PASS.

- [ ] **Step 6: Commit erstellen**

```powershell
git add frontend/src/shared/apiLayoutsAccessories.ts `
  frontend/src/shared/apiLayoutsAccessories.test.ts frontend/src/features/accessories `
  frontend/src/shared/i18n/de.ts frontend/src/shared/i18n/en.ts
git commit -m "feat: persist accessory search images"
```

---

### Task 7: Gesamtabnahme und Übergabe

**Files:**
- Modify only if verification exposes a defect: files already listed in Tasks 1 through 6.
- Verify: `docs/superpowers/specs/2026-08-13-zubehoer-artikelsuche-design.md`
- Verify: `openapi/railkeeper.yaml`

**Interfaces:**
- Consumes: alle vorherigen Tasks.
- Produces: releasefähiger Feature-Branch ohne Anlage-Commits.

- [ ] **Step 1: Diff- und Isolationsprüfung durchführen**

Run:

```powershell
git diff --check origin/main...HEAD
git log --oneline origin/main..HEAD
git diff --name-only origin/main...HEAD
```

Expected: keine Whitespacefehler; nur Zubehörsuche, gemeinsame Suchkomponenten, API-Vertrag und
zugehörige Dokumentation. Keine Datei aus den lokalen Anlage-Commits erscheint ausschließlich durch
diesen Branch.

- [ ] **Step 2: Vollständige Backendtests ausführen**

Run:

```powershell
cd backend
$env:GOCACHE='C:\Users\droth\Documents\GitHub\RailKeeper\.cache\go-build-accessory-search'
go test ./...
```

Expected: alle Pakete PASS.

- [ ] **Step 3: Vollständige Frontendtests ausführen**

Run:

```powershell
cd ..\frontend
npm.cmd run test:run
```

Expected: alle Testdateien und Tests PASS.

- [ ] **Step 4: Produktionsbuild erstellen**

Run: `npm.cmd run build`
Expected: TypeScript und Vite exit code 0.

- [ ] **Step 5: Lokalen Browser-Smoke-Test durchführen**

Prüfreihenfolge:

1. Navigation zeigt `Zubehör`, Route bleibt `/accessories`.
2. `Neuer Artikel` öffnet den Artikel-Reiter mit beiden Suchaktionen.
3. Barcode-Dialog akzeptiert Tastatureingabe und bietet Kamera-Fallbackmeldungen.
4. EAN-Suche öffnet den Ergebnisdialog.
5. Leere Werte sind vorausgewählt, Konfliktwerte nicht.
6. Übernahme ändert den Entwurf, speichert aber nicht automatisch.
7. Speichern importiert ausgewählte Bilder; ein Importfehler lässt den gespeicherten Artikel zur
   Wiederholung geöffnet.
8. Viewer sieht keine ausführbaren Suchaktionen.

- [ ] **Step 6: Abschließenden Prüfcommit nur bei notwendigen Korrekturen erstellen**

Falls Step 1 bis 5 eine Korrektur erfordern, zuerst einen reproduzierenden Test ergänzen und danach
committen:

```powershell
git add frontend/src backend/internal/api openapi/railkeeper.yaml
git commit -m "fix: finalize accessory article search"
```

Wenn keine Korrektur nötig ist, entsteht kein leerer Commit.

- [ ] **Step 7: Branch veröffentlichen und PR erstellen**

```powershell
git push -u origin dev/accessory-article-search
gh pr create --base main --head dev/accessory-article-search `
  --title "feat: add accessory article search" `
  --body "Adds accessory search with controlled field application and safe image import."
```

Vor dem Merge müssen alle GitHub-Prüfungen grün sein. Danach:

```powershell
gh pr checks --watch
gh pr merge --merge
```

Der lokale Anlage-Branch bleibt unverändert und wird nicht in `main` übernommen.
