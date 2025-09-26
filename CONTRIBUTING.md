# Contributing to AIS Demo and Spec

Thank you for improving Agentic Intent Security (AIS).

## Scope
- Code: the demo server under `cmd/aisdemo` and `internal/ais`
- Spec: documents under `spec/`

## Quickstart (dev)
```bash
# Run locally
OLLAMA_URL=http://localhost:11434 OLLAMA_MODEL=codellama:7b go run ./cmd/aisdemo
# Lint build
go build ./cmd/aisdemo
```

## Branch & PR
- Create a feature branch and open a PR.
- Include: purpose, design notes, tests or test vectors, and spec links.
- CI must pass; no linter errors; keep diffs focused.

## Style
- Go: clear names, early returns, no unused code, avoid deep nesting.
- Spec: use MUST/SHOULD/MAY; keep examples minimal and runnable.
- JSON examples: 2‑space indentation, stable key order.

## Tests & Vectors
- Put normative examples in `spec/test-vectors/` with a README.
- For new APr methods or TCA ops, add entries to registries and include pass/fail vectors.

## Security
- Do not commit secrets. Report vulnerabilities via `SECURITY.md`.

## Governance
- SemVer for spec; registries are append‑only.
- Breaking changes require a discussion issue and consensus in review.
