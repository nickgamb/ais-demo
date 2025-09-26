# AIS Error Codes

Standardized reasons to aid interop and troubleshooting.

## Categories
- IBE: envelope and signature issues
- ALIGN: alignment/verifier issues
- RISK: budget violations
- TCA: capability mismatches
- DATA: data-class violations
- INPUT: schema/argument problems
- AUTHZ: policy/authorization failures
- SYS: transient/system errors

## Codes
- IBE-EXPIRED: IBE expired
- IBE-SIG-INVALID: signature invalid or payload mismatch
- IBE-STEP-NOT-FOUND: APAStepRef not found in APA
- ALIGN-BELOW-THRESHOLD: APr coverage below threshold
- ALIGN-MISMATCH: APr evidence does not match recomputation
- RISK-WRITES-EXCEEDED: APA predictedWrites exceeds UIA maxWrites
- TCA-OP-NOT-ALLOWED: tool operation not in TCA
- TCA-EFFECTS-EXCEEDED: step effects exceed TCA declared effects
- DATA-CLASS-NOT-PERMITTED: step data class not permitted by UIA constraints
- INPUT-SCHEMA-INVALID: args fail JSON Schema
- AUTHZ-POLICY-DENY: policy engine denied request
- SYS-RETRY: transient error; retry suggested

## Representation
Errors should return HTTP 4xx/5xx with body:
```json
{ "code": "ALIGN-BELOW-THRESHOLD", "message": "alignment below threshold", "details": {}}
```
