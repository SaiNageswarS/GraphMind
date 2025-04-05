package buildcodegraph

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// BuildAstControlFlow parses all Go files in the repo to build control flow files per service.
func (a *Activities) BuildAstControlFlow(ctx context.Context, state BuildCodeGraphState) (BuildCodeGraphState, error) {
	tmpDir, err := os.MkdirTemp("", "controlFlow-*")
	if err != nil {
		return state, fmt.Errorf("failed to create temp dir: %w", err)
	}

	// Create a new file set for all files in the repo.
	fset := token.NewFileSet()

	// Walk through the repository and parse all .go files.
	var astFiles []*ast.File
	err = filepath.WalkDir(state.LocalRepoPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Skip directories and non-Go files.
		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		// Parse the file.
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			fmt.Printf("Failed to parse %s: %v\n", path, err)
			return nil // continue with other files
		}
		astFiles = append(astFiles, file)
		return nil
	})
	if err != nil {
		return state, fmt.Errorf("failed to walk repo files: %w", err)
	}

	// Use main.go to detect registered services.
	mainPath := filepath.Join(state.LocalRepoPath, "main.go")
	mainAst, err := parser.ParseFile(fset, mainPath, nil, parser.ParseComments)
	if err != nil {
		return state, fmt.Errorf("failed to parse main.go: %w", err)
	}
	registered := findRegisteredServices(mainAst)
	if len(registered) == 0 {
		fmt.Println("No registered gRPC services found in main.go")
		return state, nil
	}

	// For each registered service, look for methods across all AST files.
	var services []ServiceInfo
	for serviceName := range registered {
		methods := []*ast.FuncDecl{}
		for _, file := range astFiles {
			methods = append(methods, collectServiceMethods(file, serviceName)...)
		}
		services = append(services, ServiceInfo{
			Name:    serviceName,
			Methods: methods,
		})
	}

	// Get package name from main.go (assuming all files share the same package).
	pkgName := mainAst.Name.Name

	// Generate one file per service.
	for _, service := range services {
		if err := generateServiceFile(service, pkgName, tmpDir, fset); err != nil {
			fmt.Printf("Error generating file for service %s: %v\n", service.Name, err)
		} else {
			fmt.Printf("Generated file for service %s\n", service.Name)
		}
	}

	state.AstControlFlowFolderPath = tmpDir
	return state, nil
}

type ServiceInfo struct {
	Name    string
	Methods []*ast.FuncDecl
}

// findRegisteredServices finds function calls that register gRPC services.
func findRegisteredServices(f *ast.File) map[string]bool {
	services := make(map[string]bool)
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		// Check if the function name starts with "Register" and ends with "Server".
		if !strings.HasPrefix(selector.Sel.Name, "Register") || !strings.HasSuffix(selector.Sel.Name, "Server") {
			return true
		}
		// Expect at least two arguments.
		if len(call.Args) < 2 {
			return true
		}
		var serviceName string
		switch arg := call.Args[1].(type) {
		case *ast.SelectorExpr:
			serviceName = arg.Sel.Name
		case *ast.Ident:
			serviceName = arg.Name
		default:
			return true
		}
		services[serviceName] = true
		return true
	})
	return services
}

// collectServiceMethods collects all method declarations whose receiver matches serviceName.
func collectServiceMethods(f *ast.File, serviceName string) []*ast.FuncDecl {
	var methods []*ast.FuncDecl
	for _, decl := range f.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Recv == nil {
			continue
		}
		for _, field := range funcDecl.Recv.List {
			switch expr := field.Type.(type) {
			case *ast.StarExpr:
				if ident, ok := expr.X.(*ast.Ident); ok && ident.Name == serviceName {
					methods = append(methods, funcDecl)
				}
			case *ast.Ident:
				if expr.Name == serviceName {
					methods = append(methods, funcDecl)
				}
			}
		}
	}
	return methods
}

// generateServiceFile uses go/printer to render method AST nodes and writes them to a file.
func generateServiceFile(service ServiceInfo, pkgName, outputFolder string, fset *token.FileSet) error {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("package %s\n\n", pkgName))
	builder.WriteString("// Auto-generated control flow file for service: " + service.Name + "\n\n")
	for _, method := range service.Methods {
		var buf bytes.Buffer
		// Use go/printer to print the AST node as source code.
		err := printer.Fprint(&buf, fset, method)
		if err != nil {
			return fmt.Errorf("printer error: %w", err)
		}
		builder.WriteString(buf.String())
		builder.WriteString("\n\n")
	}
	fileName := fmt.Sprintf("%s_control_flow.go", strings.ToLower(service.Name))
	outputPath := filepath.Join(outputFolder, fileName)
	return os.WriteFile(outputPath, []byte(builder.String()), 0644)
}
