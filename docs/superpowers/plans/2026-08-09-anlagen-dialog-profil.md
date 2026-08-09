# Anlagen-Dialog und erweitertes Anlagenprofil Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Anlagen werden über einen gemeinsamen RailKeeper-Dialog angelegt und bearbeitet, während Verzeichnis und Profil die vorhandenen Anlagendaten vollständig darstellen.

**Architecture:** Zwei neue Shared-UI-Felder kapseln Textbereich und Checkbox. Ein fokussierter `LayoutFormDialog` verwaltet den lokalen Formularentwurf und die Dialoginteraktion; `LayoutsView` und `LayoutWorkspace` bleiben für API-Aufrufe, Auswahl und Versionskonflikte verantwortlich. Das Backend und der API-Vertrag bleiben unverändert.

**Tech Stack:** React 19, TypeScript 7 strict, Vite 8, Vitest 4, Testing Library, bestehende RailKeeper-CSS-Tokens und Lucide-Icons.

## Global Constraints

- Änderungen bleiben auf das Frontend begrenzt; Datenbank, Backend und OpenAPI bleiben unverändert.
- Anlegen und Bearbeiten verwenden denselben `LayoutFormDialog`.
- In der Oberfläche sind keine plattformspezifisch dargestellten Browser-Steuerelemente sichtbar.
- Deutsch und Englisch bleiben vollständig synchron.
- Rollen-, Konflikt- und Archivierungsverhalten bleiben erhalten.
- Desktop, schmale Ansicht, helle und dunkle Darstellung müssen funktionieren.
- Änderungen werden ausschließlich inline in dieser Session ausgeführt.

---

## File Map

- Create `frontend/src/shared/ui/AppTextArea.tsx`: zugänglicher app-eigener Textbereich.
- Create `frontend/src/shared/ui/AppTextArea.test.tsx`: Semantik, Hilfe, Fehler und Ref.
- Create `frontend/src/shared/ui/AppCheckbox.tsx`: app-eigene Checkbox mit kontrolliertem Wert.
- Create `frontend/src/shared/ui/AppCheckbox.test.tsx`: Semantik, Wertänderung und deaktivierter Zustand.
- Create `frontend/src/features/layouts/LayoutFormDialog.tsx`: gemeinsamer Anlage-/Bearbeiten-Dialog.
- Create `frontend/src/features/layouts/LayoutFormDialog.test.tsx`: Formular-, Fokus- und Verwerfverhalten.
- Modify `frontend/src/features/layouts/LayoutConfirmDialog.tsx`: portalfähige, tastaturbedienbare Bestätigung.
- Modify `frontend/src/features/layouts/LayoutsView.tsx`: Vollbreitenverzeichnis und Anlegedialog.
- Modify `frontend/src/features/layouts/LayoutWorkspace.tsx`: erweitertes Profil und Bearbeitungsdialog.
- Modify `frontend/src/features/layouts/LayoutsView.test.tsx`: Integrationsfälle für Anlegen, Bearbeiten und Konflikt.
- Modify `frontend/src/shared/i18n/de.ts`: neue deutsche Aktions-, Metadaten- und Dialogtexte.
- Modify `frontend/src/shared/i18n/en.ts`: entsprechende englische Texte.
- Modify `frontend/src/styles/forms-controls.css`: visuelle Shared-UI-Felder.
- Modify `frontend/src/styles/layouts.css`: Verzeichnis-, Profil- und Dialoglayout.

---

### Task 1: App-eigener Textbereich und Checkbox

**Files:**
- Create: `frontend/src/shared/ui/AppTextArea.tsx`
- Create: `frontend/src/shared/ui/AppTextArea.test.tsx`
- Create: `frontend/src/shared/ui/AppCheckbox.tsx`
- Create: `frontend/src/shared/ui/AppCheckbox.test.tsx`
- Modify: `frontend/src/styles/forms-controls.css`

**Interfaces:**
- Produces: `AppTextAreaProps extends TextareaHTMLAttributes<HTMLTextAreaElement>` mit `label`, `helpText` und `error`.
- Produces: `AppCheckboxProps extends Omit<InputHTMLAttributes<HTMLInputElement>, "type">` mit `label`.
- Consumes: bestehende `.app-field`, `.app-field-label`, `.app-field-help` und `.app-field-error`-Konventionen.

