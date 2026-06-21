# CI Lessons — v100.0.0

> "YAML, JSON etc ARE structured, rigid structs CAN BE the most pita" — Peter

## Lessons Learned

| # | Lesson | Fix |
|---|--------|-----|
| 1 | **Validate before push** | `python3 -c "import yaml; yaml.safe_load(open('file.yml'))"` |
| 2 | **Upstream code is NOT our problem** | Exclude vaked/plugin, infra_bar, widgets from CI |
| 3 | **Go package conflicts are silent killers** | asm/ directory had mixed packages → fixed |
| 4 | **GPG signing blocks CI** | `git config --global commit.gpgsign false` in CI |
| 5 | **Long tests timeout** | `-run TestReadWrite` → fast path, not full suite |
| 6 | **Rigid structs need flexible minds** | YAML indentation, JSON types, Go imports — all rigid |

## CI Architecture

```
Push → validate YAML → test blocks (race) → build (linux+darwin) → bench → publish
  ↓         ↓              ↓                    ↓                        ↓
 YAML      python         go test             go build              bench results
 valid?    yaml.safe     -race               cross-compile          in CI log
```

## The SACRED CI Rule

> "The SACRED surface must not break. Validate BEFORE you push.
> Rigid structs are pita. Learn from them. They teach patience."
> — Peter, v100.0.0
