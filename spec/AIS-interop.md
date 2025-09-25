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

### Provenance and Telemetry
- C2PA for content provenance; OpenTelemetry for traces (span links: UIA→APA→IBE).

### Hexa/IDQL Bridge (Optional)
- Map `policyProfile` to IDQL templates; compile to provider‑native policies; push via Hexa orchestrator.


