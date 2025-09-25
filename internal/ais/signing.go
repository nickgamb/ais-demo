package ais

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/base64"
    "encoding/json"
    "strings"
)

// Simple JWS HS256 (header.payload.signature) with JSON payload.
// Not a full JWT implementation; suitable for demo use.

func b64url(in []byte) string {
    s := base64.StdEncoding.EncodeToString(in)
    s = strings.TrimRight(strings.NewReplacer("+", "-", "/", "_", "=", "").Replace(s), "=")
    return s
}

func SignJWSObject(secret []byte, v any) (string, error) {
    header := map[string]string{"alg": "HS256", "typ": "JWT"}
    hb, _ := json.Marshal(header)
    pb, err := json.Marshal(v)
    if err != nil { return "", err }
    headerPart := b64url(hb)
    payloadPart := b64url(pb)
    signingInput := headerPart + "." + payloadPart
    mac := hmac.New(sha256.New, secret)
    mac.Write([]byte(signingInput))
    sig := b64url(mac.Sum(nil))
    return signingInput + "." + sig, nil
}

func VerifyJWSObject(secret []byte, v any, jws string) (bool, error) {
    parts := strings.Split(jws, ".")
    if len(parts) != 3 { return false, nil }
    signingInput := parts[0] + "." + parts[1]
    mac := hmac.New(sha256.New, secret)
    mac.Write([]byte(signingInput))
    sig := b64url(mac.Sum(nil))
    if sig != parts[2] { return false, nil }
    // Optionally compare payload; not strictly necessary here
    // but helps ensure the payload corresponds to v
    pb, err := json.Marshal(v)
    if err != nil { return false, err }
    if parts[1] != b64url(pb) { return false, nil }
    return true, nil
}


