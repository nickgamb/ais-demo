# AIS Test Vectors

Run guards and verifiers against these JSON artifacts. Expected outcomes are annotated in filenames.

Files:
- uia_minimal.json
- apa_generate_step.json
- apr_pass_semantic_entailment_v1.json (expected PASS)
- apr_fail_low_coverage.json (expected FAIL)
- ibe_expired.json (expected FAIL)

## Run the conformance checks
```bash
# Build the CLI
go build ./cmd/aisconform
# Run (uses spec/test-vectors by default)
./aisconform
```

## Emit golden JWS examples
```bash
./aisconform --emit-golden
# writes golden_*_signed.json with canonical JSON + HS256 JWS using demo secret
```
