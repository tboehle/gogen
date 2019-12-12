package importer

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/tboehle/gogen/gogenutil"
)

type customImporter struct {
	imported      map[string]*types.Package
	base          types.Importer
	skipTestFiles bool
}

type gomoduleImporter struct {
	imported      map[string]*types.Package
	base          types.Importer
	skipTestFiles bool
}

func (i *customImporter) Import(path string) (*types.Package, error) {
	var err error
	if path == "" || path[0] == '.' {
		path, err = filepath.Abs(filepath.Clean(path))
		if err != nil {
			return nil, err
		}
		path = gogenutil.StripGopath(path)
	}
	if pkg, ok := i.imported[path]; ok {
		return pkg, nil
	}
	pkg, err := i.fsPkg(path)
	if err != nil {
		return nil, err
	}
	i.imported[path] = pkg
	return pkg, nil
}

func (i *gomoduleImporter) Import(path string) (*types.Package, error) {
	var err error
	if path == "" || path[0] == '.' {
		path, err = filepath.Abs(filepath.Clean(path))
		if err != nil {
			return nil, err
		}
		path = gogenutil.StripGopath(path)
	}
	if pkg, ok := i.imported[path]; ok {
		return pkg, nil
	}
	pkg, err := i.fsPkg(path)
	if err != nil {
		return nil, err
	}
	i.imported[path] = pkg
	return pkg, nil
}

func gopathDir(pkg string) (string, error) {
	for _, gopath := range strings.Split(os.Getenv("GOPATH"), ":") {
		absPath, err := filepath.Abs(path.Join(gopath, "src", pkg))
		if err != nil {
			return "", err
		}
		if dir, err := os.Stat(absPath); err == nil && dir.IsDir() {
			return absPath, nil
		}
	}
	return "", fmt.Errorf("%s not in $GOPATH", pkg)
}

func removeGopath(p string) string {
	for _, gopath := range strings.Split(os.Getenv("GOPATH"), ":") {
		p = strings.Replace(p, path.Join(gopath, "src")+"/", "", 1)
	}
	return p
}

func (i *customImporter) fsPkg(pkg string) (*types.Package, error) {
	dir, err := gopathDir(pkg)
	if err != nil {
		return importOrErr(i.base, pkg, err)
	}

	dirFiles, err := ioutil.ReadDir(dir)
	if err != nil {
		return importOrErr(i.base, pkg, err)
	}

	fset := token.NewFileSet()
	var files []*ast.File
	for _, fileInfo := range dirFiles {
		if fileInfo.IsDir() {
			continue
		}
		n := fileInfo.Name()
		if path.Ext(fileInfo.Name()) != ".go" {
			continue
		}
		if i.skipTestFiles && strings.Contains(fileInfo.Name(), "_test.go") {
			continue
		}
		file := path.Join(dir, n)
		src, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}
		f, err := parser.ParseFile(fset, file, src, 0)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	conf := types.Config{
		Importer: i,
	}
	p, err := conf.Check(pkg, fset, files, nil)

	if err != nil {
		return importOrErr(i.base, pkg, err)
	}
	return p, nil
}

func workspaceDir() (string, error) {
	workspace, dirErr := os.Getwd()
	if dirErr != nil {
		return "", fmt.Errorf("Error while getting workspace")
	}
	if dir, err := os.Stat(workspace); err == nil && dir.IsDir() {
		// Prrove if Backslash or Slash
		return workspace, nil
	}
	return "", fmt.Errorf("Workspace is no directory or is not existing")
}

func (i *gomoduleImporter) fsPkg(pkg string) (*types.Package, error) {
	var dir, opDir string
	dir = pkg
	opDir, err := workspaceDir()
	if err != nil {
		return importOrErr(i.base, pkg, err)
	}
	// If pkg not equals opDir, it's a module from the gopath
	if pkg != opDir {
		dir, err = gopathDir(pkg)
		if err != nil {
			return importOrErr(i.base, pkg, err)
		}
	}

	dirFiles, err := ioutil.ReadDir(dir)
	if err != nil {
		return importOrErr(i.base, pkg, err)
	}

	fset := token.NewFileSet()
	var files []*ast.File
	for _, fileInfo := range dirFiles {
		if fileInfo.IsDir() {
			continue
		}
		n := fileInfo.Name()
		if path.Ext(fileInfo.Name()) != ".go" {
			continue
		}
		if i.skipTestFiles && strings.Contains(fileInfo.Name(), "_test.go") {
			continue
		}
		file := path.Join(dir, n)
		src, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}
		f, err := parser.ParseFile(fset, file, src, 0)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	conf := types.Config{
		Importer: i,
	}
	p, err := conf.Check(pkg, fset, files, nil)

	if err != nil {
		return importOrErr(i.base, pkg, err)
	}
	return p, nil
}

func importOrErr(base types.Importer, pkg string, err error) (*types.Package, error) {
	p, impErr := base.Import(pkg)
	if impErr != nil {
		return nil, err
	}
	return p, nil
}

// Default returns an importer that will try to import code from gopath before using go/importer.Default and skipping test files
func Default() types.Importer {
	return &customImporter{
		imported:      make(map[string]*types.Package),
		base:          importer.Default(),
		skipTestFiles: true,
	}
}

// Default returns an importer that will try to import go modules from the gopath and the actual package using go/importer.Default and skipping test files
func Custom() types.Importer {
	return &gomoduleImporter{
		imported:      make(map[string]*types.Package),
		base:          importer.Default(),
		skipTestFiles: true,
	}
}

// Default returns an importer that will try to import go modules from the gopath and the actual package using go/importer.Default
func CustomWithTestFile() types.Importer {
	return &gomoduleImporter{
		imported:      make(map[string]*types.Package),
		base:          importer.Default(),
		skipTestFiles: false,
	}
}

// DefaultWithTestFiles same as Default but it parses test files too
func DefaultWithTestFiles() types.Importer {
	return &customImporter{
		imported:      make(map[string]*types.Package),
		base:          importer.Default(),
		skipTestFiles: false,
	}
}
