package architecture_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

func TestMasterDataMutationsKeepTheirSingleTransactionOwner(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("..", "infrastructure", "masterdatastore", "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"UpdateCustomField": true, "ImportReconciled": true, "SetActiveMany": true, "Delete": true}
	owners := map[string]map[string]int{}
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, item := range file.Decls {
			function, ok := item.(*ast.FuncDecl)
			if !ok {
				continue
			}
			ast.Inspect(function.Body, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				selector, ok := call.Fun.(*ast.SelectorExpr)
				if !ok || selector.Sel.Name != "BeginTx" && selector.Sel.Name != "Commit" {
					return true
				}
				if !want[function.Name.Name] {
					t.Errorf("unexpected nested master-data transaction in %s", function.Name.Name)
				}
				if owners[function.Name.Name] == nil {
					owners[function.Name.Name] = map[string]int{}
				}
				owners[function.Name.Name][selector.Sel.Name]++
				return true
			})
		}
	}
	for name := range want {
		if owners[name]["BeginTx"] != 1 || owners[name]["Commit"] != 1 {
			t.Error("master-data mutation lost its single transaction owner", name, owners[name])
		}
	}
}
