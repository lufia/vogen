package main

import (
	"flag"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"unicode"
)

type Field struct {
	Name string
	Type string
}

type DefinedType struct {
	Name     string
	BaseType string
}

type TemplateData struct {
	PackageName  string
	StructName   string
	Fields       []Field
	DefinedTypes []DefinedType
}

func main() {
	sourceFile := flag.String("source", "", "Source file name")
	structNames := flag.String("structs", "", "Comma-separated list of struct names to generate")
	flag.Parse()

	if *sourceFile == "" || *structNames == "" {
		log.Fatalf("Usage: go run gen.go -source <FileName> -structs <StructName1,StructName2,...>")
	}

	filename := *sourceFile
	targetStructs := strings.Split(*structNames, ",")

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		log.Fatalf("failed to parse file: %v", err)
	}

	typeMap := make(map[string]string)
	constructors := []TemplateData{}
	definedTypes := []DefinedType{}
	constructorReturnsError := make(map[string]bool)

	for _, f := range node.Decls {
		if genDecl, ok := f.(*ast.GenDecl); ok {
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					if ident, ok := typeSpec.Type.(*ast.Ident); ok {
						typeMap[typeSpec.Name.Name] = ident.Name
						definedTypes = append(definedTypes, DefinedType{Name: typeSpec.Name.Name, BaseType: ident.Name})
					}
					if starExpr, ok := typeSpec.Type.(*ast.StarExpr); ok {
						if ident, ok := starExpr.X.(*ast.Ident); ok {
							typeMap[typeSpec.Name.Name] = "*" + ident.Name
							definedTypes = append(definedTypes, DefinedType{Name: typeSpec.Name.Name, BaseType: ident.Name})
						}
					}
					if structType, ok := typeSpec.Type.(*ast.StructType); ok {
						for _, targetStruct := range targetStructs {
							if typeSpec.Name.Name == targetStruct {
								fields := []Field{}
								for _, field := range structType.Fields.List {
									fieldType := exprToString(field.Type)
									for _, name := range field.Names {
										fields = append(fields, Field{Name: name.Name, Type: fieldType})
									}
								}
								constructors = append(constructors, TemplateData{
									PackageName:  node.Name.Name,
									StructName:   typeSpec.Name.Name,
									Fields:       fields,
									DefinedTypes: definedTypes,
								})
							}
						}
					}
				}
			}
		}
	}

	for _, f := range node.Decls {
		if funcDecl, ok := f.(*ast.FuncDecl); ok {
			if funcDecl.Recv == nil && strings.HasPrefix(funcDecl.Name.Name, "New") {
				if funcDecl.Type.Results != nil && len(funcDecl.Type.Results.List) > 1 {
					if ident, ok := funcDecl.Type.Results.List[1].Type.(*ast.Ident); ok && ident.Name == "error" {
						constructorReturnsError[funcDecl.Name.Name] = true
					}
				}
			}
		}
	}

	for _, constructor := range constructors {
		generateConstructor(filename, constructor, typeMap, constructorReturnsError)
	}
}

func exprToString(expr ast.Expr) string {
	switch v := expr.(type) {
	case *ast.Ident:
		return v.Name
	case *ast.SelectorExpr:
		return exprToString(v.X) + "." + v.Sel.Name
	case *ast.StarExpr:
		return "*" + exprToString(v.X)
	case *ast.ArrayType:
		return "[]" + exprToString(v.Elt)
	default:
		return ""
	}
}

