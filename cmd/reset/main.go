package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type templateReset struct {
	Package string
	Struct  string
	Fields  []string
}

const templateStr = `
	func (s *{{.Struct}}) Reset() {
		if s == nil {
			return
		}

		{{range .Fields}}	
			{{.}}
		{{end}}
	}
`

var tmpl = template.Must(template.New("reset").Parse(templateStr))

type astPackage struct {
	Package string
	Dir string
	Files   []*fileInfo
}

func main() {
	currDir, _ := os.Getwd()
	packages := scanPackages(currDir)

	var astPackages []astPackage
	for _, pack := range packages {
		ast := astPackage{
			Files: make([]*fileInfo, 0, len(pack.GoFiles)),
			Dir: pack.Path,
		}
		for _, file := range pack.GoFiles {
			f, err := parseFile(file)
			if err == nil {
				ast.Package = f.Package
				ast.Files = append(ast.Files, f)
			}
		}

		astPackages = append(astPackages, ast)
	}

	var resetTargets []packageStructs
	for _, pack := range astPackages {
		target := findResetableStructs(pack)
		resetTargets = append(resetTargets, target)
	}

	for _, p := range resetTargets {
		t := analyzeStruct(p)
		writeGeneratedFile(p.Pkg.Dir, p.Pkg.Package, t)
	}
}

type packageInfo struct {
	Path    string
	Name    string
	GoFiles []string
}

func scanPackages(rootDir string) []packageInfo {
	var packages []packageInfo

	filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			return nil
		}

		if isServiceDir(path) {
			return nil
		}

		pack := packageInfo{
			Path:    path,
			Name:    info.Name(),
			GoFiles: make([]string, 0),
		}

		entries, err := os.ReadDir(path)
		if err != nil {
			return nil
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			if strings.HasSuffix(entry.Name(), ".go") {
				if strings.Contains(entry.Name(), "_test") || strings.Contains(entry.Name(), ".gen") {
					continue
				}

				pack.GoFiles = append(pack.GoFiles, filepath.Join(path, entry.Name()))
			}
		}

		if len(pack.GoFiles) > 0 {
			packages = append(packages, pack)
		}

		return nil
	})

	return packages
}

func isServiceDir(dirPath string) bool {
	base := filepath.Base(dirPath)

	if base == "profiles" || base == "data" || base == ".git" {
		return true
	}

	return false
}

type fileInfo struct {
	Path    string
	Package string
	AST     *ast.File
}

func parseFile(filePath string) (*fileInfo, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	return &fileInfo{
		Path:    filePath,
		Package: f.Name.Name,
		AST:     f,
	}, nil
}

type astStruct struct {
	Name   string
	fields *ast.FieldList
}

type packageStructs struct {
	Pkg     astPackage
	Structs []astStruct
}

func findResetableStructs(pack astPackage) packageStructs {
	var structs []astStruct

	for _, file := range pack.Files {
		for _, d := range file.AST.Decls {
			genDecl, ok := d.(*ast.GenDecl)
			if !ok {
				continue
			}

			if genDecl.Tok != token.TYPE {
				continue
			}

			genHasMarker := false
			if genDecl.Doc != nil {
				for _, c := range genDecl.Doc.List {
					if strings.Contains(c.Text, "generate:reset") {
						genHasMarker = true
						break
					}
				}
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}

				specHasMarker := false
				if typeSpec.Doc != nil {
					for _, c := range typeSpec.Doc.List {
						if strings.Contains(c.Text, "generate:reset") {
							specHasMarker = true
							break
						}
					}
				}

				if genHasMarker || specHasMarker {
					fmt.Println("found resetable struc:", typeSpec.Name.Name)

					structs = append(structs, astStruct{
						Name:   typeSpec.Name.Name,
						fields: structType.Fields,
					})
				}
			}
		}
	}

	return packageStructs{
		Pkg:     pack,
		Structs: structs,
	}
}

