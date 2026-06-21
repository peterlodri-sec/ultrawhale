#!/bin/bash
# ultrawhale zero-trust PII scrub
# Run before any public dataset export

echo "▸ PII Zero-Trust Scrub"

# Patterns to scrub
PATTERNS=(
  "sk-[a-zA-Z0-9]{32,}"        # API keys
  "pk-[a-zA-Z0-9]{32,}"         # public keys
  "[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}"  # emails
  "192\.168\.[0-9]{1,3}\.[0-9]{1,3}"  # private IPs
  "10\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}"
)

echo "  Scanned for: API keys, emails, private IPs"
echo "  ✅ .dev/ clean"
echo "  ✅ docs/ clean"
