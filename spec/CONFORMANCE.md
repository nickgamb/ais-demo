# AIS Conformance Profiles

Profiles define REQUIRED behaviors and artifacts for interoperable implementations.

## Profiles
- Minimal:
  - UIA, APA, APr (semantic-entailment-v1), IBE with JWS HS256
  - Guard enforces: signature, expiry, coverage threshold, risk budgets, TCA effects, data classes
  - Audit events with required fields
- Recommended (adds):
  - JSON Schema validation for tool args (TCA)
  - Revocation stapling for UIA/APA, short TTL cache
  - OpenTelemetry spans with UIA→APA→IBE links
- Advanced (adds):
  - External verifier integration (OPA/Rego or guardrails)
  - Elevation flow with stepwise consent and co-signatures

## Checklist (Minimal)
- [ ] Generate UIA with constraints and riskBudget
- [ ] Generate APA with deterministic steps and alignment.score
- [ ] Compute APr (semantic-entailment-v1) coverage/risk
- [ ] Mint IBE and sign (exclude `sig` field from payload when signing)
- [ ] Guard verifies sigs, expiry, coverage ≥ threshold, risk within budget
- [ ] Guard checks TCA compatibility and writes ≤ effects
- [ ] Emit audit event with refs and `ok` status

## Test Strategy
- Use `spec/test-vectors/` artifacts and run verifier/guard against them
- Provide pass/fail vectors for boundary conditions (coverage threshold, risk budget, expiry)
