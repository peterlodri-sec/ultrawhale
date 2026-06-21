# File I/O Primitives Audit — v16.0.0

## Primitives Audited

| Primitive | O(T) | Uses blocks? | Journaled? | Gaps |
|-----------|------|-------------|------------|------|
| **Write** | O(1) | ✅ self | ✅ journal.Push | Large files need streaming |
| **Read** | O(n) | ✅ self | ❌ | No partial reads, no glob |
| **Sed** | O(n*m) | ❌ os.ReadFile | ❌ | Pre-allocated, no streaming |
| **Diff** | O(n) | ❌ os.ReadFile | ❌ | No streaming diff |
| **MMap** | O(1) map | ✅ self | ❌ | Unix-only, no Windows |
| **Compress** | O(n) | ❌ in-memory | ❌ | No streaming compress |
| **Export** | O(n) | ❌ json.Marshal | ❌ | No streaming export |

## Gaps

| Gap | Severity | Fix |
|-----|----------|-----|
| No streaming Write | Medium | Add WriteStream for large files |
| No partial Read | Medium | Add ReadRange(offset, length) |
| No glob support | Low | Add Glob(pattern) for file discovery |
| Sed doesn't use blocks.Write | High | Wire SedFile through blocks.Write |
| Diff doesn't use blocks.Read | Medium | Use blocks.Read for prev state |
| MMap is Unix-only | Low | Add Windows fallback |
| No streaming export | Low | Add ExportStream for large brains |

## Performance Suggestions

1. **Sed: use mmap** instead of os.ReadFile for large files
2. **Diff: use mmap** for both files when >64KB
3. **Compress: streaming** instead of in-memory for >1MB
4. **Export: streaming JSON** instead of json.Marshal for large brains
