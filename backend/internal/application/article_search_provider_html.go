package application

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

func firstFiveDigits(value string) string {
	digits := normalizedArticleNumber(value)
	if len(digits) < 5 {
		return ""
	}
	return digits[:5]
}

func pikoSparePartsFromHTML(body, pageURL string) []ArticleSearchSparePart {
	seen := map[string]bool{}
	parts := []ArticleSearchSparePart{}
	for index, block := range strings.Split(body, `<div class="artikel_ersatzteil__list_item">`) {
		if index == 0 {
			continue
		}
		block = `<div class="artikel_ersatzteil__list_item">` + block
		link := ""
		if match := linkHrefAttrPattern.FindStringSubmatch(block); len(match) >= 2 {
			link = resolveURL(pageURL, html.UnescapeString(match[1]))
		}
		description := ""
		if match := pikoSparePartTitlePattern.FindStringSubmatch(block); len(match) >= 2 {
			description = cleanHTML(match[1])
		}
		articleNumber := ""
		if match := pikoSparePartNumberPattern.FindStringSubmatch(block); len(match) >= 2 {
			articleNumber = strings.TrimSpace(html.UnescapeString(match[1]))
		}
		price := ""
		if match := pikoSparePartPriceLoosePattern.FindStringSubmatch(block); len(match) >= 2 {
			price = strings.ReplaceAll(strings.TrimSpace(match[1]), ",", ".")
		}
		availability := ""
		if match := pikoSparePartAvailabilityPattern.FindStringSubmatch(block); len(match) >= 2 {
			availability = cleanHTML(match[1])
		}
		if articleNumber == "" || (price == "" && link == "") {
			continue
		}
		key := strings.ToLower(articleNumber + "|" + link)
		if seen[key] {
			continue
		}
		seen[key] = true
		parts = append(parts, ArticleSearchSparePart{
			ArticleNumber: articleNumber,
			Description:   description,
			Price:         price,
			URL:           link,
			Source:        "PIKO",
			Availability:  availability,
		})
		if len(parts) >= 120 {
			break
		}
	}
	return parts
}

func rocoSparePartsFromHTML(body, pageURL string) []ArticleSearchSparePart {
	seen := map[string]bool{}
	parts := []ArticleSearchSparePart{}
	for index, block := range strings.Split(body, `<div class="row table-row-et">`) {
		if index == 0 {
			continue
		}
		block = `<div class="row table-row-et">` + block
		articleNumber := ""
		if match := rocoSparePartNumberPattern.FindStringSubmatch(block); len(match) >= 2 {
			articleNumber = cleanHTML(match[1])
		}
		description := ""
		if match := rocoSparePartDescriptionPattern.FindStringSubmatch(block); len(match) >= 2 {
			description = cleanHTML(match[1])
		}
		price := ""
		if match := rocoSparePartPriceLoosePattern.FindStringSubmatch(block); len(match) >= 2 {
			price = strings.ReplaceAll(strings.TrimSpace(match[1]), ",", ".")
		}
		availability := ""
		if match := rocoSparePartAvailabilityPattern.FindStringSubmatch(block); len(match) >= 2 {
			availability = cleanHTML(match[1])
		}
		if articleNumber == "" || (price == "" && availability == "") {
			continue
		}
		link := rocoSparePartURL(pageURL, articleNumber)
		key := strings.ToLower(articleNumber + "|" + link)
		if seen[key] {
			continue
		}
		seen[key] = true
		parts = append(parts, ArticleSearchSparePart{
			ArticleNumber: articleNumber,
			Description:   description,
			Price:         price,
			URL:           link,
			Source:        "ROCO",
			Availability:  availability,
		})
		if len(parts) >= 80 {
			break
		}
	}
	return parts
}

func rocoSparePartURL(pageURL, articleNumber string) string {
	parsed, err := url.Parse(pageURL)
	if err != nil {
		return pageURL
	}
	parsed.RawQuery = url.Values{"et": []string{articleNumber}}.Encode()
	return parsed.String()
}

func focusedArticleSearchQuery(input ArticleSearchInput) string {
	parts := []string{}
	for _, value := range []string{input.ArticleNumber, input.Manufacturer, input.Gauge} {
		if value != "" {
			parts = append(parts, value)
		}
	}
	return strings.Join(uniqueNonEmpty(parts), " ")
}

