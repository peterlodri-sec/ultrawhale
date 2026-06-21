# ultrawhale Public Datasets — Zero-Trust Manifest

## PII Guarantee

All public datasets are **zero-trust scrubbed** before export.
No API keys. No emails. No private IPs. No personal identifiers.

## Scrub Procedure

1. Run `.dev/scrub-pii.sh` before any export
2. Verify: `grep -r "sk-\|@\|192.168" dataset/` returns empty
3. Sign with VICE Genesis block: `SignClaim("dataset-exported-clean", "peter")`
4. Publish with MIT + CC-BY-4.0 license

## Available Datasets

| Dataset | Path | Size | License |
|---------|------|------|---------|
| DogFood v1 | vaked.dev/ultrawhale/dogfood | Growing | MIT + CC-BY-4.0 |

## Trust Verification

Every dataset export is signed by the VICE Genesis block.
Trust score: 1.0000. Verified: true.
