package idempotency

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
)

// RequestHash returns a deterministic SHA-256 hash for JSON API requests.
//
// The body is canonicalized by decoding JSON into generic Go values and encoding it again with encoding/json.
// encoding/json emits object keys in deterministic order, while preserving array order and scalar values. This is
// intentionally simple for MVP JSON requests and should be replaced by a stricter RFC 8785 canonicalizer if we later
// need cross-language byte-for-byte canonical JSON semantics.
func RequestHash(method string, path string, body []byte) (string, error) {
	canonicalBody, err := CanonicalJSON(body)
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256([]byte(strings.ToUpper(strings.TrimSpace(method)) + "\n" + path + "\n" + string(canonicalBody)))
	return hex.EncodeToString(sum[:]), nil
}

func CanonicalJSON(body []byte) ([]byte, error) {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		return []byte("null"), nil
	}

	var value any
	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}

	return json.Marshal(value)
}
