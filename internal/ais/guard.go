package ais

import (
    "errors"
    "slices"
    "strings"
    "time"
)

type GuardConfig struct { Secret []byte; MinAlignment float64 }

func VerifyIBE(cfg GuardConfig, ibe IBE, apr APr, uia UIA, apa APA, tca TCA) error {
    if time.Now().After(ibe.Exp) { return errors.New("ibe expired") }
    // Verify signature over the envelope WITHOUT the Sig field
    ibeForSig := ibe
    ibeForSig.Sig = ""
    ok, err := VerifyJWSObject(cfg.Secret, ibeForSig, ibe.Sig)
    if err != nil || !ok { return errors.New("invalid ibe signature") }
    // Verify APr deterministically based on simple semantic entailment heuristic
    cov, risk := VerifyAlignment(uia, apa)
    if apr.Method == "semantic-entailment-v1" {
        // Require that evidence matches recomputed values within tolerance
        if abs(apr.Evidence.Coverage-cov) > 1e-9 || abs(apr.Evidence.Risk-risk) > 1e-9 {
            return errors.New("apr evidence mismatch")
        }
    }
    if cov < cfg.MinAlignment { return errors.New("alignment below threshold") }
	if apa.Totals.PredictedWrites > uia.RiskBudget.MaxWrites { return errors.New("risk budget exceeded: writes") }
	var step *APAStep
	for i := range apa.Steps { if apa.Steps[i].ID == ibe.APAStepRef { step = &apa.Steps[i]; break } }
	if step == nil { return errors.New("apa step not found") }
	for _, dc := range step.Expected.DataClasses { if !slices.Contains(uia.Constraints.DataClasses, dc) { return errors.New("step data class not permitted by UIA") } }
	var op *TCAOperation
	for i := range tca.Operations { if tca.Operations[i].Name == step.Tool { op = &tca.Operations[i]; break } }
	if op == nil { return errors.New("tool op not allowed by TCA") }
	if step.Expected.Writes > op.Effects.Writes { return errors.New("step writes exceed TCA effects") }
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


