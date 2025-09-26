package ais

import "strings"

// ExternalPolicy encapsulates minimal policy inputs for external-policy-v1 verifier.
type ExternalPolicy struct {
    Keywords  []string `json:"keywords"`
    WriteRisk float64  `json:"writeRisk"`
}

// VerifyAlignmentExternalPolicy computes coverage/risk based on explicit policy keywords
// and a fixed write risk contribution. This is a deterministic profile.
func VerifyAlignmentExternalPolicy(uia UIA, apa APA, pol ExternalPolicy) (coverage float64, risk float64) {
    total := len(apa.Steps)
    if total == 0 { return 0, 1 }
    aligned := 0
    kws := make([]string, 0, len(pol.Keywords))
    for _, k := range pol.Keywords { if k != "" { kws = append(kws, strings.ToLower(k)) } }
    for _, s := range apa.Steps {
        if s.Tool == "ollama.generate" {
            if p, ok := s.Args["prompt"].(string); ok && containsAny(strings.ToLower(p), kws) { aligned++ }
        } else if s.Tool == "http.get" {
            if u, ok := s.Args["url"].(string); ok && containsAny(strings.ToLower(u), kws) { aligned++ }
        }
    }
    coverage = float64(aligned) / float64(total)
    risk = 0
    if apa.Totals.PredictedWrites > 0 { risk += pol.WriteRisk }
    if risk < 0 { risk = 0 }
    if risk > 1 { risk = 1 }
    return
}


