## AIS Protocol Flows

### A. Intent Capture and Planning
1. Client renders intent form; user confirms purpose and constraints.
2. Client issues UIA (SD‑JWT VC), signs with user/agent key; obtains user consent UX artifact.
3. Agent runtime generates APA (deterministic planner); signs; submits to verifier.
4. Verifier computes APr; on success returns proof ref.

### B. Tool/API Invocation (Per‑Call)
1. Agent selects APA step; prepares IBE with minimal UIA/APA excerpts + APr.
2. Tool server verifies: sigs, nonce, exp, TCA compatibility, alignment threshold, risk budget, data‑class policy.
3. Optional elevation: stepwise co‑signature if threshold near boundary or new destination.
4. Execute; emit audit record with redacted fields and hashes.

### C. Revocation and Expiry
1. UIA/APA revocation lists fetched via short‑TTL cache.
2. Tools staple last‑seen revocation status to audit events.

### D. Audit and Forensics
1. Append‑only JSONL with signatures: UIA_ref, APA_ref, APr_ref, IBE_id, tcaRef, result hashes, data‑class counters.
2. Privacy: store references and hashes, not raw content.

### E. Elevation and Stepwise Consent
1. Tool challenges with `needConsent` citing delta and risk.
2. Client mints ephemeral consent token (co‑signed), updates APA totals, obtains fresh APr.

### F. Multi‑Agent Cross‑Check
1. Watchguard agent computes minimal alternative APA; compare deltas.
2. If divergence > threshold, require elevation or halt.


