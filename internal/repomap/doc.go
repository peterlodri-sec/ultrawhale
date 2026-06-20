// Package repomap builds a dependency graph of the workspace using tree-sitter
// AST analysis, inspired by Aider's repository map. The graph is serialized as
// JSON and injected into the LLM context so the model understands the codebase
// structure before making edits.
//
// Architecture:
//   parser.go  — tree-sitter language detection, file parsing
//   graph.go   — symbol extraction, dependency edges
//   cache.go   — SHA-based cache with file-change invalidation
//   plugin.go  — Whale Plugin integration (HookProvider, StartupContextProvider)
package repomap
