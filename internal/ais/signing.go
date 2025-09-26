package ais

import (
    "bytes"
    "crypto/hmac"
    "crypto/sha256"
    "encoding/base64"
    "encoding/json"
    "sort"
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
    pb, err := marshalCanonical(v)
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
    // Compare canonical payload to ensure correspondence
    pb, err := marshalCanonical(v)
    if err != nil { return false, err }
    if parts[1] != b64url(pb) { return false, nil }
    return true, nil
}

// marshalCanonical produces lexicographically sorted JSON keys for stable signing.
func marshalCanonical(v any) ([]byte, error) {
    var raw any
    b, err := json.Marshal(v)
    if err != nil { return nil, err }
    if err := json.Unmarshal(b, &raw); err != nil { return nil, err }
    buf := &bytes.Buffer{}
    if err := encodeCanonical(buf, raw); err != nil { return nil, err }
    return buf.Bytes(), nil
}

func encodeCanonical(buf *bytes.Buffer, v any) error {
    switch x := v.(type) {
    case map[string]any:
        buf.WriteByte('{')
        keys := make([]string, 0, len(x))
        for k := range x { keys = append(keys, k) }
        sort.Strings(keys)
        for i, k := range keys {
            if i > 0 { buf.WriteByte(',') }
            jb, _ := json.Marshal(k)
            buf.Write(jb)
            buf.WriteByte(':')
            if err := encodeCanonical(buf, x[k]); err != nil { return err }
        }
        buf.WriteByte('}')
    case []any:
        buf.WriteByte('[')
        for i, e := range x {
            if i > 0 { buf.WriteByte(',') }
            if err := encodeCanonical(buf, e); err != nil { return err }
        }
        buf.WriteByte(']')
    default:
        jb, _ := json.Marshal(x)
        buf.Write(jb)
    }
    return nil
}


