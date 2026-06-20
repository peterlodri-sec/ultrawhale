package blocks

import (
	"fmt"
	"strings"
)

// ── Schema Primitive — Formal Block Structure Validation ─────────────
// Vaked layer: Declares. Validates that blocks conform to their schema.
// Every block primitive declares its schema here.

// BlockSchema defines the valid structure of a block type.
type BlockSchema struct {
	Name     string              // "Write", "Sed", "Diff"
	Version  string              // semver of the schema
	Fields   []SchemaField       // required fields
	Validate func(*Block) []string // custom validation
}

// SchemaField is a single field in a block schema.
type SchemaField struct {
	Name     string // "Ref", "Content", "Path"
	Type     string // "string", "[]byte", "int", "BlockKind"
	Required bool
	MinSize  int  // for []byte fields
	MaxSize  int  // for []byte fields
}

// ── Schema Registry ───────────────────────────────────────────────────

var schemaRegistry = make(map[string]*BlockSchema)

// RegisterSchema adds a block schema to the registry.
func RegisterSchema(s *BlockSchema) {
	schemaRegistry[s.Name] = s
}

// ValidateBlock validates a block against its registered schema.
func ValidateBlock(b *Block) []string {
	schema, ok := schemaRegistry[string(b.Kind)]
	if !ok {
		return nil // no schema registered — pass
	}

	var errors []string

	for _, field := range schema.Fields {
		switch field.Name {
		case "Ref":
			if b.Ref == "" && field.Required {
				errors = append(errors, fmt.Sprintf("%s: Ref is required", schema.Name))
			}
		case "Content":
			if len(b.Content) == 0 && field.Required {
				errors = append(errors, fmt.Sprintf("%s: Content is required", schema.Name))
			}
			if field.MaxSize > 0 && len(b.Content) > field.MaxSize {
				errors = append(errors, fmt.Sprintf("%s: Content exceeds max size %d", schema.Name, field.MaxSize))
			}
		case "Path":
			if b.Path == "" && field.Required {
				errors = append(errors, fmt.Sprintf("%s: Path is required", schema.Name))
			}
		}
	}

	if schema.Validate != nil {
		errors = append(errors, schema.Validate(b)...)
	}

	return errors
}

// ── Built-in Schemas ──────────────────────────────────────────────────

func init() {
	RegisterSchema(&BlockSchema{
		Name:    "Write",
		Version: "v1.0.0",
		Fields: []SchemaField{
			{Name: "Ref", Type: "string", Required: true},
			{Name: "Content", Type: "[]byte", Required: true, MaxSize: 10 * 1024 * 1024},
			{Name: "Path", Type: "string", Required: true},
		},
	})

	RegisterSchema(&BlockSchema{
		Name:    "Sed",
		Version: "v1.0.0",
		Fields: []SchemaField{
			{Name: "Ref", Type: "string", Required: true},
			{Name: "Content", Type: "[]byte", Required: true},
		},
	})

	RegisterSchema(&BlockSchema{
		Name:    "Diff",
		Version: "v1.0.0",
		Fields: []SchemaField{
			{Name: "Content", Type: "[]byte", Required: true},
		},
	})

	RegisterSchema(&BlockSchema{
		Name:    "Compress",
		Version: "v1.0.0",
		Fields: []SchemaField{
			{Name: "Content", Type: "[]byte", Required: true},
			{Name: "Kind", Type: "BlockKind", Required: true},
		},
		Validate: func(b *Block) []string {
			if b.Kind != "compress" {
				return []string{"Compress: Kind must be 'compress'"}
			}
			return nil
		},
	})
}

// SchemaStatus returns compact schema registry status.
func SchemaStatus() string {
	var names []string
	for name := range schemaRegistry {
		names = append(names, name)
	}
	return fmt.Sprintf("schema: %d block types registered (%s)", len(names), strings.Join(names, ", "))
}
