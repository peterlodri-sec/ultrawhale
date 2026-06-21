---
license: mit
task_categories: [question-answering]
language: [en]
tags: [ultrawhale, vaked, fine-tuning, research, multi-model]
pretty_name: ultrawhale-dogfood
size_categories: [n<1K]
configs:
  - config_name: default
    data_files: "*.jsonl"
    default: true
---

# 🐕 ultrawhale-dogfood

60 samples, 20 CS topics, 15 flat string fields. PII-scrubbed. VICE signed.
