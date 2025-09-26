package ais

import (
    "errors"
    "slices"
    "strings"
    "sync"
    "time"
)

type GuardConfig struct { Secret []byte; MinAlignment float64; VerifierMethod string }

var (
    nonceSeen = struct{ sync.Mutex; m map[string]time.Time }{m: map[string]time.Time{}}
    crl = struct{
        sync.RWMutex
        issued time.Time
        expires time.Time
        revoked map[string]string // id -> reason
    }{revoked: map[string]string{}}
)

// SetCRL updates an in-memory CRL used by VerifyIBE.
func SetCRL(issued, expires time.Time, ids map[string]string) {
    crl.Lock()
    crl.issued, crl.expires = issued, expires
    crl.revoked = map[string]string{}
    for k, v := range ids { crl.revoked[k] = v }
    crl.Unlock()
}

func isRevoked(id string, now time.Time) bool {
    crl.RLock()
    defer crl.RUnlock()
    if crl.expires.IsZero() || now.After(crl.expires) { return false }
    _, ok := crl.revoked[id]
    return ok
}

func VerifyIBE(cfg GuardConfig, ibe IBE, apr APr, uia UIA, apa APA, tca TCA) error {
    now := time.Now()
    if now.After(ibe.Exp) { return errors.New("IBE-EXPIRED") }
    // simple replay cache by nonce
    nonceSeen.Lock()
    if t, ok := nonceSeen.m[ibe.Nonce]; ok && now.Sub(t) < 10*time.Minute { nonceSeen.Unlock(); return errors.New("IBE-REPLAY") }
    nonceSeen.m[ibe.Nonce] = now
    nonceSeen.Unlock()
    // Verify signature over the envelope WITHOUT the Sig field
    ibeForSig := ibe
    ibeForSig.Sig = ""
    ok, err := VerifyJWSObject(cfg.Secret, ibeForSig, ibe.Sig)
    if err != nil || !ok { return errors.New("IBE-SIG-INVALID") }
    // Verify UIA/APA/APr signatures if present
    if sig, _ := uia.Proof["jws"].(string); sig != "" {
        uiaForSig := uia; uiaForSig.Proof = map[string]any{}
        if ok, _ := VerifyJWSObject(cfg.Secret, uiaForSig, sig); !ok { return errors.New("UIA-SIG-INVALID") }
    }
    if sig, _ := apa.Proof["jws"].(string); sig != "" {
        apaForSig := apa; apaForSig.Proof = map[string]any{}
        if ok, _ := VerifyJWSObject(cfg.Secret, apaForSig, sig); !ok { return errors.New("APA-SIG-INVALID") }
    }
    if sig, _ := apr.Proof["jws"].(string); sig != "" {
        aprForSig := apr; aprForSig.Proof = map[string]any{}
        if ok, _ := VerifyJWSObject(cfg.Secret, aprForSig, sig); !ok { return errors.New("APR-SIG-INVALID") }
    }
    var cov, risk float64
    switch cfg.VerifierMethod {
    case "external-policy-v1":
        cov, risk = VerifyAlignmentExternalPolicy(uia, apa, ExternalPolicy{Keywords: extractKeywords(strings.ToLower(uia.Purpose)), WriteRisk: 0.5})
    case "classifier-v1":
        cov, risk = VerifyAlignmentClassifier(uia, apa)
    default:
        cov, risk = VerifyAlignment(uia, apa)
    }
    if apr.Method == "semantic-entailment-v1" {
        // Require that evidence matches recomputed values within tolerance
        if abs(apr.Evidence.Coverage-cov) > 1e-9 || abs(apr.Evidence.Risk-risk) > 1e-9 {
            return errors.New("apr evidence mismatch")
        }
    }
    if cov < cfg.MinAlignment { return errors.New("ALIGN-BELOW-THRESHOLD") }
    // Revocation checks
    if isRevoked(uia.ID, now) { return errors.New("UIA-REVOKED") }
    if isRevoked(apa.ID, now) { return errors.New("APA-REVOKED") }
    // Risk budgets: writes, records, external calls
    if apa.Totals.PredictedWrites > uia.RiskBudget.MaxWrites { return errors.New("RISK-WRITES-EXCEEDED") }
    if apa.Totals.PredictedRecords > uia.RiskBudget.MaxRecords { return errors.New("RISK-RECORDS-EXCEEDED") }
	var step *APAStep
	for i := range apa.Steps { if apa.Steps[i].ID == ibe.APAStepRef { step = &apa.Steps[i]; break } }
    if step == nil { return errors.New("IBE-STEP-NOT-FOUND") }
    for _, dc := range step.Expected.DataClasses { if !slices.Contains(uia.Constraints.DataClasses, dc) { return errors.New("DATA-CLASS-NOT-PERMITTED") } }
    var op *TCAOperation
	for i := range tca.Operations { if tca.Operations[i].Name == step.Tool { op = &tca.Operations[i]; break } }
    if op == nil { return errors.New("TCA-OP-NOT-ALLOWED") }
    // Verify TCA operator proof if present
    if sig, _ := tca.Proof["jws"].(string); sig != "" {
        t := tca
        t.Proof = map[string]any{}
        if ok, _ := VerifyJWSObject(cfg.Secret, t, sig); !ok { return errors.New("TCA-SIG-INVALID") }
    }
    if step.Expected.Writes > op.Effects.Writes { return errors.New("TCA-EFFECTS-EXCEEDED") }
    // Optional destination membrane check
    if len(op.Effects.Destinations) > 0 {
        // naive destination extraction from url arg
        if u, ok := step.Args["url"].(string); ok {
            allowed := false
            for _, d := range op.Effects.Destinations { if strings.Contains(u, d) { allowed = true; break } }
            if !allowed { return errors.New("DESTINATION-NOT-ALLOWED") }
        }
    }
    // Basic arg validation by tool name (placeholder for JSON Schema)
    if step.Tool == "http.get" {
        if u, ok := step.Args["url"].(string); !ok || !(strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://")) { return errors.New("INPUT-SCHEMA-INVALID") }
    }
    if step.Tool == "ollama.generate" {
        if p, ok := step.Args["prompt"].(string); !ok || len(p) == 0 { return errors.New("INPUT-SCHEMA-INVALID") }
    }
	return nil
}

// VerifyAlignment computes deterministic coverage and risk for APA against UIA.
// Coverage: fraction of steps whose tool/args semantically entail the purpose text.
// Risk: increases with predicted writes and if purpose suggests write/export actions.
func VerifyAlignment(uia UIA, apa APA) (coverage float64, risk float64) {
    total := len(apa.Steps)
    if total == 0 { return 0, 1 }
    purpose := strings.ToLower(uia.Purpose)
    entailKeywords := extractKeywords(purpose)
    aligned := 0
    for _, s := range apa.Steps {
        if stepEntailsPurpose(s, entailKeywords) { aligned++ }
    }
    coverage = float64(aligned) / float64(total)
    risk = 0
    if apa.Totals.PredictedWrites > 0 { risk += 0.5 }
    writey := []string{"send","post","write","export","email","delete"}
    for _, w := range writey { if strings.Contains(purpose, w) { risk += 0.3; break } }
    if risk > 1 { risk = 1 }
    if risk < 0 { risk = 0 }
    return
}

func stepEntailsPurpose(s APAStep, purposeTerms []string) bool {
    // Read-only generate/get steps are aligned if prompts/urls mention purpose terms
    if s.Tool == "ollama.generate" {
        if p, ok := s.Args["prompt"].(string); ok {
            p = strings.ToLower(p)
            return containsAny(p, purposeTerms)
        }
    }
    if s.Tool == "http.get" {
        if u, ok := s.Args["url"].(string); ok {
            u = strings.ToLower(u)
            return containsAny(u, purposeTerms) || len(purposeTerms) == 0
        }
    }
    return false
}

func extractKeywords(text string) []string {
    // naive keyword extraction: keep alphabetic tokens >=4 chars
    words := strings.FieldsFunc(text, func(r rune) bool { return !(r >= 'a' && r <= 'z') })
    out := make([]string, 0, len(words))
    seen := map[string]bool{}
    for _, w := range words {
        if len(w) < 4 { continue }
        if seen[w] { continue }
        seen[w] = true
        out = append(out, w)
    }
    return out
}

func containsAny(s string, terms []string) bool {
    for _, t := range terms { if strings.Contains(s, t) { return true } }
    return false
}

func abs(f float64) float64 { if f < 0 { return -f }; return f }


