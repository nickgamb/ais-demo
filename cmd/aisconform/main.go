package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "math"
    "os/exec"
    "os"
    "path/filepath"
    "time"

    "ais-demo/internal/ais"
)

func mustReadJSON[T any](path string) T {
    b, err := ioutil.ReadFile(path)
    if err != nil { panic(err) }
    var v T
    if err := json.Unmarshal(b, &v); err != nil { panic(fmt.Errorf("%s: %w", path, err)) }
    return v
}

func closeEnough(a, b float64) bool { return math.Abs(a-b) <= 1e-9 }

func emitGolden() {
    // Produce canonical JWS payloads for UIA/APA/APr using the demo secret
    secret := []byte("dev-secret-change-me")
    uia := mustReadJSON[ais.UIA]("spec/test-vectors/uia_minimal.json")
    apa := mustReadJSON[ais.APA]("spec/test-vectors/apa_generate_step.json")
    apr := mustReadJSON[ais.APr]("spec/test-vectors/apr_pass_semantic_entailment_v1.json")
    // Clear proofs to avoid recursion
    uia.Proof = map[string]any{}
    apa.Proof = map[string]any{}
    apr.Proof = map[string]any{}
    if j, err := ais.SignJWSObject(secret, uia); err == nil { uia.Proof = map[string]any{"jws": j} }
    if j, err := ais.SignJWSObject(secret, apa); err == nil { apa.Proof = map[string]any{"jws": j} }
    if j, err := ais.SignJWSObject(secret, apr); err == nil { apr.Proof = map[string]any{"jws": j} }
    // Write out golden files
    _ = os.WriteFile("spec/test-vectors/golden_uia_signed.json", mustJSON(uia), 0644)
    _ = os.WriteFile("spec/test-vectors/golden_apa_signed.json", mustJSON(apa), 0644)
    _ = os.WriteFile("spec/test-vectors/golden_apr_signed.json", mustJSON(apr), 0644)
    fmt.Println("Golden JWS files written under spec/test-vectors/")
}

func mustJSON(v any) []byte {
    b, _ := json.MarshalIndent(v, "", "  ")
    return b
}

func main() {
    base := "spec/test-vectors"
    if len(os.Args) > 1 && os.Args[1] != "--emit-golden" { base = os.Args[1] }
    if len(os.Args) > 1 && os.Args[1] == "--emit-golden" {
        emitGolden()
        return
    }
    fmt.Printf("AIS Conformance — test vectors in %s\n\n", base)

    // Load vectors
    uia := mustReadJSON[ais.UIA](filepath.Join(base, "uia_minimal.json"))
    apa := mustReadJSON[ais.APA](filepath.Join(base, "apa_generate_step.json"))
    aprPass := mustReadJSON[ais.APr](filepath.Join(base, "apr_pass_semantic_entailment_v1.json"))
    aprFail := mustReadJSON[ais.APr](filepath.Join(base, "apr_fail_low_coverage.json"))
    ibeExpired := mustReadJSON[ais.IBE](filepath.Join(base, "ibe_expired.json"))

    total := 0
    failed := 0
    fail := func(name string, msg string) {
        failed++
        fmt.Printf("[FAIL] %s: %s\n", name, msg)
    }
    pass := func(name string) {
        fmt.Printf("[PASS] %s\n", name)
    }

    // Test 1: semantic-entailment-v1 pass vector matches recomputation
    total++
    cov, risk := ais.VerifyAlignment(uia, apa)
    if aprPass.Method != "semantic-entailment-v1" || !closeEnough(aprPass.Evidence.Coverage, cov) || !closeEnough(aprPass.Evidence.Risk, risk) {
        fail("apr_pass_semantic_entailment_v1", fmt.Sprintf("expected coverage=%v risk=%v got coverage=%v risk=%v", aprPass.Evidence.Coverage, aprPass.Evidence.Risk, cov, risk))
    } else { pass("apr_pass_semantic_entailment_v1") }

    // Test 2: semantic-entailment-v1 fail vector should not match recomputation
    total++
    if aprFail.Method == "semantic-entailment-v1" && (!closeEnough(aprFail.Evidence.Coverage, cov) || !closeEnough(aprFail.Evidence.Risk, risk)) {
        pass("apr_fail_low_coverage (mismatch as expected)")
    } else {
        fail("apr_fail_low_coverage", fmt.Sprintf("unexpectedly matched coverage=%v risk=%v", cov, risk))
    }

    // Test 3: IBE expired should be rejected by guard
    total++
    // Force expirations in the past if vector not already
    if time.Now().Before(ibeExpired.Exp) { ibeExpired.Exp = time.Now().Add(-time.Minute) }
    cfg := ais.GuardConfig{Secret: []byte("dev-secret-change-me"), MinAlignment: 0.8}
    // Dummy APR/UIA/APA/TCA are fine for expiry branch.
    if err := ais.VerifyIBE(cfg, ibeExpired, ais.APr{}, uia, apa, ais.TCA{}); err == nil {
        fail("ibe_expired", "expected failure but got nil")
    } else {
        pass("ibe_expired")
    }

    fmt.Printf("\nSummary: %d/%d passed\n", total-failed, total)
    if failed > 0 { os.Exit(1) }
}


