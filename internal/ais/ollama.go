package ais

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

type OllamaClient struct { BaseURL, Model string; HTTP *http.Client }

type ollamaReq struct { Model, Prompt string; Stream bool }
type ollamaResp struct { Response string }
type pullReq struct { Name string `json:"name"` }
type tagsResp struct {
    Models []struct{ Name string `json:"name"` } `json:"models"`
}

func (c *OllamaClient) Generate(prompt string) (string, error) {
	req := ollamaReq{Model: c.Model, Prompt: prompt, Stream: false}
	b, _ := json.Marshal(req)
	resp, err := c.HTTP.Post(c.BaseURL+"/api/generate", "application/json", bytes.NewReader(b))
	if err != nil { return "", err }
	defer resp.Body.Close()
    if resp.StatusCode != 200 {
        // Try to pull the model once, then retry
        if err := c.EnsureModel(); err == nil {
            // retry
            resp2, err2 := c.HTTP.Post(c.BaseURL+"/api/generate", "application/json", bytes.NewReader(b))
            if err2 != nil { return "", err2 }
            defer resp2.Body.Close()
            if resp2.StatusCode != 200 {
                body, _ := io.ReadAll(resp2.Body)
                return "", fmt.Errorf("ollama non-200 after pull: %d: %s", resp2.StatusCode, string(body))
            }
            var out2 ollamaResp
            if err := json.NewDecoder(resp2.Body).Decode(&out2); err != nil { return "", err }
            return out2.Response, nil
        }
        body, _ := io.ReadAll(resp.Body)
        return "", fmt.Errorf("ollama non-200: %d: %s", resp.StatusCode, string(body))
    }
	var out ollamaResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil { return "", err }
	return out.Response, nil
}

// EnsureModel pulls the configured model if needed, ignoring already-present cases.
func (c *OllamaClient) EnsureModel() error {
    pr := pullReq{Name: c.Model}
    b, _ := json.Marshal(pr)
    resp, err := c.HTTP.Post(c.BaseURL+"/api/pull", "application/json", bytes.NewReader(b))
    if err != nil { return err }
    defer resp.Body.Close()
    if resp.StatusCode != 200 {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("pull failed: %d: %s", resp.StatusCode, string(body))
    }
    // Consume streaming progress until EOF
    dec := json.NewDecoder(resp.Body)
    for dec.More() {
        var m map[string]any
        if err := dec.Decode(&m); err != nil { break }
    }
    return nil
}

// HasModel checks if the configured model is already available.
func (c *OllamaClient) HasModel() (bool, error) {
    resp, err := c.HTTP.Get(c.BaseURL+"/api/tags")
    if err != nil { return false, err }
    defer resp.Body.Close()
    if resp.StatusCode != 200 { return false, fmt.Errorf("tags status: %d", resp.StatusCode) }
    var tr tagsResp
    if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil { return false, err }
    for _, m := range tr.Models {
        if m.Name == c.Model { return true, nil }
    }
    return false, nil
}

// ListModels returns available model names from Ollama /api/tags
func (c *OllamaClient) ListModels() ([]string, error) {
    resp, err := c.HTTP.Get(c.BaseURL+"/api/tags")
    if err != nil { return nil, err }
    defer resp.Body.Close()
    if resp.StatusCode != 200 { return nil, fmt.Errorf("tags status: %d", resp.StatusCode) }
    var tr tagsResp
    if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil { return nil, err }
    out := make([]string, 0, len(tr.Models))
    for _, m := range tr.Models { out = append(out, m.Name) }
    return out, nil
}

// PullStream streams model pull progress, invoking cb for each progress JSON object.
func (c *OllamaClient) PullStream(cb func(map[string]any) error) error {
    pr := pullReq{Name: c.Model}
    b, _ := json.Marshal(pr)
    resp, err := c.HTTP.Post(c.BaseURL+"/api/pull", "application/json", bytes.NewReader(b))
    if err != nil { return err }
    defer resp.Body.Close()
    if resp.StatusCode != 200 {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("pull failed: %d: %s", resp.StatusCode, string(body))
    }
    dec := json.NewDecoder(resp.Body)
    for dec.More() {
        var m map[string]any
        if err := dec.Decode(&m); err != nil { return err }
        // compute percent if fields present
        if t, ok := m["total"].(float64); ok && t > 0 {
            if cpl, ok2 := m["completed"].(float64); ok2 {
                p := (cpl / t) * 100.0
                if p < 0 { p = 0 }
                if p > 100 { p = 100 }
                m["percent"] = p
            }
        }
        if err := cb(m); err != nil { return err }
    }
    return nil
}


