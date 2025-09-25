package ais

import (
    "errors"
    "io"
    "net/http"
)

// Simple HTTP GET tool
type HTTPTool struct { HTTP *http.Client }

func (h *HTTPTool) Get(url string) (string, error) {
    resp, err := h.HTTP.Get(url)
    if err != nil { return "", err }
    defer resp.Body.Close()
    if resp.StatusCode != 200 { return "", errors.New("non-200 from http.get") }
    b, _ := io.ReadAll(resp.Body)
    if len(b) > 2000 { b = b[:2000] }
    return string(b), nil
}


