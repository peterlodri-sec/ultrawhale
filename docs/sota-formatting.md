# SOTA Type-Safe Formatting — v100.1.0

## The Problem

```go
fmt.Sprintf("v%s · %s blocks", version, blockCount) // ❌ blockCount is int
// CI: %s has arg blockCount of wrong type int
```

## SOTA Methods (best to worst)

### 1. STR — Named Templates (OUR SOLUTION)

```go
S("v{version} · {blocks} blocks").
    With("version", CurrentVersion()).
    With("blocks", len(schemaRegistry)).
    Build()
// Always safe. Any type. No %d/%s mistakes. ✅
```

### 2. fmt.Sprint() — Universal Stringer

```go
fmt.Sprint(len(schemaRegistry)) // "142" ✅ Any type
```

### 3. strconv.Itoa() — Explicit Int

```go
strconv.Itoa(len(schemaRegistry)) // "142" ✅ Only int
```

### 4. fmt.Sprintf with %d — Explicit Type

```go
fmt.Sprintf("%d", len(schemaRegistry)) // "142" ✅ Must match type
```

### 5. Zod-Style Runtime Validation (Future)

```go
// Type-safe template validation at compile time
// Like TypeScript's zod, but for Go templates
// Planned for v110: Reactive Capability Graph
```

## CI Lesson

```
fmt.Sprintf("%s", int) → CI FAIL
STR("{blocks}", int)   → CI PASS
fmt.Sprint(int)         → CI PASS
```

**The answer: use STR for templates, fmt.Sprint for single values.
Never mix %s with int again.**
