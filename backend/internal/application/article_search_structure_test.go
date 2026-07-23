package application

import (
	"os"
	"strings"
	"testing"
)

func TestArticleSearchCoreStaysFocused(t *testing.T) {
	data, err := os.ReadFile("article_search.go")
	if err != nil {
		t.Fatalf("read article_search.go: %v", err)
	}
	source := string(data)
	if lines := strings.Count(source, "\n") + 1; lines > 500 {
		t.Fatalf("article_search.go has %d lines, want at most 500", lines)
	}
	for _, helper := range []string{
		"func parseDuckDuckGoResults",
		"func scoreArticleResult",
		"func extractPDFText",
		"func bestArticleDescription",
	} {
		if strings.Contains(source, helper) {
			t.Errorf("article_search.go still contains helper %q", helper)
		}
	}
}