- [ ] **Step 1: Tests für die neuen Shared-UI-Felder schreiben**

```tsx
it("connects textarea label, help, error, invalid state, and ref", () => {
  const ref = createRef<HTMLTextAreaElement>();
  render(<AppTextArea ref={ref} label="Beschreibung" helpText="Interne Notiz"
    error="Beschreibung fehlt" value="" onChange={() => undefined} />);
  const field = screen.getByRole("textbox", { name: "Beschreibung" });
  expect(ref.current).toBe(field);
  expect(field).toHaveAttribute("aria-invalid", "true");
  expect(field).toHaveAccessibleDescription("Interne Notiz Beschreibung fehlt");
});

it("emits checkbox changes and preserves disabled semantics", async () => {
  const user = userEvent.setup();
  const onChange = vi.fn();
  const { rerender } = render(<AppCheckbox label="Archiviert" checked={false} onChange={onChange} />);
  await user.click(screen.getByRole("checkbox", { name: "Archiviert" }));
  expect(onChange).toHaveBeenCalled();
  rerender(<AppCheckbox label="Archiviert" checked={false} disabled onChange={onChange} />);
  expect(screen.getByRole("checkbox", { name: "Archiviert" })).toBeDisabled();
});
```

- [ ] **Step 2: Tests ausführen und das erwartete Fehlschlagen bestätigen**

Run: `cd frontend; npm.cmd run test:run -- src/shared/ui/AppTextArea.test.tsx src/shared/ui/AppCheckbox.test.tsx`

Expected: FAIL, weil `AppTextArea` und `AppCheckbox` noch nicht existieren.

- [ ] **Step 3: Beide Komponenten implementieren**

```tsx
export const AppTextArea = forwardRef<HTMLTextAreaElement, AppTextAreaProps>(function AppTextArea(
  { label, helpText, error, className = "", id, required, ...props }, ref
) {
  const generatedId = useId();
  const inputId = id || generatedId;
  const helpId = helpText ? `${inputId}-help` : undefined;
  const errorId = error ? `${inputId}-error` : undefined;
  const describedBy = [props["aria-describedby"], helpId, errorId].filter(Boolean).join(" ") || undefined;
  return <div className={`app-field app-text-area ${error ? "has-error" : ""} ${className}`.trim()}>
    <label className="app-field-label" htmlFor={inputId}>{label}{required ? <span aria-hidden="true"> *</span> : null}</label>
    <textarea {...props} ref={ref} id={inputId} required={required} aria-describedby={describedBy}
      aria-invalid={error ? true : props["aria-invalid"]} />
    {helpText ? <span id={helpId} className="app-field-help">{helpText}</span> : null}
    {error ? <span id={errorId} className="app-field-error" role="alert">{error}</span> : null}
  </div>;
});

export const AppCheckbox = forwardRef<HTMLInputElement, AppCheckboxProps>(function AppCheckbox(
  { label, className = "", id, ...props }, ref
) {
  const generatedId = useId();
  const inputId = id || generatedId;
  return <label className={`app-checkbox ${className}`.trim()} htmlFor={inputId}>
    <input {...props} ref={ref} id={inputId} type="checkbox" />
    <span className="app-checkbox-mark" aria-hidden="true"><Check size={13} /></span>
    <span>{label}</span>
  </label>;
});
```

Add CSS that sets `appearance: none` on `.app-checkbox input`, draws focus, checked, disabled and
error states with existing tokens, and styles `.app-text-area textarea` like `.app-text-input input`.

- [ ] **Step 4: Shared-UI-Tests ausführen**

Run: `cd frontend; npm.cmd run test:run -- src/shared/ui/AppTextArea.test.tsx src/shared/ui/AppCheckbox.test.tsx`

Expected: PASS, 2 tests.

- [ ] **Step 5: Shared-UI-Felder committen**

```powershell
git add frontend/src/shared/ui/AppTextArea.tsx frontend/src/shared/ui/AppTextArea.test.tsx `
  frontend/src/shared/ui/AppCheckbox.tsx frontend/src/shared/ui/AppCheckbox.test.tsx `
  frontend/src/styles/forms-controls.css
git commit -m "feat(ui): add textarea and checkbox controls"
```

---

### Task 2: Gemeinsamen Anlagendialog bauen

