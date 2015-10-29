frugal/parser
=============

The raw parsing API for frugal. The API entry point is `CompileContext` (see `context.go`). From there, you can either call `ParseRecursive()` to parse a file and all included files, or create a parser via `NewParser()` to parse a single file.

The results of parsing are store as an abstract syntax tree (AST). The AST is documented and specified in `ast.go`. Some fields change or are only set during semantic analysis, which is a separate pass. The states of such fields are documented in `ast.go`.

The parser does not perform any semantic analysis. To do that, use the frugal/sema package. Semantic analysis can only be performed on parse trees that have been recursively parsed.

Examples:

Recursively parse a file and all of its includes:
```
context := parser.NewCompileContext()
tree := context.ParseRecursive(file)
if tree == nil {
  context.PrintErrors()
}
```

Parse a single file:
```
context := parser.NewCompileContext()
tree, err := context.Parse(file)
```
