package architecture_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestTransferWritersKeepOneTransactionOwner(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("..", "infrastructure", "data_transfer_apply*.go"))
	if err != nil {
		t.Fatal(err)
	}
	files = append(files, filepath.Join("..", "infrastructure", "data_transfer_vehicle_sets.go"))
	owners := map[string]int{}
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, decl := range file.Decls {
			function, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			ast.Inspect(function.Body, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				selector, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				if selector.Sel.Name == "BeginTx" || selector.Sel.Name == "Commit" {
					owners[selector.Sel.Name]++
					if function.Name.Name != "ApplyImportWithPolicy" {
						t.Errorf("%s starts or commits a nested transfer transaction", function.Name.Name)
					}
				}
				return true
			})
		}
	}
	if owners["BeginTx"] != 1 || owners["Commit"] != 1 {
		t.Fatal("import completion no longer has exactly one transaction owner", owners)
	}
}

func TestCompletedTransferFacadeHasNoConcreteIO(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("..", "application", "data_transfer*.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, item := range file.Imports {
			name, err := strconv.Unquote(item.Path.Value)
			if err != nil {
				t.Fatal(err)
			}
			if strings.HasPrefix(name, "database/") || strings.HasPrefix(name, "net/") ||
				name == "os" || strings.HasPrefix(name, "os/") || name == "io/fs" || name == "path/filepath" {
				t.Errorf("completed transfer facade %s imports concrete I/O: %s", path, name)
			}
		}
	}
}