**Files:**
- Create: `frontend/src/features/layouts/LayoutFormDialog.tsx`
- Create: `frontend/src/features/layouts/LayoutFormDialog.test.tsx`
- Modify: `frontend/src/features/layouts/LayoutConfirmDialog.tsx`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`
- Modify: `frontend/src/styles/layouts.css`

**Interfaces:**
- Consumes: `AppTextInput`, `AppTextArea`, `AppCheckbox`, `AppSelect`, `LayoutInput`, `LayoutKind`.
- Produces: `LayoutFormValue = Required<Pick<LayoutInput, "name" | "kind" | "gauge" | "scale">> & { description: string; archived: boolean }`.
- Produces: `LayoutFormDialogProps` mit `mode`, `initialValue`, `saving`, `message`, `conflict`, `returnFocusTo`, `onSubmit`, `onReloadConflict` und `onClose`.

- [ ] **Step 1: Dialogtests für Controls, Submit, Fokus und Verwerfen schreiben**

```tsx
const initialValue: LayoutFormValue = {
  name: "meine", kind: "private", gauge: "TT", scale: "1:120", description: "", archived: false
};

it("uses RailKeeper controls and submits a create draft", async () => {
  const user = userEvent.setup();
  const onSubmit = vi.fn();
  render(<LayoutFormDialog mode="create" initialValue={initialValue} saving={false} message=""
    conflict={false} onSubmit={onSubmit} onClose={() => undefined} />);
  expect(screen.getByRole("dialog", { name: "Anlage anlegen" })).toBeInTheDocument();
  await user.clear(screen.getByRole("textbox", { name: "Bezeichnung" }));
  await user.type(screen.getByRole("textbox", { name: "Bezeichnung" }), "Heimanlage");
  await user.click(screen.getByRole("button", { name: "Anlage speichern" }));
  expect(onSubmit).toHaveBeenCalledWith({ ...initialValue, name: "Heimanlage" });
});

it("asks before discarding a dirty draft and restores focus", async () => {
  const user = userEvent.setup();
  const trigger = document.createElement("button");
  document.body.append(trigger);
  trigger.focus();
  const onClose = vi.fn();
  render(<LayoutFormDialog mode="edit" initialValue={initialValue} saving={false} message=""
    conflict={false} returnFocusTo={trigger} onSubmit={() => undefined} onClose={onClose} />);
  await user.type(screen.getByRole("textbox", { name: "Beschreibung" }), "Geändert");
  await user.keyboard("{Escape}");
  expect(screen.getByRole("dialog", { name: "Änderungen verwerfen?" })).toBeInTheDocument();
  await user.click(screen.getByRole("button", { name: "Verwerfen" }));
  expect(onClose).toHaveBeenCalled();
  expect(trigger).toHaveFocus();
});

it("keeps the draft visible when saving fails", async () => {
  const user = userEvent.setup();
  const { rerender } = render(<LayoutFormDialog mode="edit" initialValue={initialValue}
    saving={false} message="" conflict={false} onSubmit={() => undefined} onClose={() => undefined} />);
  const name = screen.getByRole("textbox", { name: "Bezeichnung" });
  await user.clear(name);
  await user.type(name, "Lokaler Entwurf");
  rerender(<LayoutFormDialog mode="edit" initialValue={initialValue} saving={false}
    message="Speichern fehlgeschlagen" conflict={false} onSubmit={() => undefined}
    onClose={() => undefined} />);
  expect(screen.getByRole("alert")).toHaveTextContent("Speichern fehlgeschlagen");
  expect(name).toHaveValue("Lokaler Entwurf");
});
```

- [ ] **Step 2: Dialogtests ausführen und das erwartete Fehlschlagen bestätigen**

Run: `cd frontend; npm.cmd run test:run -- src/features/layouts/LayoutFormDialog.test.tsx`

Expected: FAIL, weil `LayoutFormDialog` noch nicht existiert.

- [ ] **Step 3: Dialog und verbesserte Bestätigung implementieren**

Implement `LayoutFormDialog` als Portal in `document.body`. Verwende einen `form`-Submit, eine
lokale `form`-State-Kopie und eine serialisierte Gleichheitsprüfung gegen `initialValue`. Der
Keydown-Handler behandelt Escape und zyklisches Tabben:

```tsx
const requestClose = () => {
  if (saving) return;
  if (JSON.stringify(form) !== JSON.stringify(initialValue)) setDiscardOpen(true);
  else onClose();
};

