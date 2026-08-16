package startup

import (
	"bytes"
	"html/template"
	"net/http"
)

var conflictPageTemplate = template.Must(template.New("legacy-conflict").Parse(`<!doctype html>
<html lang="de">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>RailKeeper Sicherheitsstopp</title>
  <style>
    :root { color-scheme: dark; font-family: system-ui, sans-serif; background: #0d1213; color: #edf3f1; }
    body { margin: 0; padding: 32px 20px; }
    main { max-width: 850px; margin: 0 auto; }
    section { margin: 24px 0; padding: 22px; border: 1px solid #344043; border-radius: 10px; background: #151c1e; }
    h1, h2 { margin-top: 0; }
    h1 { color: #ffcf67; }
    h2 { color: #9fdc62; font-size: 1.15rem; }
    code { display: block; overflow-wrap: anywhere; padding: 10px 12px; border-radius: 6px; background: #0b1011; color: #d7e7df; }
    li { margin: 8px 0; }
    .label { margin: 14px 0 6px; color: #afbfba; font-weight: 650; }
  </style>
</head>
<body>
<main>
  <section lang="de">
    <h1>RailKeeper Sicherheitsstopp</h1>
    <p>RailKeeper hat zwei Datenbanken gefunden und keine davon verändert.</p>
    <p>Eine automatische Zusammenführung ist absichtlich gesperrt, damit keine Bestandsdaten verloren gehen.</p>
    <p class="label">Sicherer neuer Datenordner</p>
    <code>{{.SafePath}}</code>
    <p class="label">Bisheriger Datenordner</p>
    <code>{{.LegacyPath}}</code>
    <h2>Sicheres weiteres Vorgehen</h2>
    <ol>
      <li>RailKeeper schließen.</li>
      <li>Beide Ordner vollständig sichern.</li>
      <li>Den nicht gewünschten sicheren Ordner umbenennen oder verschieben, oder den gewünschten Ordner ausdrücklich über <code>RAILKEEPER_DATA_DIR</code> konfigurieren.</li>
      <li>Die gewählte Installation prüfen und bis dahin beide Kopien behalten.</li>
    </ol>
  </section>
  <section lang="en">
    <h1>RailKeeper safety stop</h1>
    <p>RailKeeper found two databases and changed neither one.</p>
    <p>Automatic combination is intentionally unavailable so that inventory data remains protected.</p>
    <p class="label">Safe new data folder</p>
    <code>{{.SafePath}}</code>
    <p class="label">Previous data folder</p>
    <code>{{.LegacyPath}}</code>
    <h2>Safe next steps</h2>
    <ol>
      <li>Close RailKeeper.</li>
      <li>Create complete backup copies of both folders.</li>
      <li>Rename or move the safe folder you do not intend to use, or explicitly configure the chosen folder through <code>RAILKEEPER_DATA_DIR</code>.</li>
      <li>Verify the chosen installation and retain both copies until that check is complete.</li>
    </ol>
  </section>
</main>
</body>
</html>`))

func ConflictHandler(info LegacyConflictInfo) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Cache-Control", "no-store")
		response.Header().Set(
			"Content-Security-Policy",
			"default-src 'none'; style-src 'unsafe-inline'; base-uri 'none'; form-action 'none'",
		)
		response.Header().Set("X-Content-Type-Options", "nosniff")
		if request.Method != http.MethodGet {
			response.Header().Set("Allow", http.MethodGet)
			response.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var page bytes.Buffer
		if err := conflictPageTemplate.Execute(&page, info); err != nil {
			http.Error(response, "RailKeeper safety page unavailable", http.StatusInternalServerError)
			return
		}
		response.Header().Set("Content-Type", "text/html; charset=utf-8")
		response.WriteHeader(http.StatusConflict)
		_, _ = response.Write(page.Bytes())
	})
}
