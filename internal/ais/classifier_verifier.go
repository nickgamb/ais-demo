package ais

import "strings"

// VerifyAlignmentClassifier is a simple deterministic classifier-like scorer
// based on keyword frequency overlap between purpose and step prompts/urls.
func VerifyAlignmentClassifier(uia UIA, apa APA) (coverage float64, risk float64) {
    total := len(apa.Steps)
    if total == 0 { return 0, 1 }
    purpose := strings.ToLower(uia.Purpose)
    terms := extractKeywords(purpose) // reuse guard's extractor via same package
    aligned := 0
    for _, s := range apa.Steps {
        var text string
        if s.Tool == "ollama.generate" {
            if p, ok := s.Args["prompt"].(string); ok { text = p }
        } else if s.Tool == "http.get" {
            if u, ok := s.Args["url"].(string); ok { text = u }
        }
        text = strings.ToLower(text)
        matches := 0
        for _, t := range terms { if strings.Contains(text, t) { matches++ } }
        if len(terms) == 0 || matches*2 >= len(terms) { aligned++ }
    }
    coverage = float64(aligned) / float64(total)
    risk = 0
    if apa.Totals.PredictedWrites > 0 { risk += 0.5 }
    if risk > 1 { risk = 1 }
    return
}