const submit = (event: FormEvent) => {
  event.preventDefault();
  if (form.name.trim()) onSubmit({ ...form, name: form.name.trim(), gauge: form.gauge.trim(),
    scale: form.scale.trim(), description: form.description.trim() });
};
```

Render the field grid with `AppTextInput`, `AppSelect`, `AppTextArea` and `AppCheckbox`. Keep archive
hidden in create mode. Render `message` as `role="alert"`; when `conflict` is true, include the
`Serverstand neu laden` action. Upgrade `LayoutConfirmDialog` to portal into `document.body`, trap
focus, support Escape, and temporarily mark a parent dialog inert while nested.

- [ ] **Step 4: Deutsche und englische Texte ergänzen**

Add matching keys in both translation files:

```ts
"layouts.create.action": "Anlage anlegen",
"layouts.edit.action": "Bearbeiten",
"layouts.dialog.discardTitle": "Änderungen verwerfen?",
"layouts.dialog.discardBody": "Deine nicht gespeicherten Änderungen gehen verloren.",
"layouts.dialog.discardAction": "Verwerfen",
"layouts.overview.created": "Erstellt",
"layouts.overview.version": "Version",
```

English values: `Create layout`, `Edit`, `Discard changes?`, `Your unsaved changes will be lost.`,
`Discard`, `Created`, `Version`.

- [ ] **Step 5: Dialogtests ausführen**

Run: `cd frontend; npm.cmd run test:run -- src/features/layouts/LayoutFormDialog.test.tsx src/features/layouts/LayoutsView.test.tsx src/shared/i18n.test.ts`

Expected: Dialogtests PASS; bestehende `LayoutsView`-Tests dürfen nur dort FAIL melden, wo sie noch
das alte Inlineformular erwarten.

- [ ] **Step 6: Anlagendialog committen**

```powershell
git add frontend/src/features/layouts/LayoutFormDialog.tsx `
  frontend/src/features/layouts/LayoutFormDialog.test.tsx `
  frontend/src/features/layouts/LayoutConfirmDialog.tsx frontend/src/shared/i18n/de.ts `
  frontend/src/shared/i18n/en.ts frontend/src/styles/layouts.css
git commit -m "feat(layouts): add shared layout form dialog"
```

---

### Task 3: Anlagenverzeichnis und Anlageablauf umstellen

**Files:**
- Modify: `frontend/src/features/layouts/LayoutsView.tsx`
- Modify: `frontend/src/features/layouts/LayoutsView.test.tsx`
- Modify: `frontend/src/styles/layouts.css`

**Interfaces:**
- Consumes: `LayoutFormDialog` und `LayoutFormValue` aus Task 2.
- Produces: Vollbreitenverzeichnis mit `Anlage anlegen`-Button und detaillierten Karten.

- [ ] **Step 1: Integrationstest auf den neuen Anlageablauf umstellen**

```tsx
it("opens the create dialog and selects the created layout", async () => {
  const user = userEvent.setup();
  vi.mocked(api.layouts).mockResolvedValueOnce([]).mockResolvedValue([layout]);
  vi.spyOn(api, "createLayout").mockResolvedValue(layout);
  render(<LayoutsView roles={["Planner"]} />);
  await screen.findByText("Noch keine Anlage erfasst.");
  await user.click(screen.getByRole("button", { name: "Anlage anlegen" }));
  const dialog = screen.getByRole("dialog", { name: "Anlage anlegen" });
  await user.type(within(dialog).getByRole("textbox", { name: "Bezeichnung" }), "Clubanlage");
  await user.click(within(dialog).getByRole("button", { name: "Anlage speichern" }));
  await waitFor(() => expect(api.createLayout).toHaveBeenCalledWith(expect.objectContaining({
    name: "Clubanlage", kind: "private", gauge: "TT", scale: "1:120"
  })));
  expect(await screen.findByText(layout.name)).toBeInTheDocument();
});
```

Extend the viewer test to assert no create action and assert directory text for `TT`, `1:120`,
`Version 2` and the formatted update date.

- [ ] **Step 2: Angepassten Integrationstest ausführen und das erwartete Fehlschlagen bestätigen**

Run: `cd frontend; npm.cmd run test:run -- src/features/layouts/LayoutsView.test.tsx`

