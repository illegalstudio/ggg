package cmd

import (
	"encoding/json"
	"fmt"
	"os"
)

// jsonOutput is set by the persistent --json flag on the root command.
// When true, commands emit machine-readable JSON to stdout instead of
// human-formatted text, skip spinners, skip interactive confirmations,
// and refuse to disambiguate repositories interactively.
var jsonOutput bool

type jsonPayloadError struct {
	message string
	payload any
}

func (e *jsonPayloadError) Error() string {
	return e.message
}

func (e *jsonPayloadError) JSONPayload() any {
	return e.payload
}

// emitJSON encodes v as indented JSON and writes it to stdout.
func emitJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// maybeJSON emits v as JSON when --json is enabled and signals callers to
// stop their text-based output. Pattern:
//
//	if done, err := maybeJSON(data); done {
//	    return err
//	}
//	// ... text rendering ...
func maybeJSON(v any) (bool, error) {
	if !jsonOutput {
		return false, nil
	}
	return true, emitJSON(v)
}

// errString returns the error's message or an empty string if err is nil.
// Used to produce JSON-friendly fields without omitempty surprises.
func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func unsupportedJSON(command string) error {
	msg := fmt.Sprintf("--json is not supported for %s", command)
	return &jsonPayloadError{
		message: msg,
		payload: map[string]any{
			"supported": false,
			"command":   command,
			"error":     msg,
		},
	}
}
