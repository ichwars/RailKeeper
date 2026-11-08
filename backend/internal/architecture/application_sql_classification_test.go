package architecture_test

import (
	"go/parser"
	"go/token"
	"testing"
)

func TestSQLInventorySeparatesParsedURLsWithoutExemptingSQL(t *testing.T) {
	for _, test := range []struct {
		name, source string
		want         int
	}{
		{"url parameters", `package fixture;import "net/url";func f(){parsed,_:=url.Parse("https://example.test");parsed.Query()}`, 0},
		{"aliased URL import", `package fixture;import weburl "net/url";func f(){parsed,_:=weburl.Parse("https://example.test");parsed.Query()}`, 0},
		{"SQL alongside URL", `package fixture;import "net/url";func f(db DB){parsed,_:=url.Parse("https://example.test");parsed.Query();db.Query("SELECT 1")}`, 1},
		{"shadowed receiver", `package fixture;import "net/url";func f(db DB){parsed,_:=url.Parse("https://example.test");parsed.Query();{parsed:=db;parsed.Query("SELECT 1")}}`, 1},
		{"shadowed package name", `package fixture;import "net/url";func f(fake Factory){url:=fake;parsed,_:=url.Parse("value");parsed.Query("SELECT 1")}`, 1},
		{"unproven receiver", `package fixture;import "net/url";func f(db DB){parsed:=db;parsed.Query("SELECT 1")}`, 1},
		{"second result is not URL", `package fixture;import "net/url";func f(){_,parsed:=url.Parse("value");parsed.Query("SELECT 1")}`, 1},
		{"prepared statement without arguments", `package fixture;func f(db DB){statement,_:=db.Prepare("SELECT 1");statement.Query()}`, 2},
		{"SQL context and transaction", `package fixture;import "net/url";func f(db DB){parsed,_:=url.Parse("value");parsed.Query();db.QueryContext(ctx,"SELECT 1");db.BeginTx(ctx,nil)}`, 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			file, err := parser.ParseFile(token.NewFileSet(), "fixture.go", test.source, 0)
			if err != nil {
				t.Fatal(err)
			}
			if got := applicationSQLCalls(file); got != test.want {
				t.Fatalf("SQL calls %d, want %d", got, test.want)
			}
		})
	}
}