func (a *DuckDuckGoArticleSearchAdapter) searchDuckDuckGo(ctx context.Context, input ArticleSearchInput, query string, source string) ([]ArticleSearchResult, error) {
	requestURL := duckDuckGoSearchURL(query)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build article search request: %w", err)
	}
	req.Header.Set("User-Agent", "RailKeeper/0.1 article-search")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("Accept-Language", "de-DE,de;q=0.9,en;q=0.5")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("article search request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("article search returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("read article search response: %w", err)
	}
	results := parseDuckDuckGoResults(string(body), input, source)
	return results, nil
}

func duckDuckGoSearchURL(query string) string {
	values := url.Values{}
	values.Set("q", query)
	values.Set("kl", "de-de")
	values.Set("kad", "de_DE")
	return "https://duckduckgo.com/html/?" + values.Encode()
}

var (
	resultBlockPattern               = regexp.MustCompile(`(?s)<div class="result results_links.*?</div>\s*</div>`)
	resultLinkPattern                = regexp.MustCompile(`(?s)<a[^>]+class="result__a"[^>]+href="([^"]+)"[^>]*>(.*?)</a>`)
	snippetPattern                   = regexp.MustCompile(`(?s)<a[^>]+class="result__snippet"[^>]*>(.*?)</a>|<div[^>]+class="result__snippet"[^>]*>(.*?)</div>`)
	tagPattern                       = regexp.MustCompile(`(?s)<[^>]+>`)
	scriptStylePattern               = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>|<style[^>]*>.*?</style>|<noscript[^>]*>.*?</noscript>|<svg[^>]*>.*?</svg>`)
	pricePattern                     = regexp.MustCompile(`(?i)(?:hersteller[-\s]*preis|preis|uvp)?[^\d]{0,20}(\d{1,4}(?:[,.]\d{2})?)\s?(?:eur|euro|\x{20AC})`)
	lengthPattern                    = regexp.MustCompile(`(?i)(?:l[äa]nge|laenge|length|ma[ßs]|mass|lüp|luep|luep\.)[^\d]{0,30}(\d{2,4}(?:[,.]\d+)?)\s?(?:mm)?`)
	weightPattern                    = regexp.MustCompile(`(?i)(?:gewicht|weight)[^\d]{0,18}(\d{1,5}(?:[,.]\d+)?)\s?g`)
	tractionTirePattern              = regexp.MustCompile(`(?i)(?:haftreifen|traction\s*tire)[^\d]{0,18}(\d{1,2})`)
	eanPattern                       = regexp.MustCompile(`\b(\d{12,14})\b`)
	epochPattern                     = regexp.MustCompile(`(?i)(?:epoche|epoch|ep\.)[^IVX]{0,16}(I{1,3}|IV|V|VI)\b`)
	railwayPattern                   = regexp.MustCompile(`\b(DB AG|DB|DRG|DR|SBB|\x{00D6}BB|OeBB|BLS|SNCF|NS|FS)\b`)
	adapterPattern                   = regexp.MustCompile(`(?i)\b(NEM\s?651|NEM\s?652|NEM\s?658|PluX\s?16|PluX\s?22|MTC\s?21|Next\s?18|8-?polig|21-?polig|DSS\s?8pol|elektrische\s+schnittstelle)\b`)
	powerPattern                     = regexp.MustCompile(`(?i)\b(DC|AC|2-?Leiter|3-?Leiter|Gleichstrom|Wechselstrom)\b`)
	digitalPositivePattern           = regexp.MustCompile(`(?i)(?:\bdigital\s*[:=]\s*(?:ja|yes|true)\b|\bdigitaldecoder\b|\bsounddecoder\b|\bmit\s+(?:dcc\s+)?decoder\b)`)
	headlightDescriptionPattern      = regexp.MustCompile(`(?i)(?:lichtwechsel|fahrlicht|spitzenlicht|spitzenbeleuchtung|schlusslicht)[^\n:;]{0,35}[:]\s*([^.;\n]{3,220})`)
	lightingDescriptionPattern       = regexp.MustCompile(`(?i)(?:innenbeleuchtung|fuehrerstandsbeleuchtung|fuehrerstand|kabinenbeleuchtung|beleuchtung)[^\n:;]{0,35}[:]\s*([^.;\n]{3,180})`)
	soundDescriptionPattern          = regexp.MustCompile(`(?i)(?:soundgenerator|sounddecoder|\bsound\b|sound\s+laut\s+artikeldaten|geräuschmodul|geraeuschmodul|ger..uschmodul)[^\n:;]{0,35}[:]\s*([^.;\n]{3,180})`)
	imageMetaPattern                 = regexp.MustCompile(`(?is)<meta[^>]+(?:property|name)=["'](?:og:image|twitter:image|thumbnail)["'][^>]+content=["']([^"']+)["']`)
	imageMetaAltPattern              = regexp.MustCompile(`(?is)<meta[^>]+content=["']([^"']+)["'][^>]+(?:property|name)=["'](?:og:image|twitter:image|thumbnail)["']`)
	imageTagPattern                  = regexp.MustCompile(`(?is)<img\b[^>]*>`)
	linkTagPattern                   = regexp.MustCompile(`(?is)<a\b[^>]*>.*?</a>`)
	linkHrefAttrPattern              = regexp.MustCompile(`(?is)\bhref=["']([^"']+)["']`)
	rowLikePattern                   = regexp.MustCompile(`(?is)<(?:tr|li)\b[^>]*>.*?</(?:tr|li)>`)
	sparePartArticlePattern          = regexp.MustCompile(`(?i)\b([A-Z]?\d{4,8}(?:[-/][A-Z0-9]+)?)\b`)
	pikoSparePartTitlePattern        = regexp.MustCompile(`(?is)<h3[^>]*>(.*?)</h3>`)
	pikoSparePartNumberPattern       = regexp.MustCompile(`(?is)Artikelnummer:\s*([^<\s]+)`)
	pikoSparePartAvailabilityPattern = regexp.MustCompile(`(?is)<span[^>]+(?:availability|lieferstatus)[^>]*>\s*(.*?)\s*</span>`)
	rocoSparePartNumberPattern       = regexp.MustCompile(`(?is)<div[^>]+class="[^"]*\bart-nr\b[^"]*"[^>]*>\s*([^<]+?)\s*</div>`)
	rocoSparePartDescriptionPattern  = regexp.MustCompile(`(?is)<div[^>]+class="[^"]*\bart-bz\b[^"]*"[^>]*>\s*(.*?)\s*</div>`)
	rocoSparePartAvailabilityPattern = regexp.MustCompile(`(?is)<img[^>]+class="[^"]*\bprodukt-head-verfuegbarkeit\b[^"]*"[^>]+title=["']([^"']+)["']`)
	pikoSparePartPriceLoosePattern   = regexp.MustCompile(`(?is)<div class="artikel_ersatzteil__price">\s*(\d{1,4}(?:[,.]\d{2})?)\s*[^<]{0,16}</div>`)
	rocoSparePartPriceLoosePattern   = regexp.MustCompile(`(?is)<div[^>]+class="[^"]*\bart-pr\b[^"]*"[^>]*>\s*(\d{1,4}(?:[,.]\d{2})?)\s*[^<]{0,16}</div>`)
	imageURLAttrPattern              = regexp.MustCompile(`(?is)\b(?:src|data-src|data-original|data-lazy-src|data-zoom-image)=["']([^"']+)["']`)
	imageSrcSetAttrPattern           = regexp.MustCompile(`(?is)\b(?:srcset|data-srcset)=["']([^"']+)["']`)
	metaDescriptionRegex             = regexp.MustCompile(`(?is)<meta[^>]+(?:name|property)=["'](?:description|og:description)["'][^>]+content=["']([^"']+)["']`)
)

var manufacturerDomains = map[string][]string{
	"arnold":      {"hornby.com"},
	"brawa":       {"brawa.de"},
	"esu":         {"esu.eu"},
	"fleischmann": {"fleischmann.de"},
	"lgb":         {"lgb.de", "maerklin.de"},
	"maerklin":    {"maerklin.de"},
	"piko":        {"piko.de", "piko-shop.de"},
	"roco":        {"roco.cc"},
	"tillig":      {"tillig.com"},
	"trix":        {"trix.de", "maerklin.de"},
	"viessmann":   {"viessmann-modell.com"},
}

var catalogArticleDomains = []string{
	"modellbahn-fokus.de",
}

var dealerArticleDomains = []string{
	"elriwa.de",
	"modellbahnshop-lippe.com",
	"dm-toys.de",
	"haertle.de",
}

var catalogDetailLabels = []string{
	"Hersteller",
	"Art.-Nr.",
	"Artikel-Nr.",
	"Artikelnummer",
	"EAN",
	"Spur",
	"Bahn-Gesellschaft",
	"Bahngesellschaft",
	"Epoche",
	"Stromsystem",
	"Digital-Decoder",
	"Schnittstelle",
	"Motor",
	"Schwungmasse",
	"Haftreifen",
	"L?nge ?ber Puffer",
	"Laenge ueber Puffer",
	"Mindestradius",
	"Spitzenlicht",
	"Vorbild (Land)",
	"Hersteller-Preis",
	"Herstellerpreis",
	"Preis",
	"UVP",
	"Produktlinie",
	"Erscheinungsdatum",
}
