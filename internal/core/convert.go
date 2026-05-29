package core

import "strings"

// AsString extracts a string value from any type, returning empty string on mismatch.
func AsString(v any) string {
	s, _ := v.(string)
	return s
}

// AsAnySlice extracts a []any value, returning nil on mismatch or nil input.
func AsAnySlice(v any) []any {
	if v == nil {
		return nil
	}
	arr, ok := v.([]any)
	if ok {
		return arr
	}
	return nil
}

// ContainsArg reports whether want is present in argv.
func ContainsArg(argv []string, want string) bool {
	for _, got := range argv {
		if got == want {
			return true
		}
	}
	return false
}

// SkillNameDisabled reports whether name is in the disabled list.
// Comparison is case-insensitive with whitespace trimmed.
func SkillNameDisabled(name string, disabled []string) bool {
	name = strings.ToLower(strings.TrimSpace(name))
	for _, d := range disabled {
		if strings.ToLower(strings.TrimSpace(d)) == name {
			return true
		}
	}
	return false
}
