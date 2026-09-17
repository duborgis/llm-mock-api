package bedrock

import (
	"encoding/json"
	"io"
)

func writeJSON(w io.Writer, v any) error {
	return json.NewEncoder(w).Encode(v)
}

func readJSON(r io.Reader, v any) error {
	return json.NewDecoder(r).Decode(v)
}
