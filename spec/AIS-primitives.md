## AIS Primitives

This document normatively defines the core artifacts of Agentic Intent Security (AIS).

### 1. Identifiers and Keys
- Principals: user, agent runtime, tool server use stable identifiers (OIDC `sub`, DID, or mTLS DN).
- Keys: Ed25519 or P‑256 for signatures; rotation supported via `kid`.
- IDs: URN UUIDs or DIDs; references use URI fragments for step addressing.

### 2. UIA — User Intent Assertion
Purpose: Declare the user’s purpose, constraints, and risk budget.

Required fields:
- `id` (string, URI)
- `subject` ({ id })
- `purpose` (string, natural language + optional tags)
- `constraints` ({ dataClasses[], jurisdictions[], timeWindow{ notAfter }, destinations? })
- `riskBudget` ({ level:int 0–5, maxWrites:int, maxRecords:int, maxExternalCalls?:int })
- `policyProfile` (string; e.g., `research-readonly`)
- `proof` (VC‐style or JWS; selective disclosure recommended)

Semantics:
- Declarative “why/what,” not “how.”
- Minimal disclosure: only fields needed for a given tool may be revealed.
- Bind to session context where available.

### 3. APA — Agent Plan Assertion
Purpose: Concrete, stepwise plan derived from UIA.

Required fields:
- `id`, `uia`
- `model` ({ vendor, version, hash })
- `steps[]` ({ id, tool, args, expected{ dataClasses[], writes:int, externalCalls?:int }, alignment{ score:0..1, why }})
- `totals` ({ predictedWrites, predictedRecords, predictedExternalCalls? })
- `proof` (JWS by agent runtime)

Semantics:
- Steps MUST be stable and addressable; args MUST be fully explicit.
- Alignment scores MUST be computed by a deterministic procedure (profile‑specific).
- Any dynamic tool selection MUST be represented as alternative branches with guards.

### 4. APr — Alignment Proof
Purpose: Machine‑verifiable proof that APA entails UIA under constraints.

Required fields:
- `id`, `uia`, `apa`
- `method` (e.g., `semantic-entailment-v1`)
- `evidence` ({ coverage:0..1, risk:0..1, obligations?[] })
- `proof` (JWS by verifier)

Semantics:
- Coverage ≥ threshold AND risk ≤ budget required by local policy.
- Supports zero‑knowledge or redacted proofs in future profiles.

Verifier profile: `semantic-entailment-v1` (non‑normative → can be normative via test vectors)
- Inputs: UIA.purpose (string), APA (steps with args), optional policy profile.
- Procedure: extract lowercase keywords (≥4 chars) from purpose; a step entails if its prompt/url contains any keyword; coverage = alignedSteps/totalSteps; risk baseline 0, +0.5 if predictedWrites>0, +0.3 if purpose suggests outbound actions (send/post/write/export/email/delete); clamp to [0,1].
- Output: evidence.coverage, evidence.risk.
- Tolerance: implementers MUST treat evidence equal within ±1e‑9 as matching recomputation.

### 5. TCA — Tool Capability Assertion
Purpose: Formal contract of tool operations and side effects.

Required fields:
- `id` (`urn:tca:<tool>@<version>`)
- `operations[]` ({ name, argsSchema(JSON Schema), effects{ writes:int|range, dataClasses[], destinations? }})
- `operator` (entity operating the tool)
- `proof` (JWS by operator)

Semantics:
- Tools MUST reject calls that violate schema or exceed declared effects.

### 6. IBE — Intent‑Bound Envelope
Purpose: Per‑call wrapper binding call to intent/plan.

Required fields:
- `id`, `uiaRef`, `apaStepRef`, `aprRef`, `tcaRef`, `nonce`, `exp`, `sig`

Semantics:
- Servers MUST verify signatures, freshness, and compatibility before execution.
- Nonces MUST NOT be reused; expired IBEs MUST be rejected.
- Audience/tool binding: servers SHOULD bind IBE to specific tool operation via `tcaRef` and `apaStepRef`.
- Canonicalization: the `sig` MUST cover the envelope with `sig` field blanked; clock skew allowance ≤ 2 minutes.

### 7. SCA — Supply Chain Assertion
Purpose: Provenance for models, guardrails, and policies.

Fields:
- `components[]` ({ type: model|guard|policy|dataset, name, version, hash, source })
- `proof`

### 8. Revocation
- Short‑lived CRLs for UIA and APA; IBE validity MUST be bounded.
- Revocation status SHOULD be cached and stapled by clients where feasible.

### 8.1 Revocation Format (CRL)
Non-normative JSON example:
```json
{
  "@type": "CRL",
  "issued": "2025-01-01T00:00:00Z",
  "expires": "2025-01-01T00:10:00Z",
  "revoked": [
    {"type": "UIA", "id": "urn:uia:...", "reason": "user-revoked"},
    {"type": "APA", "id": "urn:apa:...", "reason": "superseded"}
  ]
}
```
Stapling: tools SHOULD include last-seen CRL `issued` and `expires` in IBE verification context and audit events.

### 9. Trust Model
- Users authorize UIA; agents propose APA; independent verifiers issue APr; tool operators publish TCA; tools enforce IBE.

---

## Appendix A: Canonicalization & Signing (Normative)

- Payload canonicalization: JSON with UTF‑8, object keys sorted lexicographically by Unicode code point, no insignificant whitespace.
- Numeric normalization: integers as digits, floats with minimal representation; timestamps RFC3339 UTC.
- IBE signing input: the IBE object with `sig` field omitted/blanked, canonicalized as above.
- JWS: HS256 (demo) or registered alg with `typ":"JWT"`. The signature covers `b64url(header) + '.' + b64url(canonicalized payload)`.
- Clock skew: verifiers MUST allow ≤ 120 seconds.

Golden example (payload excerpt, canonicalized and signed):
```json
{"@type":"IBE","aprRef":"urn:apr:abc","apaStepRef":"s1","exp":"2099-01-01T00:00:00Z","id":"urn:ibe:xyz","nonce":"abcd","tcaRef":"urn:tca:ollama.generate@1","uiaRef":"urn:uia:foo"}
```
Header:
```json
{"alg":"HS256","typ":"JWT"}
```
Signing input: `base64url(header) + '.' + base64url(payload)` → signature = `HMAC-SHA256(secret, input)`.

Implementations MUST reproduce the byte-for-byte payload when verifying.


