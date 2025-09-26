## AIS Schemas (Draft JSON Schema)

Note: Draft‑07 compatible; examples trimmed for brevity.

### UIA
{
  "$id": "https://example.org/ais/uia.schema.json",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["@type", "id", "subject", "purpose", "constraints", "riskBudget", "policyProfile", "proof"],
  "properties": {
    "@type": {"const": "UIA"},
    "id": {"type": "string", "format": "uri"},
    "subject": {"type": "object", "required": ["id"], "properties": {"id": {"type": "string"}}},
    "purpose": {"type": "string", "minLength": 3},
    "constraints": {
      "type": "object",
      "required": ["dataClasses", "timeWindow"],
      "properties": {
        "dataClasses": {"type": "array", "items": {"type": "string"}},
        "jurisdictions": {"type": "array", "items": {"type": "string"}},
        "timeWindow": {"type": "object", "properties": {"notAfter": {"type": "string", "format": "date-time"}}, "required": ["notAfter"]}
      }
    },
    "riskBudget": {
      "type": "object",
      "properties": {"level": {"type": "integer", "minimum": 0, "maximum": 5}, "maxWrites": {"type": "integer"}, "maxRecords": {"type": "integer"}, "maxExternalCalls": {"type": "integer"}},
      "required": ["level", "maxWrites", "maxRecords"]
    },
    "policyProfile": {"type": "string"},
    "proof": {"type": "object"}
  }
}

### APA
{
  "$id": "https://example.org/ais/apa.schema.json",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["@type", "id", "uia", "model", "steps", "totals", "proof"],
  "properties": {
    "@type": {"const": "APA"},
    "id": {"type": "string", "format": "uri"},
    "uia": {"type": "string", "format": "uri"},
    "model": {"type": "object", "required": ["hash"], "properties": {"vendor": {"type": "string"}, "version": {"type": "string"}, "hash": {"type": "string"}}},
    "steps": {
      "type": "array",
      "items": {"type": "object", "required": ["id", "tool", "args", "expected", "alignment"],
        "properties": {
          "id": {"type": "string"},
          "tool": {"type": "string"},
          "args": {"type": "object"},
          "expected": {"type": "object", "required": ["dataClasses", "writes"], "properties": {"dataClasses": {"type": "array", "items": {"type": "string"}}, "writes": {"type": "integer"}}},
          "alignment": {"type": "object", "required": ["score"], "properties": {"score": {"type": "number", "minimum": 0, "maximum": 1}, "why": {"type": "string"}}}
        }}
    },
    "totals": {"type": "object", "required": ["predictedWrites", "predictedRecords"], "properties": {"predictedWrites": {"type": "integer"}, "predictedRecords": {"type": "integer"}}},
    "proof": {"type": "object"}
  }
}

### APr
{
  "$id": "https://example.org/ais/apr.schema.json",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["@type", "id", "uia", "apa", "method", "evidence", "proof"],
  "properties": {
    "@type": {"const": "APr"},
    "id": {"type": "string", "format": "uri"},
    "uia": {"type": "string", "format": "uri"},
    "apa": {"type": "string", "format": "uri"},
    "method": {"type": "string"},
    "evidence": {"type": "object", "required": ["coverage", "risk"], "properties": {"coverage": {"type": "number", "minimum": 0, "maximum": 1}, "risk": {"type": "number", "minimum": 0, "maximum": 1}, "obligations": {"type": "array", "items": {"type": "string"}}}},
    "proof": {"type": "object"}
  }
}

### IBE
{
  "$id": "https://example.org/ais/ibe.schema.json",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["@type", "id", "uiaRef", "apaStepRef", "aprRef", "tcaRef", "nonce", "exp", "sig"],
  "properties": {
    "@type": {"const": "IBE"},
    "id": {"type": "string", "format": "uri"},
    "uiaRef": {"type": "string"},
    "apaStepRef": {"type": "string"},
    "aprRef": {"type": "string"},
    "tcaRef": {"type": "string"},
    "nonce": {"type": "string"},
    "exp": {"type": "string", "format": "date-time"},
    "sig": {"type": "string"}
  }
}

### TCA (excerpt)
{
  "$id": "https://example.org/ais/tca.schema.json",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["id", "operations", "operator", "proof"],
  "properties": {
    "id": {"type": "string"},
    "operations": {"type": "array"},
    "operator": {"type": "string"},
    "proof": {"type": "object"}
  }
}


