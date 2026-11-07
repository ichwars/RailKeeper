package architecture_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestIndependentAuthenticationModuleBoundaries(t *testing.T) {
	const prefix = "railkeeper/backend/internal/"
	modules := map[string][]string{
		"authentication":               {"credentials"},
		"credentials":                  {},
		"valuation":                    {},
		"transfer":                     {"domain"},
		"infrastructure/valuationdata": {"valuation", "credentials"},
		"infrastructure/authdata":      {"authentication", "credentials", "infrastructure/ratelimit"},
		"infrastructure/ratelimit":     {"credentials"},
	}
	for module, allowed := range modules {
		t.Run(module, func(t *testing.T) {
			err := filepath.WalkDir(filepath.Join("..", filepath.FromSlash(module)), func(path string, entry fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
					return nil
				}
				file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
				if err != nil {
					return err
				}
				for _, imported := range file.Imports {
					name, err := strconv.Unquote(imported.Path.Value)
					if err != nil {
						return err
					}
					if strings.HasPrefix(name, prefix) {
						valid := false
						for _, target := range allowed {
							valid = valid || name == prefix+target
						}
						if !valid {
							t.Errorf("%s imports forbidden module %s", path, name)
						}
					}
					if module == "authentication" || module == "credentials" || module == "valuation" || module == "transfer" {
						if strings.HasPrefix(name, "database/") || name == "os" || strings.HasPrefix(name, "os/") || name == "io/fs" || name == "path/filepath" ||
							strings.HasPrefix(name, "net/") || strings.HasPrefix(name, "modernc.org/") {
							t.Errorf("%s imports a transport or persistence dependency: %s", path, name)
						}
					}
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
