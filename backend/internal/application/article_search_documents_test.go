package application

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestArticleDocumentSuggestionsResolveAndDeduplicateSourceLinks(t *testing.T) {
	page := `
<a href="../files/manual.pdf"><span>Bedienungsanleitung &amp; Wartung</span></a>
<a href="../files/manual.pdf">Duplicate description</a>
<a href="/files/spare.pdf">Ersatzteilblatt</a>
<a href="/files/catalog-data.pdf">Download</a>
<a href="/items/another.html">Another vehicle</a>
<a class="disabled">Missing link</a>`
	actual := articleDocumentsFromHTML(page, "https://manufacturer.example/items/12345.html")
	want := []ArticleSearchDocument{
		{Title: "Bedienungsanleitung & Wartung", URL: "https://manufacturer.example/files/manual.pdf", Source: "manufacturer", Kind: "manual"},
		{Title: "Ersatzteilblatt", URL: "https://manufacturer.example/files/spare.pdf", Source: "manufacturer", Kind: "spare-parts"},
		{Title: "catalog-data.pdf", URL: "https://manufacturer.example/files/catalog-data.pdf", Source: "manufacturer", Kind: "document"},
	}
	if !reflect.DeepEqual(actual, want) {
		t.Fatal("document suggestions changed source identity, title, kind or deduplication", actual)
	}
}

func TestArticleDocumentSuggestionsRemainBoundedAndExcludeHiddenPageText(t *testing.T) {
	var page strings.Builder
	for index := 0; index < 20; index++ {
		fmt.Fprintf(&page, `<a href="/files/document-%d.pdf">Document %d</a>`, index, index)
	}
	if values := articleDocumentsFromHTML(page.String(), "https://manufacturer.example/12345.html"); len(values) != 12 {
		t.Fatal("page supplied an unbounded document list", len(values))
	}
	html := `<style>.hidden {content: "Inventarnummer INVALID"}</style>
<script>const model = "wrong vehicle";</script><div>Modell &amp; Zubehör</div><p>Erfasste Daten</p>`
	for _, visible := range []string{visibleArticleText(html), visibleArticleLines(html)} {
		if strings.Contains(visible, "INVALID") || strings.Contains(visible, "wrong vehicle") ||
			!strings.Contains(visible, "Modell & Zubehör") || !strings.Contains(visible, "Erfasste Daten") {
			t.Fatal("article text included hidden content or lost visible source text", visible)
		}
	}
}