Expected: FAIL, weil der Button und Dialog noch nicht integriert sind.

- [ ] **Step 3: Inlineformular durch Dialogsteuerung ersetzen**

Replace `form` with dialog state and trigger focus tracking:

```tsx
const [createOpen, setCreateOpen] = useState(false);
const [createError, setCreateError] = useState("");
const createTriggerRef = useRef<HTMLButtonElement | null>(null);

const createLayout = async (draft: LayoutFormValue) => {
  setSaving(true); setCreateError("");
  try {
    const created = await api.createLayout({ ...draft, archived: undefined,
      description: draft.description || undefined });
    await loadLayouts(created.id);
    setCreateOpen(false);
  } catch (reason) {
    setCreateError(reason instanceof Error ? reason.message : t("layouts.error.generic"));
  } finally { setSaving(false); }
};
```

Render a panel head with the create button, detailed cards and `LayoutFormDialog` outside the panel.
Each card must show name, kind, gauge, scale, version, update time and status without duplicating the
selected name in the accessible button label.

- [ ] **Step 4: Verzeichnisstyles anpassen**

Remove the two-column `.layout-catalog-grid` definition and obsolete `.layout-create-panel` rules.
Add `.layout-panel-head`, `.layout-card-meta`, `.layout-card-facts` and responsive wrapping. Preserve
the existing selected border and calm accent background.

- [ ] **Step 5: Anlagenverzeichnis testen**

Run: `cd frontend; npm.cmd run test:run -- src/features/layouts/LayoutsView.test.tsx`

Expected: all `LayoutsView` tests PASS except the edit-conflict test, which is updated in Task 4.

- [ ] **Step 6: Anlagenverzeichnis committen**

```powershell
git add frontend/src/features/layouts/LayoutsView.tsx `
  frontend/src/features/layouts/LayoutsView.test.tsx frontend/src/styles/layouts.css
git commit -m "feat(layouts): expand layout directory"
```

---

### Task 4: Anlagenprofil und Bearbeitungsablauf umstellen

**Files:**
- Modify: `frontend/src/features/layouts/LayoutWorkspace.tsx`
- Modify: `frontend/src/features/layouts/LayoutsView.test.tsx`
- Modify: `frontend/src/styles/layouts.css`

**Interfaces:**
- Consumes: `LayoutFormDialog`, `LayoutFormValue`, bestehende `api.updateLayout`- und `api.layout`-Methoden.
- Produces: Vollbreites Anlagenprofil und Bearbeiten-Dialog mit Versionskonfliktschutz.

- [ ] **Step 1: Profil- und Bearbeitungsfälle schreiben**

```tsx
it("shows the complete profile and edits through the dialog", async () => {
  const user = userEvent.setup();
  vi.spyOn(api, "updateLayout").mockResolvedValue({ ...layout, name: "Neue Bezeichnung", version: 3 });
  render(<LayoutsView roles={["Planner"]} />);
  await screen.findAllByText(layout.name);
  const profile = screen.getByText("Anlagenprofil").closest(".panel") as HTMLElement;
  expect(profile).toHaveTextContent("Clubanlage");
  expect(profile).toHaveTextContent("TT");
  expect(profile).toHaveTextContent("1:120");
  expect(profile).toHaveTextContent("Version 2");
  expect(profile).toHaveTextContent("1");
  await user.click(within(profile).getByRole("button", { name: "Bearbeiten" }));
  const dialog = screen.getByRole("dialog", { name: "Anlage bearbeiten" });
  const name = within(dialog).getByRole("textbox", { name: "Bezeichnung" });
  await user.clear(name);
  await user.type(name, "Neue Bezeichnung");
  await user.click(within(dialog).getByRole("button", { name: "Änderungen speichern" }));
  await waitFor(() => expect(api.updateLayout).toHaveBeenCalledWith(layout.id,
    expect.objectContaining({ name: "Neue Bezeichnung", expectedVersion: 2 })));
});
```

Rewrite the conflict test to open the dialog first, retain the edited name after HTTP 409, click
`Serverstand neu laden`, and assert `api.layout(layout.id)` was called.

- [ ] **Step 2: Profiltests ausführen und das erwartete Fehlschlagen bestätigen**

Run: `cd frontend; npm.cmd run test:run -- src/features/layouts/LayoutsView.test.tsx`

Expected: FAIL, weil das alte Seitenformular noch existiert und Profilfelder fehlen.

- [ ] **Step 3: Profilraster und Dialogintegration implementieren**

Add edit dialog state and map the selected layout into a `LayoutFormValue`. Keep API ownership in
`LayoutWorkspace`:

```tsx
const [editOpen, setEditOpen] = useState(false);
const editTriggerRef = useRef<HTMLButtonElement | null>(null);

