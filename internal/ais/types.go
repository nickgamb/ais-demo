package ais

import "time"

type UIA struct {
	Type          string       `json:"@type"`
	ID            string       `json:"id"`
	Subject       Principal    `json:"subject"`
	Purpose       string       `json:"purpose"`
	Constraints   Constraints  `json:"constraints"`
	RiskBudget    RiskBudget   `json:"riskBudget"`
	PolicyProfile string       `json:"policyProfile"`
	Proof         map[string]any `json:"proof"`
}

type Principal struct { ID string `json:"id"` }

type Constraints struct {
	DataClasses   []string  `json:"dataClasses"`
	Jurisdictions []string  `json:"jurisdictions,omitempty"`
	TimeWindow    TimeBound `json:"timeWindow"`
}

type TimeBound struct { NotAfter time.Time `json:"notAfter"` }

type RiskBudget struct {
    Level            int `json:"level"`
    MaxWrites        int `json:"maxWrites"`
    MaxRecords       int `json:"maxRecords"`
    MaxExternalCalls int `json:"maxExternalCalls,omitempty"`
}

type APA struct {
	Type   string       `json:"@type"`
	ID     string       `json:"id"`
	UIA    string       `json:"uia"`
	Model  ModelInfo    `json:"model"`
	Steps  []APAStep    `json:"steps"`
	Totals APATotals    `json:"totals"`
	Proof  map[string]any `json:"proof"`
}

type ModelInfo struct {
    Vendor  string `json:"vendor,omitempty"`
    Version string `json:"version,omitempty"`
    Hash    string `json:"hash"`
}

type APAStep struct {
	ID       string         `json:"id"`
	Tool     string         `json:"tool"`
	Args     map[string]any `json:"args"`
	Expected StepExpected   `json:"expected"`
	Alignment StepAlignment `json:"alignment"`
}

type StepExpected struct {
    DataClasses []string `json:"dataClasses"`
    Writes      int      `json:"writes"`
}
type StepAlignment struct {
    Score float64 `json:"score"`
    Why   string  `json:"why,omitempty"`
}
type APATotals struct {
    PredictedWrites  int `json:"predictedWrites"`
    PredictedRecords int `json:"predictedRecords"`
}

type APr struct {
	Type     string         `json:"@type"`
	ID       string         `json:"id"`
	UIA      string         `json:"uia"`
	APA      string         `json:"apa"`
	Method   string         `json:"method"`
	Evidence APrEvidence    `json:"evidence"`
	Proof    map[string]any `json:"proof"`
}

type APrEvidence struct {
    Coverage float64 `json:"coverage"`
    Risk     float64 `json:"risk"`
}

type IBE struct {
	Type      string    `json:"@type"`
	ID        string    `json:"id"`
	UIARef    string    `json:"uiaRef"`
	APAStepRef string   `json:"apaStepRef"`
	APrRef    string    `json:"aprRef"`
	TCARef    string    `json:"tcaRef"`
	Nonce     string    `json:"nonce"`
	Exp       time.Time `json:"exp"`
	Sig       string    `json:"sig"`
}

type TCA struct {
    ID        string         `json:"id"`
    Operator  string         `json:"operator"`
    Operations []TCAOperation `json:"operations"`
}
type TCAOperation struct {
    Name    string           `json:"name"`
    Effects OperationEffects `json:"effects"`
}
type OperationEffects struct {
    Writes      int      `json:"writes"`
    DataClasses []string `json:"dataClasses"`
}