func toCamelCase(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

func generateConstructor(filename string, data TemplateData, typeMap map[string]string, constructorReturnsError map[string]bool) {
	tmpl, err := template.New("constructor").Funcs(template.FuncMap{
		"camelCase": toCamelCase,
		"isDefinedType": func(typ string) bool {
			_, defined := typeMap[typ]
			return defined
		},
		"getBaseType": func(typ string) string {
			baseType, defined := typeMap[typ]
			if defined {
				return baseType
			}
			return typ
		},
		"getRootBaseType": func(typ string) string {
			/*
				type A int
				type B A
				 のように定義されている場合、BのRootBaseTypeはintになる
			*/
			for {
				isPtr := strings.HasPrefix(typ, "*")
				if isPtr {
					notPtrTyp := strings.TrimPrefix(typ, "*")
					baseType, defined := typeMap[notPtrTyp]
					if defined {
						return "*" + baseType
					}
				}
				baseType, defined := typeMap[typ]
				if !defined {
					break
				}
				typ = baseType
			}
			return typ
		},
		"shouldReturnError": func() bool {
			return len(constructorReturnsError) > 0
		},
		"constructorReturnsError": func(typ string) bool {
			return constructorReturnsError["New"+typ]
		},
		"isPointer": func(typ string) bool {
			return strings.HasPrefix(typ, "*")
		},
		"rawValueChain": func(name string, typ string) string {
			chain := "d." + name
			isPointer := strings.HasPrefix(typ, "*")
			if isPointer {
				chain = "d." + name + " != nil && *d." + name
			}
			for {
				baseType, defined := typeMap[typ]
				if !defined {
					break
				}
				chain += ".RawValue()"
				typ = baseType
			}
			return chain
		},
	}).Parse(`// Code generated by vogen DO NOT EDIT.
package {{.PackageName}}
func New{{.StructName}}({{range $index, $field := .Fields}}{{if $index}}, {{end}}{{$field.Name | camelCase}} {{if isDefinedType $field.Type}}{{getRootBaseType $field.Type}}{{else}}{{$field.Type}}{{end}}{{end}}) {{if shouldReturnError}}(*{{.StructName}}, error){{else}}*{{.StructName}}{{end}} {
 {{range $index, $field := .Fields}}
   {{if isDefinedType .Type}}
     {{if constructorReturnsError .Type}}
       tempVarByVogen{{$index}}, err := New{{.Type}}({{.Name | camelCase}})
       if err != nil {
        return nil, err
       }
     {{else}}
       tempVarByVogen{{$index}} := New{{.Type}}({{.Name | camelCase}})
     {{end}}
  {{end}}
{{end}}
{{if shouldReturnError}}
  return &{{.StructName}}{
   {{range $index, $field := .Fields}}
     {{.Name}}: {{if isDefinedType .Type}}tempVarByVogen{{$index}}{{else}}{{.Name | camelCase}}{{end}},
   {{end}}
  }, nil
{{else}}
  return &{{.StructName}}{
   {{range $index, $field := .Fields}}
     {{.Name}}: {{if isDefinedType .Type}}tempVarByVogen{{$index}}{{else}}{{.Name | camelCase}}{{end}},
   {{end}}
  }
{{end}}
}

type raw{{.StructName}} struct {
{{range .Fields}}
  {{$rootBaseType := getRootBaseType .Type}}
  {{.Name}} {{$rootBaseType}}
{{end}}
}

func (d {{.StructName}}) RawValue() raw{{.StructName}} {
{{range .Fields}}
  {{if isDefinedType .Type}}
    {{$baseType := getBaseType .Type}}
    {{$rootBaseType := getRootBaseType .Type}}
    {{if isPointer $baseType}}
      var tempVarByVogen{{.Name}} {{$rootBaseType}}
  	if d.{{.Name}} != nil {
  		tempVarByVogen{{.Name}} = ({{$rootBaseType}})(({{$baseType}})(d.{{.Name}}))
  	}
    {{end}}
  {{end}}
{{end}}

 return raw{{.StructName}}{
  {{range .Fields}}
    {{$rootBaseType := getRootBaseType .Type}}
    {{if isPointer $rootBaseType}}
      {{.Name}}: tempVarByVogen{{.Name}},
    {{else}}
      {{.Name}}: {{rawValueChain .Name .Type}},
    {{end}}
  {{end}}
 }
}

{{range .DefinedTypes}}
  {{$rootBaseType := getBaseType .Name}}
  {{if not (isPointer $rootBaseType)}}
    func (d {{.Name}}) RawValue() {{$rootBaseType}} {
        return {{$rootBaseType}}(d)
    }
  {{end}}
{{end}}
`)
	if err != nil {
		log.Fatalf("failed to parse template: %v", err)
	}

	baseFileName := strings.TrimSuffix(filename, ".go")
	outputFilename := baseFileName + "_vo_gen.go"
	f, err := os.Create(outputFilename)
	if err != nil {
		log.Fatalf("failed to create file: %v", err)
	}
	defer f.Close()

	err = tmpl.Execute(f, data)
	if err != nil {
		log.Fatalf("failed to execute template: %v", err)
	}

	runGoImports(outputFilename)
	runGoFmt(outputFilename)

	log.Printf("Constructor for %s generated successfully in %s\n", data.StructName, outputFilename)
}

func runGoImports(filename string) {
	cmd := exec.Command("go", "run", "golang.org/x/tools/cmd/goimports", "-w", filename)
	err := cmd.Run()
	if err != nil {
		log.Fatalf("failed to run goimports: %v", err)
	}
}

func runGoFmt(filename string) {
	cmd := exec.Command("go", "run", "cmd/gofmt", "-w", filename)
	err := cmd.Run()
	if err != nil {
		log.Fatalf("failed to run gofmt: %v", err)
	}
}
