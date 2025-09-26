# AIS Registries

## APr Methods
| id | description | inputs | outputs | notes |
|---|---|---|---|---|
| semantic-entailment-v1 | Deterministic keyword entailment with simple risk | UIA.purpose, APA | evidence.coverage, evidence.risk | Reference method implemented in demo |
| external-policy-v1 | Delegate to policy engine (e.g., OPA/Rego) for entailment and risk | UIA, APA, policyProfile | evidence.coverage, evidence.risk | Returns deterministic scores from policy evaluation |

Registration template:
- id (string)
- description
- inputs (fields used)
- algorithm (normative summary)
- outputs (evidence fields)
- thresholds and tolerance
- security considerations

## TCA Operations
| name | version | effects | notes |
|---|---|---|---|
| ollama.generate | 1 | { writes: 0, dataClasses:["derived"] } | text generation |
| http.get | 1 | { writes: 0, dataClasses:["derived"] } | HTTP fetch |

Registration template:
- name (tool identifier)
- version
- args schema (JSON Schema)
- effects (writes range, dataClasses, destinations?)
- operator and proof


