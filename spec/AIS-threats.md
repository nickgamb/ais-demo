## AIS Threat Model and Security Properties

### Properties
- Intent binding per call; replay protection; minimal disclosure; provenance; least privilege by risk budgets.

### Threats and Mitigations
1) Prompt/plan injection
- Mitigate with provenance filters, APA verifier, guarded tools, simulation sandboxes, human co‑sign for high‑risk.

2) Exfiltration/data drift
- Data‑class membranes, destination allowlists, quotas, watermarking, DLP hooks.

3) Replay and token theft
- Nonce + short expiry on IBE, mTLS/DPoP, audience scoping.

4) Capability escalation
- TCA schema enforcement; deny unknown args; per‑step consent; ZCAP/UCAN attenuation.

5) Supply chain compromise
- SCA provenance, model hash pinning, policy version pinning, allowlists.

6) Mis‑alignment of plans
- APr thresholds, randomized re‑verification, multi‑agent cross‑check.

7) Privacy leakage via audit
- Redaction, hashing, selective disclosure, split storage for keys vs events; retention and purpose limitation.

8) Time and revocation abuse
- CRLs with short TTL; stapling; strict clock skew; fail‑closed on stale status.