const saveLayout = async (draft: LayoutFormValue) => {
  setSaving(true); setMessage(""); setConflict(false);
  try {
    const updated = await api.updateLayout(layout.id, {
      ...draft, description: draft.description || undefined, expectedVersion: layout.version
    });
    onLayoutChanged(updated); setEditOpen(false);
  } catch (reason) {
    if (reason instanceof ApiError && reason.status === 409) {
      setConflict(true); setMessage(t("layouts.conflict.message"));
    } else setMessage(reason instanceof Error ? reason.message : t("layouts.error.generic"));
  } finally { setSaving(false); }
};
```

Move `message` into the dialog while edit mode is open. Render a full-width profile panel with a
header action and a definition-list grid for kind, status, gauge, scale, version, unit count, setup
count, created time and updated time. Keep the description in a separate block beneath the facts.

- [ ] **Step 4: Konfliktneuladen mit dem Dialog synchronisieren**

`reloadServerState` must update the parent layout and leave the dialog open. Pass the resulting
`layout` back as new `initialValue`; `LayoutFormDialog` resets its local draft when that prop changes.
Clear `message` and `conflict` only after the reload succeeds.

- [ ] **Step 5: Profil- und Konflikttests ausführen**

Run: `cd frontend; npm.cmd run test:run -- src/features/layouts/LayoutsView.test.tsx src/features/layouts/LayoutFormDialog.test.tsx`

Expected: PASS for all layout and dialog tests.

- [ ] **Step 6: Anlagenprofil committen**

```powershell
git add frontend/src/features/layouts/LayoutWorkspace.tsx `
  frontend/src/features/layouts/LayoutsView.test.tsx frontend/src/styles/layouts.css
git commit -m "feat(layouts): expand layout profile"
```

---

### Task 5: Gesamtprüfung und visuelle Abnahme

**Files:**
- Modify if required by findings: only files listed in Tasks 1 through 4.

**Interfaces:**
- Consumes: completed layout UI and existing local server at `http://127.0.0.1:18083/layouts`.
- Produces: verified frontend without backend or contract changes.

- [ ] **Step 1: Frontendtests vollständig ausführen**

Run: `cd frontend; npm.cmd run test:run`

Expected: all tests PASS.

- [ ] **Step 2: Produktionsbuild erstellen**

Run: `cd frontend; npm.cmd run build`

Expected: TypeScript and Vite complete successfully and write `frontend/dist`; `dist` remains untracked.

- [ ] **Step 3: Git-Diff auf Scope und Sauberkeit prüfen**

Run: `git diff --check; git status --short`

Expected: no whitespace errors; only intended source, test, translation, style and plan changes are
present; no `frontend/dist`, data, cache or credentials are tracked.

- [ ] **Step 4: Lokalen Server mit aktuellem Build prüfen**

Open `http://127.0.0.1:18083/layouts`, sign in with the existing local QA session if needed, and verify:

1. directory metadata and selection,
2. create dialog controls, focus, cancellation and save,
3. complete profile data,
4. edit dialog, archive control and save,
5. dark and light theme,
6. desktop and narrow viewport,
7. no console errors.

- [ ] **Step 5: Eventuelle Abnahmefunde eng korrigieren und erneut prüfen**

For each finding, add or update the narrowest failing test first, run it to FAIL, patch only the
responsible file, rerun the targeted test to PASS, then repeat Steps 1 through 4.

- [ ] **Step 6: Finalen Prüfstand committen, falls Abnahmeänderungen nötig waren**

```powershell
git add frontend/src/features/layouts frontend/src/shared/ui frontend/src/shared/i18n/de.ts `
  frontend/src/shared/i18n/en.ts frontend/src/styles/forms-controls.css frontend/src/styles/layouts.css
git commit -m "fix(layouts): finalize layout dialog presentation"
```

Skip this commit when Step 5 produced no file changes.
