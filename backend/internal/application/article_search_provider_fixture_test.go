package application

import (
	"strings"
	"testing"
)

func searchResultBlock(link, title, snippet string) string {
	return `<div class="result results_links"><a class="result__a" href="` + link + `">` + title +
		`</a><div class="result__snippet">` + snippet + `</div></div>`
}

func TestProviderFixtureKeepsDecodedSourceAndSeparatesSearchRankFromFieldConfidence(t *testing.T) {
	direct := "https://www.piko-shop.de/de/artikel/12345.html"
	body := searchResultBlock("//duckduckgo.com/l/?uddg=https%3A%2F%2Fwww.piko-shop.de%2Fde%2Fartikel%2F12345.html&amp;rut=opaque",
		`<b>PIKO</b> 12345 TT &amp; Modell`, "Neuheit PIKO 12345 TT mit einer vorbildgerechten Bedruckung") +
		searchResultBlock(direct, `PIKO 12345 TT &amp; Modell`, "Neuheit PIKO 12345 TT mit einer vorbildgerechten Bedruckung") +
		searchResultBlock(direct, "", "empty title must not become a suggestion") +
		`<div class="result results_links"><div>missing link</div></div>`
	rows := parseDuckDuckGoResults(body, ArticleSearchInput{Manufacturer: "PIKO", ArticleNumber: "12345", Gauge: "TT"}, "fixture")
	if len(rows) != 2 {
		t.Fatal("invalid search block accepted", rows)
	}
	for _, row := range rows {
		if row.URL != direct || row.Title != "PIKO 12345 TT & Modell" || row.Source != "fixture" ||
			row.Fields["articleSourceUrl"].Value != direct || row.Fields["articleSourceUrl"].Confidence != 100 ||
			row.Fields["articleNumber"].Value != "12345" {
			t.Fatal("source provenance lost", row)
		}
	}
	if rows[0].Score-rows[1].Score != 6 || rows[0].Fields["articleNumber"] != rows[1].Fields["articleNumber"] {
		t.Fatal("provider rank changed field evidence", rows)
	}
	if got := len(parseDuckDuckGoResults(strings.Repeat(searchResultBlock(direct, "PIKO 12345", "recorded snippet"), 20), ArticleSearchInput{}, "fixture")); got != 12 {
		t.Fatal("provider result bound lost", got)
	}
}

func TestPDFHexFixturePreservesUnicodeCVLabelsAndRejectsMalformedStrings(t *testing.T) {
	for _, sample := range []struct {
		raw, want string
		end       int
	}{
		{"<FEFF00430056002000330031>", "CV 31", 26},
		{"<44 E4 63 68 65 72>", "Dächer", 19},
		{"<GG>", "", 4}, {"<0>", "", 3}, {"<>", "", 2}, {"<434V", "", 0},
	} {
		got, end := readPDFHex(sample.raw, 0)
		if got != sample.want || end != sample.end {
			t.Fatalf("%q: %q at %d, want %q at %d", sample.raw, got, end, sample.want, sample.end)
		}
	}
}
