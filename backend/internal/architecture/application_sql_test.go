package architecture_test

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

type applicationSQLBoundary struct {
	File         string   `json:"file"`
	Calls        int      `json:"calls"`
	Stage        string   `json:"stage"`
	Owner        string   `json:"owner"`
	Transactions []string `json:"transactions,omitempty"`
}

// The inventory is a shrinking migration contract, not an exemption for new SQL.
func TestApplicationSQLMigrationCannotRegress(t *testing.T) {
	data, err := os.ReadFile("application_sql_baseline.json")
	if err != nil {
		t.Fatal(err)
	}
	var entries []applicationSQLBoundary
	if err := json.Unmarshal(data, &entries); err != nil {
		t.Fatal(err)
	}
	baseline := map[string]applicationSQLBoundary{}
	for _, entry := range entries {
		if _, duplicate := baseline[entry.File]; duplicate || entry.Calls < 1 || entry.Stage == "" || entry.Owner == "" {
			t.Fatalf("invalid or duplicate SQL migration entry: %q", entry.File)
		}
		baseline[entry.File] = entry
	}
	root := filepath.Join("..", "application")
	if err := filepath.WalkDir(root, func(path string, item fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if item.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		calls := applicationSQLCalls(file)
		name, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		name = filepath.ToSlash(name)
		entry, listed := baseline[name]
		delete(baseline, name)
		if calls > 0 && !listed {
			t.Errorf("new concrete database execution in application/%s; use the owning persistence port", name)
		} else if listed && calls != entry.Calls {
			t.Errorf("application/%s has %d database calls, baseline %d: migrated calls must shrink the inventory; new calls belong in %s", name, calls, entry.Calls, entry.Owner)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	for name := range baseline {
		t.Errorf("remove completed SQL migration entry %s", name)
	}
}

func applicationSQLCalls(file *ast.File) int {
	calls := 0
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		switch selector.Sel.Name {
		case "Exec", "ExecContext", "Query", "QueryContext", "QueryRow", "QueryRowContext", "Prepare", "PrepareContext", "Begin", "BeginTx":
			if selector.Sel.Name != "Query" || !isParsedURLQuery(file, selector) {
				calls++
			}
		}
		return true
	})
	return calls
}

// A direct short declaration from net/url.Parse infers *url.URL. Its Query method
// reads URL parameters. Object identity keeps SQL calls on shadowed names counted.
func isParsedURLQuery(file *ast.File, selector *ast.SelectorExpr) bool {
	receiver, ok := selector.X.(*ast.Ident)
	if !ok || receiver.Obj == nil {
		return false
	}
	assignment, ok := receiver.Obj.Decl.(*ast.AssignStmt)
	if !ok || assignment.Tok != token.DEFINE || len(assignment.Rhs) != 1 || len(assignment.Lhs) == 0 {
		return false
	}
	first, ok := assignment.Lhs[0].(*ast.Ident)
	if !ok || first.Obj != receiver.Obj {
		return false
	}
	parsed, ok := assignment.Rhs[0].(*ast.CallExpr)
	if !ok {
		return false
	}
	parse, ok := parsed.Fun.(*ast.SelectorExpr)
	if !ok || parse.Sel.Name != "Parse" {
		return false
	}
	imported, ok := parse.X.(*ast.Ident)
	if !ok || imported.Obj != nil {
		return false
	}
	for _, entry := range file.Imports {
		path, err := strconv.Unquote(entry.Path.Value)
		if err != nil || path != "net/url" {
			continue
		}
		name := "url"
		if entry.Name != nil {
			name = entry.Name.Name
		}
		if name == imported.Name {
			return true
		}
	}
	return false
}
