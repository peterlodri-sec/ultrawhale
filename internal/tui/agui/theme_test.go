package agui

import "testing"

func TestThemeCycle(t *testing.T) {
  start := Current.Name
  for i := 0; i < 3; i++ { CycleTheme() }
  if Current.Name != start {
    t.Fatalf("3x CycleTheme did not return to start")
  }
}

func TestAllThemesValid(t *testing.T) {
  names := []ThemeName{DenseMatrixGreen, CleanGraphCyberpunk, TacticalGraveyard}
  for _, name := range names {
    th, ok := Themes[name]
    if !ok { t.Fatalf("theme %s missing", name) }
    if th.Bg == "" || th.Fg == "" || th.Accent == "" {
      t.Fatalf("theme %s has empty color", name)
    }
  }
}
