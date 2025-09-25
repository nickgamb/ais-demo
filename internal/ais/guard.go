package ais

import (
    "errors"
    "slices"
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
	if apr.Evidence.Coverage < cfg.MinAlignment { return errors.New("alignment below threshold") }
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


