package structural

// The ribbon never lands on another ribbon (FR-506) because it is placed by the kit's arranger over
// the folder every running ribbon shares; the kit's own tests hold the placing. WeatherRibbon's part
// is the wiring, held here: its service arranges through the kit's arranger; the composition root
// hands that arranger the kit's shared folder as its neighbours, with nothing of WeatherRibbon's own
// standing in for it.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/oernster/ribbonkit/structure"
)

// arrangerPath and occupancyPath are the kit's arranger and its shared folder of running ribbons.
const (
	arrangerPath  = kitModule + "/application/arranger"
	occupancyPath = kitModule + "/infrastructure/occupancy"
)

// parsed answers path's syntax tree with the name each import is known by in it.
func parsed(t *testing.T, path string) (*ast.File, map[string]string) {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]string{}
	for _, spec := range file.Imports {
		imported, _ := strconv.Unquote(spec.Path.Value)
		name := filepath.Base(imported)
		if spec.Name != nil {
			name = spec.Name.Name
		}
		names[name] = imported
	}
	return file, names
}

// selects reports whether expression is pkg.Name with pkg naming the import at path.
func selects(expression ast.Expr, names map[string]string, path, name string) bool {
	selector, ok := expression.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != name {
		return false
	}
	pkg, ok := selector.X.(*ast.Ident)
	return ok && names[pkg.Name] == path
}

func TestTheRibbonUsesTheSharedPlacement(t *testing.T) {
	root := structure.Root(t)
	service, names := parsed(t, filepath.Join(root, "internal", "application", "service.go"))
	embeds := false
	ast.Inspect(service, func(node ast.Node) bool {
		spec, ok := node.(*ast.TypeSpec)
		if !ok || spec.Name.Name != "Service" {
			return true
		}
		for _, field := range spec.Type.(*ast.StructType).Fields.List {
			star, ok := field.Type.(*ast.StarExpr)
			embeds = embeds || (len(field.Names) == 0 && ok && selects(star.X, names, arrangerPath, "Arranger"))
		}
		return false
	})
	if !embeds {
		t.Error("application.Service does not embed the kit's *arranger.Arranger, so the ribbon is not placed by it")
	}

	main, mainNames := parsed(t, filepath.Join(root, "main.go"))
	if !occupied(main, mainNames) {
		t.Error("main.go does not hand the service's Neighbours a value from the kit's occupancy folder")
	}

	// Shipped code only: the root's tests of openNeighbours import the folder they exercise.
	for _, path := range structure.Shipped(goFiles(t)) {
		if filepath.Base(path) == "main.go" && filepath.Dir(path) == root {
			continue
		}
		if _, imported := parsed(t, path); containsValue(imported, occupancyPath) {
			t.Errorf("%s reaches the shared folder itself; only main.go wires it", structure.Relative(root, path))
		}
	}
}

// occupied reports whether main sets a Neighbours field from occupancy.Open itself or from a
// function answering the kit's *occupancy.Folder.
func occupied(main *ast.File, names map[string]string) bool {
	answersFolder := map[string]bool{}
	for _, declaration := range main.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Type.Results == nil {
			continue
		}
		for _, result := range function.Type.Results.List {
			if star, ok := result.Type.(*ast.StarExpr); ok && selects(star.X, names, occupancyPath, "Folder") {
				answersFolder[function.Name.Name] = true
			}
		}
	}
	values := map[string]bool{}
	found := false
	ast.Inspect(main, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.AssignStmt:
			for index, right := range node.Rhs {
				if call, ok := right.(*ast.CallExpr); ok && index < len(node.Lhs) {
					if name, ok := call.Fun.(*ast.Ident); ok && answersFolder[name.Name] {
						if left, ok := node.Lhs[index].(*ast.Ident); ok {
							values[left.Name] = true
						}
					}
				}
			}
		case *ast.KeyValueExpr:
			key, ok := node.Key.(*ast.Ident)
			if !ok || key.Name != "Neighbours" {
				return true
			}
			if value, ok := node.Value.(*ast.Ident); ok && values[value.Name] {
				found = true
			}
			if call, ok := node.Value.(*ast.CallExpr); ok && selects(call.Fun, names, occupancyPath, "Open") {
				found = true
			}
		}
		return true
	})
	return found
}

// containsValue reports whether any import in names is path.
func containsValue(names map[string]string, path string) bool {
	for _, imported := range names {
		if imported == path {
			return true
		}
	}
	return false
}
