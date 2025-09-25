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

### 7. SCA — Supply Chain Assertion
Purpose: Provenance for models, guardrails, and policies.

Fields:
- `components[]` ({ type: model|guard|policy|dataset, name, version, hash, source })
- `proof`

### 8. Revocation
- Short‑lived CRLs for UIA and APA; IBE validity MUST be bounded.
- Revocation status SHOULD be cached and stapled by clients where feasible.

### 9. Trust Model
- Users authorize UIA; agents propose APA; independent verifiers issue APr; tool operators publish TCA; tools enforce IBE.


