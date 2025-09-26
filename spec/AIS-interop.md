## AIS Interoperability Profiles

### Transports
- HTTP(S) headers with detached JWS; gRPC metadata; MCP custom headers.

### Identity
- OIDC subject, DID key, or mTLS DN as principal IDs.

### Credentials
- UIA as SD‑JWT VC; APA/APr/IBE as JWS; ZK upgrade later.

### Authorization
- GNAP/RAR: mint short‑lived grants bound to UIA purpose and constraints.
- ZCAP‑LD/UCAN: carry attenuated capabilities referencing UIA.

#### Example: GNAP + RAR (non‑normative)
Request (client → AS):
```json
{
  "access_token": {
    "label": "ais",
    "access": [{
      "type": "ais.intent",
      "actions": ["plan","invoke"],
      "limits": {"uiaRef": "urn:uia:123", "notAfter": "2025-01-01T00:10:00Z"}
    }]
  },
  "client": {"key": {"proof": "httpsig"}}
}
```
Issued token’s audience: tool server; scope encodes `uiaRef` and expiry.

#### Example: ZCAP Invocation (non‑normative)
Capability (delegated from user → agent):
```json
{
  "@context": "https://w3id.org/security/v2",
  "@type": "ZCAP",
  "parentCapability": "urn:cap:tools",
  "invoker": "did:example:agent",
  "capabilityAction": "invoke",
  "caveat": [{"type": "uiaRef", "value": "urn:uia:123"}]
}
```
Invocation carries proof and references `uiaRef` and `tcaRef`.

### Provenance and Telemetry
- C2PA for content provenance; OpenTelemetry for traces (span links: UIA→APA→IBE).

### Hexa/IDQL Bridge (Optional)
- Map `policyProfile` to IDQL templates; compile to provider‑native policies; push via Hexa orchestrator.


