# Spoken Language Examples — v100.0.0

## Human→Capability Graph

Any spoken language → same graph operation.

| Language | Utterance | Graph Operation |
|----------|-----------|----------------|
| 🇬🇧 EN | "build auth" | addMain → patch → compile → run |
| 🇭🇺 HU | "építsd meg az autentikációt" | addMain → patch → compile → run |
| 🇩🇪 DE | "baue die Authentifizierung" | addMain → patch → compile → run |
| 🇫🇮 FI | "rakenna todennus" | addMain → patch → compile → run |
| 🇯🇵 JP | "認証を構築する" | addMain → patch → compile → run |
| 🇨🇳 ZH | "构建认证" | addMain → patch → compile → run |

## Inspired by Zerolang

[zerolang](https://github.com/vercel-labs/zerolang) by Vercel Labs:
- Graph-native programming language for AI agents
- Humans speak in outcomes, agents operate on graph semantics
- Text projections (`.0` files) are view-only — the graph IS the truth

ultrawhale extends this: the SACRED surface accepts ANY language.
The capability graph IS the truth. The encoding dialect is just the medium.

## Further Reading

- [zerolang](https://github.com/vercel-labs/zerolang) — Graph-native programming language
- [Vaked Theory](docs/primitive-mapping.md) — Capability graph philosophy
- [TRANSLATE Recursion](docs/asymmetry-of-inputs.md) — The 5th recursion
- [Hieroglyphs](docs/hieroglyphs.md) — Visual meaning compression

## Try It

```
/translate en "build auth"
/translate hu "építsd meg az autentikációt"
/translate de "baue die Authentifizierung"
```