func analyzeStruct(target packageStructs) []templateReset {
	var templates []templateReset

	resetableTypes := make(map[string]struct{})
	for _, s := range target.Structs {
		resetableTypes[s.Name] = struct{}{}
	}

	for _, s := range target.Structs {
		tmpl := templateReset{
			Package: target.Pkg.Package,
			Struct:  s.Name,
			Fields:  make([]string, 0),
		}

		for _, f := range s.fields.List {
			tmpl.Fields = append(tmpl.Fields, resetField(f, resetableTypes)...)
		}

		templates = append(templates, tmpl)
	}
	return templates
}

func resetField(f *ast.Field, resetableTypes map[string]struct{}) []string {
	var resetStr []string

	for _, n := range f.Names {
		target := fmt.Sprintf("s.%s", n.Name)
		str := resetTarget(n.Name, target, f.Type, resetableTypes)
		resetStr = append(resetStr, str)
	}

	return resetStr
}

func resetTarget(name string, targetExpr string, typeExpr ast.Expr, resetableTypes map[string]struct{}) string {
	t := targetExpr
	switch x := typeExpr.(type) {
	case *ast.Ident:
		resetStr := resetPrimitive(x.Name)
		if resetStr != "" {
			return t + " = " + resetStr
		} else if _, ok := resetableTypes[x.Name]; ok {
			return "(&" + t + ").Reset()"
		} else {
			return fmt.Sprintf(`
				var z_%s %s
				%s = z_%s
			`, name, x.Name, t, name)
		}
	case *ast.ArrayType:
		if x.Len == nil {
			return fmt.Sprintf(`
			if %s != nil {
				%s = (%s)[:0]
			}`,
				t, t, t)
		}
	case *ast.MapType:
		return fmt.Sprintf("clear(%s)", t)
	case *ast.StarExpr:
		innerType := x.X

		if _, ok := innerType.(*ast.SelectorExpr); ok {
			return t + " = nil"
		}
		innerTarget := "*" + t
		innerStmt := resetTarget(name, innerTarget, innerType, resetableTypes)

		return "if " + t + " != nil {\n\t" + innerStmt + "\n}"
	}

	return ""
}

var numericTypes = map[string]struct{}{
	"int":        {},
	"int8":       {},
	"int16":      {},
	"int32":      {},
	"int64":      {},
	"uint":       {},
	"uint8":      {},
	"uint16":     {},
	"uint32":     {},
	"uint64":     {},
	"uintptr":    {},
	"byte":       {},
	"rune":       {},
	"float32":    {},
	"float64":    {},
	"complex64":  {},
	"complex128": {},
}

func resetPrimitive(t string) string {
	switch t {
	case "string":
		return "\"\""
	case "bool":
		return "false"
	default:
		if _, ok := numericTypes[t]; ok {
			return "0"
		}
	}

	return ""
}

func writeGeneratedFile(pkgDir string, pkgName string, tmpls []templateReset) {
	outPath := filepath.Join(pkgDir, "reset.gen.go")

	if len(tmpls) == 0 {
		return
	}

	title := "// Code generated by go generate; DO NOT EDIT.\n// This file was generated by cmd/reset/main.go"
	packageStr := "package " + pkgName
	space := "\n"

	var buff bytes.Buffer
	buff.Write([]byte(title))
	buff.Write([]byte(space))
	buff.Write([]byte(packageStr))
	buff.Write([]byte(space))

	for _, t := range tmpls {
		if err := tmpl.Execute(&buff, t); err != nil {
			fmt.Printf("reset: template execute failed (out=%s, pkg=%s, struct=%s): %v\n", outPath, pkgName, t.Struct, err)
			return
		}
	}
	
	formattedBytes, err := format.Source(buff.Bytes())
	if err != nil {
		fmt.Printf("reset: go/format failed (out=%s, pkg=%s): %v\n", outPath, pkgName, err)
		return
	}

	err = os.WriteFile(outPath, formattedBytes, 0644)


	if err != nil {
		fmt.Printf("reset: write file failed (out=%s, pkg=%s): %v\n", outPath, pkgName, err)
		return
	}

	fmt.Printf("reset: generated %s (%d types)\n", outPath, len(tmpls))
}	
