package buildcodegraph

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// TODO: Built for golang gRpc project. Extend it to support other languages and frameworks.
func (a *Activities) BuildAstControlFlow(ctx context.Context, state BuildCodeGraphState) (BuildCodeGraphState, error) {
	tmpDir, err := os.MkdirTemp("", "controlFlow-*")
	if err != nil {
		return state, fmt.Errorf("failed to create temp dir: %w", err)
	}

	fset := token.NewFileSet()

	entryPoint := filepath.Join(state.LocalRepoPath, "main.go")
	astFile, err := parser.ParseFile(fset, entryPoint, nil, parser.ParseComments)
	if err != nil {
		fmt.Printf("Failed to parse main.go: %v\n", err)
		return state, err
	}

	// Find the registered services in main.go.
	registered := findRegisteredServices(astFile)
	if len(registered) == 0 {
		fmt.Println("No registered gRPC services found in main.go")
		return state, nil
	}

	// For this example, assume that the current file (main.go) contains the methods for the services.
	// In a real project, you may need to parse multiple files in your package.
	var services []ServiceInfo
	for serviceName := range registered {
		methods := collectServiceMethods(astFile, serviceName)
		services = append(services, ServiceInfo{
			Name:    serviceName,
			Methods: methods,
		})
	}

	// Get the package name from the parsed file.
	pkgName := astFile.Name.Name

	// Generate one file per service.
	for _, service := range services {
		if err := generateServiceFile(service, pkgName, tmpDir); err != nil {
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

func findRegisteredServices(f *ast.File) map[string]bool {
	services := make(map[string]bool)
	fmt.Printf("Got AST as %#v\n", f)
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
		// Expect at least two arguments: e.g. RegisterXXXServer(bootServer.GrpcServer, app.SomeService)
		if len(call.Args) < 2 {
			return true
		}
		// For simplicity, assume the service instance is passed as the second argument.
		// It might be a selector (e.g., app.LoginService) or an identifier.
		var serviceName string
		switch arg := call.Args[1].(type) {
		case *ast.SelectorExpr:
			serviceName = arg.Sel.Name
		case *ast.Ident:
			serviceName = arg.Name
		default:
			// If we can't determine the service name, skip.
			return true
		}
		services[serviceName] = true
		return true
	})
	return services
}

func collectServiceMethods(f *ast.File, serviceName string) []*ast.FuncDecl {
	var methods []*ast.FuncDecl
	for _, decl := range f.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Recv == nil {
			continue
		}
		// Check each receiver.
		for _, field := range funcDecl.Recv.List {
			// The receiver type could be "*MyService" or "MyService".
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

// generateServiceFile generates a Go source file containing the methods for the given service.
func generateServiceFile(service ServiceInfo, pkgName, outputFolder string) error {
	// Build the file content with package declaration and all methods.
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("package %s\n\n", pkgName))
	builder.WriteString("// Auto-generated control flow file for service: " + service.Name + "\n\n")
	for _, method := range service.Methods {
		// Use ast.Node formatting functions or simply print the raw source code if available.
		// For simplicity, we'll use ast.Node's String() representation.
		// In production, consider using go/printer for proper formatting.
		builder.WriteString(fmt.Sprintf("%#v\n\n", method))
	}

	fileName := fmt.Sprintf("%s_control_flow.go", strings.ToLower(service.Name))
	outputPath := filepath.Join(outputFolder, fileName)

	return os.WriteFile(outputPath, []byte(builder.String()), 0644)
}
